//go:build all || resource_task_group_framework

package taskagent

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskGroupFramework_Schema verifies that NewTaskGroupResource() returns a
// resource with the correct type name and schema attributes.
func TestTaskGroupFramework_Schema(t *testing.T) {
	ctx := context.Background()
	r := NewTaskGroupResource()
	require.NotNil(t, r)

	// ── Type name ─────────────────────────────────────────────────────────
	metaResp := &resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{}, metaResp)
	assert.Equal(t, "betterado_task_group", metaResp.TypeName,
		"resource type name must be betterado_task_group")

	// ── Schema attributes ─────────────────────────────────────────────────
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not emit error diagnostics: %v", schemaResp.Diagnostics)

	attrs := schemaResp.Schema.Attributes

	expectedAttrs := []string{
		"task",
		"input",
		"version",
		"project_id",
		"name",
		"friendly_name",
		"description",
		"category",
		"author",
		"icon_url",
		"instance_name_format",
		"runs_on",
		"revision",
		"definition_type",
	}

	for _, attr := range expectedAttrs {
		_, ok := attrs[attr]
		assert.True(t, ok, "schema must contain attribute %q", attr)
	}
}
