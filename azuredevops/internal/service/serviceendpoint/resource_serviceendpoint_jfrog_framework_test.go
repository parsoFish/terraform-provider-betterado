//go:build all || resource_serviceendpoint_jfrog_artifactory_v2 || resource_serviceendpoint_jfrog_distribution_v2 || resource_serviceendpoint_jfrog_platform_v2 || resource_serviceendpoint_jfrog_xray_v2
// +build all resource_serviceendpoint_jfrog_artifactory_v2 resource_serviceendpoint_jfrog_distribution_v2 resource_serviceendpoint_jfrog_platform_v2 resource_serviceendpoint_jfrog_xray_v2

package serviceendpoint

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ──────────────────────────────────────────────────────────────────────────────
// JFrog Artifactory V2 framework resource unit tests
// ──────────────────────────────────────────────────────────────────────────────

func TestServiceEndpointJFrogArtifactoryV2Resource_Metadata(t *testing.T) {
	ctx := context.Background()
	r := NewServiceEndpointJFrogArtifactoryV2Resource()
	require.NotNil(t, r)

	resp := &resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "betterado"}, resp)
	assert.Equal(t, "betterado_serviceendpoint_jfrog_artifactory_v2", resp.TypeName)
}

func TestServiceEndpointJFrogArtifactoryV2Resource_Schema(t *testing.T) {
	ctx := context.Background()
	r := NewServiceEndpointJFrogArtifactoryV2Resource()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not emit error diagnostics: %v", schemaResp.Diagnostics)

	attrs := schemaResp.Schema.Attributes
	for _, name := range []string{"id", "project_id", "service_endpoint_name", "description", "url", "authorization"} {
		_, ok := attrs[name]
		assert.True(t, ok, "schema must contain attribute %q", name)
	}
	blocks := schemaResp.Schema.Blocks
	for _, name := range []string{"authentication_token", "authentication_basic"} {
		_, ok := blocks[name]
		assert.True(t, ok, "schema must contain block %q", name)
	}
}

func TestServiceEndpointJFrogArtifactoryV2Resource_BuildEndpoint_Token(t *testing.T) {
	plan := serviceEndpointJFrogArtifactoryV2Model{
		ServiceEndpointName: types.StringValue("test-artifactory"),
		Description:         types.StringValue("desc"),
		URL:                 types.StringValue("https://myartifactory.example.com"),
		AuthenticationToken: []seJFrogArtifactoryV2TokenModel{
			{Token: types.StringValue("mytoken")},
		},
	}
	ep, scheme, err := seJFrogArtifactoryV2BuildEndpoint(plan)
	require.NoError(t, err)
	require.NotNil(t, ep)
	assert.Equal(t, "Token", scheme)
	assert.Equal(t, "jfrogArtifactoryService", *ep.Type)
	require.NotNil(t, ep.Authorization)
	require.NotNil(t, ep.Authorization.Parameters)
	assert.Equal(t, "mytoken", (*ep.Authorization.Parameters)["apitoken"])
}

func TestServiceEndpointJFrogArtifactoryV2Resource_BuildEndpoint_BasicAuth(t *testing.T) {
	plan := serviceEndpointJFrogArtifactoryV2Model{
		ServiceEndpointName: types.StringValue("test-artifactory-basic"),
		Description:         types.StringValue("desc"),
		URL:                 types.StringValue("https://myartifactory.example.com"),
		AuthenticationBasic: []seJFrogArtifactoryV2BasicModel{
			{Username: types.StringValue("user"), Password: types.StringValue("pass")},
		},
	}
	ep, scheme, err := seJFrogArtifactoryV2BuildEndpoint(plan)
	require.NoError(t, err)
	require.NotNil(t, ep)
	assert.Equal(t, "UsernamePassword", scheme)
	assert.Equal(t, "jfrogArtifactoryService", *ep.Type)
	require.NotNil(t, ep.Authorization)
	require.NotNil(t, ep.Authorization.Parameters)
	assert.Equal(t, "user", (*ep.Authorization.Parameters)["username"])
	assert.Equal(t, "pass", (*ep.Authorization.Parameters)["password"])
}

func TestServiceEndpointJFrogArtifactoryV2Resource_BuildEndpoint_NoAuth(t *testing.T) {
	plan := serviceEndpointJFrogArtifactoryV2Model{
		ServiceEndpointName: types.StringValue("test-no-auth"),
		Description:         types.StringValue("desc"),
		URL:                 types.StringValue("https://myartifactory.example.com"),
	}
	_, _, err := seJFrogArtifactoryV2BuildEndpoint(plan)
	require.Error(t, err, "BuildEndpoint should error when no auth block is provided")
}

// ──────────────────────────────────────────────────────────────────────────────
// JFrog Distribution V2 framework resource unit tests
// ──────────────────────────────────────────────────────────────────────────────

