---
title: CI-fix initiatives must use the full CI gate, not a narrower proxy
description: When the entire purpose of an initiative is to make CI green, the WI quality_gate_cmd must be the verbatim CI gate — a narrower proxy (e.g. single-package go test) passes forge's gate while real CI stays red.
category: pattern
project: terraform-provider-betterado
created_at: 2026-06-02T09:47:24Z
updated_at: 2026-06-02T09:47:24Z
related_themes:
  - 2026-05-23-dogfood-cycle-false-pass-gate
  - 2026-05-31-forge-onboarding-findings
---

# CI-fix initiatives must use the full CI gate, not a narrower proxy

## Pattern

When an initiative's entire objective is "make CI pass", the `quality_gate_cmd` in WI specs MUST be the verbatim CI gate command — not a scoped subset.

For this project the correct CI gate is:

```bash
bash -c "make test && golangci-lint run ./... && make terrafmt-check"
```

This gate:
- Legitimately fails at iter-0 (CI is broken — that's the whole point).
- Proves green only once all three CI checks pass.
- Forge's hollow-iter0 guard (`gate.expected-fail`) handles the expected failure correctly.

## Why a proxy gate is wrong

A narrower gate (e.g. `go test -tags all -count=1 ./azuredevops/internal/service/release/...`) would:
1. Pass iter-0 green (the package had working tests before the formatting fixes).
2. Not exercise `golangci-lint` or `terrafmt-check` at all.
3. Allow forge to declare the WI complete while real GitHub CI stays red.

This is the gate-escape antipattern from [[2026-05-23-dogfood-cycle-false-pass-gate]].

## Observed in this cycle

INIT-2026-06-01-ci-green (WI-1): used `bash -c "make test && golangci-lint run ./... && make terrafmt-check"` verbatim. Ralph iter-0 produced `gate.expected-fail` as intended; iter-1 produced `gate.pass` after all four lint categories were fixed. PR #5 merged clean.

## Sources

- `_logs/2026-06-02T09-28-54_INIT-2026-06-01-ci-green/events.jsonl` (events: `gate.expected-fail` EV_mpwftbxo, `gate.pass` EV_mpwfzidr)
- `brain/cycles/_raw/2026-06-02T09-28-54_INIT-2026-06-01-ci-green.md`
- `_logs/2026-06-02T09-28-54_INIT-2026-06-01-ci-green/work-items-snapshot/WI-1.md` (quality_gate_cmd section)
