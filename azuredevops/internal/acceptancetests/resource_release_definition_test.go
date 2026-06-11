package acceptancetests

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// TestAccReleaseDefinition_basic tests creating and importing a minimal release definition.
// It uses SharedReleaseFixture (from WI-1) to obtain a pre-provisioned project, repo, and
// build definition rather than hand-rolling a betterado_project inline.
func TestAccReleaseDefinition_basic(t *testing.T) {
	fixture := SharedReleaseFixture(t)

	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionBasicFixture(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "revision"),
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttr(tfNode, "path", `\`),
					resource.TestCheckResourceAttr(tfNode, "environment.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.name", "Production"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.name", "Agent job"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccReleaseDefinition_withDeploymentInput tests that deployment_input (queue_id) is correctly
// sent to and read back from the API.
func TestAccReleaseDefinition_withDeploymentInput(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionWithDeploymentInput(name),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.#", "1"),
					resource.TestCheckResourceAttrSet(tfNode, "environment.0.deploy_phase.0.deployment_input.0.queue_id"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.0.condition", "succeeded()"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.0.timeout_in_minutes", "0"),
				),
			},
		},
	})
}

// TestAccReleaseDefinition_withApprovalOptions tests approval_options on pre/post approvals.
func TestAccReleaseDefinition_withApprovalOptions(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionWithApprovalOptions(name),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.0.approval_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.0.approval_options.0.release_creator_can_be_approver", "false"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.0.approval_options.0.timeout_in_minutes", "1440"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.0.approval_options.0.execution_order", "beforeGates"),
				),
			},
		},
	})
}

// TestAccReleaseDefinition_withEnvironmentOptions tests environment_options and execution_policy blocks.
func TestAccReleaseDefinition_withEnvironmentOptions(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionWithEnvironmentOptions(name),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "environment.0.environment_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.environment_options.0.publish_deployment_status", "true"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.environment_options.0.email_notification_type", "OnlyOnFailure"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.execution_policy.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.execution_policy.0.concurrency_count", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.execution_policy.0.queue_depth_count", "0"),
				),
			},
		},
	})
}

// TestAccReleaseDefinition_update tests that a release definition can be updated (triggers revision
// increment and verifies no revision conflicts).
func TestAccReleaseDefinition_update(t *testing.T) {
	name := testutils.GenerateResourceName()
	updatedName := name + "-updated"
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionBasic(name),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "name", name),
				),
			},
			{
				Config: hclReleaseDefinitionBasic(updatedName),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(updatedName),
					resource.TestCheckResourceAttr(tfNode, "name", updatedName),
					resource.TestCheckResourceAttrSet(tfNode, "revision"),
				),
			},
		},
	})
}

// TestAccReleaseDefinition_complete is the exhaustive acceptance test that exercises
// EVERY non-default option of betterado_release_definition in a single live ADO round-trip:
//   - real agent queue resolved from the test project (not 0)
//   - agent_specification set to "ubuntu-22.04" (non-default; AC2/WI-9)
//   - demands on the agent deploy phase
//   - skip_artifacts_download = true, enable_access_token = true
//   - non-default retention policy
//   - pre/post approvals with approval_options
//   - pre/post deployment gates with a real "Query Work Items" gate referencing a real shared query
//     (AC3/WI-9: a betterado_workitemquery resource creates the query so queryId is non-empty)
//   - cd_artifact_trigger + schedule_trigger with NO branch_filter (AC1/WI-9)
//   - a multiConfiguration parallel_execution deploy phase
//   - a runOnServer agentless deploy phase with a Delay workflow task
//   - definition-level and env-level variables
//   - idempotency (no perpetual diff) verified by ExpectNonEmptyPlan:false (default)
func TestAccReleaseDefinition_complete(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionComplete(name),
				Check: resource.ComposeTestCheckFunc(
					// basic existence + name
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttr(tfNode, "description", "Exhaustive acceptance test release definition"),
					resource.TestCheckResourceAttr(tfNode, "release_name_format", "Release-$(rev:r)"),

					// definition-level variable
					resource.TestCheckResourceAttr(tfNode, "variable.#", "1"),

					// artifact block set
					resource.TestCheckResourceAttr(tfNode, "artifact.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "artifact.0.alias", "_build"),
					resource.TestCheckResourceAttr(tfNode, "artifact.0.type", "Build"),
					resource.TestCheckResourceAttr(tfNode, "artifact.0.is_primary", "true"),

					// triggers block
					resource.TestCheckResourceAttr(tfNode, "triggers.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.cd_artifact_trigger.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.cd_artifact_trigger.0.artifact_alias", "_build"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.schedule_trigger.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.schedule_trigger.0.schedule_only_with_changes", "true"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.schedule_trigger.0.start_hours", "2"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.schedule_trigger.0.time_zone_id", "UTC"),

					// two environments
					resource.TestCheckResourceAttr(tfNode, "environment.#", "2"),

					// --- environment 0: Staging (full feature coverage) ---
					resource.TestCheckResourceAttr(tfNode, "environment.0.name", "Staging"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.rank", "1"),

					// env-level variable
					resource.TestCheckResourceAttr(tfNode, "environment.0.variable.#", "1"),

					// real queue_id set (not 0), agent_specification, demands, skip_artifacts_download, enable_access_token
					// AC2/WI-9: agent_specification persists as "ubuntu-22.04"
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.#", "2"),
					resource.TestCheckResourceAttrSet(tfNode, "environment.0.deploy_phase.0.deployment_input.0.queue_id"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.0.agent_specification", "ubuntu-22.04"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.0.demands.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.0.demands.0", "Agent.Version -gtVersion 2.0"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.0.skip_artifacts_download", "true"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.0.enable_access_token", "true"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.0.timeout_in_minutes", "60"),

					// multiConfiguration parallel_execution
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.0.parallel_execution.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.0.parallel_execution.0.type", "multiConfiguration"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.0.parallel_execution.0.max_number_of_agents", "2"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.0.parallel_execution.0.multipliers.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.0.deployment_input.0.parallel_execution.0.multipliers.0", "Configuration"),

					// runOnServer agentless phase (phase 1)
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.1.phase_type", "runOnServer"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.1.workflow_task.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.deploy_phase.1.workflow_task.0.task_id", "28782b92-5e8e-4458-9751-a71cd1492bae"),

					// environment_options non-default
					resource.TestCheckResourceAttr(tfNode, "environment.0.environment_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.environment_options.0.email_notification_type", "Always"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.environment_options.0.publish_deployment_status", "true"),

					// retention_policy non-default
					resource.TestCheckResourceAttr(tfNode, "environment.0.retention_policy.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.retention_policy.0.days_to_keep", "14"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.retention_policy.0.releases_to_keep", "5"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.retention_policy.0.retain_build", "false"),

					// pre_deploy_approval with approval_options
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.0.approver.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.0.approval_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.0.approval_options.0.release_creator_can_be_approver", "true"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.0.approval_options.0.timeout_in_minutes", "720"),

					// post_deploy_approval
					resource.TestCheckResourceAttr(tfNode, "environment.0.post_deploy_approval.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.post_deploy_approval.0.approver.#", "1"),

					// pre_deployment_gates with gate task
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deployment_gates.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deployment_gates.0.gates_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deployment_gates.0.gates_options.0.is_enabled", "true"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deployment_gates.0.gate.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deployment_gates.0.gate.0.task.0.task_id", "f1e4b0e6-017e-4819-8a48-ef19ae96e289"),

					// post_deployment_gates with gate task
					resource.TestCheckResourceAttr(tfNode, "environment.0.post_deployment_gates.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.post_deployment_gates.0.gates_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.post_deployment_gates.0.gates_options.0.is_enabled", "true"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.post_deployment_gates.0.gate.#", "1"),

					// --- environment 1: Production (condition on environment state) ---
					resource.TestCheckResourceAttr(tfNode, "environment.1.name", "Production"),
					resource.TestCheckResourceAttr(tfNode, "environment.1.rank", "2"),
					resource.TestCheckResourceAttr(tfNode, "environment.1.condition.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.1.condition.0.condition_type", "environmentState"),

					// checkReleaseDefinitionHasGates verifies gates#>0 via the API
					checkReleaseDefinitionHasGates(),
					// checkReleaseDefinitionQueueSet verifies queue_id != 0 via the API
					checkReleaseDefinitionQueueSet(),
					// checkReleaseDefinitionAgentSpecification verifies agentSpecification.identifier persisted (AC2/WI-9)
					checkReleaseDefinitionAgentSpecification("ubuntu-22.04"),
				),
			},
			// second step verifies idempotency (ExpectNonEmptyPlan defaults to false)
			{
				Config:   hclReleaseDefinitionComplete(name),
				PlanOnly: true,
			},
		},
	})
}

// checkReleaseDefinitionHasGates verifies that the persisted release definition
// has at least one gate in the pre-deployment gates of the first environment.
func checkReleaseDefinitionHasGates() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		def, err := getReleaseDefinitionFromState(s)
		if err != nil {
			return err
		}
		if def.Environments == nil || len(*def.Environments) == 0 {
			return fmt.Errorf("expected at least one environment")
		}
		env := (*def.Environments)[0]
		if env.PreDeploymentGates == nil {
			return fmt.Errorf("expected pre_deployment_gates on first environment to be set")
		}
		if env.PreDeploymentGates.Gates == nil || len(*env.PreDeploymentGates.Gates) == 0 {
			return fmt.Errorf("expected at least one gate in pre_deployment_gates; got 0")
		}
		return nil
	}
}

// checkReleaseDefinitionQueueSet verifies that the first agent-based deploy phase
// has a non-zero queueId in the API response.
func checkReleaseDefinitionQueueSet() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		def, err := getReleaseDefinitionFromState(s)
		if err != nil {
			return err
		}
		if def.Environments == nil || len(*def.Environments) == 0 {
			return fmt.Errorf("expected at least one environment")
		}
		env := (*def.Environments)[0]
		if env.DeployPhases == nil || len(*env.DeployPhases) == 0 {
			return fmt.Errorf("expected at least one deploy phase")
		}
		// DeployPhases are raw interface{} from JSON; extract queueId from first phase's deploymentInput
		firstPhase, ok := (*env.DeployPhases)[0].(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected deploy phase type")
		}
		di, ok := firstPhase["deploymentInput"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("deploymentInput not found in first phase")
		}
		queueID, ok := di["queueId"]
		if !ok {
			return fmt.Errorf("queueId not set in deploymentInput; expected a non-zero queue")
		}
		switch v := queueID.(type) {
		case float64:
			if v == 0 {
				return fmt.Errorf("queueId is 0; expected a real queue ID")
			}
		case int:
			if v == 0 {
				return fmt.Errorf("queueId is 0; expected a real queue ID")
			}
		}
		return nil
	}
}

// checkReleaseDefinitionAgentSpecification verifies that agentSpecification.identifier
// persisted in the ADO API response for the first deploy phase of the first environment.
// This is the API-level verification for AC2/WI-9.
func checkReleaseDefinitionAgentSpecification(expectedIdentifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		def, err := getReleaseDefinitionFromState(s)
		if err != nil {
			return err
		}
		if def.Environments == nil || len(*def.Environments) == 0 {
			return fmt.Errorf("expected at least one environment")
		}
		env := (*def.Environments)[0]
		if env.DeployPhases == nil || len(*env.DeployPhases) == 0 {
			return fmt.Errorf("expected at least one deploy phase")
		}
		firstPhase, ok := (*env.DeployPhases)[0].(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected deploy phase type")
		}
		di, ok := firstPhase["deploymentInput"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("deploymentInput not found in first phase")
		}
		agentSpec, ok := di["agentSpecification"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("agentSpecification not found in deploymentInput; expected identifier %q", expectedIdentifier)
		}
		identifier, ok := agentSpec["identifier"].(string)
		if !ok || identifier == "" {
			return fmt.Errorf("agentSpecification.identifier is empty; expected %q", expectedIdentifier)
		}
		if identifier != expectedIdentifier {
			return fmt.Errorf("agentSpecification.identifier = %q; want %q", identifier, expectedIdentifier)
		}
		return nil
	}
}

// --- Check functions ---

func checkReleaseDefinitionExists(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		def, err := getReleaseDefinitionFromState(s)
		if err != nil {
			return err
		}
		if def.Name == nil || *def.Name != expectedName {
			return fmt.Errorf("expected release definition name %q but got %v", expectedName, def.Name)
		}
		return nil
	}
}

func checkReleaseDefinitionDestroyed(s *terraform.State) error {
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "betterado_release_definition" {
			continue
		}
		if _, err := getReleaseDefinitionFromResource(resource); err == nil {
			return fmt.Errorf("unexpectedly found release definition that should be deleted")
		}
	}
	return nil
}

func getReleaseDefinitionFromState(s *terraform.State) (*release.ReleaseDefinition, error) {
	for _, res := range s.RootModule().Resources {
		if res.Type == "betterado_release_definition" {
			return getReleaseDefinitionFromResource(res)
		}
	}
	return nil, fmt.Errorf("no betterado_release_definition found in state")
}

func getReleaseDefinitionFromResource(res *terraform.ResourceState) (*release.ReleaseDefinition, error) {
	defID, err := strconv.Atoi(res.Primary.ID)
	if err != nil {
		return nil, err
	}
	projectID := res.Primary.Attributes["project_id"]
	clients := testutils.GetProvider().Meta().(*client.AggregatedClient)
	return clients.ReleaseClient.GetReleaseDefinition(clients.Ctx, release.GetReleaseDefinitionArgs{
		Project:      &projectID,
		DefinitionId: &defID,
	})
}

// --- HCL templates ---

func hclReleaseDefinitionProjectBase(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  description        = "%[1]s-description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}
`, name)
}

