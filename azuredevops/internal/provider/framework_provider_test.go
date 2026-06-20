//go:build all || provider_framework

package provider_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	frameworkprovider "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/provider"
	"github.com/stretchr/testify/require"
)

func TestFrameworkProvider_HasTaskGroupResource(t *testing.T) {
	p := frameworkprovider.NewFrameworkProvider()
	provWithResources, ok := p.(interface {
		Resources(context.Context) []func() resource.Resource
	})
	require.True(t, ok, "framework provider must implement Resources()")

	factories := provWithResources.Resources(context.Background())
	require.NotEmpty(t, factories, "framework provider must have at least one resource factory")

	found := false
	for _, factory := range factories {
		r := factory()
		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "azuredevops"}, &metaResp)
		if metaResp.TypeName == "azuredevops_betterado_task_group" || metaResp.TypeName == "betterado_task_group" {
			found = true
			break
		}
	}
	require.True(t, found, "framework provider must register betterado_task_group")
}
