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

// TestPipelineRunDataSource_Metadata verifies TypeName == "betterado_pipeline_run".
func TestPipelineRunDataSource_Metadata(t *testing.T) {
	ds := plpkg.NewPipelineRunDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "betterado",
	}
	resp := &datasource.MetadataResponse{}

	ds.Metadata(context.Background(), req, resp)

	assert.Equal(t, "betterado_pipeline_run", resp.TypeName)
}

// TestPipelineRunDataSource_Schema verifies required attrs pipeline_id, run_id, project_id
// and computed attrs state, result, created_date, id, name, finished_date exist in the schema.
func TestPipelineRunDataSource_Schema(t *testing.T) {
	ds := plpkg.NewPipelineRunDataSource()

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

	runID, ok := attrs["run_id"]
	require.True(t, ok, "Schema must have 'run_id' attribute")
	assert.True(t, runID.IsRequired(), "run_id must be required")
	assert.False(t, runID.IsOptional(), "run_id must not be optional")

	projectID, ok := attrs["project_id"]
	require.True(t, ok, "Schema must have 'project_id' attribute")
	assert.True(t, projectID.IsRequired(), "project_id must be required")
	assert.False(t, projectID.IsOptional(), "project_id must not be optional")

	// Computed attributes (AC2: id, name, state, result, created_date, finished_date, pipeline_id).
	id, ok := attrs["id"]
	require.True(t, ok, "Schema must have 'id' attribute")
	assert.True(t, id.IsComputed(), "id must be computed")

	name, ok := attrs["name"]
	require.True(t, ok, "Schema must have 'name' attribute")
	assert.True(t, name.IsComputed(), "name must be computed")

	state, ok := attrs["state"]
	require.True(t, ok, "Schema must have 'state' attribute")
	assert.True(t, state.IsComputed(), "state must be computed")

	result, ok := attrs["result"]
	require.True(t, ok, "Schema must have 'result' attribute")
	assert.True(t, result.IsComputed(), "result must be computed")

	createdDate, ok := attrs["created_date"]
	require.True(t, ok, "Schema must have 'created_date' attribute")
	assert.True(t, createdDate.IsComputed(), "created_date must be computed")

	finishedDate, ok := attrs["finished_date"]
	require.True(t, ok, "Schema must have 'finished_date' attribute")
	assert.True(t, finishedDate.IsComputed(), "finished_date must be computed")
}
