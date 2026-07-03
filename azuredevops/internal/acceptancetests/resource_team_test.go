package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccTeam_basic(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	teamName := testutils.GenerateResourceName()

	tfNode := "betterado_team.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclTeamBasic(projectName, teamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", teamName),
					captureTeamEvidence(tfNode),
				),
			},
		},
	})
}

// captureTeamEvidence captures live evidence of the betterado_team resource
// for the forge demo pipeline. Best-effort: never fails the test check.
func captureTeamEvidence(tfNode string) resource.TestCheckFunc {
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
		_ = testutils.CaptureLiveEvidence("team", apiURL, attrs)
		return nil
	}
}

func TestAccTeam_update(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	teamName := testutils.GenerateResourceName()
	teamDescription := "description"
	teamName2 := testutils.GenerateResourceName()
	teamDescription2 := "@@descriptionUpdate"

	tfNode := "betterado_team.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclTeamUpdate(projectName, teamName, teamDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", teamName),
					resource.TestCheckResourceAttr(tfNode, "description", teamDescription),
					resource.TestCheckResourceAttrSet(tfNode, "administrators.#"),
					resource.TestCheckResourceAttrSet(tfNode, "members.#"),
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
				),
			},
			{
				Config: hclTeamUpdate(projectName, teamName2, teamDescription2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", teamName2),
					resource.TestCheckResourceAttr(tfNode, "description", teamDescription2),
					resource.TestCheckResourceAttrSet(tfNode, "administrators.#"),
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
				),
			},
		},
	})
}

func TestAccTeam_membersUpdate(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	teamName := testutils.GenerateResourceName()
	tfNode := "betterado_team.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclTeamMembersBasic(projectName, teamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", teamName),
					resource.TestCheckResourceAttr(tfNode, "members.#", "1"),
				),
			},
			{
				Config: hclTeamMembersUpdate(projectName, teamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", teamName),
					resource.TestCheckResourceAttr(tfNode, "members.#", "2"),
				),
			},
			{
				Config: hclTeamMembersBasic(projectName, teamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", teamName),
					resource.TestCheckResourceAttr(tfNode, "members.#", "1"),
				),
			},
		},
	})
}

func TestAccTeam_administratorsUpdate(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	teamName := testutils.GenerateResourceName()

	tfNode := "betterado_team.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclTeamAdministratorsBasic(projectName, teamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", teamName),
					resource.TestCheckResourceAttrSet(tfNode, "administrators.#"),
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
				),
			},
			{
				Config: hclTeamAdministratorsUpdate(projectName, teamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", teamName),
					resource.TestCheckResourceAttr(tfNode, "administrators.#", "2"),
				),
			},
			{
				Config: hclTeamAdministratorsBasic(projectName, teamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", teamName),
					resource.TestCheckResourceAttr(tfNode, "administrators.#", "1"),
				),
			},
		},
	})
}

func TestAccTeam_complete(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	teamName := testutils.GenerateResourceName()

	tfNode := "betterado_team.test"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclTeamComplete(projectName, teamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", teamName),
					resource.TestCheckResourceAttr(tfNode, "administrators.#", "1"),
					resource.TestCheckResourceAttr(tfNode, "members.#", "1"),
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
				),
			},
		},
	})
}

func hclTeamBasic(projectName, teamName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_team" "test" {
  project_id = betterado_project.test.id
  name       = "%s"
}
`, projectName, teamName)
}

func hclTeamUpdate(projectName, teamName, description string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_team" "test" {
  project_id  = betterado_project.test.id
  name        = "%s"
  description = "%s"
}
`, projectName, teamName, description)
}

func hclTeamMembersBasic(projectName, teamName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

data "betterado_group" "test" {
  project_id = betterado_project.test.id
  name       = "Contributors"
}

resource "betterado_team" "test" {
  project_id  = betterado_project.test.id
  name        = "%s"
  description = "Test sTeam"

  members = [
    data.betterado_group.test.descriptor
  ]
}
`, projectName, teamName)
}

func hclTeamMembersUpdate(projectName, teamName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

data "betterado_group" "test" {
  project_id = betterado_project.test.id
  name       = "Contributors"
}

data "betterado_group" "test2" {
  project_id = betterado_project.test.id
  name       = "Readers"
}

resource "betterado_team" "test" {
  project_id  = betterado_project.test.id
  name        = "%s"
  description = "Test sTeam"

  members = [
    data.betterado_group.test.descriptor,
    data.betterado_group.test2.descriptor
  ]
}
`, projectName, teamName)
}

func hclTeamAdministratorsBasic(projectName, teamName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

data "betterado_group" "test" {
  project_id = betterado_project.test.id
  name       = "Contributors"
}

resource "betterado_team" "test" {
  project_id  = betterado_project.test.id
  name        = "%s"
  description = "Test sTeam"

  administrators = [
    data.betterado_group.test.descriptor
  ]
}
`, projectName, teamName)
}

func hclTeamAdministratorsUpdate(projectName, teamName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

data "betterado_group" "test" {
  project_id = betterado_project.test.id
  name       = "Contributors"
}

data "betterado_group" "test2" {
  project_id = betterado_project.test.id
  name       = "Readers"
}

resource "betterado_team" "test" {
  project_id  = betterado_project.test.id
  name        = "%s"
  description = "Test sTeam"

  administrators = [
    data.betterado_group.test.descriptor,
    data.betterado_group.test2.descriptor
  ]
}
`, projectName, teamName)
}

func hclTeamComplete(projectName, teamName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

data "betterado_group" "test" {
  project_id = betterado_project.test.id
  name       = "Contributors"
}

data "betterado_group" "test2" {
  project_id = betterado_project.test.id
  name       = "Readers"
}

resource "betterado_team" "test" {
  project_id  = betterado_project.test.id
  name        = "%s"
  description = "Test sTeam"

  administrators = [
    data.betterado_group.test.descriptor,
  ]

  members = [
    data.betterado_group.test2.descriptor
  ]
}`, projectName, teamName)
}
