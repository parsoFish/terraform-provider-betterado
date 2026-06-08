---
title: requires_env fast-fail guard confirmed working — live-acc WI exits at iteration 0 with gate.errored
description: After the requires_env fix (listing TF_ACC + AZDO_ORG_SERVICE_URL + AZDO_PERSONAL_ACCESS_TOKEN), a live-acc WI with absent env now exits immediately at gate-start (0 iterations, $0 cost) rather than burning 5 iterations in test PreCheck.
category: pattern
project: terraform-provider-betterado
created_at: 2026-06-08T12:00:00Z
updated_at: 2026-06-08T12:00:00Z
related_themes:
  - 2026-06-07-live-acc-gate-failed-but-cycle-merged-unverified
  - 2026-06-05-live-acceptance-gate-for-acceptance-wis
---

# requires_env fast-fail guard — first confirmation

## What happened

WI-3 of `INIT-2026-06-08-release-definition-schema-audit` had gate command:
```
go test -tags all -v -count=1 -run TestAccReleaseDefinition_basic ./azuredevops/internal/acceptancetests/
```

With `AZDO_PERSONAL_ACCESS_TOKEN` absent from env, the guard fired before spawning any agent:

```
gate.errored — live-env-missing
[forge gate-errored] live-acceptance gate requires TF_ACC, AZDO_ORG_SERVICE_URL,
AZDO_PERSONAL_ACCESS_TOKEN to be set so the test actually runs against the live
system; unset ⇒ the runner SKIPS and the gate would FALSE-PASS ("ok … 0.00s").
```

Result: WI-3 status `failed`, 0 iterations, $0.00 cost.

## Contrast with prior behaviour

`INIT-2026-06-07-release-folder-data-source` (pre-fix): `TF_ACC` was set but `AZDO_PERSONAL_ACCESS_TOKEN` was not. Guard stayed silent; 5 dev-loop iterations ran and all `t.Fatal`-ed in PreCheck. $~1+ wasted.

## Key fields in `.forge/project.json`

```json
"acceptance_gate": {
  "match": "acceptancetests",
  "requires_env": ["TF_ACC", "AZDO_ORG_SERVICE_URL", "AZDO_PERSONAL_ACCESS_TOKEN"]
}
```

All three must be exported for a live-acc WI to run. If any is absent: `gate.errored` immediately.

## Implication for PR status

A `gate.errored` live-acc WI still allows the unifier to open a PR (forge-side gap, tracked). For documentation-only initiatives this is acceptable (no regression risk). For implementation initiatives: must run with full creds or the live AC is unverifiable.

## Sources

- `_logs/2026-06-08T11-00-43_INIT-2026-06-08-release-definition-schema-audit/events.jsonl` (event: `gate.errored` `EV_mq5433l7_5d1cpb1q`)
- `brain/cycles/_raw/2026-06-08T11-00-43_INIT-2026-06-08-release-definition-schema-audit.md`
- `projects/terraform-provider-betterado/brain/themes/2026-06-07-live-acc-gate-failed-but-cycle-merged-unverified.md` (root cause + fix)
