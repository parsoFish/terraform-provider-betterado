data "betterado_project" "example" {
  name = "ExampleProject"
}

resource "betterado_wiki" "example" {
  project_id = data.betterado_project.example.id
  name       = "MyProjectWiki"
  type       = "projectWiki"
}

resource "betterado_wiki_page" "example" {
  project_id = data.betterado_project.example.id
  wiki_id    = betterado_wiki.example.id
  path       = "/MyPage"
  content    = "# My Page\n\nThis is the content of my wiki page."
}
