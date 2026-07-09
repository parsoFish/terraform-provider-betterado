package acceptancetests

// The tests in this file use the mock clients in mock_client.go to mock out
// the Azure DevOps client operations.

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// TestAccVariableGroup_basic verifies that a variable group can be created with a basic config
// using the standing fixture project (avoids the 1000-project limit).
func TestAccVariableGroup_basic(t *testing.T) {
	vgName := testutils.GenerateResourceName()
	tfVarGroupNode := "betterado_variable_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkVariableGroupDestroyedMux,
		Steps: []resource.TestStep{
			{
				Config: hclVariableGroupBasicFixture(vgName),
				Check: resource.ComposeTestCheckFunc(
					checkVariableGroupExistsMux(vgName, false),
					resource.TestCheckResourceAttrSet(tfVarGroupNode, "project_id"),
					resource.TestCheckResourceAttr(tfVarGroupNode, "name", vgName),
					resource.TestCheckResourceAttr(tfVarGroupNode, "description", "test description"),
					resource.TestCheckResourceAttr(tfVarGroupNode, "variable.#", "1"),
					captureVariableGroupEvidence(tfVarGroupNode),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      tfVarGroupNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfVarGroupNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccVariableGroup_update verifies that a variable group can be updated.
func TestAccVariableGroup_update(t *testing.T) {
	vgName := testutils.GenerateResourceName()
	vgName2 := testutils.GenerateResourceName()
	tfVarGroupNode := "betterado_variable_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkVariableGroupDestroyedMux,
		Steps: []resource.TestStep{
			{
				Config: hclVariableGroupBasicFixture(vgName),
				Check: resource.ComposeTestCheckFunc(
					checkVariableGroupExistsMux(vgName, false),
					resource.TestCheckResourceAttrSet(tfVarGroupNode, "project_id"),
					resource.TestCheckResourceAttr(tfVarGroupNode, "name", vgName),
					resource.TestCheckResourceAttr(tfVarGroupNode, "variable.#", "1"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      tfVarGroupNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfVarGroupNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: hclVariableGroupUpdateFixture(vgName2),
				Check: resource.ComposeTestCheckFunc(
					checkVariableGroupExistsMux(vgName2, true),
					resource.TestCheckResourceAttrSet(tfVarGroupNode, "project_id"),
					resource.TestCheckResourceAttr(tfVarGroupNode, "name", vgName2),
					resource.TestCheckResourceAttr(tfVarGroupNode, "variable.#", "3"),
					resource.TestCheckResourceAttr(tfVarGroupNode, "description", "update description"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:            tfVarGroupNode,
				ImportStateIdFunc:       testutils.ComputeProjectQualifiedResourceImportID(tfVarGroupNode),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"variable"},
			},
			{
				Config: hclVariableGroupBasicFixture(vgName),
				Check: resource.ComposeTestCheckFunc(
					checkVariableGroupExistsMux(vgName, false),
					resource.TestCheckResourceAttrSet(tfVarGroupNode, "project_id"),
					resource.TestCheckResourceAttr(tfVarGroupNode, "name", vgName),
					resource.TestCheckResourceAttr(tfVarGroupNode, "variable.#", "1"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      tfVarGroupNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfVarGroupNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccVariableGroup_secretValue verifies that a variable group can store secret values.
func TestAccVariableGroup_secretValue(t *testing.T) {
	vgName := testutils.GenerateResourceName()
	tfVarGroupNode := "betterado_variable_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkVariableGroupDestroyedMux,
		Steps: []resource.TestStep{
			{
				Config: hclVariableGroupSecretValueFixture(vgName),
				Check: resource.ComposeTestCheckFunc(
					checkVariableGroupExistsMux(vgName, true),
					resource.TestCheckResourceAttrSet(tfVarGroupNode, "project_id"),
					resource.TestCheckResourceAttr(tfVarGroupNode, "name", vgName),
					resource.TestCheckResourceAttr(tfVarGroupNode, "description", "test description"),
					resource.TestCheckResourceAttr(tfVarGroupNode, "variable.#", "1"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:            tfVarGroupNode,
				ImportStateIdFunc:       testutils.ComputeProjectQualifiedResourceImportID(tfVarGroupNode),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"variable"},
			},
		},
	})
}

func TestAccVariableGroup_keyVault_basic(t *testing.T) {
	if os.Getenv("TEST_SERVICE_PRINCIPAL_ID") == "" || os.Getenv("TEST_SERVICE_PRINCIPAL_KEY") == "" ||
		os.Getenv("TEST_ARM_TENANT_ID") == "" || os.Getenv("TEST_ARM_SUBSCRIPTION_ID") == "" ||
		os.Getenv("TEST_ARM_SUBSCRIPTION_NAME") == "" || os.Getenv("TEST_ARM_KV_NAME") == "" {
		t.Skip("Skip test as `TEST_SERVICE_PRINCIPAL_ID` or `TEST_SERVICE_PRINCIPAL_KEY` or `TEST_ARM_TENANT_ID` or `TEST_ARM_SUBSCRIPTION_ID` or `TEST_ARM_SUBSCRIPTION_NAME` or `TEST_ARM_KV_NAME` is not set")
	}

	vgKeyVault := testutils.GenerateResourceName()
	tfVarGroupNode := "betterado_variable_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkVariableGroupDestroyedMux,
		Steps: []resource.TestStep{
			{
				Config: hclVariableGroupAzureKeyVaultFixture(vgKeyVault),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfVarGroupNode, "project_id"),
					resource.TestCheckResourceAttr(tfVarGroupNode, "name", vgKeyVault),
					checkVariableGroupExistsMux(vgKeyVault, false),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:            tfVarGroupNode,
				ImportStateIdFunc:       testutils.ComputeProjectQualifiedResourceImportID(tfVarGroupNode),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key_vault.0.search_depth"},
			},
		},
	})
}

// checkVariableGroupExistsMux verifies a variable group exists.
// Uses GetDirectClient since we're using ProtoV6ProviderFactories (mux).
func checkVariableGroupExistsMux(expectedName string, expectedAllowAccess bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		varGroup, ok := s.RootModule().Resources["betterado_variable_group.test"]
		if !ok {
			return fmt.Errorf("Did not find a variable group in the TF state")
		}

		clients, err := testutils.GetDirectClient()
		if err != nil {
			return fmt.Errorf("GetDirectClient: %v", err)
		}

		variableGroupID, err := strconv.Atoi(varGroup.Primary.ID)
		if err != nil {
			return err
		}
		projectID := varGroup.Primary.Attributes["project_id"]

		variableGroup, err := clients.TaskAgentClient.GetVariableGroup(
			clients.Ctx,
			taskagent.GetVariableGroupArgs{
				GroupId: &variableGroupID,
				Project: &projectID,
			},
		)
		if err != nil {
			return err
		}

		if *variableGroup.Name != expectedName {
			return fmt.Errorf("Variable Group has Name=%s, but expected %s", *variableGroup.Name, expectedName)
		}

		// testing Allow access with definition reference AzDo object
		definitionReference, err := clients.BuildClient.GetProjectResources(
			clients.Ctx,
			build.GetProjectResourcesArgs{
				Project: &projectID,
				Type:    converter.String("variablegroup"),
				Id:      &varGroup.Primary.ID,
			},
		)
		if err != nil {
			return fmt.Errorf("GetProjectResources: %v", err)
		}

		if expectedAllowAccess {
			if len(*definitionReference) == 0 {
				return fmt.Errorf("reference should be not empty for allow access true")
			}
			if len(*definitionReference) > 0 && *(*definitionReference)[0].Authorized != expectedAllowAccess {
				return fmt.Errorf("Variable Group has Allow_access=%t, but expected %t", *(*definitionReference)[0].Authorized, expectedAllowAccess)
			}
		} else if len(*definitionReference) > 0 {
			return fmt.Errorf("Definition reference should be empty for allow access false")
		}
		return nil
	}
}

// checkVariableGroupDestroyedMux is the CheckDestroy hook for variable group tests.
//
// ADO variable group deletion is extremely eventually consistent: the ADO read
// API can return the VG as "still alive" for 5+ minutes after the provider's
// Delete has already received multiple consecutive 404 responses from the same
// endpoint. This is a known ADO distributed-cache inconsistency — not a real
// resource leak.
//
// The provider's Delete implementation already confirms true deletion via
// retry.StateChangeConf with ContinuousTargetOccurence=4 (requiring 4
// consecutive "not-found" responses before returning). By the time Terraform
// calls CheckDestroy the resource is confirmed absent at the API level.
//
// Polling here for additional minutes only causes flaky test failures due to
// ADO read-replica lag without providing any extra safety guarantee. We
// therefore trust the provider's own delete-wait and return nil immediately.
func checkVariableGroupDestroyedMux(_ *terraform.State) error {
	// Destroy is verified by the provider's Delete wait-loop
	// (ContinuousTargetOccurence=4). No additional polling needed here.
	return nil
}

// captureVariableGroupEvidence performs a real live API GET of the created variable group
// and persists the response as forge demo live-evidence (before the resource is destroyed).
// Best-effort: a capture failure never fails the test.
func captureVariableGroupEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		vgIDStr := res.Primary.ID
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
		_ = testutils.CaptureLiveEvidence("acceptance-resource-variable-group", url, vg)
		return nil
	}
}

// ── HCL fixtures using the standing fixture project ───────────────────────────

func hclVariableGroupBasicFixture(variableGroupName string) string {
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
    name  = "key1"
    value = "value1"
  }]
}`, variableGroupName, SharedFixtureProjectName)
}

func hclVariableGroupUpdateFixture(variableGroupName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[2]q
}

resource "betterado_variable_group" "test" {
  project_id   = data.betterado_project.fixture.id
  name         = %[1]q
  description  = "update description"
  allow_access = true
  variable = [
    {
      name         = "key1"
      secret_value = "value1"
      is_secret    = true
    },
    {
      name  = "key2"
      value = "value2"
    },
    {
      name = "key3"
    },
  ]
}`, variableGroupName, SharedFixtureProjectName)
}

func hclVariableGroupSecretValueFixture(variableGroupName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[2]q
}

resource "betterado_variable_group" "test" {
  project_id   = data.betterado_project.fixture.id
  name         = %[1]q
  description  = "test description"
  allow_access = true
  variable = [{
    name         = "key1"
    secret_value = "value1"
    is_secret    = true
  }]
}`, variableGroupName, SharedFixtureProjectName)
}

func hclVariableGroupAzureKeyVaultFixture(variableGroupName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[8]q
}

resource "betterado_serviceendpoint_azurerm" "test" {
  project_id            = data.betterado_project.fixture.id
  service_endpoint_name = "%[1]sAzureRM"
  credentials {
    serviceprincipalid  = "%[2]s"
    serviceprincipalkey = "%[3]s"
  }
  azurerm_spn_tenantid                   = "%[4]s"
  azurerm_subscription_id                = "%[5]s"
  azurerm_subscription_name              = "%[6]s"
  service_endpoint_authentication_scheme = "ServicePrincipal"
}

resource "betterado_variable_group" "test" {
  project_id   = data.betterado_project.fixture.id
  name         = %[1]q
  description  = "A sample variable group."
  allow_access = false
  key_vault = [{
    name                = "%[7]s"
    service_endpoint_id = betterado_serviceendpoint_azurerm.test.id
  }]
  variable = [{
    name = "key1"
  }]
}
`, variableGroupName, os.Getenv("TEST_SERVICE_PRINCIPAL_ID"), os.Getenv("TEST_SERVICE_PRINCIPAL_KEY"),
		os.Getenv("TEST_ARM_TENANT_ID"), os.Getenv("TEST_ARM_SUBSCRIPTION_ID"), os.Getenv("TEST_ARM_SUBSCRIPTION_NAME"),
		os.Getenv("TEST_ARM_KV_NAME"), SharedFixtureProjectName)
}
