//go:build all || data_source_pipeline

package pipelines_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	plpkg "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/pipelines"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPipelineDataSource_Metadata verifies TypeName == "betterado_pipeline".
func TestPipelineDataSource_Metadata(t *testing.T) {
	ds := plpkg.NewPipelineDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "betterado",
	}
	resp := &datasource.MetadataResponse{}

	ds.Metadata(context.Background(), req, resp)

	assert.Equal(t, "betterado_pipeline", resp.TypeName)
}

// TestPipelineDataSource_Schema verifies required attrs pipeline_id, project_id
// and computed attr name exist in the schema.
func TestPipelineDataSource_Schema(t *testing.T) {
	ds := plpkg.NewPipelineDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(context.Background(), req, resp)

	require.False(t, resp.Diagnostics.HasError(), "Schema() must not produce diagnostics")

	attrs := resp.Schema.Attributes
	require.NotNil(t, attrs, "Schema attributes must not be nil")

	// Required attributes.
	pipelineID, ok := attrs["pipeline_id"]
	require.True(t, ok, "Schema must have 'pipeline_id' attribute")
	assert.True(t, pipelineID.IsRequired(), "pipeline_id must be required")
	assert.False(t, pipelineID.IsOptional(), "pipeline_id must not be optional")

	projectID, ok := attrs["project_id"]
	require.True(t, ok, "Schema must have 'project_id' attribute")
	assert.True(t, projectID.IsRequired(), "project_id must be required")
	assert.False(t, projectID.IsOptional(), "project_id must not be optional")

	// Computed attributes.
	name, ok := attrs["name"]
	require.True(t, ok, "Schema must have 'name' attribute")
	assert.True(t, name.IsComputed(), "name must be computed")
	assert.False(t, name.IsRequired(), "name must not be required")

	id, ok := attrs["id"]
	require.True(t, ok, "Schema must have 'id' attribute")
	assert.True(t, id.IsComputed(), "id must be computed")

	folder, ok := attrs["folder"]
	require.True(t, ok, "Schema must have 'folder' attribute")
	assert.True(t, folder.IsComputed(), "folder must be computed")

	cfgType, ok := attrs["configuration_type"]
	require.True(t, ok, "Schema must have 'configuration_type' attribute")
	assert.True(t, cfgType.IsComputed(), "configuration_type must be computed")

	revision, ok := attrs["revision"]
	require.True(t, ok, "Schema must have 'revision' attribute")
	assert.True(t, revision.IsComputed(), "revision must be computed")

	url, ok := attrs["url"]
	require.True(t, ok, "Schema must have 'url' attribute")
	assert.True(t, url.IsComputed(), "url must be computed")
}
