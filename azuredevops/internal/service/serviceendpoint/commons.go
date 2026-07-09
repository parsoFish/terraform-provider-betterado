package serviceendpoint

// Shared enum types for the framework service endpoint resources. The SDKv2
// scaffolding that used to live in this file was deleted in the mux-free
// cutover; these string enums are the only survivors, used by the AzureRM /
// AzureCR framework resources' auth-scheme handling.

type EndpointAuthenticationScheme string

const (
	ServicePrincipal           EndpointAuthenticationScheme = "ServicePrincipal"
	ManagedServiceIdentity     EndpointAuthenticationScheme = "ManagedServiceIdentity"
	WorkloadIdentityFederation EndpointAuthenticationScheme = "WorkloadIdentityFederation"
)

type EndpointCreationMode string

const (
	Automatic EndpointCreationMode = "Automatic"
	Manual    EndpointCreationMode = "Manual"
)
