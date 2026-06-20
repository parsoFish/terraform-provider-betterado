---
title: ADO org 1000-project limit blocks live acceptance test creates
description: The davidgparsonson ADO test org is at the 1000-project cap; any live test that creates a project via TF resource or direct API fails with HTTP 400. Workaround is using GetProjects(stateFilter=wellFormed, top=1) to resolve an existing project.
category: antipattern
created_at: 2026-06-20
updated_at: 2026-06-20
---

# ADO org 1000-project limit blocks live acceptance test creates

## What happened

WI-5 (`TestAccTaskGroupStateUpgradeSmoke`) initially used `resource "betterado_project" "smoke"` in the TF HCL fixture. First live run returned:

```
Error: creating project: Failed to add a project as this organization already has 1000 projects.
Please delete unused projects to reduce total project count to under 1000 or switch to another organization.
```

Both the TF resource and a direct `QueueCreateProject` REST call reproduce this. The ADO org `davidgparsonson` is at capacity.

## Fix applied

Rewrote `smokeResolveProject()` to call `CoreClient.GetProjects(ctx, core.GetProjectsArgs{StateFilter: &projectStateValues.WellFormed, Top: &one})` — returns the first wellFormed project name. HCL fixture changed to `data "betterado_project" "smoke"` referencing that name. Test passed in 5.51s on first attempt.

## Standing rule

Any live acceptance test in this project that needs a project context MUST use `data "betterado_project"` (existing project) or resolve via `GetProjects`. NEVER create a project in TF_ACC tests against this org.

## Sources

- `_logs/2026-06-20T04-10-33_INIT-2026-06-19-framework-state-upgraders/events.jsonl` (EV_mqlvggs7_v5gozlly gate.fail, EV_mqlvrcxv_37u4bcjs gate.pass)
- `brain/cycles/_raw/2026-06-20T04-10-33_INIT-2026-06-19-framework-state-upgraders.md`
