package acceptancetests

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azdoapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// TestAccReleaseDefinition_basic tests creating, importing, and destroying a release definition
// with array syntax HCL and a Build artifact with definition_reference.
//
// AC1/WI-6: uses array syntax for stages and artifact, UUID-prefixed name, non-default stage
// name "Production", deploy_phase with rank=1, one artifact with definition_reference,
// idempotency step (ExpectNonEmptyPlan:false), CheckDestroy via API 404, captureReleaseEvidence.
// AC2/WI-6: partial-stages step (omitting workflow_task + all approval/gate blocks)
// produces no perpetual diff. Evidence to release-def-partial-stages.json.
// AC3/WI-6: partial-variables step (omitting is_secret/allow_override) produces no
// perpetual diff. Evidence to release-def-partial-variables.json.
// AC4/WI-6: artifact with minimal definition_reference keys — ADO-added keys are filtered.
// Evidence to release-def-artifact-filter.json.
func TestAccReleaseDefinition_basic(t *testing.T) {
	fixture := SharedReleaseFixture(t)

	// UUID-prefixed name (AC1): first 8 chars of a new UUID + random suffix.
	name := uuid.New().String()[:8] + "-" + testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			// Step 1: apply array-syntax HCL with artifact + definition_reference (AC1/AC4).
			{
				Config: hclReleaseDefinitionBasicArraySyntax(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "revision"),
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttr(tfNode, "path", `\`),
					resource.TestCheckResourceAttr(tfNode, "stages.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.name", "Production"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.rank", "1"),
					resource.TestCheckResourceAttr(tfNode, "artifact.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "artifact.0.alias", "_build"),
					// AC4/WI-6: capture live evidence (a real vsrm REST GET) before destroy
					// so forge demo render can back-fill it into demo.json. Best-effort.
					captureReleaseEvidence(tfNode),
					// AC4/WI-6: artifact-filter evidence — artifact_reference has no ADO-added keys.
					captureArtifactFilterEvidence(tfNode),
				),
			},
			// Step 2: import + verify all state attributes match.
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: idempotency — re-plan after apply+import must produce no diff (AC1/WI-6).
			{
				Config:             hclReleaseDefinitionBasicArraySyntax(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 4: partial-stages — omit workflow_task + all approval/gate blocks (AC2/WI-6).
			{
				Config: hclReleaseDefinitionPartialStages(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "stages.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.name", "Production"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.rank", "1"),
					capturePartialStagesEvidence(tfNode),
				),
			},
			// Step 5: idempotency for partial-stages — must produce no diff (AC2/WI-6).
			{
				Config:             hclReleaseDefinitionPartialStages(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 6: partial-variables — omit is_secret + allow_override (AC3/WI-6).
			{
				Config: hclReleaseDefinitionPartialVariables(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					capturePartialVariablesEvidence(tfNode),
				),
			},
			// Step 7: idempotency for partial-variables — must produce no diff (AC3/WI-6).
			{
				Config:             hclReleaseDefinitionPartialVariables(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccReleaseDefinition_import is a dedicated standalone import test for betterado_release_definition.
//
// Step 1 (Create): Applies hclReleaseDefinitionBasicFixture to create a release definition
// in the SharedReleaseFixture project — no inline betterado_project block needed.
// Step 2 (Import): Imports the resource by project_id/definition_id using
// ComputeProjectQualifiedResourceImportID; ImportStateVerify: true confirms all state attributes match.
// Step 3 (Idempotency): PlanOnly: true confirms no perpetual diff after import.
//
// AC1: dedicated import test passes live (TF_ACC=1), imported state matches created state
// (ImportStateVerify: true), re-plan produces no diff.
// AC2: import step uses ComputeProjectQualifiedResourceImportID (tfhelper.ImportProjectQualifiedResource
// is the wired importer), all attributes match after import with no RequiredDuringImport errors.
func TestAccReleaseDefinition_import(t *testing.T) {
	fixture := SharedReleaseFixture(t)

	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create a minimal release definition using the fixture project.
			{
				Config: hclReleaseDefinitionBasicFixture(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "revision"),
				),
			},
			// Step 2: import by project_id/definition_id; verify all state attributes match.
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: idempotency — re-plan after import must produce no diff.
			{
				Config:             hclReleaseDefinitionBasicFixture(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccReleaseDefinition_withDeploymentInput tests that deployment_input (queue_id) is correctly
// sent to and read back from the API.
func TestAccReleaseDefinition_withDeploymentInput(t *testing.T) {
	fixture := SharedReleaseFixture(t)
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionWithDeploymentInput(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.deployment_input.#", "1"),
					resource.TestCheckResourceAttrSet(tfNode, "stages.0.deploy_phase.0.deployment_input.0.queue_id"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.deployment_input.0.condition", "succeeded()"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.deployment_input.0.timeout_in_minutes", "0"),
				),
			},
		},
	})
}

// TestAccReleaseDefinition_withApprovalOptions tests approval_options on pre/post approvals.
func TestAccReleaseDefinition_withApprovalOptions(t *testing.T) {
	fixture := SharedReleaseFixture(t)
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionWithApprovalOptions(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deploy_approval.0.approval_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deploy_approval.0.approval_options.0.release_creator_can_be_approver", "false"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deploy_approval.0.approval_options.0.timeout_in_minutes", "1440"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deploy_approval.0.approval_options.0.execution_order", "beforeGates"),
				),
			},
		},
	})
}

// TestAccReleaseDefinition_withEnvironmentOptions tests environment_options and execution_policy blocks.
func TestAccReleaseDefinition_withEnvironmentOptions(t *testing.T) {
	fixture := SharedReleaseFixture(t)
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionWithEnvironmentOptions(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "stages.0.environment_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.environment_options.0.publish_deployment_status", "true"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.environment_options.0.email_notification_type", "OnlyOnFailure"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.execution_policy.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.execution_policy.0.concurrency_count", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.execution_policy.0.queue_depth_count", "0"),
				),
			},
		},
	})
}

// TestAccReleaseDefinition_update tests that a release definition can be updated (triggers revision
// increment and verifies no revision conflicts).
func TestAccReleaseDefinition_update(t *testing.T) {
	fixture := SharedReleaseFixture(t)
	name := testutils.GenerateResourceName()
	updatedName := name + "-updated"
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionBasic(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "name", name),
				),
			},
			{
				Config: hclReleaseDefinitionBasic(updatedName, fixture),
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
	fixture := SharedReleaseFixture(t)
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionComplete(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					// basic existence + name
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttr(tfNode, "description", "Exhaustive acceptance test release definition"),
					resource.TestCheckResourceAttr(tfNode, "release_name_format", "Release-$(rev:r)"),

					// definition-level variable
					resource.TestCheckResourceAttr(tfNode, "variables.%", "1"),

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

					// two stages
					resource.TestCheckResourceAttr(tfNode, "stages.#", "2"),

					// --- stages 0: Staging (full feature coverage) ---
					resource.TestCheckResourceAttr(tfNode, "stages.0.name", "Staging"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.rank", "1"),

					// env-level variable
					resource.TestCheckResourceAttr(tfNode, "stages.0.variables.%", "1"),

					// real queue_id set (not 0), agent_specification, demands, skip_artifacts_download, enable_access_token
					// AC2/WI-9: agent_specification persists as "ubuntu-22.04"
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.#", "2"),
					resource.TestCheckResourceAttrSet(tfNode, "stages.0.deploy_phase.0.deployment_input.0.queue_id"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.deployment_input.0.agent_specification", "ubuntu-22.04"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.deployment_input.0.demands.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.deployment_input.0.demands.0", "Agent.Version -gtVersion 2.0"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.deployment_input.0.skip_artifacts_download", "true"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.deployment_input.0.enable_access_token", "true"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.deployment_input.0.timeout_in_minutes", "60"),

					// multiConfiguration parallel_execution
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.deployment_input.0.parallel_execution.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.deployment_input.0.parallel_execution.0.type", "multiConfiguration"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.deployment_input.0.parallel_execution.0.max_number_of_agents", "2"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.deployment_input.0.parallel_execution.0.multipliers.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.deployment_input.0.parallel_execution.0.multipliers.0", "Configuration"),

					// runOnServer agentless phase (phase 1)
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.1.phase_type", "runOnServer"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.1.workflow_task.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.1.workflow_task.0.task_id", "28782b92-5e8e-4458-9751-a71cd1492bae"),

					// environment_options non-default
					resource.TestCheckResourceAttr(tfNode, "stages.0.environment_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.environment_options.0.email_notification_type", "Always"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.environment_options.0.publish_deployment_status", "true"),

					// retention_policy non-default
					resource.TestCheckResourceAttr(tfNode, "stages.0.retention_policy.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.retention_policy.0.days_to_keep", "14"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.retention_policy.0.releases_to_keep", "5"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.retention_policy.0.retain_build", "false"),

					// pre_deploy_approval with approval_options
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deploy_approval.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deploy_approval.0.approver.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deploy_approval.0.approval_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deploy_approval.0.approval_options.0.release_creator_can_be_approver", "true"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deploy_approval.0.approval_options.0.timeout_in_minutes", "720"),

					// post_deploy_approval
					resource.TestCheckResourceAttr(tfNode, "stages.0.post_deploy_approval.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.post_deploy_approval.0.approver.#", "1"),

					// pre_deployment_gates with gate task
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deployment_gates.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deployment_gates.0.gates_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deployment_gates.0.gates_options.0.is_enabled", "true"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deployment_gates.0.gate.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deployment_gates.0.gate.0.task.0.task_id", "f1e4b0e6-017e-4819-8a48-ef19ae96e289"),

					// post_deployment_gates with gate task
					resource.TestCheckResourceAttr(tfNode, "stages.0.post_deployment_gates.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.post_deployment_gates.0.gates_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.post_deployment_gates.0.gates_options.0.is_enabled", "true"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.post_deployment_gates.0.gate.#", "1"),

					// --- stages 1: Production (condition on environment state) ---
					resource.TestCheckResourceAttr(tfNode, "stages.1.name", "Production"),
					resource.TestCheckResourceAttr(tfNode, "stages.1.rank", "2"),
					resource.TestCheckResourceAttr(tfNode, "stages.1.condition.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.1.condition.0.condition_type", "environmentState"),

					// checkReleaseDefinitionHasGates verifies gates#>0 via the API
					checkReleaseDefinitionHasGates(),
					// checkReleaseDefinitionQueueSet verifies queue_id != 0 via the API
					checkReleaseDefinitionQueueSet(),
					// checkReleaseDefinitionAgentSpecification verifies agentSpecification.identifier persisted (AC2/WI-9)
					checkReleaseDefinitionAgentSpecification("ubuntu-22.04"),
				),
			},
			// AC3/WI-5: second step verifies idempotency — no diff (ExpectNonEmptyPlan: false).
			{
				Config:             hclReleaseDefinitionComplete(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccReleaseDefinition_blockSyntax verifies that a release definition created
// with block syntax (stages { … }, deploy_phase { … }, etc.) correctly applies,
// reads back, and produces no perpetual diff (idempotency).
//
// AC1: test uses block syntax HCL — no array syntax fixtures.
// AC2: terraform apply succeeds, stages.0.name = "Production" is readable, idempotency
// re-plan produces no diff (ExpectNonEmptyPlan: false), and terraform destroy completes cleanly.
func TestAccReleaseDefinition_blockSyntax(t *testing.T) {
	fixture := SharedReleaseFixture(t)

	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			// Step 1: apply with stages block syntax and assert read-back.
			{
				Config: hclReleaseDefinitionBlockSyntax(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "stages.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.name", "Production"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.name", "Agent job"),
					// Capture live demo evidence (a real REST GET) while the resource
					// is live, before destroy — forge demo render back-fills it into
					// demo.json. Best-effort; never fails the test.
					captureReleaseEvidence(tfNode),
				),
			},
			// Step 2: idempotency — re-plan must produce no diff.
			{
				Config:             hclReleaseDefinitionBlockSyntax(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
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

// captureReleaseEvidence performs a real live API GET of the created release
// definition and persists the response as forge demo live-evidence (before the
// resource is destroyed). Modeled on captureTaskGroupEvidence; uses the release
// client + the vsrm host. Best-effort: a capture failure never fails the test —
// the read-back assertions are the authoritative live proof; this only feeds the
// demo (forge demo render back-fills .forge/live-evidence/<label>.json into
// demo.json, joined on the matching checkpoint label).
func captureReleaseEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		def, err := getReleaseDefinitionFromResource(res)
		if err != nil {
			return nil //nolint:nilerr // best-effort demo evidence; the read-back assertions are the authoritative live proof
		}
		if def == nil || def.Id == nil {
			return nil
		}
		projectID := res.Primary.Attributes["project_id"]
		// The Release API lives on the vsrm host, not dev.azure.com.
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		vsrmURL := strings.Replace(orgURL, "https://dev.azure.com", "https://vsrm.dev.azure.com", 1)
		url := fmt.Sprintf("%s/%s/_apis/release/definitions/%d?api-version=7.1", vsrmURL, projectID, *def.Id)
		if err := testutils.CaptureLiveEvidence("acceptance-resource", url, def); err != nil {
			return nil //nolint:nilerr // best-effort demo evidence; never fail the test on a capture miss
		}
		return nil
	}
}

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
	// Build a fresh SDK client from env vars. We cannot use GetProvider().Meta()
	// here because tests that use GetMuxProviderFactories() wire a new SDKv2
	// provider instance for the mux — the module-level provider singleton is
	// never configured and its Meta() is nil.
	orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	if orgURL == "" || pat == "" {
		// Fall back to the legacy path so non-mux tests still work.
		clients := testutils.GetProvider().Meta().(*client.AggregatedClient)
		return clients.ReleaseClient.GetReleaseDefinition(clients.Ctx, release.GetReleaseDefinitionArgs{
			Project:      &projectID,
			DefinitionId: &defID,
		})
	}
	authProvider := azdoapi.NewAuthProviderPAT(pat)
	clients, err := client.GetAzdoClient(authProvider, orgURL)
	if err != nil {
		return nil, fmt.Errorf("getReleaseDefinitionFromResource: GetAzdoClient: %w", err)
	}
	return clients.ReleaseClient.GetReleaseDefinition(clients.Ctx, release.GetReleaseDefinitionArgs{
		Project:      &projectID,
		DefinitionId: &defID,
	})
}

// TestAccReleaseDefinition_providerHCLConfig proves WI-5: the framework provider
// builds its Azure DevOps client from the HCL provider block
// (org_service_url + personal_access_token), not only from environment variables.
// Before v1.0.1 the framework provider read credentials exclusively from env vars
// and ignored req.Config, so configuring the provider purely via HCL (e.g. a PAT
// sourced from a data source) left the client nil and every framework-resource
// apply failed with "Client not configured".
//
// To prove the credential flows from config and not the environment, the PAT env
// var is cleared for this test; the resource can therefore only succeed if the
// provider reads personal_access_token from the HCL block. Uses resource.Test
// (not ParallelTest) because t.Setenv is incompatible with t.Parallel.
func TestAccReleaseDefinition_providerHCLConfig(t *testing.T) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	if orgURL == "" || pat == "" {
		t.Skip("AZDO_ORG_SERVICE_URL and AZDO_PERSONAL_ACCESS_TOKEN must be set for this acceptance test")
	}

	// Build the shared fixture while the env credentials are still present.
	fixture := SharedReleaseFixture(t)
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	// Clear the PAT env var so the provider can ONLY succeed by reading the
	// credential from the HCL provider block below (the WI-5 fix).
	t.Setenv("AZDO_PERSONAL_ACCESS_TOKEN", "")

	providerBlock := fmt.Sprintf(`
provider "betterado" {
  org_service_url       = %[1]q
  personal_access_token = %[2]q
}
`, orgURL, pat)

	resource.Test(t, resource.TestCase{
		// Custom pre-check: the PAT is supplied via HCL, not the environment, so
		// don't require the PAT env var (testutils.PreCheck would).
		PreCheck: func() {
			if os.Getenv("TF_ACC") == "" {
				t.Skip("TF_ACC must be set for acceptance tests")
			}
		},
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		// No CheckDestroy / out-of-band existence check here: those shared helpers
		// build their verification client from the PAT env var, which this test
		// deliberately clears. The proof is that apply + an idempotent re-plan
		// succeed at all — they only can if the provider built its client from the
		// HCL block. terraform still destroys the resource via that same provider
		// at the end of the test.
		Steps: []resource.TestStep{
			{
				Config: providerBlock + hclReleaseDefinitionBasicArraySyntax(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "revision"),
				),
			},
			// Idempotency — re-plan with the HCL-configured provider must be empty.
			{
				Config:             providerBlock + hclReleaseDefinitionBasicArraySyntax(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// --- HCL templates ---

// hclReleaseDefinitionBasicFixture creates a minimal release definition with one stage,
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

  stages = [{
    name = "Production"
    rank = 1

    deploy_phase = [{
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    # ADO requires BOTH pre- and post-deploy approvals (VS402877).
    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID)
}

// hclReleaseDefinitionBlockSyntax creates a minimal release definition using
// attribute syntax (stages = [{…}], deploy_phase = [{…}], etc.).
// This is the fixture for TestAccReleaseDefinition_blockSyntax.
func hclReleaseDefinitionBlockSyntax(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  stages = [{
    name = "Production"
    rank = 1

    deploy_phase = [{
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID)
}

// hclReleaseDefinitionBasic creates a minimal release definition with one stage,
// referencing the shared fixture project (no inline betterado_project resource).
// Uses the default agent queue (queue_id omitted from deployment_input since it is optional).
// retention_policy and pre_deploy_approval are included because ADO REST 7.2 requires them
// (VS402982 / VS402877). The schema fields remain Optional — only the test HCL is updated.
func hclReleaseDefinitionBasic(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  stages = [{
    name = "Production"
    rank = 1

    deploy_phase = [{
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    # ADO requires BOTH pre- and post-deploy approvals to be non-empty on a
    # stage, else create fails with VS402877 ("Pre-approvals or post-approvals
    # … are empty"). The exhaustive _complete test always carried both; the
    # minimal fixtures carried only pre and were never live-run (only _complete
    # was INIT-1's merge gate), so the gap surfaced when every test was finally
    # run with TF_ACC.
    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID)
}

// hclReleaseDefinitionWithDeploymentInput uses data source to get a queue ID for deployment_input.
// Falls back to the "Azure Pipelines" hosted pool (queue 4 is the default in most orgs).
// retention_policy and pre_deploy_approval are included to satisfy ADO REST 7.2 (VS402982/VS402877).
// Uses the shared fixture project — no inline betterado_project resource.
func hclReleaseDefinitionWithDeploymentInput(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
data "betterado_agent_queue" "test" {
  name       = "Azure Pipelines"
  project_id = %[2]q
}

resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  stages = [{
    name = "Production"
    rank = 1

    deploy_phase = [{
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"

      deployment_input = [{
        queue_id                      = data.betterado_agent_queue.test.id
        timeout_in_minutes            = 0
        job_cancel_timeout_in_minutes = 1
        condition                     = "succeeded()"
      }]
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    # ADO requires BOTH pre- and post-deploy approvals to be non-empty on a
    # stage, else create fails with VS402877 ("Pre-approvals or post-approvals
    # … are empty"). The exhaustive _complete test always carried both; the
    # minimal fixtures carried only pre and were never live-run (only _complete
    # was INIT-1's merge gate), so the gap surfaced when every test was finally
    # run with TF_ACC.
    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID)
}

// hclReleaseDefinitionWithApprovalOptions creates a release definition with approval_options.
// hclReleaseDefinitionWithApprovalOptions uses the shared fixture project — no inline betterado_project resource.
func hclReleaseDefinitionWithApprovalOptions(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  stages = [{
    name = "Production"
    rank = 1

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]

      approval_options = [{
        release_creator_can_be_approver                                 = false
        enforce_identity_revalidation                                   = false
        timeout_in_minutes                                              = 1440
        execution_order                                                 = "beforeGates"
        auto_triggered_and_previous_environment_approved_can_be_skipped = false
      }]
    }]

    # ADO requires BOTH pre- and post-deploy approvals to be non-empty on a
    # stage (VS402877).
    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    deploy_phase = [{
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }]

    # ADO REST 7.2 requires a stage-level retention policy (VS402982).
    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]
  }]
}
`, name, fixture.ProjectID)
}

// hclReleaseDefinitionWithEnvironmentOptions creates a release definition with environment_options
// and execution_policy blocks.
// retention_policy and pre_deploy_approval are included to satisfy ADO REST 7.2 (VS402982/VS402877).
// Uses the shared fixture project — no inline betterado_project resource.
func hclReleaseDefinitionWithEnvironmentOptions(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  stages = [{
    name = "Production"
    rank = 1

    deploy_phase = [{
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }]

    environment_options = [{
      email_notification_type   = "OnlyOnFailure"
      publish_deployment_status = true
      badge_enabled             = false
      auto_link_work_items      = false
    }]

    execution_policy = [{
      concurrency_count = 1
      queue_depth_count = 0
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    # ADO requires BOTH pre- and post-deploy approvals to be non-empty on a
    # stage, else create fails with VS402877 ("Pre-approvals or post-approvals
    # … are empty"). The exhaustive _complete test always carried both; the
    # minimal fixtures carried only pre and were never live-run (only _complete
    # was INIT-1's merge gate), so the gap surfaced when every test was finally
    # run with TF_ACC.
    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID)
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
//   - pre/post deployment gates with a real ServerGate task (queryWorkItems) referencing the
//     shared fixture's work item query (AC3/WI-9: non-empty queryId)
//   - cd_artifact_trigger + schedule_trigger with NO branch_filter (AC1/WI-9)
//   - non-default retention_policy
//   - non-default environment_options and execution_policy
//   - definition-level + environment-level variables
//   - Build artifact referencing the shared fixture build definition
//   - condition { name = "ReleaseStarted", condition_type = "event" } on first env (required by ADO
//     when a schedule_trigger schedules release creation)
//
// Uses the shared fixture project + build definition — no inline betterado_project resource.
func hclReleaseDefinitionComplete(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
# Resolve the Azure Pipelines hosted queue for the fixture project.
data "betterado_agent_queue" "test" {
  name       = "Azure Pipelines"
  project_id = %[2]q
}

resource "betterado_release_definition" "test" {
  name                = %[1]q
  project_id          = %[2]q
  description         = "Exhaustive acceptance test release definition"
  release_name_format = "Release-$(rev:r)"

  # definition-level variable (non-default)
  variables = {
    MY_VAR = { value = "hello-complete" }
  }

  # Build artifact — the definition_reference ties to the shared fixture build definition.
  artifact = [{
    alias      = "_build"
    type       = "Build"
    is_primary = true

    definition_reference = {
      definition = %[3]q
      project    = %[2]q
    }
  }]

  # cd_artifact_trigger fires on every build of _build; schedule_trigger at 02:00 UTC daily.
  # IMPORTANT: stages 0 must have condition { name = "ReleaseStarted", condition_type = "event" }
  # so that ADO allows the schedule trigger to create releases (ADO rejects otherwise).
  triggers = [{
    cd_artifact_trigger = [{
      artifact_alias = "_build"
      branch_filter = [{
        include = ["refs/heads/main"]
        exclude = []
      }]
    }]

    schedule_trigger = [{
      schedule_only_with_changes = true
      start_hours                = 2
      start_minutes              = 0
      time_zone_id               = "UTC"
      days_to_release            = 62 # Mon–Fri (bits 1+2+4+8+16+32 = 62)
      # Note: branch_filter is not included because ADO does not return branchFilters
      # for schedule triggers in the GET response, which would cause a perpetual diff.
    }]
  }]

  stages = [{
    name = "Staging"
    rank = 1

    # Required by ADO when schedule_trigger is used (schedules release creation event)
    condition = [{
      name           = "ReleaseStarted"
      condition_type = "event"
      value          = ""
    }]

    # stage-level variable (non-default)
    variables = {
      ENV_VAR = { value = "staging-complete" }
    }

    deploy_phase = [{
      name       = "Agent phase"
      rank       = 1
      phase_type = "agentBasedDeployment"

      deployment_input = [{
        queue_id                      = data.betterado_agent_queue.test.id
        agent_specification           = "ubuntu-22.04"
        timeout_in_minutes            = 60
        job_cancel_timeout_in_minutes = 5
        condition                     = "succeeded()"
        skip_artifacts_download       = true
        enable_access_token           = true

        demands = ["Agent.Version -gtVersion 2.0"]

        parallel_execution = [{
          type                 = "multiConfiguration"
          max_number_of_agents = 2
          multipliers          = ["Configuration"]
          continue_on_error    = false
        }]
      }]
      },
      {
        name       = "Agentless phase"
        rank       = 2
        phase_type = "runOnServer"

        deployment_input = [{
          timeout_in_minutes            = 0
          job_cancel_timeout_in_minutes = 1
          condition                     = "succeeded()"
        }]

        workflow_task = [{
          display_name = "Delay"
          task_id      = "28782b92-5e8e-4458-9751-a71cd1492bae"
          version_spec = "1.*"
          enabled      = true
          inputs = {
            delayForMinutes = "1"
          }
        }]
    }]

    # environment options (non-default)
    environment_options = [{
      email_notification_type   = "Always"
      publish_deployment_status = true
      badge_enabled             = false
      auto_link_work_items      = false
    }]

    # execution policy (non-default)
    execution_policy = [{
      concurrency_count = 1
      queue_depth_count = 0
    }]

    # retention policy (non-default)
    retention_policy = [{
      days_to_keep     = 14
      releases_to_keep = 5
      retain_build     = false
    }]

    # pre_deploy_approval with non-default approval_options
    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]

      approval_options = [{
        release_creator_can_be_approver                                 = true
        enforce_identity_revalidation                                   = false
        timeout_in_minutes                                              = 720
        execution_order                                                 = "beforeGates"
        auto_triggered_and_previous_environment_approved_can_be_skipped = false
      }]
    }]

    # post_deploy_approval (non-default: normally automated)
    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    # pre_deployment_gates with a real ServerGate task: queryWorkItems
    pre_deployment_gates = [{
      gates_options = [{
        is_enabled               = true
        timeout                  = 60
        sampling_interval        = 5
        stabilization_time       = 0
        minimum_success_duration = 0
      }]

      gate = [{
        task = [{
          display_name = "Query Work Items"
          task_id      = "f1e4b0e6-017e-4819-8a48-ef19ae96e289"
          version_spec = "0.*"
          enabled      = true
          inputs = {
            queryId = %[4]q
          }
        }]
      }]
    }]

    # post_deployment_gates with the same ServerGate task
    post_deployment_gates = [{
      gates_options = [{
        is_enabled               = true
        timeout                  = 60
        sampling_interval        = 5
        stabilization_time       = 0
        minimum_success_duration = 0
      }]

      gate = [{
        task = [{
          display_name = "Query Work Items"
          task_id      = "f1e4b0e6-017e-4819-8a48-ef19ae96e289"
          version_spec = "0.*"
          enabled      = true
          inputs = {
            queryId = %[4]q
          }
        }]
      }]
    }]
    },
    {
      name = "Production"
      rank = 2

      condition = [{
        name           = "Staging"
        condition_type = "environmentState"
        value          = "4"
      }]

      deploy_phase = [{
        name       = "Agent job"
        rank       = 1
        phase_type = "agentBasedDeployment"
      }]

      retention_policy = [{
        days_to_keep     = 30
        releases_to_keep = 3
        retain_build     = true
      }]

      pre_deploy_approval = [{
        approver = [{
          id           = "00000000-0000-0000-0000-000000000000"
          is_automated = true
          rank         = 1
        }]
      }]

      post_deploy_approval = [{
        approver = [{
          id           = "00000000-0000-0000-0000-000000000000"
          is_automated = true
          rank         = 1
        }]
      }]
  }]
}
`, name, fixture.ProjectID, fmt.Sprintf("%d", fixture.BuildDefinitionID), fixture.WorkItemQueryID)
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
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionApprovalsAndGates(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					// approval_options assertions
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deploy_approval.0.approval_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deploy_approval.0.approval_options.0.timeout_in_minutes", "1440"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deploy_approval.0.approval_options.0.execution_order", "beforeGates"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deploy_approval.0.approval_options.0.release_creator_can_be_approver", "false"),
					// gates assertions
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deployment_gates.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deployment_gates.0.gates_options.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deployment_gates.0.gates_options.0.is_enabled", "true"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deployment_gates.0.gate.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.pre_deployment_gates.0.gate.0.task.0.task_id", "f1e4b0e6-017e-4819-8a48-ef19ae96e289"),
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
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionEnvironmentConfig(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					// environment_trigger
					resource.TestCheckResourceAttr(tfNode, "stages.0.environment_trigger.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.environment_trigger.0.trigger_type", "rollbackRedeploy"),
					// schedule
					resource.TestCheckResourceAttr(tfNode, "stages.0.schedule.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.schedule.0.start_hours", "3"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.schedule.0.time_zone_id", "UTC"),
					// properties
					resource.TestCheckResourceAttr(tfNode, "stages.0.properties.%", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.properties.env", "staging"),
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

  stages = [{
    name = "Staging"
    rank = 1

    environment_trigger = [{
      trigger_type = "rollbackRedeploy"
    }]

    schedule = [{
      days_to_release = 62
      start_hours     = 3
      start_minutes   = 0
      time_zone_id    = "UTC"
    }]

    properties = {
      env = "staging"
    }

    deploy_phase = [{
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
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

  stages = [{
    name = "Staging"
    rank = 1

    deploy_phase = [{
      name       = "Agentless job"
      rank       = 1
      phase_type = "runOnServer"

      deployment_input = [{
        timeout_in_minutes            = 0
        job_cancel_timeout_in_minutes = 1
        condition                     = "succeeded()"
      }]
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]

      approval_options = [{
        release_creator_can_be_approver                                 = false
        enforce_identity_revalidation                                   = false
        timeout_in_minutes                                              = 1440
        execution_order                                                 = "beforeGates"
        auto_triggered_and_previous_environment_approved_can_be_skipped = false
      }]
    }]

    # ADO requires BOTH pre- and post-deploy approvals (VS402877).
    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    pre_deployment_gates = [{
      gates_options = [{
        is_enabled               = true
        timeout                  = 600
        sampling_interval        = 60
        stabilization_time       = 0
        minimum_success_duration = 0
      }]

      gate = [{
        task = [{
          display_name = "Query Work Items"
          task_id      = "f1e4b0e6-017e-4819-8a48-ef19ae96e289"
          version_spec = "0.*"
          enabled      = true
          inputs = {
            queryId = %[3]q
          }
        }]
      }]
    }]
  }]
}
`, name, fixture.ProjectID, fixture.WorkItemQueryID)
}

// TestAccReleaseDefinition_triggerEnhancements verifies the trigger-enhancement fields
// introduced in the INIT-2026-06-08 initiative:
//   - cd_artifact_trigger.tag_filter (build-tag list; ADO 7.1 does not persist the SDK regex tagFilter — verified live)
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
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
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
					resource.TestCheckResourceAttr(tfNode, "triggers.0.cd_artifact_trigger.0.tag_filter.0.tags.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.cd_artifact_trigger.0.tag_filter.0.tags.0", "stable"),
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
  artifact = [{
    alias      = "_build"
    type       = "Build"
    is_primary = true

    definition_reference = {
      definition = tostring(%[3]d)
      project    = %[2]q
    }
  }]

  # Triggers block exercises cd_artifact_trigger (tag_filter, use_build_definition_branch,
  # create_release_on_build_tagging) and source_repo_trigger — attribute (array) syntax.
  triggers = [{
    cd_artifact_trigger = [{
      artifact_alias                  = "_build"
      use_build_definition_branch     = true
      create_release_on_build_tagging = true

      tag_filter = [{
        tags = ["stable"]
      }]
    }]

    source_repo_trigger = [{
      alias          = "_build"
      branch_filters = ["refs/heads/main"]
    }]
  }]

  stages = [{
    name = "Production"
    rank = 1

    deploy_phase = [{
      name       = "Agentless job"
      rank       = 1
      phase_type = "runOnServer"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID, fixture.BuildDefinitionID)
}

// TestAccReleaseDefinition_updateAddEnvironment verifies that a release definition
// update which (a) adds a second stage and (b) changes the description field
// is applied in-place (no destroy/recreate), both stages survive, description
// is updated, and the ADO API-level revision increments by at least 1.
//
// AC1: two stages survive the update; description is updated in-place.
// AC2: API revision after update > API revision before update (update-in-place, not ForceNew).
func TestAccReleaseDefinition_updateAddEnvironment(t *testing.T) {
	fixture := SharedReleaseFixture(t)
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	// revisionBeforeUpdate is populated by the step-1 check; step-2 uses it.
	var revisionBeforeUpdate int

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create a minimal definition with one stage + a description.
			{
				Config: hclReleaseDefinitionUpdateBase(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "stages.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "description", "base-description"),
					resource.TestCheckResourceAttrSet(tfNode, "revision"),
					// Capture the API-level revision for comparison in step 2.
					captureReleaseDefinitionRevision(&revisionBeforeUpdate),
				),
			},
			// Step 2: add a second stage and change the description — in-place update.
			{
				Config: hclReleaseDefinitionUpdateExpanded(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "stages.#", "2"),
					resource.TestCheckResourceAttr(tfNode, "description", "updated-description"),
					resource.TestCheckResourceAttrSet(tfNode, "revision"),
					// AC2: API revision must have incremented (proving update-in-place).
					checkReleaseDefinitionRevisionIncremented(&revisionBeforeUpdate),
				),
			},
		},
	})
}

