---
title: Docs-only WIs inherit a test gate from sibling code WIs — gate passes but verifies the wrong work
description: WIs that create only docs/examples files inherit the quality_gate_cmd from sibling code WIs; the gate fires required-paths-missing at iter-0 then passes at iter-1 because sibling tests are still green — but it does not verify the docs themselves.
category: antipattern
project: terraform-provider-betterado
created_at: 2026-06-06T05:10:00Z
updated_at: 2026-06-06T05:10:00Z
related_themes:
  - 2026-05-23-dogfood-cycle-false-pass-gate
  - 2026-06-02-ci-green-gate-design
---

# Docs-only WIs inherit a test gate from sibling code WIs

## Antipattern

WI-4 of INIT-2026-06-05-release-data-sources created:
- `docs/resources/release_definition.md`
- `docs/resources/release_definitions.md`
- `examples/data-sources/betterado_release_definition/main.tf`
- `examples/data-sources/betterado_release_definitions/main.tf`

Its `quality_gate_cmd` was:
```
go test -mod=vendor -tags all -count=1 -run TestDataReleaseDefinition|TestDataReleaseDefinitions ./azuredevops/internal/service/release/
```

This gate runs tests from WI-1 and WI-2 — not from WI-4's `files_in_scope`. At iter-0 the gate fires `required-paths-missing` (the doc files listed in `creates:` don't exist yet). At iter-1, after the docs are created, the gate passes because WI-1/WI-2 tests are still green. **The gate is proving WI-1/WI-2 work, not WI-4 work.**

## Risk

- If WI-4's docs are malformed, the gate doesn't catch it.
- If the examples contain invalid HCL, the gate doesn't catch it.
- A docs-only WI silently passes with a passing code gate even if docs were never written.

## Better gate for docs-only WIs

For a WI whose `files_in_scope` are only docs/examples, the quality gate should verify existence + HCL syntax, not a test suite:

```bash
# Verify files exist and HCL parses
test -f docs/resources/release_definition.md && \
test -f examples/data-sources/betterado_release_definition/main.tf && \
terraform fmt -check examples/data-sources/betterado_release_definition/main.tf
```

Or scope the gate to `make terrafmt-check` so the PM's gate spec stays simple.

## Implication for PM

When decomposing docs WIs, the PM should assign a structural gate (file existence + `terrafmt-check`) not a Go unit test gate. The initiative manifest's "Notes for PM" already specifies the default Go gate — PM should recognise docs WIs as exceptions.

## Sources

- `_logs/2026-06-06T04-41-44_INIT-2026-06-05-release-data-sources/events.jsonl` (event: `gate.expected-fail` WI-4 `required-paths-missing`)
- `_logs/2026-06-06T04-41-44_INIT-2026-06-05-release-data-sources/work-items-snapshot/WI-4.md` (quality_gate_cmd + creates fields)
- `brain/cycles/_raw/2026-06-06T04-41-44_INIT-2026-06-05-release-data-sources.md`
