//go:build (all || resource_feature_flag) && !exclude_resource_feature_flag

package featuremanagement

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	featuremanagementapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/featuremanagement"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// ── Schema / Metadata smoke-tests ─────────────────────────────────────────────

// TestFeatureFlagSchemaHasRequiredFields verifies that the schema declares
// feature_id, scope_name, scope_value, and state as attributes (required or
// optional) and overridden / reason as computed attributes.
func TestFeatureFlagSchemaHasRequiredFields(t *testing.T) {
	ctx := context.Background()
	r := NewFeatureFlagResource()
	require.NotNil(t, r)

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not emit error diagnostics: %v", schemaResp.Diagnostics)

	attrs := schemaResp.Schema.Attributes

	// Required attributes
	for _, name := range []string{"feature_id", "scope_name", "scope_value", "state"} {
		_, ok := attrs[name]
		assert.True(t, ok, "schema must contain attribute %q", name)
	}

	// Computed attributes
	for _, name := range []string{"overridden", "reason"} {
		_, ok := attrs[name]
		assert.True(t, ok, "schema must contain computed attribute %q", name)
	}
}

// TestFeatureFlagMetadata verifies the resource type name.
func TestFeatureFlagMetadata(t *testing.T) {
	ctx := context.Background()
	r := NewFeatureFlagResource()
	require.NotNil(t, r)

	metaResp := &resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{}, metaResp)
	assert.Equal(t, "betterado_feature_flag", metaResp.TypeName)
}

// ── CRUD unit tests using gomock (gate: -run TestFeatureFlagCRUD) ─────────────

// testFeatureFlagModel returns a populated featureFlagModel for use in CRUD tests.
func testFeatureFlagModel() featureFlagModel {
	return featureFlagModel{
		FeatureID:  types.StringValue("ms.vss-work.agile"),
		ScopeName:  types.StringValue("project"),
		ScopeValue: types.StringValue("00000000-0000-0000-0000-000000000001"),
		State:      types.StringValue("enabled"),
		Overridden: types.BoolValue(false),
		Reason:     types.StringValue(""),
	}
}

// make404Err returns a WrappedError that utils.ResponseWasNotFound treats as 404.
func make404Err() error {
	code := 404
	return azuredevops.WrappedError{StatusCode: &code}
}

// TestFeatureFlagCRUDCreate verifies that setFeatureFlag calls
// SetFeatureStateForScope with the correct arguments and that readFeatureFlag
// stores computed fields (overridden, reason) from the read-back response.
func TestFeatureFlagCRUDCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFC := NewMockClient(ctrl)
	ctx := context.Background()
	model := testFeatureFlagModel()

	enabledState := featuremanagementapi.ContributedFeatureEnabledValueValues.Enabled
	overridden := false
	reason := "test-reason"

	// Expect SetFeatureStateForScope called with enabled state and correct args
	mockFC.EXPECT().
		SetFeatureStateForScope(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, args featuremanagementapi.SetFeatureStateForScopeArgs) (*featuremanagementapi.ContributedFeatureState, error) {
			require.NotNil(t, args.Feature, "Feature must not be nil")
			assert.Equal(t, "ms.vss-work.agile", *args.FeatureId)
			assert.Equal(t, "project", *args.UserScope)
			assert.Equal(t, "project", *args.ScopeName)
			assert.Equal(t, "00000000-0000-0000-0000-000000000001", *args.ScopeValue)
			assert.Equal(t, featuremanagementapi.ContributedFeatureEnabledValueValues.Enabled, *args.Feature.State)
			return &featuremanagementapi.ContributedFeatureState{State: &enabledState}, nil
		})

	// Expect GetFeatureStateForScope to read back state and computed fields
	mockFC.EXPECT().
		GetFeatureStateForScope(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, args featuremanagementapi.GetFeatureStateForScopeArgs) (*featuremanagementapi.ContributedFeatureState, error) {
			assert.Equal(t, "ms.vss-work.agile", *args.FeatureId)
			assert.Equal(t, "project", *args.UserScope)
			assert.Equal(t, "project", *args.ScopeName)
			assert.Equal(t, "00000000-0000-0000-0000-000000000001", *args.ScopeValue)
			return &featuremanagementapi.ContributedFeatureState{
				State:      &enabledState,
				Overridden: &overridden,
				Reason:     converter.String(reason),
			}, nil
		})

	// Exercise the same helpers the Create CRUD method calls
	setDiags := setFeatureFlag(ctx, mockFC, &model)
	require.False(t, setDiags.HasError(), "setFeatureFlag must not error: %s", setDiags)

	readDiags := readFeatureFlag(ctx, mockFC, &model)
	require.False(t, readDiags.HasError(), "readFeatureFlag must not error: %s", readDiags)

	assert.Equal(t, "enabled", model.State.ValueString())
	assert.False(t, model.Overridden.ValueBool())
	assert.Equal(t, "test-reason", model.Reason.ValueString())
}

