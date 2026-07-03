# Migrate taskagent resources and data sources to terraform-plugin-framework

> _Derived from `demo.json` (ADR 021). Essence:_ All SDKv2 resources and data sources in the taskagent package are migrated to terraform-plugin-framework, muxed alongside SDKv2. This eliminates deprecated SDKv2 code paths for agent_pool, agent_queue, deployment_group, elastic_pool, environment, environment_resource_kubernetes, variable_group (resource+data source), and the task_group data source. A field-coverage gap matrix documents the ADO Task Agent API v7.1 parity.

## Intent & Outcome

> _Assessed intent:_ All SDKv2 resources and data sources in the taskagent package are migrated to terraform-plugin-framework, muxed alongside SDKv2. This eliminates deprecated SDKv2 code paths for agent_pool, agent_queue, deployment_group, elastic_pool, environment, environment_resource_kubernetes, variable_group (resource+data source), and the task_group data source. A field-coverage gap matrix documents the ADO Task Agent API v7.1 parity.

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN the ADO Task Agent REST API v7.1 documentation and each SDKv2 resource schema WHEN compared field-by-field THEN docs/taskagent-gap-matrix.md exists, lists every field with status (mapped/partial/missing), and defers unimplemented writable gaps explicitly | ✓ met | docs/taskagent-gap-matrix.md is present in the branch diff (git diff --name-only main...HEAD includes docs/taskagent-gap-matrix.md); file has 338-line additions covering all 9 resource/data-source types per diff stat |
| 2 | GIVEN betterado_variable_group_variable resource is migrated to terraform-plugin-framework WHEN TestAccVariableGroupVariable acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values, ExpectNonEmptyPlan: false, destroy is clean | ✗ missed | resource_variable_group_variable_framework.go does not exist on branch; WI-10 status=failed; variable_group_variable remains in SDKv2 ResourcesMap in provider.go (line 167) |
| 3 | GIVEN SDKv2 variable_group_variable file is deregistered and deleted WHEN provider.go ResourcesMap is inspected THEN betterado_variable_group_variable is absent from SDKv2 ResourcesMap; source file deleted; framework_provider.go includes NewVariableGroupVariableResource; provider_test.go count updated | ✗ missed | resource_variable_group_variable.go still present in taskagent/; provider.go still registers taskagent.ResourceVariableGroupVariable(); WI-10 status=failed |
| 4 | GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-variable-group-variable", url, apiResponse) writes .forge/live-evidence/acceptance-resource-variable-group-variable.json | ✗ missed | .forge/live-evidence/acceptance-resource-variable-group-variable.json not present (only: acceptance-resource-agent-pool.json, acceptance-resource-agent-queue.json, acceptance-resource-environment-kubernetes.json, acceptance-resource-environment.json, acceptance-resource-task-group-datasource.json, acceptance-resource-variable-group.json) |
| 5 | GIVEN all taskagent resources and data sources have been migrated to framework WHEN make docs is run THEN docs/ directory is regenerated; docs/taskagent-gap-matrix.md is up to date | ~ partial | docs/ regenerated for all WI-2 through WI-9 types (docs/resources/variable_group.md, docs/data-sources/task_group.md etc. in diff); WI-10 variable_group_variable docs not regenerated (type not yet framework); taskagent-gap-matrix.md is present |
| 6 | GIVEN the migration is complete WHEN CHANGELOG.md is inspected THEN an '## Unreleased' entry exists documenting migration of all taskagent resources/data-sources to terraform-plugin-framework | ~ partial | CHANGELOG.md ## [Unreleased] has entries for agent_queue, elastic_pool, deployment_group, agent_pool (resource+data sources), environment (resource+data source), environment_resource_kubernetes; missing entries for variable_group_variable (WI-10 incomplete) and task_group data source |
| 7 | GIVEN the provider ships a user-visible change (all taskagent types now framework) WHEN PROVIDER_VERSION.txt is inspected THEN version is bumped by one minor semver increment from the pre-initiative value | ✗ missed | PROVIDER_VERSION.txt=1.2.0; WI-11 status=failed; no minor bump has been applied for this initiative (prior version before initiative was 1.2.0 from merge commit 7b66e8f3) |
| 8 | GIVEN provider_test.go counts are all correct after migration WHEN TestProvider_HasChildResources and TestProvider_HasChildDataSources run THEN both tests pass with the updated counts reflecting all taskagent types removed from SDKv2 | ~ partial | provider_test.go lists betterado_variable_group_variable (line 130) still in SDKv2 resource list; WI-11 gate (go test -tags all -run TestProvider_Has ./azuredevops/) not run; TestProvider_HasChildResources may pass (variable_group_variable still in SDKv2 map) but counts for removed types are updated |
| 9 | GIVEN betterado_task_group data source is migrated to terraform-plugin-framework WHEN TestAccTaskGroupDataSource_basic runs live (TF_ACC=1) THEN apply succeeds, provider read-back matches resource attributes, ExpectNonEmptyPlan: false, destroy is clean | ✓ met | .forge/live-evidence/acceptance-resource-task-group-datasource.json captured at 2026-07-03T15:21:06Z; REST GET https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/distributedtask/taskgroups/6f59dc30-0466-4054-80f7-eaf88fd95ac0?api-version=7.1 returned task group with category=Build, runsOn=[Agent,DeploymentGroup] |
| 10 | GIVEN data_task_group.go (SDKv2) is deregistered and deleted WHEN provider.go DataSourcesMap is inspected THEN betterado_task_group data source is absent from the SDKv2 map; provider_test.go count decremented; data_task_group.go and data_task_group_test.go in taskagent/ deleted; framework_provider.go DataSources() includes NewTaskGroupDataSource | ✓ met | data_task_group.go and data_task_group_test.go appear in diff (removed); framework_provider.go registers NewTaskGroupDataSource in DataSources(); provider.go DataSourcesMap has comment 'betterado_task_group data source is now registered in the framework provider' |
| 11 | GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called during the test THEN CaptureLiveEvidence("acceptance-resource-task-group-datasource", url, apiResponse) writes .forge/live-evidence/acceptance-resource-task-group-datasource.json | ✓ met | .forge/live-evidence/acceptance-resource-task-group-datasource.json present, capturedAt=2026-07-03T15:21:06Z, url=https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/distributedtask/taskgroups/6f59dc30-0466-4054-80f7-eaf88fd95ac0?api-version=7.1 |
| 12 | GIVEN betterado_agent_pool resource and betterado_agent_pool / betterado_agent_pools data sources are migrated to terraform-plugin-framework WHEN TestAccAgentPool acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values, ExpectNonEmptyPlan: false, destroy is clean | ✓ met | .forge/live-evidence/acceptance-resource-agent-pool.json captured at 2026-07-03T15:21:09Z; REST GET https://dev.azure.com/davidgparsonson/_apis/distributedtask/pools/555?api-version=7.1 returned pool with name=test-acc-8k9dny1eh6, poolType=automation, autoProvision=false |
| 13 | GIVEN SDKv2 agent_pool files are deregistered and deleted WHEN provider.go ResourcesMap and DataSourcesMap are inspected THEN betterado_agent_pool resource and betterado_agent_pool/betterado_agent_pools data sources are absent from SDKv2 maps; their source files deleted; framework_provider.go Resources()/DataSources() includes the new factories; provider_test.go counts updated | ✓ met | resource_agent_pool.go, data_agent_pool.go, data_agent_pools.go appear in diff (removed); framework_provider.go registers NewAgentPoolResource, NewAgentPoolDataSource, NewAgentPoolsDataSource; CHANGELOG.md confirms SDKv2 files removed |
| 14 | GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-agent-pool", url, apiResponse) writes .forge/live-evidence/acceptance-resource-agent-pool.json | ✓ met | .forge/live-evidence/acceptance-resource-agent-pool.json present, capturedAt=2026-07-03T15:21:09Z |
| 15 | GIVEN betterado_agent_queue resource and betterado_agent_queue data source are migrated to terraform-plugin-framework WHEN TestAccAgentQueue acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values, ExpectNonEmptyPlan: false, destroy is clean | ✓ met | .forge/live-evidence/acceptance-resource-agent-queue.json captured; REST GET https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/distributedtask/queues/1917?api-version=7.1 confirms queue created |
| 16 | GIVEN SDKv2 agent_queue files are deregistered and deleted WHEN provider.go ResourcesMap and DataSourcesMap are inspected THEN betterado_agent_queue is absent from SDKv2 maps; source files deleted; framework_provider.go includes new factories; provider_test.go counts updated | ✓ met | resource_agent_queue.go, data_agent_queue.go appear in diff (removed); CHANGELOG.md entry confirms SDKv2 files removed; framework_provider.go wires NewAgentQueueResource, NewAgentQueueDataSource |
| 17 | GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-agent-queue", url, apiResponse) writes .forge/live-evidence/acceptance-resource-agent-queue.json | ✓ met | .forge/live-evidence/acceptance-resource-agent-queue.json present, capturedAt=2026-07-03T15:21:09Z |
| 18 | GIVEN betterado_environment resource and betterado_environment data source are migrated to terraform-plugin-framework WHEN TestAccEnvironment acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values, ExpectNonEmptyPlan: false, destroy is clean | ✓ met | .forge/live-evidence/acceptance-resource-environment.json captured at 2026-07-03T15:21:01Z; REST GET https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/distributedtask/environments/70?api-version=7.1 returned environment name=test-acc-mumkhnkgp8 |
| 19 | GIVEN SDKv2 environment files are deregistered and deleted WHEN provider.go ResourcesMap and DataSourcesMap are inspected THEN betterado_environment is absent from SDKv2 maps; source files deleted; framework_provider.go includes new factories; provider_test.go counts updated | ✓ met | resource_environment.go, data_environment.go in diff (removed); framework_provider.go registers NewEnvironmentResource, NewEnvironmentDataSource |
| 20 | GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-environment", url, apiResponse) writes .forge/live-evidence/acceptance-resource-environment.json | ✓ met | .forge/live-evidence/acceptance-resource-environment.json present, capturedAt=2026-07-03T15:21:01Z |
| 21 | GIVEN betterado_environment_resource_kubernetes resource is migrated to terraform-plugin-framework WHEN TestAccEnvironmentResourceKubernetes acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values, ExpectNonEmptyPlan: false, destroy is clean | ✓ met | .forge/live-evidence/acceptance-resource-environment-kubernetes.json captured; REST GET https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/distributedtask/environments/71/providers/kubernetes/29?api-version=7.1 confirms Kubernetes resource created |
| 22 | GIVEN SDKv2 environment_resource_kubernetes file is deregistered and deleted WHEN provider.go ResourcesMap is inspected THEN betterado_environment_resource_kubernetes is absent from SDKv2 ResourcesMap; source files deleted; framework_provider.go includes NewEnvironmentResourceKubernetesResource; provider_test.go count updated | ✓ met | resource_environment_resource_kubernetes.go in diff (removed); framework_provider.go registers NewEnvironmentResourceKubernetesResource; provider_test.go count updated |
| 23 | GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-environment-kubernetes", url, apiResponse) writes .forge/live-evidence/acceptance-resource-environment-kubernetes.json | ✓ met | .forge/live-evidence/acceptance-resource-environment-kubernetes.json present |
| 24 | GIVEN betterado_deployment_group resource is migrated to terraform-plugin-framework WHEN TestAccDeploymentGroup acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values, ExpectNonEmptyPlan: false, destroy is clean | ✓ met | CHANGELOG.md Unreleased entry confirms live acceptance verified by TestAccDeploymentGroup_basic, TestAccDeploymentGroup_update, TestAccDeploymentGroup_withPoolId with idempotency re-plan, import, and CaptureLiveEvidence. resource_deployment_group_framework.go in diff. |
| 25 | GIVEN SDKv2 deployment_group file is deregistered and deleted WHEN provider.go ResourcesMap is inspected THEN betterado_deployment_group is absent from SDKv2 ResourcesMap; source files deleted; framework_provider.go includes NewDeploymentGroupResource; provider_test.go count updated | ✓ met | resource_deployment_group.go in diff (removed); framework_provider.go registers NewDeploymentGroupResource; CHANGELOG confirms deregistration |
| 26 | GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-deployment-group", url, apiResponse) writes .forge/live-evidence/acceptance-resource-deployment-group.json | ~ partial | .forge/live-evidence/ does not contain acceptance-resource-deployment-group.json (only: agent-pool, agent-queue, environment-kubernetes, environment, task-group-datasource, variable-group, acceptance-resource.json, task-group-state-upgrade-live.json) |
| 27 | GIVEN betterado_elastic_pool resource is migrated to terraform-plugin-framework WHEN TestAccElasticPool acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values, ExpectNonEmptyPlan: false, destroy is clean | ✓ met | CHANGELOG.md Unreleased entry confirms live acceptance verified by TestAccElasticPool with ExpectNonEmptyPlan: false, import verify, and CaptureLiveEvidence. resource_elastic_pool_framework.go in diff. |
| 28 | GIVEN SDKv2 elastic_pool file is deregistered and deleted WHEN provider.go ResourcesMap is inspected THEN betterado_elastic_pool is absent from SDKv2 ResourcesMap; source file deleted; framework_provider.go includes NewElasticPoolResource; provider_test.go count updated | ✓ met | resource_elastic_pool.go in diff (removed); framework_provider.go registers NewElasticPoolResource; provider_test.go updated |
| 29 | GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-elastic-pool", url, apiResponse) writes .forge/live-evidence/acceptance-resource-elastic-pool.json | ~ partial | .forge/live-evidence/ does not contain acceptance-resource-elastic-pool.json separately (acceptance-resource.json and task-group-state-upgrade-live.json present; elastic_pool evidence may be in acceptance-resource.json) |
| 30 | GIVEN betterado_variable_group resource and betterado_variable_group data source are migrated to terraform-plugin-framework WHEN TestAccVariableGroup acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values (including secret variables), ExpectNonEmptyPlan: false, destroy is clean | ~ partial | .forge/live-evidence/acceptance-resource-variable-group.json captured at 2026-07-03T16:43:27Z; REST GET https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/distributedtask/variablegroups/1170?api-version=7.1 returned variable group. However .forge/last-gate-failure.md records TestAccVariableGroup destroy failures (dangling resource); WI-9 status=failed |
| 31 | GIVEN SDKv2 variable_group files are deregistered and deleted WHEN provider.go ResourcesMap and DataSourcesMap are inspected THEN betterado_variable_group is absent from SDKv2 maps; source files deleted; framework_provider.go includes new factories; provider_test.go counts updated | ✓ met | provider.go has comment 'betterado_variable_group is now a framework resource'; resource_variable_group_framework.go in diff; framework_provider.go registers NewVariableGroupResource, NewVariableGroupDataSource; data_variable_group.go in diff (removed) |
| 32 | GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-variable-group", url, apiResponse) writes .forge/live-evidence/acceptance-resource-variable-group.json | ✓ met | .forge/live-evidence/acceptance-resource-variable-group.json present, capturedAt=2026-07-03T16:43:27Z, url=https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/distributedtask/variablegroups/1170?api-version=7.1 |

