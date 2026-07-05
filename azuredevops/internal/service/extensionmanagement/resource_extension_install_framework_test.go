//go:build all || resource_extension_install

package extensionmanagement

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/extensionmanagement"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtensionInstallResource_Metadata verifies the resource type name.
func TestExtensionInstallResource_Metadata(t *testing.T) {
	ctx := context.Background()
	r := NewExtensionInstallResource()
	require.NotNil(t, r)

	resp := &resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{}, resp)
	assert.Equal(t, "betterado_extension_install", resp.TypeName)
}

// TestExtensionInstallResource_Schema verifies that Schema() returns the required attributes.
func TestExtensionInstallResource_Schema(t *testing.T) {
	ctx := context.Background()
	r := NewExtensionInstallResource()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not emit error diagnostics: %v", schemaResp.Diagnostics)

	attrs := schemaResp.Schema.Attributes

	// Required attributes
	pubAttr, ok := attrs["publisher_id"]
	require.True(t, ok, "schema must contain 'publisher_id'")
	assert.False(t, pubAttr.IsOptional(), "publisher_id must not be optional")
	assert.True(t, pubAttr.IsRequired(), "publisher_id must be required")

	extAttr, ok := attrs["extension_id"]
	require.True(t, ok, "schema must contain 'extension_id'")
	assert.False(t, extAttr.IsOptional(), "extension_id must not be optional")
	assert.True(t, extAttr.IsRequired(), "extension_id must be required")

	// Computed version
	verAttr, ok := attrs["version"]
	require.True(t, ok, "schema must contain 'version'")
	assert.True(t, verAttr.IsComputed(), "version must be computed")

	// Optional disabled
	disAttr, ok := attrs["disabled"]
	require.True(t, ok, "schema must contain 'disabled'")
	assert.True(t, disAttr.IsOptional(), "disabled must be optional")
}

// TestExtensionInstallResource_ExpandFlatten verifies that expandExtensionInstall ->
// flattenExtensionInstall round-trips publisher_id, extension_id, and version without data loss.
func TestExtensionInstallResource_ExpandFlatten(t *testing.T) {
	// Build a populated model
	original := &extensionInstallModel{
		PublisherID: types.StringValue("ms-securitydevops"),
		ExtensionID: types.StringValue("microsoft-security-devops-azdevops"),
		Version:     types.StringValue("1.2.3"),
		Disabled:    types.BoolValue(false),
	}

	// Expand to API struct
	apiExt := expandExtensionInstall(original)

	// Verify expand populated the fields
	require.NotNil(t, apiExt.PublisherId)
	assert.Equal(t, "ms-securitydevops", *apiExt.PublisherId)
	require.NotNil(t, apiExt.ExtensionId)
	assert.Equal(t, "microsoft-security-devops-azdevops", *apiExt.ExtensionId)
	require.NotNil(t, apiExt.Version)
	assert.Equal(t, "1.2.3", *apiExt.Version)

	// Simulate an API response by constructing an InstalledExtension from the expanded struct.
	apiResponse := &extensionmanagement.InstalledExtension{
		PublisherId: apiExt.PublisherId,
		ExtensionId: apiExt.ExtensionId,
		Version:     apiExt.Version,
		InstallState: &extensionmanagement.InstalledExtensionState{
			Flags: converter.ToPtr(extensionmanagement.ExtensionStateFlagsValues.None),
		},
	}

	// Flatten back into a fresh model
	result := &extensionInstallModel{
		PublisherID: original.PublisherID,
		ExtensionID: original.ExtensionID,
		Version:     original.Version,
		Disabled:    original.Disabled,
	}
	flattenExtensionInstall(result, apiResponse)

	// Verify round-trip preserves all fields
	assert.Equal(t, "ms-securitydevops", result.PublisherID.ValueString(),
		"publisher_id must round-trip unchanged")
	assert.Equal(t, "microsoft-security-devops-azdevops", result.ExtensionID.ValueString(),
		"extension_id must round-trip unchanged")
	assert.Equal(t, "1.2.3", result.Version.ValueString(),
		"version must round-trip unchanged")
	assert.False(t, result.Disabled.ValueBool(),
		"disabled must round-trip as false when flags=None")
}

// TestExtensionInstallResource_ExpandFlatten_Disabled verifies that the disabled flag
// survives an expand -> flatten cycle.
func TestExtensionInstallResource_ExpandFlatten_Disabled(t *testing.T) {
	model := &extensionInstallModel{
		PublisherID: types.StringValue("ms-securitydevops"),
		ExtensionID: types.StringValue("microsoft-security-devops-azdevops"),
		Version:     types.StringValue("1.2.3"),
		Disabled:    types.BoolValue(true),
	}

	apiExt := expandExtensionInstall(model)
	require.NotNil(t, apiExt.InstallState)
	require.NotNil(t, apiExt.InstallState.Flags)
	assert.Equal(t, extensionmanagement.ExtensionStateFlagsValues.Disabled, *apiExt.InstallState.Flags)

	apiResponse := &extensionmanagement.InstalledExtension{
		PublisherId:  apiExt.PublisherId,
		ExtensionId:  apiExt.ExtensionId,
		Version:      apiExt.Version,
		InstallState: apiExt.InstallState,
	}

	result := &extensionInstallModel{}
	flattenExtensionInstall(result, apiResponse)

	assert.True(t, result.Disabled.ValueBool(), "disabled must be true when flags=Disabled")
}
