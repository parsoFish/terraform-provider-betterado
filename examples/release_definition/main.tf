# Example: Basic Release Definition
#
# This example creates a release definition with a single Dev environment
# using the betterado provider.

terraform {
  required_providers {
    betterado = {
      source  = "local/betterado"
      version = "~> 0.1"
    }
    azuredevops = {
      source  = "microsoft/azuredevops"
      version = "~> 1.0"
    }
  }
}

# Use the official provider for core resources
provider "azuredevops" {
  org_service_url       = var.ado_org_url
  personal_access_token = var.ado_pat
}

# Use betterado for release pipeline resources
provider "betterado" {
  org_service_url       = var.ado_org_url
  personal_access_token = var.ado_pat
}

variable "ado_org_url" {
  description = "Azure DevOps organization URL"
  type        = string
}

variable "ado_pat" {
  description = "Personal Access Token"
  type        = string
  sensitive   = true
}

variable "project_name" {
  description = "Azure DevOps project name"
  type        = string
}

# Reference existing project
data "azuredevops_project" "example" {
  name = var.project_name
}

# Create a classic release definition
resource "betterado_release_definition" "example" {
  project_id          = data.azuredevops_project.example.id
  name                = "My Release Pipeline"
  description         = "Managed by Terraform"
  release_name_format = "Release-$(rev:r)"

  # Dev environment
  environment {
    name = "Dev"
    rank = 1

    pre_deploy_approvals {
      approval {
        is_automated = true
      }
    }

    post_deploy_approvals {
      approval {
        is_automated = true
      }
    }

    deploy_phase {
      name       = "Agent job"
      phase_type = "agentBasedDeployment"
      rank       = 1

      deployment_input {
        queue_id = 1
      }

      # Example task: Azure CLI
      # workflow_task {
      #   task_id  = "46e4be58-730b-4389-8a2f-ea10b3e5e815"
      #   version  = "2.*"
      #   name     = "Azure CLI"
      #   enabled  = true
      #   inputs = {
      #     scriptType     = "bash"
      #     scriptLocation = "inlineScript"
      #     inlineScript   = "echo 'Hello from release pipeline'"
      #   }
      # }
    }

    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }
  }

  # Staging environment (depends on Dev)
  environment {
    name = "Staging"
    rank = 2

    pre_deploy_approvals {
      approval {
        is_automated = false
        # approver_id = "user-guid-here"
      }
    }

    post_deploy_approvals {
      approval {
        is_automated = true
      }
    }

    deploy_phase {
      name       = "Agent job"
      phase_type = "agentBasedDeployment"
      rank       = 1

      deployment_input {
        queue_id = 1
      }
    }

    retention_policy {
      days_to_keep     = 60
      releases_to_keep = 5
      retain_build     = true
    }
  }

  # Variables shared across all environments
  variable {
    name  = "Environment"
    value = "default"
  }

  tags = ["managed-by-terraform", "classic-pipeline"]
}

output "release_definition_id" {
  value = betterado_release_definition.example.id
}

output "release_definition_url" {
  value = betterado_release_definition.example.url
}