// TestFeatureFlagCRUDRead verifies that readFeatureFlag calls
// GetFeatureStateForScope and populates the model; also checks that a 404
// response results in a "NotFound" diagnostic (Read CRUD converts to
// RemoveResource).
func TestFeatureFlagCRUDRead(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFC := NewMockClient(ctrl)
		ctx := context.Background()
		model := testFeatureFlagModel()

		disabledState := featuremanagementapi.ContributedFeatureEnabledValueValues.Disabled
		overridden := true
		pol := "policy override"

		mockFC.EXPECT().
			GetFeatureStateForScope(ctx, gomock.Any()).
			Return(&featuremanagementapi.ContributedFeatureState{
				State:      &disabledState,
				Overridden: &overridden,
				Reason:     &pol,
			}, nil)

		diags := readFeatureFlag(ctx, mockFC, &model)
		require.False(t, diags.HasError(), "readFeatureFlag must not error: %s", diags)

		assert.Equal(t, "disabled", model.State.ValueString())
		assert.True(t, model.Overridden.ValueBool())
		assert.Equal(t, "policy override", model.Reason.ValueString())
	})

	t.Run("404_returns_NotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFC := NewMockClient(ctrl)
		ctx := context.Background()
		model := testFeatureFlagModel()

		mockFC.EXPECT().
			GetFeatureStateForScope(ctx, gomock.Any()).
			Return(nil, make404Err())

		diags := readFeatureFlag(ctx, mockFC, &model)
		require.True(t, diags.HasError(), "must return error diag on 404")
		assert.True(t, hasNotFoundDiag(diags), "must be a NotFound diagnostic")
	})

	t.Run("undefined_state_returns_NotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFC := NewMockClient(ctrl)
		ctx := context.Background()
		model := testFeatureFlagModel()

		undefinedState := featuremanagementapi.ContributedFeatureEnabledValueValues.Undefined
		mockFC.EXPECT().
			GetFeatureStateForScope(ctx, gomock.Any()).
			Return(&featuremanagementapi.ContributedFeatureState{State: &undefinedState}, nil)

		diags := readFeatureFlag(ctx, mockFC, &model)
		require.True(t, diags.HasError(), "undefined state must be treated as not-found")
		assert.True(t, hasNotFoundDiag(diags), "must be a NotFound diagnostic")
	})
}

// TestFeatureFlagCRUDUpdate verifies that the Update helpers call
// SetFeatureStateForScope with the new desired state, then read back the result.
func TestFeatureFlagCRUDUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFC := NewMockClient(ctrl)
	ctx := context.Background()
	model := testFeatureFlagModel()
	model.State = types.StringValue("disabled") // new desired state

	disabledState := featuremanagementapi.ContributedFeatureEnabledValueValues.Disabled
	overridden := false

	// SetFeatureStateForScope must be called with disabled state
	mockFC.EXPECT().
		SetFeatureStateForScope(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, args featuremanagementapi.SetFeatureStateForScopeArgs) (*featuremanagementapi.ContributedFeatureState, error) {
			require.NotNil(t, args.Feature)
			assert.Equal(t, featuremanagementapi.ContributedFeatureEnabledValueValues.Disabled, *args.Feature.State)
			assert.Equal(t, "ms.vss-work.agile", *args.FeatureId)
			return &featuremanagementapi.ContributedFeatureState{State: &disabledState}, nil
		})

	// GetFeatureStateForScope reads back the new state
	mockFC.EXPECT().
		GetFeatureStateForScope(ctx, gomock.Any()).
		Return(&featuremanagementapi.ContributedFeatureState{
			State:      &disabledState,
			Overridden: &overridden,
		}, nil)

	setDiags := setFeatureFlag(ctx, mockFC, &model)
	require.False(t, setDiags.HasError(), "setFeatureFlag must not error: %s", setDiags)

	readDiags := readFeatureFlag(ctx, mockFC, &model)
	require.False(t, readDiags.HasError(), "readFeatureFlag must not error: %s", readDiags)

	assert.Equal(t, "disabled", model.State.ValueString())
	assert.False(t, model.Overridden.ValueBool())
}

// TestFeatureFlagCRUDDelete verifies that deleteFeatureFlag calls
// SetFeatureStateForScope with state "undefined" to restore the feature to its
// default (removes Terraform management without disabling the feature).
func TestFeatureFlagCRUDDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFC := NewMockClient(ctrl)
	ctx := context.Background()
	model := testFeatureFlagModel()

	undefinedState := featuremanagementapi.ContributedFeatureEnabledValueValues.Undefined

	mockFC.EXPECT().
		SetFeatureStateForScope(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, args featuremanagementapi.SetFeatureStateForScopeArgs) (*featuremanagementapi.ContributedFeatureState, error) {
			require.NotNil(t, args.Feature)
			assert.Equal(t, featuremanagementapi.ContributedFeatureEnabledValueValues.Undefined, *args.Feature.State)
			assert.Equal(t, "ms.vss-work.agile", *args.FeatureId)
			assert.Equal(t, "project", *args.UserScope)
			assert.Equal(t, "project", *args.ScopeName)
			assert.Equal(t, "00000000-0000-0000-0000-000000000001", *args.ScopeValue)
			return &featuremanagementapi.ContributedFeatureState{State: &undefinedState}, nil
		})

	diags := deleteFeatureFlag(ctx, mockFC, &model)
	require.False(t, diags.HasError(), "deleteFeatureFlag must not error: %s", diags)
}
