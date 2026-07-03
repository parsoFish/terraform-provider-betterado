# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the ADO Work Item Tracking Process REST API v7.1 schema WHEN compared against each SDKv2 resource schema in azuredevops/internal/service/workitemtrackingprocess/ THEN docs/workitemtrackingprocess-gap-matrix.md lists every API field for all 13 resources and 4 data sources; writable gaps are marked resolved or deferred with rationale
  - [x] docs/workitemtrackingprocess-gap-matrix.md created (commit 13c744b4)
- [x] Quality gate: go build -mod=vendor . passes
- [x] Live gate: TestAccWorkitemtrackingprocessProcess_CreateDisabled passes
  - Root cause: GetProcessByItsId has eventual consistency lag after EditProcess PATCH
  - Fix (iteration 2): use EditProcess PATCH response directly in Create & Update (commit 089fb845)
  - Removed unused getProcessByID helper (lint fix)
