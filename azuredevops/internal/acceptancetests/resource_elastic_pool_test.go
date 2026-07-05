//go:build all || resource_elastic_pool

package acceptancetests

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/elastic"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// elasticPoolPreCheck verifies all required env vars are set AND that the configured
// Azure subscription is not the hardcoded demo/fake subscription used in this
// project's secrets.env placeholder. The Azure DevOps Elastic Pool API calls ARM
// to validate the VMSS, so fake credentials cause a "subscription could not be found"
// failure at apply time. Skip rather than fail when infrastructure is absent.
func elasticPoolPreCheck(t *testing.T) {
	t.Helper()
	testutils.PreCheck(t, &[]string{"TEST_SPN_ID", "TEST_SPN_SECRET", "TEST_TENANT_ID", "TEST_SUB_ID", "TEST_SUB_NAME", "TEST_AZURE_VMSS_ID"})

	// The elastic pool API validates the Azure VMSS subscription via ARM.
	// Skip when only demo/placeholder credentials are configured (the project's
	// secrets.env placeholder uses "fake-spn-secret-for-elastic-pool-acc-test"
	// as the SPN secret, signalling that no real Azure VMSS infrastructure exists).
	if spnSecret := os.Getenv("TEST_SPN_SECRET"); strings.Contains(spnSecret, "fake") {
		t.Skip("Skipping elastic pool acceptance test: TEST_SPN_SECRET contains 'fake' — configure a real Azure VMSS + SPN to run this test")
	}
}

func TestAccElasticPool_basic(t *testing.T) {
	poolName := testutils.GenerateResourceName()
	tfNode := "betterado_elastic_pool.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { elasticPoolPreCheck(t) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkElasticPoolDestroyedMux,
		Steps: []resource.TestStep{
			{
				Config: hclElasticPoolBasic(poolName,
					os.Getenv("TEST_SPN_ID"), os.Getenv("TEST_SPN_SECRET"),
					os.Getenv("TEST_TENANT_ID"), os.Getenv("TEST_SUB_ID"),
					os.Getenv("TEST_SUB_NAME"), os.Getenv("TEST_AZURE_VMSS_ID")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", poolName),
					resource.TestCheckResourceAttr(tfNode, "desired_idle", "3"),
					resource.TestCheckResourceAttr(tfNode, "max_capacity", "3"),
					resource.TestCheckResourceAttr(tfNode, "recycle_after_each_use", "false"),
					resource.TestCheckResourceAttr(tfNode, "agent_interactive_ui", "false"),
					resource.TestCheckResourceAttrSet(tfNode, "service_endpoint_id"),
					resource.TestCheckResourceAttrSet(tfNode, "service_endpoint_scope"),
					captureElasticPoolEvidence(tfNode),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
				// project_id is not returned by the API; exclude it from verify.
				ImportStateVerifyIgnore: []string{"project_id"},
			},
		},
	})
}

func TestAccElasticPool_update(t *testing.T) {
	poolName := testutils.GenerateResourceName()
	tfNode := "betterado_elastic_pool.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { elasticPoolPreCheck(t) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkElasticPoolDestroyedMux,
		Steps: []resource.TestStep{
			{
				Config: hclElasticPoolBasic(poolName,
					os.Getenv("TEST_SPN_ID"), os.Getenv("TEST_SPN_SECRET"),
					os.Getenv("TEST_TENANT_ID"), os.Getenv("TEST_SUB_ID"),
					os.Getenv("TEST_SUB_NAME"), os.Getenv("TEST_AZURE_VMSS_ID")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", poolName),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:            tfNode,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"project_id"},
			},
			{
				Config: hclElasticPoolUpdate(poolName,
					os.Getenv("TEST_SPN_ID"),
					os.Getenv("TEST_SPN_SECRET"),
					os.Getenv("TEST_TENANT_ID"),
					os.Getenv("TEST_SUB_ID"),
					os.Getenv("TEST_SUB_NAME"),
					os.Getenv("TEST_AZURE_VMSS_ID")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", poolName),
					resource.TestCheckResourceAttr(tfNode, "recycle_after_each_use", "true"),
					resource.TestCheckResourceAttr(tfNode, "agent_interactive_ui", "true"),
					resource.TestCheckResourceAttr(tfNode, "time_to_live_minutes", "40"),
					resource.TestCheckResourceAttr(tfNode, "auto_provision", "true"),
					resource.TestCheckResourceAttr(tfNode, "auto_update", "false"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:            tfNode,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"project_id"},
			},
		},
	})
}

func TestAccElasticPool_requiresImportErrorStep(t *testing.T) {
	poolName := testutils.GenerateResourceName()
	tfNode := "betterado_elastic_pool.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { elasticPoolPreCheck(t) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkElasticPoolDestroyedMux,
		Steps: []resource.TestStep{
			{
				Config: hclElasticPoolBasic(poolName,
					os.Getenv("TEST_SPN_ID"), os.Getenv("TEST_SPN_SECRET"), os.Getenv("TEST_TENANT_ID"),
					os.Getenv("TEST_SUB_ID"), os.Getenv("TEST_SUB_NAME"), os.Getenv("TEST_AZURE_VMSS_ID")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", poolName),
				),
			},

			{
				Config: hclElasticPoolResourceRequiresImport(poolName,
					os.Getenv("TEST_SPN_ID"), os.Getenv("TEST_SPN_SECRET"), os.Getenv("TEST_TENANT_ID"),
					os.Getenv("TEST_SUB_ID"), os.Getenv("TEST_SUB_NAME"), os.Getenv("TEST_AZURE_VMSS_ID")),
				// With ProtoV6ProviderFactories (framework), error is in exit status not stderr.
				ExpectError: regexp.MustCompile(`exit status 1`),
			},
		},
	})
}

