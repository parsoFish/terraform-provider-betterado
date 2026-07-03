package graph

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ── stringNotEmptyValidator ──────────────────────────────────────────────────

// stringNotEmptyValidator rejects empty string values (equivalent to SDKv2
// validation.NoZeroValues on TypeString attributes).
type stringNotEmptyValidator struct{}

func stringNotEmpty() validator.String { return stringNotEmptyValidator{} }

func (v stringNotEmptyValidator) Description(_ context.Context) string {
	return "value must not be empty"
}

func (v stringNotEmptyValidator) MarkdownDescription(_ context.Context) string {
	return "value must not be empty"
}

func (v stringNotEmptyValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if req.ConfigValue.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Empty value",
			fmt.Sprintf("Attribute %s must not be an empty string.", req.Path),
		)
	}
}

// ── stringOneOfCaseInsensitiveValidator ──────────────────────────────────────

// stringOneOfCaseInsensitiveValidator rejects values that are not in the
// allowed set (case-insensitive), equivalent to SDKv2
// validation.StringInSlice(allowed, true).
type stringOneOfCaseInsensitiveValidator struct {
	allowed []string
}

func stringOneOfCaseInsensitive(allowed ...string) validator.String {
	return stringOneOfCaseInsensitiveValidator{allowed: allowed}
}

func (v stringOneOfCaseInsensitiveValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must be one of: %s (case-insensitive)", strings.Join(v.allowed, ", "))
}

func (v stringOneOfCaseInsensitiveValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value must be one of: `%s` (case-insensitive)", strings.Join(v.allowed, "`, `"))
}

func (v stringOneOfCaseInsensitiveValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	for _, a := range v.allowed {
		if strings.EqualFold(val, a) {
			return
		}
	}
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid value",
		fmt.Sprintf("Attribute %s must be one of [%s] (case-insensitive); got %q.",
			req.Path, strings.Join(v.allowed, ", "), val),
	)
}

// ── conflictingAttrsValidator ─────────────────────────────────────────────────

// conflictingAttrsValidator is a resource.ConfigValidator that rejects
// configurations where both the primary attribute (identified by attrPath) and
// any of the conflicting attribute paths are set to non-null, non-empty values.
// This mirrors SDKv2 ConflictsWith behaviour.
type conflictingAttrsValidator struct {
	// attrPath is the path expression of the attribute that owns the conflict.
	attrPath path.Expression
	// conflicts lists attribute paths that must not be set when attrPath is set.
	conflicts []path.Expression
	// description is used in Description()/MarkdownDescription().
	description string
}

// conflictingAttrs builds a resource.ConfigValidator.
//   - primary: the attribute that triggers the conflict check.
//   - conflicts: the attributes that must not be set concurrently with primary.
func conflictingAttrs(primary string, conflicts ...string) resource.ConfigValidator {
	exprs := make([]path.Expression, len(conflicts))
	for i, c := range conflicts {
		exprs[i] = path.MatchRoot(c)
	}
	conflictNames := make([]string, len(conflicts))
	copy(conflictNames, conflicts)
	return conflictingAttrsValidator{
		attrPath:    path.MatchRoot(primary),
		conflicts:   exprs,
		description: fmt.Sprintf("%s conflicts with %s", primary, strings.Join(conflictNames, ", ")),
	}
}

func (v conflictingAttrsValidator) Description(_ context.Context) string { return v.description }
func (v conflictingAttrsValidator) MarkdownDescription(_ context.Context) string {
	return v.description
}

func (v conflictingAttrsValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Check whether the primary attribute is set to a non-null, non-empty value.
	primarySet, _ := attrIsSet(ctx, req.Config, v.attrPath)
	if !primarySet {
		return
	}

	// For each conflicting attribute, error if it is also set.
	for _, conflictExpr := range v.conflicts {
		conflictSet, attrPathStr := attrIsSet(ctx, req.Config, conflictExpr)
		if conflictSet {
			resp.Diagnostics.AddError(
				"Conflicting configuration",
				fmt.Sprintf("%s — these attributes cannot be set together.", v.description+" and "+attrPathStr),
			)
		}
	}
}

// attrIsSet returns (true, pathStr) when the attribute at expr is non-null and
// (when a string) non-empty. It returns (false, "") otherwise.
func attrIsSet(ctx context.Context, config tfsdk.Config, expr path.Expression) (bool, string) {
	paths, diags := config.PathMatches(ctx, expr)
	if diags.HasError() {
		return false, ""
	}
	for _, p := range paths {
		var val types.String
		d := diag.Diagnostics{}
		config.GetAttribute(ctx, p, &val) //nolint: errcheck — we check diags separately
		_ = d
		if val.IsNull() || val.IsUnknown() {
			continue
		}
		if val.ValueString() == "" {
			continue
		}
		return true, p.String()
	}
	return false, ""
}
