//go:build (all || data_source_workitemtrackingprocess_workitemtypes) && !exclude_data_source_workitemtrackingprocess_workitemtypes

package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccWorkitemtrackingprocessWorkItemTypes_DataSource_List(t *testing.T) {
	workItemTypeName1 := testutils.GenerateWorkItemTypeName()
	workItemTypeName2 := testutils.GenerateWorkItemTypeName()
	processName := testutils.GenerateResourceName()
	tfDataNode := "data.betterado_workitemtrackingprocess_workitemtypes.test"
	tfResourceNode1 := "betterado_workitemtrackingprocess_workitemtype.test1"
	tfResourceNode2 := "betterado_workitemtrackingprocess_workitemtype.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkWorkItemTypeDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclDataSourceWorkItemTypes(workItemTypeName1, workItemTypeName2, processName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfDataNode, "id"),
					resource.TestCheckResourceAttrSet(tfDataNode, "process_id"),
					// It might contain an unknown amount of inherited work item types
					testutils.TestCheckAttrGreaterThan(tfDataNode, "work_item_types.#", 1),
					resource.TestCheckTypeSetElemNestedAttrs(tfDataNode, "work_item_types.*", map[string]string{
						"name":        workItemTypeName1,
						"description": "Test work item type 1",
					}),
					resource.TestCheckTypeSetElemAttrPair(tfDataNode, "work_item_types.*.reference_name", tfResourceNode1, "reference_name"),
					resource.TestCheckTypeSetElemNestedAttrs(tfDataNode, "work_item_types.*", map[string]string{
						"name":        workItemTypeName2,
						"description": "Test work item type 2",
					}),
					resource.TestCheckTypeSetElemAttrPair(tfDataNode, "work_item_types.*.reference_name", tfResourceNode2, "reference_name"),
				),
			},
		},
	})
}

func hclDataSourceWorkItemTypes(workItemTypeName1 string, workItemTypeName2 string, processName string) string {
	proc := process(processName)
	return fmt.Sprintf(`
%s

resource "betterado_workitemtrackingprocess_workitemtype" "test1" {
  name        = "%s"
  process_id  = betterado_workitemtrackingprocess_process.test.id
  description = "Test work item type 1"
}

resource "betterado_workitemtrackingprocess_workitemtype" "test2" {
  name        = "%s"
  process_id  = betterado_workitemtrackingprocess_process.test.id
  description = "Test work item type 2"
}

data "betterado_workitemtrackingprocess_workitemtypes" "test" {
  process_id = betterado_workitemtrackingprocess_process.test.id
  depends_on = [
    betterado_workitemtrackingprocess_workitemtype.test1,
    betterado_workitemtrackingprocess_workitemtype.test2
  ]
}
`, proc, workItemTypeName1, workItemTypeName2)
}
