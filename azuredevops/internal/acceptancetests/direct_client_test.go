package acceptancetests

import (
	"os"

	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// getDirectClient builds an AggregatedClient directly from AZDO env vars.
// Used by CheckDestroy and evidence helpers in tests that use
// ProtoV6ProviderFactories (mux provider), where the SDKv2 provider's
// Meta() is not wired and cannot be used.
//
// This function is intentionally NOT build-tag-restricted so it is available
// to any test file in this package regardless of build tags.
//
// Note: resource_task_group_test.go used to define this; it has been moved
// here so non-tagged test files (e.g. resource_git_repository_test.go) can
// use it without a duplicate-symbol error.
func getDirectClient() (*client.AggregatedClient, error) {
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	return client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
}
