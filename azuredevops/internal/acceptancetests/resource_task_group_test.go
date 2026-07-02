//go:build (all || resource_task_group) && !exclude_resource_task_group

package acceptancetests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// getDirectClient is defined in direct_client_test.go (no build tag) so it is
// available to all test files in this package, including those without build tags.

func TestAccTaskGroup_basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_task_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkTaskGroupDestroyed,
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
					// Capture a real API GET of the live resource as forge demo
					// evidence (before destroy) — proves the demo shows the actual
					// task group, not a test-name table.
					captureTaskGroupEvidence(tfNode),
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
`, name, SharedFixtureProjectName)
}

func TestAccTaskGroup_withGapFields(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_task_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkTaskGroupDestroyed,
		Steps: []resource.TestStep{
			// Step 1: create with gap fields and assert exact non-default read-back
			{
				Config: hclTaskGroupWithGapFields(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "icon_url", "https://cdn.vsassets.io/v/someicon.png"),
					resource.TestCheckResourceAttr(tfNode, "category", "Deploy"),
					resource.TestCheckResourceAttr(tfNode, "input.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "input.0.visible_rule", "targetType = filePath"),
					resource.TestCheckResourceAttr(tfNode, "input.0.aliases.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "input.0.aliases.0", "targetEnvAlias"),
					resource.TestCheckResourceAttr(tfNode, "input.0.properties.EndpointId", ""),
					resource.TestCheckResourceAttrSet(tfNode, "revision"),
					captureTaskGroupEvidence(tfNode),
				),
			},
			// Step 2: idempotency — no perpetual diff
			{
				Config:             hclTaskGroupWithGapFields(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func hclTaskGroupWithGapFields(name string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_task_group" "test" {
  project_id    = data.betterado_project.test.id
  name          = "%[1]s"
  friendly_name = "%[1]s"
  description   = "Gap-fields acceptance test"
  category      = "Deploy"
  icon_url      = "https://cdn.vsassets.io/v/someicon.png"

  version = [{
    major = 1
    minor = 0
    patch = 0
  }]

  input = [{
    name         = "targetEnv"
    label        = "Target Environment"
    type         = "string"
    visible_rule = "targetType = filePath"
    properties   = { "EndpointId" = "" }
    aliases      = ["targetEnvAlias"]
  }]

  task = [{
    display_name = "Deploy Step"
    task_id      = "d9bafed4-0b18-4f58-968d-86655b4d2ce9"
    task_version = "2.*"
  }]
}
`, name, SharedFixtureProjectName)
}

func checkTaskGroupDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("checkTaskGroupDestroyed: failed to build ADO client: %v", err)
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

// captureTaskGroupEvidence performs a real live API GET of the created task group
// and persists the response as forge demo live-evidence (before the resource is
// destroyed). Best-effort: a capture failure never fails the test — the read-back
// assertions above are the authoritative live proof; this only feeds the demo.
func captureTaskGroupEvidence(tfNode string) resource.TestCheckFunc {
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
		url := fmt.Sprintf("%s/%s/_apis/distributedtask/taskgroups/%s?api-version=7.1", orgURL, projectID, tgID)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", url, (*taskGroups)[0])
		return nil
	}
}
