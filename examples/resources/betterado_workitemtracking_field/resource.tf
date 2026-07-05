resource "betterado_workitemtracking_field" "example" {
  name           = "My Custom Field"
  reference_name = "Custom.MyCustomField"
  type           = "string"
  description    = "A custom string field for work items."
  usage          = "workItem"
}
