package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccDataReleaseDefinitionRevision_Basic creates a minimal release definition
// and reads back a specific revision via data.betterado_release_definition_revision.
// Verifies that json_content is non-empty and that a re-plan produces no diff.
func TestAccDataReleaseDefinitionRevision_Basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	fixture := SharedReleaseFixture(t)
	tfDataNode := "data.betterado_release_definition_revision.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclDataReleaseDefinitionRevision(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfDataNode, "json_content"),
				),
			},
			// Idempotency.
			{
				Config:             hclDataReleaseDefinitionRevision(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccDataReleaseDefinitionHistory_Basic creates a minimal release definition
// and reads the full revision history via data.betterado_release_definition_history.
// Verifies at least one revision entry exists and that a re-plan produces no diff.
func TestAccDataReleaseDefinitionHistory_Basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	fixture := SharedReleaseFixture(t)
	tfDataNode := "data.betterado_release_definition_history.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclDataReleaseDefinitionHistory(name, fixture),
				Check: resource.ComposeTestCheckFunc(
					// At least one revision entry must exist after creation.
					resource.TestCheckResourceAttrSet(tfDataNode, "revisions.#"),
					resource.TestCheckResourceAttrSet(tfDataNode, "revisions.0.revision"),
				),
			},
			// Idempotency.
			{
				Config:             hclDataReleaseDefinitionHistory(name, fixture),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func hclDataReleaseDefinitionRevision(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
%s

data "betterado_release_definition_revision" "test" {
  project_id            = betterado_release_definition.test.project_id
  release_definition_id = tonumber(betterado_release_definition.test.id)
  revision              = betterado_release_definition.test.revision
}
`, hclReleaseDefinitionBasic(name, fixture))
}

func hclDataReleaseDefinitionHistory(name string, fixture SharedFixtureResult) string {
	return fmt.Sprintf(`
%s

data "betterado_release_definition_history" "test" {
  project_id            = betterado_release_definition.test.project_id
  release_definition_id = tonumber(betterado_release_definition.test.id)
}
`, hclReleaseDefinitionBasic(name, fixture))
}
