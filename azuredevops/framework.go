package azuredevops

import (
	"github.com/hashicorp/terraform-plugin-framework/provider"
	internalprovider "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/provider"
)

// NewFrameworkProvider returns the terraform-plugin-framework provider for use
// in the mux setup in main.go. It delegates to the internal stub and serves as
// the public entry-point so that main.go can import from the azuredevops package
// without violating Go's internal-package access rules.
func NewFrameworkProvider() provider.Provider {
	return internalprovider.NewFrameworkProvider()
}
