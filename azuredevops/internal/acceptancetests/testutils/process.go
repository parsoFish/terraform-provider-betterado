package testutils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// CheckProcessDestroyed verifies that all processes referenced in the state are destroyed. This will be invoked
// *after* terraform destroys the resource but *before* the state is wiped clean.
//
// Note: builds the client directly from environment variables so this function works correctly
// regardless of whether the test uses ProviderFactories (SDKv2 singleton) or ProtoV6ProviderFactories
// (muxed provider), since the latter does not configure the provider singleton and
// GetProvider().Meta() would be nil.
func CheckProcessDestroyed(s *terraform.State) error {
	// Load secrets.env so credentials are available when running under forge.
	loadSecretsEnv()

	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	authProvider := azuredevops.NewAuthProviderPAT(pat)
	clients, err := client.GetAzdoClient(authProvider, orgURL)
	if err != nil {
		return fmt.Errorf("CheckProcessDestroyed: building client: %w", err)
	}

	timeout := 10 * time.Second

	// verify that every process referenced in the state does not exist in AzDO
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "betterado_workitemtrackingprocess_process" {
			continue
		}

		id, err := uuid.Parse(resource.Primary.ID)
		if err != nil {
			return err
		}

		err = retry.RetryContext(clients.Ctx, timeout, func() *retry.RetryError {
			_, err := readProcess(clients, id)
			if err == nil {
				return retry.RetryableError(fmt.Errorf("process with ID %s should not exist", id.String()))
			}
			if utils.ResponseWasNotFound(err) {
				return nil
			}

			return retry.NonRetryableError(err)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func readProcess(clients *client.AggregatedClient, identifier uuid.UUID) (*core.Process, error) {
	return clients.CoreClient.GetProcessById(clients.Ctx, core.GetProcessByIdArgs{
		ProcessId: &identifier,
	})
}

// GetDirectClient builds an AggregatedClient directly from AZDO environment
// variables. It is exported so that acceptance-test files without a build tag
// can use it without relying on the tagged resource_task_group_test.go
// definition of getDirectClient().
func GetDirectClient() (*client.AggregatedClient, error) {
	loadSecretsEnv()
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}

func GenerateWorkItemTypeName() string {
	return strings.ReplaceAll(GenerateResourceName(), "-", "")
}
