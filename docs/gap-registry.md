# Gap registry — terraform-provider-betterado

> Canonical cross-area coverage registry. Generated from the 31 per-area gap matrices in `docs/`.
> Vocabulary defined below; all matrices use ONLY these four tokens.

## Vocabulary

| Token | Meaning |
|---|---|
| `covered` | Field/resource is implemented and acceptance-tested |
| `gap-open` | Field/resource is missing and should be implemented |
| `gap-deferred` | Intentionally skipped; reason documented in the per-area matrix |
| `out-of-scope` | Non-declarative (imperative/runtime-only); will not be implemented |

**Forbidden tokens** (must not appear in any matrix): `mapped`, `supported`, `implemented`, `partial`,
`missing`, `present`, `gap-resolved`, `✅`, `⚠️`, `🚫`, `read-only`, `gap`, `breaking-deferral`,
`missing-writable-gap`, `missing-computed-gap`, and any variant not in the four above.

### Token mapping rules

- `implemented` / `mapped` / `present` / `✅ Implemented` → `covered`
- `gap` / `missing-writable-gap` / `⚠️ Gap` → `gap-open`
- `deferred` / `breaking-deferral` (when intentionally skipped) → `gap-deferred`
- `read-only` (server-assigned, never user-configurable) → `out-of-scope`
- `missing-computed-gap` → `out-of-scope` (default); `gap-deferred` only if field could usefully be Computed
- `🚫 Out of scope` / `out-of-scope` → `out-of-scope`

## Coverage index

| Area | Classification | gap-open count | gap-deferred count | v7.1→v7.2 delta |
|---|---|---|---|---|
| release-definition | betterado-net-new | 0 | 3 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| release-folder | betterado-net-new | 0 | 0 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| release-definition-permissions | betterado-net-new | 0 | 0 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| task-group | betterado-net-new | 0 | 0 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| taskagent | betterado-inherited | 0 | 5 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| approvalsandchecks | betterado-inherited | 0 | 3 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| pipelinesapproval | betterado-inherited | 0 | 3 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| pipelines-v2 | betterado-inherited | 0 | 0 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| serviceendpoint | betterado-extended | 1 | 55 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| core | betterado-inherited | 5 | 1 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| build | betterado-extended | 12 | 32 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| policy | betterado-inherited | 3 | 0 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| git | betterado-inherited | 0 | 7 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| feed | betterado-extended | 5 | 1 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| wiki | betterado-inherited | 2 | 0 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| identity | betterado-inherited | 7 | 31 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| graph | betterado-inherited | 14 | 30 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| security | betterado-inherited | 0 | 2 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| permissions | betterado-inherited | 1 | 13 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| securityroles | betterado-extended | 9 | 0 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| memberentitlementmanagement | betterado-inherited | 0 | 58 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| notification | betterado-inherited | 0 | 0 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| servicehook | betterado-inherited | 2 | 5 | v7.2 delta: sourced from learn.microsoft.com; live verification pending |
| dashboard | betterado-inherited | 0 | 8 | v7.2 delta: Position and Widgets are the primary deferred writable fields; all server-assigned fields out-of-scope |
| extension | betterado-inherited | 0 | 3 | v7.2 delta: installState sub-fields (lastUpdated, installationIssues, platform error flags) deferred; core install/disable/version covered |
| gallery-extensionmanagement | betterado-inherited | 3 | 6 | v7.2 delta: betterado_extension_install (WI-2) covers core lifecycle; gallery metadata data source and extension settings deferred |
| featuremanagement | betterado-inherited | 0 | 9 | v7.2 delta: betterado_feature_flag targets host/project scopes; user-scoped flags and feature-definition data source deferred |
| workitemtracking | betterado-inherited | 14 | 0 | v7.2 delta: AssignedTo, History, non-parent link types, query isPublic/columns/sortColumns/filterOptions and structured clause fields deferred |
| workitemtrackingprocess | betterado-inherited | 0 | 19 | v7.2 delta: behaviors, contribution pages/groups, inherited-page ordering, system-control overrides, and server-assigned enum fields deferred |
| accounts-profile | betterado-inherited | 8 | 0 | v7.2 delta: Profile data source (betterado_profile) fully deferred to follow-on WI; accountOwner/accountStatus/ownerId query param open |
| test | betterado-inherited | 5 | 3 | v7.2 delta: 5 declarative types (plan, suite, configuration, variable, retention settings) gap-open pending testplan SDK vendoring; runs/results/cases gap-deferred as data sources |

## Priority backlog

_(Populated by WI-5 after all tier normalization completes.)_

### Tier 1 — betterado-net-new resources

### Tier 2 — high-value upstream gaps

### Tier 3 — low-value computed-field gaps