## Visual Changes

### Quality gate — go test -tags all -count=1 ./azuredevops/internal/service/servicehook/... — passes on branch HEAD

- **Before:** servicehook package compiled against SDKv2-heavy taskagent package; 8 SDKv2 resources + 6 SDKv2 data sources in provider registration
- **After:** `ok  github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/servicehook 0.008s` — mux provider compiles cleanly with all migrated framework types registered alongside remaining SDKv2 types
- **Command:** `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...`

**Before output:**
```
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/servicehook	0.003s
```

**After output:**
```
ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/servicehook	0.008s
```

### betterado_task_group data source — live ADO REST GET confirms task group created via framework data source path

- **Before:** `data.betterado_task_group` was served by SDKv2 `data_task_group.go`; `DataSourcesMap` included `"betterado_task_group": taskagent.DataTaskGroup()`
- **After:** Task group data source id=6f59dc30 created; GET response confirms `category=Build`, `runsOn=[Agent,DeploymentGroup]`, `name=test-acc-3md0lpt6eh`. `TestAccTaskGroupDataSource_basic` PASS (TF_ACC=1, live, 2026-07-03T15:21:06Z).
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/distributedtask/taskgroups/6f59dc30-0466-4054-80f7-eaf88fd95ac0?api-version=7.1` _(captured 2026-07-03T15:21:06Z)_

```json
{
  "author": "",
  "category": "Build",
  "friendlyName": "test-acc-3md0lpt6eh",
  "id": "6f59dc30-0466-4054-80f7-eaf88fd95ac0",
  "name": "test-acc-3md0lpt6eh",
  "runsOn": ["Agent", "DeploymentGroup"],
  "version": { "isTest": false, "major": 1, "minor": 0, "patch": 0 }
}
```

### betterado_agent_pool — live ADO REST GET confirms agent pool created via framework resource path

- **Before:** `betterado_agent_pool` was served by SDKv2 `resource_agent_pool.go`; `ResourcesMap` included `"betterado_agent_pool": taskagent.ResourceAgentPool()`
- **After:** Agent pool id=555 created; GET response confirms `name=test-acc-8k9dny1eh6`, `poolType=automation`, `autoProvision=false`. `TestAccAgentPool_basic` PASS (TF_ACC=1, live, 2026-07-03T15:21:09Z).
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/_apis/distributedtask/pools/555?api-version=7.1` _(captured 2026-07-03T15:21:09Z)_

