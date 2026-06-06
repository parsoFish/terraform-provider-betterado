# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC3: GIVEN the environment template resource and test are present WHEN `make test` runs (no TF_ACC, offline) THEN gofmt passes, `go test ./...` (whole module) exits 0, `provider_test.go` resource-count includes `betterado_release_definition_environment_template`, and `golangci-lint run ./...` passes
- [ ] AC1: GIVEN live ADO credentials are present (TF_ACC=1, AZDO_PERSONAL_ACCESS_TOKEN, AZDO_ORG_SERVICE_URL) WHEN `TestAccReleaseDefinitionEnvironmentTemplate_Basic` runs a `terraform apply` creating a `betterado_release_definition_environment_template` resource THEN the apply succeeds, the provider Read confirms the template exists in ADO, a mandatory idempotency re-plan reports no changes (`ExpectNonEmptyPlan: false`), and `terraform destroy` removes the template cleanly
- [ ] AC2: GIVEN the template was created by the prior apply step WHEN the test imports the resource by its ADO ID THEN the imported state matches the original apply state (all computed attributes present)
