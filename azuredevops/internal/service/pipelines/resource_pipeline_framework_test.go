//go:build all || resource_pipeline

package pipelines_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	plpkg "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/pipelines"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPipelineResource_Metadata verifies that the resource TypeName is "betterado_pipeline".
func TestPipelineResource_Metadata(t *testing.T) {
	r := plpkg.NewPipelineResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "betterado",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "betterado_pipeline", resp.TypeName)
}

// TestPipelineResource_Schema verifies that the schema has:
//   - required attributes: project_id, name
//   - optional/computed attributes: folder, configuration_type
//   - computed-only attributes: id, revision, url
func TestPipelineResource_Schema(t *testing.T) {
	r := plpkg.NewPipelineResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	require.False(t, resp.Diagnostics.HasError(), "Schema() must not produce diagnostics")

	attrs := resp.Schema.Attributes
	require.NotNil(t, attrs, "Schema attributes must not be nil")

	// Required attributes.
	projectID, ok := attrs["project_id"]
	require.True(t, ok, "Schema must have 'project_id' attribute")
	assert.False(t, projectID.IsOptional(), "project_id must not be optional")
	assert.True(t, projectID.IsRequired(), "project_id must be required")

	name, ok := attrs["name"]
	require.True(t, ok, "Schema must have 'name' attribute")
	assert.False(t, name.IsOptional(), "name must not be optional")
	assert.True(t, name.IsRequired(), "name must be required")

	// Computed-only attributes.
	id, ok := attrs["id"]
	require.True(t, ok, "Schema must have 'id' attribute")
	assert.True(t, id.IsComputed(), "id must be computed")
	assert.False(t, id.IsRequired(), "id must not be required")

	revision, ok := attrs["revision"]
	require.True(t, ok, "Schema must have 'revision' attribute")
	assert.True(t, revision.IsComputed(), "revision must be computed")

	url, ok := attrs["url"]
	require.True(t, ok, "Schema must have 'url' attribute")
	assert.True(t, url.IsComputed(), "url must be computed")

	// Optional+computed attributes.
	folder, ok := attrs["folder"]
	require.True(t, ok, "Schema must have 'folder' attribute")
	assert.True(t, folder.IsOptional(), "folder must be optional")
	assert.True(t, folder.IsComputed(), "folder must be computed")

	cfgType, ok := attrs["configuration_type"]
	require.True(t, ok, "Schema must have 'configuration_type' attribute")
	assert.True(t, cfgType.IsOptional(), "configuration_type must be optional")
	assert.True(t, cfgType.IsComputed(), "configuration_type must be computed")
}
