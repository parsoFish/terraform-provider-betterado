# Fix Plan

> Checklist for WI-8. Tick items as you complete them; add items as you discover sub-problems.

- [ ] AC1: GIVEN an exhaustive TestAccReleaseDefinition_complete whose HCL sets a NON-DEFAULT value for EVERY option of betterado_release_definition — a real agent pool (queue_id resolved from the test project, NOT 0), demands, skip_artifacts_download, enable_access_token, retention, pre/post approvals, pre/post deployment gates WITH a real gate task (WI-6), cd_artifact + schedule triggers, a multiConfiguration parallel phase, a runOnServer phase WHEN TF_ACC=1 go test -run TestAccReleaseDefinition_complete -timeout 30m runs against live ADO (creds in env) THEN it applies, the Check funcs confirm every option persisted via the provider read, the default ExpectNonEmptyPlan:false idempotency check passes (NO perpetual diff), and it destroys cleanly
- [ ] AC2: GIVEN the live acceptance gate WHEN the dev-loop runs this WI's quality_gate_cmd THEN the gate exits 0 (live round-trip + idempotency proven in-cycle)
