---
title: data.betterado_release_folder — resource/data-source parity gap closed
description: The release_folder data source mirrors data_release_definition exactly — non-Context Read, reused flattenReleaseFolder, no new SDK methods — confirming the parity pattern is repeatable with minimal cost.
category: pattern
project: terraform-provider-betterado
created_at: 2026-06-07T03:25:00Z
updated_at: 2026-06-07T03:25:00Z
related_themes:
  - 2026-06-06-data-source-split-read-only-pattern
  - 2026-06-06-provider-count-test-maintenance-trap
---

# data.betterado_release_folder — resource/data-source parity gap closed

## Pattern

When a betterado resource already has a working read path (e.g., `GetFolders`), adding the matching data source costs very little:

1. Create `data_<name>.go` — non-Context Read function, plain `*schema.Resource`, reuse existing flatten helper. No new SDK methods.
2. Register in `provider.go` DataSourcesMap + update `provider_test.go` count assertion (mandatory pair — see `2026-06-06-provider-count-test-maintenance-trap`).
3. Add 2 unit tests (happy path + not-found), 1 acceptance test (resource → data source round-trip + idempotency re-plan).
4. Add example HCL + docs page section.

`data_release_folder.go` is 58 lines. `data_release_folder_test.go` is 88 lines. The acceptance test is 66 lines. Total cost for the parity gap: $1.74 (resume of prior impl; fresh would be ~$8).

## Gate pattern (confirmed for data sources)

Offline unit gate:
```bash
go test -mod=vendor -tags all -count=1 -run TestDataReleaseFolder ./azuredevops/internal/service/release/
```

Live acceptance gate (WI-2):
```bash
go test -mod=vendor -tags all -count=1 -v -run TestAccDataReleaseFolder ./azuredevops/internal/acceptancetests/
```

Both follow the established betterado gate pattern: `-run <ExactPrefix>` on a file that doesn't exist yet → gate fails at iter-0 (no-work indicator) → after implementation → gate passes.

## Acceptance test shape (idempotency required)

```go
resource.ParallelTest(t, resource.TestCase{
    Steps: []resource.TestStep{
        {Config: hcl, Check: resource.ComposeTestCheckFunc(...)},
        {Config: hcl, PlanOnly: true, ExpectNonEmptyPlan: false}, // idempotency
    },
})
```

The HCL creates the folder via resource and reads it back via data source in the same config — no pre-existing ADO state required.

## Sources

- `_logs/2026-06-07T03-20-11_INIT-2026-06-07-release-folder-data-source/artifacts/DEMO.md`
- `brain/cycles/_raw/2026-06-07T03-20-11_INIT-2026-06-07-release-folder-data-source.md`