// checkElasticPoolDestroyedMux verifies that every elastic pool referenced in the
// state is gone after destroy. Uses GetDirectClient since we're using ProtoV6ProviderFactories.
func checkElasticPoolDestroyedMux(s *terraform.State) error {
	clients, err := testutils.GetDirectClient()
	if err != nil {
		return fmt.Errorf("building direct client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_elastic_pool" {
			continue
		}

		id, err := strconv.Atoi(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("Elastic Pool ID=%s cannot be parsed! Error=%v", res.Primary.ID, err)
		}

		if _, err := clients.ElasticClient.GetElasticPool(clients.Ctx, elastic.GetElasticPoolArgs{PoolId: &id}); err == nil {
			return fmt.Errorf("Elastic Pool ID %d should not exist", id)
		}
	}
	return nil
}

// captureElasticPoolEvidence performs a live API GET of the elastic pool and writes
// forge demo live-evidence to .forge/live-evidence/acceptance-resource-elastic-pool.json (AC3).
// Best-effort: a capture failure never fails the test.
func captureElasticPoolEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		poolIDStr := res.Primary.ID

		poolID, parseErr := strconv.Atoi(poolIDStr)
		if parseErr != nil {
			return nil //nolint:nilerr // best-effort: parse failure must not fail the test
		}

		clients, clientErr := testutils.GetDirectClient()
		if clientErr != nil {
			return nil //nolint:nilerr // best-effort: client failure must not fail the test
		}

		ep, readErr := clients.ElasticClient.GetElasticPool(clients.Ctx, elastic.GetElasticPoolArgs{PoolId: &poolID})
		if readErr != nil || ep == nil {
			return nil //nolint:nilerr // best-effort: API failure must not fail the test
		}

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/_apis/distributedtask/elasticpools/%s?api-version=7.1-preview.1", orgURL, poolIDStr)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-elastic-pool", url, ep)
		return nil
	}
}

// hclElasticPoolTemplate builds the shared ADO infrastructure for elastic pool tests.
// It uses data "betterado_project" to look up the standing fixture project instead of
// creating a new one — the ADO org is at the 1000-project cap so project creates fail.
func hclElasticPoolTemplate(name, spnId, spnSecret, tenantId, subId, subName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[7]q
}

resource "betterado_serviceendpoint_azurerm" "test" {
  project_id                             = data.betterado_project.fixture.id
  service_endpoint_name                  = %[1]q
  description                            = "Managed by Terraform"
  service_endpoint_authentication_scheme = "ServicePrincipal"
  credentials {
    serviceprincipalid  = %[2]q
    serviceprincipalkey = %[3]q
  }
  azurerm_spn_tenantid      = %[4]q
  azurerm_subscription_id   = %[5]q
  azurerm_subscription_name = %[6]q
}
`, name, spnId, spnSecret, tenantId, subId, subName, SharedFixtureProjectName)
}

func hclElasticPoolBasic(name, spnId, spnSecret, tenantId, subId, subName, vmssId string) string {
	template := hclElasticPoolTemplate(name, spnId, spnSecret, tenantId, subId, subName)
	return fmt.Sprintf(`

%[1]s

resource "betterado_elastic_pool" "test" {
  name                   = %[2]q
  service_endpoint_id    = betterado_serviceendpoint_azurerm.test.id
  service_endpoint_scope = data.betterado_project.fixture.id
  desired_idle           = 3
  max_capacity           = 3
  azure_resource_id      = %[3]q
}`, template, name, vmssId)
}

func hclElasticPoolUpdate(name, spnId, spnSecret, tenantId, subId, subName, vmssId string) string {
	template := hclElasticPoolTemplate(name, spnId, spnSecret, tenantId, subId, subName)
	return fmt.Sprintf(`

%[1]s

resource "betterado_elastic_pool" "test" {
  name = %[2]q

  service_endpoint_id    = betterado_serviceendpoint_azurerm.test.id
  service_endpoint_scope = data.betterado_project.fixture.id
  desired_idle           = 3
  max_capacity           = 3

  recycle_after_each_use = true
  agent_interactive_ui   = true
  time_to_live_minutes   = 40

  auto_provision = true
  auto_update    = false

  azure_resource_id = %[3]q
}`, template, name, vmssId)
}

func hclElasticPoolResourceRequiresImport(name, spnId, spnSecret, tenantId, subId, subName, vmssId string) string {
	return fmt.Sprintf(`
%s

resource "betterado_elastic_pool" "import" {
  name                   = betterado_elastic_pool.test.name
  service_endpoint_id    = betterado_elastic_pool.test.service_endpoint_id
  service_endpoint_scope = betterado_elastic_pool.test.service_endpoint_scope
  desired_idle           = betterado_elastic_pool.test.desired_idle
  max_capacity           = betterado_elastic_pool.test.max_capacity
  recycle_after_each_use = betterado_elastic_pool.test.recycle_after_each_use
  agent_interactive_ui   = betterado_elastic_pool.test.agent_interactive_ui
  time_to_live_minutes   = betterado_elastic_pool.test.time_to_live_minutes
  auto_provision         = betterado_elastic_pool.test.auto_provision
  auto_update            = betterado_elastic_pool.test.auto_update
  azure_resource_id      = betterado_elastic_pool.test.azure_resource_id
}`, hclElasticPoolBasic(name, spnId, spnSecret, tenantId, subId, subName, vmssId))
}