// hclReleaseDefinitionBasicFixture creates a minimal release definition with one environment,
// referencing a fixture-supplied project ID directly (no inline betterado_project block).
// The fixture (from SharedReleaseFixture / WI-1) owns and cleans up the project, repo,
// and build definition — this HCL only creates the betterado_release_definition itself.
//
// AC2: project_id is the fixture-supplied UUID; no betterado_project resource is emitted here.
// AC3: retention_policy + pre/post approvals satisfy VS402982 / VS402877.
func hclReleaseDefinitionBasicFixture(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  environment {
    name = "Production"
    rank = 1

    deploy_phase {
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }

    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }

    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }

    # ADO requires BOTH pre- and post-deploy approvals (VS402877).
    post_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }
  }
}
`, name, fixture.ProjectID)
}

// hclReleaseDefinitionBasic creates a minimal release definition with one environment.
// Uses the default agent queue (queue_id omitted from deployment_input since it is optional).
// retention_policy and pre_deploy_approval are included because ADO REST 7.2 requires them
// (VS402982 / VS402877). The schema fields remain Optional — only the test HCL is updated.
func hclReleaseDefinitionBasic(name string) string {
	base := hclReleaseDefinitionProjectBase(name)
	return fmt.Sprintf(`
%s

resource "betterado_release_definition" "test" {
  name       = "%[2]s"
  project_id = betterado_project.test.id

  environment {
    name = "Production"
    rank = 1

    deploy_phase {
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }

    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }

    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }

    # ADO requires BOTH pre- and post-deploy approvals to be non-empty on a
    # stage, else create fails with VS402877 ("Pre-approvals or post-approvals
    # … are empty"). The exhaustive _complete test always carried both; the
    # minimal fixtures carried only pre and were never live-run (only _complete
    # was INIT-1's merge gate), so the gap surfaced when every test was finally
    # run with TF_ACC.
    post_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }
  }
}
`, base, name)
}

// hclReleaseDefinitionWithDeploymentInput uses data source to get a queue ID for deployment_input.
// Falls back to the "Azure Pipelines" hosted pool (queue 4 is the default in most orgs).
// retention_policy and pre_deploy_approval are included to satisfy ADO REST 7.2 (VS402982/VS402877).
func hclReleaseDefinitionWithDeploymentInput(name string) string {
	base := hclReleaseDefinitionProjectBase(name)
	return fmt.Sprintf(`
