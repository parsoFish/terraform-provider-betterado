//go:build (all || resource_extension_install) && !exclude_resource_extension_install

package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/extensionmanagement"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

func TestAccExtensionInstall_basic(t *testing.T) {
	publisherID := "ms-securitydevops"
	extensionID := "microsoft-security-devops-azdevops"
	tfNode := "betterado_extension_install.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxProviderFactories(),
		CheckDestroy:             checkExtensionInstallDestroyed,
		Steps: []resource.TestStep{
			// Step 1: apply and read-back with live evidence
			{
				Config: hclExtensionInstallBasic(publisherID, extensionID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "publisher_id"),
					resource.TestCheckResourceAttrSet(tfNode, "extension_id"),
					resource.TestCheckResourceAttrSet(tfNode, "version"),
					captureExtensionInstallEvidence(tfNode, publisherID, extensionID),
				),
			},
			// Step 2: idempotency — re-plan must produce an empty plan
			{
				Config:             hclExtensionInstallBasic(publisherID, extensionID),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclExtensionInstallBasic returns the HCL fixture for TestAccExtensionInstall_basic.
func hclExtensionInstallBasic(publisherID, extensionID string) string {
	return fmt.Sprintf(`
resource "betterado_extension_install" "test" {
  publisher_id = %q
  extension_id = %q
}
`, publisherID, extensionID)
}

// checkExtensionInstallDestroyed verifies that the extension was uninstalled after destroy.
// Uses getDirectClient() (defined in resource_task_group_test.go) because
// ProtoV6ProviderFactories does not wire the SDKv2 singleton Meta.
func checkExtensionInstallDestroyed(s *terraform.State) error {
	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_extension_install" {
			continue
		}

		parts := strings.SplitN(res.Primary.ID, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("checkExtensionInstallDestroyed: unexpected resource ID format %q", res.Primary.ID)
		}
		pubID, extID := parts[0], parts[1]

		directClient, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("checkExtensionInstallDestroyed: failed to build ADO client: %v", err)
		}

		ext, err := directClient.ExtensionManagementClient.GetInstalledExtensionByName(directClient.Ctx,
			extensionmanagement.GetInstalledExtensionByNameArgs{
				PublisherName: &pubID,
				ExtensionName: &extID,
			})
		if err != nil {
			// A 404-class error means the extension is gone — expected.
			return nil
		}
		if ext != nil {
			return fmt.Errorf("extension %s/%s still exists after destroy", pubID, extID)
		}
	}
	return nil
}

// captureExtensionInstallEvidence performs a real live API GET of the installed
// extension and persists the response as forge demo live-evidence via
// testutils.CaptureLiveEvidence. Best-effort: a capture failure never fails the test.
func captureExtensionInstallEvidence(tfNode, publisherID, extensionID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}

		directClient, err := buildExtensionDirectClient()
		if err != nil {
			return nil // best-effort
		}

		ext, err := directClient.ExtensionManagementClient.GetInstalledExtensionByName(directClient.Ctx,
			extensionmanagement.GetInstalledExtensionByNameArgs{
				PublisherName: &publisherID,
				ExtensionName: &extensionID,
			})
		if err != nil || ext == nil {
			return nil // best-effort
		}

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf(
			"%s/_apis/extensionmanagement/installedextensionsbyname/%s/%s?api-version=7.1",
			orgURL, publisherID, extensionID,
		)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-extension-install", url, ext)
		return nil
	}
}

// buildExtensionDirectClient builds an AggregatedClient directly from AZDO env vars.
// Separate from getDirectClient() (which is defined in resource_task_group_test.go)
// to avoid any naming collision risk — both produce the same result.
func buildExtensionDirectClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}
