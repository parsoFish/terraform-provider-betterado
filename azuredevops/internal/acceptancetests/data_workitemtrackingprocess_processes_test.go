//go:build (all || data_source_workitemtrackingprocess_processes) && !exclude_data_source_workitemtrackingprocess_processes

package acceptancetests

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccWorkitemtrackingprocessProcesses_DataSource_AllProcesses(t *testing.T) {
	tfNode := "data.betterado_workitemtrackingprocess_processes.all"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclDataSourceAllProcesses(),
				Check:  testutils.TestCheckAttrGreaterThan(tfNode, "processes.#", 0),
			},
		},
	})
}

func hclDataSourceAllProcesses() string {
	return `
data "betterado_workitemtrackingprocess_processes" "all" {
}
`
}
