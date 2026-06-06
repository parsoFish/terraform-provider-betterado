resource "betterado_release_definition_environment_template" "example" {
  project_id  = "00000000-0000-0000-0000-000000000001"
  name        = "My Deploy Template"
  description = "Template for deploying to production."
  category    = "Deploy"
}
