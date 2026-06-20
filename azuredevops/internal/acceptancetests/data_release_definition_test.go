package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccDataReleaseDefinition_ById creates a minimal release definition resource
// and looks it up via the betterado_release_definition data source using its numeric ID.
// It verifies that name and path attributes are resolved, and that a re-plan
// produces no diff (idempotency).
func TestAccDataReleaseDefinition_ById(t *testing.T) {
	name := testutils.GenerateResourceName()
	fixture := SharedReleaseFixture(t)
	tfDataNode := "data.betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclDataReleaseDefinitionById(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(tfDataNode, "name",
						"betterado_release_definition.test", "name"),
					resource.TestCheckResourceAttrPair(tfDataNode, "path",
						"betterado_release_definition.test", "path"),
					resource.TestCheckResourceAttrSet(tfDataNode, "id"),
				),
			},
			// Idempotency: a re-plan must produce no diff.
			{
				Config:             hclDataReleaseDefinitionById(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccDataReleaseDefinition_ByName creates a minimal release definition resource
// and looks it up via the betterado_release_definition data source using its name.
// It verifies that id and other attributes are resolved correctly, and that a
// re-plan produces no diff.
func TestAccDataReleaseDefinition_ByName(t *testing.T) {
	name := testutils.GenerateResourceName()
	fixture := SharedReleaseFixture(t)
	tfDataNode := "data.betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclDataReleaseDefinitionByName(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(tfDataNode, "id",
						"betterado_release_definition.test", "id"),
					resource.TestCheckResourceAttrPair(tfDataNode, "name",
						"betterado_release_definition.test", "name"),
					resource.TestCheckResourceAttrPair(tfDataNode, "path",
						"betterado_release_definition.test", "path"),
				),
			},
			// Idempotency: a re-plan must produce no diff.
			{
				Config:             hclDataReleaseDefinitionByName(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccDataReleaseDefinitions_List creates a minimal release definition resource
// and reads all release definitions in that project via the betterado_release_definitions
// list data source. It asserts the list is non-empty and that a re-plan produces no diff.
func TestAccDataReleaseDefinitions_List(t *testing.T) {
	name := testutils.GenerateResourceName()
	fixture := SharedReleaseFixture(t)
	tfDataNode := "data.betterado_release_definitions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclDataReleaseDefinitionsList(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					// The list reads every definition in the (shared) project, so
					// assert it is non-empty rather than an exact count.
					resource.TestCheckResourceAttrSet(tfDataNode, "release_definitions.0.id"),
				),
			},
			// Idempotency: a re-plan must produce no diff.
			{
				Config:             hclDataReleaseDefinitionsList(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// --- HCL fixtures ---

// hclDataReleaseDefinitionById builds a Terraform config that creates a project +
// minimal release definition and then reads it back using release_definition_id.
func hclDataReleaseDefinitionById(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
%s

data "betterado_release_definition" "test" {
  project_id            = betterado_release_definition.test.project_id
  release_definition_id = tonumber(betterado_release_definition.test.id)
}
`, hclReleaseDefinitionBasic(name, fixture))
}

// hclDataReleaseDefinitionByName builds a Terraform config that creates a project +
// minimal release definition and then reads it back using the name attribute.
func hclDataReleaseDefinitionByName(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
%s

data "betterado_release_definition" "test" {
  project_id = betterado_release_definition.test.project_id
  name       = betterado_release_definition.test.name
}
`, hclReleaseDefinitionBasic(name, fixture))
}

// hclDataReleaseDefinitionsList builds a Terraform config that creates a project +
// minimal release definition and then lists all definitions in the project.
func hclDataReleaseDefinitionsList(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
%s

data "betterado_release_definitions" "test" {
  project_id = betterado_release_definition.test.project_id
}
`, hclReleaseDefinitionBasic(name, fixture))
}
