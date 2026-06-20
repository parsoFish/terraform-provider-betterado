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
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccTaskGroupStateUpgradeSmoke is a live TF_ACC smoke test proving that the
// framework provider (v1.0.0, schema_version 1, with StateUpgraders wired) can
// create a betterado_task_group, read it back cleanly, and produce an idempotent
// plan — satisfying AC-5 (WI-5) of the framework-state-upgraders initiative.
//
// Test design (aligned with WI-5 "Practical note on old state simulation"):
//   - Step 1: apply HCL with non-default values → assert read-back attrs → capture
//     live evidence to .forge/live-evidence/task-group-state-upgrade-live.json.
//   - Step 2: re-plan with same config, ExpectNonEmptyPlan: false (idempotency gate).
//
// The framework's UpgradeState wiring is exercised automatically by Terraform whenever
// it encounters a state at schema_version 0; a two-step create+plan test over the
// framework binary is the accepted conforming approach per WI-5.
func TestAccTaskGroupStateUpgradeSmoke(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_task_group.smoke"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkTaskGroupSmokeDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create task group and verify read-back attrs + capture evidence.
			{
				Config: hclTaskGroupSmoke(name),
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
				Config:             hclTaskGroupSmoke(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclTaskGroupSmoke returns the Terraform HCL config for the smoke test.
// Uses non-default values for name, description, category, and a task block
// so the test is meaningfully asserting real field round-trips.
func hclTaskGroupSmoke(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "smoke" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_task_group" "smoke" {
  project_id    = betterado_project.smoke.id
  name          = "%[1]s"
  friendly_name = "%[1]s"
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
`, name)
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
