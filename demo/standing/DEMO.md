# betterado v0.2.0 — Standing Feature Showcase

> The permanent demonstration of everything the 2026-06 release-pipeline capstone
> added to `terraform-provider-betterado`. Every snippet below is from the
> applyable config in this directory, validated against the built v0.2.0 provider.

`terraform-provider-betterado` is the **only** Terraform provider that manages
Azure DevOps **classic release pipelines** — upstream `microsoft/terraform-provider-azuredevops`
ships none. v0.2.0 took that surface from "exists" to "complete."

---

## ① The headline — `environment` → `stages`

A release pipeline is a sequence of *stages*. The provider used to call them
`environment` (the raw REST name). v0.2.0 renames the block to the term the
product UI uses. Breaking, no alias.

```diff
- environment {
+ stages {
    name = "Production"
    rank = 1
    ...
  }
```

See `release-definition.tf` — two stages, `Staging` → `Production`, chained by an
`environmentState` condition.

---

## ② Coverage gaps closed

Every net-new resource was audited against the ADO REST API (gap matrices live in
`docs/`). Each writable gap below is now implemented and exercised in this demo:

| Feature | What it unlocks | Where in the demo |
|---|---|---|
| `container_image_trigger` | Release on a new container image tag | `triggers { container_image_trigger { … } }` |
| `cd_artifact_trigger` tag filters + `create_release_on_build_tagging` | Release only on tagged/filtered builds | `triggers { cd_artifact_trigger { tag_filter … } }` |
| `schedule_trigger` / `source_repo_trigger` | Time- and repo-driven releases | `triggers { schedule_trigger / source_repo_trigger }` |
| `workflow_task.timeout_in_minutes` + `retry_count_on_task_failure` | Per-task timeout + retry | Staging "Deploy artifacts" task |
| `pre_deployment_gates` / `post_deployment_gates` | Automated approval gates (REST/poll) | both stages |
| `deployment_input.override_inputs` | Phase-level task-input overrides | Staging deploy phase |
| `deployment_input.parallel_execution` | Matrix / multi-config fan-out | Staging deploy phase |
| `environment_trigger` | Rollback / redeploy triggers | Production stage |

```hcl
triggers {
  container_image_trigger {
    artifact_alias = "_ci"
    label          = "latest"
  }
}

workflow_task {
  name                        = "Deploy artifacts"
  timeout_in_minutes          = 15
  retry_count_on_task_failure = 2
  # …
}
```

---

## ③ The four release-surface resources

- **`betterado_release_definition`** — the deeply-nested pipeline (stages →
  deploy_phases → deployment_input + workflow_tasks; approvals; conditions; gates;
  options; policies; artifacts; triggers; variables + groups). The demo sets a
  non-default value on essentially every field.
- **`betterado_release_folder`** — organises pipelines into a folder tree.
- **`betterado_release_definition_permissions`** — ACL bits on the pipeline. The
  demo sets the four keys the live acceptance suite verifies (`ViewReleases`,
  `EditReleaseEnvironment`, `DeleteReleases`, `CreateReleases`); the namespace
  exposes more actions, added here only once a live test covers them.
- **`betterado_task_group`** — reusable steps, with the v0.2.0 input metadata
  (`icon_url`, `visible_rule`, `properties`, `aliases`).

---

## ④ How it's proven

| Gate | Status | How |
|---|---|---|
| Schema correctness | ✅ proven, creds-free | `terraform validate` against the built v0.2.0 provider — passes clean |
| Live round-trip (per feature) | ✅ proven | acceptance suite, `TF_ACC=1`: apply → REST `GET` → idempotency → destroy |
| Live round-trip (this combined config) | ⟳ one command | `./refresh-evidence.sh` → applies, asserts clean re-plan, writes `evidence/*.json` |

The acceptance tests capture live REST evidence via `testutils.CaptureLiveEvidence`
— the merge decision is made on a real API `GET` of the created resource, never a
table of test names.

---

## ⑤ Known limitation (by decision)

`stages` and the other nested collections use **block** syntax (`stages { … }`),
not assignable arrays (`stages = [ { … } ]`). The SDKv2 `ConfigMode:Attr` route to
arrays was tried and reverted — it forces consumers to null-fill every nested
attribute at every level (worse ergonomics, the opposite of the goal). Clean
array-with-optional syntax requires a holistic migration to the
terraform-plugin-framework, tracked in `roadmap.md` § Future.
