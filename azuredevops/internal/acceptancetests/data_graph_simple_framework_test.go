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
//
// Uses the persistent shared project (betterado-standing-demo) — the ADO org is
// at the 1000-project cap so resource "betterado_project" creates fail.
func TestAccDescriptorDataSource_Framework_Read(t *testing.T) {
	tfNode := "data.betterado_descriptor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclDescriptorFrameworkRead(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					resource.TestCheckResourceAttrSet(tfNode, "storage_key"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclDescriptorFrameworkRead(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccStorageKeyDataSource_Framework_Read verifies that data.betterado_storage_key
// populates the storage_key attribute when given a descriptor, using the muxed provider.
//
// Uses the persistent shared project (betterado-standing-demo) — the ADO org is
// at the 1000-project cap so resource "betterado_project" creates fail.
func TestAccStorageKeyDataSource_Framework_Read(t *testing.T) {
	tfNode := "data.betterado_storage_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclStorageKeyFrameworkRead(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "storage_key"),
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclStorageKeyFrameworkRead(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccGroupDataSource_Framework_Read verifies that data.betterado_group
// populates all attributes (descriptor, origin, origin_id, group_id) when given
// a group name and project_id, using the muxed provider.
//
// Uses the persistent shared project (betterado-standing-demo) — the ADO org is
// at the 1000-project cap so resource "betterado_project" creates fail.
func TestAccGroupDataSource_Framework_Read(t *testing.T) {
	groupName := "Build Administrators"
	tfNode := "data.betterado_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclGroupFrameworkRead(groupName),
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
				Config:             hclGroupFrameworkRead(groupName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccGroupMembershipDataSource_Framework_Read verifies that data.betterado_group_membership
// populates the members list when given a group descriptor, using the muxed provider.
//
// Uses the persistent shared project (betterado-standing-demo) — the ADO org is
// at the 1000-project cap so resource "betterado_project" creates fail.
func TestAccGroupMembershipDataSource_Framework_Read(t *testing.T) {
	groupName := testutils.GenerateResourceName()
	tfNode := "data.betterado_group_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclGroupMembershipFrameworkRead(groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "members.#", "2"),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclGroupMembershipFrameworkRead(groupName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclDescriptorFrameworkRead builds a config that looks up the shared project and
// reads its descriptor via data.betterado_descriptor (framework provider).
// Uses the persistent shared project — the ADO org is at the 1000-project cap.
func hclDescriptorFrameworkRead() string {
	return fmt.Sprintf(`
data "betterado_project" "shared" {
  name = %[1]q
}

data "betterado_descriptor" "test" {
  storage_key = data.betterado_project.shared.id
}
`, SharedFixtureProjectName)
}

// hclStorageKeyFrameworkRead builds a config that looks up the shared project,
// reads its descriptor, then resolves the storage key back via data.betterado_storage_key.
// Uses the persistent shared project — the ADO org is at the 1000-project cap.
func hclStorageKeyFrameworkRead() string {
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

// hclGroupFrameworkRead builds a config that looks up the shared project and then
// looks up one of its built-in groups via data.betterado_group (framework provider).
// Uses the persistent shared project — the ADO org is at the 1000-project cap.
func hclGroupFrameworkRead(groupName string) string {
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

// hclGroupMembershipFrameworkRead creates a group (in the shared project) with two members,
// then reads the membership back via data.betterado_group_membership (framework provider).
// Uses the persistent shared project — the ADO org is at the 1000-project cap.
func hclGroupMembershipFrameworkRead(groupName string) string {
	return fmt.Sprintf(`
data "betterado_project" "shared" {
  name = %[2]q
}

data "betterado_group" "admin" {
  project_id = data.betterado_project.shared.id
  name       = "Build Administrators"
}

data "betterado_group" "contributors" {
  project_id = data.betterado_project.shared.id
  name       = "Contributors"
}

resource "betterado_group" "test" {
  scope        = data.betterado_project.shared.id
  display_name = %[1]q

  members = [
    data.betterado_group.admin.descriptor,
    data.betterado_group.contributors.descriptor,
  ]
}

data "betterado_group_membership" "test" {
  group_descriptor = betterado_group.test.descriptor
}
`, groupName, SharedFixtureProjectName)
}
