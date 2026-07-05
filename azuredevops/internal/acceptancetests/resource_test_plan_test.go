//go:build (all || resource_test_plan) && !exclude_resource_test_plan

package acceptancetests

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/testplan"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// ── TestAccTestPlan_basic ──────────────────────────────────────────────────────

func TestAccTestPlan_basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_test_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkTestPlanDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclTestPlanBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					resource.TestCheckResourceAttrSet(tfNode, "root_suite_id"),
					captureTestPlanEvidence(tfNode),
				),
			},
			{Config: hclTestPlanBasic(name), PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}

// ── TestAccTestSuite_basic ─────────────────────────────────────────────────────

func TestAccTestSuite_basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfPlan := "betterado_test_plan.suite_parent"
	tfSuite := "betterado_test_suite.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkTestSuiteDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclTestSuiteBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfSuite, "name", name+"-suite"),
					resource.TestCheckResourceAttrSet(tfSuite, "id"),
					resource.TestCheckResourceAttrSet(tfPlan, "root_suite_id"),
					captureTestSuiteEvidence(tfSuite),
				),
			},
			{Config: hclTestSuiteBasic(name), PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}

// ── TestAccTestConfiguration_basic ───────────────────────────────────────────

func TestAccTestConfiguration_basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_test_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkTestConfigurationDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclTestConfigurationBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					captureTestConfigurationEvidence(tfNode),
				),
			},
			{Config: hclTestConfigurationBasic(name), PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}

// ── TestAccTestVariable_basic ─────────────────────────────────────────────────

func TestAccTestVariable_basic(t *testing.T) {
	name := testutils.GenerateResourceName()
	tfNode := "betterado_test_variable.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkTestVariableDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclTestVariableBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", name),
					resource.TestCheckResourceAttrSet(tfNode, "id"),
					captureTestVariableEvidence(tfNode),
				),
			},
			{Config: hclTestVariableBasic(name), PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}

// ── HCL configs ───────────────────────────────────────────────────────────────

func hclTestPlanBasic(name string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_test_plan" "test" {
  project_id = data.betterado_project.test.id
  name       = %[1]q
}
`, name, SharedFixtureProjectName)
}

func hclTestSuiteBasic(name string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_test_plan" "suite_parent" {
  project_id = data.betterado_project.test.id
  name       = %[1]q
}

resource "betterado_test_suite" "test" {
  project_id      = data.betterado_project.test.id
  plan_id         = betterado_test_plan.suite_parent.id
  parent_suite_id = betterado_test_plan.suite_parent.root_suite_id
  name            = "%[1]s-suite"
  suite_type      = "staticTestSuite"
}
`, name, SharedFixtureProjectName)
}

func hclTestConfigurationBasic(name string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_test_configuration" "test" {
  project_id = data.betterado_project.test.id
  name       = %[1]q
  values     = { "Browser" = "Chrome" }
}
`, name, SharedFixtureProjectName)
}

func hclTestVariableBasic(name string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_test_variable" "test" {
  project_id     = data.betterado_project.test.id
  name           = %[1]q
  allowed_values = ["value-a", "value-b"]
}
`, name, SharedFixtureProjectName)
}

// ── CheckDestroy functions ────────────────────────────────────────────────────

func checkTestPlanDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("checkTestPlanDestroyed: failed to build ADO client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_test_plan" {
			continue
		}
		planID, err := strconv.Atoi(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("test plan ID %q cannot be parsed: %v", res.Primary.ID, err)
		}
		projectID := res.Primary.Attributes["project_id"]

		plan, err := clients.TestPlanClient.GetTestPlanById(clients.Ctx, testplan.GetTestPlanByIdArgs{
			Project: &projectID,
			PlanId:  &planID,
		})
		if err != nil {
			// 404 means it's gone, which is what we want.
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
				continue
			}
			return fmt.Errorf("error reading test plan %d after destroy: %v", planID, err)
		}
		if plan != nil {
			return fmt.Errorf("test plan %d still exists after destroy", planID)
		}
	}

	return nil
}

func checkTestSuiteDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("checkTestSuiteDestroyed: failed to build ADO client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_test_suite" {
			continue
		}
		suiteID, err := strconv.Atoi(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("test suite ID %q cannot be parsed: %v", res.Primary.ID, err)
		}
		projectID := res.Primary.Attributes["project_id"]
		planIDStr := res.Primary.Attributes["plan_id"]
		planID, err := strconv.Atoi(planIDStr)
		if err != nil {
			return fmt.Errorf("test suite plan_id %q cannot be parsed: %v", planIDStr, err)
		}

		suite, err := clients.TestPlanClient.GetTestSuiteById(clients.Ctx, testplan.GetTestSuiteByIdArgs{
			Project: &projectID,
			PlanId:  &planID,
			SuiteId: &suiteID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
				continue
			}
			return fmt.Errorf("error reading test suite %d after destroy: %v", suiteID, err)
		}
		if suite != nil {
			return fmt.Errorf("test suite %d still exists after destroy", suiteID)
		}
	}

	return nil
}

func checkTestConfigurationDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("checkTestConfigurationDestroyed: failed to build ADO client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_test_configuration" {
			continue
		}
		cfgID, err := strconv.Atoi(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("test configuration ID %q cannot be parsed: %v", res.Primary.ID, err)
		}
		projectID := res.Primary.Attributes["project_id"]

		cfg, err := clients.TestPlanClient.GetTestConfigurationById(clients.Ctx, testplan.GetTestConfigurationByIdArgs{
			Project:             &projectID,
			TestConfigurationId: &cfgID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
				continue
			}
			return fmt.Errorf("error reading test configuration %d after destroy: %v", cfgID, err)
		}
		if cfg != nil {
			return fmt.Errorf("test configuration %d still exists after destroy", cfgID)
		}
	}

	return nil
}

func checkTestVariableDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("checkTestVariableDestroyed: failed to build ADO client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_test_variable" {
			continue
		}
		varID, err := strconv.Atoi(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("test variable ID %q cannot be parsed: %v", res.Primary.ID, err)
		}
		projectID := res.Primary.Attributes["project_id"]

		v, err := clients.TestPlanClient.GetTestVariableById(clients.Ctx, testplan.GetTestVariableByIdArgs{
			Project:        &projectID,
			TestVariableId: &varID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
				continue
			}
			return fmt.Errorf("error reading test variable %d after destroy: %v", varID, err)
		}
		if v != nil {
			return fmt.Errorf("test variable %d still exists after destroy", varID)
		}
	}

	return nil
}

// ── Evidence capture helpers ──────────────────────────────────────────────────

// captureTestPlanEvidence performs a live API GET of the created test plan and
// persists the response as forge demo evidence before destroy.
// Uses a per-type label ("acceptance-test-plan") so it never overwrites captures
// from other resource types. Errors from CaptureLiveEvidence are surfaced so
// the gate can detect a failed capture rather than silently missing evidence.
func captureTestPlanEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return fmt.Errorf("captureTestPlanEvidence: resource %s not found in state", tfNode)
		}
		planIDStr := res.Primary.ID
		planID, err := strconv.Atoi(planIDStr)
		if err != nil {
			return fmt.Errorf("captureTestPlanEvidence: parse plan ID %q: %w", planIDStr, err)
		}
		projectID := res.Primary.Attributes["project_id"]
		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("captureTestPlanEvidence: get direct client: %w", err)
		}
		plan, err := clients.TestPlanClient.GetTestPlanById(clients.Ctx, testplan.GetTestPlanByIdArgs{
			Project: &projectID,
			PlanId:  &planID,
		})
		if err != nil {
			return fmt.Errorf("captureTestPlanEvidence: GetTestPlanById(%d): %w", planID, err)
		}
		if plan == nil {
			return fmt.Errorf("captureTestPlanEvidence: GetTestPlanById(%d) returned nil", planID)
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/%s/_apis/testplan/plans/%s?api-version=7.1", orgURL, projectID, planIDStr)
		if err := testutils.CaptureLiveEvidence("acceptance-test-plan", url, plan); err != nil {
			return fmt.Errorf("captureTestPlanEvidence: CaptureLiveEvidence: %w", err)
		}
		return nil
	}
}

