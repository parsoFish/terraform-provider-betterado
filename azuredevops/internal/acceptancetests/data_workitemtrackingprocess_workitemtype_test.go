//go:build (all || data_source_workitemtrackingprocess_workitemtype) && !exclude_data_source_workitemtrackingprocess_workitemtype

package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccWorkitemtrackingprocessWorkItemType_DataSource_Get(t *testing.T) {
	workItemTypeName := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfResourceNode := "betterado_workitemtrackingprocess_workitemtype.test"
	tfDataNode := "data.betterado_workitemtrackingprocess_workitemtype.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkWorkItemTypeDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclDataSourceWorkItemType(workItemTypeName, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(tfDataNode, "id", tfResourceNode, "id"),
					resource.TestCheckResourceAttrPair(tfDataNode, "process_id", tfResourceNode, "process_id"),
					resource.TestCheckResourceAttrPair(tfDataNode, "reference_name", tfResourceNode, "id"),
					resource.TestCheckResourceAttrPair(tfDataNode, "name", tfResourceNode, "name"),
					resource.TestCheckResourceAttrPair(tfDataNode, "description", tfResourceNode, "description"),
					resource.TestCheckResourceAttrPair(tfDataNode, "color", tfResourceNode, "color"),
					resource.TestCheckResourceAttrPair(tfDataNode, "icon", tfResourceNode, "icon"),
					resource.TestCheckResourceAttrPair(tfDataNode, "is_enabled", tfResourceNode, "is_enabled"),
					resource.TestCheckResourceAttrPair(tfDataNode, "url", tfResourceNode, "url"),
					captureWorkItemTypeEvidence(tfResourceNode),
				),
			},
		},
	})
}

func hclDataSourceWorkItemType(workItemTypeName string, processName string) string {
	proc := process(processName)
	return fmt.Sprintf(`
%s

resource "betterado_workitemtrackingprocess_workitemtype" "test" {
  name        = "%s"
  process_id  = betterado_workitemtrackingprocess_process.test.id
  description = "Test work item type"
}

data "betterado_workitemtrackingprocess_workitemtype" "test" {
  process_id     = betterado_workitemtrackingprocess_workitemtype.test.process_id
  reference_name = betterado_workitemtrackingprocess_workitemtype.test.reference_name
}
`, proc, workItemTypeName)
}
