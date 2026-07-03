package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccServiceEndpointCheckMarxOne_apiKey(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()

	tfSvcEpNode := "betterado_serviceendpoint_checkmarx_one.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed("betterado_serviceendpoint_checkmarx_one"),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointCheckMarxOneServiceResourceApiKey(projectName, serviceEndpointName),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
					resource.TestCheckResourceAttr(tfSvcEpNode, "server_url", "https://server.com"),
				),
			},
		},
	})
}

func TestAccServiceEndpointCheckMarxOne_apiKeyUpdate(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()

	tfSvcEpNode := "betterado_serviceendpoint_checkmarx_one.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed("betterado_serviceendpoint_checkmarx_one"),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointCheckMarxOneServiceResourceApiKeyUpdate(projectName, serviceEndpointName),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "api_key"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "server_url", "https://server.com/update"),
				),
			},
		},
	})
}

func TestAccServiceEndpointCheckMarxOne_clientIdSecret(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()

	tfSvcEpNode := "betterado_serviceendpoint_checkmarx_one.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed("betterado_serviceendpoint_checkmarx_one"),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointCheckMarxOneServiceResourceClientIdSecret(projectName, serviceEndpointName),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
					resource.TestCheckResourceAttr(tfSvcEpNode, "server_url", "https://server.com"),
				),
			},
		},
	})
}

func TestAccServiceEndpointCheckMarxOne_clientIdSecretUpdate(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()

	tfSvcEpNode := "betterado_serviceendpoint_checkmarx_one.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed("betterado_serviceendpoint_checkmarx_one"),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointCheckMarxOneServiceResourceClientIdSecret(projectName, serviceEndpointName),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
					resource.TestCheckResourceAttr(tfSvcEpNode, "server_url", "https://server.com"),
				),
			},
			{
				Config: hclSvcEndpointCheckMarxOneServiceResourceClientIdSecretUpdate(projectName, serviceEndpointName),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
					resource.TestCheckResourceAttr(tfSvcEpNode, "server_url", "https://server.com/update"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "client_id", "clientidupdate"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "authorization_url", "https://authurl.com/update"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "description", "descriptionupdate"),
				),
			},
		},
	})
}

func TestAccServiceEndpointCheckMarxOne_requiresImportErrorStep(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()
	tfSvcEpNode := "betterado_serviceendpoint_checkmarx_one.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed("betterado_serviceendpoint_checkmarx_one"),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointCheckMarxOneServiceResourceApiKey(projectName, serviceEndpointName),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
				),
			},
			{
				Config:      hclSvcEndpointCheckMarxOneServiceResourceRequiresImport(projectName, serviceEndpointName),
				ExpectError: testutils.RequiresImportError(serviceEndpointName),
			},
		},
	})
}

func hclSvcEndpointCheckMarxOneServiceResourceApiKey(projectName, serviceEndpointName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_serviceendpoint_checkmarx_one" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = "%s"
  server_url            = "https://server.com"
  api_key               = "apikey"
}`, projectName, serviceEndpointName)
}

func hclSvcEndpointCheckMarxOneServiceResourceApiKeyUpdate(projectName, serviceEndpointName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_serviceendpoint_checkmarx_one" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = "%s"
  server_url            = "https://server.com/update"
  api_key               = "apikeyupdate"
}`, projectName, serviceEndpointName)
}

func hclSvcEndpointCheckMarxOneServiceResourceClientIdSecret(projectName, serviceEndpointName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_serviceendpoint_checkmarx_one" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = "%s"
  server_url            = "https://server.com"
  client_id             = "clientid"
  client_secret         = "secret"
  authorization_url     = "https://authurl.com"
}`, projectName, serviceEndpointName)
}

func hclSvcEndpointCheckMarxOneServiceResourceClientIdSecretUpdate(projectName, serviceEndpointName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_serviceendpoint_checkmarx_one" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = "%s"
  server_url            = "https://server.com/update"
  client_id             = "clientidupdate"
  client_secret         = "secretupdate"
  authorization_url     = "https://authurl.com/update"
  description           = "descriptionupdate"
}`, projectName, serviceEndpointName)
}

func hclSvcEndpointCheckMarxOneServiceResourceRequiresImport(projectName, serviceEndpointName string) string {
	template := hclSvcEndpointCheckMarxOneServiceResourceApiKey(projectName, serviceEndpointName)
	return fmt.Sprintf(`
%s

resource "betterado_serviceendpoint_checkmarx_one" "import" {
  project_id            = betterado_serviceendpoint_checkmarx_one.test.project_id
  service_endpoint_name = betterado_serviceendpoint_checkmarx_one.test.service_endpoint_name
  description           = betterado_serviceendpoint_checkmarx_one.test.description
  server_url            = betterado_serviceendpoint_checkmarx_one.test.server_url
  api_key               = betterado_serviceendpoint_checkmarx_one.test.api_key
}
`, template)
}
