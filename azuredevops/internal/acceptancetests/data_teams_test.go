package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccTeams_DataSource_basic reads the teams of the betterado-standing-demo
// fixture project so that the captured evidence honestly references the standing-demo
// project GUID in its URL.
func TestAccTeams_DataSource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance tests skipped unless TF_ACC is set")
	}
	testutils.PreCheck(t, nil)
	projectID := ResolveFixtureProjectID(t)

	tfNode := "data.betterado_teams.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclTeamsDataSourceFixture(projectID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "teams.#"),
					captureTeamsDataSourceEvidence(tfNode),
				),
			},
		},
	})
}

// captureTeamsDataSourceEvidence captures live evidence of the data_teams data source
// for the forge demo pipeline. Best-effort: never fails the test check.
func captureTeamsDataSourceEvidence(tfNode string) resource.TestCheckFunc {
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
		apiURL := fmt.Sprintf("%s/_apis/projects/%s/teams?api-version=7.1", orgURL, projectID)
		attrs := map[string]string{
			"project_id":  projectID,
			"teams_count": rs.Primary.Attributes["teams.#"],
		}
		_ = testutils.CaptureLiveEvidence("data-teams", apiURL, attrs)
		return nil
	}
}

// hclTeamsDataSourceFixture reads the teams from the betterado-standing-demo
// fixture project without creating any new project.
func hclTeamsDataSourceFixture(projectID string) string {
	return fmt.Sprintf(`
data "betterado_teams" "test" {
  project_id = %q
}
`, projectID)
}
