# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (COMPLETE — all ACs done in one iteration)

- `make fmt` — ran `gofmt -s -w` over all non-vendor *.go files → AC1 passed
- `make terrafmt` — ran `terrafmt fmt -f` over all *_test.go and website markdown → AC2 passed
- `golangci-lint run --fix ./...` — auto-fixed what it could, then remaining errors fixed manually:
  - `main.go`: removed `schema` import, simplified `ProviderFunc` lambda to direct function reference (`azuredevops.Provider`) — gocritic unlambda rule
  - `resource_release_definition.go:957`: added `//nolint:errcheck` to `uuid.Parse` call
  - `resource_release_definition.go:1037,1516-1530`: added `//nolint:staticcheck` to SA1019 deprecated field accesses (EmailRecipients, EmailNotificationType, SkipArtifactsDownload, TimeoutInMinutes, EnableAccessToken)
  - `resource_task_group.go:460`: added `//nolint:errcheck` to `uuid.Parse` call
  - `resource_task_group.go:617,630`: removed unused functions `importTaskGroupState` and `splitImportID` (also removed the now-unused `fmt` import)
- `provider_test.go`: `TestProvider_HasChildResources` expected 131 resources but provider had 132 — added `betterado_task_group` to expectedResources list
- Full quality gate `make test && golangci-lint run ./... && make terrafmt-check` → all green ✅

## What worked

- Run formatters first (`make fmt`, `make terrafmt`) before tackling lint — clears AC1+AC2 instantly
- `golangci-lint run --fix` auto-fixes some issues; check remaining output for manual fixes
- SA1019 staticcheck: use `//nolint:staticcheck` per WI instructions (not full refactor)
- errcheck: use `//nolint:errcheck` when uuid.Parse error is deliberately ignored (uuid comes from validated TF schema string)
- Dead code (unused funcs): just delete them — safer than adding `//nolint:unused`
- When provider test `TestProvider_HasChildResources` fails with count mismatch: diff the provider ResourcesMap against the expected list to find the missing entry

## What didn't work

_(none — single iteration success)_

## Open questions

_(none)_

## Notes for reflection

- The `tenv` deprecation warning from golangci-lint is harmless (not an error) — lint config references deprecated linter `tenv`, replaced by `usetesting`. Could be cleaned up in `.golangci.yml` but not required for green gate.
- `importTaskGroupState` + `splitImportID` were dead code — wired resource uses `schema.ImportStatePassthroughContext` instead. Removing them was cleaner than suppressing `unused` lint.
