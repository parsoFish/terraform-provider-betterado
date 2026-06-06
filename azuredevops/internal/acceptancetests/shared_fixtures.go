package acceptancetests

// shared_fixtures.go — canonical multi-stage release definition fixture for acceptance tests.
//
// SharedReleaseFixture provisions a real ADO project, Git repo, build definition, variable group,
// and a canonical multi-stage release definition that satisfies every known ADO REST API 7.x
// validity constraint:
//
//   - VS402877: every stage has BOTH pre_deploy_approval AND post_deploy_approval
//   - VS402982: every stage has a retention_policy block
//   - Correct permission key: EditReleaseEnvironment (not the stale EditReleaseStage)
//
// All resources are torn down via t.Cleanup so no orphaned cloud resources survive.
//
// AC3: if TF_ACC is not set the function calls t.Skip immediately, before any ADO API call.

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/operations"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// SharedFixtureResult holds the IDs returned by SharedReleaseFixture.
// Consuming tests can reference these IDs in HCL via string interpolation or
// direct SDK calls.
type SharedFixtureResult struct {
	ProjectID           string
	RepoID              string
	BuildDefinitionID   int
	VariableGroupID     int
	ReleaseDefinitionID int
}

// SharedReleaseFixture provisions a canonical multi-stage release definition
// and all its prerequisites in a fresh ADO project.
//
// REQUIRES: TF_ACC=1, AZDO_ORG_SERVICE_URL, AZDO_PERSONAL_ACCESS_TOKEN
//
// Every resource is registered with t.Cleanup for automatic teardown.
func SharedReleaseFixture(t *testing.T) SharedFixtureResult {
	t.Helper()

	// AC3: skip immediately if TF_ACC is not set — keep offline unit suite creds-free.
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping live fixture")
	}

	testutils.PreCheck(t, nil)

	// Build a dedicated ADO client from environment variables.
	// We build the client directly rather than relying on GetProvider().Meta() because
	// SharedReleaseFixture may be called from plain TestXxx functions that don't go
	// through the Terraform resource.Test lifecycle (which normally wires Meta).
	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	authProvider := azuredevops.NewAuthProviderPAT(pat)
	clients, err := client.GetAzdoClient(authProvider, orgURL)
	if err != nil {
		t.Fatalf("SharedReleaseFixture: GetAzdoClient: %v", err)
	}

	name := testutils.GenerateResourceName()

	// ── 1. Create project ────────────────────────────────────────────────────
	project := createFixtureProject(t, clients, name)
	projectID := project.Id.String()

	t.Cleanup(func() {
		if err := deleteFixtureProject(clients, projectID); err != nil {
			t.Logf("SharedReleaseFixture cleanup: failed to delete project %s: %v", projectID, err)
		}
	})

	// ── 2. Create Git repository ─────────────────────────────────────────────
	repo := createFixtureRepo(t, clients, projectID, name)
	repoID := repo.Id.String()
	// The repo is deleted as part of project deletion; register anyway for safety.
	t.Cleanup(func() {
		repoUUID, err := uuid.Parse(repoID)
		if err != nil {
			return
		}
		if err := clients.GitReposClient.DeleteRepository(clients.Ctx, git.DeleteRepositoryArgs{
			RepositoryId: &repoUUID,
		}); err != nil {
			t.Logf("fixture cleanup: DeleteRepository(%s): %v", repoID, err)
		}
	})

	// ── 3. Create build definition ───────────────────────────────────────────
	buildDef := createFixtureBuildDefinition(t, clients, projectID, repoID, name)
	buildDefID := *buildDef.Id

	t.Cleanup(func() {
		if err := clients.BuildClient.DeleteDefinition(clients.Ctx, build.DeleteDefinitionArgs{
			Project:      &projectID,
			DefinitionId: &buildDefID,
		}); err != nil {
			t.Logf("fixture cleanup: DeleteDefinition(%d): %v", buildDefID, err)
		}
	})

	// ── 4. Create variable group ─────────────────────────────────────────────
	vg := createFixtureVariableGroup(t, clients, projectID, name)
	vgID := *vg.Id

	t.Cleanup(func() {
		if err := clients.TaskAgentClient.DeleteVariableGroup(clients.Ctx, taskagent.DeleteVariableGroupArgs{
			ProjectIds: &[]string{projectID},
			GroupId:    &vgID,
		}); err != nil {
			t.Logf("fixture cleanup: DeleteVariableGroup(%d): %v", vgID, err)
		}
	})

	// ── 5. Create canonical multi-stage release definition ───────────────────
	relDef := createFixtureReleaseDefinition(t, clients, projectID, buildDefID, name)
	relDefID := *relDef.Id

	t.Cleanup(func() {
		if err := clients.ReleaseClient.DeleteReleaseDefinition(clients.Ctx, releaseapi.DeleteReleaseDefinitionArgs{
			Project:      &projectID,
			DefinitionId: &relDefID,
		}); err != nil {
			t.Logf("fixture cleanup: DeleteReleaseDefinition(%d): %v", relDefID, err)
		}
	})

	return SharedFixtureResult{
		ProjectID:           projectID,
		RepoID:              repoID,
		BuildDefinitionID:   buildDefID,
		VariableGroupID:     vgID,
		ReleaseDefinitionID: relDefID,
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func createFixtureProject(t *testing.T, clients *client.AggregatedClient, name string) *core.TeamProject {
	t.Helper()

	// Look up the "Agile" process template ID.
	processes, err := clients.CoreClient.GetProcesses(clients.Ctx, core.GetProcessesArgs{})
	if err != nil {
		t.Fatalf("SharedReleaseFixture: GetProcesses: %v", err)
	}
	var processTemplateID string
	for _, p := range *processes {
		if p.Name != nil && *p.Name == "Agile" {
			processTemplateID = p.Id.String()
			break
		}
	}
	if processTemplateID == "" {
		t.Fatalf("SharedReleaseFixture: could not find Agile process template")
	}

	visibility := core.ProjectVisibilityValues.Private
	vcType := "Git"
	project := &core.TeamProject{
		Name:        converter.String(name),
		Description: converter.String(name + "-shared-fixture"),
		Visibility:  &visibility,
		Capabilities: &map[string]map[string]string{
			"versioncontrol": {
				"sourceControlType": vcType,
			},
			"processTemplate": {
				"templateTypeId": processTemplateID,
			},
		},
	}

	operationRef, err := clients.CoreClient.QueueCreateProject(clients.Ctx, core.QueueCreateProjectArgs{
		ProjectToCreate: project,
	})
	if err != nil {
		t.Fatalf("SharedReleaseFixture: QueueCreateProject: %v", err)
	}

	stateConf := &retry.StateChangeConf{
		ContinuousTargetOccurence: 1,
		Delay:                     5 * time.Second,
		MinTimeout:                10 * time.Second,
		Timeout:                   10 * time.Minute,
		Pending: []string{
			string(operations.OperationStatusValues.InProgress),
			string(operations.OperationStatusValues.Queued),
			string(operations.OperationStatusValues.NotSet),
		},
		Target: []string{
			string(operations.OperationStatusValues.Failed),
			string(operations.OperationStatusValues.Succeeded),
			string(operations.OperationStatusValues.Cancelled),
		},
		Refresh: func() (interface{}, string, error) {
			ret, err := clients.OperationsClient.GetOperation(clients.Ctx, operations.GetOperationArgs{
				OperationId: operationRef.Id,
				PluginId:    operationRef.PluginId,
			})
			if err != nil {
				return nil, string(operations.OperationStatusValues.Failed), err
			}
			return ret, string(*ret.Status), nil
		},
	}
	if _, err := stateConf.WaitForStateContext(clients.Ctx); err != nil {
		t.Fatalf("SharedReleaseFixture: waiting for project creation: %v", err)
	}

	// Fetch the created project by name to obtain its UUID.
	created, err := clients.CoreClient.GetProject(clients.Ctx, core.GetProjectArgs{
		ProjectId: &name,
	})
	if err != nil {
		t.Fatalf("SharedReleaseFixture: GetProject after create: %v", err)
	}
	return created
}

func deleteFixtureProject(clients *client.AggregatedClient, projectID string) error {
	id, err := uuid.Parse(projectID)
	if err != nil {
		return fmt.Errorf("invalid project UUID %s: %w", projectID, err)
	}

	var operationRef *operations.OperationReference
	err = retry.RetryContext(clients.Ctx, 5*time.Minute, func() *retry.RetryError {
		var deleteErr error
		operationRef, deleteErr = clients.CoreClient.QueueDeleteProject(clients.Ctx, core.QueueDeleteProjectArgs{
			ProjectId: &id,
		})
		if deleteErr != nil {
			return retry.RetryableError(deleteErr)
		}
		return nil
	})
	if err != nil {
		return err
	}

	stateConf := &retry.StateChangeConf{
		ContinuousTargetOccurence: 1,
		Delay:                     10 * time.Second,
		MinTimeout:                10 * time.Second,
		Timeout:                   10 * time.Minute,
		Pending: []string{
			string(operations.OperationStatusValues.InProgress),
			string(operations.OperationStatusValues.Queued),
			string(operations.OperationStatusValues.NotSet),
		},
		Target: []string{
			string(operations.OperationStatusValues.Failed),
			string(operations.OperationStatusValues.Succeeded),
			string(operations.OperationStatusValues.Cancelled),
		},
		Refresh: func() (interface{}, string, error) {
			ret, err := clients.OperationsClient.GetOperation(clients.Ctx, operations.GetOperationArgs{
				OperationId: operationRef.Id,
				PluginId:    operationRef.PluginId,
			})
			if err != nil {
				return nil, string(operations.OperationStatusValues.Failed), err
			}
			return ret, string(*ret.Status), nil
		},
	}
	if _, err := stateConf.WaitForStateContext(clients.Ctx); err != nil {
		return fmt.Errorf("waiting for project delete: %w", err)
	}
	return nil
}

// mustParseUUID parses an ADO project/resource UUID inside a fixture, failing the
// test on a malformed value. Keeps the error checked (errcheck check-blank clean)
// while still being usable inline in a struct literal.
func mustParseUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("fixture: parsing UUID %q: %v", s, err)
	}
	return id
}

