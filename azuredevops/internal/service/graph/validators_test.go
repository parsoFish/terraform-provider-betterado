package graph_test

// Unit tests for validator wiring on betterado_group and betterado_group_membership.
//
// These tests use resource.UnitTest (no TF_ACC required) to verify that the
// conflict-triangle on betterado_group and the mode-enum on
// betterado_group_membership raise the expected diagnostics.
//
// The minimal provider factory serves only the two resources under test so the
// tests have no network dependency.

import (
	"context"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	tftest "github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/graph"
)

// ── minimal test provider ─────────────────────────────────────────────────────

// validatorTestProvider is a bare-minimum provider that exposes the two graph
// resources. It never calls Configure so it works without credentials.
type validatorTestProvider struct{}

func (p *validatorTestProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "betterado"
}

func (p *validatorTestProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

func (p *validatorTestProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

func (p *validatorTestProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		graph.NewGroupResource,
		graph.NewGroupMembershipResource,
	}
}

func (p *validatorTestProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func providerFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"betterado": providerserver.NewProtocol6WithError(&validatorTestProvider{}),
	}
}

// ── betterado_group conflict-triangle tests ───────────────────────────────────

// TestGroupResource_ConflictOriginIDAndMail verifies that setting origin_id
// together with mail produces a validation error.
func TestGroupResource_ConflictOriginIDAndMail(t *testing.T) {
	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: providerFactories(),
		Steps: []tftest.TestStep{
			{
				Config: `
resource "betterado_group" "test" {
  origin_id = "some-origin-id"
  mail      = "group@example.com"
}
`,
				ExpectError: regexp.MustCompile(`(?i)(Invalid Attribute Combination|cannot be configured together)`),
			},
		},
	})
}

// TestGroupResource_ConflictOriginIDAndDisplayName verifies that setting
// origin_id together with display_name produces a validation error.
func TestGroupResource_ConflictOriginIDAndDisplayName(t *testing.T) {
	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: providerFactories(),
		Steps: []tftest.TestStep{
			{
				Config: `
resource "betterado_group" "test" {
  origin_id    = "some-origin-id"
  display_name = "My Group"
}
`,
				ExpectError: regexp.MustCompile(`(?i)(Invalid Attribute Combination|cannot be configured together)`),
			},
		},
	})
}

// TestGroupResource_ConflictMailAndDisplayName verifies that setting mail
// together with display_name produces a validation error.
func TestGroupResource_ConflictMailAndDisplayName(t *testing.T) {
	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: providerFactories(),
		Steps: []tftest.TestStep{
			{
				Config: `
resource "betterado_group" "test" {
  mail         = "group@example.com"
  display_name = "My Group"
}
`,
				ExpectError: regexp.MustCompile(`(?i)(Invalid Attribute Combination|cannot be configured together)`),
			},
		},
	})
}

// TestGroupResource_ConflictOriginIDAndScope verifies that origin_id and scope
// cannot be set together.
func TestGroupResource_ConflictOriginIDAndScope(t *testing.T) {
	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: providerFactories(),
		Steps: []tftest.TestStep{
			{
				Config: `
resource "betterado_group" "test" {
  origin_id = "some-origin-id"
  scope     = "00000000-0000-0000-0000-000000000001"
}
`,
				ExpectError: regexp.MustCompile(`(?i)(Invalid Attribute Combination|cannot be configured together)`),
			},
		},
	})
}

// TestGroupResource_ConflictMailAndScope verifies that mail and scope cannot be
// set together.
func TestGroupResource_ConflictMailAndScope(t *testing.T) {
	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: providerFactories(),
		Steps: []tftest.TestStep{
			{
				Config: `
resource "betterado_group" "test" {
  mail  = "group@example.com"
  scope = "00000000-0000-0000-0000-000000000001"
}
`,
				ExpectError: regexp.MustCompile(`(?i)(Invalid Attribute Combination|cannot be configured together)`),
			},
		},
	})
}

// ── betterado_group_membership mode-enum tests ────────────────────────────────

// TestGroupMembershipResource_InvalidMode verifies that an invalid mode value
// produces a validation error.
func TestGroupMembershipResource_InvalidMode(t *testing.T) {
	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: providerFactories(),
		Steps: []tftest.TestStep{
			{
				Config: `
resource "betterado_group_membership" "test" {
  group   = "vssgp.some-descriptor"
  mode    = "replace"
  members = []
}
`,
				ExpectError: regexp.MustCompile(`(?i)(invalid|value must be one of)`),
			},
		},
	})
}

// TestGroupMembershipResource_EmptyGroupDescriptor verifies that an empty group
// descriptor produces a validation error.
func TestGroupMembershipResource_EmptyGroupDescriptor(t *testing.T) {
	tftest.UnitTest(t, tftest.TestCase{
		ProtoV6ProviderFactories: providerFactories(),
		Steps: []tftest.TestStep{
			{
				Config: `
resource "betterado_group_membership" "test" {
  group   = ""
  members = []
}
`,
				ExpectError: regexp.MustCompile(`(?i)(invalid|length|must be at least|empty)`),
			},
		},
	})
}
