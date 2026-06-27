# terraform-provider-betterado — agent instructions

A GitHub fork of `microsoft/terraform-provider-azuredevops` that adds the
resources Microsoft never shipped — chiefly **classic release pipelines**
(`vsrm.dev.azure.com` Release API) and **task groups** — on top of the full
upstream provider. Go + Terraform Plugin SDK v2. Module
`github.com/parsoFish/terraform-provider-betterado`, resource prefix `betterado_`.

## Build, test, lint — exact invocations

```bash
# Build/test ONLY the package you're changing — never the whole tree.
go test -tags all -count=1 ./azuredevops/internal/service/release/...      # release pkg
go test -tags all -count=1 ./azuredevops/internal/service/taskagent/...    # task-group pkg
go build -mod=vendor .                 # compile the entry point (fast) — verify it builds

# Full CI-equivalent gate (mirrors GitHub; run before opening a PR):
make test && golangci-lint run ./... && make terrafmt-check
make fmt && make terrafmt               # auto-formatters (run before the gate)

# Live acceptance (needs creds — see below):
TF_ACC=1 go test -tags all -run TestAcc<Name> ./azuredevops/internal/acceptancetests/...

# Reap leaked test projects (from killed/timed-out runs that skipped cleanup):
make sweep            # deletes test-acc-* / AccTest* projects; real projects allowlisted
```

> **Test cleanup:** each acceptance test tears down its own resources via
> `t.Cleanup` / `CheckDestroy` on the normal path. A killed/timed-out/panicked run
> skips that, leaking the fixture project. `make sweep` (TF SDK sweeper, see
> `acceptancetests/sweeper_test.go`) is the backstop — run it after an aborted
> acceptance run. The package path MUST precede `-sweep` or `go test` ignores it.

> **Disk discipline:** `go build ./...` / `go vet ./...` (whole tree) generate
> multi-GB build cache and fill the drive. Build/vet only the package under change.
> `make clean-cache` reclaims space.

Live creds for acceptance/demo come from `secrets.env` (gitignored):
`AZDO_ORG_SERVICE_URL`, `AZDO_PERSONAL_ACCESS_TOKEN`, `TF_ACC=1`.

## How agents should work this project

- Build the implementation to satisfy the tests as written; write fixtures with
  **non-default** values and assert Create/Update via a **read-back** (apply →
  provider read → idempotency re-plan `ExpectNonEmptyPlan:false` → clean destroy).
- Commit your work; leave git history intact (no resets).
- Two gates must hold for every change: (1) a `TF_ACC` live-acceptance test
  against real ADO; (2) the CI-equivalent gate green. Both are injected as
  standing ACs from `.forge/project.json`.
- Record each initiative's plan + demo under `forge/history/<initiative-id>/`.

## Skills (project-action recipes, under `forge/skills/`)

- `forge/skills/resource-scaffolder/` — CRUD boilerplate + the schema-shape rules.
- `forge/skills/schema-refactor/` — rename a schema field + convert nested blocks to
  assignable list-of-object attributes (`ConfigMode: SchemaConfigModeAttr`).
- `forge/skills/ado-api-explorer/` — systematically map an ADO REST endpoint before implementing.
- `forge/skills/ado-browser-inspector/` — capture portal network traces for undocumented calls.
- `forge/skills/ado-demo/` — the live-evidence demo: apply → API GET round-trip → portal screenshot → destroy.

## Resource implementation pattern

`resource_<name>.go` returns `*schema.Resource`; Create/Read/Update/Delete with a
`Context` suffix; `expand*` (state→API) and `flatten*` (API→state) per nested
layer; acceptance tests in `azuredevops/internal/acceptancetests/`; register in
`azuredevops/provider.go`; add example configs under `examples/`.

Schema shapes: single nested object → `TypeList` + `MaxItems:1`; ordered list of
objects → `TypeList` (or a list-of-object attribute via `ConfigMode:
SchemaConfigModeAttr` for readability); unordered collection → `TypeSet`; simple
map → `TypeMap`.

## Key technical notes (gotchas paid for in prior cycles)

- **Two API hosts:** core `dev.azure.com`, release `vsrm.dev.azure.com`. The Go
  SDK routes by client (`release.Client` vs `TaskAgentClient`).
- **Stale-revision update returns HTTP 400, not 409** (`typeKey:
  InvalidRequestException`, "old copy of the release pipeline"). Update must
  detect this, re-read for the current revision, and retry once.
- **404 in Read** ⇒ `d.SetId("")` + return nil (external delete), never error.
- **Artifact `definition_reference`** returns extra API-computed keys not in user
  config; `flattenArtifacts` filters to user-set keys to avoid a perpetual diff.
- Release definitions are deeply nested (stages → deploy_phases → deployment_input
  + workflow_tasks; approvals; conditions; options; policies; artifacts;
  variables; variable_groups). Each expand/flatten handles one layer.

## Layout

- `azuredevops/` — provider source (no forge metadata).
- `docs/` — API references + per-resource gap matrices.
- `forge/` — committed forge artifacts: `forge/brain/` (project brain),
  `forge/skills/` (project-action skills), `forge/history/<initiative-id>/` (plan + demo per cycle).
- `roadmap.md` — the planning frontier. `.forge/project.json` — the forge contract config.

## Release — declared as `releaseProcess`, tagged by CI

A merged schema/feature change is not *delivered* until it's published, or
Terraform consumers can't use the new fields. This project declares the repo-side
prep as `releaseProcess` in `.forge/project.json`, so forge performs it in the
cycle; **tagging + publishing stay in CI** — forge never runs them.

The flow:

1. **In-cycle — docs current** (`releaseProcess` `docs` step): `make docs`
   (tfplugindocs regenerates `docs/resources/` + `docs/data-sources/` from the
   schema), then `git checkout -- docs/guides/` (tfplugindocs deletes the
   hand-written guides — restore them). Refresh the `examples/` the docs embed.
2. **In-cycle — draft changelog** (`releaseProcess` `changelog` step): add a
   DRAFT entry under `## Unreleased` in `CHANGELOG.md` (one bullet per
   user-visible change, categorised). The post-approval finaliser promotes it to
   a versioned section.
3. **Pre-merge — bump the version** (`releaseProcess` `version` step): edit
   `PROVIDER_VERSION.txt` (semver: minor for new features, patch for fixes).
4. **CI — tag + release** (forge ships, never runs): on merge to `main`,
   `.github/workflows/tag-on-changelog.yml` reads `PROVIDER_VERSION.txt` and
   pushes the `vX.Y.Z` tag; that tag push triggers `.github/workflows/release.yml`
   (GoReleaser, `.goreleaser.yml`), which builds the signed artifacts the
   Terraform Registry ingests.

## Fork workflow

`main` carries all betterado work. Pull upstream via `git fetch upstream && git
merge upstream/main` (expect conflicts: `microsoft/`→`parsoFish/` import renames +
`betterado_` resource prefixes). `upstream` = the Microsoft repo; `origin` =
`parsoFish/terraform-provider-betterado`.

## God files — third_party constraint (C3)

The `third_party/azure-devops-go-api/` directory contains generated code from upstream
(Microsoft's ADO Go SDK). Files in this tree exceed 800 LOC and some exceed 1600 LOC
(e.g., `workitemtracking/client.go`, `workitemtrackingprocess/models.go`). These are
excluded from structural refactoring and treated as opaque upstream dependencies.
Do not attempt to split or restructure them; they are managed externally.
