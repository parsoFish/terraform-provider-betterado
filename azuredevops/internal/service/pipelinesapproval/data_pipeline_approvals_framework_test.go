//go:build all || data_pipeline_approvals
// +build all data_pipeline_approvals

package pipelinesapproval_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	plpkg "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/pipelinesapproval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPipelineApprovalsDataSource_Metadata verifies that the TypeName is
// "betterado_pipeline_approvals".
func TestPipelineApprovalsDataSource_Metadata(t *testing.T) {
	d := plpkg.NewPipelineApprovalsDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "betterado",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	assert.Equal(t, "betterado_pipeline_approvals", resp.TypeName)
}

// TestPipelineApprovalsDataSource_Schema verifies that the schema has the
// required and computed attributes expected by AC1 and AC2:
//   - id:              computed
//   - project_id:      required
//   - pipeline_run_id: required
//   - approvals:       computed list containing id, status, comment, instructions, approved_by_id
func TestPipelineApprovalsDataSource_Schema(t *testing.T) {
	d := plpkg.NewPipelineApprovalsDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	require.False(t, resp.Diagnostics.HasError(), "Schema() must not produce diagnostics")

	attrs := resp.Schema.Attributes
	require.NotNil(t, attrs, "Schema attributes must not be nil")

	// id — computed only.
	id, ok := attrs["id"]
	require.True(t, ok, "Schema must have 'id' attribute")
	assert.True(t, id.IsComputed(), "id must be computed")
	assert.False(t, id.IsRequired(), "id must not be required")

	// project_id — required.
	projectID, ok := attrs["project_id"]
	require.True(t, ok, "Schema must have 'project_id' attribute")
	assert.True(t, projectID.IsRequired(), "project_id must be required")
	assert.False(t, projectID.IsOptional(), "project_id must not be optional")

	// pipeline_run_id — required.
	pipelineRunID, ok := attrs["pipeline_run_id"]
	require.True(t, ok, "Schema must have 'pipeline_run_id' attribute")
	assert.True(t, pipelineRunID.IsRequired(), "pipeline_run_id must be required")
	assert.False(t, pipelineRunID.IsOptional(), "pipeline_run_id must not be optional")

	// approvals — computed list.
	approvalsRaw, ok := attrs["approvals"]
	require.True(t, ok, "Schema must have 'approvals' attribute")
	assert.True(t, approvalsRaw.IsComputed(), "approvals must be computed")
	assert.False(t, approvalsRaw.IsRequired(), "approvals must not be required")

	// Verify approvals is a ListNestedAttribute with expected nested attributes.
	approvalsList, ok := approvalsRaw.(schema.ListNestedAttribute)
	require.True(t, ok, "approvals must be a schema.ListNestedAttribute")

	nestedAttrs := approvalsList.NestedObject.Attributes

	for _, name := range []string{"id", "status", "comment", "instructions", "approved_by_id"} {
		attr, exists := nestedAttrs[name]
		require.True(t, exists, "approvals nested object must have %q attribute", name)
		assert.True(t, attr.IsComputed(), "approvals.%s must be computed", name)
	}
}
