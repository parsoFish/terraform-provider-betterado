resource "betterado_workitemtrackingprocess_inherited_page" "example" {
  process_id        = betterado_workitemtrackingprocess_process.example.id
  work_item_type_id = "Microsoft.VSTS.WorkItemTypes.Bug" # inherited work item type reference name
  page_id           = "details-page-id"                  # ID of the inherited page to manage
  label             = "Details"
}
