# ── betterado_task_group ──────────────────────────────────────────────────────
# Reusable deploy steps, referenced from the release definition as a metaTask.
# Showcases the v0.2.0 input metadata: icon_url, visible_rule, properties, aliases.
resource "betterado_task_group" "deploy_steps" {
  project_id           = var.project_id
  name                 = "betterado-standing-deploy-steps"
  friendly_name        = "Standing Demo — Deploy Steps"
  description          = "Reusable deployment task group managed by Terraform (standing demo)"
  category             = "Deploy"
  author               = "platform-team"
  instance_name_format = "Deploy: $(target)"
  icon_url             = "https://cdn.vsassets.io/v/M183_20210324.1/_content/Header/default-avatar.png"

  version {
    major   = 1
    minor   = 0
    patch   = 0
    is_test = false
  }

  # Parameterised input — exercises every input-metadata field added in v0.2.0.
  input {
    name          = "target"
    label         = "Deployment Target"
    type          = "string"
    default_value = "staging"
    required      = true
    help_markdown = "The environment slot to deploy to."
    group_name    = "Targets"
    visible_rule  = "type = string"
    properties    = { "EndpointFilterRule" = "" }
    aliases       = ["targetEnv", "deployTarget"]
  }

  input {
    name          = "dryRun"
    label         = "Dry run"
    type          = "boolean"
    default_value = "false"
    help_markdown = "When true, print actions without applying them."
  }

  task {
    display_name = "Health check"
    task_id      = "e213ff0f-5d5c-4791-802d-52ea3e7be1f1" # PowerShell v2
    task_version = "2.*"
    inputs = {
      targetType = "inline"
      script     = "Write-Host \"Health check for $(target)\""
    }
  }

  task {
    display_name      = "Run deployment"
    task_id           = "6c731c3c-3c68-459a-a5c9-bde6e6595b5b" # Bash v3
    task_version      = "3.*"
    continue_on_error = true
    inputs = {
      targetType = "inline"
      script     = "echo 'Deploying to $(target) (dryRun=$(dryRun))'"
    }
  }
}
