package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccTeams_DataSource_basic(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	tfNode := "data.betterado_teams.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclTeamsDataSourceBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttr(tfNode, "teams.#", "1"),
					resource.TestCheckResourceAttrSet(tfNode, "teams.0.project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "teams.0.id"),
					resource.TestCheckResourceAttrSet(tfNode, "teams.0.name"),
					resource.TestCheckResourceAttrSet(tfNode, "teams.0.description"),
					resource.TestCheckResourceAttrSet(tfNode, "teams.0.administrators.#"),
					resource.TestCheckResourceAttrSet(tfNode, "teams.0.members.#"),
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

func hclTeamsDataSourceBasic(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  description        = "description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

data "betterado_teams" "test" {
  project_id = betterado_project.test.id
}
`, name)
}
