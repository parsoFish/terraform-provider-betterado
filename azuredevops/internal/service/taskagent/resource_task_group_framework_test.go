//go:build all || resource_task_group_framework

package taskagent

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskGroupFramework_UpgradeState verifies that UpgradeState() returns a map with
// exactly one entry keyed at version 0.
func TestTaskGroupFramework_UpgradeState(t *testing.T) {
	ctx := context.Background()
	r := &TaskGroupResource{}

	upgraders := r.UpgradeState(ctx)
	require.Len(t, upgraders, 1, "UpgradeState must return exactly one upgrader")
	_, ok := upgraders[0]
	assert.True(t, ok, "UpgradeState map must contain key 0")
}

// TestTaskGroupFramework_SchemaVersion verifies that Schema() declares Version 1.
func TestTaskGroupFramework_SchemaVersion(t *testing.T) {
	ctx := context.Background()
	r := &TaskGroupResource{}

	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	require.False(t, resp.Diagnostics.HasError(), "Schema() must not error: %s", resp.Diagnostics)
	assert.Equal(t, int64(1), resp.Schema.Version, "Schema.Version must be 1")
}

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
