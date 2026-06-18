# Standing demo — provider + version pin.
#
# Published use: Terraform downloads parsoFish/betterado >= 0.2.0 from the
# registry. Local/dev use: a dev_overrides block in a CLI config file points the
# same source at a locally-built binary (see README.md) — skip `terraform init`
# and run `terraform validate` / `plan` / `apply` directly.
terraform {
  required_version = ">= 1.5"
  required_providers {
    betterado = {
      source  = "parsoFish/betterado"
      version = ">= 0.2.0" # every feature below landed in v0.2.0
    }
  }
}

provider "betterado" {
  # Credentials come from the environment (AZDO_ORG_SERVICE_URL,
  # AZDO_PERSONAL_ACCESS_TOKEN) or these explicit fields.
  org_service_url       = var.ado_org_url
  personal_access_token = var.ado_pat
}
