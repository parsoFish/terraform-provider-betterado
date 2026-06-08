# Release Definition Implementation Roadmap

> **Initiative:** `INIT-2026-06-08-release-definition-schema-audit`
> **Derived from:** `docs/release-definition-gap-matrix.md` (WI-1 output)
> **Priority axes:** (a) ADO 7.2 required-for-create, (b) operator config-surface parity, (c) complexity

---

## How to Read This Document

Each entry below is a **logical gap cluster** — a group of related missing fields or capabilities that should be implemented together in a single work item. Entries are ordered by decreasing priority: P1 first, P3 last.

**Iteration budget calibration** is based on the Go provider domain baseline from `brain/cycles/themes/work-item-completion-by-domain.md`:

- Moderate complexity (new schema attributes wired to existing expand/flatten helpers): **3–4 iterations**
- Deeply nested / novel sub-schema (new resource sub-block, new type mapping): **4–5 iterations**
- Acceptance-test only (no schema change, real ADO fixture required): **2–3 iterations**
- New data source (stub → full read-back): **4–5 iterations**

---

## P1 — Required-for-create and high-parity gaps

### WI-A: Acceptance-test coverage for retention_policy (env-level)

**Gap:** `environment[].retention_policy` (`days_to_keep`, `releases_to_keep`, `retain_build`) is wired in the resource schema and expand/flatten helpers but has no acceptance test. The code path is unvalidated end-to-end.

**ADO 7.2 relevance:** `retentionPolicy` is returned on every `GetReleaseDefinition` response; a missing round-trip test can silently mask a diff-loop bug.

**Fields in scope:**
- `environment[].retention_policy.days_to_keep`
- `environment[].retention_policy.releases_to_keep`
- `environment[].retention_policy.retain_build`

**Deliverable:** New acceptance test `TestAccReleaseDefinition_withRetentionPolicy` asserting all three fields round-trip correctly.

**Estimated iteration budget:** 2–3

**depends_on:** _(none — schema is already complete)_

---

### WI-B: Acceptance-test coverage for post_deploy_approval with live approver

**Gap:** `TestAccReleaseDefinition_withApprovalOptions` exercises `approval_options` but does not wire a real approver identity. `approver[].id`, `approver[].is_automated`, and `approver[].rank` are untested live.

**ADO 7.2 relevance:** Approval workflows are a core enterprise release-gate mechanism; the schema already exists but the round-trip is unverified.

**Fields in scope:**
- `environment[].pre_deploy_approval.approver[].id`
- `environment[].pre_deploy_approval.approver[].is_automated`
- `environment[].pre_deploy_approval.approver[].rank`
- `environment[].post_deploy_approval` (full block, with approver)

**Deliverable:** New acceptance test `TestAccReleaseDefinition_withPostDeployApproval` using a real approver identity from env-var, asserting `post_deploy_approval` and `pre_deploy_approval.approver[]` round-trip.

**Estimated iteration budget:** 3

**depends_on:** _(none — schema is already complete)_

---

### WI-C: workflow_task.timeoutInMinutes and retryCountOnTaskFailure

**Gap:** Two writable `WorkflowTask` fields are missing from the TF schema: `timeoutInMinutes` (per-task timeout) and `retryCountOnTaskFailure` (task retry count). Both are commonly used in enterprise pipelines.

**ADO 7.2 relevance:** Required-for-create? No. Config-surface parity: Yes — the ADO web UI exposes both in the task editor; operators expect them in IaC.

**Fields in scope:**
- `deploy_phase[].workflow_task[].timeout_in_minutes`
- `deploy_phase[].workflow_task[].retry_count_on_task_failure`
- Same fields for gate tasks in `pre/post_deployment_gates[].gate[].task[]`

**Schema changes:** Extend `workflowTaskSchema()` with two new optional+computed integer attributes. Extend `expandWorkflowTask` and `flattenWorkflowTask`.

**Deliverable:** Schema additions + new acceptance test `TestAccReleaseDefinition_withWorkflowTaskAllFields`.

**Estimated iteration budget:** 3–4

**depends_on:** _(none)_

---

## P2 — Config-surface parity gaps

### WI-D: deploymentInput.overrideInputs

**Gap:** `deploy_phase[].deployment_input.override_inputs` (`*map[string]string`) is not surfaced. This field allows overriding task input values at the phase level — commonly used for parameterising shared task groups.

