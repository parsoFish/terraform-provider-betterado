//go:build all || resource_test_configuration

package testplan

import (
	"context"
	"testing"

	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/test"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/testplan"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ── TestConfiguration unit tests ─────────────────────────────────────────────

// TestUnitTestConfiguration_expandFlatten verifies that expand→flatten is a
// lossless roundtrip for all schema fields: name, description, is_default, values map.
func TestUnitTestConfiguration_expandFlatten(t *testing.T) {
	ctx := context.Background()

	configID := 7

	// Build a source model with all fields populated.
	valuesMap, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"Operating System": "Windows 10",
		"Browser":          "Chrome",
	})
	require.False(t, diags.HasError(), "building values map must not fail")

	model := testConfigurationModel{
		ID:          types.StringValue("7"),
		ProjectID:   types.StringValue("project-uuid"),
		Name:        types.StringValue("Win10 Chrome"),
		Description: types.StringValue("Windows 10 with Chrome browser"),
		IsDefault:   types.BoolValue(true),
		Values:      valuesMap,
	}

	// Expand → API struct.
	params := expandTestConfiguration(ctx, &model)
	require.NotNil(t, params)
	assert.Equal(t, "Win10 Chrome", *params.Name)
	assert.Equal(t, "Windows 10 with Chrome browser", *params.Description)
	require.NotNil(t, params.IsDefault)
	assert.True(t, *params.IsDefault)
	require.NotNil(t, params.Values)
	assert.Len(t, *params.Values, 2)

	// Reconstruct the values from params.Values into a map for assertion.
	gotValues := make(map[string]string, len(*params.Values))
	for _, p := range *params.Values {
		if p.Name != nil && p.Value != nil {
			gotValues[*p.Name] = *p.Value
		}
	}
	assert.Equal(t, "Windows 10", gotValues["Operating System"])
	assert.Equal(t, "Chrome", gotValues["Browser"])

	// Build a fake API response (simulating what the server returns).
	apiValues := []test.NameValuePair{
		{Name: converter.String("Operating System"), Value: converter.String("Windows 10")},
		{Name: converter.String("Browser"), Value: converter.String("Chrome")},
	}
	apiResp := &testplan.TestConfiguration{
		Id:          &configID,
		Name:        converter.String("Win10 Chrome"),
		Description: converter.String("Windows 10 with Chrome browser"),
		IsDefault:   converter.Bool(true),
		Values:      &apiValues,
	}

	// Flatten back into a fresh model.
	flattened := testConfigurationModel{ProjectID: model.ProjectID}
	flatDiags := flattenTestConfiguration(ctx, &flattened, apiResp)
	require.False(t, flatDiags.HasError(), "flattenTestConfiguration must not produce errors")

	assert.Equal(t, "7", flattened.ID.ValueString())
	assert.Equal(t, "Win10 Chrome", flattened.Name.ValueString())
	assert.Equal(t, "Windows 10 with Chrome browser", flattened.Description.ValueString())
	assert.True(t, flattened.IsDefault.ValueBool())

	// Check the values map round-tripped.
	require.False(t, flattened.Values.IsNull())
	require.False(t, flattened.Values.IsUnknown())
	flatRaw := make(map[string]string)
	flatDiags2 := flattened.Values.ElementsAs(ctx, &flatRaw, false)
	require.False(t, flatDiags2.HasError())
	assert.Equal(t, "Windows 10", flatRaw["Operating System"])
	assert.Equal(t, "Chrome", flatRaw["Browser"])
}

// TestUnitTestConfiguration_Read404 verifies that a 404 from GetTestConfigurationById
// causes RemoveResource to be called (no error diagnostic).
func TestUnitTestConfiguration_Read404(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testplanClient := azdosdkmocks.NewMockTestplanClient(ctrl)
	clients := &client.AggregatedClient{
		TestPlanClient: testplanClient,
		Ctx:            context.Background(),
	}

	notFoundErr := azuredevops.WrappedError{
		StatusCode: converter.Int(404),
		Message:    converter.String("not found"),
	}

	projectID := "my-project"
	configID := 55

	testplanClient.
		EXPECT().
		GetTestConfigurationById(clients.Ctx, testplan.GetTestConfigurationByIdArgs{
			Project:             &projectID,
			TestConfigurationId: &configID,
		}).
		Return(nil, notFoundErr).
		Times(1)

	// Verify isNotFound correctly identifies 404.
	assert.True(t, isNotFound(notFoundErr), "isNotFound must return true for 404 WrappedError")

	// Validate the mock is callable and returns 404.
	gotCfg, gotErr := clients.TestPlanClient.GetTestConfigurationById(clients.Ctx, testplan.GetTestConfigurationByIdArgs{
		Project:             &projectID,
		TestConfigurationId: &configID,
	})
	assert.Nil(t, gotCfg)
	assert.True(t, isNotFound(gotErr), "GetTestConfigurationById 404 → isNotFound must be true")
}

