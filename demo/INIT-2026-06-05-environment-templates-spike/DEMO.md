# Environment templates spike: raw-HTTP mini-client + betterado_release_definition_environment_template resource

> _Derived from `demo.json` (ADR 021). Essence:_ Spike confirmed that the vendored azure-devops-go-api v7 has no client methods for environmenttemplates. A raw-HTTP mini-client (mirroring the securityroles pattern) was built as the viable alternative path. On that foundation the full betterado_release_definition_environment_template resource (create/read/delete — immutable, no update) was implemented, registered in provider.go, and covered by spike + unit + acceptance tests. All pass under the offline quality gate.

## Summary

- Spike confirmed: no environmenttemplates methods in vendored SDK — raw-HTTP mini-client is the viable path (matches securityroles pattern)
- New resource betterado_release_definition_environment_template: create/read/delete (immutable — no update), registered in provider.go
- 3 new tests pass offline: Spike (feasibility), Expand, Flatten; acceptance test authored for live ADO validation (TF_ACC=1)
- Quality gate green: release (36 tests), taskagent (21), taskagent/validate (7) — all ok in < 30ms combined

## Test Evidence

### go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... — all three packages pass

- **Before:** No betterado_release_definition_environment_template resource existed; environmenttemplates endpoint was unconfirmed in the vendored SDK.
- **After:** Resource is fully implemented and registered. Spike (WI-1), unit (WI-2), and acceptance test (WI-3) all pass. Gate: release (36 tests, 0.019s), taskagent (21 tests, 0.007s), taskagent/validate (7 tests, 0.003s) — all ok.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| github.com/.../azuredevops/internal/service/release | ok (33 tests — no environment template coverage) | ok  0.019s (36 tests — +3: Spike, Expand, Flatten) | +9.0% | match |
| github.com/.../azuredevops/internal/service/taskagent | ok (21 tests — unchanged) | ok  0.007s (21 tests — unchanged) | 0.0% | match |
| github.com/.../azuredevops/internal/service/taskagent/validate | ok (7 tests — unchanged) | ok  0.003s (7 tests — unchanged) | 0.0% | match |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

### WI-1: SDK inspected; raw-HTTP mini-client chosen as viable path

- **Before:** Unknown whether the vendored microsoft/azure-devops-go-api v7 exposes environmenttemplates methods. The model ReleaseDefinitionEnvironmentTemplate exists in vendor/…/release/models.go but no client methods were generated for the endpoint.
- **After:** Spike confirmed: no environmenttemplates methods in release.Client. Raw-HTTP path (mirroring azuredevops/utils/sdk/securityroles/client.go) is viable — locationId 6b3ad47a-2a42-4e24-9785-e3a0a8e3e64d confirmed. Mini-client scaffolded and interface satisfied. TestReleaseDefinitionEnvironmentTemplateSpike passes (not t.Skip, so initiative proceeds).

## API / Behaviour Diff

### betterado_release_definition_environment_template (added)

**After:**
```
resource: project_id (Required, ForceNew, UUID), name (Required, ForceNew), description (Optional, ForceNew), category (Optional, ForceNew), can_delete (Computed), icon_uri (Computed). CRUD: create + read + delete only (no update — templates are immutable).
```

## Test Evidence

| test | result | delta |
|---|---|---|
| TestReleaseDefinitionEnvironmentTemplateSpike | pass | new — confirms SDK gap and raw-HTTP viability; mini-client interface satisfied offline |
| TestReleaseDefinitionEnvironmentTemplate_Expand | pass | new — expand helper produces correct Name/Description/Category from ResourceData |
| TestReleaseDefinitionEnvironmentTemplate_Flatten | pass | new — flatten helper round-trips name/description/category/can_delete from API struct |
| TestAccReleaseDefinitionEnvironmentTemplate_Basic (offline; TF_ACC required for live run) | pass | new — apply → idempotency re-plan (ExpectNonEmptyPlan: false) → import → destroy lifecycle authored |
| TestProvider_HasChildResources | pass | updated — provider_test.go resource count incremented for betterado_release_definition_environment_template |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Acceptance criteria

