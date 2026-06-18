# Provider state — post-capstone (v0.2.0, 2026-06-18)

Where `terraform-provider-betterado` stands after the release/task-group capstone.

## What it is now
The only Terraform provider with **classic ADO release pipeline** support
(upstream `microsoft/terraform-provider-azuredevops` has zero). v0.2.0 published.
Net-new surface, all live-proven against real ADO and documented in the registry:

- **`betterado_release_definition`** — deeply nested: `stages` (renamed from
  `environment`), each with `deploy_phase` (agent / agentless / deployment-group),
  `deployment_input`, `workflow_task` (incl. `timeout_in_minutes` +
  `retry_count_on_task_failure`), approvals, pre/post `*_deployment_gates`,
  `condition`, `environment_options`, `execution_policy`, `retention_policy`,
  variables + variable groups; `artifact` with triggers incl.
  `container_image_trigger`, artifact-tag triggers, `createReleaseOnBuildTagging`,
  schedule + source-repo triggers, `environment_trigger`. 8/8 writable gaps closed.
- **`betterado_release_folder`** — reviewed → confirmed complete (both writable fields).
- **`betterado_release_definition_permissions`** — ACL bits on a pipeline. Four
  are **live-proven** by the acceptance suite (`ViewReleases`,
  `EditReleaseEnvironment`, `DeleteReleases`, `CreateReleases`); the ReleaseManagement
  namespace exposes more actions, untested (keys are validated by ADO at apply,
  not the schema). NOTE: `docs/release-definition-permissions-gap-matrix.md` + the
  legacy permissions example use a DIFFERENT name set (`ViewReleasePipeline`,
  `ManageReleasesSettings`, `QueueRelease`…) — a doc/code inconsistency to reconcile
  against the live namespace.
- **`betterado_task_group`** — incl. input metadata: `icon_url`, `visible_rule`,
  `properties`, `aliases`.
- Data sources: release_definition (+ history, revision, list), release_folder, task_group.

## Coverage posture
Every net-new resource has a gap matrix in `docs/`. The remaining unmapped API
fields are read-only / computed / deprecated (no action). Writable parity reached.

## Known limitation (deferred, by decision)
`stages` (and nested collections) use **block** syntax, not assignable arrays.
SDKv2 `ConfigMode:Attr` would give `stages = [{…}]` but forces consumers to
null-fill every nested attribute at every level (cty.Object, not
ObjectWithOptionalAttrs) — worse ergonomics. Clean array+optional needs a
**holistic terraform-plugin-framework migration** (roadmap.md § Future).

## Quality posture
Two-gate model enforced as standing ACs: a live `TF_ACC` acceptance test
(apply→read-back→idempotency→destroy) + the CI-equivalent gate (make test +
golangci-lint + terrafmt). Live-evidence demos (`CaptureLiveEvidence` → real REST
GET). Registry docs regenerated + a release/version contract (tag after merge).
