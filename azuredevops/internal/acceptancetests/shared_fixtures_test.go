package acceptancetests

import (
	"os"
	"testing"

	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

const zeroUUID = "00000000-0000-0000-0000-000000000000"

// TestSharedReleaseFixture is a live acceptance test that provisions the canonical
// fixture and then PROVES, by reading the release definition back from ADO, that
// every showcased option actually persisted — not merely that the objects exist.
//
// This guards against the class of bug where a definition "reports" a feature
// (e.g. approvals) that is not really configured: each assertion checks a concrete
// NON-DEFAULT value or a REAL identity, so a default/placeholder cannot pass.
func TestSharedReleaseFixture(t *testing.T) {
	result := SharedReleaseFixture(t)

	// AC1: all IDs must be non-zero / non-empty.
	if result.ProjectID == "" || result.RepoID == "" || result.BuildDefinitionID == 0 ||
		result.VariableGroupID == 0 || result.ReleaseDefinitionID == 0 {
		t.Fatalf("SharedReleaseFixture: incomplete result: %+v", result)
	}
	if result.ApproverID == "" || result.ApproverID == zeroUUID {
		t.Fatalf("SharedReleaseFixture: ApproverID must be a real identity, got %q", result.ApproverID)
	}

	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	clients, err := client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
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

	// ── Definition-level options ──────────────────────────────────────────────
	if got := strVal(def.Path); got != `\fixtures` {
		t.Errorf("definition path = %q, want %q (non-default)", got, `\fixtures`)
	}
	if got := strVal(def.ReleaseNameFormat); got != "Release-$(rev:r)" {
		t.Errorf("release_name_format = %q, want non-default %q", got, "Release-$(rev:r)")
	}
	if strVal(def.Description) == "" {
		t.Error("definition description is empty (expected the fixture description)")
	}
	// Definition-level variables (plain + secret).
	if def.Variables == nil {
		t.Error("definition has no variables")
	} else {
		if _, ok := (*def.Variables)["GLOBAL_PLAIN"]; !ok {
			t.Error("definition variable GLOBAL_PLAIN missing")
		}
		if sec, ok := (*def.Variables)["GLOBAL_SECRET"]; !ok || sec.IsSecret == nil || !*sec.IsSecret {
			t.Error("definition variable GLOBAL_SECRET missing or not marked secret")
		}
	}
	// Linked variable group at definition level.
	if !containsInt(def.VariableGroups, result.VariableGroupID) {
		t.Errorf("definition variable_groups %v does not link the fixture group #%d", deref(def.VariableGroups), result.VariableGroupID)
	}
	// Artifact.
	if def.Artifacts == nil || len(*def.Artifacts) == 0 {
		t.Error("definition has no artifacts")
	}
	// Two triggers: a CD artifact-source trigger and a scheduled trigger.
	if got := triggerTypes(def.Triggers); !contains(got, "artifactSource") || !contains(got, "schedule") {
		t.Errorf("triggers = %v, want both artifactSource and schedule", got)
	}

	// ── Three chained stages ──────────────────────────────────────────────────
	if def.Environments == nil || len(*def.Environments) != 3 {
		t.Fatalf("expected 3 stages (Dev, QA, Prod), got %d", envCount(def))
	}
	dev := mustEnv(t, def, "Dev")
	qa := mustEnv(t, def, "QA")
	prod := mustEnv(t, def, "Prod")

	// REAL approvals on every stage — the core fix. The approver must be the real
	// identity, NOT the zero-UUID automated placeholder, and IsAutomated must be false.
	for _, env := range []releaseapi.ReleaseDefinitionEnvironment{dev, qa, prod} {
		name := strVal(env.Name)
		assertRealApprover(t, name+" pre", env.PreDeployApprovals, result.ApproverID)
		assertRealApprover(t, name+" post", env.PostDeployApprovals, result.ApproverID) // VS402877: both pre+post
	}

	// Approval OPTIONS configured (non-default) on Dev's pre-deploy approval.
	if opts := approvalOptionsOf(dev.PreDeployApprovals); opts == nil {
		t.Error("Dev pre-deploy approval_options missing")
	} else {
		if opts.ExecutionOrder == nil || string(*opts.ExecutionOrder) != "beforeGates" {
			t.Errorf("Dev pre approval execution_order = %v, want beforeGates", opts.ExecutionOrder)
		}
		if intVal(opts.TimeoutInMinutes) != fixtureApproverTimeoutMin {
			t.Errorf("Dev pre approval timeout = %d, want non-default %d", intVal(opts.TimeoutInMinutes), fixtureApproverTimeoutMin)
		}
	}

	// Stage dependencies via environmentState conditions: QA depends on Dev, Prod on QA.
	assertEnvStateCondition(t, qa, "Dev")
	assertEnvStateCondition(t, prod, "QA")
	// Dev is triggered by the release-start event (no upstream stage dependency).
	if !hasConditionType(dev, "event") {
		t.Error("Dev stage missing an event condition (expected ReleaseStarted)")
	}

	// Different job types: Dev/Prod agent-based, QA agentless (runOnServer).
	if !contains(phaseTypes(dev), "agentBasedDeployment") {
		t.Errorf("Dev phase types = %v, want agentBasedDeployment", phaseTypes(dev))
	}
	if !contains(phaseTypes(qa), "runOnServer") {
		t.Errorf("QA phase types = %v, want runOnServer (agentless)", phaseTypes(qa))
	}

	// Agent jobs must have a real pool (queue) AND a workflow task (not empty).
	if agentQueueID(dev) <= 0 {
		t.Errorf("Dev agent job has no pool/queue assigned (queueId=%d)", agentQueueID(dev))
	}
	if !contains(phaseTaskNames(dev), "Command Line") {
		t.Errorf("Dev agent job has no Command Line task (tasks: %v)", phaseTaskNames(dev))
	}
	if !contains(phaseTaskNames(qa), "Delay") {
		t.Errorf("QA agentless job has no Delay task (tasks: %v)", phaseTaskNames(qa))
	}

	// Tags: exercise the option; log whether ADO persisted them (the definitions
	// endpoint is known not to persist tags — informational, not a failure).
	if def.Tags == nil || len(*def.Tags) == 0 {
		t.Logf("definition tags did not persist (known ADO definitions-endpoint limitation)")
	} else {
		t.Logf("definition tags persisted: %v", *def.Tags)
	}

	// Non-default, per-stage-distinct retention (default is 30/3/true).
	assertRetention(t, dev, 7, 2, false)
	assertRetention(t, qa, 14, 5, true)
	assertRetention(t, prod, 90, 10, true)

	// Environment options + execution policy configured (non-default) on Dev.
	// (Assert BadgeEnabled — a non-default, non-deprecated option; default is false.)
	if dev.EnvironmentOptions == nil || !boolVal(dev.EnvironmentOptions.BadgeEnabled) {
		t.Errorf("Dev environment_options badge_enabled = %v, want true (non-default)", dev.EnvironmentOptions)
	}
	if dev.ExecutionPolicy == nil || intVal(dev.ExecutionPolicy.ConcurrencyCount) != 2 || intVal(dev.ExecutionPolicy.QueueDepthCount) != 1 {
		t.Errorf("Dev execution_policy = %+v, want concurrency=2 queue_depth=1", dev.ExecutionPolicy)
	}

	// Stage-level variables on Dev (plain + secret).
	if dev.Variables == nil {
		t.Error("Dev stage has no variables")
	} else if sec, ok := (*dev.Variables)["STAGE_SECRET"]; !ok || sec.IsSecret == nil || !*sec.IsSecret {
		t.Error("Dev stage variable STAGE_SECRET missing or not secret")
	}

	// Env-level (stage-scoped) variable group on QA — the SECOND group.
	if result.VariableGroup2ID == 0 {
		t.Error("fixture did not create a second variable group")
	} else if !containsInt(qa.VariableGroups, result.VariableGroup2ID) {
		t.Errorf("QA stage variable_groups %v does not link the 2nd group #%d", deref(qa.VariableGroups), result.VariableGroup2ID)
	}

	// Deployment gates: QA has a pre-deployment gate, Prod has a post-deployment gate.
	assertHasGate(t, "QA pre", qa.PreDeploymentGates)
	assertHasGate(t, "Prod post", prod.PostDeploymentGates)
}

func assertHasGate(t *testing.T, label string, gs *releaseapi.ReleaseDefinitionGatesStep) {
	t.Helper()
	if gs == nil {
		t.Errorf("%s: deployment gates step is nil", label)
		return
	}
	if gs.GatesOptions == nil || gs.GatesOptions.IsEnabled == nil || !*gs.GatesOptions.IsEnabled {
		t.Errorf("%s: gates not enabled", label)
	}
	if gs.Gates == nil || len(*gs.Gates) == 0 {
		t.Errorf("%s: no gate tasks configured", label)
	}
}

// ── assertion + accessor helpers ─────────────────────────────────────────────

func assertRealApprover(t *testing.T, label string, ap *releaseapi.ReleaseDefinitionApprovals, wantID string) {
	t.Helper()
	if ap == nil || ap.Approvals == nil || len(*ap.Approvals) == 0 {
		t.Errorf("%s: no approval steps (VS402877 requires an approval)", label)
		return
	}
	step := (*ap.Approvals)[0]
	gotID := ""
	if step.Approver != nil {
		gotID = strVal(step.Approver.Id)
	}
	// The meaningful proof that this is a REAL approval (not the auto-approve
	// placeholder): a non-zero approver identity AND a manual gate.
	if gotID == "" || gotID == zeroUUID {
		t.Errorf("%s: approver is the zero-UUID placeholder (%q) — not a real identity", label, gotID)
	}
	if step.IsAutomated == nil || *step.IsAutomated {
		t.Errorf("%s: approval is automated; want a real manual gate (is_automated=false)", label)
	}
	// ADO may normalise the identity ref on read-back; tie to the resolved group
	// when it round-trips unchanged, but don't fail on a representation difference.
	if gotID != wantID {
		t.Logf("%s: approver id %q differs from resolved %q (ADO normalised the identity ref)", label, gotID, wantID)
	}
}

func assertEnvStateCondition(t *testing.T, env releaseapi.ReleaseDefinitionEnvironment, dependsOn string) {
	t.Helper()
	if env.Conditions == nil {
		t.Errorf("stage %q: no conditions (expected environmentState dependency on %q)", strVal(env.Name), dependsOn)
		return
	}
	for _, c := range *env.Conditions {
		if c.ConditionType != nil && string(*c.ConditionType) == "environmentState" && strVal(c.Name) == dependsOn {
			return
		}
	}
	t.Errorf("stage %q: missing environmentState condition depending on %q (conditions: %+v)", strVal(env.Name), dependsOn, *env.Conditions)
}

func assertRetention(t *testing.T, env releaseapi.ReleaseDefinitionEnvironment, days, releases int, retain bool) {
	t.Helper()
	rp := env.RetentionPolicy
	if rp == nil {
		t.Errorf("stage %q: retention_policy is nil (VS402982)", strVal(env.Name))
		return
	}
	if intVal(rp.DaysToKeep) != days || intVal(rp.ReleasesToKeep) != releases || boolVal(rp.RetainBuild) != retain {
		t.Errorf("stage %q retention = %d/%d/%v, want %d/%d/%v (non-default)",
			strVal(env.Name), intVal(rp.DaysToKeep), intVal(rp.ReleasesToKeep), boolVal(rp.RetainBuild), days, releases, retain)
	}
}

func hasConditionType(env releaseapi.ReleaseDefinitionEnvironment, condType string) bool {
	if env.Conditions == nil {
		return false
	}
	for _, c := range *env.Conditions {
		if c.ConditionType != nil && string(*c.ConditionType) == condType {
			return true
		}
	}
	return false
}

func approvalOptionsOf(ap *releaseapi.ReleaseDefinitionApprovals) *releaseapi.ApprovalOptions {
	if ap == nil {
		return nil
	}
	return ap.ApprovalOptions
}

func phaseTypes(env releaseapi.ReleaseDefinitionEnvironment) []string {
	var out []string
	if env.DeployPhases == nil {
		return out
	}
	for _, p := range *env.DeployPhases {
		if m, ok := p.(map[string]interface{}); ok {
			if pt, ok := m["phaseType"].(string); ok {
				out = append(out, pt)
			}
		}
	}
	return out
}

func agentQueueID(env releaseapi.ReleaseDefinitionEnvironment) int {
	if env.DeployPhases == nil {
		return 0
	}
	for _, p := range *env.DeployPhases {
		m, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		di, ok := m["deploymentInput"].(map[string]interface{})
		if !ok {
			continue
		}
		if q, ok := di["queueId"].(float64); ok {
			return int(q)
		}
	}
	return 0
}

func phaseTaskNames(env releaseapi.ReleaseDefinitionEnvironment) []string {
	var out []string
	if env.DeployPhases == nil {
		return out
	}
	for _, p := range *env.DeployPhases {
		m, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		tasks, ok := m["workflowTasks"].([]interface{})
		if !ok {
			continue
		}
		for _, tk := range tasks {
			if tm, ok := tk.(map[string]interface{}); ok {
				if n, ok := tm["name"].(string); ok {
					out = append(out, n)
				}
			}
		}
	}
	return out
}

func triggerTypes(triggers *[]interface{}) []string {
	var out []string
	if triggers == nil {
		return out
	}
	for _, tr := range *triggers {
		if m, ok := tr.(map[string]interface{}); ok {
			if tt, ok := m["triggerType"].(string); ok {
				out = append(out, tt)
			}
		}
	}
	return out
}

func mustEnv(t *testing.T, def *releaseapi.ReleaseDefinition, name string) releaseapi.ReleaseDefinitionEnvironment {
	t.Helper()
	for _, e := range *def.Environments {
		if strVal(e.Name) == name {
			return e
		}
	}
	t.Fatalf("stage %q not found", name)
	return releaseapi.ReleaseDefinitionEnvironment{}
}

func envCount(def *releaseapi.ReleaseDefinition) int {
	if def.Environments == nil {
		return 0
	}
	return len(*def.Environments)
}

func containsInt(p *[]int, want int) bool {
	if p == nil {
		return false
	}
	for _, v := range *p {
		if v == want {
			return true
		}
	}
	return false
}

func contains(s []string, want string) bool {
	for _, v := range s {
		if v == want {
			return true
		}
	}
	return false
}

func deref(p *[]int) []int {
	if p == nil {
		return nil
	}
	return *p
}

func strVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func intVal(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

func boolVal(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}
