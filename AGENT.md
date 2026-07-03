# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (complete)

**Root cause of gate failure:** The gate ran `go test -tags all -run TestFrameworkProvider_HasExtensionInstallResource ./azuredevops/internal/provider/...` and got `[no tests to run]`. The test `TestFrameworkProvider_HasExtensionInstallResource` did not exist in `framework_provider_test.go`. The framework registration itself (`extensionmanagement.NewExtensionInstallResource`) was already committed by WI-3 in `framework_provider.go`.

**Fix applied:**
1. Added `TestFrameworkProvider_HasExtensionInstallResource` to `azuredevops/internal/provider/framework_provider_test.go` (after `TestFrameworkProvider_HasTaskGroupResource`). Build tag: `//go:build all || provider_framework`. Test pattern exactly matches `TestFrameworkProvider_HasReleaseFolderResource` from the same file.
2. Created `examples/resources/betterado_extension_install/resource.tf` with publisher_id + extension_id example.
3. Ran `make docs` → generated `docs/resources/extension_install.md`. The Makefile post-step automatically runs `git checkout -- docs/guides/` to restore hand-written guides.
4. Added CHANGELOG.md entry under `## [Unreleased]`.
5. Bumped PROVIDER_VERSION.txt from 1.2.0 → 1.3.0 (minor bump for new resource).

**Gate result:** `PASS TestFrameworkProvider_HasExtensionInstallResource` — test runs and passes.
**make test:** No failures.
**golangci-lint --new-from-rev=main:** 0 issues.
**make terrafmt-check:** Exit 0.

## What worked

- The extensionmanagement registration was already in `framework_provider.go` from WI-3. WI-4 only needed the test, docs, example, changelog, and version bump.
- The Makefile's `make docs` target includes `git checkout -- docs/guides/` automatically — no manual restoration needed.
- The `//go:build all || provider_framework` tag is correct; `-tags all` picks it up.

## What didn't work

_(none — resolved in single iteration)_

## Open questions

_(none)_

## Notes for reflection

- The gate's "no tests to run" error is a common failure mode when WI-2/WI-3 commits the resource implementation but WI-4 (which owns `framework_provider_test.go`) hasn't been run yet to add the provider-level test. The forge gate-tightening logic correctly rejects `ok ... 0.00s [no tests to run]` as a false pass.
- WI-4 is the correct owner of `framework_provider_test.go` changes — all prior WIs should implement only their service-package files.
