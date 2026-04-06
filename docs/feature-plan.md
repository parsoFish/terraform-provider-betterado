# Feature Plan: Deepening betterado_release_definition

Based on live API exploration on 2026-04-06 against `davidgparsonson/DPLife`.

## Prioritization Criteria

- **P0 (Must Have)**: Causes drift, data loss, or broken plans without it
- **P1 (Should Have)**: Common real-world usage, users will need it quickly
- **P2 (Nice to Have)**: Advanced features, less commonly configured via Terraform

---

## Phase 1: Fix Correctness Issues (P0)

These are bugs or inaccuracies in the current implementation.

### 1.1 Fix revision conflict error handling
**Current**: CLAUDE.md says retry on 409. **Actual**: API returns 400 with `InvalidRequestException`.
- Update error detection in `resourceReleaseDefinitionUpdate` to match on message text or typeKey
- Update CLAUDE.md documentation

### 1.2 Fix tags behavior
**Current**: Schema accepts tags, expand sends them, but API silently ignores them on create and update.
- Option A: Remove `tags` from schema (honest but loses future compat)
- Option B: Keep in schema but document as known limitation, suppress diff
- **Recommendation**: Option B — keep the schema, add a note in docs, investigate separate tags API

### 1.3 Add deploymentInput to deploy phases
**Current**: `expandDeployPhases` creates `AgentBasedDeployPhase` but never sets `DeploymentInput`. This means `queueId` (agent pool) is never sent — the API picks a default.
- Add `deployment_input` block to `deploy_phase` schema
- Essential fields: `queue_id`, `demands`, `timeout_in_minutes`, `job_cancel_timeout_in_minutes`, `condition`
- This is P0 because without `queue_id`, users can't target specific agent pools

---

## Phase 2: Approval & Gate Enhancements (P1)

### 2.1 Add approvalOptions to pre/post deploy approvals
Fields to add to approval schema:
- `required_approver_count` (int, Optional, null = all required)
- `release_creator_can_be_approver` (bool, Optional, default false)
- `enforce_identity_revalidation` (bool, Optional, default false)
- `timeout_in_minutes` (int, Optional, default 0)
- `execution_order` (string, Optional, "beforeGates"/"afterSuccessfulGates")
- `auto_triggered_and_previous_environment_approved_can_be_skipped` (bool, Optional)

Also add `is_notification_on` to each approver step.

### 2.2 Add pre/post deployment gates
New schema blocks: `pre_deployment_gates` and `post_deployment_gates`
```
gates_options {
  is_enabled         = true
  timeout            = 1440
  sampling_interval  = 5
  stabilization_time = 5
  minimum_success_duration = 0
}
gate { ... }  # Future: individual gate task definitions
```

---

## Phase 3: Environment Configuration (P1)

### 3.1 Add environmentOptions block
New `environment_options` block on each environment:
- `email_notification_type` (string: "OnlyOnFailure", "Always", "Never")
- `email_recipients` (string)
- `skip_artifacts_download` (bool)
- `timeout_in_minutes` (int)
- `enable_access_token` (bool)
- `publish_deployment_status` (bool)
- `badge_enabled` (bool)
- `auto_link_work_items` (bool)
- `pull_request_deployment_enabled` (bool)

### 3.2 Add executionPolicy block
- `concurrency_count` (int, default 1)
- `queue_depth_count` (int, validation: 0 or 1 only)

### 3.3 Add top-level isDisabled
- `enabled` attribute (bool, Optional, default true)
- Maps to `isDisabled` (inverted) in expand/flatten

---

## Phase 4: Deploy Phase Enhancements (P1-P2)

### 4.1 Remaining deploymentInput fields (P2)
- `parallel_execution` block (`parallel_execution_type`: "none", "multiConfiguration", "multiMachine")
- `agent_specification` block (for specifying vmImage)
- `skip_artifacts_download` (bool)
- `enable_access_token` (bool)
- `override_inputs` (map)
- `artifacts_download_input` block

### 4.2 Workflow task enhancements (P2)
- `timeout_in_minutes` (int)
- `retry_count_on_task_failure` (int)
- `ref_name` (string)

