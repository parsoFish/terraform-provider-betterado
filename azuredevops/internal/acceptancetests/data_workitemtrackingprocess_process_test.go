//go:build (all || data_source_workitemtrackingprocess_process) && !exclude_data_source_workitemtrackingprocess_process

package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccWorkitemtrackingprocessProcess_DataSource_Get(t *testing.T) {
	tfNode := "data.betterado_workitemtrackingprocess_process.agile"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclDataSourceAgileSystemProcess(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "id", agileSystemProcessTypeId),
				),
			},
		},
	})
}

func hclDataSourceAgileSystemProcess() string {
	return fmt.Sprintf(`
data "betterado_workitemtrackingprocess_process" "agile" {
  id = "%s"
}
`, agileSystemProcessTypeId)
}
