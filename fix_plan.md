# Fix Plan

> Checklist for WI-5. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN all policy and checks resources have been migrated to framework (WI-2/3/4 done) WHEN make docs runs and docs/guides/ is restored THEN docs/resources/ and docs/data-sources/ reflect the current provider schema; docs/guides/ hand-written guides are intact
- [x] AC2: GIVEN CHANGELOG.md is inspected WHEN looking for the Unreleased section THEN a '## Unreleased' entry describes the policy/approvalsandchecks framework migration with the resource names
- [x] AC3: GIVEN PROVIDER_VERSION.txt is inspected WHEN compared to main branch THEN the semver patch (or minor) version has been bumped
- [x] AC4: GIVEN the forge demo evidence is inspected WHEN checking forge/history/INIT-2026-07-01-migrate-framework-policy-branch/demo/ THEN demo.json contains a checkpoint with liveEvidence.url pointing to a real ADO Policy or Checks REST GET response
