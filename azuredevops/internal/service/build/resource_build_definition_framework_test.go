//go:build all || resource_build_definition_framework

package build

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildDefinitionFramework_Schema calls Schema() on the framework resource and
// asserts that the schema declares all required attributes with no error diagnostics.
//
// This is the quality-gate test; the gate command is:
//
//	go test -tags all -run TestBuildDefinitionFramework_Schema ./azuredevops/internal/service/build/...
func TestBuildDefinitionFramework_Schema(t *testing.T) {
	t.Parallel()

	r := &BuildDefinitionResource{}

	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}

	r.Schema(context.Background(), schemaReq, schemaResp)

	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not return error diagnostics: %s", schemaResp.Diagnostics)

	attrs := schemaResp.Schema.Attributes

	// Top-level required attributes per AC1.
	requiredAttrs := []string{
		"name",
		"project_id",
		"revision",
		"path",
		"agent_pool_name",
		"repository",
		"variable",
		"ci_trigger",
		"pull_request_trigger",
		"agent_specification",
		"job_authorization_scope",
		"queue_status",
		"skip_first_run",
	}
	for _, attr := range requiredAttrs {
		_, ok := attrs[attr]
		assert.True(t, ok, "schema must declare %q attribute", attr)
	}

	// Spot-check that the nested attributes exist (as schema.Attribute values).
	for _, nestedAttrName := range []string{"repository", "variable", "ci_trigger", "pull_request_trigger"} {
		a, exists := attrs[nestedAttrName]
		assert.True(t, exists, "schema must declare %q attribute", nestedAttrName)
		assert.NotNil(t, a, "schema attribute %q must not be nil", nestedAttrName)
	}
}
