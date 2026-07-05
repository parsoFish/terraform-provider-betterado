package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccCheckRestAPI_basic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	serviceConnectionName := testutils.GenerateResourceName()
	displayName := testutils.GenerateResourceName()

	tfCheckNode := "betterado_check_rest_api.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed("betterado_check_rest_api"),
		Steps: []resource.TestStep{
			{
				Config: hclCheckRestAPIResourceBasic(projectID, serviceConnectionName, displayName),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckPipelineCheckExistsWithName(tfCheckNode, displayName),
					resource.TestCheckResourceAttr(tfCheckNode, "connected_service_name_selector", "connectedServiceName"),
					resource.TestCheckResourceAttr(tfCheckNode, "connected_service_name", "se_"+serviceConnectionName),
					resource.TestCheckResourceAttr(tfCheckNode, "method", "GET"),
				),
			},
		},
	})
}

func TestAccCheckRestAPI_complete(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	variableGroupName := testutils.GenerateResourceName()
	displayName := testutils.GenerateResourceName()
	serviceConnectionName := testutils.GenerateResourceName()

	tfCheckNode := "betterado_check_rest_api.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed("betterado_check_rest_api"),
		Steps: []resource.TestStep{
			{
				Config: hclCheckRestAPIResourceComplete(projectID, serviceConnectionName, displayName, variableGroupName),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckPipelineCheckExistsWithName(tfCheckNode, displayName),
					resource.TestCheckResourceAttr(tfCheckNode, "connected_service_name_selector", "connectedServiceName"),
					resource.TestCheckResourceAttr(tfCheckNode, "connected_service_name", "se_"+serviceConnectionName),
					resource.TestCheckResourceAttr(tfCheckNode, "method", "POST"),
					resource.TestCheckResourceAttr(tfCheckNode, "headers", "{\"contentType\":\"application/json\"}"),
					resource.TestCheckResourceAttr(tfCheckNode, "body", "{\"params\":\"value\"}"),
					resource.TestCheckResourceAttr(tfCheckNode, "completion_event", "ApiResponse"),
					resource.TestCheckResourceAttr(tfCheckNode, "success_criteria", "eq(root['status'], '200')"),
					resource.TestCheckResourceAttr(tfCheckNode, "url_suffix", "user/1"),
					resource.TestCheckResourceAttr(tfCheckNode, "retry_interval", "4000"),
					resource.TestCheckResourceAttr(tfCheckNode, "variable_group_name", variableGroupName),
					resource.TestCheckResourceAttr(tfCheckNode, "timeout", "40000"),
				),
			},
		},
	})
}

func TestAccCheckRestAPI_update(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	variableGroupName := testutils.GenerateResourceName()
	displayName := testutils.GenerateResourceName()
	serviceConnectionName := testutils.GenerateResourceName()

	tfCheckNode := "betterado_check_rest_api.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed("betterado_check_rest_api"),
		Steps: []resource.TestStep{
			{
				Config: hclCheckRestAPIResourceBasic(projectID, serviceConnectionName, displayName),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckPipelineCheckExistsWithName(tfCheckNode, displayName),
					resource.TestCheckResourceAttr(tfCheckNode, "connected_service_name_selector", "connectedServiceName"),
					resource.TestCheckResourceAttr(tfCheckNode, "connected_service_name", "se_"+serviceConnectionName),
					resource.TestCheckResourceAttr(tfCheckNode, "method", "GET"),
				),
			},
			{
				Config: hclCheckRestAPIResourceComplete(projectID, serviceConnectionName, displayName, variableGroupName),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckPipelineCheckExistsWithName(tfCheckNode, displayName),
					resource.TestCheckResourceAttr(tfCheckNode, "connected_service_name_selector", "connectedServiceName"),
					resource.TestCheckResourceAttr(tfCheckNode, "connected_service_name", "se_"+serviceConnectionName),
					resource.TestCheckResourceAttr(tfCheckNode, "method", "POST"),
					resource.TestCheckResourceAttr(tfCheckNode, "headers", "{\"contentType\":\"application/json\"}"),
					resource.TestCheckResourceAttr(tfCheckNode, "body", "{\"params\":\"value\"}"),
					resource.TestCheckResourceAttr(tfCheckNode, "completion_event", "ApiResponse"),
					resource.TestCheckResourceAttr(tfCheckNode, "success_criteria", "eq(root['status'], '200')"),
					resource.TestCheckResourceAttr(tfCheckNode, "url_suffix", "user/1"),
					resource.TestCheckResourceAttr(tfCheckNode, "retry_interval", "4000"),
					resource.TestCheckResourceAttr(tfCheckNode, "variable_group_name", variableGroupName),
					resource.TestCheckResourceAttr(tfCheckNode, "timeout", "40000"),
				),
			},
		},
	})
}

func hclCheckRestAPIResourceBasic(projectID, serviceConnectionName, displayName string) string {
	return fmt.Sprintf(`
resource "betterado_serviceendpoint_generic" "test" {
  project_id            = %q
  service_endpoint_name = "%s"
  server_url            = "https://dev.azure.com/"
  username              = "username"
  password              = "dummy"
}

resource "betterado_serviceendpoint_generic" "test2" {
  project_id            = %q
  service_endpoint_name = "se_%s"
  server_url            = "https://dev.azure.com/"
  username              = "username"
  password              = "dummy"
}

resource "betterado_check_rest_api" "test" {
  project_id                      = %q
  target_resource_id              = betterado_serviceendpoint_generic.test.id
  target_resource_type            = "endpoint"
  display_name                    = "%s"
  connected_service_name_selector = "connectedServiceName"
  connected_service_name          = betterado_serviceendpoint_generic.test2.service_endpoint_name
  method                          = "GET"
}`, projectID, serviceConnectionName, projectID, serviceConnectionName, projectID, displayName)
}

func hclCheckRestAPIResourceComplete(projectID, serviceConnectionName, displayName, variableGroupName string) string {
	return fmt.Sprintf(`
resource "betterado_serviceendpoint_generic" "test" {
  project_id            = %q
  service_endpoint_name = "%s"
  server_url            = "https://dev.azure.com/"
  username              = "username"
  password              = "dummy"
}

resource "betterado_serviceendpoint_generic" "test2" {
  project_id            = %q
  service_endpoint_name = "se_%s"
  server_url            = "https://dev.azure.com/"
  username              = "username"
  password              = "dummy"
}

resource "betterado_variable_group" "test" {
  project_id   = %q
  name         = "%s"
  allow_access = true
  variable = [{
    name  = "FOO"
    value = "BAR"
  }]
}

resource "betterado_check_rest_api" "test" {
  project_id           = %q
  target_resource_id   = betterado_serviceendpoint_generic.test.id
  target_resource_type = "endpoint"

  display_name                    = "%s"
  connected_service_name_selector = "connectedServiceName"
  connected_service_name          = betterado_serviceendpoint_generic.test2.service_endpoint_name
  method                          = "POST"
  headers                         = "{\"contentType\":\"application/json\"}"
  body                            = "{\"params\":\"value\"}"
  completion_event                = "ApiResponse"
  success_criteria                = "eq(root['status'], '200')"
  url_suffix                      = "user/1"
  retry_interval                  = 4000
  variable_group_name             = betterado_variable_group.test.name
  timeout                         = "40000"
}`, projectID, serviceConnectionName, projectID, serviceConnectionName, projectID, variableGroupName, projectID, displayName)
}
