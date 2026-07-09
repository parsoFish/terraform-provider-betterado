//go:build (all || resource_extension) && !exclude_resource_extension

package acceptancetests

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/extensionmanagement"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// getDirectExtensionClient builds an AggregatedClient directly from AZDO env vars.
// Used by CheckDestroy and evidence helpers because ProtoV6ProviderFactories does
// not wire the SDKv2 provider singleton's Meta, so testutils.GetProvider().Meta()
// would be nil when the test uses the mux provider factory.
func getDirectExtensionClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

// captureExtensionEvidence fetches the live extension GET URL and writes forge live evidence.
// Best-effort: never fails the test — errors are silently swallowed.
func captureExtensionEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_ = tryCaptureLiveExtensionEvidence(tfNode, s) //nolint:errcheck
		return nil
	}
}

// tryCaptureLiveExtensionEvidence performs the actual evidence capture and returns
// an error on failure (caller ignores it).
func tryCaptureLiveExtensionEvidence(tfNode string, s *terraform.State) error {
	res, ok := s.RootModule().Resources[tfNode]
	if !ok {
		return fmt.Errorf("resource %s not found in state", tfNode)
	}

	publisherID := res.Primary.Attributes["publisher_id"]
	extensionID := res.Primary.Attributes["extension_id"]

	orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
	// Extension Management GET URL per WI-2 spec.
	getURL := fmt.Sprintf(
		"https://extmgmt.dev.azure.com/%s/_apis/extensionmanagement/installedextensionsbyname/%s/%s?api-version=7.1",
		extractExtensionOrgName(orgURL),
		publisherID,
		extensionID,
	)

	clients, err := getDirectExtensionClient()
	if err != nil {
		return err
	}

	resp, err := clients.ExtensionManagementClient.GetInstalledExtensionByName(clients.Ctx, extensionmanagement.GetInstalledExtensionByNameArgs{
		PublisherName: &publisherID,
		ExtensionName: &extensionID,
	})
	if err != nil {
		return err
	}

	return testutils.CaptureLiveEvidence("acceptance-resource", getURL, resp)
}

// extractExtensionOrgName returns the last path segment from an ADO org URL like
// https://dev.azure.com/myorg → "myorg".
func extractExtensionOrgName(orgURL string) string {
	parts := strings.Split(strings.TrimRight(orgURL, "/"), "/")
	if len(parts) == 0 {
		return orgURL
	}
	return parts[len(parts)-1]
}

