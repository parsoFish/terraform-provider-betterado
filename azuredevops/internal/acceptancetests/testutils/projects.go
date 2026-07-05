package testutils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// getProjectClient returns an AggregatedClient suitable for project lookups.
// When the SDKv2 provider singleton has not been configured (e.g. tests using
// ProtoV6ProviderFactories / GetProviderFactories), Meta() is nil; in
// that case we build a fresh client directly from the environment.
func getProjectClient() (*client.AggregatedClient, error) {
	if meta := GetProvider().Meta(); meta != nil {
		return meta.(*client.AggregatedClient), nil
	}
	return GetDirectClient()
}

// CheckProjectExists Given the name of an AzDO project, this will return a function that will check whether
// or not the project (1) exists in the state and (2) exist in AzDO and (3) has the correct name
func CheckProjectExists(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources["betterado_project.project"]
		if !ok {
			return fmt.Errorf("Did not find a project in the TF state")
		}

		clients, err := getProjectClient()
		if err != nil {
			return fmt.Errorf("getProjectClient: %v", err)
		}
		id := resource.Primary.ID
		project, err := readProject(clients, id)
		if err != nil {
			return fmt.Errorf("Project with ID=%s cannot be found!. Error=%v", id, err)
		}

		if *project.Name != expectedName {
			return fmt.Errorf("Project with ID=%s has Name=%s, but expected Name=%s", id, *project.Name, expectedName)
		}

		return nil
	}
}

// CheckProjectDestroyed verifies that all projects referenced in the state are destroyed. This will be invoked
// *after* terraform destroys the resource but *before* the state is wiped clean.
func CheckProjectDestroyed(s *terraform.State) error {
	clients, err := getProjectClient()
	if err != nil {
		return fmt.Errorf("getProjectClient: %v", err)
	}

	// verify that every project referenced in the state does not exist in AzDO
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "betterado_project" {
			continue
		}

		id := resource.Primary.ID

		// indicates the project still exists - this should fail the test
		if _, err := readProject(clients, id); err == nil {
			return fmt.Errorf("project with ID %s should not exist", id)
		}
	}

	return nil
}

func readProject(clients *client.AggregatedClient, identifier string) (*core.TeamProject, error) {
	return clients.CoreClient.GetProject(clients.Ctx, core.GetProjectArgs{
		ProjectId:           &identifier,
		IncludeCapabilities: converter.Bool(true),
		IncludeHistory:      converter.Bool(false),
	})
}
