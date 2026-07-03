//go:build all || resource_build_folder_framework

package build

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildFolderFramework_Schema calls Schema() on the framework resource and asserts
// that the schema declares project_id, path, and description attributes with no error
// diagnostics.
//
// This is the quality-gate test: the gate command is
//
//	go test -tags all -run TestBuildFolderFramework_Schema ./azuredevops/internal/service/build/...
func TestBuildFolderFramework_Schema(t *testing.T) {
	t.Parallel()

	r := &BuildFolderResource{}

	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}

	r.Schema(context.Background(), schemaReq, schemaResp)

	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not return error diagnostics: %s", schemaResp.Diagnostics)

	attrs := schemaResp.Schema.Attributes

	_, hasProjectID := attrs["project_id"]
	assert.True(t, hasProjectID, "schema must declare 'project_id' attribute")

	_, hasPath := attrs["path"]
	assert.True(t, hasPath, "schema must declare 'path' attribute")

	_, hasDescription := attrs["description"]
	assert.True(t, hasDescription, "schema must declare 'description' attribute")
}
