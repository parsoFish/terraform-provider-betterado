package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccTeamMembers_CreateAndUpdate creates a team inside the betterado-standing-demo
// fixture project so that the captured evidence honestly references the standing-demo
// project GUID in its URL.
func TestAccTeamMembers_CreateAndUpdate(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance tests skipped unless TF_ACC is set")
	}
	testutils.PreCheck(t, nil)
	projectID := ResolveFixtureProjectID(t)
	teamName := testutils.GenerateResourceName()

	config1 := hclTeamMembersInFixtureProject(projectID, teamName, 1)
	config2 := hclTeamMembersInFixtureProject(projectID, teamName, 2)
	config3 := hclTeamMembersInFixtureProject(projectID, teamName, 1)

	tfNode := "betterado_team_members.team_members"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
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
			"project_id":    projectID,
			"team_id":       teamID,
			"members_count": rs.Primary.Attributes["members.#"],
		}
		_ = testutils.CaptureLiveEvidence("team-members", apiURL, attrs)
		return nil
	}
}

// hclTeamMembersInFixtureProject creates a team in the fixture project and assigns
// memberCount members from the project's built-in groups.
func hclTeamMembersInFixtureProject(projectID, teamName string, memberCount int) string {
	membersHCL := `data.betterado_group.contributors.descriptor`
	if memberCount >= 2 {
		membersHCL = `data.betterado_group.contributors.descriptor,
    data.betterado_group.readers.descriptor`
	}
	return fmt.Sprintf(`
resource "betterado_team" "team" {
  project_id = %q
  name       = %q
}

data "betterado_group" "contributors" {
  project_id = %q
  name       = "Contributors"
}

data "betterado_group" "readers" {
  project_id = %q
  name       = "Readers"
}

resource "betterado_team_members" "team_members" {
  project_id = betterado_team.team.project_id
  team_id    = betterado_team.team.id
  members = [
    %s
  ]
}
`, projectID, teamName, projectID, projectID, membersHCL)
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
