# Environment-specific inputs. Copy terraform.tfvars.example → terraform.tfvars
# and fill in real values from your Azure DevOps org.

variable "ado_org_url" {
  description = "Azure DevOps organization URL, e.g. https://dev.azure.com/your-org"
  type        = string
}

variable "ado_pat" {
  description = "Personal Access Token with Release (read/write/manage) scopes"
  type        = string
  sensitive   = true
}

variable "project_id" {
  description = "Azure DevOps project ID (UUID) the demo resources are created in"
  type        = string
}

variable "project_name" {
  description = "Project name — used only to build human-friendly portal URLs in outputs"
  type        = string
}

variable "queue_id" {
  description = "Agent pool queue ID for deploy phases (e.g. the hosted 'Azure Pipelines' pool)"
  type        = number
}

variable "build_definition_id" {
  description = "An existing build definition ID that produces the release artifact"
  type        = number
}

variable "approver_id" {
  description = "UUID of a real user/group that approves the production stage"
  type        = string
}

variable "permission_principal_id" {
  description = "Identity descriptor (group/user) the permissions resource grants on the pipeline"
  type        = string
}
