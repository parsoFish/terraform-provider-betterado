//go:build all || memberentitlementmanagement

package memberentitlementmanagement

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/require"
)

func TestNewUserEntitlementResource_Metadata(t *testing.T) {
	r := NewUserEntitlementResource()
	require.NotNil(t, r, "NewUserEntitlementResource() should return a non-nil resource")

	var metaResp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)

	require.Equal(t, "betterado_user_entitlement", metaResp.TypeName,
		"resource TypeName must be betterado_user_entitlement")
}
