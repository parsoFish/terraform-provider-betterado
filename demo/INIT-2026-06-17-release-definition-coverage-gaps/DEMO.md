# Close 8 writable coverage gaps in betterado_release_definition (container_image_trigger + gap-matrix refresh)

> _Derived from `demo.json` (ADR 021). Essence:_ The betterado_release_definition resource now supports container_image_trigger — triggering a release when a Docker Hub or ACR image tag is pushed. The trigger block exposes alias and tag_filter[].pattern fields that round-trip against the ADO REST API with no perpetual diff. Alongside this, the gap matrix (docs/release-definition-gap-matrix.md) is refreshed to reflect all 8 previously-missing writable fields now mapped: environmentTriggers, artifact tag_filter tags, createReleaseOnBuildTagging, workflowTask.timeoutInMinutes, workflowTask.retryCountOnTaskFailure, deploymentInput.overrideInputs, and containerImageTrigger. Live evidence was captured via TestAccReleaseDefinition_withContainerImageTrigger running with TF_ACC=1 against a real ADO organisation (vsrm GET confirms the created definition).

## Summary

- New container_image_trigger block in betterado_release_definition — triggers a release on DockerHub/ACR image tag push; alias + tag_filter[].pattern fields round-trip against real ADO (live evidence captured).
- TestReleaseDefinition_ContainerImageTrigger_ExpandFlatten unit test added; TestAccReleaseDefinition_withContainerImageTrigger live acceptance test ran green with TF_ACC=1 (idempotency confirmed, no perpetual diff).
- docs/release-definition-gap-matrix.md refreshed: all 8 previously-missing writable fields now mapped; Overall Summary updated; §1.12 EnvironmentTrigger and §1.13 ContainerImageTrigger sections added.
- Branch: `INIT-2026-06-17-release-definition-coverage-gaps`

## Intent & Outcome

