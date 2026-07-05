package acceptancetests

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

func TestAccProject_basic(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	tfNode := "betterado_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testutils.PreCheck(t, nil) },
		ProviderFactories: testutils.GetSDKv2ProviderFactories(),
		CheckDestroy:      testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclProjectBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					checkProjectExists(projectName),
					resource.TestCheckResourceAttrSet(tfNode, "process_template_id"),
					resource.TestCheckResourceAttr(tfNode, "name", projectName),
					resource.TestCheckResourceAttr(tfNode, "version_control", "Git"),
					resource.TestCheckResourceAttr(tfNode, "visibility", "private"),
					resource.TestCheckResourceAttr(tfNode, "work_item_template", "Agile"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  checkImportProject(),
			},
		},
	})
}

// TestAccProject_importByName exercises the framework betterado_project resource
// via import of the EXISTING betterado-standing-demo project. No new project is
// created — the ADO org is at the 1000-project soft-delete cap so any create
// attempt fails immediately. The test:
//
//  1. Imports betterado-standing-demo by name via the framework provider.
//  2. Asserts read-back attributes (name, visibility, version_control, work_item_template,
//     process_template_id) against known values for the standing-demo project.
//  3. Captures live evidence to .forge/live-evidence/acceptance-resource.json.
//  4. Re-plans with the same config (ExpectNonEmptyPlan: false — idempotency gate).
func TestAccProject_importByName(t *testing.T) {
	tfNode := "betterado_project.test"
	// betterado-standing-demo is the persistent shared fixture project that is
	// NEVER deleted (see shared_fixtures.go SharedFixtureProjectName).
	standingDemoName := SharedFixtureProjectName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		// No CheckDestroy: the standing-demo project must never be deleted.
		Steps: []resource.TestStep{
			// Step 1 — import by name; assert read-back + capture live evidence.
			// NOTE: ImportStateVerify is intentionally omitted here. ImportStateVerify
			// compares the imported state against the pre-import state, but since this
			// test imports an existing project (no prior terraform apply), there is no
			// pre-import state to compare against. The checkProjectImportByName function
			// verifies all required attributes (name, visibility, version_control,
			// work_item_template, process_template_id). Step 2's PlanOnly +
			// ExpectNonEmptyPlan: false verifies idempotency.
			// ImportStatePersist: true is required so that the imported state is written
			// to the main test working directory (testCaseWorkingDir) rather than a
			// temporary directory that is discarded after this step. Without this flag,
			// terraform-plugin-testing runs the import in a throw-away workingDir, the
			// state is never propagated back, and Step 2 plans against empty state,
			// showing "Plan: 1 to add" (false non-empty plan failure).
			{
				Config:             hclProjectStandingDemo(),
				ResourceName:       tfNode,
				ImportState:        true,
				ImportStateId:      standingDemoName,
				ImportStatePersist: true,
				ImportStateCheck:   checkProjectImportByName(standingDemoName),
			},
			// Step 2 — idempotency: re-plan must show no diff.
			{
				Config:             hclProjectStandingDemo(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			// Step 3 — forget, don't destroy. ImportStatePersist leaves the REAL
			// standing-demo project in the test state, and the harness always runs
			// terraform destroy on whatever state remains after the last step — on
			// 2026-07-02 that soft-deleted the live shared fixture project. A
			// `removed` block (terraform >= 1.7) drops the project from state
			// without touching the remote resource, so the post-test destroy has
			// nothing to delete.
			{
				Config: hclProjectRemovedFromState(),
			},
		},
	})
}

// hclProjectRemovedFromState removes betterado_project.test from state WITHOUT
// destroying the remote project (see Step 3 of TestAccProject_importByName).
func hclProjectRemovedFromState() string {
	return `
removed {
  from = betterado_project.test

  lifecycle {
    destroy = false
  }
}`
}