**ADO 7.2 relevance:** Config-surface parity — visible in the ADO UI under "Deployment input overrides"; moderate usage in enterprise pipelines.

**Fields in scope:**
- `deploy_phase[].deployment_input.override_inputs` (TypeMap `map[string]string`)

**Schema changes:** Add optional TypeMap attribute to `deploymentInputSchema()`; extend `expandDeploymentInput` / `flattenDeploymentInput`.

**Deliverable:** Schema addition + new acceptance test `TestAccReleaseDefinition_withOverrideInputs`.

**Estimated iteration budget:** 3

**depends_on:** _(none)_

---

### WI-E: environment[].environmentTriggers (cross-stage deploy triggers)

**Gap:** `environment[].environmentTriggers` (`*[]EnvironmentTrigger`) is missing. This enables a stage to trigger automatically after another stage succeeds — a key multi-stage pipeline pattern.

**ADO 7.2 relevance:** High config-surface parity; IaC users routinely configure "Deploy to Production after Staging succeeds" chains.

**Fields in scope:**
- `environment[].environment_trigger[]` — new optional list block
  - `trigger_type` (string; e.g. `"environmentState"`)
  - `trigger_content` (string JSON; ADO returns a JSON-encoded object)

**Schema changes:** New sub-block `environment_trigger` inside the environment schema; new `expandEnvironmentTrigger` / `flattenEnvironmentTrigger` helpers.

**Deliverable:** Schema addition + new acceptance test `TestAccReleaseDefinition_withEnvironmentTriggers`.

**Estimated iteration budget:** 4–5

**depends_on:** _(none; but naturally tested alongside WI-B)_

---

### WI-F: Trigger tag-filter and createReleaseOnBuildTagging

