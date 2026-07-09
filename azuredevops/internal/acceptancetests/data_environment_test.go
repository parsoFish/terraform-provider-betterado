//go:build (all || data_environment || resource_environment) && !(exclude_data_environment && exclude_resource_environment)

package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccEnvironment_dataSource(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkEnvironmentDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclDataSourceEnvironmentBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", name),
				),
			},
			{
				Config:             hclDataSourceEnvironmentBasic(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccEnvironment_dataSource_by_name(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkEnvironmentDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclDataSourceEnvironmentBasicByName(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", name),
				),
			},
			{
				Config:             hclDataSourceEnvironmentBasicByName(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclDataSourceEnvironmentBasic looks up environment by ID using the standing fixture project.
func hclDataSourceEnvironmentBasic(name string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[2]q
}

resource "betterado_environment" "test" {
  project_id  = data.betterado_project.fixture.id
  name        = %[1]q
  description = "Managed by Terraform"
}

data "betterado_environment" "test" {
  project_id     = data.betterado_project.fixture.id
  environment_id = betterado_environment.test.id
}
`, name, SharedFixtureProjectName)
}

// hclDataSourceEnvironmentBasicByName looks up environment by name using the standing fixture project.
func hclDataSourceEnvironmentBasicByName(name string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[2]q
}

resource "betterado_environment" "test" {
  project_id  = data.betterado_project.fixture.id
  name        = %[1]q
  description = "Managed by Terraform"
}

data "betterado_environment" "test" {
  project_id = data.betterado_project.fixture.id
  name       = betterado_environment.test.name
}
`, name, SharedFixtureProjectName)
}
