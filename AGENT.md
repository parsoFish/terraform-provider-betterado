# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1

**Goal:** Update acceptance tests from SDKv2 block HCL syntax to framework array HCL syntax; wire the muxed provider factory for acceptance tests.

**Changes made:**

1. **`azuredevops/internal/provider/framework_provider.go`** — Fixed the framework provider's `Configure()` (was a no-op). Now reads `AZDO_ORG_SERVICE_URL` + `AZDO_PERSONAL_ACCESS_TOKEN` from env vars, calls `client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)`, and stores the result as `resp.ResourceData`. This is how the `TaskGroupResource.Configure()` gets its `*client.AggregatedClient`.

2. **`azuredevops/internal/acceptancetests/testutils/mux_provider.go`** (new file) — Added `GetMuxedProviderFactories()` which returns a `map[string]func() (tfprotov6.ProviderServer, error)` for use in `resource.TestCase.ProtoV6ProviderFactories`. It builds a proto-v6 mux server: SDKv2 provider (upgraded via `tf5to6server.UpgradeServer`) + framework provider stub (via `providerserver.NewProtocol6`).

3. **`azuredevops/internal/acceptancetests/resource_task_group_test.go`** — Multiple changes:
   - `hclTaskGroupBasic`: replaced block syntax `version {}`, `input {}`, `task {}` with array syntax `version = [{}]`, `input = [{}]`, `task = [{}]`
   - `hclTaskGroupWithGapFields`: same block→array conversion
   - `TestAccTaskGroup_basic`: `Providers: testutils.GetProviders()` → `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`
   - `TestAccTaskGroup_withGapFields`: same swap
   - `checkTaskGroupDestroyed`: replaced `testutils.GetProvider().Meta().(*client.AggregatedClient)` with `getDirectClient()` (env-var-based client)
   - `captureTaskGroupEvidence`: same replacement
   - Added `getDirectClient()` helper function

**Why `GetProvider().Meta()` doesn't work with `ProtoV6ProviderFactories`:** The `testutils.GetProviders()` approach returns a singleton SDKv2 provider, whose `Meta` field is set during test execution. The `ProtoV6ProviderFactories` approach builds a fresh mux server per-factory-call — the SDKv2 provider inside the mux is NOT the singleton, so `testutils.GetProvider().Meta()` remains nil.

**Why the framework provider needs its own `Configure()`:** In the mux, both the SDKv2 and framework providers receive the Configure RPC. The SDKv2 handles it with the full HCL schema. The framework provider's schema is empty (`schema.Schema{}`), so it reads credentials from env vars directly. The `resp.ResourceData` value becomes `req.ProviderData` in each framework resource's `Configure()` call.

**Compile status:** All packages compile. `go build -tags all ./...` passes. Unit tests pass.

### Iteration 2

**Goal:** Standing acceptance criteria from WI-3 spec — docs, examples, changelog, lint.

**Changes made:**

1. **`examples/resources/betterado_task_group/resource.tf`** — Converted from SDKv2 block syntax to framework array HCL syntax (`version = [{...}]`, `input = [{...}]`, `task = [{...}]`).

2. **`docs/resources/task_group.md`** — Rewrote example and schema section using array HCL syntax and framework attribute terminology (replaced "Block List" SDKv2 terminology with "List of Object"). Added typed defaults documentation for optional task-step fields.

3. **`CHANGELOG.md`** — Added draft entry under `## Unreleased` describing `betterado_task_group` migration to terraform-plugin-framework with array-of-objects HCL syntax and perpetual-diff-free optional fields.

4. **`resource_task_group_framework.go`** — `gofumpt -w` fix (golangci-lint was reporting 1 gofumpt issue in the file; now 0 issues).

**Quality gate status:**
- `go build -tags all ./...` — PASS
- `go test -tags all -count=1 ./azuredevops/internal/service/taskagent/ ./azuredevops/internal/provider/` — PASS
- `golangci-lint run ./...` — 0 issues (after gofumpt fix)
- `terrafmt diff -c -q -f azuredevops/internal/acceptancetests/resource_task_group_test.go` — PASS (no diff)

**Remaining:** The live acceptance gate (TF_ACC=1) has not yet been run in this iteration — that requires live ADO credentials and will be validated by the orchestrator quality gate. All code-level work for WI-3 is complete.

## What worked

- `azuredevops.NewAuthProviderPAT(pat)` is the correct PAT auth provider constructor (used elsewhere in the codebase: sweeper_test.go, shared_fixtures.go, provider.go)
- `tf5to6server.UpgradeServer(ctx, func() tfprotov5.ProviderServer { return schema.NewGRPCProviderServer(sdkv2Provider) })` is the correct signature (matches main.go usage)
- Framework `ListNestedAttribute` requires array HCL syntax `attr = [{ field = value }]`, NOT block syntax `attr { field = value }`
- `resp.ResourceData = agg` in framework provider `Configure()` flows to `req.ProviderData` in resource `Configure()` calls
- `getDirectClient()` from env vars works regardless of whether the SDKv2 singleton is wired
- `gofumpt -w <file>` fixes gofumpt issues in-place (separate tool from gofmt; golangci-lint uses gofumpt)

