//go:build (all || resource_feature_flag) && !exclude_resource_feature_flag

package featuremanagement

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