**Gap:** `triggers[artifactSource].triggerConditions[].tags` and `createReleaseOnBuildTagging` are not surfaced. Tag-based CD trigger conditions are used in git-flow scenarios (release/* branches + tags).

**ADO 7.2 relevance:** Config-surface parity — the ADO web UI exposes both checkboxes under "Artifact trigger conditions".

**Fields in scope:**
- `trigger[].cd_artifact_trigger[].tag_filter[]` — new optional list sub-block
  - `include[]` / `exclude[]` (list of string, mirroring `branch_filter` pattern)
- `trigger[].cd_artifact_trigger[].create_release_on_build_tagging` (optional bool)

**Schema changes:** Extend `cdArtifactTriggerSchema()`; extend `expandArtifactTrigger` / `flattenArtifactTrigger`.

**Deliverable:** Schema additions + new acceptance test `TestAccReleaseDefinition_withCdArtifactTriggerTagFilter`.

**Estimated iteration budget:** 3–4

**depends_on:** _(none)_

---

### WI-G: variable_groups acceptance-test coverage

**Gap:** `variable_groups` at definition level and at `environment[]` level is wired in the schema but has no acceptance test. The code path is unvalidated.

**ADO 7.2 relevance:** Config-surface parity — variable groups are a fundamental shared-secrets mechanism.

**Fields in scope:**
- `variable_groups` (definition-level)
- `environment[].variable_groups`

**Deliverable:** New acceptance test `TestAccReleaseDefinition_withVariableGroups` requiring a real ADO variable group fixture.

**Estimated iteration budget:** 3

**depends_on:** _(none — schema is already complete)_

---

### WI-H: Schedule-trigger and CD artifact-trigger acceptance-test coverage

**Gap:** `schedule_trigger` and `cd_artifact_trigger.branch_filter` have no dedicated acceptance tests. Both are wired in the schema but the live round-trip is unconfirmed.

**Fields in scope:**
- `trigger[].schedule_trigger[]` — all fields
- `trigger[].cd_artifact_trigger[].branch_filter[]`

**Deliverable:**
- `TestAccReleaseDefinition_withScheduleTrigger`
- `TestAccReleaseDefinition_withCdArtifactTriggerBranchFilter`

**Estimated iteration budget:** 2–3

**depends_on:** _(none)_

---

## P3 — Deferred / complex gaps

### WI-I: containerImageTrigger

**Gap:** `triggers[containerImageTrigger]` exists in the ADO API but not in the Go SDK type system (no `ContainerImageTrigger` struct). Implementing it requires deserialising the raw JSON interface.

**ADO 7.2 relevance:** Niche — used only when pipelines pull from Azure Container Registry.

**Deliverable:** New `container_image_trigger` sub-block + `TestAccReleaseDefinition_withContainerImageTrigger` (requires ACR fixture).

**Estimated iteration budget:** 5–6

**depends_on:** _(none; but should be done after WI-F to reuse trigger plumbing patterns)_

---

### WI-J: Full read-back data source (data_release_definition expansion)

**Gap:** The existing `data_release_definition` data source surfaces only 4 fields (`name`, `path`, `description`, `release_name_format`). A full read-back data source matching the resource schema would enable cross-definition references and import-time inspection.

**ADO 7.2 relevance:** Config-surface parity — operators frequently use data sources to reference existing definitions.

**Deliverable:** Expanded `data_release_definition` schema surfacing all mapped resource fields + `TestAccDataReleaseDefinition_full`.

**Estimated iteration budget:** 4–5

**depends_on:** WI-C (workflow_task fields should match between resource and data source), WI-E (environment triggers should be readable)

---

### WI-K: Definition-revision and history data sources

**Gap:** `GetDefinitionRevision` and `GetReleaseDefinitionHistory` SDK methods are not surfaced as data sources. Useful for compliance/audit dashboards and version-pinning.

**Fields in scope:**
- New `data_release_definition_revision` data source — returns JSON blob of a specific revision
- New `data_release_definition_history` data source — returns list of revision metadata

**Deliverable:** Two new data sources + acceptance tests.

**Estimated iteration budget:** 4–5 per data source (8–10 total, or split into two WIs)

**depends_on:** WI-J (data source infrastructure patterns established)

---

## Dependency Graph

```
WI-A  ──────────────────────────────────────────────────────────────────▶ (standalone)
WI-B  ──────────────────────────────────────────────────────────────────▶ (standalone)
WI-C  ──────────────────────────────────────────────────────────────────▶ (standalone)
WI-D  ──────────────────────────────────────────────────────────────────▶ (standalone)
WI-E  ──────────────────────────────────────────────────────────────────▶ (standalone)
WI-F  ──────────────────────────────────────────────────────────────────▶ (standalone)
WI-G  ──────────────────────────────────────────────────────────────────▶ (standalone)
WI-H  ──────────────────────────────────────────────────────────────────▶ (standalone)
WI-I  depends_on WI-F (trigger plumbing patterns)
WI-J  depends_on WI-C, WI-E
WI-K  depends_on WI-J
```

### Explicit depends_on table

| Work item | depends_on |
|---|---|
| WI-A | _(none)_ |
| WI-B | _(none)_ |
| WI-C | _(none)_ |
| WI-D | _(none)_ |
| WI-E | _(none)_ |
| WI-F | _(none)_ |
| WI-G | _(none)_ |
| WI-H | _(none)_ |
| WI-I | WI-F |
| WI-J | WI-C, WI-E |
| WI-K | WI-J |

---

## Recommended Execution Order

For a team working serially, the following order maximises value delivery per iteration:

1. **WI-A** — Fastest win; fixes known stale test gap; no schema change required.
2. **WI-B** — Fixes the other known stale test gap; validates approval round-trip.
3. **WI-C** — High parity; common user request; moderate schema change.
4. **WI-H** — Validates existing trigger schema with live tests.
5. **WI-D** — Moderate parity; straightforward TypeMap addition.
6. **WI-G** — Validates existing variable-group wiring.
7. **WI-F** — Extends trigger conditions; enables WI-I.
8. **WI-E** — Multi-stage triggers; larger sub-block.
9. **WI-I** — Niche; complex raw-JSON handling.
10. **WI-J** — Data source expansion; depends on WI-C + WI-E.
11. **WI-K** — Audit data sources; depends on WI-J.

---

## Out of Scope

The following capabilities are explicitly **out of scope** for this initiative. They involve **read-only / computed values** or **imperative runtime operations** that are not part of the declarative `betterado_release_definition` resource lifecycle.

### Read-only / Computed fields (never writable via TF)

These fields exist in the ADO SDK but are server-assigned and must not be written by a Terraform provider:

| Field | Reason excluded |
|---|---|
| `ReleaseDefinition._links` | REST navigation links; server-computed |
| `ReleaseDefinition.url` | REST canonical URL; server-computed |
| `ReleaseDefinition.projectReference` | Project tracked as `project_id` string; nested struct is redundant |
| `ReleaseDefinition.createdBy` / `createdOn` | Audit metadata; server-assigned |
| `ReleaseDefinition.modifiedBy` / `modifiedOn` | Audit metadata; server-assigned |
| `ReleaseDefinition.isDeleted` | Soft-delete flag; managed by ADO, not TF |
| `ReleaseDefinition.lastRelease` | Reference to last release instance; runtime state |
| `ReleaseDefinition.source` | Enum set by ADO on create (e.g. `restApi`); read-only |
| `ReleaseDefinition.retentionPolicy` | **Deprecated** in ADO SDK; use env-level retention |
| `ReleaseDefinition.properties` | Opaque internal bag; not user-configurable |
| `ReleaseDefinitionEnvironment.badgeUrl` | Server-generated URL |
| `ReleaseDefinitionEnvironment.currentRelease` | Runtime state reference |
| `ReleaseDefinitionEnvironment.deployStep` | Internal gate step ID; read-only |
| `ReleaseDefinitionEnvironment.queueId` | Legacy field; superseded by `deploymentInput.queueId` |
| `ReleaseDefinitionEnvironment.runOptions` | **Deprecated** in ADO SDK |
| `ReleaseDefinitionEnvironment.processParameters` | Internal pipeline parameters; not user-configurable via TF |
| `ReleaseDefinitionApprovalStep.id` | Internal step ID (not the approver UUID); server-assigned |
| `ReleaseDefinitionApprovalStep.isNotificationOn` | Server-computed notification flag |
| `Artifact.isRetained` | Set by release runtime; not the definition |
| `Artifact.sourceId` | **Deprecated** in ADO SDK |
| `ReleaseDefinitionGatesStep.id` | Server-assigned step ID |
| `AgentBasedDeployPhase.refName` | Internal reference name; read-only |
| `AgentDeploymentInput.imageId` | Legacy image ID |

### Imperative runtime operations (out of scope by design)

The following SDK methods are **imperative runtime operations** on release instances, not on release definitions. They do not belong in a declarative Terraform provider:

| Operation | Reason excluded |
|---|---|
| `CreateRelease` | Creates a release instance — imperative runtime action, not IaC lifecycle |
| `UpdateApproval` / `GetApprovals` | Approval state management — runtime, human-in-the-loop workflow |
| `ManualInterventions` (resume/update) | Interactive pipeline control — not declarative |
| `GetDeployments` / deployment lifecycle | Deployment execution state — runtime, not definition |
| `GetReleaseEnvironment` (runtime state) | Live environment execution state — not a definition concern |
| `GetRelease` / `GetReleases` | Release instance retrieval — separate runtime domain |

> These operations are not "missing" from the TF provider — they are architecturally incompatible with the declarative IaC lifecycle that Terraform enforces. A `betterado_release_definition` resource manages the _definition_ of a release pipeline; it does not manage the _execution_ of releases.

---

## Appendix: Gap Summary (from WI-1 gap matrix)

| Area | Mapped | Partial | Missing | Writable gaps |
|---|---|---|---|---|
| ReleaseDefinition (top-level) | 9 | 1 | 12 | 0 (comment — deferred) |
| ReleaseDefinitionEnvironment | 15 | 0 | 10 | 1 (`environmentTriggers` — WI-E) |
| Approvals / ApprovalOptions | 9 | 0 | 2 | 0 (both read-only) |
| Artifact | 4 | 0 | 2 | 0 (both read-only/deprecated) |
| Triggers | 9 | 0 | 5 | 3 (`tags`, `createReleaseOnBuildTagging` — WI-F; `containerImageTrigger` — WI-I) |
| DeployPhase / DeploymentInput | 15 | 0 | 4 | 1 (`overrideInputs` — WI-D) |
| WorkflowTask | 9 | 0 | 3 | 2 (`timeoutInMinutes`, `retryCountOnTaskFailure` — WI-C) |
| EnvironmentRetentionPolicy | 3 | 0 | 0 | 0 ✅ |
| EnvironmentOptions | 9 | 0 | 0 | 0 ✅ |
| EnvironmentExecutionPolicy | 2 | 0 | 0 | 0 ✅ |
| GatesStep / GatesOptions | 6 | 0 | 1 | 0 (read-only) |
| ConfigurationVariableValue | 3 | 0 | 0 | 0 ✅ |
| **TOTAL** | **93** | **1** | **39** | **7 writable** |

> Of 39 missing fields: **32 are read-only / computed / deprecated** (zero implementation required). **7 are writable gaps** addressed by WI-C through WI-I above.
