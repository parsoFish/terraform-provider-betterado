# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (this iteration — WI-3 completes)

- CHANGELOG.md already had `## [Unreleased]` with `### FEATURES` section documenting both dashboard and extension migrations (written by prior WI-1/WI-2 iterations).
- The WI spec required `### Changed` and `### Added` subsections in addition to the detailed FEATURES entries.
- Added `### Changed` entries for both migrations and `### Added` entries for the two gap-matrix docs files under `## [Unreleased]`, preserving the existing `### FEATURES` block.
- Bumped `PROVIDER_VERSION.txt` from `1.2.0` → `1.2.1` (patch increment).
- `make test` passed (gofmt clean, all non-acceptance tests passed).
- Committed as `docs(changelog): add WI-3 Unreleased entries for dashboard/extension migration; bump patch to 1.2.1`.

## What worked

- The CHANGELOG.md `## [Unreleased]` section already existed with full detail from WI-1/WI-2 work; only needed to prepend the canonical `### Changed` / `### Added` entries required by the WI spec.
- `make test` passes trivially for docs-only changes (no Go code changed).

## What didn't work

_(none — completed in first iteration)_

## Open questions

_(none)_

## Notes for reflection

_(none)_
