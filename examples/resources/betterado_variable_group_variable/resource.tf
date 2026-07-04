# Add a plain-text variable to an existing variable group.
resource "betterado_variable_group_variable" "example" {
  project_id        = var.project_id
  variable_group_id = betterado_variable_group.example.id
  name              = "MY_VAR"
  value             = "hello"
}

# Add a secret variable to the same group.
resource "betterado_variable_group_variable" "secret_example" {
  project_id        = var.project_id
  variable_group_id = betterado_variable_group.example.id
  name              = "MY_SECRET"
  secret_value      = "super-secret"
}

resource "betterado_variable_group" "example" {
  project_id   = var.project_id
  name         = "my-variable-group"
  allow_access = true

  variable = [{
    name  = "seed"
    value = "seed"
  }]
}

variable "project_id" {
  type = string
}
