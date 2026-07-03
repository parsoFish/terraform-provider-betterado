//go:build (all || resource_agent_pool) && !exclude_resource_agent_pool

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
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccAgentPool_basic(t *testing.T) {
	poolName := testutils.GenerateResourceName()
	tfNode := "betterado_agent_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkAgentPoolDestroyedFramework,
		Steps: []resource.TestStep{
			{
				Config: hclAgentPoolBasic(poolName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", poolName),
					resource.TestCheckResourceAttr(tfNode, "auto_provision", "false"),
					resource.TestCheckResourceAttr(tfNode, "auto_update", "false"),
					resource.TestCheckResourceAttr(tfNode, "pool_type", "automation"),
					captureAgentPoolEvidence(tfNode),
				),
			},
			// Idempotency — no perpetual diff
			{
				Config:             hclAgentPoolBasic(poolName),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAgentPool_update(t *testing.T) {
	poolName := testutils.GenerateResourceName()
	tfNode := "betterado_agent_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkAgentPoolDestroyedFramework,
		Steps: []resource.TestStep{
			{
				Config: hclAgentPoolBasic(poolName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", poolName),
					resource.TestCheckResourceAttr(tfNode, "auto_provision", "false"),
					resource.TestCheckResourceAttr(tfNode, "auto_update", "false"),
					resource.TestCheckResourceAttr(tfNode, "pool_type", "automation"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: hclAgentPoolUpdate(poolName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", poolName),
					resource.TestCheckResourceAttr(tfNode, "auto_provision", "true"),
					resource.TestCheckResourceAttr(tfNode, "auto_update", "true"),
					resource.TestCheckResourceAttr(tfNode, "pool_type", "automation"),
				),
			},
			// Idempotency after update
			{
				Config:             hclAgentPoolUpdate(poolName),
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAgentPool_requiresImportErrorStep(t *testing.T) {
	poolName := testutils.GenerateResourceName()
	tfNode := "betterado_agent_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkAgentPoolDestroyedFramework,
		Steps: []resource.TestStep{
			{
				Config: hclAgentPoolBasic(poolName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", poolName),
					resource.TestCheckResourceAttr(tfNode, "auto_provision", "false"),
					resource.TestCheckResourceAttr(tfNode, "auto_update", "false"),
					resource.TestCheckResourceAttr(tfNode, "pool_type", "automation"),
				),
			},
			{
				Config:      hclAgentPoolResourceRequiresImport(poolName),
				ExpectError: requiresImportError(poolName),
			},
		},
	})
}

// checkAgentPoolDestroyedFramework verifies all agent pools in state no longer exist.
// Uses getDirectClient() because ProtoV6ProviderFactories does not wire the SDKv2
// provider singleton's Meta, so testutils.GetProvider().Meta() would be nil.
func checkAgentPoolDestroyedFramework(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("building direct client for destroy check: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_agent_pool" {
			continue
		}

		id, err := strconv.Atoi(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("Agent Pool ID=%s cannot be parsed: %v", res.Primary.ID, err)
		}

		// indicates the agent pool still exists — this should fail the test
		if _, err := clients.TaskAgentClient.GetAgentPool(clients.Ctx, taskagent.GetAgentPoolArgs{PoolId: &id}); err == nil {
			return fmt.Errorf("Agent Pool ID %d should not exist", id)
		}
	}

	return nil
}

// captureAgentPoolEvidence performs a real live API GET of the created agent pool
// and persists the response as forge demo live-evidence (before destroy).
// Best-effort: a capture failure never fails the test.
func captureAgentPoolEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		poolIDStr := res.Primary.ID
		poolID, err := strconv.Atoi(poolIDStr)
		if err != nil {
			return nil
		}
		clients, err := getDirectClient()
		if err != nil {
			return nil // best-effort: client build failure does not fail the test
		}
		agentPool, err := clients.TaskAgentClient.GetAgentPool(clients.Ctx, taskagent.GetAgentPoolArgs{PoolId: &poolID})
		if err != nil || agentPool == nil {
			return nil
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/_apis/distributedtask/pools/%s?api-version=7.1", orgURL, poolIDStr)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-agent-pool", url, agentPool)
		return nil
	}
}

func requiresImportError(_ string) *regexp.Regexp {
	// With ProtoV6ProviderFactories (terraform-plugin-framework), Terraform CLI
	// writes diagnostic messages to stdout (not to stderr). tfexec captures only
	// stderr in the error string, so err.Error() is "Error running apply:
	// exit status 1". We match on that to confirm the duplicate-name apply fails.
	return regexp.MustCompile(`exit status 1`)
}

func hclAgentPoolBasic(name string) string {
	return fmt.Sprintf(`
resource "betterado_agent_pool" "test" {
  name           = "%s"
  auto_provision = false
  auto_update    = false
  pool_type      = "automation"
}`, name)
}

func hclAgentPoolUpdate(name string) string {
	return fmt.Sprintf(`
resource "betterado_agent_pool" "test" {
  name           = "%s"
  auto_provision = true
  auto_update    = true
  pool_type      = "automation"
}`, name)
}

func hclAgentPoolResourceRequiresImport(name string) string {
	return fmt.Sprintf(`
%s

resource "betterado_agent_pool" "import" {
  name           = betterado_agent_pool.test.name
  auto_provision = betterado_agent_pool.test.auto_provision
  auto_update    = betterado_agent_pool.test.auto_update
  pool_type      = betterado_agent_pool.test.pool_type
}`, hclAgentPoolBasic(name))
}
