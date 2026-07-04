package acceptancetests

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	taskagentsvc "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent"
)

func TestAccVariableGroupVariable_basic(t *testing.T) {
	vgName := testutils.GenerateResourceName()
	node := "betterado_variable_group_variable.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkVariableGroupVariableDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclVariableGroupVariableBasic(vgName, "foo"),
				Check: resource.ComposeTestCheckFunc(
					checkVariableGroupVariableExists(node),
					resource.TestCheckResourceAttr(node, "name", "test-key"),
					resource.TestCheckResourceAttr(node, "value", "foo"),
					resource.TestCheckResourceAttrSet(node, "project_id"),
					resource.TestCheckResourceAttrSet(node, "variable_group_id"),
					captureVariableGroupVariableEvidence(node),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      node,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: hclVariableGroupVariableBasic(vgName, "bar"),
				Check: resource.ComposeTestCheckFunc(
					checkVariableGroupVariableExists(node),
					resource.TestCheckResourceAttr(node, "name", "test-key"),
					resource.TestCheckResourceAttr(node, "value", "bar"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      node,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVariableGroupVariable_secret(t *testing.T) {
	vgName := testutils.GenerateResourceName()
	node := "betterado_variable_group_variable.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkVariableGroupVariableDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclVariableGroupVariableSecret(vgName, "foo"),
				Check: resource.ComposeTestCheckFunc(
					checkVariableGroupVariableExists(node),
					resource.TestCheckResourceAttr(node, "name", "test-key"),
					resource.TestCheckResourceAttr(node, "secret_value", "foo"),
					resource.TestCheckResourceAttrSet(node, "project_id"),
					resource.TestCheckResourceAttrSet(node, "variable_group_id"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:            node,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_value"},
			},
			{
				Config: hclVariableGroupVariableSecret(vgName, "bar"),
				Check: resource.ComposeTestCheckFunc(
					checkVariableGroupVariableExists(node),
					resource.TestCheckResourceAttr(node, "name", "test-key"),
					resource.TestCheckResourceAttr(node, "secret_value", "bar"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:            node,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_value"},
			},
		},
	})
}

func checkVariableGroupVariableDestroyed(s *terraform.State) error {
	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_variable_group_variable" {
			continue
		}

		ok, err := checkVariableGroupVariableFromState(res)
		if err == nil && ok {
			return fmt.Errorf("variable still exists")
		}
	}

	return nil
}

func checkVariableGroupVariableExists(node string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		state, ok := s.RootModule().Resources[node]
		if !ok {
			return fmt.Errorf("Did not find a variable group in the TF state")
		}

		ok, err := checkVariableGroupVariableFromState(state)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("%q doesn't exist", node)
		}
		return nil
	}
}

func checkVariableGroupVariableFromState(res *terraform.ResourceState) (bool, error) {
	projectId, groupId, name, err := taskagentsvc.ResourceVariableGroupVariableParseId(res.Primary.ID)
	if err != nil {
		return false, err
	}

	clients, err := testutils.GetDirectClient()
	if err != nil {
		return false, fmt.Errorf("GetDirectClient: %v", err)
	}

	resp, err := clients.TaskAgentClient.GetVariableGroup(
		clients.Ctx,
		taskagent.GetVariableGroupArgs{
			GroupId: &groupId,
			Project: &projectId,
		},
	)
	if err != nil {
		return false, err
	}

	if resp.Variables == nil {
		return false, fmt.Errorf("unexpected null variables in group response")
	}

	vars := *resp.Variables
	_, ok := vars[name]
	return ok, nil
}

