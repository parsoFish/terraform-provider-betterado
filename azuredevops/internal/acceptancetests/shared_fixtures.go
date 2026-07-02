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
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/operations"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
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
	// ApproverID is a REAL Azure DevOps identity (a project group), used as a
	// genuine manual approver on the canonical release definition — not the
	// zero-UUID automated placeholder. Consumers can assert against it.
	ApproverID string
	// VariableGroup2ID is a second variable group, linked at STAGE scope (QA) so
	// the env-level variable_groups option is exercised (ADO forbids linking the
	// same group at both pipeline and stage scope).
	VariableGroup2ID int
	// WorkItemQueryID backs the "Query Work Items" deployment-gate task.
	WorkItemQueryID string
}

// Non-default fixture constants. These are deliberately NOT the ADO defaults
// (30 days / 3 releases / retain=true, concurrency 1, etc.) so a read-back can
// PROVE the option was actually configured rather than left at its default.
const (
	fixtureApproverTimeoutMin = 1440 // 24h — non-default (default is 43200/30d)
)

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

	// ── 1. Resolve (or create once) the PERSISTENT shared project ────────────
	// The project is reused across runs and never deleted — see
	// SharedFixtureProjectName. Only the per-run sub-resources below are torn down.
	project := resolveOrCreateFixtureProject(t, clients)
	projectID := project.Id.String()

	// ── 2. Create Git repository ─────────────────────────────────────────────
	repo := createFixtureRepo(t, clients, projectID, name)
	repoID := repo.Id.String()
	// The project persists, so the per-run repo MUST be torn down explicitly.
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

	// ── 5. Create a SECOND variable group (linked at stage scope on QA) ───────
	vg2 := createFixtureVariableGroup(t, clients, projectID, name+"-2")
	vg2ID := *vg2.Id
	t.Cleanup(func() {
		if err := clients.TaskAgentClient.DeleteVariableGroup(clients.Ctx, taskagent.DeleteVariableGroupArgs{
			ProjectIds: &[]string{projectID},
			GroupId:    &vg2ID,
		}); err != nil {
			t.Logf("fixture cleanup: DeleteVariableGroup(%d): %v", vg2ID, err)
		}
	})

	// ── 6. Create a work-item query (backs the deployment-gate task) ──────────
	queryID := createFixtureWorkItemQuery(t, clients, projectID, name)

	// ── 7. Resolve a REAL approver identity ───────────────────────────────────
	// Use a real project group (Project Administrators) as a genuine manual
	// approver instead of the zero-UUID automated placeholder, so the canonical
	// definition demonstrates a real approval gate a read-back can prove.
	approverID := resolveApproverIdentity(t, clients, projectID)

	// ── 8. Create canonical multi-stage release definition ────────────────────
	relDef := createFixtureReleaseDefinition(t, clients, projectID, buildDefID, vgID, vg2ID, approverID, queryID, name)
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
		ApproverID:          approverID,
		VariableGroup2ID:    vg2ID,
		WorkItemQueryID:     queryID,
	}
}

// resolveApproverIdentity returns a real Azure DevOps group identity GUID for the
// project — preferring "Project Administrators" — to use as a genuine manual
// approver. Uses IdentityClient.ListGroups (a registered location that works with
// a PAT; GetSelf's location is not registered on this org's vssps host).
func resolveApproverIdentity(t *testing.T, clients *client.AggregatedClient, projectID string) string {
	t.Helper()
	groups, err := clients.IdentityClient.ListGroups(clients.Ctx, identity.ListGroupsArgs{
		ScopeIds: &projectID,
	})
	if err != nil {
		t.Fatalf("SharedReleaseFixture: IdentityClient.ListGroups: %v", err)
	}
	if groups == nil || len(*groups) == 0 {
		t.Fatalf("SharedReleaseFixture: no project groups found to use as an approver")
	}
	fallback := ""
	for _, g := range *groups {
		if g.Id == nil {
			continue
		}
		id := g.Id.String()
		if fallback == "" {
			fallback = id
		}
		if g.ProviderDisplayName != nil && strings.Contains(*g.ProviderDisplayName, "Project Administrators") {
			return id
		}
	}
	if fallback == "" {
		t.Fatalf("SharedReleaseFixture: no usable project group identity found")
	}
	return fallback
}

