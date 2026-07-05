//go:build (all || resource_serviceendpoint_github) && !exclude_resource_serviceendpoint_github

package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/serviceendpoint"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccServiceEndpointGitHub_basic verifies the framework-migrated
// betterado_serviceendpoint_github resource end-to-end:
//
//  1. apply — creates a GitHub (PAT) service endpoint in the pre-existing shared project
//  2. read-back — asserts name + project_id + captures live evidence
//  3. idempotency — re-plan produces no diff (ExpectNonEmptyPlan: false)
//  4. destroy — endpoint is removed cleanly
//
// Uses SharedFixtureProjectName (betterado-standing-demo) so no new ADO project
// is created — the org is at the 1000-project cap; project creates would fail.
func TestAccServiceEndpointGitHub_basic(t *testing.T) {
	serviceEndpointName := testutils.GenerateResourceName()

	resourceType := "betterado_serviceendpoint_github"
	tfSvcEpNode := resourceType + ".test"

	config := hclSvcEndpointGitHubBasic(serviceEndpointName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, &[]string{"AZDO_GITHUB_SERVICE_CONNECTION_PAT"}) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkServiceEndpointGitHubDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointName),
					resource.TestCheckResourceAttr(tfSvcEpNode, "description", "Managed by Terraform"),
					captureServiceEndpointGitHubEvidence(tfSvcEpNode),
				),
			},
			{
				// idempotency: re-plan must produce no diff
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccServiceEndpointGitHub_PersonalTokenBasic(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()

	resourceType := "betterado_serviceendpoint_github"
	tfSvcEpNode := resourceType + ".serviceendpoint"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, &[]string{"AZDO_GITHUB_SERVICE_CONNECTION_PAT"}) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: personTokenConfigBasic(projectName, serviceEndpointName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointName),
					resource.TestCheckResourceAttr(tfSvcEpNode, "description", "Managed by Terraform"),
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
				),
			},
		},
	})
}

func TestAccServiceEndpointGitHub_PersonalTokenUpdate(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointNameFirst := testutils.GenerateResourceName()
	serviceEndpointNameSecond := testutils.GenerateResourceName()
	description := "Manage by Terraform Update"

	resourceType := "betterado_serviceendpoint_github"
	tfSvcEpNode := resourceType + ".serviceendpoint"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, &[]string{"AZDO_GITHUB_SERVICE_CONNECTION_PAT"}) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: personTokenConfigBasic(projectName, serviceEndpointNameFirst),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameFirst),
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointNameFirst),
				),
			}, {
				Config: personTokenConfigUpdate(projectName, serviceEndpointNameSecond, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameSecond),
					resource.TestCheckResourceAttr(tfSvcEpNode, "description", description),
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointNameSecond),
				),
			},
		},
	})
}

func TestAccServiceEndpointGitHub_OauthBasic(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()

	resourceType := "betterado_serviceendpoint_github"
	tfSvcEpNode := resourceType + ".serviceendpoint"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, &[]string{"AZDO_GITHUB_SERVICE_CONNECTION_PAT"}) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: oauthConfigBasic(projectName, serviceEndpointName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointName),
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointName),
				),
			},
		},
	})
}

func TestAccServiceEndpointGitHub_OauthUpdate(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointNameFirst := testutils.GenerateResourceName()
	serviceEndpointNameSecond := testutils.GenerateResourceName()
	description := "Manage by Terraform Update"

	resourceType := "betterado_serviceendpoint_github"
	tfSvcEpNode := resourceType + ".serviceendpoint"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, &[]string{"AZDO_GITHUB_SERVICE_CONNECTION_PAT"}) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: oauthConfigBasic(projectName, serviceEndpointNameFirst),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameFirst),
					resource.TestCheckResourceAttr(tfSvcEpNode, "description", "Managed by Terraform"),
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointNameFirst),
				),
			}, {
				Config: oauthConfigUpdate(projectName, serviceEndpointNameSecond, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameSecond),
					resource.TestCheckResourceAttr(tfSvcEpNode, "description", description),
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointNameSecond),
				),
			},
		},
	})
}

func TestAccServiceEndpointGitHub_CreateAndUpdate(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointNameFirst := testutils.GenerateResourceName()
	serviceEndpointNameSecond := testutils.GenerateResourceName()

	resourceType := "betterado_serviceendpoint_github"
	tfSvcEpNode := resourceType + ".serviceendpoint"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, &[]string{"AZDO_GITHUB_SERVICE_CONNECTION_PAT"}) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckServiceEndpointDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: testutils.HclServiceEndpointGitHubResource(projectName, serviceEndpointNameFirst),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameFirst),
					resource.TestCheckResourceAttr(tfSvcEpNode, "description", "Managed by Terraform"),
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointNameFirst),
				),
			}, {
				Config: testutils.HclServiceEndpointGitHubResource(projectName, serviceEndpointNameSecond),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameSecond),
					resource.TestCheckResourceAttr(tfSvcEpNode, "description", "Managed by Terraform"),
					testutils.CheckServiceEndpointExistsWithName(tfSvcEpNode, serviceEndpointNameSecond),
				),
			},
		},
	})
}

