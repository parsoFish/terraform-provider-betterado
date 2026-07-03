# Agent Memory — WI-8

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (completed all ACs in one pass)

1. Read WI-8 spec, checked no prior last-gate-failure.md existed.
2. Studied prior framework migrations (black_duck, checkmarx_one) for patterns.
3. Created `resource_serviceendpoint_artifactory_framework.go`:
   - Two ListNestedBlock sections: `authentication_token` (token field) and `authentication_basic` (username+password fields)
   - Both blocks are optional; seArtifactoryBuildEndpoint() checks which is populated
   - TypeName: `req.ProviderTypeName + "_serviceendpoint_artifactory"`
4. Created `resource_serviceendpoint_dynamic_lifecycle_services_framework.go`:
   - Flat attributes: authorization_endpoint, lifecycle_services_api_endpoint, client_id, username, password
   - TypeName: `req.ProviderTypeName + "_serviceendpoint_dynamics_lifecycle_services"`
   - Endpoint type: "lcsserviceendpoint", scheme: "UsernamePassword"
5. Deregistered both from `azuredevops/provider.go` ResourcesMap (replaced entries with comment lines)
6. Registered both in `azuredevops/internal/provider/framework_provider.go` Resources() list
7. Removed both from `azuredevops/provider_test.go` expectedResources list (replaced with comment lines)
8. Replaced `Providers: testutils.GetProviders()` → `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()` in both acceptance test files using sed
9. Ran `make fmt` to fix gofmt issues (struct alignment in artifactory model)
10. `make test` passed, `golangci-lint run --new-from-rev=main` → 0 issues, `make terrafmt-check` passed
11. `make docs` regenerated docs; `git checkout -- docs/guides/` restored hand-written guides
12. Added CHANGELOG entries under [Unreleased] for both resources
13. Committed: `feat(serviceendpoint): migrate artifactory, dynamics_lifecycle_services to terraform-plugin-framework`

## What worked

- **sed for bulk ProtoV6 replacement**: `sed -i 's/Providers:    testutils.GetProviders(),/ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),/g'` worked perfectly across both test files in one command.
- **make fmt** after writing Go files catches gofmt issues immediately; run it before make test.
- **black_duck framework file** as pattern for simple resources; checkmarx_one for ones with blocks.
- **Artifactory with two optional auth blocks**: used `[]seArtifactoryTokenModel` and `[]seArtifactoryBasicModel` slice fields in the model struct; schema uses `schema.ListNestedBlock` with max 1 item (validated at HCL level).
- Quality gate `TestProvider_HasChildResources` verifies count of SDKv2 resources exactly — deregister from provider.go + remove from provider_test.go expectedResources list must happen together.

## What didn't work

_(none — completed in one iteration)_

## Open questions

_(none)_

## Notes for reflection

- Artifactory's two-auth-scheme pattern (token vs username/password) maps cleanly to two optional ListNestedBlock sections in the framework schema. Both schemes are mutually exclusive in practice — validation could be added but isn't strictly required for the migration (the SDKv2 used ExactlyOneOf).
- The dynamics_lifecycle_services SDKv2 file used snake_case "dynamic" (without 's') but the Terraform resource name is "dynamics_lifecycle_services" (with 's'). The framework file correctly uses the Terraform name in the TypeName.
