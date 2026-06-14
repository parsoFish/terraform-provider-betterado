
terraform {
  required_providers {
    betterado = {
      source = "parsoFish/betterado"
      version = ">=0.1.0"
    }
  }
}

resource "betterado_project" "project" {
  name               = "terraform-provider-betterado"
  description        = ""
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_serviceendpoint_github" "github_serviceendpoint" {
  project_id            = betterado_project.project.id
  service_endpoint_name = "GitHub Service Connection"

  auth_personal {
    # personal_access_token = "..." Or set with `AZDO_GITHUB_SERVICE_CONNECTION_PAT` env var
  }
}

resource "betterado_serviceendpoint_dockerregistry" "dockerregistry_serviceendpoint" {
  project_id            = betterado_project.project.id
  service_endpoint_name = "DockerRegistry Service Connection"

  # docker_username = "..." - Or set with `AZDO_DOCKERREGISTRY_SERVICE_CONNECTION_USERNAME` env var
  # docker_email    = "..." - Or set with `AZDO_DOCKERREGISTRY_SERVICE_CONNECTION_EMAIL` env var
  # docker_password = "..." - Or set with `AZDO_DOCKERREGISTRY_SERVICE_CONNECTION_PASSWORD` env var

}

resource "betterado_build_definition" "nightly_build" {
  project_id      = betterado_project.project.id
  agent_pool_name = "Azure Pipelines"
  name            = "Nightly Build"

  repository {
    repo_type             = "GitHub"
    repo_id               = "parsoFish/terraform-provider-betterado"
    branch_name           = "main"
    yml_path              = ".azdo/azure-pipeline-nightly.yml"
    service_connection_id = betterado_serviceendpoint_github.github_serviceendpoint.id
  }
}
