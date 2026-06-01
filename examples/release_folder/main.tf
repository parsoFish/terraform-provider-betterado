terraform {
  required_providers {
    betterado = {
      source = "parsoFish/betterado"
    }
  }
}

provider "betterado" {
  org_service_url = "https://dev.azure.com/myorg"
}

resource "betterado_project" "example" {
  name               = "Example Project"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_release_folder" "example" {
  project_id = betterado_project.example.id
  path       = "\\MyFolder"
}

resource "betterado_release_folder" "nested" {
  project_id = betterado_project.example.id
  path       = "\\MyFolder\\Nested"
}
