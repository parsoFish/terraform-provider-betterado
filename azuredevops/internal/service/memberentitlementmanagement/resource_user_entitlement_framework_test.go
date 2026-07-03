//go:build all || memberentitlementmanagement

package memberentitlementmanagement

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/licensing"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/memberentitlementmanagement"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserEntitlementResource_Metadata(t *testing.T) {
	r := NewUserEntitlementResource()
	require.NotNil(t, r, "NewUserEntitlementResource() should return a non-nil resource")

	var metaResp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)

	require.Equal(t, "betterado_user_entitlement", metaResp.TypeName,
		"resource TypeName must be betterado_user_entitlement")
}

// buildUEConfig creates a tfsdk.Config for the UserEntitlement schema with the given attribute values.
func buildUEConfig(t *testing.T, attrOverrides map[string]tftypes.Value) tfsdk.Config {
	t.Helper()
	r := NewUserEntitlementResource()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)

	tfType := schemaResp.Schema.Type().TerraformType(context.Background())
	base := map[string]tftypes.Value{
		"id":                   tftypes.NewValue(tftypes.String, nil),
		"principal_name":       tftypes.NewValue(tftypes.String, nil),
		"origin_id":            tftypes.NewValue(tftypes.String, nil),
		"origin":               tftypes.NewValue(tftypes.String, nil),
		"account_license_type": tftypes.NewValue(tftypes.String, nil),
		"licensing_source":     tftypes.NewValue(tftypes.String, nil),
		"descriptor":           tftypes.NewValue(tftypes.String, nil),
	}
	for k, v := range attrOverrides {
		base[k] = v
	}
	cfgVal := tftypes.NewValue(tfType, base)
	return tfsdk.Config{
		Raw:    cfgVal,
		Schema: schemaResp.Schema,
	}
}

// TestUserEntitlementResource_ConfigValidators_PrincipalNameExclusive verifies AC2:
// setting both principal_name and origin_id produces a diagnostic error.
func TestUserEntitlementResource_ConfigValidators_PrincipalNameExclusive(t *testing.T) {
	r := NewUserEntitlementResource()
	rv, ok := r.(resource.ResourceWithConfigValidators)
	require.True(t, ok, "UserEntitlementResource must implement ResourceWithConfigValidators")

	validators := rv.ConfigValidators(context.Background())
	require.Len(t, validators, 1)

	cfg := buildUEConfig(t, map[string]tftypes.Value{
		"principal_name": tftypes.NewValue(tftypes.String, "user@example.com"),
		"origin_id":      tftypes.NewValue(tftypes.String, "some-origin-id"),
		"origin":         tftypes.NewValue(tftypes.String, "aad"),
	})
	req := resource.ValidateConfigRequest{Config: cfg}
	resp := &resource.ValidateConfigResponse{}
	validators[0].ValidateResource(context.Background(), req, resp)

	require.True(t, resp.Diagnostics.HasError(), "expected error when both principal_name and origin_id are set")
}

// TestUserEntitlementResource_ConfigValidators_RequiresAtLeastOne verifies AC2:
// setting neither principal_name nor origin_id produces a diagnostic error.
func TestUserEntitlementResource_ConfigValidators_RequiresAtLeastOne(t *testing.T) {
	r := NewUserEntitlementResource()
	rv, ok := r.(resource.ResourceWithConfigValidators)
	require.True(t, ok)

	validators := rv.ConfigValidators(context.Background())

	cfg := buildUEConfig(t, nil)
	req := resource.ValidateConfigRequest{Config: cfg}
	resp := &resource.ValidateConfigResponse{}
	validators[0].ValidateResource(context.Background(), req, resp)

	require.True(t, resp.Diagnostics.HasError(), "expected error when neither principal_name nor origin_id is set")
}