func createFixtureRepo(t *testing.T, clients *client.AggregatedClient, projectID, name string) *git.GitRepository {
	t.Helper()
	repo, err := clients.GitReposClient.CreateRepository(clients.Ctx, git.CreateRepositoryArgs{
		GitRepositoryToCreate: &git.GitRepositoryCreateOptions{
			Name: converter.String(name + "-repo"),
			Project: &core.TeamProjectReference{
				Id: func() *uuid.UUID { id := mustParseUUID(t, projectID); return &id }(),
			},
		},
		Project: &projectID,
	})
	if err != nil {
		t.Fatalf("SharedReleaseFixture: CreateRepository: %v", err)
	}
	return repo
}

func createFixtureBuildDefinition(t *testing.T, clients *client.AggregatedClient, projectID, repoID, name string) *build.BuildDefinition {
	t.Helper()
	repoType := "TfsGit"
	ymlPath := "azure-pipelines.yml"
	defaultBranch := "refs/heads/main"
	agentPool := "Azure Pipelines"
	defName := name + "-build"
	def := &build.BuildDefinition{
		Name: converter.String(defName),
		Queue: &build.AgentPoolQueue{
			Pool: &build.TaskAgentPoolReference{
				Name: converter.String(agentPool),
			},
		},
		Repository: &build.BuildRepository{
			Type:          converter.String(repoType),
			Id:            converter.String(repoID),
			DefaultBranch: converter.String(defaultBranch),
			Properties: &map[string]string{
				"reportBuildStatus": "true",
			},
		},
		Process: &build.YamlProcess{
			YamlFilename: converter.String(ymlPath),
		},
	}
	created, err := clients.BuildClient.CreateDefinition(clients.Ctx, build.CreateDefinitionArgs{
		Definition: def,
		Project:    &projectID,
	})
	if err != nil {
		t.Fatalf("SharedReleaseFixture: CreateDefinition: %v", err)
	}
	return created
}

