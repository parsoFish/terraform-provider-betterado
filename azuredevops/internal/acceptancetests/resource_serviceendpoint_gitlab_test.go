package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccServiceEndpointGitLab_basic(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()

	resourceType := "betterado_serviceendpoint_gitlab"
	tfSvcEpNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointGitLabResourceBasic(projectName, serviceEndpointName),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointName),
					resource.TestCheckResourceAttr(tfSvcEpNode, "url", "https://gitlab.com"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "username", "username"),
				),
			},
		},
	})
}

func TestAccServiceEndpointGitLab_complete(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()
	description := testutils.GenerateResourceName()

	resourceType := "betterado_serviceendpoint_gitlab"
	tfSvcEpNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointGitLabResourceComplete(projectName, serviceEndpointName, description),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointName),
					resource.TestCheckResourceAttr(tfSvcEpNode, "url", "https://gitlab.com"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "username", "username"),
				),
			},
		},
	})
}

func TestAccServiceEndpointGitLab_update(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointNameFirst := testutils.GenerateResourceName()

	description := testutils.GenerateResourceName()
	serviceEndpointNameSecond := testutils.GenerateResourceName()

	resourceType := "betterado_serviceendpoint_gitlab"
	tfSvcEpNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointGitLabResourceBasic(projectName, serviceEndpointNameFirst),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointNameFirst), resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameFirst),
					resource.TestCheckResourceAttr(tfSvcEpNode, "url", "https://gitlab.com"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "username", "username"),
				),
			},
			{
				Config: hclSvcEndpointGitLabResourceUpdate(projectName, serviceEndpointNameSecond, description),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointNameSecond),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameSecond),
					resource.TestCheckResourceAttr(tfSvcEpNode, "url", "https://gitlab.com/update"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "username", "usernameupdate"),
				),
			},
		},
	})
}

func TestAccServiceEndpointGitLab_requiresImportErrorStep(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()
	resourceType := "betterado_serviceendpoint_gitlab"
	tfSvcEpNode := resourceType + ".test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointGitLabResourceBasic(projectName, serviceEndpointName),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
				),
			},
			{
				Config:      hclSvcEndpointGitLabResourceRequiresImport(projectName, serviceEndpointName),
				ExpectError: testutils.RequiresImportError(serviceEndpointName),
			},
		},
	})
}

func hclSvcEndpointGitLabResourceBasic(projectName string, serviceEndpointName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_serviceendpoint_gitlab" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = "%s"
  url                   = "https://gitlab.com"
  username              = "username"
  api_token             = "token"
}`, projectName, serviceEndpointName)
}

func hclSvcEndpointGitLabResourceComplete(projectName string, serviceEndpointName string, description string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_serviceendpoint_gitlab" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = "%s"
  description           = "%s"
  url                   = "https://gitlab.com"
  username              = "username"
  api_token             = "token"
}`, projectName, serviceEndpointName, description)
}

func hclSvcEndpointGitLabResourceUpdate(projectName string, serviceEndpointName string, description string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = "%s"
}

resource "betterado_serviceendpoint_gitlab" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = "%s"
  description           = "%s"
  url                   = "https://gitlab.com/update"
  username              = "usernameupdate"
  api_token             = "tokenupdate"
}`, projectName, serviceEndpointName, description)
}

func hclSvcEndpointGitLabResourceRequiresImport(projectName string, serviceEndpointName string) string {
	template := hclSvcEndpointGitLabResourceBasic(projectName, serviceEndpointName)
	return fmt.Sprintf(`
%s

resource "betterado_serviceendpoint_gitlab" "import" {
  project_id            = betterado_serviceendpoint_gitlab.test.project_id
  service_endpoint_name = betterado_serviceendpoint_gitlab.test.service_endpoint_name
  description           = betterado_serviceendpoint_gitlab.test.description
  url                   = betterado_serviceendpoint_gitlab.test.url
  username              = betterado_serviceendpoint_gitlab.test.username
  api_token             = betterado_serviceendpoint_gitlab.test.api_token
}
`, template)
}
