//go:build (all || resource_wiki_page) && !exclude_resource_wiki_page

package acceptancetests

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	azwiki "github.com/microsoft/azure-devops-go-api/azuredevops/v7/wiki"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// getWikiPageDirectClient builds an AggregatedClient from AZDO env vars.
// Used by CheckDestroy because ProtoV6ProviderFactories does not wire the
// SDKv2 provider singleton's Meta.
func getWikiPageDirectClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

// checkWikiPageDestroyedByAttrs verifies that the wiki page resource is gone
// after destroy by reading path/project_id/wiki_id from the state entry.
func checkWikiPageDestroyedByAttrs(s *terraform.State) error {
	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_wiki_page" {
			continue
		}
		projectID := res.Primary.Attributes["project_id"]
		wikiID := res.Primary.Attributes["wiki_id"]
		path := res.Primary.Attributes["path"]

		clients, err := getWikiPageDirectClient()
		if err != nil {
			return nil // best-effort
		}
		_, err = clients.WikiClient.GetPage(clients.Ctx, azwiki.GetPageArgs{
			Project:        converter.String(projectID),
			WikiIdentifier: converter.String(wikiID),
			Path:           converter.String(path),
			IncludeContent: converter.Bool(false),
		})
		if err != nil {
			// Any error including 404 means the page is gone — pass.
			return nil
		}
		return fmt.Errorf("wiki page at path %q still exists after destroy", path)
	}
	return nil
}

// captureWikiPageEvidence performs a real live API GET of the wiki page and
// persists the response as forge demo live-evidence. Best-effort: failure
// never fails the test.
func captureWikiPageEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		pageID := res.Primary.ID
		projectID := res.Primary.Attributes["project_id"]
		wikiID := res.Primary.Attributes["wiki_id"]

		clients, err := getWikiPageDirectClient()
		if err != nil {
			return nil
		}

		id, err := strconv.Atoi(pageID)
		if err != nil {
			return nil
		}
		pageResp, err := clients.WikiClient.GetPageById(clients.Ctx, azwiki.GetPageByIdArgs{
			Project:        converter.String(projectID),
			WikiIdentifier: converter.String(wikiID),
			Id:             &id,
			IncludeContent: converter.Bool(true),
		})
		if err != nil || pageResp == nil {
			return nil
		}

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/%s/_apis/wiki/wikis/%s/pages/%s?includeContent=true&api-version=7.1",
			orgURL, projectID, wikiID, pageID)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-wiki-page", url, pageResp)
		return nil
	}
}

// TestAccWikiPageResource_basic verifies creating and reading a wiki page using
// the framework provider (betterado_wiki_page registered via GetMuxedProviderFactories).
// Uses the standing fixture project (SharedFixtureProjectName) to avoid the
// org's 1000-project cap.
func TestAccWikiPageResource_basic(t *testing.T) {
	wikiName := testutils.GenerateResourceName()
	tf := "betterado_wiki_page.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkWikiPageDestroyedByAttrs,
		Steps: []resource.TestStep{
			{
				Config: hclWikiPageBasic(wikiName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tf, "id"),
					resource.TestCheckResourceAttrSet(tf, "project_id"),
					resource.TestCheckResourceAttrSet(tf, "wiki_id"),
					resource.TestCheckResourceAttr(tf, "path", "/page-path"),
					resource.TestCheckResourceAttr(tf, "content", "content"),
					captureWikiPageEvidence(tf),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclWikiPageBasic(wikiName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccWikiPageResource_update verifies that updating a wiki page's content
// succeeds and the updated value is reflected in state.
// Uses the standing fixture project (SharedFixtureProjectName) to avoid the
// org's 1000-project cap.
func TestAccWikiPageResource_update(t *testing.T) {
	wikiName := testutils.GenerateResourceName()
	tf := "betterado_wiki_page.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkWikiPageDestroyedByAttrs,
		Steps: []resource.TestStep{
			{
				Config: hclWikiPageBasic(wikiName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tf, "id"),
					resource.TestCheckResourceAttrSet(tf, "project_id"),
					resource.TestCheckResourceAttrSet(tf, "wiki_id"),
					resource.TestCheckResourceAttr(tf, "path", "/page-path"),
					resource.TestCheckResourceAttr(tf, "content", "content"),
				),
			},
			{
				Config: hclWikiPageUpdate(wikiName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tf, "id"),
					resource.TestCheckResourceAttrSet(tf, "project_id"),
					resource.TestCheckResourceAttrSet(tf, "wiki_id"),
					resource.TestCheckResourceAttr(tf, "path", "/page-path"),
					resource.TestCheckResourceAttr(tf, "content", "contentupdate"),
					captureWikiPageEvidence(tf),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclWikiPageUpdate(wikiName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclWikiPageBasic creates a betterado_wiki (codeWiki) backed by a new git repo
// in the standing fixture project, and a betterado_wiki_page within that wiki.
// Uses codeWiki (not projectWiki) so parallel runs don't collide on the
// one-project-wiki-per-project ADO limit. Uses SharedFixtureProjectName to
// avoid the org's 1000-project cap.
func hclWikiPageBasic(wikiName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[2]q
}

resource "betterado_git_repository" "wiki_repo" {
  project_id = data.betterado_project.fixture.id
  name       = %[1]q
  initialization {
    init_type = "Clean"
  }
}

resource "betterado_wiki" "test" {
  project_id    = data.betterado_project.fixture.id
  repository_id = betterado_git_repository.wiki_repo.id
  name          = %[1]q
  version       = "master"
  type          = "codeWiki"
  mapped_path   = "/"
}

resource "betterado_wiki_page" "test" {
  project_id = data.betterado_project.fixture.id
  wiki_id    = betterado_wiki.test.id
  path       = "/page-path"
  content    = "content"
  version    = "master"
}
`, wikiName, SharedFixtureProjectName)
}

// hclWikiPageUpdate updates the content of the wiki page in the standing fixture project.
func hclWikiPageUpdate(wikiName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[2]q
}

resource "betterado_git_repository" "wiki_repo" {
  project_id = data.betterado_project.fixture.id
  name       = %[1]q
  initialization {
    init_type = "Clean"
  }
}

resource "betterado_wiki" "test" {
  project_id    = data.betterado_project.fixture.id
  repository_id = betterado_git_repository.wiki_repo.id
  name          = %[1]q
  version       = "master"
  type          = "codeWiki"
  mapped_path   = "/"
}

resource "betterado_wiki_page" "test" {
  project_id = data.betterado_project.fixture.id
  wiki_id    = betterado_wiki.test.id
  path       = "/page-path"
  content    = "contentupdate"
  version    = "master"
}
`, wikiName, SharedFixtureProjectName)
}