func createFixtureVariableGroup(t *testing.T, clients *client.AggregatedClient, projectID, name string) *taskagent.VariableGroup {
	t.Helper()
	vgType := "Vsts"
	vgName := name + "-vg"
	vgParams := &taskagent.VariableGroupParameters{
		Name:        converter.String(vgName),
		Description: converter.String("shared fixture variable group"),
		Type:        converter.String(vgType),
		Variables: &map[string]interface{}{
			"FIXTURE_VAR": map[string]interface{}{
				"value":    "fixture-value",
				"isSecret": false,
			},
		},
		VariableGroupProjectReferences: &[]taskagent.VariableGroupProjectReference{
			{
				Name:        converter.String(vgName),
				Description: converter.String("shared fixture variable group"),
				ProjectReference: &taskagent.ProjectReference{
					Id: func() *uuid.UUID { id := mustParseUUID(t, projectID); return &id }(),
				},
			},
		},
	}
	created, err := clients.TaskAgentClient.AddVariableGroup(clients.Ctx, taskagent.AddVariableGroupArgs{
		VariableGroupParameters: vgParams,
	})
	if err != nil {
		t.Fatalf("SharedReleaseFixture: AddVariableGroup: %v", err)
	}
	return created
}

// canonicalApproval returns a ReleaseDefinitionApprovals with a single automated approver.
// The automated-approver pattern (UUID all-zeros + IsAutomated=true) is the pattern
// already used throughout the existing HCL acceptance tests.
// AC2: this is used for BOTH pre- and post-deploy approvals to satisfy VS402877.
func canonicalApproval() *releaseapi.ReleaseDefinitionApprovals {
	automated := true
	rank := 1
	return &releaseapi.ReleaseDefinitionApprovals{
		Approvals: &[]releaseapi.ReleaseDefinitionApprovalStep{
			{
				Approver: &webapi.IdentityRef{
					Id: converter.String("00000000-0000-0000-0000-000000000000"),
				},
				IsAutomated: &automated,
				Rank:        &rank,
			},
		},
	}
}

