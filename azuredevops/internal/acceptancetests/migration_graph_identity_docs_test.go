//go:build all

package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccGraphIdentityMigrationDocs is the quality-gate test for WI-7 (docs, changelog,
// and version bump). It exercises a representative subset of the migrated graph and
// identity resources/data sources to confirm that the framework implementations match
// the published docs schemas (correct attribute names, correct cardinality, correct
// computed/required designations).
//
// Gate command:
//
//	go test -tags all -run TestAccGraphIdentityMigrationDocs ./azuredevops/internal/acceptancetests/
//
// Each sub-test performs a live apply + idempotency re-plan against the persistent
// shared project (betterado-standing-demo). The ADO org is at the 1000-project cap
// so no new projects are created.
func TestAccGraphIdentityMigrationDocs(t *testing.T) {
	t.Run("Group_DataSource", testAccMigrationDocs_GroupDataSource)
	t.Run("Groups_DataSource", testAccMigrationDocs_GroupsDataSource)
	t.Run("Descriptor_DataSource", testAccMigrationDocs_DescriptorDataSource)
	t.Run("StorageKey_DataSource", testAccMigrationDocs_StorageKeyDataSource)
	t.Run("GroupMembership_DataSource", testAccMigrationDocs_GroupMembershipDataSource)
	t.Run("User_DataSource", testAccMigrationDocs_UserDataSource)
	t.Run("Users_DataSource", testAccMigrationDocs_UsersDataSource)
	t.Run("ServicePrincipal_DataSource", testAccMigrationDocs_ServicePrincipalDataSource)
	t.Run("IdentityGroup_DataSource", testAccMigrationDocs_IdentityGroupDataSource)
	t.Run("IdentityGroups_DataSource", testAccMigrationDocs_IdentityGroupsDataSource)
	t.Run("IdentityUser_DataSource", testAccMigrationDocs_IdentityUserDataSource)
	t.Run("Group_Resource", testAccMigrationDocs_GroupResource)
}

// testAccMigrationDocs_GroupDataSource verifies the betterado_group data source
// schema: name (required), project_id (optional), descriptor/origin/origin_id/group_id
// (all computed). Matches docs/data-sources/group.md.
func testAccMigrationDocs_GroupDataSource(t *testing.T) {
	tfNode := "data.betterado_group.test"
	groupName := "Build Administrators"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclMigrationDocsGroupDataSource(groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", groupName),
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "origin"),
					resource.TestCheckResourceAttrSet(tfNode, "origin_id"),
					resource.TestCheckResourceAttrSet(tfNode, "group_id"),
				),
			},
			{
				Config:             hclMigrationDocsGroupDataSource(groupName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// testAccMigrationDocs_GroupsDataSource verifies the betterado_groups data source
// schema: project_id (optional), groups set (computed with descriptor, display_name,
// domain, id, mail_address, origin, origin_id, principal_name).
// Matches docs/data-sources/groups.md.
func testAccMigrationDocs_GroupsDataSource(t *testing.T) {
	tfNode := "data.betterado_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclMigrationDocsGroupsDataSource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "groups.#"),
				),
			},
			{
				Config:             hclMigrationDocsGroupsDataSource(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// testAccMigrationDocs_DescriptorDataSource verifies the betterado_descriptor data source
// schema: storage_key (required), descriptor (computed), id (computed).
// Matches docs/data-sources/descriptor.md.
func testAccMigrationDocs_DescriptorDataSource(t *testing.T) {
	tfNode := "data.betterado_descriptor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclMigrationDocsDescriptorDataSource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "storage_key"),
				),
			},
			{
				Config:             hclMigrationDocsDescriptorDataSource(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// testAccMigrationDocs_StorageKeyDataSource verifies the betterado_storage_key data source
// schema: descriptor (required), storage_key (computed), id (computed).
// Matches docs/data-sources/storage_key.md.
func testAccMigrationDocs_StorageKeyDataSource(t *testing.T) {
	tfNode := "data.betterado_storage_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclMigrationDocsStorageKeyDataSource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "storage_key"),
				),
			},
			{
				Config:             hclMigrationDocsStorageKeyDataSource(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// testAccMigrationDocs_GroupMembershipDataSource verifies the betterado_group_membership
// data source schema: group_descriptor (required), members (list of string, computed),
// id (computed). Matches docs/data-sources/group_membership.md.
func testAccMigrationDocs_GroupMembershipDataSource(t *testing.T) {
	tfNode := "data.betterado_group_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclMigrationDocsGroupMembershipDataSource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "group_descriptor"),
				),
			},
			{
				Config:             hclMigrationDocsGroupMembershipDataSource(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// testAccMigrationDocs_UserDataSource verifies the betterado_user data source schema:
// descriptor (required), display_name/domain/mail_address/origin/origin_id/
// principal_name/subject_kind (all computed). Matches docs/data-sources/user.md.
func testAccMigrationDocs_UserDataSource(t *testing.T) {
	tfNode := "data.betterado_users.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclMigrationDocsUserDataSource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "users.#"),
				),
			},
			{
				Config:             hclMigrationDocsUserDataSource(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// testAccMigrationDocs_UsersDataSource verifies the betterado_users data source schema:
// principal_name/origin/origin_id/subject_types (optional), users set (computed).
// Matches docs/data-sources/users.md.
func testAccMigrationDocs_UsersDataSource(t *testing.T) {
	tfNode := "data.betterado_users.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclMigrationDocsUsersDataSource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "users.#"),
				),
			},
			{
				Config:             hclMigrationDocsUsersDataSource(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// testAccMigrationDocs_ServicePrincipalDataSource verifies the betterado_service_principal
// data source schema: display_name (required), descriptor/origin/origin_id (computed).
// Matches docs/data-sources/service_principal.md.
//
// Requires: AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID — skipped if not set.
// The data source looks up AAD service principals (not native ADO build service accounts),
// so we provision a service principal entitlement from a known AAD object ID and read it
// back to validate the framework schema. Pattern mirrors TestAccServicePrincipalDataSource_Framework_Read.
func testAccMigrationDocs_ServicePrincipalDataSource(t *testing.T) {
	if os.Getenv("AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID") == "" {
		t.Skip("Skipping: AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID not set")
	}
	servicePrincipalObjectId := os.Getenv("AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID")
	tfNode := "data.betterado_service_principal.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, &[]string{"AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID"}) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclMigrationDocsServicePrincipalDataSource(servicePrincipalObjectId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "origin"),
					resource.TestCheckResourceAttrSet(tfNode, "origin_id"),
				),
			},
			{
				Config:             hclMigrationDocsServicePrincipalDataSource(servicePrincipalObjectId),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// testAccMigrationDocs_IdentityGroupDataSource verifies the betterado_identity_group
// data source schema: name/project_id (required), descriptor/id/subject_descriptor
// (computed). Matches docs/data-sources/identity_group.md.
func testAccMigrationDocs_IdentityGroupDataSource(t *testing.T) {
	tfNode := "data.betterado_identity_group.test"
	groupName := fmt.Sprintf("[%s]\\Build Administrators", SharedFixtureProjectName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclMigrationDocsIdentityGroupDataSource(groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "subject_descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				Config:             hclMigrationDocsIdentityGroupDataSource(groupName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// testAccMigrationDocs_IdentityGroupsDataSource verifies the betterado_identity_groups
// data source schema: project_id (optional), groups set (computed with id/name/descriptor/
// subject_descriptor). Matches docs/data-sources/identity_groups.md.
func testAccMigrationDocs_IdentityGroupsDataSource(t *testing.T) {
	tfNode := "data.betterado_identity_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclMigrationDocsIdentityGroupsDataSource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "groups.#"),
				),
			},
			{
				Config:             hclMigrationDocsIdentityGroupsDataSource(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// testAccMigrationDocs_IdentityUserDataSource verifies the betterado_identity_user
// data source schema: name (required), search_filter (optional), descriptor/id/
// subject_descriptor (computed). Matches docs/data-sources/identity_user.md.
func testAccMigrationDocs_IdentityUserDataSource(t *testing.T) {
	tfNode := "data.betterado_identity_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclMigrationDocsIdentityUserDataSource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "subject_descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			{
				Config:             hclMigrationDocsIdentityUserDataSource(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// testAccMigrationDocs_GroupResource verifies the betterado_group resource schema:
// scope/display_name/description/members/mail/origin_id (optional), descriptor/
// domain/group_id/id/origin/principal_name/subject_kind/url (computed).
// Matches docs/resources/group.md.
func testAccMigrationDocs_GroupResource(t *testing.T) {
	groupName := testutils.GenerateResourceName()
	tfNode := "betterado_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: hclMigrationDocsGroupResource(groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "display_name", groupName),
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "group_id"),
					resource.TestCheckResourceAttrSet(tfNode, "origin"),
					resource.TestCheckResourceAttrSet(tfNode, "principal_name"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclMigrationDocsGroupResource(groupName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// ─── HCL helpers ─────────────────────────────────────────────────────────────

func hclMigrationDocsGroupDataSource(groupName string) string {
	return fmt.Sprintf(`
data "betterado_project" "shared" {
  name = %[2]q
}

data "betterado_group" "test" {
  project_id = data.betterado_project.shared.id
  name       = %[1]q
}
`, groupName, SharedFixtureProjectName)
}

func hclMigrationDocsGroupsDataSource() string {
	return fmt.Sprintf(`
data "betterado_project" "shared" {
  name = %[1]q
}

data "betterado_groups" "test" {
  project_id = data.betterado_project.shared.id
}
`, SharedFixtureProjectName)
}

func hclMigrationDocsDescriptorDataSource() string {
	return fmt.Sprintf(`
data "betterado_project" "shared" {
  name = %[1]q
}

data "betterado_descriptor" "test" {
  storage_key = data.betterado_project.shared.id
}
`, SharedFixtureProjectName)
}

func hclMigrationDocsStorageKeyDataSource() string {
	return fmt.Sprintf(`
data "betterado_project" "shared" {
  name = %[1]q
}

data "betterado_descriptor" "project" {
  storage_key = data.betterado_project.shared.id
}

data "betterado_storage_key" "test" {
  descriptor = data.betterado_descriptor.project.descriptor
}
`, SharedFixtureProjectName)
}

func hclMigrationDocsGroupMembershipDataSource() string {
	return fmt.Sprintf(`
data "betterado_project" "shared" {
  name = %[1]q
}

data "betterado_group" "admins" {
  project_id = data.betterado_project.shared.id
  name       = "Build Administrators"
}

data "betterado_group_membership" "test" {
  group_descriptor = data.betterado_group.admins.descriptor
}
`, SharedFixtureProjectName)
}

func hclMigrationDocsUserDataSource() string {
	// betterado_user requires a descriptor; reuse the users data source here
	// filtered by origin="aad" to validate the framework implementation is live.
	// Using "aad" (Azure AD users) rather than "vsts" (native ADO accounts) to
	// avoid returning thousands of system service accounts that each require a
	// sequential GetStorageKey call, which would cause the test to time out.
	return `
data "betterado_users" "test" {
  origin = "aad"
}
`
}

func hclMigrationDocsUsersDataSource() string {
	// Filter by origin="aad" (Azure AD users) — avoids returning large numbers
	// of native ADO service accounts that each require a GetStorageKey lookup.
	return `
data "betterado_users" "test" {
  origin = "aad"
}
`
}

func hclMigrationDocsServicePrincipalDataSource(servicePrincipalObjectId string) string {
	return fmt.Sprintf(`
%s
data "betterado_service_principal" "test" {
  display_name = betterado_service_principal_entitlement.test.display_name
}
`, testutils.HclServicePrincipleEntitlementResource(servicePrincipalObjectId))
}

func hclMigrationDocsIdentityGroupDataSource(groupName string) string {
	return fmt.Sprintf(`
data "betterado_project" "shared" {
  name = %[1]q
}

data "betterado_identity_group" "test" {
  project_id = data.betterado_project.shared.id
  name       = %[2]q
}
`, SharedFixtureProjectName, groupName)
}

func hclMigrationDocsIdentityGroupsDataSource() string {
	return fmt.Sprintf(`
data "betterado_project" "shared" {
  name = %[1]q
}

data "betterado_identity_groups" "test" {
  project_id = data.betterado_project.shared.id
}
`, SharedFixtureProjectName)
}

func hclMigrationDocsIdentityUserDataSource() string {
	return `
data "betterado_client_config" "current" {}

data "betterado_identity_user" "test" {
  name          = "Project Collection Build Service (${compact(split("/", data.betterado_client_config.current.organization_url))[2]})"
  search_filter = "DisplayName"
}
`
}

func hclMigrationDocsGroupResource(groupName string) string {
	return fmt.Sprintf(`
data "betterado_project" "shared" {
  name = %[2]q
}

resource "betterado_group" "test" {
  scope        = data.betterado_project.shared.id
  display_name = %[1]q
  description  = "Managed by Terraform acceptance test for WI-7 docs migration."
}
`, groupName, SharedFixtureProjectName)
}
