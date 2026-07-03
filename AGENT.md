# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 4 (this iteration)

**Symptom**: All three live acceptance test failures had the same pattern:
```
Error running post-test destroy, there may be dangling resources:
error reading work item type <name> after destroy: VS1640142: Work item type not found or you do not have permission in the process <id>
```

**Root cause**: `checkWorkItemTypeDestroyed` in the acceptance test calls `GetProcessWorkItemType` for each `betterado_workitemtrackingprocess_workitemtype` resource in state. After the parent process is deleted, calling `GetProcessWorkItemType` returns HTTP 400 with `VS1640142: Work item type not found...`. The existing `utils.ResponseWasNotFound` function only recognized HTTP 404 and HTTP 400 with `VS800075` or `VS402806` — it did NOT handle `VS1640142`.

**Fix applied**: Added `VS1640142` to the 400 Bad Request not-found codes in `ResponseWasNotFound` (`azuredevops/internal/utils/HttpResponse.go`). Also added a corresponding unit test in `HttpResponse_test.go`. All unit tests pass; build clean.

### Prior iterations (0–3)

- Framework resource (`resource_work_item_type_framework.go`) — fully implemented
- Framework data sources (`data_work_item_type_framework.go`, `data_work_item_types_framework.go`) — fully implemented
- SDKv2 files deleted, provider registrations updated
- Acceptance tests updated to use `ProtoV6ProviderFactories` / direct clients
- `captureWorkItemTypeEvidence` + `CaptureLiveEvidence` wired (AC4)

## What worked

- Pattern: Azure DevOps API returns HTTP 400 with vendor-specific codes (VS######) for "resource not found" when the parent process is already gone — not HTTP 404
- `ResponseWasNotFound` in `azuredevops/internal/utils/HttpResponse.go` is the single place to add new vendor error codes — adding `VS1640142` there fixes all callers (acceptance test destroy check AND resource refreshModel)

## What didn't work

- Prior iterations focused on the framework migration itself; the destroy check error only surfaces in live acceptance tests (not unit tests or offline runs)

## Open questions

- None blocking.

## Notes for reflection

- `VS1640142` is the AzDO error code for "work item type not found or no permission" (HTTP 400). Should be documented in the project brain alongside `VS800075` and `VS402806` as known pseudo-404 error codes.