%s

data "betterado_agent_queue" "test" {
  name       = "Azure Pipelines"
  project_id = betterado_project.test.id
}

resource "betterado_release_definition" "test" {
  name       = "%[2]s"
  project_id = betterado_project.test.id

  environment {
    name = "Production"
    rank = 1

    deploy_phase {
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"

      deployment_input {
        queue_id                      = data.betterado_agent_queue.test.id
        timeout_in_minutes            = 0
        job_cancel_timeout_in_minutes = 1
        condition                     = "succeeded()"
      }
    }

    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }

    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }

    # ADO requires BOTH pre- and post-deploy approvals to be non-empty on a
    # stage, else create fails with VS402877 ("Pre-approvals or post-approvals
    # … are empty"). The exhaustive _complete test always carried both; the
    # minimal fixtures carried only pre and were never live-run (only _complete
    # was INIT-1's merge gate), so the gap surfaced when every test was finally
    # run with TF_ACC.
    post_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }
  }
}
`, base, name)
}

// hclReleaseDefinitionWithApprovalOptions creates a release definition with approval_options.
func hclReleaseDefinitionWithApprovalOptions(name string) string {
	base := hclReleaseDefinitionProjectBase(name)
	return fmt.Sprintf(`
%s

resource "betterado_release_definition" "test" {
  name       = "%[2]s"
  project_id = betterado_project.test.id

  environment {
    name = "Production"
    rank = 1

    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }

      approval_options {
        release_creator_can_be_approver                                 = false
        enforce_identity_revalidation                                   = false
        timeout_in_minutes                                              = 1440
        execution_order                                                 = "beforeGates"
        auto_triggered_and_previous_environment_approved_can_be_skipped = false
      }
    }

    # ADO requires BOTH pre- and post-deploy approvals to be non-empty on a
    # stage (VS402877).
    post_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }

    deploy_phase {
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }

    # ADO REST 7.2 requires a stage-level retention policy (VS402982).
    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }
  }
}
`, base, name)
}

// hclReleaseDefinitionWithEnvironmentOptions creates a release definition with environment_options
// and execution_policy blocks.
// retention_policy and pre_deploy_approval are included to satisfy ADO REST 7.2 (VS402982/VS402877).
func hclReleaseDefinitionWithEnvironmentOptions(name string) string {
	base := hclReleaseDefinitionProjectBase(name)
	return fmt.Sprintf(`