// ── config helpers ────────────────────────────────────────────────────────────

// hclSvcEndpointGitHubBasic returns HCL for a minimal GitHub PAT service endpoint
// using the pre-existing shared project (no new ADO projects).
func hclSvcEndpointGitHubBasic(serviceEndpointName string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_serviceendpoint_github" "test" {
  project_id            = data.betterado_project.test.id
  service_endpoint_name = %[1]q
  personal_access_token = "test_pat_token"
}
`, serviceEndpointName, SharedFixtureProjectName)
}

func personTokenConfigBasic(projectName string, serviceEndpointName string) string {
	projectResource := testutils.HclProjectResource(projectName)

	serviceEndpointResource := fmt.Sprintf(`
resource "betterado_serviceendpoint_github" "serviceendpoint" {
  project_id            = betterado_project.project.id
  service_endpoint_name = "%[1]s"
  personal_access_token = "test_token_basic"
}`, serviceEndpointName)

	return fmt.Sprintf("%s\n%s", projectResource, serviceEndpointResource)
}

func personTokenConfigUpdate(projectName string, serviceEndpointName string, description string) string {
	projectResource := testutils.HclProjectResource(projectName)

	serviceEndpointResource := fmt.Sprintf(`
resource "betterado_serviceendpoint_github" "serviceendpoint" {
  project_id            = betterado_project.project.id
  service_endpoint_name = "%[1]s"
  personal_access_token = "test_token_update"
  description           = "%[2]s"
}`, serviceEndpointName, description)

	return fmt.Sprintf("%s\n%s", projectResource, serviceEndpointResource)
}

func oauthConfigBasic(projectName string, serviceEndpointName string) string {
	projectResource := testutils.HclProjectResource(projectName)
	serviceEndpointResource := fmt.Sprintf(`
resource "betterado_serviceendpoint_github" "serviceendpoint" {
  project_id             = betterado_project.project.id
  service_endpoint_name  = "%[1]s"
  oauth_configuration_id = "xxxx-xxxx-xxxx-xxxx-xxxx"
}`, serviceEndpointName)

	return fmt.Sprintf("%s\n%s", projectResource, serviceEndpointResource)
}

func oauthConfigUpdate(projectName string, serviceEndpointName string, description string) string {
	projectResource := testutils.HclProjectResource(projectName)
	serviceEndpointResource := fmt.Sprintf(`
resource "betterado_serviceendpoint_github" "serviceendpoint" {
  project_id             = betterado_project.project.id
  service_endpoint_name  = "%[1]s"
  oauth_configuration_id = "xxx-xxxx-xxxx-xxxx-xxxx"
  description            = "%[2]s"
}`, serviceEndpointName, description)

	return fmt.Sprintf("%s\n%s", projectResource, serviceEndpointResource)
}

// ── CheckDestroy / evidence ───────────────────────────────────────────────���───

// checkServiceEndpointGitHubDestroyed verifies all GitHub service endpoints in
// the state have been deleted. Uses getDirectClient (not testutils.GetProvider().Meta())
// because ProtoV6ProviderFactories does not wire the SDKv2 provider singleton's Meta.
func checkServiceEndpointGitHubDestroyed(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("checkServiceEndpointGitHubDestroyed: build client: %w", err)
		}
		for _, res := range s.RootModule().Resources {
			if res.Type != resourceType {
				continue
			}
			id, err := uuid.Parse(res.Primary.ID)
			if err != nil {
				return fmt.Errorf("invalid service endpoint ID %q: %w", res.Primary.ID, err)
			}
			projectID := res.Primary.Attributes["project_id"]
			ep, err := clients.ServiceEndpointClient.GetServiceEndpointDetails(clients.Ctx,
				serviceendpoint.GetServiceEndpointDetailsArgs{
					EndpointId: &id,
					Project:    &projectID,
				})
			if err == nil && ep != nil && ep.Id != nil {
				return fmt.Errorf("service endpoint %s still exists after destroy", id)
			}
		}
		return nil
	}
}

// captureServiceEndpointGitHubEvidence performs a real live API GET of the
// created GitHub service endpoint and persists the response as forge demo
// live-evidence (before destroy). Label "acceptance-resource-github" matches
// the forge unifier checkpoint. Best-effort: a capture failure never fails the test.
func captureServiceEndpointGitHubEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		id, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return nil
		}
		projectID := res.Primary.Attributes["project_id"]
		clients, err := getDirectClient()
		if err != nil {
			return nil
		}
		ep, err := clients.ServiceEndpointClient.GetServiceEndpointDetails(clients.Ctx,
			serviceendpoint.GetServiceEndpointDetailsArgs{
				EndpointId: &id,
				Project:    &projectID,
			})
		if err != nil || ep == nil {
			return nil
		}
		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		endpointURL := fmt.Sprintf("%s/_apis/serviceendpoint/endpoints/%s?api-version=7.1", orgURL, id)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-github", endpointURL, ep)
		return nil
	}
}
