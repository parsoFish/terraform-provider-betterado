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
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkEnvironmentDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclDataSourceEnvironmentBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", name),
				),
			},
		},
	})
}

func TestAccEnvironment_dataSource_by_name(t *testing.T) {
	name := testutils.GenerateResourceName()

	tfNode := "betterado_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.GetProviders(),
		CheckDestroy: checkEnvironmentDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclDataSourceEnvironmentBasicByName(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", name),
				),
			},
		},
	})
}

func hclDataSourceEnvironmentBasic(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_environment" "test" {
  project_id  = betterado_project.test.id
  name        = "%[1]s"
  description = "Managed by Terraform"
}

data "betterado_environment" "test" {
  project_id     = betterado_project.test.id
  environment_id = betterado_environment.test.id
}
`, name)
}

func hclDataSourceEnvironmentBasicByName(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_environment" "test" {
  project_id  = betterado_project.test.id
  name        = "%[1]s"
  description = "Managed by Terraform"
}

data "betterado_environment" "test" {
  project_id = betterado_project.test.id
  name       = betterado_environment.test.name
}
`, name)
}