---

## Phase 5: Advanced Features (P2)

### 5.1 Environment demands
- Flat string list at environment level (name/value alternating pairs)
- May need a custom schema type or key-value list

### 5.2 Schedules
- Scheduled deployment triggers per environment

### 5.3 Environment triggers
- Auto-redeploy triggers, rollback triggers

### 5.4 Triggers (top-level)
- Artifact triggers, scheduled triggers, pull request triggers

---

## Phase 6: New Resource — betterado_task_group (P1)

Full API reference: `docs/api-reference/task-groups.md`

### Overview
Task groups are reusable collections of build/release tasks with parameterized inputs. They're referenced from release definition deploy phases as `definitionType: "metaTask"` workflow tasks. This makes them a natural companion resource.

### Key Design Decisions
- **API host:** `dev.azure.com` (core host, NOT vsrm.dev.azure.com)
- **SDK client:** `TaskAgentClient` — already initialized in client.go
- **ID type:** UUID (not int like release definitions)
- **Versioning:** `version.major` should be ForceNew; version bumping uses a separate publish workflow
- **Revision:** Optimistic concurrency via 409 Conflict (standard, unlike release defs which use 400)
- **Update:** PUT requires full object (omitted fields are wiped)

### Schema (TF attributes)
- `name` (string, Required)
- `friendly_name` (string, Required)
- `project_id` (string, Required, ForceNew)
- `description` (string, Optional)
- `category` (string, Required — "Deploy", "Build", "Utility", etc.)
- `instance_name_format` (string, Optional)
- `author` (string, Optional)
- `runs_on` (list of string, Optional — default ["Agent"])
- `version` block (Required) — `major` (ForceNew), `minor`, `patch`
- `input` block list (Optional) — parameterized inputs with name, label, type, default_value, required
- `task` block list (Required, MinItems 1) — steps with display_name, task_id, task_version, inputs, condition, etc.
- Computed: `id` (UUID), `revision`, `definition_type`

### Effort: Medium-Large (new resource, but pattern is well-established)

---

## Implementation Order (Recommended)

| Order | Item | Phase | Effort | Impact |
|-------|------|-------|--------|--------|
| 1 | deploymentInput (queue_id, demands, timeouts, condition) | 1.3 | Medium | High — unblocks agent pool targeting |
| 2 | Fix revision conflict handling | 1.1 | Small | High — correctness |
| 3 | Fix tags behavior | 1.2 | Small | Medium — stops confusing drift |
| 4 | approvalOptions | 2.1 | Medium | High — real pipelines need approval config |
| 5 | environmentOptions | 3.1 | Medium | Medium — notification/badge config |
| 6 | **betterado_task_group resource** | **6** | **Med-Large** | **High — reusable task definitions** |
| 7 | isDisabled (enabled) | 3.3 | Small | Medium — pipeline lifecycle |
| 8 | executionPolicy | 3.2 | Small | Medium — concurrency control |
| 9 | Deployment gates | 2.2 | Large | Medium — enterprise feature |
| 10 | Remaining deploymentInput | 4.1 | Medium | Low-Med |
| 11 | Workflow task enhancements | 4.2 | Small | Low |
| 12 | Environment demands | 5.1 | Small | Low |
| 13 | Schedules & triggers | 5.2-5.4 | Large | Low |

---

## API Behavioral Notes for Implementation

1. **PUT requires full object** — always read-modify-write, never partial update
2. **Revision is optimistic concurrency** — must include current revision on PUT
3. **Stale revision = 400** (not 409) with typeKey `InvalidRequestException`
4. **Demands are flat string arrays** — `["name", "value", "name2", "value2"]`
5. **queueDepthCount** only allows 0 or 1
6. **Secret variable values** return as null on read — flatten must preserve user-provided value
7. **Tags** are not settable via definitions API — separate endpoint needed
8. **Environment IDs, approval IDs, gate IDs** are all server-assigned on create
9. **environmentOptions** respect explicit values on create; absent = server defaults
10. **Delete returns 204** with no body
