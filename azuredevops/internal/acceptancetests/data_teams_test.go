package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccTeams_DataSource_basic(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	tfNode := "data.betterado_teams.test"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
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
				),
			},
		},
	})
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
