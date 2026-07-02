package acceptancetests

import (
	"os"
	"testing"

	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
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

// preCheckGitRepository is the standard PreCheck for git repository tests.
// It verifies required env vars and ensures the shared fixture project
// (SharedFixtureProjectName) exists in the live ADO environment, creating
// it on the first run if needed. This prevents the
// "Project with name betterado-standing-demo does not exist" error when
// the HCL uses `data "betterado_project"` with the shared project name.
//
// Called in place of testutils.PreCheck(t, nil) for all git tests that use
// data "betterado_project" { name = SharedFixtureProjectName }.
func preCheckGitRepository(t *testing.T) {
	t.Helper()
	testutils.PreCheck(t, nil)
	if os.Getenv("TF_ACC") == "" {
		return
	}
	clients, err := getDirectClient()
	if err != nil {
		t.Fatalf("preCheckGitRepository: build direct client: %v", err)
	}
	resolveOrCreateFixtureProject(t, clients)
}

// preCheckGitRepositoryWithEnvVars is like preCheckGitRepository but also
// checks for additional required environment variables (e.g. service connection
// credentials). Used by tests that need extra env vars beyond the standard set.
func preCheckGitRepositoryWithEnvVars(t *testing.T, extraVars []string) {
	t.Helper()
	testutils.PreCheck(t, &extraVars)
	if os.Getenv("TF_ACC") == "" {
		return
	}
	clients, err := getDirectClient()
	if err != nil {
		t.Fatalf("preCheckGitRepository: build direct client: %v", err)
	}
	resolveOrCreateFixtureProject(t, clients)
}
