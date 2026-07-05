package profile

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/stretchr/testify/require"
)

// TestDataProfileSchema verifies the schema and metadata of the betterado_profile data source
// without any network calls (offline unit test).
func TestDataProfileSchema(t *testing.T) {
	t.Run("NewProfileDataSource returns non-nil", func(t *testing.T) {
		ds := NewProfileDataSource()
		require.NotNil(t, ds)
	})

	t.Run("Metadata sets TypeName to betterado_profile", func(t *testing.T) {
		ds := NewProfileDataSource()
		ctx := context.Background()
		metaReq := datasource.MetadataRequest{ProviderTypeName: "betterado"}
		metaResp := &datasource.MetadataResponse{}
		ds.Metadata(ctx, metaReq, metaResp)
		require.Equal(t, "betterado_profile", metaResp.TypeName)
	})

	t.Run("Schema has id as required attribute", func(t *testing.T) {
		ds := NewProfileDataSource()
		ctx := context.Background()
		schemaReq := datasource.SchemaRequest{}
		schemaResp := &datasource.SchemaResponse{}
		ds.Schema(ctx, schemaReq, schemaResp)

		require.False(t, schemaResp.Diagnostics.HasError(), "Schema() should not return errors: %v", schemaResp.Diagnostics)

		attrs := schemaResp.Schema.Attributes
		require.Contains(t, attrs, "id", "schema must have 'id' attribute")

		idAttr, ok := attrs["id"].(schema.StringAttribute)
		require.True(t, ok, "'id' must be a StringAttribute")
		require.True(t, idAttr.IsRequired(), "'id' must be Required")
	})

	t.Run("Schema has display_name as computed attribute", func(t *testing.T) {
		ds := NewProfileDataSource()
		ctx := context.Background()
		schemaReq := datasource.SchemaRequest{}
		schemaResp := &datasource.SchemaResponse{}
		ds.Schema(ctx, schemaReq, schemaResp)

		require.False(t, schemaResp.Diagnostics.HasError())

		attrs := schemaResp.Schema.Attributes
		require.Contains(t, attrs, "display_name", "schema must have 'display_name' attribute")
		require.True(t, attrs["display_name"].IsComputed(), "'display_name' must be Computed")
	})

	t.Run("Schema has email_address as computed attribute", func(t *testing.T) {
		ds := NewProfileDataSource()
		ctx := context.Background()
		schemaReq := datasource.SchemaRequest{}
		schemaResp := &datasource.SchemaResponse{}
		ds.Schema(ctx, schemaReq, schemaResp)

		require.False(t, schemaResp.Diagnostics.HasError())

		attrs := schemaResp.Schema.Attributes
		require.Contains(t, attrs, "email_address", "schema must have 'email_address' attribute")
		require.True(t, attrs["email_address"].IsComputed(), "'email_address' must be Computed")
	})

	t.Run("Schema has public_alias as computed attribute", func(t *testing.T) {
		ds := NewProfileDataSource()
		ctx := context.Background()
		schemaReq := datasource.SchemaRequest{}
		schemaResp := &datasource.SchemaResponse{}
		ds.Schema(ctx, schemaReq, schemaResp)

		require.False(t, schemaResp.Diagnostics.HasError())

		attrs := schemaResp.Schema.Attributes
		require.Contains(t, attrs, "public_alias", "schema must have 'public_alias' attribute")
		require.True(t, attrs["public_alias"].IsComputed(), "'public_alias' must be Computed")
	})

	t.Run("Schema has avatar_url as computed optional attribute", func(t *testing.T) {
		ds := NewProfileDataSource()
		ctx := context.Background()
		schemaReq := datasource.SchemaRequest{}
		schemaResp := &datasource.SchemaResponse{}
		ds.Schema(ctx, schemaReq, schemaResp)

		require.False(t, schemaResp.Diagnostics.HasError())

		attrs := schemaResp.Schema.Attributes
		require.Contains(t, attrs, "avatar_url", "schema must have 'avatar_url' attribute")
		require.True(t, attrs["avatar_url"].IsComputed(), "'avatar_url' must be Computed")
	})

	t.Run("profileDataModel struct compiles correctly", func(t *testing.T) {
		m := profileDataModel{}
		require.True(t, m.ID.IsNull())
		require.True(t, m.DisplayName.IsNull())
		require.True(t, m.EmailAddress.IsNull())
		require.True(t, m.PublicAlias.IsNull())
		require.True(t, m.AvatarURL.IsNull())
	})
}
