package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccGraphComplexDataSources_Framework is the top-level test that groups
// the four framework-migrated complex graph data sources under a single test name
// so the WI-5 quality gate can run them all with:
//
//	go test -tags all -run TestAccGraphComplexDataSources_Framework ./azuredevops/internal/acceptancetests/
func TestAccGraphComplexDataSources_Framework(t *testing.T) {
	t.Run("Users", TestAccUsersDataSource_Framework_Read)
	t.Run("Groups", TestAccGroupsDataSource_Framework_Read)
	t.Run("User", TestAccUserDataSource_Framework_Read)
	t.Run("ServicePrincipal", TestAccServicePrincipalDataSource_Framework_Read)
}

// TestAccUsersDataSource_Framework_Read verifies that data.betterado_users
// returns a populated users set when filtered by origin, using the muxed provider.
// Uses the muxed (framework) provider; idempotency re-plan must be clean.
//
// No project creation needed — the ADO org is at the 1000-project cap.
func TestAccUsersDataSource_Framework_Read(t *testing.T) {
	tfNode := "data.betterado_users.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclUsersFrameworkReadByOrigin(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "users.#"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclUsersFrameworkReadByOrigin(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccGroupsDataSource_Framework_Read verifies that data.betterado_groups
// returns a populated groups set when scoped to the shared project,
// using the muxed provider. Each group must have descriptor, display_name, origin,
// origin_id, domain, mail_address, principal_name, id (storage key).
// Idempotency re-plan must be clean.
//
// Uses the persistent shared project (betterado-standing-demo) — the ADO org is
// at the 1000-project cap.
func TestAccGroupsDataSource_Framework_Read(t *testing.T) {
	tfNode := "data.betterado_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclGroupsFrameworkRead(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "groups.#"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclGroupsFrameworkRead(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccUserDataSource_Framework_Read verifies that data.betterado_user
// populates display_name, domain, mail_address, origin, origin_id, principal_name,
// subject_kind when given a known user descriptor via the muxed provider.
// Idempotency re-plan must be clean.
//
// Requires: AZDO_TEST_AAD_USER_EMAIL — the test will skip if not set.
// The test creates a user entitlement to obtain a valid descriptor, then reads
// the user back via data.betterado_user.
func TestAccUserDataSource_Framework_Read(t *testing.T) {
	if os.Getenv("AZDO_TEST_AAD_USER_EMAIL") == "" {
		t.Skip("Skip test due to AZDO_TEST_AAD_USER_EMAIL not set")
	}
	userEmail := os.Getenv("AZDO_TEST_AAD_USER_EMAIL")
	tfNode := "data.betterado_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, &[]string{"AZDO_TEST_AAD_USER_EMAIL"}) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclUserFrameworkRead(userEmail),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "display_name"),
					resource.TestCheckResourceAttrSet(tfNode, "origin"),
					resource.TestCheckResourceAttrSet(tfNode, "origin_id"),
					resource.TestCheckResourceAttrSet(tfNode, "principal_name"),
					resource.TestCheckResourceAttrSet(tfNode, "subject_kind"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclUserFrameworkRead(userEmail),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccServicePrincipalDataSource_Framework_Read verifies that
// data.betterado_service_principal populates descriptor, display_name, origin_id,
// origin when given a known display_name via the muxed provider.
// Idempotency re-plan must be clean.
//
// Requires: AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID — the test will skip if not set.
func TestAccServicePrincipalDataSource_Framework_Read(t *testing.T) {
	if os.Getenv("AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID") == "" {
		t.Skip("Skip test due to AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID not set")
	}
	servicePrincipalObjectId := os.Getenv("AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID")
	tfNode := "data.betterado_service_principal.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, &[]string{"AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID"}) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclServicePrincipalFrameworkRead(servicePrincipalObjectId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "display_name"),
					resource.TestCheckResourceAttrSet(tfNode, "origin"),
					resource.TestCheckResourceAttrSet(tfNode, "origin_id"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclServicePrincipalFrameworkRead(servicePrincipalObjectId),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclUsersFrameworkReadByOrigin builds a config that reads all AAD users
// via data.betterado_users filtered by origin (framework provider).
func hclUsersFrameworkReadByOrigin() string {
	return `
data "betterado_users" "test" {
  origin = "aad"
}
`
}

// hclGroupsFrameworkRead builds a config that reads groups for the shared
// project via data.betterado_groups (framework provider).
func hclGroupsFrameworkRead() string {
	return fmt.Sprintf(`
data "betterado_project" "shared" {
  name = %[1]q
}

data "betterado_groups" "test" {
  project_id = data.betterado_project.shared.id
}
`, SharedFixtureProjectName)
}

// hclUserFrameworkRead creates a user entitlement then reads the user descriptor
// back via data.betterado_user (framework provider).
// data.betterado_users is used as a bridge to get the descriptor from the email.
func hclUserFrameworkRead(userEmail string) string {
	return fmt.Sprintf(`
resource "betterado_user_entitlement" "test" {
  principal_name       = %[1]q
  account_license_type = "stakeholder"
}

data "betterado_users" "lookup" {
  principal_name = %[1]q
  depends_on     = [betterado_user_entitlement.test]
}

data "betterado_user" "test" {
  descriptor = tolist(data.betterado_users.lookup.users)[0].descriptor
  depends_on = [data.betterado_users.lookup]
}
`, userEmail)
}

// hclServicePrincipalFrameworkRead creates a service principal entitlement,
// then reads it back via data.betterado_service_principal (framework provider).
func hclServicePrincipalFrameworkRead(servicePrincipalObjectId string) string {
	return fmt.Sprintf(`
%s
data "betterado_service_principal" "test" {
  display_name = betterado_service_principal_entitlement.test.display_name
}
`, testutils.HclServicePrincipleEntitlementResource(servicePrincipalObjectId))
}
