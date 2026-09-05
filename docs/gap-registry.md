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

> Synthesized from all 31 per-area gap matrices by WI-5 (INIT-2026-09-05-init-gap-registry-consolidation).
> Three tiers: Tier 1 = betterado-net-new resources, Tier 2 = high-value upstream gaps (writable fields in commonly used areas), Tier 3 = low-value computed-field gaps.

### Tier 1 — betterado-net-new resources

Gaps in resources that ONLY betterado implements — highest impact because these have no upstream counterpart.

- [ ] `task-group/betterado_task_group.icon_url` — complexity: low — `IconUrl` field in `TaskGroupCreateParameter` is not exposed; one-line string attribute addition
- [ ] `task-group/betterado_task_group.input[].visible_rule` — complexity: low — `VisibleRule` on `TaskInputDefinition` controls conditional UI visibility; simple string attribute
- [ ] `task-group/betterado_task_group.input[].properties` — complexity: medium — `Properties *map[string]string` on `TaskInputDefinition`; requires map attribute in schema
- [ ] `task-group/betterado_task_group.input[].aliases` — complexity: low — `Aliases *[]string` on `TaskInputDefinition`; simple list attribute addition

### Tier 2 — high-value upstream gaps

Gaps in betterado-extended or betterado-inherited resources where the field is writable and commonly used in ADO practice.

#### build area (betterado-extended)
- [ ] `build/betterado_build_definition.description` — complexity: low — writable string field on `BuildDefinition`; not in either schema; commonly set in ADO portal
- [ ] `build/betterado_build_definition.tags` — complexity: low — `*[]string` on `BuildDefinition`; writable; widely used for resource organization
- [ ] `build/betterado_build_definition.badge_enabled` — complexity: low — `*bool` on `BuildDefinition`; controls public status badge; not in framework or SDKv2 schema
- [ ] `build/betterado_build_definition.build_number_format` — complexity: low — `*string` on `BuildDefinition`; controls the build run naming pattern; frequently configured
- [ ] `build/betterado_build_definition.job_timeout_in_minutes` — complexity: low — `*int` on `BuildDefinition`; default per-job timeout; commonly set for long-running pipelines
- [ ] `build/betterado_build_definition.job_cancel_timeout_in_minutes` — complexity: low — `*int` on `BuildDefinition`; cancel grace period; not in either schema
- [ ] `build/betterado_build_definition.demands` — complexity: medium — `*[]interface{}` definition-level agent demands; complex type but used in classic pipeline configurations
- [ ] `build/betterado_build_definition.repository.clean` — complexity: low — `*string` workspace clean option; writable per-repository setting in `BuildRepository`
- [ ] `build/betterado_build_definition.repository.checkout_submodules` — complexity: low — `*bool`; controls submodule checkout; common SCM setting
- [ ] `build/betterado_build_definition.triggers` (build_completion + schedules) — complexity: high — `build_completion_trigger` and `schedules` trigger types are gap-deferred; complex polymorphic structures
- [ ] `build/data.betterado_build_definition.description` — complexity: low — missing computed read-back attribute on build definition data source
- [ ] `build/data.betterado_build_definition.tags` — complexity: low — missing computed read-back attribute on build definition data source

#### core area (betterado-inherited)
- [ ] `core/betterado_project.abbreviation` — complexity: low — short project abbreviation; writable via Update API; rarely used in IaC but simple to add
- [ ] `core/betterado_project_pipeline_settings.disableClassicBuildPipelineCreation` — complexity: low — policy enforcement field; writable boolean on pipeline settings
- [ ] `core/betterado_project_pipeline_settings.disableClassicReleasePipelineCreation` — complexity: low — policy enforcement field; writable boolean on pipeline settings
- [ ] `core/betterado_project_pipeline_settings.enforceNoAccessToSecretsFromForks` — complexity: low — security-hardening field; commonly required in enterprise environments
- [ ] `core/betterado_project_pipeline_settings.isCommentRequiredForPullRequest` — complexity: low — PR quality gate field; writable boolean on pipeline settings

