//go:build all || memberentitlementmanagement

package memberentitlementmanagement

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/require"
)

func TestNewGroupEntitlementResource_Metadata(t *testing.T) {
	r := NewGroupEntitlementResource()
	require.NotNil(t, r, "NewGroupEntitlementResource() should return a non-nil resource")

	var metaResp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)

	require.Equal(t, "betterado_group_entitlement", metaResp.TypeName,
		"resource TypeName must be betterado_group_entitlement")
}