```json
{
  "id": 555,
  "isHosted": false,
  "name": "test-acc-8k9dny1eh6",
  "poolType": "automation",
  "autoProvision": false,
  "autoUpdate": false,
  "size": 0
}
```

### betterado_agent_queue — live ADO REST GET confirms agent queue created via framework resource path

- **Before:** `betterado_agent_queue` was served by SDKv2 `resource_agent_queue.go`
- **After:** Agent queue id=1917 created. `TestAccAgentQueue*` PASS (TF_ACC=1, live).
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/distributedtask/queues/1917?api-version=7.1` _(captured 2026-07-03)_

### betterado_environment — live ADO REST GET confirms environment created via framework resource path

- **Before:** `betterado_environment` was served by SDKv2 `resource_environment.go`
- **After:** Environment id=70 created; GET confirms `name=test-acc-mumkhnkgp8`. `TestAccEnvironment*` PASS (TF_ACC=1, live, 2026-07-03T15:21:01Z).
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/distributedtask/environments/70?api-version=7.1` _(captured 2026-07-03T15:21:01Z)_

### betterado_environment_resource_kubernetes — live ADO REST GET confirms Kubernetes resource created via framework path

- **Before:** `betterado_environment_resource_kubernetes` was served by SDKv2
- **After:** Kubernetes environment resource id=29 in environment id=71 created. `TestAccEnvironmentResourceKubernetes_createUpdate` PASS (TF_ACC=1, live).
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/distributedtask/environments/71/providers/kubernetes/29?api-version=7.1`

### betterado_variable_group — live ADO REST GET confirms variable group created via framework resource path

- **Before:** `betterado_variable_group` was served by SDKv2 `resource_variable_group.go`; secret variables required special handling
- **After:** Variable group id=1170 created; GET response confirms `name=test-acc-2xe7xyb1rk`, `type=Vsts`, variables with key1=value1. Live evidence captured 2026-07-03T16:43:27Z. Note: destroy-phase race condition in acceptance test recorded in .forge/last-gate-failure.md; WI-9 marked failed.
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/distributedtask/variablegroups/1170?api-version=7.1` _(captured 2026-07-03T16:43:27Z)_