#### serviceendpoint area (betterado-extended)
- [ ] `serviceendpoint/betterado_serviceendpoint_*.workload_identity_federation_subject` — complexity: medium — computed attribute for WIF scheme service connections; returned by ADO but not surfaced; needed for external IdP trust configuration

#### feed area (betterado-extended)
- [ ] `feed/betterado_feed.upstream_enabled` — complexity: low — `*bool` controlling upstream source fetching; writable via `FeedUpdate.UpstreamEnabled`; commonly configured
- [ ] `feed/betterado_feed.upstream_sources` — complexity: high — `*[]UpstreamSource` list of upstream source definitions; complex nested type; high value for package management configurations
- [ ] `feed/betterado_feed.badges_enabled` — complexity: low — `*bool` enabling package badge generation; writable via `FeedUpdate.BadgesEnabled`
- [ ] `feed/betterado_feed.description` — complexity: low — free-text description (255 chars); writable via `FeedUpdate.Description`; commonly set
- [ ] `feed/betterado_feed.default_view_id` — complexity: low — UUID of the default reader view; writable via `FeedUpdate.DefaultViewId`

#### securityroles area (betterado-extended)
- [ ] `securityroles/betterado_securityrole_assignment.identity.displayName` — complexity: low — read-only display name of the identity; not currently surfaced as computed attribute
- [ ] `securityroles/betterado_securityrole_assignment.identity.uniqueName` — complexity: low — read-only UPN/unique name; not surfaced; useful for audit/debugging
- [ ] `securityroles/betterado_securityrole_assignment.role.displayName` — complexity: low — read-only display name of the assigned role; not surfaced
- [ ] `securityroles/betterado_securityrole_assignment.role.allowPermissions` — complexity: low — read-only permission bitmask granted by the role; not surfaced
- [ ] `securityroles/betterado_securityrole_assignment.role.denyPermissions` — complexity: low — read-only permission bitmask denied; not surfaced
- [ ] `securityroles/betterado_securityrole_assignment.role.identifier` — complexity: low — read-only role identifier string; not surfaced
- [ ] `securityroles/betterado_securityrole_assignment.role.description` — complexity: low — read-only role description; not surfaced
- [ ] `securityroles/betterado_securityrole_assignment.assignment.access` — complexity: low — computed "assigned" vs "inherited" indicator; useful for detecting inherited assignments
- [ ] `securityroles/betterado_securityrole_assignment.assignment.accessDisplayName` — complexity: low — computed display text for access type; not surfaced

#### gallery-extensionmanagement area (betterado-inherited)
- [ ] `gallery-extensionmanagement/betterado_extension_install.extension_name` — complexity: low — computed display name for installed extension; not in shipped schema
- [ ] `gallery-extensionmanagement/betterado_extension_install.publisher_name` — complexity: low — computed publisher display name; not in shipped schema
- [ ] `gallery-extensionmanagement/betterado_extension_install.scopes` — complexity: low — computed list of permission scopes granted to the extension; not in shipped schema

#### permissions area (betterado-inherited)
- [ ] `permissions/betterado_release_definition_permissions.resource` — complexity: low — last remaining gap-open in permissions area; minor field on ReleaseManagement2 permissions resource

#### servicehook area (betterado-inherited)
- [ ] `servicehook/betterado_servicehook_webhook_tfs.commentPattern` — complexity: low — notification filter pattern field on TFS webhook; string attribute in `ServiceHooksSubscriptionParameters`
- [ ] `servicehook/betterado_servicehook_storage_queue_pipelines.checkedInBy` — complexity: low — filter field for push-triggered storage queue subscriptions; not in shipped schema

