//go:build (all || resource_dashboard) && !exclude_resource_dashboard

package acceptancetests

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/dashboard"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// preCheckDashboard extends testutils.PreCheck by verifying the shared
// fixture project (SharedFixtureProjectName / "betterado-standing-demo")
// exists.  If the project is missing the test fails loudly via
// resolveOrCreateFixtureProject so a missing fixture can never silently pass
// the gate as a skip.
func preCheckDashboard(t *testing.T) {
	t.Helper()
	testutils.PreCheck(t, nil)

	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	authProvider := azuredevops.NewAuthProviderPAT(pat)
	clients, err := client.GetAzdoClient(authProvider, orgURL)
	if err != nil {
		t.Fatalf("preCheckDashboard: GetAzdoClient: %v", err)
	}

	// Fail loudly if the shared fixture project is missing — never skip.
	// If the project is absent, restore it from the ADO recycle bin
	// (Organization settings → Projects) and re-run.
	resolveOrCreateFixtureProject(t, clients)
}

// getDirectDashboardClient builds an AggregatedClient directly from AZDO env vars.
// Used by CheckDestroy and evidence helpers because ProtoV6ProviderFactories
// does not wire the SDKv2 provider singleton's Meta, so testutils.GetProvider().Meta()
// would be nil when the test uses the mux provider factory.
func getDirectDashboardClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

// captureDashboardEvidence fetches the live dashboard GET URL and writes forge live evidence.
// Best-effort: never fails the test — errors are silently swallowed.
func captureDashboardEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_ = tryCaptureDashboardEvidence(tfNode, s) //nolint:errcheck
		return nil
	}
}

// tryCaptureDashboardEvidence performs the actual evidence capture and returns
// an error on failure (caller ignores it).
func tryCaptureDashboardEvidence(tfNode string, s *terraform.State) error {
	res, ok := s.RootModule().Resources[tfNode]
	if !ok {
		return fmt.Errorf("resource %s not found in state", tfNode)
	}

	projectID := res.Primary.Attributes["project_id"]
	dashboardID := res.Primary.ID

	orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
	getURL := fmt.Sprintf("%s/%s/_apis/dashboard/dashboards/%s?api-version=7.1", orgURL, projectID, dashboardID)

	clients, err := getDirectDashboardClient()
	if err != nil {
		return err
	}

	id, err := uuid.Parse(dashboardID)
	if err != nil {
		return err
	}

	args := dashboard.GetDashboardArgs{
		Project:     converter.String(projectID),
		DashboardId: &id,
	}
	if teamID, ok := res.Primary.Attributes["team_id"]; ok && teamID != "" {
		args.Team = converter.String(teamID)
	}

	resp, err := clients.DashboardClient.GetDashboard(clients.Ctx, args)
	if err != nil {
		return err
	}

	return testutils.CaptureLiveEvidence("dashboard-acceptance-resource", getURL, resp)
}

