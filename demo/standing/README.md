# Standing demo — the permanent betterado v0.2.0 showcase

A single, applyable Terraform configuration that exercises **every feature shipped
in betterado v0.2.0** across all four release-surface resource types. It is the
canonical, living reference for the provider: copy it, point it at your org, and
you have a working pipeline that touches the whole surface.

Unlike the per-initiative demos under `demo/INIT-*/` (each proves one cycle's
slice), this is the **standing** demo — comprehensive, kept current, and meant to
be left deployed for review.

## What it showcases

| File | Resource | Features exercised |
|---|---|---|
| `release-definition.tf` | `betterado_release_definition` | **`stages`** (the rename from `environment`); the full `triggers` block — `cd_artifact_trigger` (branch + tag filters, `create_release_on_build_tagging`), `schedule_trigger`, `source_repo_trigger`, **`container_image_trigger`**; `workflow_task` **`timeout_in_minutes` + `retry_count_on_task_failure`**; pre/post **`*_deployment_gates`**; `deployment_input` **`override_inputs` + `parallel_execution`**; **`environment_trigger`**; approvals + `approval_options`, conditions, variables, env options, execution + retention policy |
| `release-folder.tf` | `betterado_release_folder` | both writable fields (`path`, `description`) |
| `permissions.tf` | `betterado_release_definition_permissions` | the four **live-proven** ACL keys (`ViewReleases`, `EditReleaseEnvironment`, `DeleteReleases`, `CreateReleases`) |
| `task-group.tf` | `betterado_task_group` | `icon_url` + input metadata: `visible_rule`, `properties`, `aliases`, `group_name` |
| `data-sources.tf` | data sources | read-back of the definition, folder, and task group |

Bold = a coverage gap closed or the headline rename during the 2026-06 capstone.

## Run it

Published provider:

```bash
cp terraform.tfvars.example terraform.tfvars   # then edit with your org values
terraform init
terraform apply
```

Local/dev build (no registry download — used to validate against an unreleased
schema):

```bash
go build -mod=vendor -o /tmp/tf-betterado-devbin/terraform-provider-betterado .  # from repo root
cat > /tmp/dev.tfrc <<'EOF'
provider_installation {
  dev_overrides { "parsoFish/betterado" = "/tmp/tf-betterado-devbin" }
  direct {}
}
EOF
export TF_CLI_CONFIG_FILE=/tmp/dev.tfrc
terraform validate          # schema-only — no creds, no API calls
terraform apply             # needs creds + terraform.tfvars
```

## Refresh the live evidence

`./refresh-evidence.sh` (creds from the repo's gitignored `secrets.env`) applies
the showcase, asserts a clean idempotency re-plan, then GETs each resource from
the live REST API into `evidence/*.json` — the round-trip proof. It leaves the
resources **standing** for portal review; run `terraform destroy` when done.

## Validation status

- **Schema-valid** against the built v0.2.0 provider — `terraform validate` passes
  with zero errors (run in CI-equivalent form with the dev build above). This is
  the creds-free correctness gate every commit to this dir must hold.
- **Live round-trip** of each individual feature is proven by the acceptance
  suite (`azuredevops/internal/acceptancetests/`, `TF_ACC=1`) — apply → API GET →
  idempotency → destroy. `refresh-evidence.sh` reproduces that for this combined
  config and captures the evidence here.
