# Agent Memory — WI-5

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (WI-5 iter 0) — COMPLETE

All four ACs satisfied in one iteration.

1. **AC1 — Docs:** Prior docs (pre-iter 0) reflected old SDKv2 schema with `timeouts`
   block. Created minimal example HCL stubs under
   `examples/resources/betterado_{user,group,service_principal}_entitlement/resource.tf`,
   then ran `make docs` (calls `tfplugindocs generate` + `git checkout -- docs/guides/`
   automatically). Regenerated docs now document the correct framework schema (no timeouts,
   proper attribute descriptions).

2. **AC2 — demo.json:** Created
   `forge/history/INIT-2026-07-01-migrate-framework-member-entitlement/demo/demo.json`
   with an `acceptance-resource` checkpoint carrying `liveEvidence.url` sourced from
   `.forge/live-evidence/acceptance-resource-group-entitlement.json` (captured by WI-3's
   `TestAccGroupEntitlement_Create` live run). URL:
   `https://dev.azure.com/davidgparsonson/_apis/memberentitlementmanagement/groupentitlements/7bece247-5904-4d44-b7e3-29f96502d9ae?api-version=7.1`

3. **AC3 — CHANGELOG:** Added `user_entitlement` and `group_entitlement` FEATURES entries
   under `[Unreleased]` (service_principal_entitlement entry was already present from WI-4).

4. **AC4 — PROVIDER_VERSION.txt:** Bumped `1.2.0` → `1.3.0` (minor bump; three resources
   migrated in this initiative).

5. **Gate:** `go test -tags all -run TestFrameworkProvider_HasUserEntitlementResource
   ./azuredevops/internal/provider/` PASS (0.004s). `make test` PASS.
   `golangci-lint run --new-from-rev=main ./azuredevops/...` 0 issues.
   `make terrafmt-check` PASS.

## What worked

- `make docs` handles both tfplugindocs regeneration AND `git checkout -- docs/guides/`
  in a single call (see GNUmakefile `docs:` target) — no need to restore guides separately.
- Live evidence from `.forge/live-evidence/` can be used directly in demo.json checkpoint
  `liveEvidence` field.

## What didn't work

_(none — straightforward housekeeping iteration)_

## Open questions

_(none)_

## Notes for reflection

- The GNUmakefile `docs:` target is self-contained for doc regeneration. Future WIs can
  just `make docs`.
- demo.json lives at `forge/history/<initiative-id>/demo/demo.json`.
- tfplugindocs requires `examples/resources/<resource>/resource.tf` stubs to embed
  example usage in the generated markdown. Without them, no Example Usage section appears.
