resource "betterado_test_plan" "example" {
  project_id = var.project_id
  name       = "Sprint 1 Test Plan"
}

resource "betterado_test_suite" "example" {
  project_id      = var.project_id
  plan_id         = betterado_test_plan.example.id
  parent_suite_id = betterado_test_plan.example.root_suite_id
  name            = "API Tests"
  suite_type      = "staticTestSuite"
}
