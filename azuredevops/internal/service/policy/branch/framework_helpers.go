package branch

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ── Policy type IDs ───────────────────────────────────────────────────────────
// These UUIDs identify each branch policy type in Azure DevOps.
// https://docs.microsoft.com/en-us/rest/api/azure/devops/policy/types/list?view=azure-devops-rest-5.1
var (
	MinReviewerCount  = uuid.MustParse("fa4e907d-c16b-4a4c-9dfa-4906e5d171dd")
	BuildValidation   = uuid.MustParse("0609b952-1397-4640-95ec-e00a01b2c241")
	AutoReviewers     = uuid.MustParse("fd2167ab-b0be-447a-8ec8-39368250530e")
	WorkItemLinking   = uuid.MustParse("40e92b44-2fe1-4dd6-b3d8-74a9c21d0c6e")
	CommentResolution = uuid.MustParse("c6a1889d-b943-4856-b76f-9e46bb6b0df2")
	MergeTypes        = uuid.MustParse("fa4e907d-c16b-4a4c-9dfa-4916e5d171ab")
	StatusCheck       = uuid.MustParse("cbdc66da-9728-4af8-aada-9a5a32e4a226")
)

// ── Policy settings structs (JSON wire format) ────────────────────────────────

type autoReviewerPolicySettings struct {
	SubmitterCanVote     bool     `json:"creatorVoteCounts"`
	AutoReviewerIds      []string `json:"requiredReviewerIds"`
	PathFilters          []string `json:"filenamePatterns"`
	DisplayMessage       string   `json:"message"`
	MinimumApproverCount int      `json:"minimumApproverCount"`
}

type buildValidationPolicySettings struct {
	BuildDefinitionID       int      `json:"buildDefinitionId"`
	PolicyDisplayName       string   `json:"displayName"`
	ManualQueueOnly         bool     `json:"manualQueueOnly"`
	QueueOnSourceUpdateOnly bool     `json:"queueOnSourceUpdateOnly"`
	ValidDuration           int      `json:"validDuration"`
	FilenamePatterns        []string `json:"filenamePatterns"`
}

type mergeTypePolicySettings struct {
	AllowSquash        bool `json:"allowSquash"`
	AllowRebase        bool `json:"allowRebase"`
	AllowNoFastForward bool `json:"allowNoFastForward"`
	AllowRebaseMerge   bool `json:"allowRebaseMerge"`
}

// ── Scope helpers shared across all branch policy framework resources ─────────

// scopeModel is the framework state model for a single scope entry.
type scopeModel struct {
	RepositoryID  types.String `tfsdk:"repository_id"`
	RepositoryRef types.String `tfsdk:"repository_ref"`
	MatchType     types.String `tfsdk:"match_type"`
}

// scopeAttrTypes returns the attr.Type map for scopeModel.
func scopeAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"repository_id":  types.StringType,
		"repository_ref": types.StringType,
		"match_type":     types.StringType,
	}
}

// expandScopesFramework converts a types.List of scopeModel into the API
// representation ([]map[string]interface{}) as expected by the policy settings.
func expandScopesFramework(ctx context.Context, scopeList types.List) ([]map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	var scopes []scopeModel
	diags.Append(scopeList.ElementsAs(ctx, &scopes, false)...)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]map[string]interface{}, len(scopes))
	for i, s := range scopes {
		entry := map[string]interface{}{}
		repoID := s.RepositoryID.ValueString()
		repoRef := s.RepositoryRef.ValueString()
		matchType := s.MatchType.ValueString()

		if repoID != "" {
			entry["repositoryId"] = repoID
		} else {
			entry["repositoryId"] = nil
		}
		if repoRef != "" {
			entry["refName"] = repoRef
		} else {
			entry["refName"] = nil
		}
		if matchType != "" {
			entry["matchKind"] = matchType
		} else {
			entry["matchKind"] = nil
		}

		if strings.EqualFold(matchType, "DefaultBranch") && (repoID != "" || repoRef != "") {
			diags.AddError("Invalid scope configuration",
				"neither 'repository_id' nor 'repository_ref' can be set when 'match_type=DefaultBranch'")
			return nil, diags
		}
		result[i] = entry
	}
	return result, diags
}

