---
title: Vendor UnmarshalJSON patch required when ADO returns numeric enum the SDK struct expects as string
description: ADO REST can return a bitmask integer for fields the SDK declares as a string enum; absent a custom UnmarshalJSON the field silently zeroes on every read, causing a perpetual diff.
category: pattern
created_at: 2026-06-11
updated_at: 2026-06-11
---

## What happened

`ReleaseSchedule.DaysToRelease` — ADO REST API returns `daysToRelease` as a JSON integer (e.g. `62` for Mon–Fri bitmask). The SDK struct declares `ScheduleDays` as a `string` enum. `json.Unmarshal` fails silently: the integer cannot unmarshal into a string, so the field zeroes. On every plan, Terraform sees `schedule.days_to_release = 0` vs the HCL value → perpetual diff.

**Fix**: add `vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/release/schedule_unmarshal.go` with a custom `UnmarshalJSON` for `ReleaseSchedule` that reads the field as `json.Number`, converts integer bitmask to the `ScheduleDays` string the struct expects.

## Generalisation

This class of bug — numeric ADO response field + string SDK field — is distinct from the silent-discard antipattern (`2026-06-05-ado-silent-field-discard-idempotency.md`). The write succeeds; the read mis-decodes. Pattern for identifying it:

1. Acceptance test shows the field set correctly in ADO portal.
2. `go test` round-trip test passes (gomock returns what you put in).
3. Live plan still shows perpetual diff → the real ADO response has a different JSON type.

**When to apply**: any vendored SDK struct field where `go test` passes but live acceptance produces drift — check whether ADO returns a different JSON type than the struct declares.

**Tradeoff**: vendor patch is load-bearing; upstream SDK struct changes will conflict. Document vendor overrides in a known-overrides file (open question, see `user-questions.md`).

## Sources

- `_logs/2026-06-08T12-01-16_INIT-2026-06-08-release-definition-environment-config-surface/events.jsonl` — `gate.pass` WI-2 `iteration=0` 22.163s (live idempotency confirmed after patch)
- `/home/parso/forge/brain/cycles/_raw/2026-06-08T12-01-16_INIT-2026-06-08-release-definition-environment-config-surface.md`
- Related: `brain/themes/2026-06-05-ado-silent-field-discard-idempotency.md`
