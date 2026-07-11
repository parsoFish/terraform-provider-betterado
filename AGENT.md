# Unifier Agent Memory — INIT-2026-07-10-framework-auth-parity

> Institutional memory across unifier-Ralph iterations. Read at the start of every iteration; updated at the end.

## What I tried

_(updated by each iteration — most recent at the top)_

### Iteration 1 (UWI-3 — terminal re-prep)

**State on entry:** No `last-gate-failure.md`. All 3 WIs complete. 9 commits on branch since main. Previous unifier pass (UWI-1 → UWI-2 review cycle) already authored demo.json + DEMO.md + pr-description.md and committed capture outputs. UWI-2 was a code-fix that addressed 5 review findings including a security-relevant `use_cli *bool` null/false distinction fix.

**Problem found:** Running the hollow gate `go test -tags all -count=1 -run TestAccAuthParity ./azuredevops/internal/acceptancetests/` **FAILED** with:
```
--- FAIL: TestAccAuthParity_CredentialConstruction/ClientCertificate
    aztfauth.NewCredential(ClientCertificate): unexpected error: sources must contain at least one TokenCredential
```
Root cause: UWI-2 added a `ClientCertificate` subtest to `TestAccAuthParity_CredentialConstruction` with an incorrect assumption that aztfauth accepts empty cert content at construction time. In fact, `buildClientCertificateCredOpt` calls `getClientCert()` eagerly, which returns `"no client certificate available"` for empty `ClientCertBase64` + `ClientCertPfxFile`. With no cred option added to the chain, `entrauth.NewCredential` fails with "sources must contain at least one TokenCredential". The test comment was wrong ("credential will fail at token acquisition, not at construction").

**Fix applied:** Removed `ClientCertificate` subtest from `TestAccAuthParity_CredentialConstruction`. Updated doc comment to document why it's excluded and where cert auth IS tested (auth_test.go stub-based approach). Test now passes with 6 subtests.

**Other changes this iteration:**
- Re-authored `demo.json`: updated essence with UWI-2 security fix mention; split broken `cat ... && cat ...` Protocol checkpoint (shell operators don't work in non-shell contexts, orchestrator saw stderr) into two clean checkpoints (`cat terraform-registry-manifest.json` and `cat PROVIDER_VERSION.txt`); updated AC17 evidence to reflect 6-subtest coverage.
- Re-rendered `DEMO.md` via `forge demo render`.
- Updated `.forge/pr-description.md` What/How sections to describe UWI-2 changes (*bool use_cli, ARM_AUXILIARY_TENANT_IDS trimming, parseBoolEnv warnings, revised CredentialConstruction coverage).
- Committed as `feat(INIT-2026-07-10-framework-auth-parity): unify and demo (UWI-3 terminal re-prep)` and pushed.

**Gates verified locally:**
- `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → ok
- `go test -tags all -count=1 -run TestAccAuthParity ./azuredevops/internal/acceptancetests/` → ok (was failing before fix)
- `go test -tags all -count=1 ./azuredevops/internal/provider/...` → ok (27/27 pass)
- `go vet -tags all ./azuredevops/internal/provider/...` → clean

## Notes for reflection

- **Compound commands in demo checkpoints fail:** The orchestrator's capture runs each `command` as a direct exec (not through bash shell), so `&&` and `;` operators don't work. Commands in demo.json must be single, directly executable invocations. Splitting into separate checkpoints is the correct pattern.
- **aztfauth.NewCredential cert construction is eager:** The ClientCertificate path calls `getClientCert()` at construction time, not lazily. Credential construction tests must supply real cert data or skip the cert path. The WI-3 implementation incorrectly assumed lazy validation.
- **UWI-2 introduced a test regression:** The per-WI dev phase introduced a broken subtest. The unifier re-prep phase (UWI-3) is the right place to catch and fix integration issues like this — the unifier runs the hollow gate first and fixes within scope.
