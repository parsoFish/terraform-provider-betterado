//go:build (all || resource_task_group) && !exclude_resource_task_group

package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

func TestAccTaskGroup_basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_task_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkTaskGroupDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create + assert read-back
			{
				Config: hclTaskGroupBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttr(tfNode, "description", "Acceptance test task group"),
					resource.TestCheckResourceAttr(tfNode, "category", "Build"),
					resource.TestCheckResourceAttrSet(tfNode, "revision"),
					resource.TestCheckResourceAttr(tfNode, "input.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "input.0.name", "myParam"),
					resource.TestCheckResourceAttr(tfNode, "task.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "task.0.display_name", "Echo Step"),
				),
			},
			// Step 2: idempotency — no perpetual diff
			{
				Config:             hclTaskGroupBasic(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func hclTaskGroupBasic(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_task_group" "test" {
  project_id    = betterado_project.test.id
  name          = "%[1]s"
  friendly_name = "%[1]s"
  description   = "Acceptance test task group"
  category      = "Build"

  version {
    major = 1
    minor = 0
    patch = 0
  }

  input {
    name  = "myParam"
    label = "My Parameter"
    type  = "string"
  }

  task {
    display_name  = "Echo Step"
    task_id       = "d9bafed4-0b18-4f58-968d-86655b4d2ce9"
    task_version  = "2.*"
  }
}
`, name)
}

func checkTaskGroupDestroyed(s *terraform.State) error {
	clients := testutils.GetProvider().Meta().(*client.AggregatedClient)

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
			if utils.ResponseWasNotFound(err) {
				// resource is gone — expected
				continue
			}
			return fmt.Errorf("error reading task group %s after destroy: %v", tgID, err)
		}
		if taskGroups != nil && len(*taskGroups) > 0 {
			return fmt.Errorf("task group %s still exists after destroy", tgID)
		}
	}

	return nil
}
