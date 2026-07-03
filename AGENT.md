# Agent Memory — WI-6

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (complete — all 4 ACs done)

1. **Created example HCL files:**
   - `examples/resources/betterado_servicehook_storage_queue_pipelines/resource.tf` — minimal resource block with `stage_state_changed_event {}`
   - `examples/resources/betterado_servicehook_webhook_tfs/resource.tf` — minimal resource block with `git_push {}`

2. **Ran `make docs`** — tfplugindocs regenerated both resource doc files with embedded examples and framework schema descriptions. `git checkout -- docs/guides/` is baked into the Makefile target so hand-written guides are automatically restored.

3. **Verified `make terrafmt-check` passes** — the script only inspects `_test.go` files.

4. **Updated CHANGELOG.md** — added `### Changed` entries under `## [Unreleased]` for both resource migrations.

5. **Bumped PROVIDER_VERSION.txt** — `1.2.0` → `1.2.1`.

6. **Quality gate passed:** `go test -tags all -count=1 -run TestProvider_HasChildResources ./azuredevops/` → `ok 0.005s`

7. **Committed** as `a83d26e9 docs(servicehook): regenerate docs, add examples, update changelog and version`

## What worked

- `make docs` handles everything: generates schema-annotated docs, embeds the `examples/resources/<resource>/resource.tf` content.
- tfplugindocs names resource doc files by stripping the provider prefix: `betterado_servicehook_storage_queue_pipelines` → `docs/resources/servicehook_storage_queue_pipelines.md`.
- `make terrafmt-check` only checks `_test.go` files, not `examples/` — example HCL just needs to be valid.

## What didn't work

_(nothing failed)_

## Open questions

_(none)_

## Notes for reflection

_(none — straightforward finalization WI)_
