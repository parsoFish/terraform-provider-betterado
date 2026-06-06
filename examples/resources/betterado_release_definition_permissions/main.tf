data "betterado_project" "example" {
  name = "ExampleProject"
}

resource "betterado_release_definition" "example" {
  project_id = data.betterado_project.example.id
  name       = "ExamplePipeline"
}

resource "betterado_release_definition_permissions" "example" {
  project_id            = data.betterado_project.example.id
  release_definition_id = betterado_release_definition.example.id
  principal             = "ExampleGroup"

  permissions = {
    "ViewReleaseDefinition"    = "Allow"
    "EditReleaseDefinition"    = "Allow"
    "DeleteReleasePipeline"    = "Deny"
    "QueueRelease"             = "Allow"
    "EditReleaseEnvironment"   = "NotSet"
  }
}
