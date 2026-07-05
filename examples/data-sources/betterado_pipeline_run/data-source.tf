data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_pipeline_run" "example" {
  project_id  = data.betterado_project.example.id
  pipeline_id = 42
  run_id      = 1001
}

output "run_state" {
  value = data.betterado_pipeline_run.example.state
}

output "run_result" {
  value = data.betterado_pipeline_run.example.result
}
