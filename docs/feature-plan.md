# Feature Plan: Deepening betterado_release_definition

Based on live API exploration on 2026-04-06 against `davidgparsonson/DPLife`.

## Prioritization Criteria

- **P0 (Must Have)**: Causes drift, data loss, or broken plans without it
- **P1 (Should Have)**: Common real-world usage, users will need it quickly
- **P2 (Nice to Have)**: Advanced features, less commonly configured via Terraform

---

## Phase 1: Fix Correctness Issues (P0) — ✅ COMPLETE

All P0 issues resolved.

### 1.1 Fix revision conflict error handling — ✅ Done
Update function detects HTTP 400 with "old copy of the release pipeline" message, re-reads definition for current revision, retries once.

### 1.2 Fix tags behavior — ✅ Done
Tags kept in schema, documented as known limitation. API silently ignores on create/update.

### 1.3 Add deploymentInput to deploy phases — ✅ Done
`deployment_input` block with `queue_id`, `timeout_in_minutes`, `job_cancel_timeout_in_minutes`, `condition`, `skip_artifacts_download`, `enable_access_token`.

---

## Phase 2: Approval & Gate Enhancements (P1) — Approvals ✅ / Gates Remaining

### 2.1 Add approvalOptions to pre/post deploy approvals — ✅ Done
All 6 approval_options fields implemented:
- `required_approver_count`, `release_creator_can_be_approver`, `enforce_identity_revalidation`
- `timeout_in_minutes`, `execution_order`, `auto_triggered_and_previous_environment_approved_can_be_skipped`
Also added `is_notification_on` to each approver step.

### 2.2 Add pre/post deployment gates — 🔲 Not started
New schema blocks needed: `pre_deployment_gates` and `post_deployment_gates`
```
gates_options {
  is_enabled         = true
  timeout            = 1440
  sampling_interval  = 5
  stabilization_time = 5
  minimum_success_duration = 0
}
gate { ... }  # Individual gate task definitions
```

---

## Phase 3: Environment Configuration (P1) — ✅ COMPLETE

### 3.1 Add environmentOptions block — ✅ Done
All fields implemented: `email_notification_type`, `email_recipients`, `badge_enabled`, `auto_link_work_items`, `pull_request_deployment_enabled`, `publish_deployment_status`, `timeout_in_minutes`, `enable_access_token`, `skip_artifacts_download`.

### 3.2 Add executionPolicy block — ✅ Done
`concurrency_count` and `queue_depth_count` implemented.

### 3.3 Add top-level isDisabled — ✅ Done
Schema has `enabled` attribute mapping to `isDisabled` (inverted) in expand/flatten.

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

## Phase 6: New Resource — betterado_task_group (P1) — ✅ COMPLETE

**File:** `azuredevops/internal/service/taskagent/resource_task_group.go`

Full CRUD lifecycle implemented and tested idempotent. Registered in provider.go.

### Implemented Features
- Name, description, category, author, instance_name_format
- Version block (major ForceNew, minor, patch, is_test)
- Parameterized inputs (name, label, type, default_value, required, help, group_name)
- Task steps (task_id, version, display_name, inputs, condition, enabled, always_run, continue_on_error, etc.)
- runs_on configuration
- Computed: id (UUID), revision, definition_type
- API host: `dev.azure.com` (core host, uses TaskAgentClient)
- Update: PUT requires full object (read-modify-write pattern)

---

## Implementation Order (Recommended)

| Order | Item | Phase | Status |
|-------|------|-------|--------|
| 1 | deploymentInput (queue_id, demands, timeouts, condition) | 1.3 | ✅ Done |
| 2 | Fix revision conflict handling | 1.1 | ✅ Done |
| 3 | Fix tags behavior | 1.2 | ✅ Done (documented limitation) |
| 4 | approvalOptions | 2.1 | ✅ Done |
| 5 | environmentOptions | 3.1 | ✅ Done |
| 6 | **betterado_task_group resource** | **6** | **✅ Done** |
| 7 | isDisabled (enabled) | 3.3 | ✅ Done |
| 8 | executionPolicy | 3.2 | ✅ Done |
| 9 | Deployment gates | 2.2 | 🔲 Not started |
| 10 | Remaining deploymentInput | 4.1 | 🔲 Not started |
| 11 | Workflow task enhancements | 4.2 | 🔲 Not started |
| 12 | Environment demands | 5.1 | 🔲 Not started |
| 13 | Schedules & triggers | 5.2-5.4 | 🔲 Not started |

### New Items from PR #178 Analysis (April 2026)

| Priority | Feature | Notes |
|----------|---------|-------|
| High | Agentless jobs (RunOnServer) | Server-side tasks: Delay, InvokeRESTAPI, ManualIntervention |
| High | Deployment group jobs | Deploy to on-prem/VM machines via deployment groups |
| High | Gates (pre/post deploy) | Automated quality checks before/after deployment |
| High | Multi-config / multi-agent parallelism | Variable multipliers, parallel agent execution |
| Medium | Artifact filters | Branch/tag conditions per artifact |
| Medium | Tags (separate API) | Definitions endpoint doesn't persist; needs separate endpoint |
| Medium | Demands per environment | Agent capability matching |
| Medium | Properties | Definition/environment level for Jira/Boards integration |
| Medium | Schedules | Timed release triggers per environment |
| Lower | override_inputs on tasks | Override task inputs at release time |
| Lower | Environment type | production/staging/testing metadata classification |
| Lower | Build artifact download control | Per-job selective/skip/all artifact download |
| Lower | Release triggers | Definition-level trigger configuration |

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
