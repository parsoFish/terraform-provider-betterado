//go:build all || resource_test_plan

package testplan

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnitTestPlanDataSource_Schema verifies that NewTestPlanDataSource() exposes
// the correct type name and required schema attributes.
func TestUnitTestPlanDataSource_Schema(t *testing.T) {
	ctx := context.Background()
	ds := NewTestPlanDataSource()
	require.NotNil(t, ds)

	metaResp := &datasource.MetadataResponse{}
	ds.Metadata(ctx, datasource.MetadataRequest{}, metaResp)
	assert.Equal(t, "betterado_test_plan", metaResp.TypeName)

	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not emit diagnostics: %v", schemaResp.Diagnostics)

	for _, attr := range []string{"id", "project_id", "plan_id", "name", "area_path", "iteration_path", "start_date", "end_date"} {
		_, ok := schemaResp.Schema.Attributes[attr]
		assert.True(t, ok, "attribute %q must exist in datasource schema", attr)
	}
}