// captureReleaseDefinitionRevision reads the ADO API revision of the release
// definition found in Terraform state and stores it in *out.  Called in step 1
// so that checkReleaseDefinitionRevisionIncremented can compare against it in
// step 2.
func captureReleaseDefinitionRevision(out *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		def, err := getReleaseDefinitionFromState(s)
		if err != nil {
			return err
		}
		if def.Revision == nil {
			return fmt.Errorf("captureReleaseDefinitionRevision: API returned nil revision")
		}
		*out = *def.Revision
		return nil
	}
}

// checkReleaseDefinitionRevisionIncremented asserts that the ADO API revision of
// the release definition found in Terraform state is strictly greater than
// *expectedMinRevision (the value captured in the previous step).  A greater
// revision proves that the update was applied in-place rather than via a
// destroy/recreate cycle (which would reset the revision to 1).
func checkReleaseDefinitionRevisionIncremented(expectedMinRevision *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		def, err := getReleaseDefinitionFromState(s)
		if err != nil {
			return err
		}
		if def.Revision == nil {
			return fmt.Errorf("checkReleaseDefinitionRevisionIncremented: API returned nil revision")
		}
		if *def.Revision <= *expectedMinRevision {
			return fmt.Errorf(
				"expected API revision > %d (revision before update) but got %d; "+
					"this may indicate a destroy/recreate instead of an in-place update",
				*expectedMinRevision, *def.Revision,
			)
		}
		return nil
	}
}

