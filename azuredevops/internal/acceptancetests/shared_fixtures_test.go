package acceptancetests

import (
	"os"
	"testing"

	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// TestSharedReleaseFixture is a live acceptance smoke test that validates:
//   - SharedReleaseFixture provisions all required resources (non-zero IDs) — AC1
//   - the returned release definition has pre AND post approvals on every stage — AC2 / VS402877
//   - every stage has a retention_policy — AC2 / VS402982
//   - the test skips without TF_ACC — AC3 (enforced inside SharedReleaseFixture)
func TestSharedReleaseFixture(t *testing.T) {
	result := SharedReleaseFixture(t)

	// AC1: all IDs must be non-zero / non-empty
	if result.ProjectID == "" {
		t.Error("SharedReleaseFixture: ProjectID is empty")
	}
	if result.RepoID == "" {
		t.Error("SharedReleaseFixture: RepoID is empty")
	}
	if result.BuildDefinitionID == 0 {
		t.Error("SharedReleaseFixture: BuildDefinitionID is 0")
	}
	if result.VariableGroupID == 0 {
		t.Error("SharedReleaseFixture: VariableGroupID is 0")
	}
	if result.ReleaseDefinitionID == 0 {
		t.Error("SharedReleaseFixture: ReleaseDefinitionID is 0")
	}

	// AC2: verify the release definition structure via the ADO API.
	// Build the client directly using the same env vars as SharedReleaseFixture.
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	authProvider := azuredevops.NewAuthProviderPAT(pat)
	clients, err := client.GetAzdoClient(authProvider, orgURL)
	if err != nil {
		t.Fatalf("TestSharedReleaseFixture: GetAzdoClient: %v", err)
	}
	def, apiErr := clients.ReleaseClient.GetReleaseDefinition(clients.Ctx, releaseapi.GetReleaseDefinitionArgs{
		Project:      &result.ProjectID,
		DefinitionId: &result.ReleaseDefinitionID,
	})
	if apiErr != nil {
		t.Fatalf("SharedReleaseFixture: GetReleaseDefinition: %v", apiErr)
	}
	if def.Environments == nil || len(*def.Environments) == 0 {
		t.Fatal("SharedReleaseFixture: release definition has no environments")
	}

	for i, env := range *def.Environments {
		stageName := "(unknown)"
		if env.Name != nil {
			stageName = *env.Name
		}

		// AC2: pre_deploy_approval must be present with at least one approver (VS402877)
		if env.PreDeployApprovals == nil {
			t.Errorf("stage[%d] %q: PreDeployApprovals is nil (VS402877)", i, stageName)
		} else if env.PreDeployApprovals.Approvals == nil || len(*env.PreDeployApprovals.Approvals) == 0 {
			t.Errorf("stage[%d] %q: PreDeployApprovals.Approvals is empty (VS402877)", i, stageName)
		}

		// AC2: post_deploy_approval must be present with at least one approver (VS402877)
		if env.PostDeployApprovals == nil {
			t.Errorf("stage[%d] %q: PostDeployApprovals is nil (VS402877)", i, stageName)
		} else if env.PostDeployApprovals.Approvals == nil || len(*env.PostDeployApprovals.Approvals) == 0 {
			t.Errorf("stage[%d] %q: PostDeployApprovals.Approvals is empty (VS402877)", i, stageName)
		}

		// AC2: retention_policy must be present (VS402982)
		if env.RetentionPolicy == nil {
			t.Errorf("stage[%d] %q: RetentionPolicy is nil (VS402982)", i, stageName)
		}
	}
}
