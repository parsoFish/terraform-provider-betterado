# Pipelines v2 API — Gap Matrix

> **Initiative:** INIT-2026-07-01-new-api-pipelines-v2
> **Resource:** `betterado_pipeline`
> **API surface:** `_apis/pipelines` (Pipelines v2, `dev.azure.com`)
> **SDK:** `github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines`

---

## Field mapping

Every field returned by the Pipelines v2 REST API is documented below with its
Terraform attribute, mapping status, and rationale.

| Field | API path | Terraform attr | Status | Notes |
|---|---|---|---|---|
| id | `Pipeline.Id` | `id` (computed) | **Mapped** | Stored as a string resource ID. Integer from API. |
| name | `Pipeline.Name` | `name` (required) | **Mapped** | The pipeline display name. |
| folder | `Pipeline.Folder` | `folder` (optional/computed) | **Mapped** | Root folder defaults to `\`. |
| revision | `Pipeline.Revision` | `revision` (computed) | **Mapped** | Optimistic-concurrency token; updated on every write. |
| configuration.type | `Pipeline.Configuration.Type` | `configuration_type` (optional/computed) | **Mapped** | `ConfigurationType` enum: `yaml`, `designerJson`, `justInTime`, `designerHyphenJson`, `unknown`. Defaults to `yaml`. |
| url | `Pipeline.Url` | `url` (computed) | **Mapped** | Self-referential REST URL of the pipeline. |

### Fields listed in `CreatePipelineParameters`

| Field | API path | Terraform attr | Status | Notes |
|---|---|---|---|---|
| name | `CreatePipelineParameters.Name` | `name` | **Mapped** | Passed to Create. |
| folder | `CreatePipelineParameters.Folder` | `folder` | **Mapped** | Passed to Create. |
| configuration.type | `CreatePipelineParameters.Configuration.Type` | `configuration_type` | **Mapped** | Passed to Create. |

### Fields not surfaced (internal / read-only metadata)

| Field | API path | Rationale |
|---|---|---|
| `_links` | `Pipeline.Links` | HAL link bag — internal hypermedia, not a user-facing value. |

---

## Overlap analysis: `betterado_pipeline` vs `betterado_build_definition`

### API surface comparison

| Dimension | `betterado_pipeline` | `betterado_build_definition` |
|---|---|---|
| REST API | Pipelines v2 — `_apis/pipelines` | Build Definitions v1/v2 — `_apis/build/definitions` |
| Primary use case | YAML pipelines; Pipelines v2 object model | Classic build pipelines; extended trigger/queue/retention schema |
| SDK package | `azuredevops/v7/pipelines` | `azuredevops/v7/build` |
| Resource type name | `betterado_pipeline` | `betterado_build_definition` |

### Decision: **coexist — different API surface**

`betterado_build_definition` uses the Build API v1/v2 (`_apis/build/definitions`),
which manages **classic build pipelines** with the full Build Definition object model
(triggers, queue settings, retention, variables, steps).

`betterado_pipeline` uses the **Pipelines v2 API** (`_apis/pipelines`), which provides
a lighter-weight, modern object model designed primarily for YAML pipeline definitions.
Although Azure DevOps internally maps both API surfaces to the same underlying pipeline
concept, the two resources expose **different fields** and serve **different audiences**:

- **`betterado_pipeline` is preferred** for new YAML-first pipelines where users only
  need to register the pipeline definition and do not need to manage classic build
  settings (triggers, queues, retention schedules) via Terraform.
- **`betterado_build_definition` is retained** for classic build pipeline management and
  for users migrating from the upstream community provider, which exposes the full Build
  Definitions API surface.

There is **no functional overlap** at the schema level: `betterado_pipeline` exposes
`configuration_type`, `folder`, `revision`, and `url`; `betterado_build_definition`
exposes hundreds of additional fields (agents, pools, variables, steps, triggers, etc.)
that have no equivalent in the Pipelines v2 API. Both resources can coexist in a single
Terraform configuration without conflict.

---

## References

- [Pipelines v2 REST API reference](https://learn.microsoft.com/en-us/rest/api/azure/devops/pipelines/)
- [Build Definitions REST API reference](https://learn.microsoft.com/en-us/rest/api/azure/devops/build/definitions/)
- SDK source: `vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines/models.go`
- SDK source: `vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines/client.go`
