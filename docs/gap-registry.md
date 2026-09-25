# Gap Registry

Tracks implementation coverage for the `betterado` Terraform provider against the ADO API surface.

## Vocabulary

Canonical status tokens used throughout this registry:

- `covered` — implemented and acceptance-tested
- `gap-open` — missing; should be implemented
- `gap-deferred` — intentionally skipped; reason documented
- `out-of-scope` — non-declarative / imperative-only; will not be implemented

## Classification

Three classification values for resources in this provider:

- `betterado-net-new` — resources the fork ships that upstream never released
- `betterado-extended` — upstream resources the fork has extended with additional fields/capabilities
- `betterado-inherited` — resources inherited unchanged from upstream

## Area index

Stub table — WI-2 through WI-4b fill the data cells.

| Area | Classification | Resources / Data Sources | Gap-Open | Gap-Deferred |
|------|----------------|--------------------------|----------|--------------|
| release-definition | | | | |
| release-folder | | | | |
| release-definition-permissions | | | | |
| task-group | | | | |
| taskagent | | | | |
| approvalsandchecks | | | | |
| pipelinesapproval | | | | |
| pipelines-v2 | | | | |
| serviceendpoint | | | | |
| core | | | | |
| build | | | | |
| policy | | | | |
| git | | | | |
| feed | | | | |
| wiki | | | | |
| identity | | | | |
| graph | | | | |
| security | | | | |
| permissions | | | | |
| securityroles | | | | |
| memberentitlementmanagement | | | | |
| notification | | | | |
| servicehook | | | | |
| dashboard | | | | |
| extension | | | | |
| gallery-extensionmanagement | | | | |
| featuremanagement | | | | |
| workitemtracking | | | | |
| workitemtrackingprocess | | | | |
| accounts-profile | | | | |
| test | | | | |

## Priority backlog

Populated by WI-5.
