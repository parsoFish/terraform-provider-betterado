//go:build (all || resource_wiki) && !exclude_resource_wiki

package acceptancetests

import (
	"fmt"
	"os"
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

// getWikiDirectClient builds an AggregatedClient from AZDO env vars.
// Used by CheckDestroy because ProtoV6ProviderFactories does not wire the
// SDKv2 provider singleton's Meta.
func getWikiDirectClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

// TestAccWikiResource_projectWiki verifies creating a project wiki resource
// against the standing fixture project (betterado-standing-demo), which avoids
// hitting the org's 1000-project cap. Project wikis are auto-created for ADO
// projects; we create a uniquely-named one here and destroy it after the test.
func TestAccWikiResource_projectWiki(t *testing.T) {
	wikiName := testutils.GenerateResourceName()
	tf := "betterado_wiki.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkWikiDestroyedFramework,
		Steps: []resource.TestStep{
			{
				Config: hclWikiProjectWiki(wikiName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tf, "project_id"),
					resource.TestCheckResourceAttrSet(tf, "name"),
					resource.TestCheckResourceAttr(tf, "type", "projectWiki"),
					captureWikiEvidence(tf),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclWikiProjectWiki(wikiName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccWikiResource_codeWiki verifies creating a code wiki resource in the
// standing fixture project using a new repo created within it.
func TestAccWikiResource_codeWiki(t *testing.T) {
	repoName := testutils.GenerateResourceName()
	wikiName := testutils.GenerateResourceName()
	tf := "betterado_wiki.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkWikiDestroyedFramework,
		Steps: []resource.TestStep{
			{
				Config: hclWikiCodeWiki(repoName, wikiName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tf, "project_id"),
					resource.TestCheckResourceAttr(tf, "type", "codeWiki"),
					resource.TestCheckResourceAttrSet(tf, "name"),
					resource.TestCheckResourceAttrSet(tf, "repository_id"),
					resource.TestCheckResourceAttrSet(tf, "version"),
					resource.TestCheckResourceAttrSet(tf, "mapped_path"),
					captureWikiEvidence(tf),
				),
			},
			// Idempotency: re-plan must produce no diff.
			{
				Config:             hclWikiCodeWiki(repoName, wikiName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func checkWikiDestroyedFramework(s *terraform.State) error {
	clients, err := getWikiDirectClient()
	if err != nil {
		return fmt.Errorf("checkWikiDestroyedFramework: failed to build ADO client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_wiki" {
			continue
		}

		_, err := clients.WikiClient.GetWiki(clients.Ctx, azwiki.GetWikiArgs{WikiIdentifier: converter.String(res.Primary.ID)})
		if err == nil {
			return fmt.Errorf("found wiki %s that should have been deleted", res.Primary.ID)
		}
	}
	return nil
}

// captureWikiEvidence performs a real live API GET of the created wiki and
// persists the response as forge demo live-evidence. Best-effort: a failure
// never fails the test.
func captureWikiEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		wikiID := res.Primary.ID
		clients, err := getWikiDirectClient()
		if err != nil {
			return nil
		}
		wikiResp, err := clients.WikiClient.GetWiki(clients.Ctx, azwiki.GetWikiArgs{WikiIdentifier: converter.String(wikiID)})
		if err != nil || wikiResp == nil {
			return nil
		}

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		projectID := res.Primary.Attributes["project_id"]
		url := fmt.Sprintf("%s/%s/_apis/wiki/wikis/%s?api-version=7.1", orgURL, projectID, wikiID)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-wiki", url, wikiResp)
		return nil
	}
}

// hclWikiProjectWiki creates a project wiki in the standing fixture project.
// Uses SharedFixtureProjectName to avoid the 1000-project org cap.
func hclWikiProjectWiki(wikiName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[2]q
}

resource "betterado_wiki" "test" {
  project_id = data.betterado_project.fixture.id
  name       = %[1]q
  type       = "projectWiki"
}
`, wikiName, SharedFixtureProjectName)
}

// hclWikiCodeWiki creates a code wiki backed by a new git repo in the standing
// fixture project.
func hclWikiCodeWiki(repoName, wikiName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[3]q
}

resource "betterado_git_repository" "test" {
  project_id = data.betterado_project.fixture.id
  name       = %[1]q
  initialization {
    init_type = "Clean"
  }
}

resource "betterado_wiki" "test" {
  project_id    = data.betterado_project.fixture.id
  repository_id = betterado_git_repository.test.id
  name          = %[2]q
  version       = "master"
  type          = "codeWiki"
  mapped_path   = "/"
}
`, repoName, wikiName, SharedFixtureProjectName)
}
