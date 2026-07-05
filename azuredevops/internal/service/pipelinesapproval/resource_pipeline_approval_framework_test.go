//go:build all || resource_pipeline_approval
// +build all resource_pipeline_approval

package pipelinesapproval_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	plpkg "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/pipelinesapproval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPipelineApprovalResource_Metadata verifies that the resource TypeName is
// "betterado_pipeline_approval".
func TestPipelineApprovalResource_Metadata(t *testing.T) {
	r := plpkg.NewPipelineApprovalResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "betterado",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "betterado_pipeline_approval", resp.TypeName)
}

// TestPipelineApprovalResource_Schema verifies that the schema has the expected
// attributes with the correct cardinalities:
//   - id:          computed
//   - project_id:  required
//   - approval_id: required
//   - status:      required
//   - comment:     optional
func TestPipelineApprovalResource_Schema(t *testing.T) {
	r := plpkg.NewPipelineApprovalResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	require.False(t, resp.Diagnostics.HasError(), "Schema() must not produce diagnostics")

	attrs := resp.Schema.Attributes
	require.NotNil(t, attrs, "Schema attributes must not be nil")

	// id — computed only.
	id, ok := attrs["id"]
	require.True(t, ok, "Schema must have 'id' attribute")
	assert.True(t, id.IsComputed(), "id must be computed")
	assert.False(t, id.IsRequired(), "id must not be required")
	assert.False(t, id.IsOptional(), "id must not be optional")

	// project_id — required.
	projectID, ok := attrs["project_id"]
	require.True(t, ok, "Schema must have 'project_id' attribute")
	assert.True(t, projectID.IsRequired(), "project_id must be required")
	assert.False(t, projectID.IsOptional(), "project_id must not be optional")

	// approval_id — required.
	approvalID, ok := attrs["approval_id"]
	require.True(t, ok, "Schema must have 'approval_id' attribute")
	assert.True(t, approvalID.IsRequired(), "approval_id must be required")
	assert.False(t, approvalID.IsOptional(), "approval_id must not be optional")

	// status — required.
	status, ok := attrs["status"]
	require.True(t, ok, "Schema must have 'status' attribute")
	assert.True(t, status.IsRequired(), "status must be required")
	assert.False(t, status.IsOptional(), "status must not be optional")

	// comment — optional (not required, not computed-only).
	comment, ok := attrs["comment"]
	require.True(t, ok, "Schema must have 'comment' attribute")
	assert.True(t, comment.IsOptional(), "comment must be optional")
	assert.False(t, comment.IsRequired(), "comment must not be required")
}