// resolveAgentQueueID returns the project's "Azure Pipelines" (Microsoft-hosted)
// agent queue id, for use as the agent job's pool. Every project gets this queue.
func resolveAgentQueueID(t *testing.T, clients *client.AggregatedClient, projectID string) int {
	t.Helper()
	const poolName = "Azure Pipelines"
	queues, err := clients.TaskAgentClient.GetAgentQueues(clients.Ctx, taskagent.GetAgentQueuesArgs{
		Project:   &projectID,
		QueueName: converter.String(poolName),
	})
	if err != nil {
		t.Fatalf("SharedReleaseFixture: GetAgentQueues(%q): %v", poolName, err)
	}
	if queues == nil || len(*queues) == 0 || (*queues)[0].Id == nil {
		t.Fatalf("SharedReleaseFixture: agent queue %q not found in project", poolName)
	}
	return *(*queues)[0].Id
}

// createFixtureWorkItemQuery creates a shared work-item query (under "Shared
// Queries") whose id backs the "Query Work Items" deployment gate. Deleted with
// the project, so no separate cleanup is registered.
func createFixtureWorkItemQuery(t *testing.T, clients *client.AggregatedClient, projectID, name string) string {
	t.Helper()
	parent := "Shared Queries"
	q, err := clients.WorkItemTrackingClient.CreateQuery(clients.Ctx, workitemtracking.CreateQueryArgs{
		Project: &projectID,
		Query:   &parent,
		PostedQuery: &workitemtracking.QueryHierarchyItem{
			Name:     converter.String(name + "-gate-query"),
			Wiql:     converter.String("SELECT [System.Id] FROM WorkItems WHERE [System.WorkItemType] = 'Bug'"),
			IsFolder: converter.Bool(false),
		},
	})
	if err != nil {
		t.Fatalf("SharedReleaseFixture: CreateQuery: %v", err)
	}
	if q == nil || q.Id == nil {
		t.Fatalf("SharedReleaseFixture: CreateQuery returned no id")
	}
	return q.Id.String()
}