// captureVariableGroupVariableEvidence fetches the variable group via the ADO REST API
// and persists the response as forge demo live-evidence (before the resource is destroyed).
// Best-effort: a capture failure never fails the test.
func captureVariableGroupVariableEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		vgIDStr := res.Primary.Attributes["variable_group_id"]
		vgID, parseErr := strconv.Atoi(vgIDStr)
		if parseErr != nil {
			return nil //nolint:nilerr // best-effort evidence capture
		}
		projectID := res.Primary.Attributes["project_id"]
		clients, clientErr := testutils.GetDirectClient()
		if clientErr != nil {
			return nil //nolint:nilerr // best-effort evidence capture
		}
		vg, getErr := clients.TaskAgentClient.GetVariableGroup(clients.Ctx, taskagent.GetVariableGroupArgs{
			GroupId: &vgID,
			Project: &projectID,
		})
		if getErr != nil || vg == nil {
			return nil //nolint:nilerr // best-effort evidence capture
		}
		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		if len(orgURL) > 0 && orgURL[len(orgURL)-1] == '/' {
			orgURL = orgURL[:len(orgURL)-1]
		}
		url := fmt.Sprintf("%s/%s/_apis/distributedtask/variablegroups/%s?api-version=7.1", orgURL, projectID, vgIDStr)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-variable-group-variable", url, vg)
		return nil
	}
}

func TestAccVariableGroupVariable_ForEach_ConcurrentCreate(t *testing.T) {
	vgName := testutils.GenerateResourceName()

	var nodes []string

	for i := 0; i < 20; i++ {
		nodes = append(nodes, fmt.Sprintf(`betterado_variable_group_variable.test.%d`, i))
	}

	var checks []resource.TestCheckFunc
	for _, n := range nodes {
		checks = append(checks, checkVariableGroupVariableExists(n))
	}

	steps := []resource.TestStep{
		{
			Config:             hclVariableGroupVariableForEach(vgName),
			Check:              resource.ComposeTestCheckFunc(checks...),
			ExpectNonEmptyPlan: false,
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps:                    steps,
	})
}

// ── HCL fixtures using the standing fixture project ───────────────────────────

func hclVariableGroupVariableBasic(variableGroupName, val string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[3]q
}

resource "betterado_variable_group" "test" {
  project_id   = data.betterado_project.fixture.id
  name         = %[1]q
  description  = "test description"
  allow_access = false
  variable = [
    {
      name  = "key1"
      value = "value1"
    },
    {
      name         = "skey1"
      secret_value = "svalue1"
      is_secret    = true
    },
  ]
  lifecycle {
    ignore_changes = [variable]
  }
}
resource "betterado_variable_group_variable" "test" {
  project_id        = data.betterado_project.fixture.id
  variable_group_id = betterado_variable_group.test.id
  name              = "test-key"
  value             = %[2]q
}
`, variableGroupName, val, SharedFixtureProjectName)
}

func hclVariableGroupVariableSecret(variableGroupName, val string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[3]q
}

resource "betterado_variable_group" "test" {
  project_id   = data.betterado_project.fixture.id
  name         = %[1]q
  description  = "test description"
  allow_access = false
  variable = [
    {
      name  = "key1"
      value = "value1"
    },
    {
      name         = "skey1"
      secret_value = "svalue1"
      is_secret    = true
    },
  ]
  lifecycle {
    ignore_changes = [variable]
  }
}
resource "betterado_variable_group_variable" "test" {
  project_id        = data.betterado_project.fixture.id
  variable_group_id = betterado_variable_group.test.id
  name              = "test-key"
  secret_value      = %[2]q
}
`, variableGroupName, val, SharedFixtureProjectName)
}

func hclVariableGroupVariableForEach(variableGroupName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[2]q
}

resource "betterado_variable_group" "test" {
  project_id   = data.betterado_project.fixture.id
  name         = %[1]q
  description  = "test description"
  allow_access = false

  variable = [{
    name  = "seed"
    value = "seed"
  }]
  lifecycle {
    ignore_changes = [variable]
  }
}

resource "betterado_variable_group_variable" "test" {
  count             = 20
  project_id        = data.betterado_project.fixture.id
  variable_group_id = betterado_variable_group.test.id
  name              = "key${count.index}"
  value             = "val${count.index}"
}
`, variableGroupName, SharedFixtureProjectName)
}
