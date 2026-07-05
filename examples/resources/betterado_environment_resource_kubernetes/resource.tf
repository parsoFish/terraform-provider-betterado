resource "betterado_environment_resource_kubernetes" "example" {
  project_id          = "00000000-0000-0000-0000-000000000000"
  environment_id      = 1
  service_endpoint_id = "00000000-0000-0000-0000-000000000001"
  name                = "my-k8s-resource"
  namespace           = "default"
  cluster_name        = "my-cluster"
  tags                = ["env:prod", "team:platform"]
}
