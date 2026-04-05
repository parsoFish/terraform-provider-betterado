package acceptancetests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/wiki"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

func TestAccWikiResource_projectWiki(t *testing.T) {
	projectName := testutils.GenerateResourceName()

	tf := "betterado_wiki.test"
	resourceType := "betterado_wiki"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkWikiDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclWikiResourceProjectWiki(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tf, "project_id"),
					resource.TestCheckResourceAttrSet(tf, "name"),
				),
			},
			{
				ResourceName:      tf,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWikiResource_codeWiki(t *testing.T) {
	projectName := testutils.GenerateResourceName()

	tf := "betterado_wiki.test"
	resourceType := "betterado_wiki"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkWikiDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclWikiResourceCodeWiki(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tf, "project_id"),
					resource.TestCheckResourceAttrSet(tf, "type"),
					resource.TestCheckResourceAttrSet(tf, "name"),
					resource.TestCheckResourceAttrSet(tf, "repository_id"),
					resource.TestCheckResourceAttrSet(tf, "version"),
					resource.TestCheckResourceAttrSet(tf, "mapped_path"),
				),
			},
			{
				ResourceName:      tf,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWikiResource_importErrorStep(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	tf := "betterado_wiki.test"
	resourceType := "betterado_wiki"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkWikiDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclWikiResourceProjectWiki(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betterado_wiki.test", "project_id"),
					resource.TestCheckResourceAttrSet("betterado_wiki.test", "type"),
					resource.TestCheckResourceAttrSet("betterado_wiki.test", "name"),
				),
			},
			{
				ResourceName:      tf,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:      hclWikiResourceRequiresImport(projectName),
				ExpectError: wikiRequiresImportError(),
			},
		},
	})
}

func checkWikiDestroyed(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, resource := range s.RootModule().Resources {
			if resource.Type != resourceType {
				continue
			}

			// indicates the resource exists - this should fail the test
			clients := testutils.GetProvider().Meta().(*client.AggregatedClient)
			_, err := clients.WikiClient.GetWiki(clients.Ctx, wiki.GetWikiArgs{WikiIdentifier: converter.String(resource.Primary.ID)})
			if err == nil {
				return fmt.Errorf("found wiki that should have been deleted")
			}
		}
		return nil
	}
}

func wikiRequiresImportError() *regexp.Regexp {
	message := "Error: Wiki already exists with name ('codeWikiRepo'|'projectWikiRepo')."
	return regexp.MustCompile(message)
}

func hclWikiResourceProjectWiki(projectName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_wiki" "test" {
  project_id = betterado_project.test.id
  name       = "projectWikiRepo"
  type       = "projectWiki"
}
`, projectName)
}

func hclWikiResourceCodeWiki(projectName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "Repo"
  initialization {
    init_type = "Clean"
  }
}

resource "betterado_wiki" "test" {
  project_id    = betterado_project.test.id
  repository_id = betterado_git_repository.test.id
  name          = "codeWikiRepo"
  version       = "master"
  type          = "codeWiki"
  mapped_path   = "/"
}`, projectName)
}

func hclWikiResourceRequiresImport(projectName string) string {
	return fmt.Sprintf(`
%s

resource "betterado_wiki" "import" {
  project_id = betterado_wiki.test.project_id
  name       = betterado_wiki.test.name
  type       = betterado_wiki.test.type
}
`, hclWikiResourceProjectWiki(projectName))
}
