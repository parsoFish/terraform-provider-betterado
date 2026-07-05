resource "betterado_test_variable" "example" {
  project_id     = var.project_id
  name           = "Browser"
  description    = "The browser to run tests against"
  allowed_values = ["Chrome", "Firefox", "Edge"]
}