## What didn't work

- `make test` runs `go test -v ./...` which includes all 600+ packages and takes too long without TF_ACC — use targeted package tests instead

## Open questions

- Will the live acceptance test expose any idempotency diffs? The framework resource has defaults on all optional task/input fields — if ADO returns different values than the defaults, the second plan will show a diff. The schema uses `Default:` values for omitted fields so this should be fine.
- The framework provider's `Configure()` only handles PAT auth. If ADO auth is via AAD tokens, the framework provider won't be able to build a client. This is acceptable for the current scope (PAT is the standard for acceptance tests) but worth noting for future AAD support.

### Iteration 3

**Goal:** Fix the live acceptance gate (TF_ACC=1) — was failing with "Invalid Provider Server Combination".

**Root cause:** `tf6muxserver.NewMuxServer` requires **both** muxed providers to expose **identical** schemas at the protocol level. The SDKv2 provider (upgraded to proto-v6 via `tf5to6server`) had 18 provider attributes (`org_service_url`, `personal_access_token`, `client_id`, etc.); the framework provider had `schema.Schema{}` (empty). The mux rejected this at plan load time.

**Fix:** Added all 18 SDKv2 provider attributes to the framework provider's `Schema()` method as matching framework attribute types:
- `schema.TypeString` → `schema.StringAttribute{}`
- `schema.TypeBool` → `schema.BoolAttribute{}`
- `schema.TypeList` with `Elem: schema.TypeString` → `schema.ListAttribute{ElementType: types.StringType}`

The `Sensitive: true` flag was preserved for the same fields as SDKv2 (`personal_access_token`, `client_certificate`, `client_certificate_password`, `client_secret`, `oidc_token`).

**Results:**
- `TestAccTaskGroup_basic`: PASS (23.67s) — create, read-back assertions, idempotency (no diff on plan-only step), destroy.
- `TestAccTaskGroup_withGapFields`: PASS (23.40s) — same cycle including gap fields.
- Live evidence captured to `.forge/live-evidence/acceptance-resource.json` (label `acceptance-resource` per project contract).

**Quality gates:**
- `go build -tags all ./...` — PASS
- `go test -tags all -count=1 ./azuredevops/internal/service/taskagent/ ./azuredevops/internal/provider/` — PASS
- `golangci-lint run ./azuredevops/internal/provider/...` — 0 issues
- `terrafmt diff -c -q -f resource_task_group_test.go` — PASS
- `TF_ACC=1 go test -tags all -run TestAccTaskGroup_basic` — PASS
- `TF_ACC=1 go test -tags all -run TestAccTaskGroup_withGapFields` — PASS

**All WI-3 ACs are complete.**

## What worked

- `azuredevops.NewAuthProviderPAT(pat)` is the correct PAT auth provider constructor (used elsewhere in the codebase: sweeper_test.go, shared_fixtures.go, provider.go)
- `tf5to6server.UpgradeServer(ctx, func() tfprotov5.ProviderServer { return schema.NewGRPCProviderServer(sdkv2Provider) })` is the correct signature (matches main.go usage)
- Framework `ListNestedAttribute` requires array HCL syntax `attr = [{ field = value }]`, NOT block syntax `attr { field = value }`
- `resp.ResourceData = agg` in framework provider `Configure()` flows to `req.ProviderData` in resource `Configure()` calls
- `getDirectClient()` from env vars works regardless of whether the SDKv2 singleton is wired
- `gofumpt -w <file>` fixes gofumpt issues in-place (separate tool from gofmt; golangci-lint uses gofumpt)
- **Mux schema parity**: `tf6muxserver` compares provider schemas at the tftypes level — framework and SDKv2 must have the same attribute names, types, and sensitivity flags. The framework schema `Schema()` must mirror the SDKv2 `Schema:` map exactly.

## What didn't work

- `make test` runs `go test -v ./...` which includes all 600+ packages and takes too long without TF_ACC — use targeted package tests instead
- Empty `schema.Schema{}` in framework provider breaks the mux — must mirror the SDKv2 schema.

## Open questions

- ~~Will the live acceptance test expose any idempotency diffs?~~ Resolved: No diffs. All optional fields with defaults are handled correctly.
- The framework provider's `Configure()` only handles PAT auth. If ADO auth is via AAD tokens, the framework provider won't be able to build a client. Acceptable for current scope.

## Notes for reflection

- The mux provider factory pattern (`GetMuxedProviderFactories`) should probably live in testutils as a general utility, not per-test. Done: it's in `testutils/mux_provider.go`.
- The framework provider needs a real `Configure()` implementation — the no-op stub was always going to be a problem for live tests. Fixed.
- **Key lesson:** When muxing SDKv2 + framework providers, the framework provider's `Schema()` must return an identical schema to the SDKv2 provider's schema. This is a non-obvious requirement of `tf6muxserver` — it validates schema identity at mux creation time during the test's pre-plan phase.
- All WI-3 ACs are now satisfied: array HCL syntax ✓, framework resource ✓, docs current ✓, examples updated ✓, CHANGELOG entry ✓, lint clean ✓, terrafmt clean ✓, live ADO gate passing ✓, idempotency ✓, evidence captured ✓.