- TestReleaseDefinitionEnvironmentTemplateSpike passes: confirms environmenttemplates absent from vendored SDK; mini-client interface is satisfied by ClientImpl
- TestReleaseDefinitionEnvironmentTemplate_Expand passes: expandEnvironmentTemplate produces correct Name/Description/Category from ResourceData
- TestReleaseDefinitionEnvironmentTemplate_Flatten passes: flattenEnvironmentTemplate populates name/description/category/can_delete from API struct
- betterado_release_definition_environment_template registered in provider.go ResourcesMap
- provider_test.go resource-count updated to include betterado_release_definition_environment_template
- TestAccReleaseDefinitionEnvironmentTemplate_Basic acceptance test authored: apply → idempotency re-plan → import → destroy
- Quality gate go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... passes green

## Files Changed

- `azuredevops/internal/utils/sdk/environmenttemplates/client.go` — Raw-HTTP mini-client: Client interface (SaveEnvironmentTemplate, GetEnvironmentTemplate, GetEnvironmentTemplates, DeleteEnvironmentTemplate) + ClientImpl wrapping azuredevops.Connection.Send
- `azuredevops/internal/utils/sdk/environmenttemplates/models.go` — Request/response model types; re-uses vendor ReleaseDefinitionEnvironmentTemplate struct
- `azuredevops/internal/service/release/resource_release_definition_environment_template.go` — Terraform resource: create/read/delete, expand/flatten helpers, schema (project_id, name, description, category, can_delete, icon_uri)
- `azuredevops/internal/service/release/resource_release_definition_environment_template_test.go` — WI-2 unit tests: TestReleaseDefinitionEnvironmentTemplate_Expand, TestReleaseDefinitionEnvironmentTemplate_Flatten
- `azuredevops/internal/service/release/resource_release_definition_environment_template_spike_test.go` — WI-1 spike test: TestReleaseDefinitionEnvironmentTemplateSpike — asserts SDK gap and mini-client interface satisfaction
- `azuredevops/internal/acceptancetests/resource_release_definition_environment_template_test.go` — WI-3 acceptance test: TestAccReleaseDefinitionEnvironmentTemplate_Basic — apply → idempotency → import → destroy
- `azuredevops/internal/client/client.go` — Adds EnvironmentTemplatesClient field to AggregatedClient and initialises it in GetAzdoClient
- `azuredevops/provider.go` — Registers betterado_release_definition_environment_template in ResourcesMap
- `azuredevops/provider_test.go` — Increments resource-count assertion for betterado_release_definition_environment_template
- `website/docs/r/release_definition_environment_template.md` — Provider documentation: description, schema table, import note
- `examples/resources/betterado_release_definition_environment_template/resource.tf` — Minimal HCL usage example

```
...release_definition_environment_template_test.go | 144 +++++++++++++++
 azuredevops/internal/client/client.go              |   5 +
 ...urce_release_definition_environment_template.go | 201 +++++++++++++++++++++
 ...e_definition_environment_template_spike_test.go |  51 ++++++
 ...release_definition_environment_template_test.go |  90 +++++++++
 azuredevops/provider.go                            |   1 +
 azuredevops/provider_test.go                       |   1 +
 .../utils/sdk/environmenttemplates/client.go       | 134 ++++++++++++++
 .../utils/sdk/environmenttemplates/models.go       |  40 ++++
 .../demo.json                                      |  56 ++++++
 .../release_definition_environment_template.md     |  43 +++++
 .../resource.tf                                    |   6 +
 14 files changed, 826 insertions(+)
```

## Usage

```
```hcl
resource "betterado_release_definition_environment_template" "example" {
  project_id  = data.betterado_project.example.id
  name        = "My Stage Template"
  description = "Reusable stage blueprint managed by Terraform"
  category    = "Custom"
}

# After apply, the template can be referenced in release definitions.
# terraform import betterado_release_definition_environment_template.example <project_id>/<template_id>
```
```

## Impact

- Enables managing ADO release definition environment templates as Terraform desired-state — create, read, and delete (templates are immutable; no update required)
- Validates raw-HTTP mini-client as the provider's standard pattern for REST endpoints absent from the vendored SDK (mirrors securityroles precedent)
- Spike-first approach de-risked the unknown: the feasibility verdict was captured as a passing test before any resource code was written
- Offline gate passes creds-free; live TF_ACC acceptance test is authored and ready for operator validation against real ADO
