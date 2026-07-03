package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ── Policy type IDs ───────────────────────────────────────────────────────────
// These UUIDs identify each repository policy type in Azure DevOps.
// https://docs.microsoft.com/en-us/rest/api/azure/devops/policy/types/list?view=azure-devops-rest-5.1
var (
	AuthorEmailPattern = uuid.MustParse("77ed4bd3-b063-4689-934a-175e4d0a78d7")
	FilePathPattern    = uuid.MustParse("51c78909-e838-41a2-9496-c647091e3c61")
	CaseEnforcement    = uuid.MustParse("7ed39669-655c-494e-b4a0-a08b4da0fcce")
	ReservedNames      = uuid.MustParse("db2b9b4c-180d-4529-9701-01541d19f36b")
	PathLength         = uuid.MustParse("001a79cf-fda1-4c4e-9e7c-bac40ee5ead8")
	FileSize           = uuid.MustParse("2e26e725-8201-4edd-8bf5-978563c34a80")
	CheckCredentials   = uuid.MustParse("e67ae10f-cf9a-40bc-8e66-6b3a8216956e")
)

// ── Import state helper ──────────────────────────────────────────────────────

// importRepoPolicyState provides a generic ImportState for repository policy resources.
// Import ID format: <project_id>/<policy_id>
func importRepoPolicyState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Expected format: <project_id>/<policy_id>, got: %s", req.ID))
		return
	}
	projectID := parts[0]
	policyIDStr := parts[1]
	if _, err := strconv.Atoi(policyIDStr); err != nil {
		resp.Diagnostics.AddError("Invalid policy ID in import",
			fmt.Sprintf("policy ID must be an integer, got: %s", policyIDStr))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), projectID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), policyIDStr)...)
}

// ── Static default helpers ────────────────────────────────────────────────────

// staticRepoPolicyBoolDefault implements defaults.Bool for a constant bool value.
type staticRepoPolicyBoolDefault struct{ value bool }

func staticRepoPolicyBool(v bool) defaults.Bool { return staticRepoPolicyBoolDefault{value: v} }

func (d staticRepoPolicyBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", d.value)
}

func (d staticRepoPolicyBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%v`", d.value)
}

func (d staticRepoPolicyBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.value)
}

// emptyRepoPolicyListDefault implements defaults.List returning an empty list of strings.
// Without this default, omitting repository_ids from config causes Terraform to mark the
// planned value as unknown (computed but not yet known), which causes ElementsAs to fail
// with "Received unknown value, however the target type cannot handle unknown values."
type emptyRepoPolicyListDefault struct{}

func emptyRepoPolicyList() defaults.List { return emptyRepoPolicyListDefault{} }

func (emptyRepoPolicyListDefault) Description(_ context.Context) string {
	return "defaults to empty list (project-wide policy)"
}

func (emptyRepoPolicyListDefault) MarkdownDescription(_ context.Context) string {
	return "defaults to empty list (project-wide policy)"
}

func (emptyRepoPolicyListDefault) DefaultList(ctx context.Context, _ defaults.ListRequest, resp *defaults.ListResponse) {
	listVal, diags := types.ListValueFrom(ctx, types.StringType, []string{})
	resp.Diagnostics.Append(diags...)
	resp.PlanValue = listVal
}

// ── Plan modifier helpers ─────────────────────────────────────────────────────

// repoPolicyUseStateForUnknownModifier is a plan modifier that uses the prior
// state value when the plan value is unknown.
type repoPolicyUseStateForUnknownModifier struct{}

func repoPolicyUseStateForUnknown() planmodifier.String {
	return repoPolicyUseStateForUnknownModifier{}
}

func (m repoPolicyUseStateForUnknownModifier) Description(_ context.Context) string {
	return "use prior state value when unknown"
}

func (m repoPolicyUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "use prior state value when unknown"
}

func (m repoPolicyUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// repoPolicyRequiresReplaceModifier marks the resource for replacement when
// the attribute changes.
type repoPolicyRequiresReplaceModifier struct{}

func repoPolicyRequiresReplace() planmodifier.String { return repoPolicyRequiresReplaceModifier{} }

func (m repoPolicyRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m repoPolicyRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m repoPolicyRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ── Scope (repository_ids) helpers ───────────────────────────────────────────

// expandRepositoryIDs converts a types.List of strings into the API representation:
// a []map[string]interface{} where each entry has a "repositoryId" key.
// If the list is empty, returns a single scope entry with empty repositoryId (project-wide policy).
func expandRepositoryIDs(ctx context.Context, repoIDs types.List) ([]map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	var ids []string
	diags.Append(repoIDs.ElementsAs(ctx, &ids, false)...)
	if diags.HasError() {
		return nil, diags
	}

	if len(ids) == 0 {
		return []map[string]interface{}{
			{"repositoryId": ""},
		}, diags
	}

	scopes := make([]map[string]interface{}, len(ids))
	for i, id := range ids {
		scopes[i] = map[string]interface{}{
			"repositoryId": id,
		}
	}
	return scopes, diags
}

// flattenRepositoryIDs converts the raw policy settings map into a types.List of string IDs.
func flattenRepositoryIDs(ctx context.Context, raw map[string]interface{}) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	scopesRaw, ok := raw["scope"]
	if !ok || scopesRaw == nil {
		emptyList, d := types.ListValueFrom(ctx, types.StringType, []string{})
		diags.Append(d...)
		return emptyList, diags
	}

	scopesSlice, ok := scopesRaw.([]interface{})
	if !ok {
		emptyList, d := types.ListValueFrom(ctx, types.StringType, []string{})
		diags.Append(d...)
		return emptyList, diags
	}

	var ids []string
	for _, s := range scopesSlice {
		sm, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		if v, ok := sm["repositoryId"]; ok && v != nil {
			id := fmt.Sprintf("%v", v)
			if id != "" {
				ids = append(ids, id)
			}
		}
	}

	// Ensure we always return an empty list (not null) when there are no repo IDs.
	// types.ListValueFrom with a nil slice returns a null list, which causes
	// "provider produced inconsistent result after apply" when the plan had an
	// empty list. Explicitly convert nil → []string{} to get a known empty list.
	if ids == nil {
		ids = []string{}
	}
	listVal, d := types.ListValueFrom(ctx, types.StringType, ids)
	diags.Append(d...)
	return listVal, diags
}

// boolPtr is a convenience helper to get a bool pointer.
func boolPtr(b bool) *bool { return &b }
