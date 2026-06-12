---
title: Live ADO acceptance tests as a standalone WI gate
description: Splitting live acceptance test verification into its own WI lets ralph discover secrets.env independently and run real ADO tests (~25s each) as the WI quality gate.
category: pattern
created_at: 2026-06-11
updated_at: 2026-06-11
---

## Observation

In INIT-2026-06-08-release-data-sources-completion, WI-2's quality gate was:

```
go test -tags all -v -count=1 -run TestAccDataReleaseDefinitionRevision_Basic|TestAccDataReleaseDefinitionHistory_Basic ./azuredevops/internal/acceptancetests/
```

This ran real ADO acceptance tests (~25s each) against `https://dev.azure.com/davidgparsonson` using credentials from `secrets.env`. Both tests passed first try; zero code changes were needed in WI-2.

## Why it works

- WI-1 ships the implementation; WI-2's gate forces proof that the implementation works live.
- By isolating in a separate WI, ralph's context is fresh and focused on credential discovery (`source secrets.env`) without implementation noise.
- Gate-tightening still applies: `data_release_definition_revision_history_test.go` was in WI-1's required paths, so WI-2 only needed to demonstrate passing, not create files.

## When to apply

Any data source that reads from the ADO REST API should have a matching acceptance test WI. The gate command pattern: `source secrets.env && go test -tags all -v -count=1 -run <TestName> ./azuredevops/internal/acceptancetests/`.

## Sources

- `_logs/2026-06-08T11-43-56_INIT-2026-06-08-release-data-sources-completion/events.jsonl` — EV_mq9gdigl_kmezz4sg (WI-2 ralph.end, gate.pass), EV_mq9g9aig_52f1x4vu (gate.pass iteration 0)
- `/home/parso/forge/brain/cycles/_raw/2026-06-08T11-43-56_INIT-2026-06-08-release-data-sources-completion.md`