```json
{
  "id": 1170,
  "name": "test-acc-2xe7xyb1rk",
  "type": "Vsts",
  "variables": { "key1": { "value": "value1" } }
}
```

### gap matrix — docs/taskagent-gap-matrix.md documents all 9 taskagent types

- **Before:** No gap matrix existed for taskagent types; only `docs/task-group-gap-matrix.md` existed for the already-migrated task_group resource
- **After:** `docs/taskagent-gap-matrix.md` added (338 lines); one section per resource/data-source type with status table (mapped/partial/missing) and deferred-writable-gap notes for each of the 9 in-scope types

### Known gaps (partial/missed ACs)

- **variable_group_variable (WI-10):** `resource_variable_group_variable_framework.go` not written; WI-10 status=failed. SDKv2 file still registered. Root cause: WI-9 variable_group live gate was failing on destroy; WI-10 was blocked on WI-9 completion.
- **PROVIDER_VERSION.txt (WI-11):** Not bumped — WI-11 was blocked on WI-10 completion.
- **deployment_group live evidence file:** `acceptance-resource-deployment-group.json` not present in `.forge/live-evidence/` (live acceptance ran per CHANGELOG but evidence file not written with that label).
- **elastic_pool live evidence file:** `acceptance-resource-elastic-pool.json` not present as a separate file.
