# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the docs/ directory is inspected WHEN looking for a policy gap matrix THEN docs/policy-gap-matrix.md exists and lists every ADO Policy API type (branch + repository) with a coverage column indicating whether betterado exposes it
- [x] AC2: GIVEN the docs/ directory is inspected WHEN looking for an approvals-and-checks gap matrix THEN docs/approvalsandchecks-gap-matrix.md exists and lists every ADO Checks/Approvals API type with a coverage column indicating whether betterado exposes it

## Status: COMPLETE

Both files were created in commit `bc358c12` and are present in `git diff --name-only main...HEAD`.
WI-1 quality gate (`go test -tags all -run TestProvider_HasChildResources ./azuredevops/`) passes.

Note: `last-gate-failure.md` in .forge/ is from WI-4/5 live acceptance tests (TestAccCheck*) — NOT WI-1's gate.
