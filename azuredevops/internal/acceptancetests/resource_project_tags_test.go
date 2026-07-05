package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// buildTestClient creates an AggregatedClient from environment variables.
// Used by framework-based tests that cannot call testutils.GetProvider().Meta().
func buildTestClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

// TestAccProjectTags_basic operates on the betterado-standing-demo fixture project
// so that the captured evidence honestly references the standing-demo project GUID.
func TestAccProjectTags_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance tests skipped unless TF_ACC is set")
	}
	testutils.PreCheck(t, nil)
	projectID := ResolveFixtureProjectID(t)

	tfNode := "betterado_project_tags.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclProjectTagsFixture(projectID, []string{"tag1", "tag2"}),
				Check: resource.ComposeTestCheckFunc(
					CheckProjectTagsExist(),
					resource.TestCheckResourceAttr(tfNode, "tags.#", "2"),
					resource.TestCheckResourceAttr(tfNode, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(tfNode, "tags.1", "tag2"),
					captureProjectTagsEvidence(tfNode),
				),
			},
			{
				// Clean up: remove tags from the fixture project.
				Config: hclProjectTagsFixture(projectID, []string{}),
			},
		},
	})
}

// captureProjectTagsEvidence captures live evidence of project_tags
// for the forge demo pipeline. Best-effort: never fails the test check.
func captureProjectTagsEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		if orgURL == "" {
			return nil
		}
		projectID := rs.Primary.Attributes["project_id"]
		apiURL := fmt.Sprintf("%s/_apis/projects/%s/properties?api-version=7.1", orgURL, projectID)
		attrs := map[string]string{
			"project_id": projectID,
			"tags_count": rs.Primary.Attributes["tags.#"],
		}
		_ = testutils.CaptureLiveEvidence("project-tags", apiURL, attrs)
		return nil
	}
}

func TestAccProjectTags_update(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_project_tags.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkProjectTagsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclProjectTagsBasic(name),
				Check: resource.ComposeTestCheckFunc(
					CheckProjectTagsExist(),
					resource.TestCheckResourceAttr(tfNode, "tags.#", "2"),
					resource.TestCheckResourceAttr(tfNode, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(tfNode, "tags.1", "tag2"),
				),
			},
			{
				Config: hclProjectTagsUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					CheckProjectTagsExist(),
					resource.TestCheckResourceAttr(tfNode, "tags.#", "2"),
					resource.TestCheckResourceAttr(tfNode, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(tfNode, "tags.1", "tag3"),
				),
			},
		},
	})
}

func TestAccProjectTags_requiresImportError(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_project_tags.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkProjectTagsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclProjectTagsBasic(name),
				Check: resource.ComposeTestCheckFunc(
					CheckProjectTagsExist(),
					resource.TestCheckResourceAttr(tfNode, "tags.#", "2"),
				),
			},
			{
				Config: hclProjectTagsImport(name),
				Check: resource.ComposeTestCheckFunc(
					CheckProjectTagsExist(),
					resource.TestCheckResourceAttr(tfNode, "tags.#", "2"),
					resource.TestCheckResourceAttr(tfNode, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(tfNode, "tags.1", "tag2"),
				),
			},
		},
	})
}

func checkProjectTagsDestroyed(s *terraform.State) error {
	clients, err := buildTestClient()
	if err != nil {
		return fmt.Errorf("checkProjectTagsDestroyed: build client: %v", err)
	}
	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_project_tags" {
			continue
		}
		id := res.Primary.ID
		projectID, err := uuid.Parse(id)
		if err != nil {
			return err
		}

		tags, err := clients.CoreClient.GetProjectProperties(clients.Ctx, core.GetProjectPropertiesArgs{
			ProjectId: &projectID,
			Keys:      &[]string{"Microsoft.TeamFoundation.Project.Tag.*"},
		})
		if err != nil {
			if utils.ResponseWasNotFound(err) {
				return nil
			}
			return fmt.Errorf("Get project Tags (Project ID: %s). Error: %+v", id, err)
		}

		if tags != nil && len(*tags) != 0 {
			return fmt.Errorf("Project Tags (Project ID: %s) should not exist", id)
		}
	}
	return nil
}

func CheckProjectTagsExist() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources["betterado_project_tags.test"]
		if !ok {
			return fmt.Errorf("Did not find a `betterado_project_tags` in the TF state")
		}

		clients, err := buildTestClient()
		if err != nil {
			return fmt.Errorf("CheckProjectTagsExist: build client: %v", err)
		}
		id := res.Primary.ID
		projectID, err := uuid.Parse(id)
		if err != nil {
			return err
		}

		_, err = clients.CoreClient.GetProjectProperties(clients.Ctx, core.GetProjectPropertiesArgs{
			ProjectId: &projectID,
			Keys:      &[]string{"Microsoft.TeamFoundation.Project.Tag.*"},
		})
		if err != nil {
			return fmt.Errorf("Project Tags with Project ( Project ID=%s ) not found!. Error=%v", id, err)
		}
		return nil
	}
}

func hclProjectTagsTemplate(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%[1]s"
}`, name)
}

func hclProjectTagsBasic(name string) string {
	return fmt.Sprintf(`
%s

resource "betterado_project_tags" "test" {
  project_id = betterado_project.test.id
  tags       = ["tag1", "tag2"]
}
`, hclProjectTagsTemplate(name))
}

func hclProjectTagsUpdate(name string) string {
	return fmt.Sprintf(`
%s

resource "betterado_project_tags" "test" {
  project_id = betterado_project.test.id
  tags       = ["tag1", "tag3"]
}
`, hclProjectTagsTemplate(name))
}

func hclProjectTagsImport(name string) string {
	return fmt.Sprintf(`
%s

resource "betterado_project_tags" "import" {
  project_id = betterado_project_tags.test.project_id
  tags       = betterado_project_tags.test.tags
}
`, hclProjectTagsBasic(name))
}

// hclProjectTagsFixture sets tags on the betterado-standing-demo fixture project
// without creating a new project.
func hclProjectTagsFixture(projectID string, tags []string) string {
	tagsHCL := "["
	for i, tag := range tags {
		if i > 0 {
			tagsHCL += ", "
		}
		tagsHCL += fmt.Sprintf("%q", tag)
	}
	tagsHCL += "]"
	return fmt.Sprintf(`
resource "betterado_project_tags" "test" {
  project_id = %q
  tags       = %s
}
`, projectID, tagsHCL)
}
