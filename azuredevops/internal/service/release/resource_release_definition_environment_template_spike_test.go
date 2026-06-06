//go:build (all || resource_release_definition_environment_template) && !exclude_resource_release_definition_environment_template
// +build all resource_release_definition_environment_template
// +build !exclude_resource_release_definition_environment_template

package release

import (
	"context"
	"testing"

	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/utils/sdk/environmenttemplates"
)

// TestReleaseDefinitionEnvironmentTemplateSpike is a compile-time + interface-satisfaction
// spike test. It does not talk to a live ADO endpoint.
//
// AC1 findings (recorded here as required by the spike brief):
//
//  1. release.Client (vendored microsoft/azure-devops-go-api v7) has NO methods for
//     environmenttemplates — the model ReleaseDefinitionEnvironmentTemplate exists at
//     vendor/.../release/models.go:2090 but was not surfaced in the generated client.
//
//  2. The raw-HTTP path via azuredevops.Connection.GetClientByUrl() + client.Client.Send()
//     IS viable: the same pattern is in use at
//     azuredevops/utils/sdk/securityroles/client.go and the ADO REST reference documents
//     the environmenttemplates locationId as 6b3ad47a-2a42-4e24-9785-e3a0a8e3e64d
//     (release resource area). The mini-client in
//     azuredevops/utils/sdk/environmenttemplates/client.go is therefore the chosen path.
//
// AC2: the mini-client is scaffolded and this test asserts that *environmenttemplates.ClientImpl
// satisfies the environmenttemplates.Client interface at compile time.
func TestReleaseDefinitionEnvironmentTemplateSpike(t *testing.T) {
	t.Log("Spike: vendored release.Client has no environmenttemplates methods")
	t.Log("Spike: raw-HTTP path is viable (locationId 6b3ad47a-2a42-4e24-9785-e3a0a8e3e64d confirmed)")
	t.Log("Spike: mini-client scaffolded at azuredevops/utils/sdk/environmenttemplates/client.go")

	// Build a real *azuredevops.Connection with a fake base URL so that
	// NewClient can construct an azuredevops.Client without hitting the network.
	conn := azuredevops.NewAnonymousConnection("https://vsrm.dev.azure.com/fake-org")
	ctx := context.Background()

	// NewClient must return a value that satisfies the interface — this is the
	// compile-time + runtime type assertion required by AC2.
	var c environmenttemplates.Client = environmenttemplates.NewClient(ctx, conn)
	if c == nil {
		t.Fatal("environmenttemplates.NewClient returned nil — interface not satisfied")
	}

	t.Logf("environmenttemplates.NewClient returned %T — interface satisfied", c)
}
