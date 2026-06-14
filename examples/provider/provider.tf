terraform {
  required_providers {
    betterado = {
      source  = "parsoFish/betterado"
      version = "~> 0.1.0"
    }
  }
}

# Authentication uses a Personal Access Token. The token is read from the
# AZDO_PERSONAL_ACCESS_TOKEN environment variable when not set inline.
provider "betterado" {
  org_service_url       = "https://dev.azure.com/my-organization"
  personal_access_token = var.ado_pat # or export AZDO_PERSONAL_ACCESS_TOKEN
}

variable "ado_pat" {
  description = "Azure DevOps Personal Access Token"
  type        = string
  sensitive   = true
}
