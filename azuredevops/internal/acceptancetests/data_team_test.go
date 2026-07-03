package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccTeam_DataSource_Basic(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	tfNode := "data.betterado_team.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclTeamDataSourceBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "name"),
					resource.TestCheckResourceAttrSet(tfNode, "description"),
					resource.TestCheckResourceAttrSet(tfNode, "administrators.#"),
					resource.TestCheckResourceAttrSet(tfNode, "members.#"),
					captureTeamDataSourceEvidence(tfNode),
				),
			},
		},
	})
}

// captureTeamDataSourceEvidence captures live evidence of the data_team data source
// for the forge demo pipeline. Best-effort: never fails the test check.
func captureTeamDataSourceEvidence(tfNode string) resource.TestCheckFunc {
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
		teamID := rs.Primary.Attributes["id"]
		apiURL := fmt.Sprintf("%s/_apis/projects/%s/teams/%s?api-version=7.1", orgURL, projectID, teamID)
		attrs := map[string]string{
			"id":         teamID,
			"name":       rs.Primary.Attributes["name"],
			"project_id": projectID,
		}
		_ = testutils.CaptureLiveEvidence("data-team", apiURL, attrs)
		return nil
	}
}

func hclTeamDataSourceBasic(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "AccTest %[1]s"
  description        = "description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

data "betterado_team" "test" {
  project_id = betterado_project.test.id
  name       = "${betterado_project.test.name} Team"
}


`, name)
}
