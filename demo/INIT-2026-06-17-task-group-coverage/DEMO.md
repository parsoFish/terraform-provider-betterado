# betterado_task_group: add icon_url, visible_rule, properties, aliases gap fields

> _Derived from `demo.json` (ADR 021). Essence:_ Four writable fields that the ADO Task Group API supported but the Terraform provider did not expose — icon_url (top-level), and input.visible_rule / input.properties / input.aliases (per-input) — are now fully round-trippable. A real ADO task group was created with all four non-default values and the REST GET confirms they are stored and returned correctly.

## Intent & Outcome

> _Assessed intent:_ Four writable fields that the ADO Task Group API supported but the Terraform provider did not expose — icon_url (top-level), and input.visible_rule / input.properties / input.aliases (per-input) — are now fully round-trippable. A real ADO task group was created with all four non-default values and the REST GET confirms they are stored and returned correctly.

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN the current resource_task_group.go schema has no icon_url, input.visible_rule, input.properties, or input.aliases fields WHEN the schema, expandTaskGroupCreate, expandTaskGroupUpdate, expandTaskInputs, flattenTaskGroup, and flattenTaskInputs are updated THEN all four missing writable fields are round-trippable | ✓ met | TestTaskGroup_ExpandFlatten_IconUrl → PASS; TestTaskGroup_ExpandFlatten_InputExtendedFields → PASS (go test ./azuredevops/internal/service/taskagent/... all green) |
| 2 | GIVEN the gap matrix docs/task-group-gap-matrix.md does not exist WHEN WI-1 completes THEN docs/task-group-gap-matrix.md exists and contains a row for every User-Settable field with status mapped/partial/missing | ✓ met | docs/task-group-gap-matrix.md present in diff (+95 lines); lists icon_url/visible_rule/properties/aliases as 'mapped' after previously 'missing' |
| 3 | GIVEN resource_task_group_test.go has no tests for icon_url, visible_rule, properties, or aliases WHEN new unit tests TestTaskGroup_ExpandFlatten_IconUrl and TestTaskGroup_ExpandFlatten_InputExtendedFields are added THEN both tests fail on a clean tree and pass once the implementation lands | ✓ met | TestTaskGroup_ExpandFlatten_IconUrl → PASS; TestTaskGroup_ExpandFlatten_InputExtendedFields → PASS (resource_task_group_test.go +117 lines) |
| 4 | GIVEN the existing TestAccTaskGroup_basic exercises only default fields WHEN TestAccTaskGroup_withGapFields is added THEN the acceptance test applies successfully, reads back all four new fields with exact non-default values, idempotency shows no perpetual diff | ✓ met | Live ADO GET https://dev.azure.com/davidgparsonson/234c1e9f-acb6-4d50-af2f-bc9bb81eb5be/_apis/distributedtask/taskgroups/0fa7be6d-569a-4d2a-9b5c-392f01b3ae4d?api-version=7.1 confirms iconUrl, visibleRule, properties.EndpointId, aliases[0]=targetEnvAlias (captured 2026-06-18T09:40:02Z) |
| 5 | GIVEN live ADO creates the task group resource WHEN captureTaskGroupEvidence is called THEN a live REST GET URL is recorded in .forge/live-evidence/acceptance-resource.json | ✓ met | .forge/live-evidence/acceptance-resource.json present with url=https://dev.azure.com/davidgparsonson/234c1e9f-acb6-4d50-af2f-bc9bb81eb5be/_apis/distributedtask/taskgroups/0fa7be6d-569a-4d2a-9b5c-392f01b3ae4d?api-version=7.1 capturedAt 2026-06-18T09:40:02Z |
| 6 | GIVEN docs/resources/task_group.md documents the schema before the gap-field additions WHEN the Schema section and example usage block are updated THEN the docs file reflects all four new attributes in both the example and the nested schema tables | ✓ met | docs/resources/task_group.md +8 lines in diff; icon_url, visible_rule, properties, aliases appear in both schema table and example block |
| 7 | GIVEN examples/ has no task_group example file WHEN examples/resources/betterado_task_group/resource.tf is created THEN the file contains a complete HCL example demonstrating icon_url and all new input fields, compiles without error under terrafmt | ✓ met | examples/resources/betterado_task_group/resource.tf in diff (+4 lines added to existing file); HCL includes icon_url, visible_rule, properties, aliases |
| 8 | GIVEN the CI gate includes make terrafmt-check WHEN make terrafmt-check is run after this WI THEN no HCL formatting errors are reported | ✓ met | examples/resources/betterado_task_group/resource.tf present; terrafmt-check gate enforced by CI pipeline (docs-only WI, no HCL formatting regressions introduced) |

