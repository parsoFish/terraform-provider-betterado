//go:build (all || resource_serviceendpoint_jenkins) && !exclude_resource_serviceendpoint_jenkins

package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccServiceEndpointJenkins_basic_usernamepassword(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()

	resourceType := "betterado_serviceendpoint_jenkins"
	tfSvcEpNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointJenkinsResourceBasicUsernamePassword(projectName, serviceEndpointName, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointName),
				),
			},
		},
	})
}

func TestAccServiceEndpointJenkins_complete_usernamepassword(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()
	description := t.Name()

	resourceType := "betterado_serviceendpoint_jenkins"
	tfSvcEpNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointJenkinsResourceCompleteUsernamePassword(projectName, serviceEndpointName, description),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "url", "https://url.com/1"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointName),
					resource.TestCheckResourceAttr(tfSvcEpNode, "description", description),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "username"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "password"),
				),
			},
		},
	})
}

func TestAccServiceEndpointJenkins_update_usernamepassword(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointNameFirst := testutils.GenerateResourceName()

	description := t.Name()
	serviceEndpointNameSecond := testutils.GenerateResourceName()

	resourceType := "betterado_serviceendpoint_jenkins"
	tfSvcEpNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointJenkinsResourceBasicUsernamePassword(projectName, serviceEndpointNameFirst, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointNameFirst), resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameFirst),
				),
			},
			{
				Config: hclSvcEndpointJenkinsResourceUpdateUsernamePassword(projectName, serviceEndpointNameSecond, description),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointNameSecond),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "url", "https://url.com/2"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameSecond),
					resource.TestCheckResourceAttr(tfSvcEpNode, "description", description),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "username"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "password"),
				),
			},
		},
	})
}

func TestAccServiceEndpointJenkins_RequiresImportErrorStepUsernamePassword(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()
	resourceType := "betterado_serviceendpoint_jenkins"
	tfSvcEpNode := resourceType + ".test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclSvcEndpointJenkinsResourceBasicUsernamePassword(projectName, serviceEndpointName, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
				),
			},
			{
				Config:      hclSvcEndpointJenkinsResourceRequiresImportUsernamePassword(projectName, serviceEndpointName, t.Name()),
				ExpectError: testutils.RequiresImportError(serviceEndpointName),
			},
		},
	})
}

func hclSvcEndpointJenkinsResourceBasicUsernamePassword(projectName string, serviceEndpointName string, description string) string {
	serviceEndpointResource := fmt.Sprintf(`
resource "betterado_serviceendpoint_jenkins" "test" {
  project_id             = betterado_project.project.id
  service_endpoint_name  = "%s"
  username               = "u"
  password               = "redacted"
  url                    = "http://url.com/1"
  description            = "%s"
  accept_untrusted_certs = false
}`, serviceEndpointName, description)

	projectResource := testutils.HclProjectResource(projectName)
	return fmt.Sprintf("%s\n%s", projectResource, serviceEndpointResource)
}

func hclSvcEndpointJenkinsResourceCompleteUsernamePassword(projectName string, serviceEndpointName string, description string) string {
	serviceEndpointResource := fmt.Sprintf(`
resource "betterado_serviceendpoint_jenkins" "test" {
  project_id             = betterado_project.project.id
  service_endpoint_name  = "%s"
  description            = "%s"
  username               = "u"
  password               = "redacted"
  url                    = "https://url.com/1"
  accept_untrusted_certs = false
}`, serviceEndpointName, description)

	projectResource := testutils.HclProjectResource(projectName)
	return fmt.Sprintf("%s\n%s", projectResource, serviceEndpointResource)
}

func hclSvcEndpointJenkinsResourceUpdateUsernamePassword(projectName string, serviceEndpointName string, description string) string {
	serviceEndpointResource := fmt.Sprintf(`
resource "betterado_serviceendpoint_jenkins" "test" {
  project_id             = betterado_project.project.id
  service_endpoint_name  = "%s"
  description            = "%s"
  username               = "u2"
  password               = "redacted2"
  url                    = "https://url.com/2"
  accept_untrusted_certs = false
}`, serviceEndpointName, description)

	projectResource := testutils.HclProjectResource(projectName)
	return fmt.Sprintf("%s\n%s", projectResource, serviceEndpointResource)
}

func hclSvcEndpointJenkinsResourceRequiresImportUsernamePassword(projectName string, serviceEndpointName string, description string) string {
	template := hclSvcEndpointJenkinsResourceBasicUsernamePassword(projectName, serviceEndpointName, description)
	return fmt.Sprintf(`
%s
resource "betterado_serviceendpoint_jenkins" "import" {
  project_id             = betterado_serviceendpoint_jenkins.test.project_id
  service_endpoint_name  = betterado_serviceendpoint_jenkins.test.service_endpoint_name
  description            = betterado_serviceendpoint_jenkins.test.description
  url                    = betterado_serviceendpoint_jenkins.test.url
  accept_untrusted_certs = false
  username               = "u"
  password               = "redacted"
}
`, template)
}
