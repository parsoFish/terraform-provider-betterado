//go:build all || resource_pipeline_authorization_framework

package build

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResourceAuthorizationFramework_Schema calls Schema() on the framework resource and
// asserts that the schema declares all required attributes with no error diagnostics.
func TestResourceAuthorizationFramework_Schema(t *testing.T) {
	t.Parallel()

	r := &ResourceAuthorizationResource{}

	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}

	r.Schema(context.Background(), schemaReq, schemaResp)

	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not return error diagnostics: %s", schemaResp.Diagnostics)

	attrs := schemaResp.Schema.Attributes

	// AC1: resource_authorization schema declares project_id, resource_id,
	// definition_id, type, authorized.
	requiredAttrs := []string{
		"project_id",
		"resource_id",
		"definition_id",
		"type",
		"authorized",
	}
	for _, attr := range requiredAttrs {
		_, ok := attrs[attr]
		assert.True(t, ok, "schema must declare %q attribute", attr)
	}
}
