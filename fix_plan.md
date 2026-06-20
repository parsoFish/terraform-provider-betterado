# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a new file azuredevops/internal/service/taskagent/resource_task_group_framework.go exporting NewTaskGroupResource() that returns a resource.Resource WHEN go test -tags all -run TestTaskGroupFramework_Schema ./azuredevops/internal/service/taskagent/ runs THEN the test passes: schema attributes task, input, version, project_id, name, friendly_name, description, category, author, icon_url, instance_name_format, runs_on, revision, definition_type are all present and the resource type name is betterado_task_group
- [x] AC2: GIVEN the framework resource implements Create, Read, Update, Delete with Context methods using the same TaskAgentClient API calls as the SDKv2 resource WHEN go build -mod=vendor . compiles THEN compilation succeeds with no errors