%s

resource "betterado_release_definition" "test" {
  name       = "%[2]s"
  project_id = betterado_project.test.id

  environment {
    name = "Production"
    rank = 1

    deploy_phase {
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }

    environment_options {
      email_notification_type   = "OnlyOnFailure"
      publish_deployment_status = true
      badge_enabled             = false
      auto_link_work_items      = false
    }

    execution_policy {
      concurrency_count = 1
      queue_depth_count = 0
    }

    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }

    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }

    # ADO requires BOTH pre- and post-deploy approvals to be non-empty on a
    # stage, else create fails with VS402877 ("Pre-approvals or post-approvals
    # … are empty"). The exhaustive _complete test always carried both; the
    # minimal fixtures carried only pre and were never live-run (only _complete
    # was INIT-1's merge gate), so the gap surfaced when every test was finally
    # run with TF_ACC.
    post_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }
  }
}
`, base, name)
}

// hclReleaseDefinitionComplete creates an exhaustive release definition that sets a NON-DEFAULT
// value for EVERY option of betterado_release_definition. This is the standing acceptance test
// merge gate as described in WI-8.
//
// Features exercised:
//   - real agent queue (resolved via data source, not 0)
//   - agent_specification set to "ubuntu-22.04" (non-default; AC2/WI-9)
//   - demands on the agent phase
//   - skip_artifacts_download = true, enable_access_token = true
//   - multiConfiguration parallel_execution with multipliers
//   - runOnServer agentless phase with a Delay workflow task
//   - pre/post approvals with full approval_options
//   - pre/post deployment gates with a real ServerGate task (queryWorkItems) referencing a real
//     shared query created by betterado_workitemquery (AC3/WI-9: non-empty queryId)
//   - cd_artifact_trigger + schedule_trigger with NO branch_filter (AC1/WI-9)
//   - non-default retention_policy
//   - non-default environment_options and execution_policy
//   - definition-level + environment-level variables
//   - Build artifact referencing the git repo + build definition created in this test project
//   - condition { name = "ReleaseStarted", condition_type = "event" } on first env (required by ADO
//     when a schedule_trigger schedules release creation)
func hclReleaseDefinitionComplete(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = %[1]q
  description        = "%[1]s-description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

# Create a Git repository so we have a real repo for the build definition artifact.
resource "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s-repo"
  initialization {
    init_type = "Clean"
  }
}

# Build definition — provides the artifact source for the cd_artifact_trigger.
resource "betterado_build_definition" "test" {
  project_id      = betterado_project.test.id
  name            = "%[1]s-build"
  agent_pool_name = "Azure Pipelines"

  repository {
    repo_type   = "TfsGit"
    repo_id     = betterado_git_repository.test.id
    branch_name = betterado_git_repository.test.default_branch
    yml_path    = "azure-pipelines.yml"
  }
}

# Resolve the Azure Pipelines hosted queue for this test project.
data "betterado_agent_queue" "test" {
  name       = "Azure Pipelines"
  project_id = betterado_project.test.id
}

# Create a shared work-item query so the "Query Work Items" gate task has a real queryId.
# AC3/WI-9: betterado_workitemquery creates the query under "Shared Queries".
resource "betterado_workitemquery" "gate_query" {
  project_id = betterado_project.test.id
  name       = "All Work Items - Gate Check"
  area       = "Shared Queries"
  wiql       = "SELECT [System.Id] FROM WorkItems WHERE [System.TeamProject] = @project ORDER BY [System.Id]"
}

resource "betterado_release_definition" "test" {
  name                = %[1]q
  project_id          = betterado_project.test.id
  description         = "Exhaustive acceptance test release definition"
  release_name_format = "Release-$(rev:r)"

  # definition-level variable (non-default)
  variable {
    name  = "MY_VAR"
    value = "hello-complete"
  }

  # Build artifact — the definition_reference ties to the build definition above.
  artifact {
    alias      = "_build"
    type       = "Build"
    is_primary = true

    definition_reference = {
      definition = tostring(betterado_build_definition.test.id)
      project    = betterado_project.test.id
    }
  }

  # cd_artifact_trigger fires on every build of _build; schedule_trigger at 02:00 UTC daily.
  # IMPORTANT: environment 0 must have condition { name = "ReleaseStarted", condition_type = "event" }
  # so that ADO allows the schedule trigger to create releases (ADO rejects otherwise).
  triggers {
    cd_artifact_trigger {
      artifact_alias = "_build"
      branch_filter {
        include = ["refs/heads/main"]
        exclude = []
      }
    }

    schedule_trigger {
      schedule_only_with_changes = true
      start_hours                = 2
      start_minutes              = 0
      time_zone_id               = "UTC"
      days_to_release            = 62 # Mon–Fri (bits 1+2+4+8+16+32 = 62)
      # Note: branch_filter is not included because ADO does not return branchFilters
      # for schedule triggers in the GET response, which would cause a perpetual diff.
    }
  }

  # ─── Environment 0: Staging — exercises all per-environment options ───────────
  environment {
    name = "Staging"
    rank = 1

    # Required by ADO when schedule_trigger is used (schedules release creation event)
    condition {
      name           = "ReleaseStarted"
      condition_type = "event"
      value          = ""
    }

    # environment-level variable (non-default)
    variable {
      name  = "ENV_VAR"
      value = "staging-complete"
    }

    # --- Agent-based phase: real queue + demands + skip_artifacts_download + enable_access_token
    #     + multiConfiguration parallel execution ---
    deploy_phase {
      name       = "Agent phase"
      rank       = 1
      phase_type = "agentBasedDeployment"

      deployment_input {
        queue_id                      = data.betterado_agent_queue.test.id
        agent_specification           = "ubuntu-22.04"
        timeout_in_minutes            = 60
        job_cancel_timeout_in_minutes = 5
        condition                     = "succeeded()"
        skip_artifacts_download       = true
        enable_access_token           = true

        demands = ["Agent.Version -gtVersion 2.0"]

        parallel_execution {
          type                 = "multiConfiguration"
          max_number_of_agents = 2
          multipliers          = ["Configuration"]
          continue_on_error    = false
        }
      }
    }

    # --- runOnServer agentless phase with a Delay workflow task (task GUID from /_apis/distributedtask/tasks) ---
    deploy_phase {
      name       = "Agentless phase"
      rank       = 2
      phase_type = "runOnServer"

      deployment_input {
        timeout_in_minutes            = 0
        job_cancel_timeout_in_minutes = 1
        condition                     = "succeeded()"
      }

      workflow_task {
        name    = "Delay"
        task_id = "28782b92-5e8e-4458-9751-a71cd1492bae"
        version = "1.*"
        enabled = true
        inputs = {
          delayForMinutes = "1"
        }
      }
    }

    # environment options (non-default)
    environment_options {
      email_notification_type   = "Always"
      publish_deployment_status = true
      badge_enabled             = false
      auto_link_work_items      = false
    }

    # execution policy (non-default)
    execution_policy {
      concurrency_count = 1
      queue_depth_count = 0
    }

    # retention policy (non-default)
    retention_policy {
      days_to_keep     = 14
      releases_to_keep = 5
      retain_build     = false
    }

    # pre_deploy_approval with non-default approval_options
    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }

      approval_options {
        release_creator_can_be_approver                                 = true
        enforce_identity_revalidation                                   = false
        timeout_in_minutes                                              = 720
        execution_order                                                 = "beforeGates"
        auto_triggered_and_previous_environment_approved_can_be_skipped = false
      }
    }

    # post_deploy_approval (non-default: normally automated)
    post_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }

    # pre_deployment_gates with a real ServerGate task: queryWorkItems (f1e4b0e6-017e-4819-8a48-ef19ae96e289)
    # AC3/WI-9: betterado_workitemquery.gate_query provides a real shared query ID so queryId is not empty.
    pre_deployment_gates {
      gates_options {
        is_enabled               = true
        timeout                  = 60
        sampling_interval        = 5
        stabilization_time       = 0
        minimum_success_duration = 0
      }

      gate {
        task {
          name    = "Query Work Items"
          task_id = "f1e4b0e6-017e-4819-8a48-ef19ae96e289"
          version = "0.*"
          enabled = true
          inputs = {
            queryId = betterado_workitemquery.gate_query.id
          }
        }
      }
    }

    # post_deployment_gates with the same ServerGate task
    post_deployment_gates {
      gates_options {
        is_enabled               = true
        timeout                  = 60
        sampling_interval        = 5
        stabilization_time       = 0
        minimum_success_duration = 0
      }

      gate {
        task {
          name    = "Query Work Items"
          task_id = "f1e4b0e6-017e-4819-8a48-ef19ae96e289"
          version = "0.*"
          enabled = true
          inputs = {
            queryId = betterado_workitemquery.gate_query.id
          }
        }
      }
    }
  }

  # ─── Environment 1: Production — environment-state condition ─────────────────
  environment {
    name = "Production"
    rank = 2

    condition {
      name           = "Staging"
      condition_type = "environmentState"
      value          = "4"
    }

    deploy_phase {
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }

    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }

    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }

    post_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }
  }
}
`, name)
}

