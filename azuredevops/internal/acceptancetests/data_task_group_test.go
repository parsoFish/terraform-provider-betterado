//go:build (all || data_task_group) && !exclude_data_task_group

package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccTaskGroupDataSource_basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfResNode := "betterado_task_group.test"
	tfDataNode := "data.betterado_task_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkTaskGroupDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclTaskGroupDataSourceBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(tfDataNode, "name", tfResNode, "name"),
					resource.TestCheckResourceAttrPair(tfDataNode, "description", tfResNode, "description"),
					resource.TestCheckResourceAttrPair(tfDataNode, "category", tfResNode, "category"),
					resource.TestCheckResourceAttrSet(tfDataNode, "id"),
				),
			},
			// Idempotency — no perpetual diff
			{
				Config:             hclTaskGroupDataSourceBasic(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func hclTaskGroupDataSourceBasic(name string) string {
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
    display_name = "Echo Step"
    task_id      = "d9bafed4-0b18-4f58-968d-86655b4d2ce9"
    task_version = "2.*"
  }
}

data "betterado_task_group" "test" {
  project_id = betterado_task_group.test.project_id
  id         = betterado_task_group.test.id
}
`, name)
}
