//go:build all || resource_build_definition_framework

package acceptancetests

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccBuildDefinition_Framework_basic exercises the terraform-plugin-framework
// implementation of betterado_build_definition via the mux provider path:
//
//  1. apply  — creates the build definition with a TfsGit repo source, a
//     non-default path, and a variable.
//  2. read-back — asserts attributes match the configuration.
//  3. idempotency — re-plan produces no diff (ExpectNonEmptyPlan: false).
//  4. destroy — cleans up cleanly.
//
// Live evidence is captured via CaptureLiveEvidence("acceptance-resource", …)
// during the read-back step so forge demo render can back-fill demo.json.
//
// Uses SharedReleaseFixture to obtain a pre-existing persistent project
// (betterado-standing-demo), so no new ADO project is created.
func TestAccBuildDefinition_Framework_basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	fixture := SharedReleaseFixture(t)
	tfNode := "betterado_build_definition.fw_test"
	path := `\AccTestBD-` + name

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkBuildDefinitionFrameworkDestroyed(fixture.ProjectID),
		Steps: []resource.TestStep{
			// Step 1: create + assert read-back + capture live evidence.
			{
				Config: hclBuildDefinitionFrameworkBasic(name, fixture.ProjectID, fixture.RepoID, path),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttr(tfNode, "project_id", fixture.ProjectID),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "revision"),
					resource.TestCheckResourceAttr(tfNode, "path", path),
					resource.TestCheckResourceAttr(tfNode, "variable.#", "1"),
					captureBuildDefinitionFrameworkEvidence(tfNode),
				),
			},
			// Step 2: idempotency — no perpetual diff.
			{
				Config:             hclBuildDefinitionFrameworkBasic(name, fixture.ProjectID, fixture.RepoID, path),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclBuildDefinitionFrameworkBasic returns Terraform HCL for a build definition
// using a TfsGit repository source, a non-default path, and a single variable.
func hclBuildDefinitionFrameworkBasic(name, projectID, repoID, path string) string {
	return fmt.Sprintf(`
resource "betterado_build_definition" "fw_test" {
  project_id = %[2]q
  name       = %[1]q
  path       = %[4]q

  repository {
    repo_type   = "TfsGit"
    repo_id     = %[3]q
    branch_name = "master"
    yml_path    = "azure-pipelines.yml"
  }

  variable {
    name  = "AccTestVar"
    value = "hello"
  }
}
`, name, projectID, repoID, path)
}

// checkBuildDefinitionFrameworkDestroyed returns a CheckDestroyFunc that verifies
// the build definition no longer exists after destroy.
func checkBuildDefinitionFrameworkDestroyed(projectID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "betterado_build_definition" {
				continue
			}
			idStr := rs.Primary.Attributes["id"]
			if idStr == "" {
				continue
			}
			defID, err := strconv.Atoi(idStr)
			if err != nil {
				continue
			}
			clients, err := getDirectClient()
			if err != nil {
				return fmt.Errorf("checkBuildDefinitionFrameworkDestroyed: build client: %w", err)
			}
			_, apiErr := clients.BuildClient.GetDefinition(clients.Ctx, build.GetDefinitionArgs{
				Project:      &projectID,
				DefinitionId: &defID,
			})
			if apiErr == nil {
				return fmt.Errorf("build definition (id: %d) still exists after destroy", defID)
			}
			// A 404-style error means it's gone — that's what we want.
		}
		return nil
	}
}

// captureBuildDefinitionFrameworkEvidence performs a real live API GET of the
// created build definition and persists the response as forge demo live-evidence.
// Best-effort: a capture failure never fails the test.
func captureBuildDefinitionFrameworkEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		idStr := res.Primary.Attributes["id"]
		projectID := res.Primary.Attributes["project_id"]
		if idStr == "" || projectID == "" {
			return nil
		}

		defID, err := strconv.Atoi(idStr)
		if err != nil || defID == 0 {
			return nil
		}

		clients, clientErr := getDirectClient()
		if clientErr != nil {
			return nil // best-effort; do not fail the test on evidence capture
		}

		defResp, apiErr := clients.BuildClient.GetDefinition(clients.Ctx, build.GetDefinitionArgs{
			Project:      &projectID,
			DefinitionId: &defID,
		})
		if apiErr != nil || defResp == nil {
			return nil
		}

		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		for len(orgURL) > 0 && orgURL[len(orgURL)-1] == '/' {
			orgURL = orgURL[:len(orgURL)-1]
		}
		apiURL := fmt.Sprintf(
			"%s/%s/_apis/build/definitions/%d?api-version=7.1",
			orgURL, projectID, defID,
		)
		_ = testutils.CaptureLiveEvidence("acceptance-resource", apiURL, defResp)
		return nil
	}
}