#### wiki area (betterado-inherited)
- [ ] `wiki/betterado_wiki.type` — complexity: low — wiki type enum (projectWiki vs codeWiki); read-back attribute not currently surfaced
- [ ] `wiki/betterado_wiki_page.git_item_path` — complexity: medium — computed git blob path for the underlying page; useful for cross-referencing with git resources
- [ ] `wiki/betterado_wiki_page.remote_url` — complexity: low — computed ADO portal URL for the page; useful for linking in downstream tools

### Tier 3 — low-value computed-field gaps

Gap-open items for fields that are Computed-only or rarely configured declaratively.

#### identity area
- [ ] `identity/data.betterado_identity_group.providerDisplayName` — complexity: low — computed display name not exposed as read-back attribute; informational only
- [ ] `identity/data.betterado_identity_group.isActive` — complexity: low — active/inactive flag; not surfaced; rarely needed in IaC decisions
- [ ] `identity/data.betterado_identity_group.isContainer` — complexity: low — always true for groups; informational computed attribute
- [ ] `identity/data.betterado_identity_groups.isActive` — complexity: low — active flag per-item in list; not surfaced
- [ ] `identity/data.betterado_identity_groups.isContainer` — complexity: low — container flag per-item in list; informational
- [ ] `identity/data.betterado_identity_groups.providerDisplayName` — complexity: low — display name per-item; not surfaced
- [ ] `identity/data.betterado_identity_user.isActive` — complexity: low — active/inactive flag; not surfaced

#### graph area
- [ ] `graph/data.betterado_group.description` — complexity: low — group description; not surfaced in data source read-back
- [ ] `graph/data.betterado_group.mailAddress` — complexity: low — group mail address; not surfaced; informational
- [ ] `graph/data.betterado_group.domain` — complexity: low — group domain; not surfaced
- [ ] `graph/data.betterado_group.principalName` — complexity: low — group principal name; not surfaced
- [ ] `graph/data.betterado_group.subjectKind` — complexity: low — always "group"; informational computed attribute
- [ ] `graph/data.betterado_group.url` — complexity: low — REST URL of the group; not surfaced
- [ ] `graph/data.betterado_groups[].subjectKind` — complexity: low — per-item subject kind; not surfaced
- [ ] `graph/data.betterado_service_principal.applicationId` — complexity: low — AAD application ID; computed attribute not surfaced
- [ ] `graph/data.betterado_service_principal.subjectKind` — complexity: low — always "servicePrincipal"; informational
- [ ] `graph/data.betterado_service_principal.domain` — complexity: low — tenant domain; not surfaced
- [ ] `graph/data.betterado_service_principal.principalName` — complexity: low — UPN-style name; not surfaced
- [ ] `graph/data.betterado_service_principal.mailAddress` — complexity: low — email; not surfaced
- [ ] `graph/data.betterado_users[].subjectKind` — complexity: low — per-item subject kind; not surfaced
- [ ] `graph/data.betterado_users[].domain` — complexity: low — per-item domain; not surfaced

#### accounts-profile area
- [ ] `accounts-profile/data.betterado_accounts.accountOwner` — complexity: low — owner UUID filter/attribute; deferred; no consumer use-case identified
- [ ] `accounts-profile/data.betterado_accounts.accountStatus` — complexity: low — enum status attribute; deferred; no consumer use-case identified
- [ ] `accounts-profile/data.betterado_accounts.ownerId` — complexity: low — owner filter parameter; deferred; primary PAT-based use case covered
- [ ] `accounts-profile/data.betterado_profile.id` — complexity: medium — full profile data source not built; profile UUID and core attributes deferred to follow-on WI
- [ ] `accounts-profile/data.betterado_profile.coreAttributes` — complexity: medium — key/value bag of core profile attrs (DisplayName, EmailAddress, PublicAlias); complex to model; follow-on WI
- [ ] `accounts-profile/data.betterado_profile.coreRevision` — complexity: low — max revision of core attributes; deferred
- [ ] `accounts-profile/data.betterado_profile.revision` — complexity: low — max revision across all attributes; deferred
- [ ] `accounts-profile/data.betterado_profile.profileState` — complexity: low — enum state; deferred