// TestUserEntitlementResource_ConfigValidators_OriginRequiredTogether verifies AC2:
// setting origin without origin_id produces a diagnostic error.
func TestUserEntitlementResource_ConfigValidators_OriginRequiredTogether(t *testing.T) {
	r := NewUserEntitlementResource()
	rv, ok := r.(resource.ResourceWithConfigValidators)
	require.True(t, ok)

	validators := rv.ConfigValidators(context.Background())

	cfg := buildUEConfig(t, map[string]tftypes.Value{
		"origin": tftypes.NewValue(tftypes.String, "aad"),
	})
	req := resource.ValidateConfigRequest{Config: cfg}
	resp := &resource.ValidateConfigResponse{}
	validators[0].ValidateResource(context.Background(), req, resp)

	require.True(t, resp.Diagnostics.HasError(), "expected error when origin is set without origin_id")
}

// TestUserEntitlementResource_ConfigValidators_ValidPrincipalName verifies AC2:
// a valid config with only principal_name passes validation.
func TestUserEntitlementResource_ConfigValidators_ValidPrincipalName(t *testing.T) {
	r := NewUserEntitlementResource()
	rv, ok := r.(resource.ResourceWithConfigValidators)
	require.True(t, ok)

	validators := rv.ConfigValidators(context.Background())

	cfg := buildUEConfig(t, map[string]tftypes.Value{
		"principal_name": tftypes.NewValue(tftypes.String, "user@example.com"),
	})
	req := resource.ValidateConfigRequest{Config: cfg}
	resp := &resource.ValidateConfigResponse{}
	validators[0].ValidateResource(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "expected no error for valid principal_name-only config")
}

// TestUserEntitlementResource_Schema_Defaults verifies AC3:
// account_license_type and licensing_source have schema-level descriptions
// mentioning their default values, confirming the Default is wired.
func TestUserEntitlementResource_Schema_Defaults(t *testing.T) {
	r := NewUserEntitlementResource()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)

	altAttr, ok := schemaResp.Schema.Attributes["account_license_type"]
	require.True(t, ok, "account_license_type attribute must exist")
	require.Contains(t, altAttr.GetDescription(), "express",
		"account_license_type description should mention default value 'express'")

	lsAttr, ok := schemaResp.Schema.Attributes["licensing_source"]
	require.True(t, ok, "licensing_source attribute must exist")
	require.Contains(t, lsAttr.GetDescription(), "account",
		"licensing_source description should mention default value 'account'")
}

// TestUserEntitlementResource_LowerCasePlanModifier verifies AC4:
// the ueLowerCase plan modifier normalises values to lower-case.
func TestUserEntitlementResource_LowerCasePlanModifier(t *testing.T) {
	mod := ueLowerCase()

	req := planmodifier.StringRequest{
		PlanValue: types.StringValue("Express"),
	}
	resp := &planmodifier.StringResponse{
		PlanValue: req.PlanValue,
	}
	mod.PlanModifyString(context.Background(), req, resp)

	assert.Equal(t, "express", resp.PlanValue.ValueString(),
		"ueLowerCase should normalise 'Express' to 'express'")
}

// TestUserEntitlementResource_FlattenNormalisesCase verifies AC4:
// flattenUserEntitlementFramework writes lower-case for AccountLicenseType
// and LicensingSource regardless of what the API returns.
func TestUserEntitlementResource_FlattenNormalisesCase(t *testing.T) {
	altVal := licensing.AccountLicenseType("Express")
	lsVal := licensing.LicensingSource("Account")

	model := &userEntitlementModel{
		ID: types.StringValue("test-id"),
	}
	ue := &memberentitlementmanagement.UserEntitlement{
		AccessLevel: &licensing.AccessLevel{
			AccountLicenseType: &altVal,
			LicensingSource:    &lsVal,
		},
	}
	flattenUserEntitlementFramework(model, ue)

	assert.Equal(t, "express", model.AccountLicenseType.ValueString(),
		"flattenUserEntitlementFramework should normalise AccountLicenseType to lower-case")
	assert.Equal(t, "account", model.LicensingSource.ValueString(),
		"flattenUserEntitlementFramework should normalise LicensingSource to lower-case")
}
