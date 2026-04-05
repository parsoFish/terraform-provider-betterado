package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccTeamAdministrators_CreateAndUpdate(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	teamName := testutils.GenerateResourceName()

	config1 := fmt.Sprintf(`


%s

data "betterado_group" "builtin_project_contributors" {
  project_id = betterado_project.project.id
  name       = "Contributors"
}

resource "betterado_team_administrators" "team_administrators" {
  project_id = betterado_team.team.project_id
  team_id    = betterado_team.team.id
  administrators = [
    data.betterado_group.builtin_project_contributors.descriptor
  ]
}




	`, testutils.HclTeamConfiguration(projectName, teamName, "", nil, nil))

	config2 := fmt.Sprintf(`


%s

data "betterado_group" "builtin_project_contributors" {
  project_id = betterado_project.project.id
  name       = "Contributors"
}

data "betterado_group" "builtin_project_readers" {
  project_id = betterado_project.project.id
  name       = "Readers"
}

resource "betterado_team_administrators" "team_administrators" {
  project_id = betterado_team.team.project_id
  team_id    = betterado_team.team.id
  administrators = [
    data.betterado_group.builtin_project_contributors.descriptor,
    data.betterado_group.builtin_project_readers.descriptor
  ]
}


		`, testutils.HclTeamConfiguration(projectName, teamName, "", nil, nil))

	tfNode := "betterado_team_administrators.team_administrators"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "team_id"),
					resource.TestCheckResourceAttr(tfNode, "administrators.#", "1"),
				),
			},
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "team_id"),
					resource.TestCheckResourceAttr(tfNode, "administrators.#", "2"),
				),
			},
		},
	})
}

func TestAccTeamAdministrators_CreateAndUpdate_Overwrite(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	teamName := testutils.GenerateResourceName()

	config1 := fmt.Sprintf(`


%s

data "betterado_group" "builtin_project_contributors" {
  project_id = betterado_project.project.id
  name       = "Contributors"
}

resource "betterado_team_administrators" "team_administrators" {
  project_id = betterado_team.team.project_id
  team_id    = betterado_team.team.id
  mode       = "overwrite"
  administrators = [
    data.betterado_group.builtin_project_contributors.descriptor
  ]
}




	`, testutils.HclTeamConfiguration(projectName, teamName, "", nil, nil))

	config2 := fmt.Sprintf(`


%s

data "betterado_group" "builtin_project_contributors" {
  project_id = betterado_project.project.id
  name       = "Contributors"
}

data "betterado_group" "builtin_project_readers" {
  project_id = betterado_project.project.id
  name       = "Readers"
}

resource "betterado_team_administrators" "team_administrators" {
  project_id = betterado_team.team.project_id
  team_id    = betterado_team.team.id
  mode       = "overwrite"
  administrators = [
    data.betterado_group.builtin_project_contributors.descriptor,
    data.betterado_group.builtin_project_readers.descriptor
  ]
}


		`, testutils.HclTeamConfiguration(projectName, teamName, "", nil, nil))

	tfNode := "betterado_team_administrators.team_administrators"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: config1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "team_id"),
					resource.TestCheckResourceAttr(tfNode, "administrators.#", "1"),
				),
			},
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "team_id"),
					resource.TestCheckResourceAttr(tfNode, "administrators.#", "2"),
				),
			},
		},
	})
}
