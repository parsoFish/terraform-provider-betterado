# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_agent_pool resource and betterado_agent_pool / betterado_agent_pools data sources are migrated to terraform-plugin-framework WHEN TestAccAgentPool acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values, ExpectNonEmptyPlan: false, destroy is clean
- [x] AC2: GIVEN SDKv2 agent_pool files are deregistered and deleted WHEN provider.go ResourcesMap and DataSourcesMap are inspected THEN betterado_agent_pool resource and betterado_agent_pool/betterado_agent_pools data sources are absent from SDKv2 maps; their source files deleted; framework_provider.go Resources()/DataSources() includes the new factories; provider_test.go counts updated
- [x] AC3: GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-agent-pool", url, apiResponse) writes .forge/live-evidence/acceptance-resource-agent-pool.json

## Sub-tasks completed in iteration 0

- [x] Create resource_agent_pool_framework.go with AgentPoolResource (CRUD + ImportState + syncPoolStatus)
- [x] Create data_agent_pool_framework.go with AgentPoolDataSource (lookup by name)
- [x] Create data_agent_pools_framework.go with AgentPoolsDataSource (list all pools as nested list)
- [x] Register NewAgentPoolResource, NewAgentPoolDataSource, NewAgentPoolsDataSource in framework_provider.go
- [x] Remove betterado_agent_pool from provider.go ResourcesMap
- [x] Remove betterado_agent_pool/betterado_agent_pools from provider.go DataSourcesMap
- [x] Delete SDKv2 files: resource_agent_pool.go, data_agent_pool.go, data_agent_pools.go, data_agent_pool_test.go, data_agent_pools_test.go
- [x] Update provider_test.go: remove betterado_agent_pool from resources list; remove betterado_agent_pool/betterado_agent_pools from data sources list
- [x] Update resource_agent_pool_test.go: ProtoV6ProviderFactories, getDirectClient() for CheckDestroy, ExpectNonEmptyPlan: false, captureAgentPoolEvidence -> CaptureLiveEvidence("acceptance-resource-agent-pool", ...)
- [x] Update data_agent_pool_test.go and data_agent_pools_test.go: ProtoV6ProviderFactories, ExpectNonEmptyPlan: false
- [x] make docs (tfplugindocs) + git checkout -- docs/guides/
- [x] CHANGELOG.md [Unreleased] entries
- [x] make test: all pass; golangci-lint --new-from-rev=main: 0 issues
- [x] Committed: 8453169e feat(taskagent): migrate betterado_agent_pool resource + data sources to framework

## Awaiting live gate

The live quality gate (TF_ACC=1, `go test -tags all -run TestAccAgentPool ./azuredevops/internal/acceptancetests/`) must pass in the forge environment.
