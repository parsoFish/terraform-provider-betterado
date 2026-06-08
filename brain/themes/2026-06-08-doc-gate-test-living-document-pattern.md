---
title: Living-document gate test — Go test asserting doc file exists + minimum line count
description: A lightweight Go test file in the service package asserts that a documentation file exists in the repo root and has ≥N non-empty lines; runs in the standard go test suite without TF_ACC, permanently binding the doc to the code.
category: pattern
project: terraform-provider-betterado
created_at: 2026-06-08T12:00:00Z
updated_at: 2026-06-08T12:00:00Z
related_themes:
  - 2026-06-02-ci-green-gate-design
  - 2026-06-06-acceptance-test-compile-only-gate
---

# Living-document gate test

## Pattern

For documentation-only initiatives (gap matrices, roadmaps, ADR files) that must stay tracked with the code, place a small Go test file in the relevant service package:

```go
// azuredevops/internal/service/release/doc_audit_test.go
func TestAuditGapMatrixDocExists(t *testing.T) {
    _, thisFile, _, _ := runtime.Caller(0)
    repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", ".."))
    docPath := filepath.Join(repoRoot, "docs", "release-definition-gap-matrix.md")
    // stat + count non-empty lines; fail if < 50
}
```

Key properties:
- Uses `runtime.Caller(0)` to locate repo root — survives worktree / CI path variations.
- Counts *non-empty lines* (not bytes) — catches trivially-empty files created as placeholders.
- Runs in `go test ./azuredevops/internal/service/release/...` with no `TF_ACC`, no network, ≈4ms.
- Gate-tightener compatible: the test function name matches the `quality_gate_cmd` regex, so the gate sees real execution (not `[no tests to run]`).

## Applied in

`INIT-2026-06-08-release-definition-schema-audit` — WI-1 wrote `TestAuditGapMatrixDocExists` (≥50 lines), WI-2 appended `TestAuditRoadmapDocExists` (≥20 lines) to the same file. Both gates passed in ≈4ms. File: `azuredevops/internal/service/release/doc_audit_test.go` (123 lines, 2 tests).

## When to use

- Any WI that produces a `docs/*.md` file as its primary deliverable.
- The threshold (`minLines`) should match the expected minimum content density; 50 for detailed tables, 20 for shorter roadmaps.

## Sources

- `_logs/2026-06-08T11-00-43_INIT-2026-06-08-release-definition-schema-audit/events.jsonl` (events: `gate.pass` WI-1 `EV_mq53yru0_4s0m1wt9`, `gate.pass` WI-2 `EV_mq543301_2kvs8j5x`)
- `brain/cycles/_raw/2026-06-08T11-00-43_INIT-2026-06-08-release-definition-schema-audit.md`
