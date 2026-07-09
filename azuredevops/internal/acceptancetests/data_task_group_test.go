//go:build (all || data_task_group) && !exclude_data_task_group

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

func TestAccTaskGroupDataSource_basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfResNode := "betterado_task_group.test"
	tfDataNode := "data.betterado_task_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkTaskGroupDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclTaskGroupDataSourceBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(tfDataNode, "name", tfResNode, "name"),
					resource.TestCheckResourceAttrPair(tfDataNode, "description", tfResNode, "description"),
					resource.TestCheckResourceAttrPair(tfDataNode, "category", tfResNode, "category"),
					resource.TestCheckResourceAttrSet(tfDataNode, "id"),
					captureTaskGroupDataSourceEvidence(tfResNode),
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
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_task_group" "test" {
  project_id    = data.betterado_project.test.id
  name          = "%[1]s"
  friendly_name = "%[1]s"
  description   = "Acceptance test task group"
  category      = "Build"

  version = [{
    major = 1
    minor = 0
    patch = 0
  }]

  input = [{
    name  = "myParam"
    label = "My Parameter"
    type  = "string"
  }]

  task = [{
    display_name = "Echo Step"
    task_id      = "d9bafed4-0b18-4f58-968d-86655b4d2ce9"
    task_version = "2.*"
  }]
}

data "betterado_task_group" "test" {
  project_id = betterado_task_group.test.project_id
  id         = betterado_task_group.test.id
}
`, name, SharedFixtureProjectName)
}

// captureTaskGroupDataSourceEvidence performs a real live API GET of the task group
// resource (before destroy) and writes forge demo evidence to
// .forge/live-evidence/acceptance-resource-task-group-datasource.json, satisfying the
// AC3 evidence requirement. Best-effort: never fails the test.
func captureTaskGroupDataSourceEvidence(tfResNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfResNode]
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
			return nil
		}
		taskGroups, err := clients.TaskAgentClient.GetTaskGroups(clients.Ctx, taskagent.GetTaskGroupsArgs{
			Project:     &projectID,
			TaskGroupId: &tgID,
		})
		if err != nil || taskGroups == nil || len(*taskGroups) == 0 {
			return nil
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/%s/_apis/distributedtask/taskgroups/%s?api-version=7.1", orgURL, projectID, tgID)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-task-group-datasource", url, (*taskGroups)[0])
		return nil
	}
}
