data "betterado_project" "example" {
  name = "MyProject"
}

resource "betterado_feature_flag" "example" {
  feature_id  = "ms.vss-work.agile"
  scope_name  = "project"
  scope_value = data.betterado_project.example.id
  state       = "enabled"
}