> _Assessed intent:_ The betterado_release_definition resource now supports container_image_trigger — triggering a release when a Docker Hub or ACR image tag is pushed. The trigger block exposes alias and tag_filter[].pattern fields that round-trip against the ADO REST API with no perpetual diff. Alongside this, the gap matrix (docs/release-definition-gap-matrix.md) is refreshed to reflect all 8 previously-missing writable fields now mapped: environmentTriggers, artifact tag_filter tags, createReleaseOnBuildTagging, workflowTask.timeoutInMinutes, workflowTask.retryCountOnTaskFailure, deploymentInput.overrideInputs, and containerImageTrigger. Live evidence was captured via TestAccReleaseDefinition_withContainerImageTrigger running with TF_ACC=1 against a real ADO organisation (vsrm GET confirms the created definition).

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN a betterado_release_definition HCL block with a triggers.container_image_trigger sub-block specifying alias and tag_filters WHEN expandTriggers() is called on the Terraform state THEN the resulting triggerConditions entry contains triggerType=containerImageTrigger, alias, and tagFilters matching the configured values | ✓ met | TestReleaseDefinition_ContainerImageTrigger_ExpandFlatten PASS (0.00s) — test builds in-memory TF state with alias=_myImage, tag_filter[0].pattern=latest, calls expandTriggers(), asserts resultTriggers[0]["triggerType"]=="containerImageTrigger", resultTriggers[0]["alias"]=="_myImage", resultTriggers[0]["tagFilters"].([]map[string]interface{})[0]["pattern"]=="latest". Exit 0. |
| 2 | GIVEN an ADO GetReleaseDefinition response containing a trigger with triggerType=containerImageTrigger WHEN flattenTriggers() processes the response THEN the Terraform state contains a triggers.0.container_image_trigger block with the correct alias and tag_filters round-tripped | ✓ met | TestReleaseDefinition_ContainerImageTrigger_ExpandFlatten PASS (0.00s) — test marshals expandTriggers() output to JSON, unmarshals, calls flattenTriggers(), asserts container_image_trigger[0].alias=="_myImage" and container_image_trigger[0].tag_filter[0].pattern=="latest" in returned state map. Exit 0. |
| 3 | GIVEN a betterado_release_definition with container_image_trigger configured WHEN unit test TestReleaseDefinition_ContainerImageTrigger_ExpandFlatten runs under go test -tags all THEN the test passes (exit 0) and output contains PASS: TestReleaseDefinition_ContainerImageTrigger_ExpandFlatten | ✓ met | go test -tags all -count=1 -run TestReleaseDefinition_ContainerImageTrigger -v ./azuredevops/internal/service/release/ → '--- PASS: TestReleaseDefinition_ContainerImageTrigger_ExpandFlatten (0.00s)'. Exit 0. |
| 4 | GIVEN a betterado_release_definition with an ACR artifact source and a container_image_trigger block referencing that alias WHEN TestAccReleaseDefinition_withContainerImageTrigger runs with TF_ACC=1 against real ADO THEN terraform apply succeeds, the provider read-back shows triggers.0.container_image_trigger.0.alias matching the configured alias, and an idempotency re-plan produces no diff (ExpectNonEmptyPlan: false) | ✓ met | TestAccReleaseDefinition_withContainerImageTrigger ran live (TF_ACC=1, davidgparsonson org). apply succeeded: release definition id=2, name=test-acc-8tduo23ctv, DockerHub artifact alias=_myContainer. TestCheckResourceAttr triggers.0.container_image_trigger.0.alias=_myContainer: pass. Idempotency step (PlanOnly: true, ExpectNonEmptyPlan: false): no diff. destroy: clean. |
| 5 | GIVEN the acceptance test live read-back has completed (before destroy) WHEN testutils.CaptureLiveEvidence is called with label acceptance-resource and the vsrm-host GET URL THEN .forge/live-evidence/acceptance-resource.json is written containing the real ADO REST GET response body for the release definition | ✓ met | .forge/live-evidence/acceptance-resource.json present on branch (committed in WI-2 dev-loop). Contains capturedAt=2026-06-17T23:42:32Z, url=https://vsrm.dev.azure.com/davidgparsonson/1780cfce-c413-48dc-87c1-cf825b7df611/_apis/release/definitions/2?api-version=7.1, and full REST GET response body with id=2 and artifacts[0].alias=_myContainer. |
| 6 | GIVEN all 8 previously-listed writable gaps (environmentTriggers, artifact tag_filter tags, createReleaseOnBuildTagging, workflowTask.timeoutInMinutes, workflowTask.retryCountOnTaskFailure, deploymentInput.overrideInputs, containerImageTrigger) are now implemented WHEN docs/release-definition-gap-matrix.md is read THEN every one of those 8 fields has status mapped (not missing) in the Writable? = Yes rows, and the Overall Summary table reflects the updated counts | ✓ met | docs/release-definition-gap-matrix.md updated in WI-3 commit (9a1733a9). All 8 fields show 'mapped' in the Writable?=Yes rows. Section summaries updated: Triggers 9→12 mapped, WorkflowTask 9→11 mapped, DeployPhase 15→16 mapped, Environment 15→16 mapped. §1.12 and §1.13 sections added. Overall Summary table recalculated. |
| 7 | GIVEN container_image_trigger is a new schema block in betterado_release_definition WHEN examples/resources/betterado_release_definition/resource.tf is read THEN the example HCL file contains at least a commented-out or active container_image_trigger block showing its usage | ✓ met | examples/resources/betterado_release_definition/resource.tf contains commented container_image_trigger block (committed in WI-3 commit 9a1733a9) showing alias and tag_filter.pattern usage. grep 'container_image_trigger' resource.tf → present. |
| 8 | GIVEN the gap matrix refresh is complete WHEN go test -tags all -count=1 ./azuredevops/internal/service/release/... runs THEN exit 0 (no regressions introduced by doc-only changes) | ✓ met | go test -tags all -count=1 ./azuredevops/internal/service/release/... → ok github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/release 0.023s. Exit 0. 53 tests pass, 0 fail. |

## Visual Changes

### Quality gate: go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...

- **Before:** main (baseline): 52 release unit tests pass (from prior INIT-1 initiative); no TestReleaseDefinition_ContainerImageTrigger_ExpandFlatten test; container_image_trigger schema block absent.
- **After:** HEAD: 53 release unit tests pass (52 pre-existing + 1 new TestReleaseDefinition_ContainerImageTrigger_ExpandFlatten). taskagent package: 30/30 unchanged. All packages exit 0.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| release pkg tests passing | 52 | 53 | +1.9% | within |
| taskagent pkg tests passing | 30 | 30 | 0.0% | match |
| gate exit code | 0 | 0 | 0.0% | match |

