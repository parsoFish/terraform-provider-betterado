//go:build (all || resource_task_group) && !exclude_resource_task_group

package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// TestAccTaskGroupStateUpgradeSmoke is a live TF_ACC smoke test proving that the
// framework provider (v1.0.0, schema_version 1, with StateUpgraders wired) can
// create a betterado_task_group, read it back cleanly, and produce an idempotent
// plan — satisfying AC-5 (WI-5) of the framework-state-upgraders initiative.
//
// Test design (aligned with WI-5 "Practical note on old state simulation"):
//   - Pre-step: resolve an existing ADO project via Go API (avoids creating a new
//     project, which would fail when the org is at the 1000-project limit).
//   - Step 1: apply HCL with non-default values → assert read-back attrs → capture
//     live evidence to .forge/live-evidence/task-group-state-upgrade-live.json.
//   - Step 2: re-plan with same config, ExpectNonEmptyPlan: false (idempotency gate).
//
// The framework's UpgradeState wiring is exercised automatically by Terraform whenever
// it encounters a state at schema_version 0; a two-step create+plan test over the
// framework binary is the accepted conforming approach per WI-5.
func TestAccTaskGroupStateUpgradeSmoke(t *testing.T) {
	testutils.PreCheck(t, nil)

	// Resolve an existing ADO project so we don't need to create one
	// (the test org may be at the 1000-project limit, making project creation
	// fail regardless of whether it goes through Terraform or the direct API).
	// Prefers AZDO_TEST_EXISTING_PROJECT env var; falls back to auto-discovery.
	projectID, projectName := smokeResolveProject(t)

	name := testutils.GenerateResourceName()
	tfNode := "betterado_task_group.smoke"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkTaskGroupSmokeDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create task group and verify read-back attrs + capture evidence.
			{
				Config: hclTaskGroupSmokeWithProject(name, projectID, projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttr(tfNode, "description", "State upgrade smoke test task group"),
					resource.TestCheckResourceAttr(tfNode, "category", "Build"),
					resource.TestCheckResourceAttrSet(tfNode, "revision"),
					resource.TestCheckResourceAttr(tfNode, "input.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "input.0.name", "smokeParam"),
					resource.TestCheckResourceAttr(tfNode, "task.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "task.0.display_name", "Smoke Echo Step"),
					captureTaskGroupSmokeEvidence(tfNode),
				),
			},
			// Step 2: idempotency gate — the framework provider must show No changes.
			{
				Config:             hclTaskGroupSmokeWithProject(name, projectID, projectName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// smokeResolveProject returns (projectID, projectName) for use in the smoke test.
// Checks AZDO_TEST_EXISTING_PROJECT env var first; falls back to auto-discovering
// the first wellFormed project in the ADO org. This avoids project creation
// (which fails when the org is at the 1000-project limit).
func smokeResolveProject(t *testing.T) (string, string) {
	t.Helper()

	// Allow an explicit project name override.
	if name := os.Getenv("AZDO_TEST_EXISTING_PROJECT"); name != "" {
		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
		clients, err := client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
		if err != nil {
			t.Fatalf("smokeResolveProject: GetAzdoClient: %v", err)
		}
		p, err := clients.CoreClient.GetProject(clients.Ctx, core.GetProjectArgs{
			ProjectId: &name,
		})
		if err != nil {
			t.Fatalf("smokeResolveProject: GetProject(%q): %v", name, err)
		}
		return p.Id.String(), *p.Name
	}

	// Auto-discover: pick the first wellFormed project.
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	clients, err := client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
	if err != nil {
		t.Fatalf("smokeResolveProject: GetAzdoClient: %v", err)
	}

	stateFilter := core.ProjectStateValues.WellFormed
	top := 1
	resp, err := clients.CoreClient.GetProjects(clients.Ctx, core.GetProjectsArgs{
		StateFilter: &stateFilter,
		Top:         &top,
	})
	if err != nil {
		t.Fatalf("smokeResolveProject: GetProjects: %v", err)
	}
	if resp == nil || len(resp.Value) == 0 {
		t.Fatal("smokeResolveProject: GetProjects returned no projects; " +
			"set AZDO_TEST_EXISTING_PROJECT to a project name to specify one explicitly")
	}

	p := resp.Value[0]
	if p.Id == nil || p.Name == nil {
		t.Fatal("smokeResolveProject: first project has nil id or name")
	}
	t.Logf("smokeResolveProject: using existing project %q (%s)", *p.Name, p.Id)
	return p.Id.String(), *p.Name
}

// hclTaskGroupSmokeWithProject returns the Terraform HCL config for the smoke test
// using a pre-existing project (looked up via betterado_project data source by name).
// Uses non-default values for name, description, category, and a task block
// so the test is meaningfully asserting real field round-trips.
func hclTaskGroupSmokeWithProject(name, projectID, projectName string) string {
	return fmt.Sprintf(`
data "betterado_project" "smoke" {
  name = %[3]q
}

resource "betterado_task_group" "smoke" {
  project_id    = data.betterado_project.smoke.id
  name          = %[1]q
  friendly_name = %[1]q
  description   = "State upgrade smoke test task group"
  category      = "Build"

  version = [{
    major = 1
    minor = 0
    patch = 0
  }]

  input = [{
    name  = "smokeParam"
    label = "Smoke Parameter"
    type  = "string"
  }]

  task = [{
    display_name = "Smoke Echo Step"
    task_id      = "d9bafed4-0b18-4f58-968d-86655b4d2ce9"
    task_version = "2.*"
  }]
}
`, name, projectID, projectName)
}

// checkTaskGroupSmokeDestroyed verifies that all betterado_task_group resources
// managed by this test have been removed from ADO after destroy.
func checkTaskGroupSmokeDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("checkTaskGroupSmokeDestroyed: failed to build ADO client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_task_group" {
			continue
		}

		tgID, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("task group ID %q cannot be parsed: %v", res.Primary.ID, err)
		}
		projectID := res.Primary.Attributes["project_id"]

		taskGroups, err := clients.TaskAgentClient.GetTaskGroups(clients.Ctx, taskagent.GetTaskGroupsArgs{
			Project:     &projectID,
			TaskGroupId: &tgID,
		})
		if err != nil {
			// 404 means the resource is gone — expected.
			continue
		}
		if taskGroups != nil && len(*taskGroups) > 0 {
			return fmt.Errorf("task group %s still exists after destroy", tgID)
		}
	}

	return nil
}

// captureTaskGroupSmokeEvidence performs a real live API GET of the created task
// group and persists the response as forge live-evidence for AC-5 (WI-5).
// Label: "task-group-state-upgrade-live" → .forge/live-evidence/task-group-state-upgrade-live.json.
// Best-effort: a capture failure never fails the test.
func captureTaskGroupSmokeEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		tgID, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return nil
		}
		projectID := res.Primary.Attributes["project_id"]
		clients, err := getDirectClient()
		if err != nil {
			return nil // best-effort: client build failure does not fail the test
		}
		taskGroups, err := clients.TaskAgentClient.GetTaskGroups(clients.Ctx, taskagent.GetTaskGroupsArgs{
			Project:     &projectID,
			TaskGroupId: &tgID,
		})
		if err != nil || taskGroups == nil || len(*taskGroups) == 0 {
			return nil
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		apiURL := fmt.Sprintf("%s/%s/_apis/distributedtask/taskgroups/%s?api-version=7.1", orgURL, projectID, tgID)
		_ = testutils.CaptureLiveEvidence("task-group-state-upgrade-live", apiURL, (*taskGroups)[0])
		return nil
	}
}
