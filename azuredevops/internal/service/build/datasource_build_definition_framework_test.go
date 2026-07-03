//go:build all || datasource_build_definition_framework

package build

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildDefinitionDataSourceFramework_Schema calls Schema() on the framework data source and
// asserts that the schema declares all required attributes with no error diagnostics.
//
// This is the quality-gate test; the gate command is:
//
//	go test -tags all -run TestBuildDefinitionDataSourceFramework_Schema ./azuredevops/internal/service/build/...
func TestBuildDefinitionDataSourceFramework_Schema(t *testing.T) {
	t.Parallel()

	d := &BuildDefinitionDataSource{}

	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), schemaReq, schemaResp)

	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not return error diagnostics: %s", schemaResp.Diagnostics)

	attrs := schemaResp.Schema.Attributes

	// AC1: project_id (Required), name (Required), path (Optional), revision (Computed),
	// repository (Computed), ci_trigger (Computed), pull_request_trigger (Computed),
	// variable (Computed), agent_pool_name (Computed), agent_specification (Computed),
	// job_authorization_scope (Computed), queue_status (Computed), schedules (Computed).
	expectedAttrs := []string{
		"project_id",
		"name",
		"path",
		"revision",
		"repository",
		"ci_trigger",
		"pull_request_trigger",
		"variable",
		"agent_pool_name",
		"agent_specification",
		"job_authorization_scope",
		"queue_status",
		"schedules",
	}
	for _, attrName := range expectedAttrs {
		a, ok := attrs[attrName]
		assert.True(t, ok, "schema must declare %q attribute", attrName)
		assert.NotNil(t, a, "schema attribute %q must not be nil", attrName)
	}

	// project_id must be Required.
	if projectIDAttr, ok := attrs["project_id"]; ok {
		assert.False(t, projectIDAttr.IsOptional(), "project_id must not be Optional")
		assert.False(t, projectIDAttr.IsComputed(), "project_id must not be Computed")
	}

	// name must be Required.
	if nameAttr, ok := attrs["name"]; ok {
		assert.False(t, nameAttr.IsOptional(), "name must not be Optional")
		assert.False(t, nameAttr.IsComputed(), "name must not be Computed")
	}

	// path must be Optional.
	if pathAttr, ok := attrs["path"]; ok {
		assert.True(t, pathAttr.IsOptional(), "path must be Optional")
	}

	// revision must be Computed.
	if revAttr, ok := attrs["revision"]; ok {
		assert.True(t, revAttr.IsComputed(), "revision must be Computed")
	}

	// Nested computed attributes.
	for _, nestedName := range []string{"repository", "ci_trigger", "pull_request_trigger", "variable", "schedules"} {
		a, ok := attrs[nestedName]
		assert.True(t, ok, "schema must declare %q attribute", nestedName)
		if ok {
			assert.True(t, a.IsComputed(), "%q must be Computed", nestedName)
		}
	}
}