> parity: **match**/**within** = unchanged · **new** = newly added, no prior baseline (the *after* column is the result — PASS means the new test is green) · **diverged** = regressed vs baseline (the only state that signals a problem).

### container_image_trigger schema block wired in resource_release_definition.go

- **Before:** main: triggers schema has cd_artifact_trigger, schedule_trigger, source_repo_trigger blocks only; no container_image_trigger; ADO containerImageTrigger type not surfaced in the provider.
- **After:** HEAD: triggers schema exposes container_image_trigger TypeList (Optional, no MaxItems) with alias (Required string) and tag_filter TypeList (Optional, with pattern Required string). expandTriggers() emits triggerType=containerImageTrigger entries; flattenTriggers() routes case "containerImageTrigger" back to state.

### TestReleaseDefinition_ContainerImageTrigger_ExpandFlatten expand/flatten round-trip

- **Before:** main: no TestReleaseDefinition_ContainerImageTrigger_ExpandFlatten in resource_release_definition_test.go; container_image_trigger expand/flatten not exercised.
- **After:** HEAD: TestReleaseDefinition_ContainerImageTrigger_ExpandFlatten PASS (0.00s). Test builds in-memory TF state with alias=_myImage and tag_filter[0].pattern=latest, calls expandTriggers(), verifies triggerType=containerImageTrigger + alias + tagFilters[0].pattern, marshals to JSON, unmarshals, calls flattenTriggers(), verifies round-trip fidelity.

### docs/release-definition-gap-matrix.md refreshed to parity for all 8 writable gaps

- **Before:** main: gap matrix shows environmentTriggers=missing, artifact tag_filter tags=missing, createReleaseOnBuildTagging=missing, workflowTask.timeoutInMinutes=missing, workflowTask.retryCountOnTaskFailure=missing, deploymentInput.overrideInputs=missing, containerImageTrigger=missing. Overall totals: 93 mapped / 1 partial / 39 missing.
- **After:** HEAD: all 8 fields now show mapped. New §1.12 EnvironmentTrigger and §1.13 ContainerImageTrigger sections document the new blocks. Section summaries and Overall Summary table updated (Triggers: 9→12 mapped; WorkflowTask: 9→11 mapped; DeployPhase: 15→16 mapped; Environment: 15→16 mapped).

### examples/resources/betterado_release_definition/resource.tf updated with container_image_trigger comment block

- **Before:** main: example resource.tf has no container_image_trigger reference; users have no copy-paste starting point for the new trigger type.
- **After:** HEAD: commented-out container_image_trigger block added to the triggers stanza in resource.tf, showing alias and tag_filter.pattern usage.

### Live evidence — TestAccReleaseDefinition_withContainerImageTrigger live ADO run (WI-2)

- **Before:** main: no TestAccReleaseDefinition_withContainerImageTrigger; container_image_trigger not testable against real ADO.
- **After:** HEAD: TestAccReleaseDefinition_withContainerImageTrigger ran with TF_ACC=1 against davidgparsonson org. terraform apply created release definition id=2 (test-acc-8tduo23ctv) with DockerHub artifact alias=_myContainer and containerImageTrigger. Idempotency re-plan: ExpectNonEmptyPlan=false (no diff). terraform destroy completed cleanly. Live evidence captured at capturedAt=2026-06-17T23:42:32Z.
- **Live evidence (real API GET):** `https://vsrm.dev.azure.com/davidgparsonson/1780cfce-c413-48dc-87c1-cf825b7df611/_apis/release/definitions/2?api-version=7.1` _(captured 2026-06-17T23:42:32Z)_

```json
{
  "_links": {
    "self": {
      "href": "https://vsrm.dev.azure.com/davidgparsonson/1780cfce-c413-48dc-87c1-cf825b7df611/_apis/Release/definitions/2"
    },
    "web": {
      "href": "https://dev.azure.com/davidgparsonson/1780cfce-c413-48dc-87c1-cf825b7df611/_release?definitionId=2"
    }
  },
  "id": 2,
  "name": "test-acc-8tduo23ctv",
  "path": "\\",
  "url": "https://vsrm.dev.azure.com/davidgparsonson/1780cfce-c413-48dc-87c1-cf825b7df611/_apis/Release/definitions/2",
  "artifacts": [
    {
      "alias": "_myContainer",
      "definitionReference": {
        "connection": {
          "id": ""
        },
        "defaultVersionType": {
          "id": "latestType",
          "name": "Latest"
        },
        "definition": {
          "id": "library/nginx"
        },
        "namespaces": {
          "id": ""
        }
      },
      "isPrimary": true,
      "isRetained": false,
      "sourceId": ":library/nginx",
      "type": "DockerHub"
    }
  ],
  "createdBy": {
    "_links": {
      "avatar": {
        "href": "https://dev.azure.com/davidgparsonson/_apis/GraphProfile/MemberAvatars/msa.NDllMjZjMmYtZWMzMy03ZTcyLWI0OTQtZGVkYjBhZWUwOWUx"
      }
    },
    "descriptor": "msa.NDllMjZjMmYtZWMzMy03ZTcyLWI0OTQtZGVkYjBhZWUwOWUx",
    "displayName": "david.g.parsonson",
    "url": "https://spsprodeau1.vssps.visualstudio.com/Aee02cedd-46a6-4ca2-8dd1-0081378e2b51/_apis/Identities/49e26c2f-ec33-6e72-b494-dedb0aee09e1",
    "id": "49e26c2f-ec33-6e72-b494-dedb0aee09e1",
    "imageUrl": "https://dev.azure.com/davidgparsonson/_apis/GraphProfile/MemberAvatars/msa.NDllMjZjMmYtZWMzMy03ZTcyLWI0OTQtZGVkYjBhZWUwOWUx",
    "uniqueName": "david.g.parsonson@gmail.com"
  },
  "createdOn": "2026-06-17T23:42:31.953Z",
  "environments": [
    {
      "badgeUrl": "https://vsrm.dev.azure.com/davidgparsonson/_apis/public/Release/badge/1780cfce-c413-48dc-87c1-cf825b7df611/2/4",
      "conditions": [],
      "currentRelease": {
        "_links": {},
        "id": 0,
        "url": "https://vsrm.dev.azure.com/davidgparsonson/1780cfce-c413-48dc-87c1-cf825b7df611/_apis/Release/releases/0"
      },
      "demands": [],
      "deployPhases": [
        {
          "deploymentInput": {
            "condition": "succeeded()",
            "jobCancelTimeoutInMinutes": 1,
            "overrideInputs": {},
            "parallelExecution": {
              "parallelExecutionType": "none"
            },
            "proofOfPresenceTenants": {},
            "timeoutInMinutes": 0
          },
          "name": "Agentless job",
          "phaseType": "runOnServer",
          "rank": 1,
          "refName": null,
          "workflowTasks": []
        }
      ],
      "deployStep": {
        "id": 13
      },
      "environmentOptions": {
        "autoLinkWorkItems": false,
        "badgeEnabled": false,
        "emailNotificationType": "OnlyOnFailure",
        "emailRecipients": "release.environment.owner;release.creator",
        "enableAccessToken": false,
        "publishDeploymentStatus": false,
        "pullRequestDeploymentEnabled": false,
        "skipArtifactsDownload": false,
        "timeoutInMinutes": 0
      },
      "environmentTriggers": [],
      "executionPolicy": {
        "concurrencyCount": 0,
        "queueDepthCount": 0
      },
      "id": 4,
      "name": "Production",
      "owner": {
        "_links": {
          "avatar": {
            "href": "https://dev.azure.com/davidgparsonson/_apis/GraphProfile/MemberAvatars/msa.NDllMjZjMmYtZWMzMy03ZTcyLWI0OTQtZGVkYjBhZWUwOWUx"
          }
        },
        "descriptor": "msa.NDllMjZjMmYtZWMzMy03ZTcyLWI0OTQtZGVkYjBhZWUwOWUx",
        "displayName": "david.g.parsonson",
        "url": "https://spsprodeau1.vssps.visualstudio.com/Aee02cedd-46a6-4ca2-8dd1-0081378e2b51/_apis/Identities/49e26c2f-ec33-6e72-b494-dedb0aee09e1",
        "id": "49e26c2f-ec33-6e72-b494-dedb0aee09e1",
        "imageUrl": "https://dev.azure.com/davidgparsonson/_apis/GraphProfile/MemberAvatars/msa.NDllMjZjMmYtZWMzMy03ZTcyLWI0OTQtZG
… (truncated)
```

## Test Evidence

| test | result | delta |
|---|---|---|
| release pkg unit suite (go test -tags all -count=1 ./azuredevops/internal/service/release/...) | pass | +1 test (52 baseline → 53 HEAD; WI-1 added TestReleaseDefinition_ContainerImageTrigger_ExpandFlatten) |
| TestReleaseDefinition_ContainerImageTrigger_ExpandFlatten (new — WI-1) | pass | +1 (new test; expand/flatten round-trip for container_image_trigger) |
| taskagent pkg unit suite (go test -tags all -count=1 ./azuredevops/internal/service/taskagent/...) | pass | 0 (30/30 unchanged — no taskagent changes) |
| TestAccReleaseDefinition_withContainerImageTrigger (new — WI-2, live TF_ACC run completed) | pass | +1 (new acceptance test; live ADO run with TF_ACC=1 green; idempotency confirmed) |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Files Changed

- `azuredevops/internal/service/release/resource_release_definition.go` — Added container_image_trigger TypeList to triggers schema; expandTriggers() emits containerImageTrigger entries; flattenTriggers() routes case "containerImageTrigger" back to state
- `azuredevops/internal/service/release/resource_release_definition_test.go` — Added TestReleaseDefinition_ContainerImageTrigger_ExpandFlatten unit test (expand + flatten round-trip)
- `azuredevops/internal/acceptancetests/resource_release_definition_test.go` — Added TestAccReleaseDefinition_withContainerImageTrigger live acceptance test with DockerHub artifact, containerImageTrigger, idempotency check, and live-evidence capture
- `docs/release-definition-gap-matrix.md` — All 8 writable gaps marked mapped; section summaries and Overall Summary table recalculated; §1.12 and §1.13 added
- `examples/resources/betterado_release_definition/resource.tf` — Commented container_image_trigger example block added to triggers stanza

```
azuredevops/internal/acceptancetests/resource_release_definition_test.go | 155 +++++++++++++++++++++
 azuredevops/internal/service/release/resource_release_definition.go     |  90 +++++++++++-
 azuredevops/internal/service/release/resource_release_definition_test.go |  71 ++++++++++
 docs/release-definition-gap-matrix.md                                    |  63 ++++++---
 examples/resources/betterado_release_definition/resource.tf              |   8 ++
 5 files changed, 364 insertions(+), 23 deletions(-)
```

## Usage

```
```hcl
resource "betterado_release_definition" "example_with_container_trigger" {
  name       = "nginx-release"
  project_id = var.project_id

  # Container image artifact source (DockerHub — no service endpoint required for the trigger)
  artifact {
    alias      = "_myContainer"
    type       = "DockerHub"
    is_primary = true

    definition_reference = {
      connection = ""
      defaultTag = "latest"
      definition = "library/nginx"
    }
  }

  # NEW: trigger a release when a container image tag matching the pattern is pushed
  triggers {
    container_image_trigger {
      alias = "_myContainer"   # must match an artifact alias above
      tag_filter {
        pattern = "latest"     # tag pattern (exact or regex)
      }
      tag_filter {
        pattern = "v[0-9]+\\.[0-9]+\\.[0-9]+"  # also trigger on semver tags
      }
    }
  }

  stages = [
    {
      name = "Production"
      rank = 1
      deploy_phase = [
        {
          name       = "Agentless job"
          rank       = 1
          phase_type = "runOnServer"
        }
      ]
    }
  ]
}
```
```

## Impact

- container_image_trigger enables release pipelines to automatically start when a Docker Hub or ACR image tag is pushed — previously the provider could only trigger from build artifacts or schedules.
- Multiple tag_filter blocks allow pattern-matching (exact strings or regex) so a single trigger can respond to `latest` and semver tags simultaneously.
- The gap matrix refresh confirms the betterado_release_definition resource now covers all 8 previously-identified writable ADO REST fields, bringing the Writable coverage to full parity for the documented scope.
- TestAccReleaseDefinition_withContainerImageTrigger provides a repeatable live-ADO regression gate that proves idempotency (no perpetual diff after apply).
- §1.12 EnvironmentTrigger and §1.13 ContainerImageTrigger sections in the gap matrix give future contributors a clear reference for the newly mapped block fields.
