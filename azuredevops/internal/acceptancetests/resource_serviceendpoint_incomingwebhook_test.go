//go:build (all || resource_serviceendpoint_incomingwebhook) && !exclude_resource_serviceendpoint_incomingwebhook

package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccServiceEndpointIncomingWebhook_Basic(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()

	resourceType := "betterado_serviceendpoint_incomingwebhook"
	tfSvcEpNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointIncomingWebhookResource(projectName, serviceEndpointName),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointName),
					resource.TestCheckResourceAttr(tfSvcEpNode, "webhook_name", "test_webhook_name"),
				),
			},
		},
	})
}

func TestAccServiceEndpointIncomingWebhook_Complete(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()
	description := testutils.GenerateResourceName()
	webhookName := "test_webhook_name"
	secret := "hsdhj23r8/3aefh1!"
	httpHeader := "X-Header-Test"

	resourceType := "betterado_serviceendpoint_incomingwebhook"
	tfSvcEpNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointIncomingWebhookResourceComplete(projectName, serviceEndpointName, webhookName, secret, httpHeader, description),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointName),
					resource.TestCheckResourceAttr(tfSvcEpNode, "webhook_name", webhookName),
					resource.TestCheckResourceAttr(tfSvcEpNode, "secret", secret),
					resource.TestCheckResourceAttr(tfSvcEpNode, "http_header", httpHeader),
					resource.TestCheckResourceAttr(tfSvcEpNode, "description", description),
				),
			},
		},
	})
}

func TestAccServiceEndpointIncomingWebhook_update(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointNameFirst := testutils.GenerateResourceName()
	webhookName := "test_webhook_name"

	description := testutils.GenerateResourceName()
	serviceEndpointNameSecond := testutils.GenerateResourceName()

	resourceType := "betterado_serviceendpoint_incomingwebhook"
	tfSvcEpNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointIncomingWebhookResource(projectName, serviceEndpointNameFirst),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointNameFirst), resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameFirst),
				),
			},
			{
				Config: hclSvcEndpointIncomingWebhookResourceUpdate(projectName, webhookName, serviceEndpointNameSecond, description),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointNameSecond),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameSecond),
					resource.TestCheckResourceAttr(tfSvcEpNode, "description", description),
				),
			},
		},
	})
}

func TestAccServiceEndpointIncomingWebhook_requiresImportErrorStep(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()

	resourceType := "betterado_serviceendpoint_incomingwebhook"
	tfSvcEpNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointIncomingWebhookResource(projectName, serviceEndpointName),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointName),
				),
			},
			{
				Config:      hclSvcEndpointIncomingWebhookResourceRequiresImport(projectName, serviceEndpointName),
				ExpectError: testutils.RequiresImportError(serviceEndpointName),
			},
		},
	})
}

func hclSvcEndpointIncomingWebhookResource(projectName string, serviceEndpointName string) string {
	return hclSvcEndpointIncomingWebhookResourceUpdate(projectName, "test_webhook_name", serviceEndpointName, "description")
}

func hclSvcEndpointIncomingWebhookResourceUpdate(projectName string, webhookName string, serviceEndpointName string, description string) string {
	serviceEndpointResource := fmt.Sprintf(`
resource "betterado_serviceendpoint_incomingwebhook" "test" {
  project_id            = betterado_project.project.id
  webhook_name          = "%s"
  secret                = "secret1!"
  http_header           = "X-Header"
  service_endpoint_name = "%s"
  description           = "%s"
}`, webhookName, serviceEndpointName, description)

	projectResource := testutils.HclProjectResource(projectName)
	return fmt.Sprintf("%s\n%s", projectResource, serviceEndpointResource)
}

func hclSvcEndpointIncomingWebhookResourceComplete(projectName string, serviceEndpointName string, webhookName string, secret string, httpHeader string, description string) string {
	serviceEndpointResource := fmt.Sprintf(`
resource "betterado_serviceendpoint_incomingwebhook" "test" {
  project_id            = betterado_project.project.id
  service_endpoint_name = "%s"
  webhook_name          = "%s"
  secret                = "%s"
  http_header           = "%s"
  description           = "%s"
}`, serviceEndpointName, webhookName, secret, httpHeader, description)

	projectResource := testutils.HclProjectResource(projectName)
	return fmt.Sprintf("%s\n%s", projectResource, serviceEndpointResource)
}

func hclSvcEndpointIncomingWebhookResourceRequiresImport(projectName string, serviceEndpointName string) string {
	template := hclSvcEndpointIncomingWebhookResource(projectName, serviceEndpointName)
	return fmt.Sprintf(`
	%s
resource "betterado_serviceendpoint_incomingwebhook" "import" {
  project_id            = betterado_serviceendpoint_incomingwebhook.test.project_id
  webhook_name          = "test_webhook_name"
  secret                = "hsdhj23r8/3aefh1!"
  http_header           = "X-Header-Test"
  service_endpoint_name = betterado_serviceendpoint_incomingwebhook.test.service_endpoint_name
  description           = betterado_serviceendpoint_incomingwebhook.test.description
}
	`, template)
}
