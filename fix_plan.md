# Fix Plan

> Checklist for WI-4. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN examples/resources/betterado_release_definition/resource.tf after update WHEN the file is inspected THEN it uses 'stages = [{ … }]' array syntax with no 'environment { }' blocks; the file is terrafmt-clean
- [x] AC2: GIVEN docs/resources/release_definition.md after update WHEN the file is inspected THEN all HCL code blocks use 'stages' and array syntax; no reference to 'environment' as the top-level pipeline-stage key remains
- [x] AC3: GIVEN make terrafmt-check run after the file updates WHEN executed against the updated files THEN exits 0 (no HCL formatting violations)
- [x] AC4: GIVEN make test run (gofmt + whole-module go test, no TF_ACC) after the examples/docs changes WHEN executed THEN exits 0 (the doc_audit_test.go audit test and provider_test.go resource-count test still pass)