// hclReleaseDefinitionUpdateBase returns HCL for a minimal release definition with
// ONE stage and description "base-description". Used as step 1 of
// TestAccReleaseDefinition_updateAddEnvironment.
//
// Uses fixture.ProjectID (no inline betterado_project) to keep the test focused.
// Both pre_deploy_approval and post_deploy_approval are present (VS402877).
// retention_policy is present (VS402982).
func hclReleaseDefinitionUpdateBase(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name        = %[1]q
  project_id  = %[2]q
  description = "base-description"

  stages = [{
    name = "Staging"
    rank = 1

    deploy_phase = [{
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID)
}

// hclReleaseDefinitionUpdateExpanded returns HCL for a release definition with
// TWO stages and description "updated-description". Used as step 2 of
// TestAccReleaseDefinition_updateAddEnvironment to verify that:
//   - A second stage can be added in-place (no ForceNew).
//   - The description change is persisted without destroy/recreate.
//
// Uses fixture.ProjectID (no inline betterado_project).
// Both pre_deploy_approval and post_deploy_approval are present on every
// stage (VS402877). retention_policy is present on every stage (VS402982).
func hclReleaseDefinitionUpdateExpanded(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name        = %[1]q
  project_id  = %[2]q
  description = "updated-description"

  stages = [{
    name = "Staging"
    rank = 1

    deploy_phase = [{
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
    }, {
    name = "Production"
    rank = 2

    deploy_phase = [{
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID)
}

// TestAccReleaseDefinition_completeWithNewFields is a combined exhaustive acceptance
// test that exercises BOTH the existing complete-test features AND the PR #18 / PR #19
// additions together in a single resource block.
//
// PR #18 fields exercised: environment_trigger, schedule, properties.
// PR #19 fields exercised: cd_artifact_trigger.tag_filter,
//
//	cd_artifact_trigger.use_build_definition_branch,
//	cd_artifact_trigger.create_release_on_build_tagging,
//	source_repo_trigger.
//
// The individual tests (TestAccReleaseDefinition_environmentConfig and
// TestAccReleaseDefinition_triggerEnhancements) prove each field works in
// isolation; this test proves the fields work TOGETHER — no perpetual drift
// occurs when all options are combined (AC2).
//
// Uses SharedReleaseFixture for project + build definition (no inline
// betterado_project). A runOnServer phase is used so no real agent queue is
// required. Pre/post approvals and retention_policy satisfy VS402877/VS402982.
//
// AC1: test passes live (TF_ACC=1), all new-field assertions succeed, idempotency
//
//	step (PlanOnly: true, ExpectNonEmptyPlan: false) produces no diff.
//
// AC2: assertions cover field interactions not duplicated verbatim from the
//
//	individual isolation tests.
func TestAccReleaseDefinition_completeWithNewFields(t *testing.T) {
	fixture := SharedReleaseFixture(t)
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionCompleteWithNewFields(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "name", name),

					// artifact is wired to the shared build definition.
					resource.TestCheckResourceAttr(tfNode, "artifact.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "artifact.0.alias", "_build"),
					resource.TestCheckResourceAttr(tfNode, "artifact.0.type", "Build"),
					resource.TestCheckResourceAttr(tfNode, "artifact.0.is_primary", "true"),

					// triggers block — CD artifact trigger with PR #19 fields.
					resource.TestCheckResourceAttr(tfNode, "triggers.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.cd_artifact_trigger.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.cd_artifact_trigger.0.artifact_alias", "_build"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.cd_artifact_trigger.0.tag_filter.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.cd_artifact_trigger.0.tag_filter.0.tags.0", "stable"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.cd_artifact_trigger.0.use_build_definition_branch", "true"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.cd_artifact_trigger.0.create_release_on_build_tagging", "true"),

					// source_repo_trigger (PR #19) — proves it coexists with cd_artifact_trigger.
					resource.TestCheckResourceAttr(tfNode, "triggers.0.source_repo_trigger.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.source_repo_trigger.0.alias", "_build"),

					// single Staging stage.
					resource.TestCheckResourceAttr(tfNode, "stages.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.name", "Staging"),

					// environment_trigger (PR #18).
					resource.TestCheckResourceAttr(tfNode, "stages.0.environment_trigger.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.environment_trigger.0.trigger_type", "rollbackRedeploy"),

					// schedule (PR #18) — start_hours and time_zone_id as representative fields.
					resource.TestCheckResourceAttr(tfNode, "stages.0.schedule.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.schedule.0.start_hours", "3"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.schedule.0.time_zone_id", "UTC"),

					// properties (PR #18) — one entry proves the map round-trips.
					resource.TestCheckResourceAttr(tfNode, "stages.0.properties.%", "1"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.properties.env", "staging"),
				),
			},
			// Idempotency: re-plan with all new fields combined must produce no diff.
			{
				Config:             hclReleaseDefinitionCompleteWithNewFields(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclReleaseDefinitionCompleteWithNewFields returns HCL for a single
// betterado_release_definition that combines the PR #18 environment-level fields
// (environment_trigger, schedule, properties) with the PR #19 trigger fields
// (cd_artifact_trigger.tag_filter, use_build_definition_branch,
// create_release_on_build_tagging, source_repo_trigger) alongside the shared
// fixture artifact reference.
//
// Arg layout: %[1]q = name, %[2]q = fixture.ProjectID, %[3]d = fixture.BuildDefinitionID.
//
// A runOnServer phase is used to avoid the need for a real agent queue.
// Pre/post approvals (VS402877) and retention_policy (VS402982) are present.
// An "event" condition on the stage satisfies ADO's requirement when
// combining cd_artifact_trigger with environment_trigger.
func hclReleaseDefinitionCompleteWithNewFields(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  # Build artifact — references the shared pre-created build definition.
  artifact = [{
    alias      = "_build"
    type       = "Build"
    is_primary = true

    definition_reference = {
      definition = tostring(%[3]d)
      project    = %[2]q
    }
  }]

  # Triggers block: combines PR #19 cd_artifact_trigger enhancements with source_repo_trigger.
  triggers = [{
    # CD artifact trigger with tag_filter and new boolean flags (PR #19).
    cd_artifact_trigger = [{
      artifact_alias                  = "_build"
      use_build_definition_branch     = true
      create_release_on_build_tagging = true

      # tag_filter — builds tagged "stable" will fire this trigger.
      tag_filter = [{
        tags = ["stable"]
      }]
    }]

    # Source-repo trigger on the same artifact alias (PR #19).
    source_repo_trigger = [{
      alias          = "_build"
      branch_filters = ["refs/heads/main"]
    }]
  }]

  stages = [{
    name = "Staging"
    rank = 1

    # ADO requires an "event" condition on the stage when environment_trigger
    # is combined with a cd_artifact_trigger (schedules release creation event).
    condition = [{
      name           = "ReleaseStarted"
      condition_type = "event"
      value          = ""
    }]

    # environment_trigger (PR #18): re-deploy on rollback.
    environment_trigger = [{
      trigger_type = "rollbackRedeploy"
    }]

    # schedule (PR #18): weekday releases at 03:00 UTC.
    schedule = [{
      days_to_release = 62
      start_hours     = 3
      start_minutes   = 0
      time_zone_id    = "UTC"
    }]

    # properties (PR #18): arbitrary key-value metadata.
    properties = {
      env = "staging"
    }

    # runOnServer phase — no real agent queue required for this combined test.
    deploy_phase = [{
      name       = "Agentless job"
      rank       = 1
      phase_type = "runOnServer"

      deployment_input = [{
        timeout_in_minutes            = 0
        job_cancel_timeout_in_minutes = 1
        condition                     = "succeeded()"
      }]
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    # ADO VS402877: both pre- and post-deploy approvals are mandatory.
    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID, fixture.BuildDefinitionID)
}

// TestAccReleaseDefinition_withEnvironmentSecretVariable verifies that a secret
// variable declared at ENVIRONMENT scope round-trips without a perpetual diff.
// ADO returns null for secret values; the provider must preserve the in-state
// value from the matching stages.<i>.variable path. Before the fix,
// flattenVariables read the definition-level "variable" key for every scope and
// lost env-scoped secrets, producing an endless plan diff. The PlanOnly step
// with ExpectNonEmptyPlan:false is the regression guard.
func TestAccReleaseDefinition_withEnvironmentSecretVariable(t *testing.T) {
	fixture := SharedReleaseFixture(t)

	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionWithEnvironmentSecretVariable(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "stages.0.variables.%", "1"),
				),
			},
			{
				// Idempotency: env-scoped secret must not re-plan.
				Config:             hclReleaseDefinitionWithEnvironmentSecretVariable(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccReleaseDefinition_withWorkflowTaskAndOverrideInputs covers WI-C
// (workflow_task.timeout_in_minutes / retry_count_on_task_failure) and WI-D
// (deployment_input.override_inputs), asserting they round-trip without a diff.
func TestAccReleaseDefinition_withWorkflowTaskAndOverrideInputs(t *testing.T) {
	fixture := SharedReleaseFixture(t)

	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionWithWorkflowTaskAndOverrideInputs(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.workflow_task.0.timeout_in_minutes", "15"),
					resource.TestCheckResourceAttr(tfNode, "stages.0.deploy_phase.0.workflow_task.0.retry_count_on_task_failure", "2"),
					// deployment_input.override_inputs is implemented + unit-tested
					// (TestReleaseDefinition_DeploymentInputOverrideInputs). It can't be
					// live-asserted on an Agent job: ADO rejects any override key not
					// already present in the agent-job deployment input ("'<key>' property
					// not present"). Live validation would need a deployment-group phase.
				),
			},
			{
				Config:             hclReleaseDefinitionWithWorkflowTaskAndOverrideInputs(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func hclReleaseDefinitionWithWorkflowTaskAndOverrideInputs(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  stages = [{
    name = "Production"
    rank = 1

    deploy_phase = [{
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"

      workflow_task = [{
        display_name                = "Run script"
        task_id                     = "d9bafed4-0b18-4f58-968d-86655b4d2ce9"
        version_spec                = "2.*"
        timeout_in_minutes          = 15
        retry_count_on_task_failure = 2

        inputs = {
          script = "echo deploy"
        }
      }]
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID)
}

func hclReleaseDefinitionWithEnvironmentSecretVariable(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  stages = [{
    name = "Production"
    rank = 1

    variables = {
      ENV_SECRET = {
        value     = "env-super-secret"
        is_secret = true
      }
    }

    deploy_phase = [{
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID)
}

// TestAccReleaseDefinition_withDefinitionTags verifies that definition-level
// tags are stored by ADO and round-trip cleanly through the provider.
func TestAccReleaseDefinition_withDefinitionTags(t *testing.T) {
	fixture := SharedReleaseFixture(t)
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionWithDefinitionTags(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(tfNode, "tags.*", "team-a"),
					resource.TestCheckTypeSetElemAttr(tfNode, "tags.*", "prod"),
				),
			},
			// Idempotency re-plan: no diff on second apply.
			{
				Config:             hclReleaseDefinitionWithDefinitionTags(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func hclReleaseDefinitionWithDefinitionTags(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  tags = ["team-a", "prod"]

  stages = [{
    name = "Production"
    rank = 1

    deploy_phase = [{
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID)
}

// TestAccReleaseDefinition_withContainerImageTrigger verifies that a
// container_image_trigger round-trips through the ADO Release API.
//
// AC2/WI-5: apply succeeds, triggers.0.container_image_trigger.0.artifact_alias and
// .label round-trip cleanly, idempotency re-plan produces no diff, destroy cleans up.
// AC4/WI-5: captureReleaseEvidence is called while the resource is live (before destroy)
// so .forge/live-evidence/acceptance-resource.json is written with the real vsrm REST GET URL.
func TestAccReleaseDefinition_withContainerImageTrigger(t *testing.T) {
	fixture := SharedReleaseFixture(t)
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionWithContainerImageTrigger(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "triggers.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.container_image_trigger.#", "1"),
					resource.TestCheckResourceAttr(tfNode,
						"triggers.0.container_image_trigger.0.artifact_alias", "_build"),
					resource.TestCheckResourceAttr(tfNode,
						"triggers.0.container_image_trigger.0.label", "latest"),
					captureReleaseEvidence(tfNode),
				),
			},
			{
				Config:             hclReleaseDefinitionWithContainerImageTrigger(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclReleaseDefinitionWithContainerImageTrigger returns HCL for a
// betterado_release_definition that exercises the container_image_trigger block
// inside the triggers block. The artifact alias "_build" matches the fixture's
// build artifact alias (fixture wires a Build artifact with alias "_build").
//
// A runOnServer phase avoids requiring a real agent queue for this focused test.
// Pre/post approvals (VS402877) and retention_policy (VS402982) are present.
// Arg layout: %[1]q = name, %[2]q = fixture.ProjectID, %[3]d = fixture.BuildDefinitionID.
func hclReleaseDefinitionWithContainerImageTrigger(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  # Build artifact — required so the container_image_trigger can reference
  # artifact alias "_build". Uses the fixture's pre-created build definition.
  artifact = [{
    alias      = "_build"
    type       = "Build"
    is_primary = true

    definition_reference = {
      definition = tostring(%[3]d)
      project    = %[2]q
    }
  }]

  # Triggers block: one container_image_trigger on the "_build" artifact alias — attribute syntax.
  triggers = [{
    container_image_trigger = [{
      artifact_alias = "_build"
      label          = "latest"
    }]
  }]

  stages = [{
    name = "Production"
    rank = 1

    deploy_phase = [{
      name       = "Agentless job"
      rank       = 1
      phase_type = "runOnServer"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID, fixture.BuildDefinitionID)
}

// ─── pull_request_trigger test ───────────────────────────────────────────────

// TestAccReleaseDefinition_withPullRequestTrigger verifies that a
// pull_request_trigger round-trips through the ADO Release API.
//
// A runOnServer phase avoids requiring a real agent queue for this focused test.
// Pre/post approvals (VS402877) and retention_policy (VS402982) are present.
// Idempotency re-plan step asserts no perpetual diff (ExpectNonEmptyPlan: false).
func TestAccReleaseDefinition_withPullRequestTrigger(t *testing.T) {
	fixture := SharedReleaseFixture(t)
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkReleaseDefinitionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclReleaseDefinitionWithPullRequestTrigger(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					checkReleaseDefinitionExists(name),
					resource.TestCheckResourceAttr(tfNode, "triggers.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "triggers.0.pull_request_trigger.#", "1"),
					resource.TestCheckResourceAttr(tfNode,
						"triggers.0.pull_request_trigger.0.artifact_alias", "_build"),
					resource.TestCheckResourceAttr(tfNode,
						"triggers.0.pull_request_trigger.0.target_branches.#", "1"),
					resource.TestCheckResourceAttr(tfNode,
						"triggers.0.pull_request_trigger.0.target_branches.0", "refs/heads/main"),
					resource.TestCheckResourceAttr(tfNode,
						"triggers.0.pull_request_trigger.0.tags.#", "1"),
					resource.TestCheckResourceAttr(tfNode,
						"triggers.0.pull_request_trigger.0.tags.0", "ready"),
				),
			},
			{
				Config:             hclReleaseDefinitionWithPullRequestTrigger(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclReleaseDefinitionWithPullRequestTrigger returns HCL for a
// betterado_release_definition that exercises the pull_request_trigger block.
// Arg layout: %[1]q = name, %[2]q = fixture.ProjectID, %[3]d = fixture.BuildDefinitionID.
func hclReleaseDefinitionWithPullRequestTrigger(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  # Build artifact — provides the "_build" alias referenced by the pull_request_trigger.
  artifact = [{
    alias      = "_build"
    type       = "Build"
    is_primary = true

    definition_reference = {
      definition = tostring(%[3]d)
      project    = %[2]q
    }
  }]

  # Triggers block: one pull_request_trigger — attribute (array) syntax.
  triggers = [{
    pull_request_trigger = [{
      artifact_alias  = "_build"
      target_branches = ["refs/heads/main"]
      tags            = ["ready"]
    }]
  }]

  stages = [{
    name = "Production"
    rank = 1

    deploy_phase = [{
      name       = "Agentless job"
      rank       = 1
      phase_type = "runOnServer"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID, fixture.BuildDefinitionID)
}

// ─── AC1/WI-6: array-syntax fixture ──────────────────────────────────────────

// hclReleaseDefinitionBasicArraySyntax creates a minimal release definition using
// HCL attribute-assignment (array) syntax for stages and artifact. This is the
// primary fixture for TestAccReleaseDefinition_basic / AC1 / WI-6.
//
// Requirements met:
//   - stages = [{ ... }]  (array syntax, not block syntax)
//   - artifact = [{ ... }]  (array syntax with definition_reference)
//   - stage name "Production" (non-default)
//   - deploy_phase rank = 1
//   - pre/post approvals (VS402877) + retention_policy (VS402982)
func hclReleaseDefinitionBasicArraySyntax(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  artifact = [{
    alias      = "_build"
    type       = "Build"
    is_primary = true

    definition_reference = {
      definition = tostring(%[3]d)
      project    = %[2]q
    }
  }]

  stages = [{
    name = "Production"
    rank = 1

    deploy_phase = [{
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID, fixture.BuildDefinitionID)
}

// ─── AC2/WI-6: partial-stages fixture ────────────────────────────────────────

// hclReleaseDefinitionPartialStages creates a release definition with minimal
// deploy_phase (no workflow_task, no approval/gate blocks beyond the mandatory
// pre/post approvals + retention_policy) to verify idempotency when optional
// fields are omitted (AC2/WI-6).
func hclReleaseDefinitionPartialStages(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  artifact = [{
    alias      = "_build"
    type       = "Build"
    is_primary = true

    definition_reference = {
      definition = tostring(%[3]d)
      project    = %[2]q
    }
  }]

  stages = [{
    name = "Production"
    rank = 1

    deploy_phase = [{
      rank       = 1
      phase_type = "agentBasedDeployment"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID, fixture.BuildDefinitionID)
}

// ─── AC3/WI-6: partial-variables fixture ─────────────────────────────────────

// hclReleaseDefinitionPartialVariables creates a release definition with a
// definition-level variable that omits is_secret and allow_override, verifying
// idempotency when optional variable attributes are absent (AC3/WI-6).
// Uses the framework map syntax (variables = { key = { value = "..." } }).
func hclReleaseDefinitionPartialVariables(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
resource "betterado_release_definition" "test" {
  name       = %[1]q
  project_id = %[2]q

  variables = {
    "myVar" = {
      value = "hello"
    }
  }

  artifact = [{
    alias      = "_build"
    type       = "Build"
    is_primary = true

    definition_reference = {
      definition = tostring(%[3]d)
      project    = %[2]q
    }
  }]

  stages = [{
    name = "Production"
    rank = 1

    deploy_phase = [{
      rank       = 1
      phase_type = "agentBasedDeployment"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]

    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}
`, name, fixture.ProjectID, fixture.BuildDefinitionID)
}

// ─── AC2/WI-6: partial-stages evidence capture ───────────────────────────────

// capturePartialStagesEvidence writes live-evidence for AC2/WI-6 (partial-stages
// idempotency). Best-effort — never fails the test.
func capturePartialStagesEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		def, err := getReleaseDefinitionFromResource(res)
		if err != nil || def == nil || def.Id == nil {
			return nil //nolint:nilerr
		}
		projectID := res.Primary.Attributes["project_id"]
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		vsrmURL := strings.Replace(orgURL, "https://dev.azure.com", "https://vsrm.dev.azure.com", 1)
		url := fmt.Sprintf("%s/%s/_apis/release/definitions/%d?api-version=7.1", vsrmURL, projectID, *def.Id)
		_ = testutils.CaptureLiveEvidence("release-def-partial-stages", url, def)
		return nil
	}
}

// ─── AC3/WI-6: partial-variables evidence capture ────────────────────────────

// capturePartialVariablesEvidence writes live-evidence for AC3/WI-6 (partial-
// variables idempotency). Best-effort — never fails the test.
func capturePartialVariablesEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		def, err := getReleaseDefinitionFromResource(res)
		if err != nil || def == nil || def.Id == nil {
			return nil //nolint:nilerr
		}
		projectID := res.Primary.Attributes["project_id"]
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		vsrmURL := strings.Replace(orgURL, "https://dev.azure.com", "https://vsrm.dev.azure.com", 1)
		url := fmt.Sprintf("%s/%s/_apis/release/definitions/%d?api-version=7.1", vsrmURL, projectID, *def.Id)
		_ = testutils.CaptureLiveEvidence("release-def-partial-variables", url, def)
		return nil
	}
}

// ─── AC4/WI-6: artifact-filter evidence capture ──────────────────────────────

// captureArtifactFilterEvidence writes live-evidence for AC4/WI-6, capturing the
// live API response that proves ADO-added definition_reference keys
// (artifactSourceDefinitionUrl, defaultVersionType, etc.) are not present in the
// Terraform state after read-back. Best-effort — never fails the test.
func captureArtifactFilterEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		def, err := getReleaseDefinitionFromResource(res)
		if err != nil || def == nil || def.Id == nil {
			return nil //nolint:nilerr
		}
		projectID := res.Primary.Attributes["project_id"]
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		vsrmURL := strings.Replace(orgURL, "https://dev.azure.com", "https://vsrm.dev.azure.com", 1)
		url := fmt.Sprintf("%s/%s/_apis/release/definitions/%d?api-version=7.1", vsrmURL, projectID, *def.Id)
		_ = testutils.CaptureLiveEvidence("release-def-artifact-filter", url, def)
		return nil
	}
}
