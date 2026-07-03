# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the ADO Graph REST API v7.1 and the current SDKv2 schema in azuredevops/internal/service/graph/ WHEN docs/graph-gap-matrix.md is read THEN every field returned by the ADO Graph API (groups, users, service principals, storage keys, descriptors, memberships) is listed with coverage status (supported/gap/deferred)
- [x] AC2: GIVEN the ADO Identity REST API v7.1 and the current SDKv2 schema in azuredevops/internal/service/identity/ WHEN docs/identity-gap-matrix.md is read THEN every field returned by the ADO Identity API (identity groups, identity users) is listed with coverage status (supported/gap/deferred)
- [x] AC3: GIVEN writable gaps are identified WHEN the gap matrix is reviewed THEN each writable gap is marked either 'implement in this initiative' or 'deferred' with a rationale

## Completion notes

- `docs/graph-gap-matrix.md`: covers all 10 resources/data-sources (betterado_group resource, betterado_group data source, betterado_descriptor, betterado_storage_key, betterado_group_membership resource+data source, betterado_groups, betterado_service_principal, betterado_user, betterado_users). 63 supported / 14 gaps / 29 deferred.
- `docs/identity-gap-matrix.md`: covers all 3 data sources (betterado_identity_group, betterado_identity_groups, betterado_identity_user). 15 supported / 7 gaps / 30 deferred.
- All writable gaps are read-only computed fields; each is marked 'implement in this initiative' or 'deferred' with rationale (AC3).
- `TestAccGraphGapMatrix` test added in `azuredevops/internal/acceptancetests/graph_gap_matrix_test.go` — runs without TF_ACC, verifies document presence and required sections. Gate command passes.