// TestAccReleaseDefinition_approvalsAndGates verifies that non-default approval_options
// (timeout_in_minutes, execution_order, release_creator_can_be_approver) and a
// pre_deployment_gates block (with a Query Work Items gate task) are persisted to ADO,
// read back without drift, and that a re-plan produces no diff (idempotency).
func TestAccReleaseDefinition_approvalsAndGates(t *testing.T) {
	fixture := SharedReleaseFixture(t)
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionApprovalsAndGates(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					// approval_options assertions
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.0.approval_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.0.approval_options.0.timeout_in_minutes", "1440"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.0.approval_options.0.execution_order", "beforeGates"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deploy_approval.0.approval_options.0.release_creator_can_be_approver", "false"),
					// gates assertions
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deployment_gates.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deployment_gates.0.gates_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deployment_gates.0.gates_options.0.is_enabled", "true"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deployment_gates.0.gate.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.pre_deployment_gates.0.gate.0.task.0.task_id", "f1e4b0e6-017e-4819-8a48-ef19ae96e289"),
				),
			},
			// idempotency step: re-plan must emit no diff
			{
				Config:   hclReleaseDefinitionApprovalsAndGates(name, fixture),
				PlanOnly: true,
			},
		},
	})
}

// TestAccReleaseDefinition_environmentConfig verifies that environment_trigger, schedule, and
// properties blocks are correctly persisted to ADO and read back without drift.
// A PlanOnly step with ExpectNonEmptyPlan:false confirms idempotency (AC1/WI-2).
func TestAccReleaseDefinition_environmentConfig(t *testing.T) {
	fixture := SharedReleaseFixture(t)
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionEnvironmentConfig(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					// environment_trigger
					resource.TestCheckResourceAttr(tfNode, "environment.0.environment_trigger.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.environment_trigger.0.trigger_type", "rollbackRedeploy"),
					// schedule
					resource.TestCheckResourceAttr(tfNode, "environment.0.schedule.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.schedule.0.start_hours", "3"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.schedule.0.time_zone_id", "UTC"),
					// properties
					resource.TestCheckResourceAttr(tfNode, "environment.0.properties.%", "1"),
					resource.TestCheckResourceAttr(tfNode, "environment.0.properties.env", "staging"),
				),
			},
			{
				// Idempotency: no perpetual diff (AC1/WI-2)
				Config:             hclReleaseDefinitionEnvironmentConfig(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclReleaseDefinitionEnvironmentConfig returns HCL for a betterado_release_definition that
// exercises environment_trigger, schedule, and properties blocks (WI-2).
// Uses fixture.ProjectID (no inline betterado_project) to keep the test focused on the
// new environment-level blocks rather than project lifecycle.
func hclReleaseDefinitionEnvironmentConfig(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  environment {
    name = "Staging"
    rank = 1

    environment_trigger {
      trigger_type = "rollbackRedeploy"
    }

    schedule {
      days_to_release = 62
      start_hours     = 3
      start_minutes   = 0
      time_zone_id    = "UTC"
    }

    properties = {
      env = "staging"
    }

    deploy_phase {
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }

    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }

    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }

    # ADO requires BOTH pre- and post-deploy approvals (VS402877).
    post_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }
  }
}
`, name, fixture.ProjectID)
}

// hclReleaseDefinitionApprovalsAndGates returns HCL for a betterado_release_definition with
// non-default approval_options and a pre_deployment_gates block containing a Query Work Items
// gate task. Uses fixture.ProjectID (no inline betterado_project) and fixture.WorkItemQueryID
// so the gate references a real shared query in the project.
// A runOnServer phase is used to avoid the need for a real agent queue in this focused test.
func hclReleaseDefinitionApprovalsAndGates(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  environment {
    name = "Staging"
    rank = 1

    deploy_phase {
      name       = "Agentless job"
      rank       = 1
      phase_type = "runOnServer"

      deployment_input {
        timeout_in_minutes            = 0
        job_cancel_timeout_in_minutes = 1
        condition                     = "succeeded()"
      }
    }

    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }

    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }

      approval_options {
        release_creator_can_be_approver                                 = false
        enforce_identity_revalidation                                   = false
        timeout_in_minutes                                              = 1440
        execution_order                                                 = "beforeGates"
        auto_triggered_and_previous_environment_approved_can_be_skipped = false
      }
    }

    # ADO requires BOTH pre- and post-deploy approvals (VS402877).
    post_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }

    pre_deployment_gates {
      gates_options {
        is_enabled               = true
        timeout                  = 600
        sampling_interval        = 60
        stabilization_time       = 0
        minimum_success_duration = 0
      }

      gate {
        task {
          name    = "Query Work Items"
          task_id = "f1e4b0e6-017e-4819-8a48-ef19ae96e289"
          version = "0.*"
          enabled = true
          inputs = {
            queryId = %[3]q
          }
        }
      }
    }
  }
}
`, name, fixture.ProjectID, fixture.WorkItemQueryID)
}

