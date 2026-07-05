resource "betterado_test_configuration" "example" {
  project_id  = var.project_id
  name        = "Chrome on Windows"
  description = "Test configuration for Chrome browser on Windows"
  values = {
    "Browser"         = "Chrome"
    "Operating System" = "Windows 10"
  }
}
