## Why

The v1.0.0 breaking change introduced array-of-objects syntax for `betterado_release_definition` (`stages`, `artifact`, `deploy_phase`, …) and `betterado_task_group` (`task`, `input`, `version`). Consumers landing on the Terraform registry page after the 1.0.0 release must see HCL examples that use the new syntax — otherwise they copy-paste block syntax that will fail with the new schema. Additionally, `docs/guides/` hand-written authentication guides were being silently deleted by `tfplugindocs` on every `make docs` run with no automated recovery, creating a documentation maintenance hazard. Finally, the roadmap's holistic migration section lacked the explicit list of phase-2 resource candidates and the identification of the terraform-plugin-mux scaffold as the extension point, leaving contributors without a clear next step.

## What

Files changed in this initiative (`git diff --name-only main...HEAD`):

- `GNUmakefile` — Added `git checkout -- docs/guides/` guard immediately after the `tfplugindocs` invocation in the `make docs` target, so hand-written guides survive doc regeneration.
- `azuredevops/internal/provider/framework_provider.go` — Fixed the framework provider `TypeName` to match the expected resource address in the mux.
- `azuredevops/internal/provider/framework_provider_test.go` — Updated test expectations to match the corrected `TypeName`.
- `roadmap.md` — Renamed the holistic migration section heading from `Future — holistic Plugin Framework migration` (em-dash) to `Future: holistic terraform-plugin-framework migration` (colon, exact AC text). Expanded body with Phase 1 completion notice, explicit Phase 2 candidate list (`betterado_release_folder`, `betterado_release_definition_permissions`, upstream-inherited resources), and a note identifying the `terraform-plugin-mux` scaffold from `INIT-2026-06-19-framework-state-upgraders` as the extension point for future migrations.

Pre-existing (not in this diff, already correct from prior initiatives):
- `docs/resources/release_definition.md` — array syntax already present (`stages = [{…}]`, `artifact = [{…}]`).
- `docs/resources/task_group.md` — array syntax already present (`task = [{…}]`, `input = [{…}]`, `version = [{…}]`).
- `examples/resources/betterado_release_definition/resource.tf` — array syntax already in use.
- `examples/resources/betterado_task_group/resource.tf` — array syntax already in use.

## How

1. **GNUmakefile guard**: The `docs` make target was amended to call `git checkout -- docs/guides/` after `tfplugindocs generate`. This is idempotent: if no guides exist (fresh checkout), git silently no-ops; if guides exist and were deleted by tfplugindocs, they are restored from HEAD.

2. **Framework provider TypeName fix**: `framework_provider.go` had a mismatched `TypeName` that caused the mux to fail routing requests to the framework resource implementations. Fixed to match the ADO resource address used in acceptance tests. The corresponding provider test was updated to assert the corrected value.

3. **Roadmap expansion**: Section heading changed from em-dash to colon to satisfy the grep-able AC requirement. Phase 1 completion (this cycle) and Phase 2 candidate enumeration added to give contributors an actionable next step. The mux scaffold (`INIT-2026-06-19-framework-state-upgraders`) is named as the extension point — new framework resources register in `framework_provider.go` `Resources()` / `DataSources()` without touching `main.go`.

4. **Quality gate**: `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` → all three packages green. `make terrafmt-check` → exit 0.