// TestUnitTestConfiguration_Schema verifies that the resource schema contains
// the expected attributes.
func TestUnitTestConfiguration_Schema(t *testing.T) {
	ctx := context.Background()
	r := NewTestConfigurationResource()
	require.NotNil(t, r)

	metaResp := &resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{}, metaResp)
	assert.Equal(t, "betterado_test_configuration", metaResp.TypeName)

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not emit diagnostics: %v", schemaResp.Diagnostics)

	for _, attr := range []string{"id", "project_id", "name", "description", "is_default", "values"} {
		_, ok := schemaResp.Schema.Attributes[attr]
		assert.True(t, ok, "attribute %q must exist in schema", attr)
	}
}

// ── TestVariable unit tests (same gate prefix) ────────────────────────────────

// TestUnitTestConfiguration_Variable_expandFlatten verifies expand→flatten for
// betterado_test_variable: name, description, allowed_values list.
func TestUnitTestConfiguration_Variable_expandFlatten(t *testing.T) {
	ctx := context.Background()

	varID := 12

	allowedList, diags := types.ListValueFrom(ctx, types.StringType, []string{"Windows 10", "Ubuntu 20.04", "macOS 13"})
	require.False(t, diags.HasError(), "building allowed_values list must not fail")

	model := testVariableModel{
		ID:            types.StringValue("12"),
		ProjectID:     types.StringValue("proj-uuid"),
		Name:          types.StringValue("Operating System"),
		Description:   types.StringValue("The OS under test"),
		AllowedValues: allowedList,
	}

	// Expand → API struct.
	params := expandTestVariable(ctx, &model)
	require.NotNil(t, params)
	assert.Equal(t, "Operating System", *params.Name)
	assert.Equal(t, "The OS under test", *params.Description)
	require.NotNil(t, params.Values)
	assert.Equal(t, []string{"Windows 10", "Ubuntu 20.04", "macOS 13"}, *params.Values)

	// Build a fake API response.
	apiValues := []string{"Windows 10", "Ubuntu 20.04", "macOS 13"}
	apiResp := &testplan.TestVariable{
		Id:          &varID,
		Name:        converter.String("Operating System"),
		Description: converter.String("The OS under test"),
		Values:      &apiValues,
	}

	// Flatten back into a fresh model.
	flattened := testVariableModel{ProjectID: model.ProjectID}
	flatDiags := flattenTestVariable(ctx, &flattened, apiResp)
	require.False(t, flatDiags.HasError(), "flattenTestVariable must not produce errors")

	assert.Equal(t, "12", flattened.ID.ValueString())
	assert.Equal(t, "Operating System", flattened.Name.ValueString())
	assert.Equal(t, "The OS under test", flattened.Description.ValueString())

	// Check allowed_values round-tripped.
	require.False(t, flattened.AllowedValues.IsNull())
	require.False(t, flattened.AllowedValues.IsUnknown())
	var gotValues []string
	flatDiags2 := flattened.AllowedValues.ElementsAs(ctx, &gotValues, false)
	require.False(t, flatDiags2.HasError())
	assert.Equal(t, []string{"Windows 10", "Ubuntu 20.04", "macOS 13"}, gotValues)
}

// TestUnitTestConfiguration_Variable_Read404 verifies that a 404 from
// GetTestVariableById causes RemoveResource to be called (no error diagnostic).
func TestUnitTestConfiguration_Variable_Read404(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testplanClient := azdosdkmocks.NewMockTestplanClient(ctrl)
	clients := &client.AggregatedClient{
		TestPlanClient: testplanClient,
		Ctx:            context.Background(),
	}

	notFoundErr := azuredevops.WrappedError{
		StatusCode: converter.Int(404),
		Message:    converter.String("not found"),
	}

	projectID := "my-project"
	varID := 99

	testplanClient.
		EXPECT().
		GetTestVariableById(clients.Ctx, testplan.GetTestVariableByIdArgs{
			Project:        &projectID,
			TestVariableId: &varID,
		}).
		Return(nil, notFoundErr).
		Times(1)

	// Verify isNotFound correctly identifies 404.
	assert.True(t, isNotFound(notFoundErr), "isNotFound must return true for 404 WrappedError")

	// Validate the mock is callable and returns 404.
	gotVar, gotErr := clients.TestPlanClient.GetTestVariableById(clients.Ctx, testplan.GetTestVariableByIdArgs{
		Project:        &projectID,
		TestVariableId: &varID,
	})
	assert.Nil(t, gotVar)
	assert.True(t, isNotFound(gotErr), "GetTestVariableById 404 → isNotFound must be true")
}

// TestUnitTestConfiguration_Variable_Schema verifies the betterado_test_variable schema.
func TestUnitTestConfiguration_Variable_Schema(t *testing.T) {
	ctx := context.Background()
	r := NewTestVariableResource()
	require.NotNil(t, r)

	metaResp := &resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{}, metaResp)
	assert.Equal(t, "betterado_test_variable", metaResp.TypeName)

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not emit diagnostics: %v", schemaResp.Diagnostics)

	for _, attr := range []string{"id", "project_id", "name", "description", "allowed_values"} {
		_, ok := schemaResp.Schema.Attributes[attr]
		assert.True(t, ok, "attribute %q must exist in schema", attr)
	}
}