// canonicalRetentionPolicy returns the default retention policy for a stage.
// AC2: every stage must carry a retention_policy to satisfy VS402982.
func canonicalRetentionPolicy() *releaseapi.EnvironmentRetentionPolicy {
	return &releaseapi.EnvironmentRetentionPolicy{
		DaysToKeep:     converter.Int(30),
		ReleasesToKeep: converter.Int(3),
		RetainBuild:    converter.Bool(true),
	}
}

// canonicalStage builds a single ReleaseDefinitionEnvironment that satisfies
// all ADO REST API 7.x validity constraints:
//   - pre AND post deploy approvals (VS402877)
//   - retention_policy (VS402982)
//   - agentBasedDeployment phase (no queue needed — ADO defaults to any available)
func canonicalStage(stageName string, rank int) releaseapi.ReleaseDefinitionEnvironment {
	agentPhaseType := "agentBasedDeployment"
	phaseName := "Agent job"
	phaseRank := 1

	phase := map[string]interface{}{
		"name":      phaseName,
		"rank":      phaseRank,
		"phaseType": agentPhaseType,
	}
	phases := []interface{}{phase}

	return releaseapi.ReleaseDefinitionEnvironment{
		Name:                converter.String(stageName),
		Rank:                converter.Int(rank),
		PreDeployApprovals:  canonicalApproval(),
		PostDeployApprovals: canonicalApproval(),
		RetentionPolicy:     canonicalRetentionPolicy(),
		DeployPhases:        &phases,
	}
}

func createFixtureReleaseDefinition(t *testing.T, clients *client.AggregatedClient, projectID string, buildDefID int, name string) *releaseapi.ReleaseDefinition {
	t.Helper()

	defName := name + "-release"
	buildDefIDStr := fmt.Sprintf("%d", buildDefID)

	// Two-stage canonical release definition: Staging → Production.
	// AC2: each stage has both pre/post approvals AND a retention_policy.
	stagingStage := canonicalStage("Staging", 1)
	productionStage := canonicalStage("Production", 2)

	envs := []releaseapi.ReleaseDefinitionEnvironment{stagingStage, productionStage}

	// Build artifact ties the release definition to the build definition.
	artifact := releaseapi.Artifact{
		Alias:     converter.String("_build"),
		Type:      converter.String("Build"),
		IsPrimary: converter.Bool(true),
		DefinitionReference: &map[string]releaseapi.ArtifactSourceReference{
			"definition": {Id: converter.String(buildDefIDStr)},
			"project":    {Id: converter.String(projectID)},
		},
	}

	relDef := &releaseapi.ReleaseDefinition{
		Name:         converter.String(defName),
		Path:         converter.String(`\`),
		Environments: &envs,
		Artifacts:    &[]releaseapi.Artifact{artifact},
	}

	created, err := clients.ReleaseClient.CreateReleaseDefinition(clients.Ctx, releaseapi.CreateReleaseDefinitionArgs{
		ReleaseDefinition: relDef,
		Project:           &projectID,
	})
	if err != nil {
		t.Fatalf("SharedReleaseFixture: CreateReleaseDefinition: %v", err)
	}
	return created
}