// hclProjectStandingDemo returns HCL for the standing betterado-standing-demo project
// with the known attribute values (no create — import only).
func hclProjectStandingDemo() string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = %q
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}`, SharedFixtureProjectName)
}

// checkProjectImportByName asserts read-back attributes for an imported project and
// captures live evidence. It is used as ImportStateCheck in TestAccProject_importByName.
func checkProjectImportByName(expectedName string) resource.ImportStateCheckFunc {
	return func(states []*terraform.InstanceState) error {
		if len(states) != 1 {
			return fmt.Errorf("expected 1 import state, got %d", len(states))
		}
		state := states[0]

		// Assert ID is a valid UUID.
		if err := uuid.Validate(state.ID); err != nil {
			return fmt.Errorf("project ID should be a valid UUID, got %q: %v", state.ID, err)
		}

		// Assert key attributes are set.
		if got := state.Attributes["name"]; got != expectedName {
			return fmt.Errorf("expected name=%q, got %q", expectedName, got)
		}
		if got := state.Attributes["visibility"]; got == "" {
			return fmt.Errorf("visibility attribute is not set after import")
		}
		if got := state.Attributes["version_control"]; got == "" {
			return fmt.Errorf("version_control attribute is not set after import")
		}
		if got := state.Attributes["work_item_template"]; got == "" {
			return fmt.Errorf("work_item_template attribute is not set after import")
		}
		if got := state.Attributes["process_template_id"]; got == "" {
			return fmt.Errorf("process_template_id attribute is not set after import")
		}

		// Capture live evidence (best-effort) — build client directly from env vars
		// rather than using getDirectClient() (defined in a build-tagged file) so
		// that this file compiles without any build tags.
		projectID := state.ID
		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
		if orgURL != "" && pat != "" {
			clients, err := client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
			if err == nil {
				project, err := clients.CoreClient.GetProject(clients.Ctx, core.GetProjectArgs{
					ProjectId:           &projectID,
					IncludeCapabilities: converter.Bool(true),
					IncludeHistory:      converter.Bool(false),
				})
				if err == nil {
					apiURL := fmt.Sprintf("%s/_apis/projects/%s?includeCapabilities=true&api-version=7.1",
						orgURL, projectID)
					_ = testutils.CaptureLiveEvidence("acceptance-resource", apiURL, project)
				}
			}
		}

		return nil
	}
}

func TestAccProject_update(t *testing.T) {
	projectNameFirst := testutils.GenerateResourceName()
	projectNameSecond := testutils.GenerateResourceName()
	tfNode := "betterado_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testutils.PreCheck(t, nil) },
		ProviderFactories: testutils.GetSDKv2ProviderFactories(),
		CheckDestroy:      testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclProjectBasic(projectNameFirst),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_template_id"),
					resource.TestCheckResourceAttr(tfNode, "name", projectNameFirst),
					resource.TestCheckResourceAttr(tfNode, "version_control", "Git"),
					resource.TestCheckResourceAttr(tfNode, "visibility", "private"),
					resource.TestCheckResourceAttr(tfNode, "work_item_template", "Agile"),
					checkProjectExists(projectNameFirst),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  checkImportProject(),
			},
			{
				Config: hclProjectUpdate(projectNameSecond),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_template_id"),
					resource.TestCheckResourceAttr(tfNode, "name", projectNameSecond),
					resource.TestCheckResourceAttr(tfNode, "version_control", "Git"),
					resource.TestCheckResourceAttr(tfNode, "visibility", "public"),
					resource.TestCheckResourceAttr(tfNode, "work_item_template", "Agile"),
					checkProjectExists(projectNameSecond),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  checkImportProject(),
			},
		},
	})
}

func TestAccProject_features(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	tfNode := "betterado_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testutils.PreCheck(t, nil) },
		ProviderFactories: testutils.GetSDKv2ProviderFactories(),
		CheckDestroy:      testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclProjectFeature(projectName, "disabled", "disabled"),
				Check: resource.ComposeTestCheckFunc(
					checkProjectExists(projectName),
					resource.TestCheckResourceAttrSet(tfNode, "process_template_id"),
					resource.TestCheckResourceAttr(tfNode, "name", projectName),
					resource.TestCheckResourceAttr(tfNode, "version_control", "Git"),
					resource.TestCheckResourceAttr(tfNode, "visibility", "private"),
					resource.TestCheckResourceAttr(tfNode, "work_item_template", "Agile"),
					resource.TestCheckResourceAttr(tfNode, "features.testplans", "disabled"),
					resource.TestCheckResourceAttr(tfNode, "features.artifacts", "disabled"),
				),
			},
			{
				ResourceName:            tfNode,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"features.testplans", "features.artifacts", "features.%"},
				ImportStateCheck:        checkImportProject(),
			},
			{
				Config: hclProjectFeature(projectName, "enabled", "disabled"),
				Check: resource.ComposeTestCheckFunc(
					checkProjectExists(projectName),
					resource.TestCheckResourceAttrSet(tfNode, "process_template_id"),
					resource.TestCheckResourceAttr(tfNode, "name", projectName),
					resource.TestCheckResourceAttr(tfNode, "version_control", "Git"),
					resource.TestCheckResourceAttr(tfNode, "visibility", "private"),
					resource.TestCheckResourceAttr(tfNode, "work_item_template", "Agile"),
					resource.TestCheckResourceAttr(tfNode, "features.testplans", "enabled"),
					resource.TestCheckResourceAttr(tfNode, "features.artifacts", "disabled"),
				),
			},
			{
				ResourceName:            tfNode,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"features.testplans", "features.artifacts", "features.%"},
				ImportStateCheck:        checkImportProject(),
			},
		},
	})
}

func TestAccProject_requireImportError(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	tfNode := "betterado_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testutils.PreCheck(t, nil) },
		ProviderFactories: testutils.GetSDKv2ProviderFactories(),
		CheckDestroy:      testutils.CheckProjectDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclProjectBasic(projectName),
				Check: resource.ComposeTestCheckFunc(
					checkProjectExists(projectName),
					resource.TestCheckResourceAttrSet(tfNode, "process_template_id"),
					resource.TestCheckResourceAttr(tfNode, "name", projectName),
					resource.TestCheckResourceAttr(tfNode, "version_control", "Git"),
					resource.TestCheckResourceAttr(tfNode, "visibility", "private"),
					resource.TestCheckResourceAttr(tfNode, "work_item_template", "Agile"),
				),
			},
			{
				Config:      hclProjectImport(projectName),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`Error:  creating project: TF200019: The following project already exists on the Azure DevOps Server: %s. You cannot create a new project with the same name as an existing project. Provide a different name`, projectName)),
			},
		},
	})
}

func checkProjectExists(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		state, ok := s.RootModule().Resources["betterado_project.test"]
		if !ok {
			return fmt.Errorf("Did not find a project in the TF state")
		}

		clients := testutils.GetProvider().Meta().(*client.AggregatedClient)
		id := state.Primary.ID
		project, err := clients.CoreClient.GetProject(clients.Ctx, core.GetProjectArgs{
			ProjectId:           &id,
			IncludeCapabilities: converter.Bool(true),
			IncludeHistory:      converter.Bool(false),
		})
		if err != nil {
			return fmt.Errorf("Project with ID=%s cannot be found!. Error=%v", id, err)
		}

		if *project.Name != expectedName {
			return fmt.Errorf("Project with ID=%s has Name=%s, but expected Name=%s", id, *project.Name, expectedName)
		}

		return nil
	}
}

func checkImportProject() resource.ImportStateCheckFunc {
	return func(states []*terraform.InstanceState) error {
		if len(states) != 1 {
			return fmt.Errorf("Expected project imported but not found in the state.")
		}
		state := states[0]
		projectId := state.ID
		if err := uuid.Validate(projectId); err != nil {
			return fmt.Errorf("Project ID should be a valid UUID. Got: %s", projectId)
		}
		return nil
	}
}

func hclProjectBasic(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  description        = "%[1]s-description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}`, name)
}

func hclProjectUpdate(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  description        = "%[1]s-description-update"
  visibility         = "public"
  version_control    = "Git"
  work_item_template = "Agile"
}`, name)
}

func hclProjectFeature(projectName, testPlans, stateArtifacts string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%s"
  description        = "%s-description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"

  features = {
    "testplans" = "%s"
    "artifacts" = "%s"
  }
}`, projectName, projectName, testPlans, stateArtifacts)
}

func hclProjectImport(name string) string {
	template := hclProjectBasic(name)
	return fmt.Sprintf(`
%s

resource "betterado_project" "import" {
  name               = betterado_project.test.name
  description        = betterado_project.test.description
  visibility         = betterado_project.test.visibility
  version_control    = betterado_project.test.version_control
  work_item_template = betterado_project.test.work_item_template
}`, template)
}
