package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccGraphSimpleDataSources_Framework is the top-level test that groups
// the four framework-migrated graph data sources under a single test name
// so the WI-4 quality gate can run them all with:
//
//	go test -tags all -run TestAccGraphSimpleDataSources_Framework ./azuredevops/internal/acceptancetests/
func TestAccGraphSimpleDataSources_Framework(t *testing.T) {
	t.Run("Descriptor", TestAccDescriptorDataSource_Framework_Read)
	t.Run("StorageKey", TestAccStorageKeyDataSource_Framework_Read)
	t.Run("Group", TestAccGroupDataSource_Framework_Read)
	t.Run("GroupMembership", TestAccGroupMembershipDataSource_Framework_Read)
}

// TestAccDescriptorDataSource_Framework_Read verifies that data.betterado_descriptor
// populates the descriptor attribute when given a project's storage key UUID,
// using the muxed (framework) provider. The idempotency re-plan must be clean.
func TestAccDescriptorDataSource_Framework_Read(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	tfNode := "data.betterado_descriptor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclDescriptorFrameworkRead(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "storage_key"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclDescriptorFrameworkRead(projectName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccStorageKeyDataSource_Framework_Read verifies that data.betterado_storage_key
// populates the storage_key attribute when given a descriptor, using the muxed provider.
func TestAccStorageKeyDataSource_Framework_Read(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	tfNode := "data.betterado_storage_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclStorageKeyFrameworkRead(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "storage_key"),
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclStorageKeyFrameworkRead(projectName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccGroupDataSource_Framework_Read verifies that data.betterado_group
// populates all attributes (descriptor, origin, origin_id, group_id) when given
// a group name and project_id, using the muxed provider.
func TestAccGroupDataSource_Framework_Read(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	groupName := "Build Administrators"
	tfNode := "data.betterado_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclGroupFrameworkRead(projectName, groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "origin"),
					resource.TestCheckResourceAttrSet(tfNode, "origin_id"),
					resource.TestCheckResourceAttrSet(tfNode, "group_id"),
					resource.TestCheckResourceAttr(tfNode, "name", groupName),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclGroupFrameworkRead(projectName, groupName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccGroupMembershipDataSource_Framework_Read verifies that data.betterado_group_membership
// populates the members list when given a group descriptor, using the muxed provider.
func TestAccGroupMembershipDataSource_Framework_Read(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	groupName := testutils.GenerateResourceName()
	tfNode := "data.betterado_group_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclGroupMembershipFrameworkRead(projectName, groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "members.#", "2"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclGroupMembershipFrameworkRead(projectName, groupName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclDescriptorFrameworkRead builds a config that creates a project and then
// reads its descriptor via data.betterado_descriptor (framework provider).
func hclDescriptorFrameworkRead(projectName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = %[1]q
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

data "betterado_descriptor" "test" {
  storage_key = betterado_project.test.id
}
`, projectName)
}

// hclStorageKeyFrameworkRead builds a config that creates a project, reads its
// descriptor, then resolves the storage key back via data.betterado_storage_key.
func hclStorageKeyFrameworkRead(projectName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = %[1]q
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

data "betterado_descriptor" "project" {
  storage_key = betterado_project.test.id
}

data "betterado_storage_key" "test" {
  descriptor = data.betterado_descriptor.project.descriptor
}
`, projectName)
}

// hclGroupFrameworkRead builds a config that creates a project and looks up one
// of its built-in groups via data.betterado_group (framework provider).
func hclGroupFrameworkRead(projectName, groupName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = %[1]q
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

data "betterado_group" "test" {
  project_id = betterado_project.test.id
  name       = %[2]q
}
`, projectName, groupName)
}

// hclGroupMembershipFrameworkRead creates a group with two members, then reads
// the membership back via data.betterado_group_membership (framework provider).
func hclGroupMembershipFrameworkRead(projectName, groupName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = %[1]q
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

data "betterado_group" "admin" {
  project_id = betterado_project.test.id
  name       = "Build Administrators"
}

data "betterado_group" "contributors" {
  project_id = betterado_project.test.id
  name       = "Contributors"
}

resource "betterado_group" "test" {
  scope        = betterado_project.test.id
  display_name = %[2]q

  members = [
    data.betterado_group.admin.descriptor,
    data.betterado_group.contributors.descriptor,
  ]
}

data "betterado_group_membership" "test" {
  group_descriptor = betterado_group.test.descriptor
}
`, projectName, groupName)
}
