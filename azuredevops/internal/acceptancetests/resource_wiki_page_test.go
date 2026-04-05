package acceptancetests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccWikiPageResource_basic(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	tf := "betterado_wiki.test"
	resourceType := "betterado_wiki"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkWikiDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclProjectWikiPageBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betterado_wiki_page.test", "project_id"),
					resource.TestCheckResourceAttrSet("betterado_wiki_page.test", "wiki_id"),
					resource.TestCheckResourceAttrSet("betterado_wiki_page.test", "path"),
					resource.TestCheckResourceAttrSet("betterado_wiki_page.test", "content"),
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

func TestAccWikiPageResource_update(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	tf := "betterado_wiki.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkWikiDestroyed("betterado_wiki"),
		Steps: []resource.TestStep{
			{
				Config: hclProjectWikiPageBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betterado_wiki_page.test", "project_id"),
					resource.TestCheckResourceAttrSet("betterado_wiki_page.test", "wiki_id"),
					resource.TestCheckResourceAttr("betterado_wiki_page.test", "path", "/path"),
					resource.TestCheckResourceAttr("betterado_wiki_page.test", "content", "content"),
				),
			},
			{
				ResourceName:      tf,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: hclProjectWikiPageUpdate(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betterado_wiki_page.test", "project_id"),
					resource.TestCheckResourceAttrSet("betterado_wiki_page.test", "wiki_id"),
					resource.TestCheckResourceAttr("betterado_wiki_page.test", "path", "/path"),
					resource.TestCheckResourceAttr("betterado_wiki_page.test", "content", "contentupdate"),
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

func TestAccWikiPageResource_requireImportError(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	tf := "betterado_wiki.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkWikiDestroyed("betterado_wiki"),
		Steps: []resource.TestStep{
			{
				Config: hclProjectWikiPageBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("betterado_wiki_page.test", "project_id"),
					resource.TestCheckResourceAttrSet("betterado_wiki_page.test", "wiki_id"),
					resource.TestCheckResourceAttr("betterado_wiki_page.test", "path", "/path"),
					resource.TestCheckResourceAttr("betterado_wiki_page.test", "content", "content"),
				),
			},
			{
				ResourceName:      tf,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:      hclProjectWikiPageImport(projectName),
				ExpectError: regexp.MustCompile(`The page '/path' specified in the add operation already exists in the wiki. Please specify a new page path.`),
			},
		},
	})
}

func hclProjectWikiPageBasic(projectName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_wiki" "test" {
  project_id = betterado_project.test.id
  name       = "projectWikiRepo"
  type       = "projectWiki"
}

resource "betterado_wiki_page" "test" {
  project_id = betterado_project.test.id
  wiki_id    = betterado_wiki.test.id
  path       = "/path"
  content    = "content"
}
`, projectName)
}

func hclProjectWikiPageUpdate(projectName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_wiki" "test" {
  project_id = betterado_project.test.id
  name       = "projectWikiRepo"
  type       = "projectWiki"
}

resource "betterado_wiki_page" "test" {
  project_id = betterado_project.test.id
  wiki_id    = betterado_wiki.test.id
  path       = "/path"
  content    = "contentupdate"
}
`, projectName)
}

func hclProjectWikiPageImport(projectName string) string {
	return fmt.Sprintf(`
%s

resource "betterado_wiki_page" "import" {
  project_id = betterado_wiki_page.test.project_id
  wiki_id    = betterado_wiki_page.test.wiki_id
  path       = betterado_wiki_page.test.path
  content    = betterado_wiki_page.test.content
}
`, hclProjectWikiPageBasic(projectName))
}
