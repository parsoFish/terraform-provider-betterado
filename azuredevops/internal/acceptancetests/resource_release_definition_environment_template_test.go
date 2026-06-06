//go:build (all || resource_release_definition_environment_template) && !exclude_resource_release_definition_environment_template

package acceptancetests

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/utils/sdk/environmenttemplates"
)

// TestAccReleaseDefinitionEnvironmentTemplate_Basic tests creating, reading, idempotency, importing,
// and destroying a betterado_release_definition_environment_template resource.
func TestAccReleaseDefinitionEnvironmentTemplate_Basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_release_definition_environment_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkReleaseDefinitionEnvironmentTemplateDestroyed,
		Steps: []resource.TestStep{
			// Step 1 — apply: create the template and verify it exists in state and in ADO.
			{
				Config: hclReleaseDefinitionEnvironmentTemplateBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttrSet(tfNode, "can_delete"),
					checkReleaseDefinitionEnvironmentTemplateExists(name),
				),
			},
			// Step 2 — idempotency: re-plan must report no changes.
			{
				Config:             hclReleaseDefinitionEnvironmentTemplateBasic(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 3 — import: imported state must match apply state.
			{
				ResourceName:      tfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(tfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// checkReleaseDefinitionEnvironmentTemplateExists verifies that the template with the given name
// exists in ADO after a terraform apply.
func checkReleaseDefinitionEnvironmentTemplateExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tmpl, err := getReleaseDefinitionEnvironmentTemplateFromState(s)
		if err != nil {
			return err
		}
		if tmpl.Name == nil || *tmpl.Name != name {
			return fmt.Errorf("expected environment template name %q in ADO, got %v", name, tmpl.Name)
		}
		return nil
	}
}

// checkReleaseDefinitionEnvironmentTemplateDestroyed verifies that the template has been removed
// from ADO after terraform destroy.
func checkReleaseDefinitionEnvironmentTemplateDestroyed(s *terraform.State) error {
	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_release_definition_environment_template" {
			continue
		}
		projectID := res.Primary.Attributes["project_id"]
		templateID, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("parsing template ID %q: %w", res.Primary.ID, err)
		}
		clients := testutils.GetProvider().Meta().(*client.AggregatedClient)
		tmpl, err := clients.EnvironmentTemplatesClient.GetEnvironmentTemplate(
			context.Background(),
			&environmenttemplates.GetEnvironmentTemplateArgs{
				Project:    &projectID,
				TemplateId: &templateID,
			},
		)
		if err == nil && tmpl != nil && tmpl.Id != nil {
			return fmt.Errorf("expected environment template %q to be destroyed, but it still exists in ADO", res.Primary.ID)
		}
		// Any error (including 404) means it's gone — that's success.
	}
	return nil
}

// getReleaseDefinitionEnvironmentTemplateFromState reads the first environment template resource
// from the state and retrieves it from ADO.
func getReleaseDefinitionEnvironmentTemplateFromState(s *terraform.State) (*environmenttemplates.ReleaseDefinitionEnvironmentTemplate, error) {
	for _, res := range s.RootModule().Resources {
		if res.Type == "betterado_release_definition_environment_template" {
			return getReleaseDefinitionEnvironmentTemplateFromResource(res)
		}
	}
	return nil, fmt.Errorf("no betterado_release_definition_environment_template found in state")
}

func getReleaseDefinitionEnvironmentTemplateFromResource(res *terraform.ResourceState) (*environmenttemplates.ReleaseDefinitionEnvironmentTemplate, error) {
	projectID := res.Primary.Attributes["project_id"]
	templateID, err := uuid.Parse(res.Primary.ID)
	if err != nil {
		return nil, fmt.Errorf("parsing template ID %q: %w", res.Primary.ID, err)
	}
	clients := testutils.GetProvider().Meta().(*client.AggregatedClient)
	return clients.EnvironmentTemplatesClient.GetEnvironmentTemplate(
		context.Background(),
		&environmenttemplates.GetEnvironmentTemplateArgs{
			Project:    &projectID,
			TemplateId: &templateID,
		},
	)
}

// hclReleaseDefinitionEnvironmentTemplateBasic produces HCL that creates a project and a
// betterado_release_definition_environment_template referencing it.
func hclReleaseDefinitionEnvironmentTemplateBasic(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  description        = "%[1]s-description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_release_definition_environment_template" "test" {
  project_id  = betterado_project.test.id
  name        = "%[1]s"
  description = "acceptance test environment template"
}
`, name)
}
