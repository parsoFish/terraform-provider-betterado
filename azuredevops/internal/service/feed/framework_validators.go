package feed

// framework_validators.go — inline implementations of terraform-plugin-framework
// schema validators used by the feed resource, permission, retention-policy and
// data-source. These replicate the subset of
// terraform-plugin-framework-validators that was present in the SDKv2
// implementations, so plan-time validation parity is restored without adding a
// new module dependency.

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// ── stringNotWhiteSpace ───────────────────────────────────────────────────────

// stringNotWhiteSpace rejects values that are empty or consist entirely of
// whitespace, matching the SDKv2 StringIsNotWhiteSpace helper.
type stringNotWhiteSpaceValidator struct{}

func stringNotWhiteSpace() validator.String { return stringNotWhiteSpaceValidator{} }

func (v stringNotWhiteSpaceValidator) Description(_ context.Context) string {
	return "value must not be empty or whitespace"
}

func (v stringNotWhiteSpaceValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringNotWhiteSpaceValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	if strings.TrimFunc(val, unicode.IsSpace) == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid value",
			fmt.Sprintf("Attribute %s must not be empty or whitespace, got: %q", req.Path, val),
		)
	}
}

// ── stringIsUUID ──────────────────────────────────────────────────────────────

var uuidRegexp = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// stringIsUUID rejects values that are not a valid UUID, matching the SDKv2
// IsUUID helper.
type stringIsUUIDValidator struct{}

func stringIsUUID() validator.String { return stringIsUUIDValidator{} }

func (v stringIsUUIDValidator) Description(_ context.Context) string {
	return "value must be a valid UUID"
}

func (v stringIsUUIDValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsUUIDValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	if !uuidRegexp.MatchString(val) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid UUID",
			fmt.Sprintf("Attribute %s must be a valid UUID, got: %q", req.Path, val),
		)
	}
}

// ── stringOneOf ───────────────────────────────────────────────────────────────

// stringOneOf rejects values not in the provided set, matching the SDKv2
// StringInSlice helper.
type stringOneOfValidator struct {
	valid []string
}

func stringOneOf(values ...string) validator.String { return stringOneOfValidator{valid: values} }

func (v stringOneOfValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must be one of: %s", strings.Join(v.valid, ", "))
}

func (v stringOneOfValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringOneOfValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	for _, allowed := range v.valid {
		if val == allowed {
			return
		}
	}
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid value",
		fmt.Sprintf("Attribute %s must be one of [%s], got: %q", req.Path, strings.Join(v.valid, ", "), val),
	)
}

// ── int64Between ──────────────────────────────────────────────────────────────

// int64Between rejects values outside [min, max], matching the SDKv2
// IntBetween helper.
type int64BetweenValidator struct {
	min, max int64
}

func int64Between(min, max int64) validator.Int64 { return int64BetweenValidator{min: min, max: max} }

func (v int64BetweenValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must be between %d and %d (inclusive)", v.min, v.max)
}

func (v int64BetweenValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v int64BetweenValidator) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueInt64()
	if val < v.min || val > v.max {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Value out of range",
			fmt.Sprintf("Attribute %s must be between %d and %d (inclusive), got: %d", req.Path, v.min, v.max, val),
		)
	}
}