// flattenScopesFramework converts the raw policy settings map into a types.List of scopeModel.
func flattenScopesFramework(ctx context.Context, raw map[string]interface{}) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: scopeAttrTypes()}

	scopesRaw, ok := raw["scope"]
	if !ok || scopesRaw == nil {
		emptyList, d := types.ListValueFrom(ctx, elemType, []scopeModel{})
		diags.Append(d...)
		return emptyList, diags
	}

	scopesSlice, ok := scopesRaw.([]interface{})
	if !ok {
		emptyList, d := types.ListValueFrom(ctx, elemType, []scopeModel{})
		diags.Append(d...)
		return emptyList, diags
	}

	models := make([]scopeModel, len(scopesSlice))
	for i, s := range scopesSlice {
		sm := s.(map[string]interface{})
		var repoID, repoRef, matchType string
		if v, ok := sm["repositoryId"]; ok && v != nil {
			repoID = fmt.Sprintf("%v", v)
		}
		if v, ok := sm["refName"]; ok && v != nil {
			repoRef = fmt.Sprintf("%v", v)
		}
		if v, ok := sm["matchKind"]; ok && v != nil {
			matchType = fmt.Sprintf("%v", v)
		}
		models[i] = scopeModel{
			RepositoryID:  types.StringValue(repoID),
			RepositoryRef: types.StringValue(repoRef),
			MatchType:     types.StringValue(matchType),
		}
	}

	listVal, d := types.ListValueFrom(ctx, elemType, models)
	diags.Append(d...)
	return listVal, diags
}

// ── Coerce helpers ───────────────────────────────────────────────────────────

// boolCoerce safely converts an interface{} from JSON to bool.
func boolCoerce(v interface{}) bool {
	if v == nil {
		return false
	}
	switch b := v.(type) {
	case bool:
		return b
	case float64:
		return b != 0
	}
	return false
}

// int64Coerce safely converts an interface{} from JSON (float64) to int64.
func int64Coerce(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case int64:
		return n
	}
	return 0
}

// stringCoerce safely converts an interface{} from JSON to string.
func stringCoerce(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// ── Import state helper ──────────────────────────────────────────────────────

// importPolicyState provides a generic ImportState for policy resources.
// Import ID format: <project_id>/<policy_id>
func importPolicyState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

// staticPolicyBoolDefault implements defaults.Bool for a constant bool value.
type staticPolicyBoolDefault struct{ value bool }

func staticPolicyBool(v bool) defaults.Bool { return staticPolicyBoolDefault{value: v} }

func (d staticPolicyBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", d.value)
}

func (d staticPolicyBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%v`", d.value)
}

func (d staticPolicyBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.value)
}

// staticPolicyInt64Default implements defaults.Int64 for a constant int64 value.
type staticPolicyInt64Default struct{ value int64 }

func staticPolicyInt64(v int64) defaults.Int64 { return staticPolicyInt64Default{value: v} }

func (d staticPolicyInt64Default) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %d", d.value)
}

func (d staticPolicyInt64Default) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%d`", d.value)
}

func (d staticPolicyInt64Default) DefaultInt64(_ context.Context, _ defaults.Int64Request, resp *defaults.Int64Response) {
	resp.PlanValue = types.Int64Value(d.value)
}

// staticPolicyStringDefault implements defaults.String for a constant string value.
type staticPolicyStringDefault struct{ value string }

func staticPolicyString(v string) defaults.String { return staticPolicyStringDefault{value: v} }

func (d staticPolicyStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d staticPolicyStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%q`", d.value)
}

func (d staticPolicyStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// ── Plan modifier helpers ─────────────────────────────────────────────────────

// policyUseStateForUnknownModifier is a plan modifier that uses the prior
// state value when the plan value is unknown.
type policyUseStateForUnknownModifier struct{}

func policyUseStateForUnknown() planmodifier.String { return policyUseStateForUnknownModifier{} }

func (m policyUseStateForUnknownModifier) Description(_ context.Context) string {
	return "use prior state value when unknown"
}

func (m policyUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "use prior state value when unknown"
}

func (m policyUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// policyRequiresReplaceModifier marks the resource for replacement when
// the attribute changes.
type policyRequiresReplaceModifier struct{}

func policyRequiresReplace() planmodifier.String { return policyRequiresReplaceModifier{} }

func (m policyRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m policyRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m policyRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}
