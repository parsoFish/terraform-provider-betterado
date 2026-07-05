//go:build all || resource_pipeline_authorization_framework

package build

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPipelineAuthorizationFramework_Schema calls Schema() on the framework resource and
// asserts that the schema declares all required attributes with no error diagnostics.
//
// This is the quality-gate test; the gate command is:
//
//	go test -tags all -run TestPipelineAuthorizationFramework_Schema ./azuredevops/internal/service/build/...
func TestPipelineAuthorizationFramework_Schema(t *testing.T) {
	t.Parallel()

	r := &PipelineAuthorizationResource{}

	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}

	r.Schema(context.Background(), schemaReq, schemaResp)

	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not return error diagnostics: %s", schemaResp.Diagnostics)

	attrs := schemaResp.Schema.Attributes

	// AC1: pipeline_authorization schema declares project_id, pipeline_project_id,
	// resource_id, type, pipeline_id.
	requiredAttrs := []string{
		"project_id",
		"pipeline_project_id",
		"resource_id",
		"type",
		"pipeline_id",
	}
	for _, attr := range requiredAttrs {
		_, ok := attrs[attr]
		assert.True(t, ok, "schema must declare %q attribute", attr)
	}
}
