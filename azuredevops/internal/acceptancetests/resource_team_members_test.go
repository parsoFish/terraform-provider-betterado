package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccTeamMembers_CreateAndUpdate(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	teamName := testutils.GenerateResourceName()

	config1 := fmt.Sprintf(`


%s

data "betterado_group" "builtin_project_contributors" {
  project_id = betterado_project.project.id
  name       = "Contributors"
}

resource "betterado_team_members" "team_members" {
  project_id = betterado_team.team.project_id
  team_id    = betterado_team.team.id
  members = [
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

resource "betterado_team_members" "team_members" {
  project_id = betterado_team.team.project_id
  team_id    = betterado_team.team.id
  members = [
    data.betterado_group.builtin_project_contributors.descriptor,
    data.betterado_group.builtin_project_readers.descriptor
  ]
}


	`, testutils.HclTeamConfiguration(projectName, teamName, "", nil, nil))

	config3 := fmt.Sprintf(`


%s

data "betterado_group" "builtin_project_readers" {
  project_id = betterado_project.project.id
  name       = "Readers"
}

resource "betterado_team_members" "team_members" {
  project_id = betterado_team.team.project_id
  team_id    = betterado_team.team.id
  members = [
    data.betterado_group.builtin_project_readers.descriptor
  ]
}


	`, testutils.HclTeamConfiguration(projectName, teamName, "", nil, nil))

	tfNode := "betterado_team_members.team_members"
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
					resource.TestCheckResourceAttr(tfNode, "members.#", "1"),
					captureTeamMembersEvidence(tfNode),
				),
			},
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "team_id"),
					resource.TestCheckResourceAttr(tfNode, "members.#", "2"),
				),
			},
			{
				Config: config3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "team_id"),
					resource.TestCheckResourceAttr(tfNode, "members.#", "1"),
				),
			},
		},
	})
}

// captureTeamMembersEvidence captures live evidence of the team_members resource
// for the forge demo pipeline. Best-effort: never fails the test check.
func captureTeamMembersEvidence(tfNode string) resource.TestCheckFunc {
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
		teamID := rs.Primary.Attributes["team_id"]
		apiURL := fmt.Sprintf("%s/_apis/projects/%s/teams/%s/members?api-version=7.1", orgURL, projectID, teamID)
		attrs := map[string]string{
			"project_id":   projectID,
			"team_id":      teamID,
			"members_count": rs.Primary.Attributes["members.#"],
		}
		_ = testutils.CaptureLiveEvidence("team-members", apiURL, attrs)
		return nil
	}
}

func TestAccTeamMembers_CreateAndUpdate_Overwrite(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	teamName := testutils.GenerateResourceName()

	config1 := fmt.Sprintf(`


%s

data "betterado_group" "builtin_project_contributors" {
  project_id = betterado_project.project.id
  name       = "Contributors"
}

resource "betterado_team_members" "team_members" {
  project_id = betterado_team.team.project_id
  team_id    = betterado_team.team.id
  mode       = "overwrite"
  members = [
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

resource "betterado_team_members" "team_members" {
  project_id = betterado_team.team.project_id
  team_id    = betterado_team.team.id
  mode       = "overwrite"
  members = [
    data.betterado_group.builtin_project_contributors.descriptor,
    data.betterado_group.builtin_project_readers.descriptor
  ]
}


	`, testutils.HclTeamConfiguration(projectName, teamName, "", nil, nil))

	config3 := fmt.Sprintf(`


%s

data "betterado_group" "builtin_project_readers" {
  project_id = betterado_project.project.id
  name       = "Readers"
}

resource "betterado_team_members" "team_members" {
  project_id = betterado_team.team.project_id
  team_id    = betterado_team.team.id
  mode       = "overwrite"
  members = [
    data.betterado_group.builtin_project_readers.descriptor
  ]
}


	`, testutils.HclTeamConfiguration(projectName, teamName, "", nil, nil))

	tfNode := "betterado_team_members.team_members"
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
					resource.TestCheckResourceAttr(tfNode, "members.#", "1"),
				),
			},
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "team_id"),
					resource.TestCheckResourceAttr(tfNode, "members.#", "2"),
				),
			},
			{
				Config: config3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "team_id"),
					resource.TestCheckResourceAttr(tfNode, "members.#", "1"),
				),
			},
		},
	})
}