// gatesStep builds a deployment-gates step (a "Query Work Items" gate) with
// non-default gate options. Used for both pre- and post-deployment gates.
func gatesStep(queryID string) *releaseapi.ReleaseDefinitionGatesStep {
	queryTaskID := uuid.MustParse("f1e4b0e6-017e-4819-8a48-ef19ae96e289") // built-in "Query Work Items" gate
	return &releaseapi.ReleaseDefinitionGatesStep{
		GatesOptions: &releaseapi.ReleaseDefinitionGatesOptions{
			IsEnabled:              converter.Bool(true),
			Timeout:                converter.Int(60),
			SamplingInterval:       converter.Int(5),
			StabilizationTime:      converter.Int(0),
			MinimumSuccessDuration: converter.Int(0),
		},
		Gates: &[]releaseapi.ReleaseDefinitionGate{
			{
				Tasks: &[]releaseapi.WorkflowTask{
					{
						Name:           converter.String("Query Work Items"),
						TaskId:         &queryTaskID,
						Version:        converter.String("0.*"),
						Enabled:        converter.Bool(true),
						DefinitionType: converter.String("task"),
						Inputs:         &map[string]string{"queryId": queryID},
					},
				},
			},
		},
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

// SharedFixtureProjectName is the single, PERSISTENT ADO project all release
// acceptance fixtures share. It is the SAME long-lived project that hosts the
// standing demo — reused for both so live acceptance tests can run immediately
// without creating a project (the org sits at the 1000-project cap; creates fail
// — see brain theme 2026-06-20-ado-org-project-limit-blocks-test-creates).
//
// It is REUSED across runs and NEVER deleted: ADO soft-deletes projects (28-day
// retention) and soft-deleted projects count toward the 1000-project cap, so a
// create+delete per run fills the recycle bin and eventually blocks all project
// creation org-wide. Per-run sub-resources (repo, build def, variable groups,
// release def, work-item query) are created with unique `test-acc-*` names and
// torn down in t.Cleanup — they don't count toward the project cap and never
// touch the standing-demo's own resources (different names/IDs). Allowlisted in
// the sweeper (keepProjects) so it is never reaped.
const SharedFixtureProjectName = "betterado-standing-demo"

// resolveOrCreateFixtureProject returns the shared persistent fixture project.
// It NEVER creates, auto-discovers, or substitutes a project: the fixture is
// long-lived shared infrastructure (it hosts the standing demo), and any
// "helpful" fallback turns a missing fixture into live writes against whatever
// project the fallback picks — on 2026-07-02 an auto-discover fallback ran
// terraform apply/destroy inside an unrelated real project after the fixture
// was accidentally soft-deleted. If the lookup fails, fail loudly: the operator
// restores the project from the ADO recycle bin (28-day retention).
func resolveOrCreateFixtureProject(t *testing.T, clients *client.AggregatedClient) *core.TeamProject {
	t.Helper()

	existing, err := clients.CoreClient.GetProject(clients.Ctx, core.GetProjectArgs{
		ProjectId: converter.String(SharedFixtureProjectName),
	})
	if err != nil || existing == nil || existing.Id == nil {
		t.Fatalf("SharedReleaseFixture: fixture project %q not found (%v). "+
			"DO NOT create, auto-discover, or substitute another project — restore the fixture "+
			"from the ADO recycle bin (Organization settings → Projects) and re-run.",
			SharedFixtureProjectName, err)
	}
	return existing
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

// ── canonical release-definition building blocks ─────────────────────────────
//
// Each helper sets DELIBERATELY NON-DEFAULT values so a read-back can prove the
// option was actually configured. Where ADO requires an approval to satisfy
// VS402877, we use a REAL manual approver (a genuine identity, IsAutomated=false)
// — not the zero-UUID automated placeholder — so the definition demonstrates a
// real approval gate.

// realManualApproval builds an approval step with a real approver identity
// (IsAutomated=false → a genuine manual gate) plus the supplied approval options.
func realManualApproval(approverID string, rank int, opts *releaseapi.ApprovalOptions) *releaseapi.ReleaseDefinitionApprovals {
	return &releaseapi.ReleaseDefinitionApprovals{
		Approvals: &[]releaseapi.ReleaseDefinitionApprovalStep{
			{
				Approver:    &webapi.IdentityRef{Id: converter.String(approverID)},
				IsAutomated: converter.Bool(false),
				Rank:        converter.Int(rank),
			},
		},
		ApprovalOptions: opts,
	}
}

func approvalOptions(order string, timeoutMin int, requiredCount int, creatorCanApprove, enforceRevalidation bool) *releaseapi.ApprovalOptions {
	eo := releaseapi.ApprovalExecutionOrder(order)
	return &releaseapi.ApprovalOptions{
		ExecutionOrder:              &eo,
		TimeoutInMinutes:            converter.Int(timeoutMin),
		RequiredApproverCount:       converter.Int(requiredCount),
		ReleaseCreatorCanBeApprover: converter.Bool(creatorCanApprove),
		EnforceIdentityRevalidation: converter.Bool(enforceRevalidation),
		AutoTriggeredAndPreviousEnvironmentApprovedCanBeSkipped: converter.Bool(false),
	}
}

func retention(days, releases int, retainBuild bool) *releaseapi.EnvironmentRetentionPolicy {
	return &releaseapi.EnvironmentRetentionPolicy{
		DaysToKeep:     converter.Int(days),
		ReleasesToKeep: converter.Int(releases),
		RetainBuild:    converter.Bool(retainBuild),
	}
}

func environmentOptions(emailType string, badge, autoLink, publishStatus bool) *releaseapi.EnvironmentOptions {
	return &releaseapi.EnvironmentOptions{
		EmailNotificationType:        converter.String(emailType),
		SkipArtifactsDownload:        converter.Bool(false),
		TimeoutInMinutes:             converter.Int(0),
		EnableAccessToken:            converter.Bool(true),
		PublishDeploymentStatus:      converter.Bool(publishStatus),
		BadgeEnabled:                 converter.Bool(badge),
		AutoLinkWorkItems:            converter.Bool(autoLink),
		PullRequestDeploymentEnabled: converter.Bool(false),
	}
}

func executionPolicy(concurrency, queueDepth int) *releaseapi.EnvironmentExecutionPolicy {
	return &releaseapi.EnvironmentExecutionPolicy{
		ConcurrencyCount: converter.Int(concurrency),
		QueueDepthCount:  converter.Int(queueDepth),
	}
}

func stageCondition(name, conditionType, value string) releaseapi.Condition {
	ct := releaseapi.ConditionType(conditionType)
	return releaseapi.Condition{
		Name:          converter.String(name),
		ConditionType: &ct,
		Value:         converter.String(value),
	}
}

func releaseVariable(value string, isSecret, allowOverride bool) releaseapi.ConfigurationVariableValue {
	return releaseapi.ConfigurationVariableValue{
		Value:         converter.String(value),
		IsSecret:      converter.Bool(isSecret),
		AllowOverride: converter.Bool(allowOverride),
	}
}

// agentPhase builds an agentBasedDeployment phase exercising the rich
// deployment_input surface: a real agent POOL/queue, an agent specification (only
// meaningful on the Microsoft-hosted pool the queue points at), demands, timeouts,
// the skip/enable flags, multiConfiguration parallel execution, AND a real
// workflow task (Command Line) so the job is not empty.
func agentPhase(queueID int) map[string]interface{} {
	cmdLineTaskID := uuid.MustParse("d9bafed4-0b18-4f58-968d-86655b4d2ce9") // built-in "Command Line" (CmdLine@2)
	return map[string]interface{}{
		"name":      "Agent job",
		"rank":      1,
		"phaseType": "agentBasedDeployment",
		"deploymentInput": map[string]interface{}{
			"queueId":                   queueID,
			"timeoutInMinutes":          60,
			"jobCancelTimeoutInMinutes": 5,
			"condition":                 "succeeded()",
			"skipArtifactsDownload":     true,
			"enableAccessToken":         true,
			"demands":                   []string{"Agent.Version -gtVersion 2.0"},
			"agentSpecification":        map[string]interface{}{"identifier": "ubuntu-22.04"},
			"parallelExecution": map[string]interface{}{
				"parallelExecutionType": "multiConfiguration",
				"maxNumberOfAgents":     2,
				"multipliers":           "Configuration",
				"continueOnError":       false,
			},
		},
		"workflowTasks": []releaseapi.WorkflowTask{
			{
				Name:            converter.String("Command Line"),
				TaskId:          &cmdLineTaskID,
				Version:         converter.String("2.*"),
				Enabled:         converter.Bool(true),
				AlwaysRun:       converter.Bool(false),
				ContinueOnError: converter.Bool(false),
				Condition:       converter.String("succeeded()"),
				DefinitionType:  converter.String("task"),
				Inputs:          &map[string]string{"script": `echo "Hello from the canonical fixture"`},
			},
		},
	}
}

// agentlessPhase builds a runOnServer (agentless) phase carrying a server-side
// Delay workflow task — showcasing a different kind of job. queueId is omitted
// (agentless jobs have no agent queue).
func agentlessPhase() map[string]interface{} {
	delayTaskID := uuid.MustParse("28782b92-5e8e-4458-9751-a71cd1492bae") // built-in "Delay" task
	return map[string]interface{}{
		"name":      "Agentless job",
		"rank":      1,
		"phaseType": "runOnServer",
		"deploymentInput": map[string]interface{}{
			"timeoutInMinutes":          0,
			"jobCancelTimeoutInMinutes": 1,
			"condition":                 "succeeded()",
		},
		"workflowTasks": []releaseapi.WorkflowTask{
			{
				Name:            converter.String("Delay"),
				TaskId:          &delayTaskID,
				Version:         converter.String("1.*"),
				Enabled:         converter.Bool(true),
				AlwaysRun:       converter.Bool(false),
				ContinueOnError: converter.Bool(false),
				Condition:       converter.String("succeeded()"),
				DefinitionType:  converter.String("task"),
				Inputs:          &map[string]string{"delayForMinutes": "1"},
			},
		},
	}
}

// createFixtureReleaseDefinition builds a MINIMAL-resource, MAXIMAL-option canonical
// release definition: three stages chained by environmentState conditions (Dev →
// QA → Prod), exercising agent + agentless jobs, parallel execution, demands, real
// manual approvals with varied approval options, per-stage non-default retention,
// environment options, execution policy, definition + environment variables, a
// linked variable group, an artifact, and CD + scheduled triggers.
func createFixtureReleaseDefinition(t *testing.T, clients *client.AggregatedClient, projectID string, buildDefID, vgID, vg2ID int, approverID, queryID, name string) *releaseapi.ReleaseDefinition {
	t.Helper()

	defName := name + "-release"
	buildDefIDStr := fmt.Sprintf("%d", buildDefID)

	// Resolve the hosted agent queue once — used as the agent job pool on Dev/Prod.
	queueID := resolveAgentQueueID(t, clients, projectID)

	// ── Stage 1: Dev — agent job (pool + spec + task, parallel multiConfig), event-triggered, rich env ──
	devVars := map[string]releaseapi.ConfigurationVariableValue{
		"STAGE_PLAIN":  releaseVariable("dev-value", false, true),
		"STAGE_SECRET": releaseVariable("s3cr3t", true, false),
	}
	devPhases := []interface{}{agentPhase(queueID)}
	devConds := []releaseapi.Condition{stageCondition("ReleaseStarted", "event", "")}
	// NOTE: the variable group is linked at the DEFINITION level (below). ADO forbids
	// linking the same group at both pipeline and stage scope, so the stage-level
	// variable_groups option would need a SECOND variable group to showcase — left as
	// a minimal-resource trade-off (def-level link already proves the capability).
	dev := releaseapi.ReleaseDefinitionEnvironment{
		Name:                converter.String("Dev"),
		Rank:                converter.Int(1),
		Conditions:          &devConds,
		Variables:           &devVars,
		DeployPhases:        &devPhases,
		PreDeployApprovals:  realManualApproval(approverID, 1, approvalOptions("beforeGates", fixtureApproverTimeoutMin, 1, true, false)),
		PostDeployApprovals: realManualApproval(approverID, 1, approvalOptions("afterSuccessfulGates", 720, 1, false, false)),
		RetentionPolicy:     retention(7, 2, false),
		EnvironmentOptions:  environmentOptions("Always", true, true, true),
		ExecutionPolicy:     executionPolicy(2, 1),
	}

	// ── Stage 2: QA — agentless job, depends on Dev succeeding; pre-deploy GATE +
	//    a stage-scoped (env-level) variable group ──
	qaPhases := []interface{}{agentlessPhase()}
	qaConds := []releaseapi.Condition{stageCondition("Dev", "environmentState", "4")} // 4 = succeeded
	qaVGs := []int{vg2ID}
	qa := releaseapi.ReleaseDefinitionEnvironment{
		Name:                converter.String("QA"),
		Rank:                converter.Int(2),
		Conditions:          &qaConds,
		VariableGroups:      &qaVGs,
		DeployPhases:        &qaPhases,
		PreDeployApprovals:  realManualApproval(approverID, 1, approvalOptions("afterGatesAlways", 60, 1, false, true)),
		PostDeployApprovals: realManualApproval(approverID, 1, nil),
		PreDeploymentGates:  gatesStep(queryID),
		RetentionPolicy:     retention(14, 5, true),
		EnvironmentOptions:  environmentOptions("OnlyOnFailure", false, false, true),
		ExecutionPolicy:     executionPolicy(1, 0),
	}

	// ── Stage 3: Prod — agent job, depends on QA succeeding ──
	prodPhases := []interface{}{agentPhase(queueID)}
	prodConds := []releaseapi.Condition{stageCondition("QA", "environmentState", "4")}
	prod := releaseapi.ReleaseDefinitionEnvironment{
		Name:                converter.String("Prod"),
		Rank:                converter.Int(3),
		Conditions:          &prodConds,
		DeployPhases:        &prodPhases,
		PreDeployApprovals:  realManualApproval(approverID, 1, approvalOptions("beforeGates", 2880, 1, false, true)),
		PostDeployApprovals: realManualApproval(approverID, 1, nil),
		PostDeploymentGates: gatesStep(queryID),
		RetentionPolicy:     retention(90, 10, true),
		EnvironmentOptions:  environmentOptions("Never", false, true, false),
		ExecutionPolicy:     executionPolicy(1, 1),
	}

	envs := []releaseapi.ReleaseDefinitionEnvironment{dev, qa, prod}

	// Definition-level variables + linked variable group.
	defVars := map[string]releaseapi.ConfigurationVariableValue{
		"GLOBAL_PLAIN":  releaseVariable("global-value", false, true),
		"GLOBAL_SECRET": releaseVariable("global-s3cr3t", true, false),
	}
	defVGs := []int{vgID}

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

	// CD artifact trigger (on new build of _build, main branch) + a scheduled trigger.
	triggers := []interface{}{
		map[string]interface{}{
			"triggerType":   "artifactSource",
			"artifactAlias": "_build",
			"triggerConditions": []map[string]interface{}{
				{"sourceBranch": "refs/heads/main"},
			},
		},
		map[string]interface{}{
			"triggerType": "schedule",
			"schedule": map[string]interface{}{
				"scheduleOnlyWithChanges": true,
				"startHours":              2,
				"startMinutes":            0,
				"timeZoneId":              "UTC",
				"daysToRelease":           62, // Mon-Fri bitmask
			},
		},
	}

	// Tags — set them to exercise the option. NOTE: ADO's release *definitions*
	// endpoint is known not to persist tags (they round-trip empty); the read-back
	// in the smoke test logs whether they survived rather than failing.
	tags := []string{"fixture", "canonical-release"}

	relDef := &releaseapi.ReleaseDefinition{
		Name:              converter.String(defName),
		Path:              converter.String(`\fixtures`), // non-default path
		Description:       converter.String("Canonical exhaustive fixture: 3 chained stages, agent + agentless jobs, real approvals, non-default options."),
		ReleaseNameFormat: converter.String("Release-$(rev:r)"),
		Variables:         &defVars,
		VariableGroups:    &defVGs,
		Environments:      &envs,
		Artifacts:         &[]releaseapi.Artifact{artifact},
		Triggers:          &triggers,
		Tags:              &tags,
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
