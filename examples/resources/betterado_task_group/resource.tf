# A reusable task group referenced from release/build definitions as a
# definition_type = "metaTask" workflow task.
resource "betterado_task_group" "example" {
  project_id           = var.project_id
  name                 = "deploy-webapp"
  friendly_name        = "Deploy Web App"
  description          = "Reusable deployment steps for the web application"
  category             = "Deploy"
  author               = "platform-team"
  instance_name_format = "Deploy $(environment)"
  icon_url             = "https://cdn.vsassets.io/v/someicon.png"

  version = [{
    major = 1
    minor = 0
    patch = 0
  }]

  # Parameterized input surfaced to consumers of the task group.
  input = [{
    name          = "environment"
    label         = "Target environment"
    type          = "string"
    default_value = "staging"
    required      = true
    help_markdown = "The environment slot to deploy to."
    visible_rule  = "targetType = filePath"
    properties    = { "EndpointId" = "" }
    aliases       = ["targetEnvAlias"]
  }]

  # Task steps executed when the group runs.
  task = [{
    display_name = "Run deploy script"
    task_id      = "d9bafed4-0b18-4f58-968d-86655b4d2ce9" # CmdLine@2
    task_version = "2.*"

    inputs = {
      script = "echo Deploying to $(environment)"
    }
  }]
}

variable "project_id" {
  type = string
}