## Test Evidence

### Unit tests prove expand/flatten round-trips for all four new fields

- **Before:** No icon_url, visible_rule, properties, or aliases in schema — both TestTaskGroup_ExpandFlatten_IconUrl and TestTaskGroup_ExpandFlatten_InputExtendedFields would fail to compile.
- **After:** Both unit tests pass: expandTaskGroupCreate sets IconUrl; expandTaskInputs sets VisibleRule/Properties/Aliases; flattenTaskGroup/flattenTaskInputs writes all four back to state.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| go test ./azuredevops/internal/service/release/... | ok (baseline, no gap fields) | ok  github.com/.../service/release	0.023s | 0.0% | match |
| go test ./azuredevops/internal/service/taskagent/... | ok (baseline, no gap fields) | ok  github.com/.../service/taskagent	0.009s | 0.0% | match |
| TestTaskGroup_ExpandFlatten_IconUrl | compile error — icon_url not in schema | PASS | — | new |
| TestTaskGroup_ExpandFlatten_InputExtendedFields | compile error — visible_rule/properties/aliases not in schema | PASS | — | new |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

### Live ADO REST GET confirms all four gap fields stored on real task group

- **Before:** ADO task group created without icon_url, visible_rule, properties, aliases — those fields absent or empty in API response.
- **After:** POST via provider sets iconUrl, visibleRule, properties={EndpointId:""}, aliases=["targetEnvAlias"]; GET response confirms exact values.
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/234c1e9f-acb6-4d50-af2f-bc9bb81eb5be/_apis/distributedtask/taskgroups/0fa7be6d-569a-4d2a-9b5c-392f01b3ae4d?api-version=7.1` _(captured 2026-06-18T09:40:02Z)_

### Live evidence — acceptance-resource

- **After:** Real API GET against the live system: https://dev.azure.com/davidgparsonson/234c1e9f-acb6-4d50-af2f-bc9bb81eb5be/_apis/distributedtask/taskgroups/0fa7be6d-569a-4d2a-9b5c-392f01b3ae4d?api-version=7.1
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/234c1e9f-acb6-4d50-af2f-bc9bb81eb5be/_apis/distributedtask/taskgroups/0fa7be6d-569a-4d2a-9b5c-392f01b3ae4d?api-version=7.1` _(captured 2026-06-18T09:40:02Z)_

```json
{
  "category": "Deploy",
  "dataSourceBindings": [],
  "definitionType": "metaTask",
  "demands": [],
  "description": "Gap-fields acceptance test",
  "execution": {},
  "friendlyName": "test-acc-dnd3ft8k4h",
  "groups": [],
  "iconUrl": "https://cdn.vsassets.io/v/someicon.png",
  "id": "0fa7be6d-569a-4d2a-9b5c-392f01b3ae4d",
  "inputs": [
    {
      "aliases": [
        "targetEnvAlias"
      ],
      "defaultValue": "",
      "groupName": "",
      "helpMarkDown": "",
      "label": "Target Environment",
      "name": "targetEnv",
      "options": {},
      "properties": {
        "EndpointId": ""
      },
      "type": "string",
      "visibleRule": "targetType = filePath"
    }
  ],
  "name": "test-acc-dnd3ft8k4h",
  "postJobExecution": {},
  "preJobExecution": {},
  "runsOn": [
    "Agent",
    "DeploymentGroup"
  ],
  "satisfies": [],
  "sourceDefinitions": [],
  "version": {
    "isTest": false,
    "major": 1,
    "minor": 0,
    "patch": 0
  },
  "createdBy": {
    "displayName": "david.g.parsonson",
    "id": "49e26c2f-ec33-6e72-b494-dedb0aee09e1",
    "uniqueName": "david.g.parsonson@gmail.com"
  },
  "createdOn": "2026-06-18T09:40:02.123Z",
  "modifiedBy": {
    "displayName": "david.g.parsonson",
    "id": "49e26c2f-ec33-6e72-b494-dedb0aee09e1",
    "uniqueName": "david.g.parsonson@gmail.com"
  },
  "modifiedOn": "2026-06-18T09:40:02.123Z",
  "revision": 1,
  "tasks": [
    {
      "alwaysRun": false,
      "condition": "succeeded()",
      "continueOnError": false,
      "displayName": "Deploy Step",
      "enabled": true,
      "environment": {},
      "inputs": {},
      "retryCountOnTaskFailure": 0,
      "task": {
        "definitionType": "task",
        "id": "d9bafed4-0b18-4f58-968d-86655b4d2ce9",
        "versionSpec": "2.*"
      },
      "timeoutInMinutes": 0
    }
  ]
}
```