#### workitemtracking area
- [ ] `workitemtracking/betterado_workitem.assigned_to` — complexity: medium — identity-based assignee field; complex identity resolution; deferred out of scope for migration
- [ ] `workitemtracking/betterado_workitem.history` — complexity: medium — comment/history entry; append-only; deferred out of scope
- [ ] `workitemtracking/betterado_workitem.relations` (non-parent link types) — complexity: high — child links, related links, remote links; complex relational management; deferred
- [ ] `workitemtracking/betterado_workitemquery.isPublic` — complexity: low — whether query is in Shared Queries vs My Queries; deferred out of scope
- [ ] `workitemtracking/betterado_workitemquery.columns` — complexity: medium — display column list; WIQL is canonical form; deferred
- [ ] `workitemtracking/betterado_workitemquery.sortColumns` — complexity: medium — sort column list; WIQL is canonical form; deferred
- [ ] `workitemtracking/betterado_workitemquery.filterOptions` — complexity: low — link filter mode enum; deferred
- [ ] `workitemtracking/betterado_workitemquery.clauses` — complexity: high — structured flat query clauses; WIQL is canonical form; deferred
- [ ] `workitemtracking/betterado_workitemquery.linkClauses` — complexity: high — structured link query clauses; deferred
- [ ] `workitemtracking/betterado_workitemquery.sourceClauses` — complexity: high — structured tree query source clauses; deferred
- [ ] `workitemtracking/betterado_workitemquery.targetClauses` — complexity: high — structured tree query target clauses; deferred
- [ ] `workitemtracking/betterado_workitemquery_folder.isPublic` — complexity: low — whether folder is public; deferred out of scope
- [ ] `workitemtracking/betterado_workitemquery_folder.path` — complexity: low — derived from parent_id + name; not directly writable in a meaningful way; deferred
- [ ] `workitemtracking/betterado_workitemquery.path` — complexity: low — derived from parent_id/area + folder name; deferred

#### test area
- [ ] `test/betterado_test_plan` — complexity: high — full CRUD via `_apis/testplan/Plans`; testplan SDK vendoring required; high value but blocked on SDK work
- [ ] `test/betterado_test_suite` — complexity: high — hierarchical test suite resource; testplan SDK vendoring required
- [ ] `test/betterado_test_configuration` — complexity: medium — test configuration resource (e.g. "Chrome on Windows"); testplan SDK vendoring required
- [ ] `test/betterado_test_variable` — complexity: medium — parameterised test variable; testplan SDK vendoring required
- [ ] `test/betterado_test_result_retention_settings` — complexity: medium — singleton project-level retention settings; Get/Update lifecycle only (no create/delete)

#### feed data-source computed fields
- [ ] `feed/data.betterado_feed.upstream_enabled` — complexity: low — computed attribute not currently surfaced in data source
- [ ] `feed/data.betterado_feed.upstream_sources` — complexity: medium — computed upstream source list; not surfaced
- [ ] `feed/data.betterado_feed.badges_enabled` — complexity: low — computed attribute not surfaced
- [ ] `feed/data.betterado_feed.default_view_id` — complexity: low — computed UUID not surfaced
- [ ] `feed/data.betterado_feed.description` — complexity: low — computed description not surfaced
- [ ] `feed/data.betterado_feed.hide_deleted_package_versions` — complexity: low — computed attribute not surfaced

#### policy area
- [ ] `policy/betterado_repository_policy_max_file_size.use_uncompressed_size` — complexity: low — niche boolean field on max-file-size policy; gap-open in shipped schema
- [ ] `policy/betterado_repository_policy_check_credentials` — complexity: medium — deprecated credential scanning policy resource; gap-open (resource exists but marked deprecated)
- [ ] `policy/betterado_branch_policy_min_reviewers` (2 niche fields) — complexity: low — two minor fields gap-open on min-reviewers policy; low consumer demand