func TestAccDashboard_project_basic(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_dashboard.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { preCheckDashboard(t) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkDashboardDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclDashboardProjectBasic(name),
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExist(name),
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttrSet(tfNode, "name"),
					resource.TestCheckResourceAttrSet(tfNode, "owner_id"),
					// Capture live evidence before destroy.
					captureDashboardEvidence(tfNode),
				),
			},
			// Idempotency: re-plan produces no diff.
			{
				Config:             hclDashboardProjectBasic(name),
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateIdFunc: importDashboardId(tfNode),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDashboard_project_update(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_dashboard.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { preCheckDashboard(t) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkDashboardDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclDashboardProjectBasic(name),
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExist(name),
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttrSet(tfNode, "name"),
					resource.TestCheckResourceAttrSet(tfNode, "owner_id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateIdFunc: importDashboardId(tfNode),
				ImportStateVerify: true,
			},
			{
				Config: hclDashboardProjectUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExist(name+"update"),
					resource.TestCheckResourceAttr(tfNode, "name", name+"update"),
					resource.TestCheckResourceAttr(tfNode, "description", "description"),
					resource.TestCheckResourceAttr(tfNode, "refresh_interval", "5"),
				),
			},
			// Idempotency after update.
			{
				Config:             hclDashboardProjectUpdate(name),
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateIdFunc: importDashboardId(tfNode),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDashboard_project_complete(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_dashboard.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { preCheckDashboard(t) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkDashboardDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclDashboardProjectComplete(name),
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExist(name),
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttrSet(tfNode, "name"),
					resource.TestCheckResourceAttrSet(tfNode, "owner_id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateIdFunc: importDashboardId(tfNode),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDashboard_team_basic(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_dashboard.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { preCheckDashboard(t) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkDashboardDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclDashboardTeamBasic(name),
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExist(name),
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttrSet(tfNode, "name"),
					resource.TestCheckResourceAttrSet(tfNode, "owner_id"),
				),
			},
			// Idempotency.
			{
				Config:             hclDashboardTeamBasic(name),
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateIdFunc: importDashboardId(tfNode),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDashboard_team_update(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_dashboard.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { preCheckDashboard(t) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkDashboardDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclDashboardTeamBasic(name),
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExist(name),
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttrSet(tfNode, "name"),
					resource.TestCheckResourceAttrSet(tfNode, "team_id"),
					resource.TestCheckResourceAttrSet(tfNode, "owner_id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateIdFunc: importDashboardId(tfNode),
				ImportStateVerify: true,
			},
			{
				Config: hclDashboardTeamUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExist(name+"update"),
					resource.TestCheckResourceAttr(tfNode, "name", name+"update"),
					resource.TestCheckResourceAttrSet(tfNode, "name"),
					resource.TestCheckResourceAttrSet(tfNode, "team_id"),
					resource.TestCheckResourceAttrSet(tfNode, "owner_id"),
					resource.TestCheckResourceAttr(tfNode, "description", "description"),
					resource.TestCheckResourceAttr(tfNode, "refresh_interval", "5"),
				),
			},
			// Idempotency after update.
			{
				Config:             hclDashboardTeamUpdate(name),
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateIdFunc: importDashboardId(tfNode),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDashboard_team_complete(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_dashboard.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { preCheckDashboard(t) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkDashboardDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclDashboardTeamComplete(name),
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExist(name),
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttrSet(tfNode, "name"),
					resource.TestCheckResourceAttrSet(tfNode, "team_id"),
					resource.TestCheckResourceAttrSet(tfNode, "description"),
					resource.TestCheckResourceAttrSet(tfNode, "refresh_interval"),
					resource.TestCheckResourceAttrSet(tfNode, "owner_id"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateIdFunc: importDashboardId(tfNode),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDashboard_team_requireImportError(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_dashboard.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { preCheckDashboard(t) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkDashboardDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclDashboardTeamBasic(name),
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExist(name),
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttrSet(tfNode, "name"),
					resource.TestCheckResourceAttrSet(tfNode, "team_id"),
					resource.TestCheckResourceAttrSet(tfNode, "owner_id"),
				),
			},
			{
				ExpectError: regexp.MustCompile("Creating Dashboard in Azure DevOps: VS403345: Duplicate name on dashboard. Each dashboard held by a team must use a distinct name"),
				Config:      hclDashboardTeamRequireImport(name),
			},
		},
	})
}

func checkDashboardDestroyed(s *terraform.State) error {
	clients, err := getDirectDashboardClient()
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	// verify that every dashboard referenced in the state does not exist in AzDO
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "betterado_dashboard" {
			continue
		}

		id := resource.Primary.ID
		dashboardId, err := uuid.Parse(id)
		if err != nil {
			return fmt.Errorf("Parsing dashboard ID: %+v", err)
		}

		args := dashboard.GetDashboardArgs{
			Project:     converter.String(resource.Primary.Attributes["project_id"]),
			DashboardId: &dashboardId,
		}

		if v, ok := resource.Primary.Attributes["team_id"]; ok && v != "" {
			args.Team = converter.String(v)
		}
		// indicates the dashboard still exists - this should fail the test
		if _, err = clients.DashboardClient.GetDashboard(clients.Ctx, args); err == nil {
			return fmt.Errorf("Dashboard with ID %s should not exist", id)
		}
	}
	return nil
}

func checkDashboardExist(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources["betterado_dashboard.test"]
		if !ok {
			return fmt.Errorf("Did not find a `betterado_dashboard` in the Terraform state")
		}

		clients, err := getDirectDashboardClient()
		if err != nil {
			return fmt.Errorf("getting client: %w", err)
		}

		id := res.Primary.ID
		dashboardId, err := uuid.Parse(id)
		if err != nil {
			return fmt.Errorf("Parsing dashboard ID %s", id)
		}

		args := dashboard.GetDashboardArgs{
			Project:     converter.String(res.Primary.Attributes["project_id"]),
			DashboardId: &dashboardId,
		}

		if v, ok := res.Primary.Attributes["team_id"]; ok && v != "" {
			args.Team = converter.String(v)
		}

		resp, err := clients.DashboardClient.GetDashboard(clients.Ctx, args)
		if err != nil {
			return fmt.Errorf("Dashboard with ID: %s cannot be found!. Error: %v", id, err)
		}

		if *resp.Name != expectedName {
			return fmt.Errorf("Dashboard with ID: %s has Name: %s, but expected Name: %s", id, *resp.Name, expectedName)
		}
		return nil
	}
}

func importDashboardId(resourceType string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		if res, ok := s.RootModule().Resources[resourceType]; ok {
			projectId := res.Primary.Attributes["project_id"]
			if teamId, ok := res.Primary.Attributes["team_id"]; ok && teamId != "" {
				return fmt.Sprintf("%s/%s/%s", projectId, teamId, res.Primary.ID), nil
			}
			return fmt.Sprintf("%s/%s", projectId, res.Primary.ID), nil
		}
		return "", fmt.Errorf("Not found: %s", resourceType)
	}
}

// hclDashboardProjectBasic uses the shared persistent project (SharedFixtureProjectName)
// as a data source instead of creating a new project, to avoid the ADO 1000-project cap.
func hclDashboardProjectBasic(name string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_dashboard" "test" {
  project_id = data.betterado_project.test.id
  name       = "%[1]s"
}
`, name, SharedFixtureProjectName)
}

func hclDashboardProjectUpdate(name string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_dashboard" "test" {
  project_id       = data.betterado_project.test.id
  name             = "%[1]supdate"
  description      = "description"
  refresh_interval = 5
}
`, name, SharedFixtureProjectName)
}

func hclDashboardProjectComplete(name string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_dashboard" "test" {
  project_id       = data.betterado_project.test.id
  name             = "%[1]s"
  description      = "description"
  refresh_interval = 5
}
`, name, SharedFixtureProjectName)
}

func hclDashboardTeamBasic(name string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_team" "test" {
  project_id = data.betterado_project.test.id
  name       = "%[1]s dashboard"
}

resource "betterado_dashboard" "test" {
  project_id = data.betterado_project.test.id
  name       = "%[1]s"
  team_id    = betterado_team.test.id
}
`, name, SharedFixtureProjectName)
}

func hclDashboardTeamUpdate(name string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_team" "test" {
  project_id = data.betterado_project.test.id
  name       = "%[1]s dashboard"
}

resource "betterado_dashboard" "test" {
  project_id       = data.betterado_project.test.id
  name             = "%[1]supdate"
  team_id          = betterado_team.test.id
  description      = "description"
  refresh_interval = 5
}
`, name, SharedFixtureProjectName)
}

func hclDashboardTeamComplete(name string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_team" "test" {
  project_id = data.betterado_project.test.id
  name       = "%[1]s dashboard"
}

resource "betterado_dashboard" "test" {
  project_id       = data.betterado_project.test.id
  name             = "%[1]s"
  team_id          = betterado_team.test.id
  description      = "description"
  refresh_interval = 5
}
`, name, SharedFixtureProjectName)
}

func hclDashboardTeamRequireImport(name string) string {
	return fmt.Sprintf(`
%s

resource "betterado_dashboard" "import" {
  project_id = betterado_dashboard.test.project_id
  name       = betterado_dashboard.test.name
  team_id    = betterado_dashboard.test.team_id
}
`, hclDashboardTeamBasic(name))
}
