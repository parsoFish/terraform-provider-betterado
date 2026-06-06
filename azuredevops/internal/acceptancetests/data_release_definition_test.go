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
	tfDataNode := "data.betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclDataReleaseDefinitionById(name),
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
				Config:             hclDataReleaseDefinitionById(name),
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
	tfDataNode := "data.betterado_release_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclDataReleaseDefinitionByName(name),
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
				Config:             hclDataReleaseDefinitionByName(name),
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
	tfDataNode := "data.betterado_release_definitions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclDataReleaseDefinitionsList(name),
				Check: resource.ComposeTestCheckFunc(
					// At least the definition created in the fixture must appear.
					resource.TestCheckResourceAttr(tfDataNode, "release_definitions.#", "1"),
				),
			},
			// Idempotency: a re-plan must produce no diff.
			{
				Config:             hclDataReleaseDefinitionsList(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// --- HCL fixtures ---

// hclDataReleaseDefinitionById builds a Terraform config that creates a project +
// minimal release definition and then reads it back using release_definition_id.
func hclDataReleaseDefinitionById(name string) string {
	return fmt.Sprintf(`
%s

data "betterado_release_definition" "test" {
  project_id            = betterado_release_definition.test.project_id
  release_definition_id = tonumber(betterado_release_definition.test.id)
}
`, hclReleaseDefinitionBasic(name))
}

// hclDataReleaseDefinitionByName builds a Terraform config that creates a project +
// minimal release definition and then reads it back using the name attribute.
func hclDataReleaseDefinitionByName(name string) string {
	return fmt.Sprintf(`
%s

data "betterado_release_definition" "test" {
  project_id = betterado_release_definition.test.project_id
  name       = betterado_release_definition.test.name
}
`, hclReleaseDefinitionBasic(name))
}

// hclDataReleaseDefinitionsList builds a Terraform config that creates a project +
// minimal release definition and then lists all definitions in the project.
func hclDataReleaseDefinitionsList(name string) string {
	return fmt.Sprintf(`
%s

data "betterado_release_definitions" "test" {
  project_id = betterado_release_definition.test.project_id
}
`, hclReleaseDefinitionBasic(name))
}
