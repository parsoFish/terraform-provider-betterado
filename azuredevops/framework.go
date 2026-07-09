package azuredevops

import (
	"github.com/hashicorp/terraform-plugin-framework/provider"
	internalprovider "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/provider"
)

// NewFrameworkProvider returns the terraform-plugin-framework provider.
// This shim allows main.go to call NewFrameworkProvider() from the azuredevops
// package without violating Go's internal-package access rules.
func NewFrameworkProvider(version string) provider.Provider {
	return internalprovider.NewFrameworkProvider(version)
}