func TestServiceEndpointJFrogDistributionV2Resource_Metadata(t *testing.T) {
	ctx := context.Background()
	r := NewServiceEndpointJFrogDistributionV2Resource()
	require.NotNil(t, r)

	resp := &resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "betterado"}, resp)
	assert.Equal(t, "betterado_serviceendpoint_jfrog_distribution_v2", resp.TypeName)
}

func TestServiceEndpointJFrogDistributionV2Resource_Schema(t *testing.T) {
	ctx := context.Background()
	r := NewServiceEndpointJFrogDistributionV2Resource()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not emit error diagnostics: %v", schemaResp.Diagnostics)

	attrs := schemaResp.Schema.Attributes
	for _, name := range []string{"id", "project_id", "service_endpoint_name", "description", "url", "authorization"} {
		_, ok := attrs[name]
		assert.True(t, ok, "schema must contain attribute %q", name)
	}
	blocks := schemaResp.Schema.Blocks
	for _, name := range []string{"authentication_token", "authentication_basic"} {
		_, ok := blocks[name]
		assert.True(t, ok, "schema must contain block %q", name)
	}
}

func TestServiceEndpointJFrogDistributionV2Resource_BuildEndpoint_Token(t *testing.T) {
	plan := serviceEndpointJFrogDistributionV2Model{
		ServiceEndpointName: types.StringValue("test-distribution"),
		Description:         types.StringValue("desc"),
		URL:                 types.StringValue("https://mydistribution.example.com"),
		AuthenticationToken: []seJFrogDistributionV2TokenModel{
			{Token: types.StringValue("disttoken")},
		},
	}
	ep, scheme, err := seJFrogDistributionV2BuildEndpoint(plan)
	require.NoError(t, err)
	require.NotNil(t, ep)
	assert.Equal(t, "Token", scheme)
	assert.Equal(t, "jfrogDistributionService", *ep.Type)
	require.NotNil(t, ep.Authorization)
	require.NotNil(t, ep.Authorization.Parameters)
	assert.Equal(t, "disttoken", (*ep.Authorization.Parameters)["apitoken"])
}

func TestServiceEndpointJFrogDistributionV2Resource_BuildEndpoint_BasicAuth(t *testing.T) {
	plan := serviceEndpointJFrogDistributionV2Model{
		ServiceEndpointName: types.StringValue("test-distribution-basic"),
		Description:         types.StringValue("desc"),
		URL:                 types.StringValue("https://mydistribution.example.com"),
		AuthenticationBasic: []seJFrogDistributionV2BasicModel{
			{Username: types.StringValue("distuser"), Password: types.StringValue("distpass")},
		},
	}
	ep, scheme, err := seJFrogDistributionV2BuildEndpoint(plan)
	require.NoError(t, err)
	require.NotNil(t, ep)
	assert.Equal(t, "UsernamePassword", scheme)
	assert.Equal(t, "jfrogDistributionService", *ep.Type)
	require.NotNil(t, ep.Authorization)
	require.NotNil(t, ep.Authorization.Parameters)
	assert.Equal(t, "distuser", (*ep.Authorization.Parameters)["username"])
	assert.Equal(t, "distpass", (*ep.Authorization.Parameters)["password"])
}

// ──────────────────────────────────────────────────────────────────────────────
// JFrog Platform V2 framework resource unit tests
// ──────────────────────────────────────────────────────────────────────────────

func TestServiceEndpointJFrogPlatformV2Resource_Metadata(t *testing.T) {
	ctx := context.Background()
	r := NewServiceEndpointJFrogPlatformV2Resource()
	require.NotNil(t, r)

	resp := &resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "betterado"}, resp)
	assert.Equal(t, "betterado_serviceendpoint_jfrog_platform_v2", resp.TypeName)
}

func TestServiceEndpointJFrogPlatformV2Resource_Schema(t *testing.T) {
	ctx := context.Background()
	r := NewServiceEndpointJFrogPlatformV2Resource()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not emit error diagnostics: %v", schemaResp.Diagnostics)

	attrs := schemaResp.Schema.Attributes
	for _, name := range []string{"id", "project_id", "service_endpoint_name", "description", "url", "authorization"} {
		_, ok := attrs[name]
		assert.True(t, ok, "schema must contain attribute %q", name)
	}
	blocks := schemaResp.Schema.Blocks
	for _, name := range []string{"authentication_token", "authentication_basic"} {
		_, ok := blocks[name]
		assert.True(t, ok, "schema must contain block %q", name)
	}
}

func TestServiceEndpointJFrogPlatformV2Resource_BuildEndpoint_Token(t *testing.T) {
	plan := serviceEndpointJFrogPlatformV2Model{
		ServiceEndpointName: types.StringValue("test-platform"),
		Description:         types.StringValue("desc"),
		URL:                 types.StringValue("https://myplatform.example.com"),
		AuthenticationToken: []seJFrogPlatformV2TokenModel{
			{Token: types.StringValue("platformtoken")},
		},
	}
	ep, scheme, err := seJFrogPlatformV2BuildEndpoint(plan)
	require.NoError(t, err)
	require.NotNil(t, ep)
	assert.Equal(t, "Token", scheme)
	assert.Equal(t, "jfrogPlatformService", *ep.Type)
	require.NotNil(t, ep.Authorization)
	require.NotNil(t, ep.Authorization.Parameters)
	assert.Equal(t, "platformtoken", (*ep.Authorization.Parameters)["apitoken"])
}

