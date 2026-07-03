package feed

// framework_validators.go — wrappers around terraform-plugin-framework-validators
// and minimal inline validators for the feed resource, permission,
// retention-policy and data-source schemas.

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// ── stringNotWhiteSpace ───────────────────────────────────────────────────────

// stringNotWhiteSpace rejects values that are empty or consist entirely of
// whitespace, matching the SDKv2 StringIsNotWhiteSpace helper.
// There is no equivalent in terraform-plugin-framework-validators, so we
// keep this as a minimal inline implementation.
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

// stringIsUUID returns a validator.String that rejects non-UUID values.
// Uses stringvalidator.RegexMatches from terraform-plugin-framework-validators.
func stringIsUUID() validator.String {
	return stringvalidator.RegexMatches(
		uuidRegexp,
		"value must be a valid UUID (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)",
	)
}

// ── stringOneOf ───────────────────────────────────────────────────────────────

// stringOneOf returns a validator.String that rejects values not in the set.
// Delegates to stringvalidator.OneOf from terraform-plugin-framework-validators.
func stringOneOf(values ...string) validator.String {
	return stringvalidator.OneOf(values...)
}

// ── int64Between ──────────────────────────────────────────────────────────────

// int64Between returns a validator.Int64 that rejects values outside [min, max].
// Delegates to int64validator.Between from terraform-plugin-framework-validators.
func int64Between(min, max int64) validator.Int64 {
	return int64validator.Between(min, max)
}