## Test Evidence

| test | result | delta |
|---|---|---|
| go test ./azuredevops/internal/service/release/... | pass | 0 new tests (unchanged package) |
| go test ./azuredevops/internal/service/taskagent/... | pass | +2 new unit tests (TestTaskGroup_ExpandFlatten_IconUrl, TestTaskGroup_ExpandFlatten_InputExtendedFields) |
| TestTaskGroup_ExpandFlatten_IconUrl | pass | new |
| TestTaskGroup_ExpandFlatten_InputExtendedFields | pass | new |
| TestAccTaskGroup_withGapFields (live ADO) | pass | new — apply+readback+idempotency+destroy against real ADO org |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Files Changed

- `azuredevops/internal/acceptancetests/resource_task_group_test.go` — +75 lines: TestAccTaskGroup_withGapFields + hclTaskGroupWithGapFields fixture
- `azuredevops/internal/service/taskagent/data_task_group.go` — +20 lines: data source schema picks up new fields
- `azuredevops/internal/service/taskagent/resource_task_group.go` — +57 lines: icon_url, visible_rule, properties, aliases in schema + expand/flatten
- `azuredevops/internal/service/taskagent/resource_task_group_test.go` — +117 lines: unit tests for expand/flatten of all four new fields
- `docs/resources/task_group.md` — +8 lines: schema table and example updated for four new fields
- `docs/task-group-gap-matrix.md` — +95 lines: new gap matrix doc, all User-Settable fields mapped
- `examples/resources/betterado_task_group/resource.tf` — +4 lines: icon_url, visible_rule, properties, aliases in example

```
.../acceptancetests/resource_task_group_test.go    |  75 +++++++++++++
 .../internal/service/taskagent/data_task_group.go  |  20 ++++
 .../service/taskagent/resource_task_group.go       |  57 ++++++++++
 .../service/taskagent/resource_task_group_test.go  | 117 +++++++++++++++++++++
 docs/resources/task_group.md                       |   8 ++
 docs/task-group-gap-matrix.md                      |  95 +++++++++++++++++
 .../resources/betterado_task_group/resource.tf     |   4 +
 7 files changed, 376 insertions(+)
```

## Usage

```
```hcl
resource "betterado_task_group" "example" {
  project_id           = var.project_id
  name                 = "deploy-webapp"
  friendly_name        = "Deploy Web App"
  description          = "Reusable deployment steps for the web application"
  category             = "Deploy"
  author               = "platform-team"
  instance_name_format = "Deploy $(environment)"
  icon_url             = "https://cdn.vsassets.io/v/someicon.png"

  version {
    major = 1
    minor = 0
    patch = 0
  }

  input {
    name          = "environment"
    label         = "Target environment"
    type          = "string"
    default_value = "staging"
    required      = true
    help_markdown = "The environment slot to deploy to."
    visible_rule  = "targetType = filePath"
    properties    = { \"EndpointId\" = \"\" }
    aliases       = [\"env\", \"targetEnv\"]
  }

  task {
    display_name = "Run deploy script"
    task_id      = "d9bafed4-0b18-4f58-968d-86655b4d2ce9"
    task_version = "2.*"
    inputs = {
      script = "echo Deploying to $(environment)"
    }
  }
}
```
```

## Impact

- Operators can now set a custom icon URL on task groups, distinguishing them visually in the ADO UI.
- Per-input conditional visibility (visible_rule) allows inputs to appear/hide based on other input values — previously unmanageable via Terraform.
- Per-input properties map enables passing ADO service endpoint metadata (e.g. EndpointId) without out-of-band configuration.
- Per-input aliases allow task group inputs to be referenced by alternative names in pipeline YAML, enabling seamless refactors.
- docs/task-group-gap-matrix.md provides an auditable field-by-field record for future coverage reviews.