func TestServiceEndpointJFrogPlatformV2Resource_BuildEndpoint_BasicAuth(t *testing.T) {
	plan := serviceEndpointJFrogPlatformV2Model{
		ServiceEndpointName: types.StringValue("test-platform-basic"),
		Description:         types.StringValue("desc"),
		URL:                 types.StringValue("https://myplatform.example.com"),
		AuthenticationBasic: []seJFrogPlatformV2BasicModel{
			{Username: types.StringValue("platuser"), Password: types.StringValue("platpass")},
		},
	}
	ep, scheme, err := seJFrogPlatformV2BuildEndpoint(plan)
	require.NoError(t, err)
	require.NotNil(t, ep)
	assert.Equal(t, "UsernamePassword", scheme)
	assert.Equal(t, "jfrogPlatformService", *ep.Type)
	require.NotNil(t, ep.Authorization)
	require.NotNil(t, ep.Authorization.Parameters)
	assert.Equal(t, "platuser", (*ep.Authorization.Parameters)["username"])
	assert.Equal(t, "platpass", (*ep.Authorization.Parameters)["password"])
}

// ──────────────────────────────────────────────────────────────────────────────
// JFrog XRay V2 framework resource unit tests
// ──────────────────────────────────────────────────────────────────────────────

func TestServiceEndpointJFrogXRayV2Resource_Metadata(t *testing.T) {
	ctx := context.Background()
	r := NewServiceEndpointJFrogXRayV2Resource()
	require.NotNil(t, r)

	resp := &resource.MetadataResponse{}
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "betterado"}, resp)
	assert.Equal(t, "betterado_serviceendpoint_jfrog_xray_v2", resp.TypeName)
}

func TestServiceEndpointJFrogXRayV2Resource_Schema(t *testing.T) {
	ctx := context.Background()
	r := NewServiceEndpointJFrogXRayV2Resource()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not emit error diagnostics: %v", schemaResp.Diagnostics)

	attrs := schemaResp.Schema.Attributes
	for _, name := range []string{"id", "project_id", "service_endpoint_name", "description", "url", "authorization"} {
		_, ok := attrs[name]
		assert.True(t, ok, "schema must contain attribute %q", name)
	}
	blocks := schemaResp.Schema.Blocks
	for _, name := range []string{"authentication_token", "authentication_basic"} {
		_, ok := blocks[name]
		assert.True(t, ok, "schema must contain block %q", name)
	}
}

func TestServiceEndpointJFrogXRayV2Resource_BuildEndpoint_Token(t *testing.T) {
	plan := serviceEndpointJFrogXRayV2Model{
		ServiceEndpointName: types.StringValue("test-xray"),
		Description:         types.StringValue("desc"),
		URL:                 types.StringValue("https://myxray.example.com"),
		AuthenticationToken: []seJFrogXRayV2TokenModel{
			{Token: types.StringValue("xraytoken")},
		},
	}
	ep, scheme, err := seJFrogXRayV2BuildEndpoint(plan)
	require.NoError(t, err)
	require.NotNil(t, ep)
	assert.Equal(t, "Token", scheme)
	assert.Equal(t, "jfrogXrayService", *ep.Type)
	require.NotNil(t, ep.Authorization)
	require.NotNil(t, ep.Authorization.Parameters)
	assert.Equal(t, "xraytoken", (*ep.Authorization.Parameters)["apitoken"])
}

func TestServiceEndpointJFrogXRayV2Resource_BuildEndpoint_BasicAuth(t *testing.T) {
	plan := serviceEndpointJFrogXRayV2Model{
		ServiceEndpointName: types.StringValue("test-xray-basic"),
		Description:         types.StringValue("desc"),
		URL:                 types.StringValue("https://myxray.example.com"),
		AuthenticationBasic: []seJFrogXRayV2BasicModel{
			{Username: types.StringValue("xrayuser"), Password: types.StringValue("xraypass")},
		},
	}
	ep, scheme, err := seJFrogXRayV2BuildEndpoint(plan)
	require.NoError(t, err)
	require.NotNil(t, ep)
	assert.Equal(t, "UsernamePassword", scheme)
	assert.Equal(t, "jfrogXrayService", *ep.Type)
	require.NotNil(t, ep.Authorization)
	require.NotNil(t, ep.Authorization.Parameters)
	assert.Equal(t, "xrayuser", (*ep.Authorization.Parameters)["username"])
	assert.Equal(t, "xraypass", (*ep.Authorization.Parameters)["password"])
}
