data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_pipeline" "example" {
  project_id  = data.betterado_project.example.id
  pipeline_id = 42
}

output "pipeline_name" {
  value = data.betterado_pipeline.example.name
}
