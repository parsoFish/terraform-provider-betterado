data "betterado_project" "example" {
  name = "MyProject"
}

data "betterado_feature_flag" "example" {
  feature_id  = "ms.vss-work.agile"
  scope_name  = "project"
  scope_value = data.betterado_project.example.id
}

output "feature_state" {
  value = data.betterado_feature_flag.example.state
}
