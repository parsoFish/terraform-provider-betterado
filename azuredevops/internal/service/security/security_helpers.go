package security

// security_helpers.go holds shared helper functions used by the framework
// implementation of betterado_security_permissions.

import (
	"fmt"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// resolveIdentityDescriptor resolves an AAD / VSIDS subject descriptor to a
// legacy identity descriptor string required by the Security REST API.
func resolveIdentityDescriptor(clients *client.AggregatedClient, subjectDescriptor string) (string, error) {
	identities, err := clients.IdentityClient.ReadIdentities(clients.Ctx, identity.ReadIdentitiesArgs{
		SubjectDescriptors: &subjectDescriptor,
	})
	if err != nil {
		return "", fmt.Errorf("reading identity for subject descriptor: %v", err)
	}

	if identities == nil || len(*identities) == 0 {
		return "", fmt.Errorf("no identity found for subject descriptor '%s'", subjectDescriptor)
	}

	id := (*identities)[0]
	if id.Descriptor == nil {
		return "", fmt.Errorf("identity descriptor is nil for subject descriptor '%s'", subjectDescriptor)
	}

	// Check if identity is active
	if id.IsActive != nil && !*id.IsActive {
		return "", fmt.Errorf("identity for subject descriptor '%s' is not active", subjectDescriptor)
	}

	return *id.Descriptor, nil
}
