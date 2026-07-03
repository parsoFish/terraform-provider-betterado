package testutils

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/utils/sdk/pipelineschecksextras"
)

// CheckPipelineCheckExistsWithName verifies that a service endpoint of a particular type exists in the state,
// and that it has the expected name when compared against the data in Azure DevOps.
func CheckPipelineCheckExistsWithName(tfNode string, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return fmt.Errorf("Did not find a check in the state")
		}

		check, err := getCheckFromState(resourceState)
		if err != nil {
			return err
		}

		if DisplayName, found := check.Settings.(map[string]interface{})["displayName"]; found {
			if DisplayName != expectedName {
				return fmt.Errorf("Check has Name=%s, but expected Name=%s", DisplayName, expectedName)
			}
		} else {
			return fmt.Errorf("displayName setting not found")
		}

		return nil
	}
}

// CheckPipelineCheckDestroyed verifies that all checks of the given type in the state are destroyed.
// This will be invoked *after* terraform destroys the resource but *before* the state is wiped clean.
func CheckPipelineCheckDestroyed(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, res := range s.RootModule().Resources {
			if res.Type != resourceType {
				continue
			}

			// indicates the resource exists - this should fail the test
			if _, err := getCheckFromState(res); err == nil {
				return fmt.Errorf("Unexpectedly found a check that should have been deleted")
			}
		}

		return nil
	}
}

// GetADOClientsFromEnv builds an AggregatedClient directly from environment variables.
// This is necessary when using GetMuxedProviderFactories() because the SDKv2 provider
// singleton does not have its Meta() configured in that case.
// Exported so acceptance test files in the parent package can call it directly for live evidence capture.
func GetADOClientsFromEnv() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	if orgURL == "" || pat == "" {
		return nil, fmt.Errorf("AZDO_ORG_SERVICE_URL and AZDO_PERSONAL_ACCESS_TOKEN must be set")
	}
	authProvider := azuredevops.NewAuthProviderPAT(pat)
	return client.GetAzdoClient(authProvider, orgURL)
}

// getADOClientsFromEnv is an unexported alias of GetADOClientsFromEnv for use within
// this package (testutils internal callers).
func getADOClientsFromEnv() (*client.AggregatedClient, error) {
	return GetADOClientsFromEnv()
}

// given a resource from the state, return a check (and error)
func getCheckFromState(resource *terraform.ResourceState) (*pipelineschecksextras.CheckConfiguration, error) {
	branchControlCheckID, err := strconv.Atoi(resource.Primary.ID)
	if err != nil {
		return nil, err
	}

	projectID := resource.Primary.Attributes["project_id"]

	// Build the client directly from environment variables rather than using
	// GetProvider().Meta(), which is nil when tests use GetMuxedProviderFactories().
	clients, err := getADOClientsFromEnv()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	return clients.PipelinesChecksClientExtras.GetCheckConfiguration(ctx, pipelineschecksextras.GetCheckConfigurationArgs{
		Project: &projectID,
		Id:      &branchControlCheckID,
		Expand:  converter.ToPtr(pipelineschecksextras.CheckConfigurationExpandParameterValues.Settings),
	})
}
