---
slug: council-constraints
project: terraform-provider-betterado
date_added: 2026-05-22T00:00:00.000Z
related_themes:
  - 2026-05-18-stack-and-test-layout.md
  - 2026-05-18-go-test-harness-demos.md
---

# terraform-provider-betterado — binding council constraints

Project-level constraints shared by every betterado initiative (substrate +
createable-surface). The architect references this theme from each manifest
rather than copy-pasting the same five bullets into all 20 initiatives.

## Quality gate (two tiers)

1. **Offline unit gate (default)** — creds-free, fast (~1s). For schema/unit WIs:
   ```
   go test -tags all -count=1 -run <NewPrefix> ./azuredevops/internal/service/<area>/...
   ```
   `go build -mod=vendor .` (entry point only — never `./...`) to verify compilation.
   Each new `betterado_*` registered in `azuredevops/provider.go`. A test pkg that
   compiles but asserts nothing is a FAIL.

2. **Live acceptance gate (for live-ADO-behaviour WIs only)** — slow (~30m), requires creds:
   ```
   go test -tags all -count=1 -run TestAcc<Name> -timeout 30m ./azuredevops/internal/acceptancetests/...
   ```
   Creds (`AZDO_ORG_SERVICE_URL` + `AZDO_PERSONAL_ACCESS_TOKEN`) are stored in
   gitignored `secrets.env` using these canonical names. PreCheck self-loads `secrets.env`
   via `loadSecretsEnv()` (commit 3de384f0) — agents must NOT manually export vars or
   re-map `ADO_PAT`. `TF_ACC=1` is provided by the forge serve env; do NOT prefix it in
   the gate command itself. The `acceptance_gate.requires_env` guard in `.forge/project.json`
   errors fast if any of the three vars are absent before the gate runs.
   Required for any WI whose ACs involve live ADO behaviour (resource/data-source CRUD,
   idempotency, field round-trips). Analysis/audit/docs/CI/unit-only WIs use the offline
   gate only.

## Per-resource test substrate

Each createable resource ships five mock unit tests:

1. expand ↔ flatten roundtrip
2. create API-error
3. read-404-clears-state
4. update-calls-SDK-with-args
5. delete API-error

Pattern mirrors upstream `resource_environment_test.go`.

## Docs

`docs/resources/<name>.md` (description, basic + complex example, argument
& attribute reference, import) + runnable `examples/<name>/`. Edit
`docs/resources/` + `examples/` only, never `website/`.

## Fixtures

Inline if <20 lines else `testdata/*.json`. Never hand-edit
`azdosdkmocks/` (regenerate + commit if an SDK signature changes).

## Additive & atomic

Absent config reproduces prior behaviour; a quality-gate failure marks the
initiative BLOCKED (no cascade to independents).

## PM scope-guard

terraform-provider-betterado is a large vendored Go monorepo (286+
`*_test.go`, a huge `vendor/`). PM plans work-items ONLY against:

- `azuredevops/internal/service/<area>/` for the initiative's area.
- `azuredevops/provider.go` to register each new `betterado_*` name.
- The matching mock `azdosdkmocks/<area>_sdk_mock.go` (read, never edit).
- ONE existing upstream `*_test.go` in a sibling package as gomock pattern.

PM does NOT `Glob`/scan `vendor/`, the repo root, `website/`, or `docs/`
trees. Brain-query is mandatory but bounded: 1–2 targeted queries, not
broad exploration.
