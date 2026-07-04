//go:build all || memberentitlementmanagement

package memberentitlementmanagement

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/require"
)

func TestNewServicePrincipalEntitlementResource_Metadata(t *testing.T) {
	r := NewServicePrincipalEntitlementResource()
	require.NotNil(t, r, "NewServicePrincipalEntitlementResource() should return a non-nil resource")

	var metaResp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "betterado"}, &metaResp)

	require.Equal(t, "betterado_service_principal_entitlement", metaResp.TypeName,
		"resource TypeName must be betterado_service_principal_entitlement")
}
