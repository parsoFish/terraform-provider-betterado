package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

func TestAccWorkitemtrackingprocessRule_Basic(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: multipleConditionsRule(workItemTypeName, processName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(tfNode, tfjsonpath.New("id"), knownvalue.NotNull()),
				},
				Check: resource.ComposeTestCheckFunc(
					captureRuleEvidence(tfNode),
				),
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: ruleImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkitemtrackingprocessRule_Update(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: multipleConditionsRule(workItemTypeName, processName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(tfNode, tfjsonpath.New("id"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: ruleImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: updatedRule(workItemTypeName, processName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(tfNode, tfjsonpath.New("id"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: ruleImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkitemtrackingprocessRule_ConditionTypes(t *testing.T) {
	testCases := []struct {
		conditionType string
		field         string
		value         string
	}{
		{"when", "System.State", "New"},
		{"whenNot", "System.State", "Closed"},
		{"whenChanged", "System.Title", ""},
		{"whenNotChanged", "System.Title", ""},
		{"whenWas", "System.State", "New"},
	}

	for _, tc := range testCases {
		t.Run(tc.conditionType, func(t *testing.T) {
			workItemTypeName := testutils.GenerateWorkItemTypeName()
			processName := testutils.GenerateResourceName()
			tfNode := "betterado_workitemtrackingprocess_rule.test"

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { testutils.PreCheck(t, nil) },
				ProtoV6ProviderFactories: testutils.GetProviderFactories(),
				CheckDestroy:             testutils.CheckProcessDestroyed,
				Steps: []resource.TestStep{
					{
						Config: ruleWithConditionType(workItemTypeName, processName, tc.conditionType, tc.field, tc.value),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttrSet(tfNode, "id"),
						),
					},
					{
						ResourceName:      tfNode,
						ImportStateIdFunc: ruleImportStateIdFunc(tfNode),
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}

func TestAccWorkitemtrackingprocessRule_ConditionGroupMembership(t *testing.T) {
	testCases := []string{
		"whenCurrentUserIsMemberOfGroup",
		"whenCurrentUserIsNotMemberOfGroup",
	}

	for _, conditionType := range testCases {
		t.Run(conditionType, func(t *testing.T) {
			workItemTypeName := testutils.GenerateWorkItemTypeName()
			processName := testutils.GenerateResourceName()
			groupName := testutils.GenerateResourceName()
			tfNode := "betterado_workitemtrackingprocess_rule.test"

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { testutils.PreCheck(t, nil) },
				ProtoV6ProviderFactories: testutils.GetProviderFactories(),
				CheckDestroy:             testutils.CheckProcessDestroyed,
				Steps: []resource.TestStep{
					{
						Config: ruleWithGroupMembershipCondition(workItemTypeName, processName, "", groupName, conditionType),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttrSet(tfNode, "id"),
						),
					},
					{
						ResourceName:      tfNode,
						ImportStateIdFunc: ruleImportStateIdFunc(tfNode),
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}

func TestAccWorkitemtrackingprocessRule_ActionTypes(t *testing.T) {
	testCases := []struct {
		actionType  string
		targetField string
		value       string
	}{
		{"makeRequired", "System.Title", ""},
		{"makeReadOnly", "System.Title", ""},
		{"setDefaultValue", "System.Title", "Default Title"},
		{"setDefaultFromClock", "System.ChangedDate", ""},
		{"setDefaultFromField", "System.Description", "System.Title"},
		{"copyValue", "System.Title", "Copied Value"},
		{"copyFromClock", "System.ChangedDate", ""},
		{"copyFromCurrentUser", "System.AssignedTo", ""},
		{"copyFromField", "System.Description", "System.Title"},
		{"setValueToEmpty", "System.Description", ""},
		{"copyFromServerClock", "System.ChangedDate", ""},
		{"copyFromServerCurrentUser", "System.AssignedTo", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.actionType, func(t *testing.T) {
			workItemTypeName := testutils.GenerateWorkItemTypeName()
			processName := testutils.GenerateResourceName()
			tfNode := "betterado_workitemtrackingprocess_rule.test"

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { testutils.PreCheck(t, nil) },
				ProtoV6ProviderFactories: testutils.GetProviderFactories(),
				CheckDestroy:             testutils.CheckProcessDestroyed,
				Steps: []resource.TestStep{
					{
						Config: ruleWithActionType(workItemTypeName, processName, tc.actionType, tc.targetField, tc.value),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttrSet(tfNode, "id"),
						),
					},
					{
						ResourceName:      tfNode,
						ImportStateIdFunc: ruleImportStateIdFunc(tfNode),
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}

func TestAccWorkitemtrackingprocessRule_HideTargetField(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	groupName := testutils.GenerateResourceName()
	fieldName := generateFieldName()
	tfNode := "betterado_workitemtrackingprocess_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: ruleWithHideTargetField(workItemTypeName, processName, "", groupName, fieldName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(tfNode, tfjsonpath.New("id"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: ruleImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWorkitemtrackingprocessRule_DisallowValue(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfNode := "betterado_workitemtrackingprocess_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckProcessDestroyed,
		Steps: []resource.TestStep{
			{
				Config: ruleWithDisallowValue(workItemTypeName, processName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(tfNode, tfjsonpath.New("id"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: ruleImportStateIdFunc(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func updatedRule(workItemTypeName string, processName string) string {
	workItemType := basicWorkItemType(workItemTypeName, processName)
	return fmt.Sprintf(`
%s

resource "betterado_workitemtrackingprocess_rule" "test" {
  process_id        = betterado_workitemtrackingprocess_process.test.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  name              = "Updated Rule"
  is_enabled        = false

  condition {
    condition_type = "when"
    field          = "System.State"
    value          = "Active"
  }

  action {
    action_type  = "makeReadOnly"
    target_field = "System.Title"
  }
}
`, workItemType)
}

func multipleConditionsRule(workItemTypeName string, processName string) string {
	workItemType := basicWorkItemType(workItemTypeName, processName)
	return fmt.Sprintf(`
%s

resource "betterado_workitemtrackingprocess_rule" "test" {
  process_id        = betterado_workitemtrackingprocess_process.test.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  name              = "Multiple Conditions Rule"

  condition {
    condition_type = "whenWas"
    field          = "System.State"
    value          = "New"
  }

  condition {
    condition_type = "when"
    field          = "System.State"
    value          = "Active"
  }

  condition {
    condition_type = "whenChanged"
    field          = "System.Title"
  }

  action {
    action_type  = "makeRequired"
    target_field = "System.Title"
  }

  action {
    action_type  = "makeReadOnly"
    target_field = "System.Description"
  }
}
`, workItemType)
}

func ruleWithGroupMembershipCondition(workItemTypeName, processName, _, groupName, conditionType string) string {
	workItemType := basicWorkItemType(workItemTypeName, processName)
	return fmt.Sprintf(`
%s

resource "betterado_group_entitlement" "test" {
  display_name = "%s"
}

resource "betterado_workitemtrackingprocess_rule" "test" {
  process_id        = betterado_workitemtrackingprocess_process.test.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  name              = "Test %s Rule"

  condition {
    condition_type = "%s"
    value          = betterado_group_entitlement.test.id
  }

  action {
    action_type  = "makeRequired"
    target_field = "System.Title"
  }
}
`, workItemType, groupName, conditionType, conditionType)
}

func ruleWithConditionType(workItemTypeName, processName, conditionType, field, value string) string {
	workItemType := basicWorkItemType(workItemTypeName, processName)

	fieldAttr := ""
	if field != "" {
		fieldAttr = fmt.Sprintf(`field = "%s"`, field)
	}
	valueAttr := ""
	if value != "" {
		valueAttr = fmt.Sprintf(`value = "%s"`, value)
	}

	return fmt.Sprintf(`
%s

resource "betterado_workitemtrackingprocess_rule" "test" {
  process_id        = betterado_workitemtrackingprocess_process.test.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  name              = "Test %s Rule"

  condition {
    condition_type = "%s"
    %s
    %s
  }

  action {
    action_type  = "makeRequired"
    target_field = "System.Title"
  }
}
`, workItemType, conditionType, conditionType, fieldAttr, valueAttr)
}

func ruleWithHideTargetField(workItemTypeName, processName, _, groupName, fieldName string) string {
	workItemType := basicWorkItemType(workItemTypeName, processName)
	return fmt.Sprintf(`
%s

resource "betterado_group_entitlement" "test" {
  display_name = "%s"
}

resource "betterado_workitemtracking_field" "test" {
  name           = "%s"
  reference_name = "Custom.%s"
  type           = "string"
}

resource "betterado_workitemtrackingprocess_field" "test" {
  process_id        = betterado_workitemtrackingprocess_process.test.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.test.id
  field_id          = betterado_workitemtracking_field.test.id
}

resource "betterado_workitemtrackingprocess_rule" "test" {
  process_id        = betterado_workitemtrackingprocess_process.test.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  name              = "Test hideTargetField Rule"

  condition {
    condition_type = "whenCurrentUserIsNotMemberOfGroup"
    value          = betterado_group_entitlement.test.id
  }

  action {
    action_type  = "hideTargetField"
    target_field = betterado_workitemtrackingprocess_field.test.field_id
  }
}
`, workItemType, groupName, fieldName, fieldName)
}

func ruleWithDisallowValue(workItemTypeName, processName string) string {
	workItemType := basicWorkItemType(workItemTypeName, processName)
	return fmt.Sprintf(`
%s

resource "betterado_workitemtrackingprocess_rule" "test" {
  process_id        = betterado_workitemtrackingprocess_process.test.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  name              = "Test disallowValue Rule"

  condition {
    condition_type = "whenWas"
    field          = "System.State"
    value          = "New"
  }

  action {
    action_type  = "disallowValue"
    target_field = "System.State"
    value        = "Closed"
  }
}
`, workItemType)
}

func ruleWithActionType(workItemTypeName, processName, actionType, targetField, value string) string {
	workItemType := basicWorkItemType(workItemTypeName, processName)

	valueAttr := ""
	if value != "" {
		valueAttr = fmt.Sprintf(`value = "%s"`, value)
	}

	return fmt.Sprintf(`
%s

resource "betterado_workitemtrackingprocess_rule" "test" {
  process_id        = betterado_workitemtrackingprocess_process.test.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.test.reference_name
  name              = "Test %s Rule"

  condition {
    condition_type = "when"
    field          = "System.State"
    value          = "New"
  }

  action {
    action_type  = "%s"
    target_field = "%s"
    %s
  }
}
`, workItemType, actionType, actionType, targetField, valueAttr)
}

// captureRuleEvidence performs a real live API GET of the created rule and
// persists the response as forge demo live-evidence (before the resource is
// destroyed). Best-effort: a capture failure never fails the test.
func captureRuleEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		ruleID, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return nil //nolint:nilerr // best-effort evidence capture; parse failure must not fail the test
		}
		processID := res.Primary.Attributes["process_id"]
		witRefName := res.Primary.Attributes["work_item_type_id"]
		ruleURL := res.Primary.Attributes["url"]

		// Build a direct client from env vars (GetProviderFactories does not
		// expose the underlying AggregatedClient via Meta).
		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
		clients, err := client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
		if err != nil {
			return nil //nolint:nilerr // best-effort: client build failure must not fail the test
		}

		rule, err := clients.WorkItemTrackingProcessClient.GetProcessWorkItemTypeRule(
			clients.Ctx,
			workitemtrackingprocess.GetProcessWorkItemTypeRuleArgs{
				ProcessId:  converter.UUID(processID),
				WitRefName: &witRefName,
				RuleId:     &ruleID,
			},
		)
		if err != nil || rule == nil {
			return nil //nolint:nilerr // best-effort: API failure must not fail the test
		}

		_ = testutils.CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-rule", ruleURL, rule)
		return nil
	}
}

func ruleImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		processId := rs.Primary.Attributes["process_id"]
		witRefName := rs.Primary.Attributes["work_item_type_id"]
		ruleId := rs.Primary.ID
		return fmt.Sprintf("%s/%s/%s", processId, witRefName, ruleId), nil
	}
}