// TestAccReleaseDefinition_triggerEnhancements verifies the trigger-enhancement fields
// introduced in the INIT-2026-06-08 initiative:
//   - cd_artifact_trigger.tag_filter (pattern + tags)
//   - cd_artifact_trigger.use_build_definition_branch
//   - cd_artifact_trigger.create_release_on_build_tagging
//   - source_repo_trigger (alias + branch_filters)
//
// Uses SharedReleaseFixture for the project + build definition so no inline
// betterado_project is needed.  A runOnServer phase avoids requiring a real
// agent queue.  Pre/post approvals and retention_policy satisfy VS402877/VS402982.
func TestAccReleaseDefinition_triggerEnhancements(t *testing.T) {
	fixture := SharedReleaseFixture(t)
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionTriggerEnhancements(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					// Triggers block: 1 entry containing both trigger types.
					resource.TestCheckResourceAttr(tfNode, "triggers.#", "1"),
					// CD artifact trigger assertions.
					resource.TestCheckResourceAttr(tfNode, "triggers.0.cd_artifact_trigger.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.cd_artifact_trigger.0.artifact_alias", "_build"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.cd_artifact_trigger.0.tag_filter.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.cd_artifact_trigger.0.tag_filter.0.pattern", "v*"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.cd_artifact_trigger.0.use_build_definition_branch", "true"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.cd_artifact_trigger.0.create_release_on_build_tagging", "true"),
					// Source repo trigger assertions.
					resource.TestCheckResourceAttr(tfNode, "triggers.0.source_repo_trigger.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.source_repo_trigger.0.alias", "_build"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.source_repo_trigger.0.branch_filters.#", "1"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclReleaseDefinitionTriggerEnhancements(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclReleaseDefinitionTriggerEnhancements returns HCL for a betterado_release_definition
// that exercises cd_artifact_trigger (tag_filter, use_build_definition_branch,
// create_release_on_build_tagging) and source_repo_trigger.
// Arg layout: %[1]q = name, %[2]q = fixture.ProjectID, %[3]d = fixture.BuildDefinitionID.
func hclReleaseDefinitionTriggerEnhancements(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  # Build artifact — links to the fixture's pre-created build definition.
  artifact {
    alias      = "_build"
    type       = "Build"
    is_primary = true

    definition_reference = {
      definition = tostring(%[3]d)
      project    = %[2]q
    }
  }

  # Triggers block exercises all trigger-enhancement fields.
  triggers {
    # CD artifact trigger with tag_filter and new boolean flags.
    cd_artifact_trigger {
      artifact_alias                  = "_build"
      use_build_definition_branch     = true
      create_release_on_build_tagging = true

      tag_filter {
        pattern = "v*"
        tags    = ["stable"]
      }
    }

    # Source-repo trigger on the same artifact alias.
    source_repo_trigger {
      alias          = "_build"
      branch_filters = ["refs/heads/main"]
    }
  }

  environment {
    name = "Production"
    rank = 1

    # runOnServer phase — no real agent queue needed for this focused trigger test.
    deploy_phase {
      name       = "Agentless job"
      rank       = 1
      phase_type = "runOnServer"
    }

    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }

    # ADO VS402877: both pre- and post-deploy approvals are mandatory.
    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }

    post_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }
  }
}
`, name, fixture.ProjectID, fixture.BuildDefinitionID)
}
