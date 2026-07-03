package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccIdentityDataSources_Framework is the top-level test that groups the three
// framework-migrated identity data sources under a single test name so the WI-6
// quality gate can run them all with:
//
//	go test -tags all -run TestAccIdentityDataSources_Framework ./azuredevops/internal/acceptancetests/
func TestAccIdentityDataSources_Framework(t *testing.T) {
	t.Run("IdentityGroup", TestAccIdentityGroupDataSource_Framework_Read)
	t.Run("IdentityGroups", TestAccIdentityGroupsDataSource_Framework_Read)
	t.Run("IdentityUser", TestAccIdentityUserDataSource_Framework_Read)
}

// TestAccIdentityGroupDataSource_Framework_Read verifies that data.betterado_identity_group
// populates descriptor and subject_descriptor when given a group name and project_id,
// using the muxed (framework) provider. The idempotency re-plan must be clean.
//
// Uses the persistent shared project (betterado-standing-demo) — the ADO org is
// at the 1000-project cap so resource "betterado_project" creates fail.
func TestAccIdentityGroupDataSource_Framework_Read(t *testing.T) {
	tfNode := "data.betterado_identity_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclIdentityGroupFrameworkRead(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "subject_descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclIdentityGroupFrameworkRead(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccIdentityGroupsDataSource_Framework_Read verifies that data.betterado_identity_groups
// populates the groups set (each entry has id, name, descriptor, subject_descriptor)
// when given a project_id, using the muxed (framework) provider.
//
// Uses the persistent shared project (betterado-standing-demo).
func TestAccIdentityGroupsDataSource_Framework_Read(t *testing.T) {
	tfNode := "data.betterado_identity_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclIdentityGroupsFrameworkRead(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "groups.#"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclIdentityGroupsFrameworkRead(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccIdentityUserDataSource_Framework_Read verifies that data.betterado_identity_user
// populates descriptor and subject_descriptor when given a name (and optional search_filter),
// using the muxed (framework) provider.
//
// Uses "Project Collection Build Service ({OrgName})" — the collection-level ADO build
// service identity that is always present in every organisation, regardless of whether
// any pipelines have been run. The org name is extracted at apply time from
// betterado_client_config.organization_url so no hard-coded org name is needed.
func TestAccIdentityUserDataSource_Framework_Read(t *testing.T) {
	tfNode := "data.betterado_identity_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclIdentityUserFrameworkRead(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "subject_descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclIdentityUserFrameworkRead(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclIdentityGroupFrameworkRead looks up "Build Administrators" in the shared project.
// The identity group name uses the ADO format [ProjectName]\GroupName.
// SharedFixtureProjectName is "betterado-standing-demo" (a constant), so the name
// is constructed in Go before being embedded as a literal HCL string.
func hclIdentityGroupFrameworkRead() string {
	groupName := fmt.Sprintf("[%s]\\Build Administrators", SharedFixtureProjectName)
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

// hclIdentityGroupsFrameworkRead lists all identity groups in the shared project.
func hclIdentityGroupsFrameworkRead() string {
	return fmt.Sprintf(`
data "betterado_project" "shared" {
  name = %[1]q
}

data "betterado_identity_groups" "test" {
  project_id = data.betterado_project.shared.id
}
`, SharedFixtureProjectName)
}

// hclIdentityUserFrameworkRead looks up the "Project Collection Build Service ({OrgName})"
// identity user by DisplayName. This is the collection-level build service account that is
// a permanent system identity present in every ADO organisation, so it is always findable
// without creating new projects or user entitlements. The org name is extracted at apply
// time from betterado_client_config.organization_url.
func hclIdentityUserFrameworkRead() string {
	return `
data "betterado_client_config" "current" {}

data "betterado_identity_user" "test" {
  name          = "Project Collection Build Service (${compact(split("/", data.betterado_client_config.current.organization_url))[2]})"
  search_filter = "DisplayName"
}
`
}