// captureTestSuiteEvidence performs a live API GET of the created test suite
// and persists the response as forge demo evidence before destroy.
// Uses a per-type label ("acceptance-test-suite") so it never overwrites captures
// from other resource types. Errors from CaptureLiveEvidence are surfaced.
func captureTestSuiteEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return fmt.Errorf("captureTestSuiteEvidence: resource %s not found in state", tfNode)
		}
		suiteIDStr := res.Primary.ID
		suiteID, err := strconv.Atoi(suiteIDStr)
		if err != nil {
			return fmt.Errorf("captureTestSuiteEvidence: parse suite ID %q: %w", suiteIDStr, err)
		}
		projectID := res.Primary.Attributes["project_id"]
		planIDStr := res.Primary.Attributes["plan_id"]
		planID, err := strconv.Atoi(planIDStr)
		if err != nil {
			return fmt.Errorf("captureTestSuiteEvidence: parse plan ID %q: %w", planIDStr, err)
		}
		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("captureTestSuiteEvidence: get direct client: %w", err)
		}
		suite, err := clients.TestPlanClient.GetTestSuiteById(clients.Ctx, testplan.GetTestSuiteByIdArgs{
			Project: &projectID,
			PlanId:  &planID,
			SuiteId: &suiteID,
		})
		if err != nil {
			return fmt.Errorf("captureTestSuiteEvidence: GetTestSuiteById(%d): %w", suiteID, err)
		}
		if suite == nil {
			return fmt.Errorf("captureTestSuiteEvidence: GetTestSuiteById(%d) returned nil", suiteID)
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/%s/_apis/testplan/plans/%s/suites/%s?api-version=7.1", orgURL, projectID, planIDStr, suiteIDStr)
		if err := testutils.CaptureLiveEvidence("acceptance-test-suite", url, suite); err != nil {
			return fmt.Errorf("captureTestSuiteEvidence: CaptureLiveEvidence: %w", err)
		}
		return nil
	}
}

// captureTestConfigurationEvidence performs a live API GET of the created test
// configuration and persists the response as forge demo evidence before destroy.
// Uses a per-type label ("acceptance-test-configuration") so it never overwrites
// captures from other resource types. Errors from CaptureLiveEvidence are surfaced.
func captureTestConfigurationEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return fmt.Errorf("captureTestConfigurationEvidence: resource %s not found in state", tfNode)
		}
		cfgIDStr := res.Primary.ID
		cfgID, err := strconv.Atoi(cfgIDStr)
		if err != nil {
			return fmt.Errorf("captureTestConfigurationEvidence: parse cfg ID %q: %w", cfgIDStr, err)
		}
		projectID := res.Primary.Attributes["project_id"]
		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("captureTestConfigurationEvidence: get direct client: %w", err)
		}
		cfg, err := clients.TestPlanClient.GetTestConfigurationById(clients.Ctx, testplan.GetTestConfigurationByIdArgs{
			Project:             &projectID,
			TestConfigurationId: &cfgID,
		})
		if err != nil {
			return fmt.Errorf("captureTestConfigurationEvidence: GetTestConfigurationById(%d): %w", cfgID, err)
		}
		if cfg == nil {
			return fmt.Errorf("captureTestConfigurationEvidence: GetTestConfigurationById(%d) returned nil", cfgID)
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/%s/_apis/test/configurations/%s?api-version=7.1", orgURL, projectID, cfgIDStr)
		if err := testutils.CaptureLiveEvidence("acceptance-test-configuration", url, cfg); err != nil {
			return fmt.Errorf("captureTestConfigurationEvidence: CaptureLiveEvidence: %w", err)
		}
		return nil
	}
}

// captureTestVariableEvidence performs a live API GET of the created test
// variable and persists the response as forge demo evidence before destroy.
// Uses a per-type label ("acceptance-test-variable") so it never overwrites
// captures from other resource types. Errors from CaptureLiveEvidence are surfaced.
func captureTestVariableEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return fmt.Errorf("captureTestVariableEvidence: resource %s not found in state", tfNode)
		}
		varIDStr := res.Primary.ID
		varID, err := strconv.Atoi(varIDStr)
		if err != nil {
			return fmt.Errorf("captureTestVariableEvidence: parse var ID %q: %w", varIDStr, err)
		}
		projectID := res.Primary.Attributes["project_id"]
		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("captureTestVariableEvidence: get direct client: %w", err)
		}
		v, err := clients.TestPlanClient.GetTestVariableById(clients.Ctx, testplan.GetTestVariableByIdArgs{
			Project:        &projectID,
			TestVariableId: &varID,
		})
		if err != nil {
			return fmt.Errorf("captureTestVariableEvidence: GetTestVariableById(%d): %w", varID, err)
		}
		if v == nil {
			return fmt.Errorf("captureTestVariableEvidence: GetTestVariableById(%d) returned nil", varID)
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/%s/_apis/testplan/variables/%s?api-version=7.1", orgURL, projectID, varIDStr)
		if err := testutils.CaptureLiveEvidence("acceptance-test-variable", url, v); err != nil {
			return fmt.Errorf("captureTestVariableEvidence: CaptureLiveEvidence: %w", err)
		}
		return nil
	}
}