func TestAccExtension_basic(t *testing.T) {
	publisherId := "ms-securitydevops"
	extensionId := "microsoft-security-devops-azdevops"
	tfNode := "betterado_extension.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkExtensionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclExtensionBasic(publisherId, extensionId),
				Check: resource.ComposeTestCheckFunc(
					checkExtensionExist(extensionId),
					captureExtensionEvidence(tfNode),
					resource.TestCheckResourceAttrSet(tfNode, "extension_id"),
					resource.TestCheckResourceAttrSet(tfNode, "publisher_id"),
					resource.TestCheckResourceAttrSet(tfNode, "publisher_name"),
					resource.TestCheckResourceAttrSet(tfNode, "extension_name"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", publisherId, extensionId),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccExtension_complete(t *testing.T) {
	publisherId := "ms-securitydevops"
	extensionId := "microsoft-security-devops-azdevops"
	tfNode := "betterado_extension.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkExtensionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclExtensionComplete(publisherId, extensionId),
				Check: resource.ComposeTestCheckFunc(
					checkExtensionExist(extensionId),
					captureExtensionEvidence(tfNode),
					resource.TestCheckResourceAttrSet(tfNode, "extension_id"),
					resource.TestCheckResourceAttrSet(tfNode, "publisher_id"),
					resource.TestCheckResourceAttrSet(tfNode, "publisher_name"),
					resource.TestCheckResourceAttrSet(tfNode, "extension_name"),
					resource.TestCheckResourceAttrSet(tfNode, "scope.#"),
					resource.TestCheckResourceAttrSet(tfNode, "version"),
					resource.TestCheckResourceAttrSet(tfNode, "disabled"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", publisherId, extensionId),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccExtension_update(t *testing.T) {
	publisherId := "ms-securitydevops"
	extensionId := "microsoft-security-devops-azdevops"
	tfNode := "betterado_extension.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkExtensionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclExtensionBasic(publisherId, extensionId),
				Check: resource.ComposeTestCheckFunc(
					checkExtensionExist(extensionId),
					captureExtensionEvidence(tfNode),
					resource.TestCheckResourceAttrSet(tfNode, "extension_id"),
					resource.TestCheckResourceAttrSet(tfNode, "publisher_id"),
					resource.TestCheckResourceAttrSet(tfNode, "publisher_name"),
					resource.TestCheckResourceAttrSet(tfNode, "extension_name"),
					resource.TestCheckResourceAttrSet(tfNode, "scope.#"),
					resource.TestCheckResourceAttrSet(tfNode, "version"),
					resource.TestCheckResourceAttrSet(tfNode, "disabled"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", publisherId, extensionId),
				ImportStateVerify: true,
			},
			{
				Config: hclExtensionUpdate(publisherId, extensionId, true),
				Check: resource.ComposeTestCheckFunc(
					checkExtensionExist(extensionId),
					resource.TestCheckResourceAttrSet(tfNode, "extension_id"),
					resource.TestCheckResourceAttrSet(tfNode, "publisher_id"),
					resource.TestCheckResourceAttrSet(tfNode, "publisher_name"),
					resource.TestCheckResourceAttrSet(tfNode, "extension_name"),
					resource.TestCheckResourceAttr(tfNode, "disabled", "true"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", publisherId, extensionId),
				ImportStateVerify: true,
			},
			{
				Config: hclExtensionUpdate(publisherId, extensionId, false),
				Check: resource.ComposeTestCheckFunc(
					checkExtensionExist(extensionId),
					resource.TestCheckResourceAttrSet(tfNode, "extension_id"),
					resource.TestCheckResourceAttrSet(tfNode, "publisher_id"),
					resource.TestCheckResourceAttrSet(tfNode, "publisher_name"),
					resource.TestCheckResourceAttrSet(tfNode, "extension_name"),
					resource.TestCheckResourceAttrSet(tfNode, "scope.#"),
					resource.TestCheckResourceAttr(tfNode, "disabled", "false"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", publisherId, extensionId),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccExtension_requireImportError(t *testing.T) {
	publisherId := "ms-securitydevops"
	extensionId := "microsoft-security-devops-azdevops"
	tfNode := "betterado_extension.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkExtensionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclExtensionBasic(publisherId, extensionId),
				Check: resource.ComposeTestCheckFunc(
					checkExtensionExist(extensionId),
					captureExtensionEvidence(tfNode),
					resource.TestCheckResourceAttrSet(tfNode, "extension_id"),
					resource.TestCheckResourceAttrSet(tfNode, "publisher_id"),
					resource.TestCheckResourceAttrSet(tfNode, "publisher_name"),
					resource.TestCheckResourceAttrSet(tfNode, "extension_name"),
					resource.TestCheckResourceAttrSet(tfNode, "scope.#"),
					resource.TestCheckResourceAttrSet(tfNode, "version"),
					resource.TestCheckResourceAttrSet(tfNode, "disabled"),
				),
			},
			{
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", publisherId, extensionId),
				ImportStateVerify: true,
			},
			{
				Config:      hclExtensionImportError(publisherId, extensionId),
				ExpectError: requiresExtensionImportError(publisherId, extensionId),
			},
		},
	})
}

func checkExtensionDestroyed(s *terraform.State) error {
	clients, err := getDirectExtensionClient()
	if err != nil {
		return fmt.Errorf("checkExtensionDestroyed: failed to build client: %v", err)
	}
	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_extension" {
			continue
		}
		ids := strings.Split(res.Primary.ID, "/")
		if len(ids) != 2 {
			return fmt.Errorf("unexpected extension resource ID format: %s", res.Primary.ID)
		}

		_, err := clients.ExtensionManagementClient.GetInstalledExtensionByName(clients.Ctx, extensionmanagement.GetInstalledExtensionByNameArgs{
			PublisherName: &ids[0],
			ExtensionName: &ids[1],
		})

		if err == nil {
			return fmt.Errorf("Extension with Publisher ID=%s , Extension ID: %s should not exist", ids[0], ids[1])
		}
	}
	return nil
}

func checkExtensionExist(expectedExtensionId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources["betterado_extension.test"]
		if !ok {
			return fmt.Errorf("Did not find `betterado_extension` in the TF state")
		}

		clients, err := getDirectExtensionClient()
		if err != nil {
			return fmt.Errorf("checkExtensionExist: failed to build client: %v", err)
		}
		ids := strings.Split(res.Primary.ID, "/")
		if len(ids) != 2 {
			return fmt.Errorf("unexpected extension resource ID format: %s", res.Primary.ID)
		}

		extension, err := clients.ExtensionManagementClient.GetInstalledExtensionByName(clients.Ctx, extensionmanagement.GetInstalledExtensionByNameArgs{
			PublisherName: &ids[0],
			ExtensionName: &ids[1],
		})
		if err != nil {
			return fmt.Errorf("Extension with Publisher ID=%s , Extension ID: %s cannot be found!. Error=%v", ids[0], ids[1], err)
		}

		if *extension.ExtensionId != expectedExtensionId {
			return fmt.Errorf("Extension with Publisher ID=%s has Extension ID=%s, but expected Extension ID=%s", *extension.PublisherId, *extension.ExtensionId, expectedExtensionId)
		}
		return nil
	}
}

func requiresExtensionImportError(publisherId, extensionId string) *regexp.Regexp {
	message := "Installing extension for Publisher: %s, Name: %s. Error: TF1590010: Extension %s.%s is already installed in this organization"
	return regexp.MustCompile(fmt.Sprintf(message, publisherId, extensionId, publisherId, extensionId))
}

func hclExtensionBasic(publisherId, extensionId string) string {
	return fmt.Sprintf(`
resource "betterado_extension" "test" {
  publisher_id = "%s"
  extension_id = "%s"
}`, publisherId, extensionId)
}

func hclExtensionComplete(publisherId, extensionId string) string {
	return fmt.Sprintf(`
resource "betterado_extension" "test" {
  publisher_id = "%s"
  extension_id = "%s"
  disabled     = false
}`, publisherId, extensionId)
}

func hclExtensionUpdate(publisherId, extensionId string, disabled bool) string {
	return fmt.Sprintf(`
resource "betterado_extension" "test" {
  publisher_id = "%s"
  extension_id = "%s"
  disabled     = %t
}`, publisherId, extensionId, disabled)
}

func hclExtensionImportError(publisherId, extensionId string) string {
	return fmt.Sprintf(`
%s

resource "betterado_extension" "import" {
  publisher_id = betterado_extension.test.publisher_id
  extension_id = betterado_extension.test.extension_id
}`, hclExtensionBasic(publisherId, extensionId))
}
