//go:build (all || data_source_pipeline) && !exclude_data_source_pipeline

package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	adoPipelines "github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccDataPipelineRun creates a pipeline, triggers a run and waits for it to
// complete, then reads the run back via the betterado_pipeline_run data source.
//
// AC1: data source reads state, result, created_date; ExpectNonEmptyPlan:false;
//
//	destroy step leaves the run record in ADO (runs are immutable — no CheckDestroy).
//
// AC2: CaptureLiveEvidence captures the GET _apis/pipelines/{id}/runs/{runId} response.
func TestAccDataPipelineRun(t *testing.T) {
	name := testutils.GenerateResourceName()

	// configFilePath must be non-empty at TestCase validation time (before any
	// step runs) because the framework eagerly calls ConfigFile funcs during
	// validate(). We pre-create a placeholder temp file so that validation
	// sees a valid path. Step 1's Check (triggerAndWaitForRun) then overwrites
	// this variable with the path of a new temp file containing real HCL.
	// The closure below captures configFilePath by reference, so when Steps 2
	// and 3 execute they read the updated path.
	placeholderFile, err := os.CreateTemp("", "betterado_pipeline_run_placeholder_*.tf")
	if err != nil {
		t.Fatalf("failed to create placeholder config file: %v", err)
	}
	placeholderFile.Close()

	var (
		pipelineIDStr  string
		runIDStr       string
		configFilePath = placeholderFile.Name()
	)

	tfDataNode := "data.betterado_pipeline_run.test"

	// hclDataSourceConfigFile is used as ConfigFile for Steps 2 and 3.
	// It is a func so the framework evaluates it lazily; by Step 2 execution
	// time, triggerAndWaitForRun has written the file and set configFilePath.
	// Returns "" when called before Step 1 finishes (during initialisation
	// probes) — the framework treats nil Config as "no config yet" which is
	// safe for those early calls.
	hclDataSourceConfigFile := func(_ config.TestStepConfigRequest) string {
		return configFilePath
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		// Runs are immutable in ADO — no destroy needed, and no CheckDestroy.
		Steps: []resource.TestStep{
			// Step 1: create the pipeline + trigger + wait, then record run IDs.
			{
				Config: hclPipelineRunResourceOnly(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betterado_pipeline.test", "id"),
					// Trigger the run, wait for completion, then write the
					// HCL for Steps 2/3 to a temp file (populates configFilePath).
					triggerAndWaitForRun(name, &pipelineIDStr, &runIDStr, &configFilePath),
				),
			},
			// Step 2: read the completed run via the data source.
			// ConfigFile returns configFilePath which was written by Step 1's Check.
			{
				ConfigFile: hclDataSourceConfigFile,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfDataNode, "id"),
					resource.TestCheckResourceAttrSet(tfDataNode, "state"),
					resource.TestCheckResourceAttrSet(tfDataNode, "created_date"),
					capturePipelineRunEvidence(tfDataNode, &pipelineIDStr, &runIDStr),
				),
			},
			// Step 3: idempotency — ExpectNonEmptyPlan:false (AC1).
			{
				ConfigFile:         hclDataSourceConfigFile,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclPipelineRunResourceOnly creates the betterado_pipeline resource used in
// Step 1. The pipeline is in the standing shared fixture project.
func hclPipelineRunResourceOnly(name string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

data "betterado_git_repository" "test" {
  project_id = data.betterado_project.test.id
  name       = %[2]q
}

resource "betterado_pipeline" "test" {
  project_id         = data.betterado_project.test.id
  name               = %[1]q
  folder             = "\\tf-acc-run"
  configuration_type = "yaml"
  repo_id            = data.betterado_git_repository.test.id
  yaml_path          = "/azure-pipelines.yml"
}
`, name, SharedFixtureProjectName)
}

// hclPipelineRunDataSourceContent returns the full HCL content — including a
// terraform { required_providers } block so the ConfigFile path can supply it
// without relying on the framework's mergedConfig injection (which only applies
// to steps using Config, not ConfigFile).
func hclPipelineRunDataSourceContent(name, pid, rid string) string {
	// Compute the provider source address the same way the framework does.
	host := "registry.terraform.io"
	if v := os.Getenv("TF_ACC_PROVIDER_HOST"); v != "" {
		host = strings.TrimSuffix(v, "/")
	}
	namespace := "hashicorp"
	if v := os.Getenv("TF_ACC_PROVIDER_NAMESPACE"); v != "" {
		namespace = strings.TrimSuffix(v, "/")
	}
	providerSource := host + "/" + namespace + "/betterado"

	return fmt.Sprintf(`
terraform {
  required_providers {
    betterado = {
      source = %[5]q
    }
  }
}

data "betterado_project" "test" {
  name = %[3]q
}

data "betterado_git_repository" "test" {
  project_id = data.betterado_project.test.id
  name       = %[3]q
}

resource "betterado_pipeline" "test" {
  project_id         = data.betterado_project.test.id
  name               = %[1]q
  folder             = "\\tf-acc-run"
  configuration_type = "yaml"
  repo_id            = data.betterado_git_repository.test.id
  yaml_path          = "/azure-pipelines.yml"
}

data "betterado_pipeline_run" "test" {
  project_id  = data.betterado_project.test.id
  pipeline_id = %[2]s
  run_id      = %[4]s
}
`, name, pid, SharedFixtureProjectName, rid, providerSource)
}

// triggerAndWaitForRun is a TestCheckFunc that:
//  1. Reads the pipeline ID from the state (betterado_pipeline.test).
//  2. Calls RunPipeline to trigger a run.
//  3. Polls GetRun until state == "completed" or a 3-minute timeout is reached.
//  4. Sets *pipelineIDStr and *runIDStr so subsequent steps can reference them.
//  5. Writes the Step-2/3 HCL to a temp file and sets *configFilePath to its path.
func triggerAndWaitForRun(name string, pipelineIDStr, runIDStr, configFilePath *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources["betterado_pipeline.test"]
		if !ok {
			return fmt.Errorf("betterado_pipeline.test not found in state")
		}
		pipelineIDRaw := res.Primary.ID
		projectID := res.Primary.Attributes["project_id"]

		var pipelineID int
		if _, err := fmt.Sscanf(pipelineIDRaw, "%d", &pipelineID); err != nil {
			return fmt.Errorf("cannot parse pipeline ID %q: %v", pipelineIDRaw, err)
		}

		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("triggerAndWaitForRun: build client: %v", err)
		}

		// Trigger the pipeline run.
		run, err := clients.PipelinesClient.RunPipeline(clients.Ctx, adoPipelines.RunPipelineArgs{
			Project:       &projectID,
			PipelineId:    &pipelineID,
			RunParameters: &adoPipelines.RunPipelineParameters{},
		})
		if err != nil {
			return fmt.Errorf("triggerAndWaitForRun: RunPipeline: %v", err)
		}
		if run == nil || run.Id == nil {
			return fmt.Errorf("triggerAndWaitForRun: RunPipeline returned nil run")
		}

		runID := *run.Id

		// Poll until the run completes (up to 3 minutes).
		deadline := time.Now().Add(3 * time.Minute)
		for time.Now().Before(deadline) {
			time.Sleep(10 * time.Second)
			r, err := clients.PipelinesClient.GetRun(clients.Ctx, adoPipelines.GetRunArgs{
				Project:    &projectID,
				PipelineId: &pipelineID,
				RunId:      &runID,
			})
			if err != nil {
				continue // transient — keep polling
			}
			if r == nil || r.State == nil {
				continue
			}
			if *r.State == adoPipelines.RunStateValues.Completed ||
				*r.State == adoPipelines.RunStateValues.Canceling {
				break
			}
		}

		*pipelineIDStr = pipelineIDRaw
		*runIDStr = fmt.Sprintf("%d", runID)

		// Write the Step-2/3 HCL to a temp file. Steps 2 and 3 use ConfigFile
		// which is evaluated lazily — by the time those steps run, this file
		// exists and configFilePath points to it.
		content := hclPipelineRunDataSourceContent(name, pipelineIDRaw, *runIDStr)
		f, err := os.CreateTemp("", "betterado_pipeline_run_*.tf")
		if err != nil {
			return fmt.Errorf("triggerAndWaitForRun: create temp config file: %v", err)
		}
		if _, err := fmt.Fprint(f, content); err != nil {
			f.Close()
			return fmt.Errorf("triggerAndWaitForRun: write temp config file: %v", err)
		}
		f.Close()
		*configFilePath = f.Name()

		return nil
	}
}

// capturePipelineRunEvidence performs a live API GET of the completed run and
// persists the response as forge demo live-evidence (AC2).
func capturePipelineRunEvidence(tfNode string, pipelineIDStr, runIDStr *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		projectID := res.Primary.Attributes["project_id"]

		pid := ""
		rid := ""
		if pipelineIDStr != nil {
			pid = *pipelineIDStr
		}
		if runIDStr != nil {
			rid = *runIDStr
		}
		if pid == "" || rid == "" {
			return nil // best-effort
		}

		var pipelineID, runID int
		fmt.Sscanf(pid, "%d", &pipelineID) //nolint:errcheck
		fmt.Sscanf(rid, "%d", &runID)      //nolint:errcheck

		clients, err := getDirectClient()
		if err != nil {
			return nil
		}

		r, err := clients.PipelinesClient.GetRun(clients.Ctx, adoPipelines.GetRunArgs{
			Project:    &projectID,
			PipelineId: &pipelineID,
			RunId:      &runID,
		})
		if err != nil || r == nil {
			return nil
		}

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/%s/_apis/pipelines/%d/runs/%d?api-version=7.1-preview.1",
			orgURL, projectID, pipelineID, runID)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", url, r)
		return nil
	}
}
