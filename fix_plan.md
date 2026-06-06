# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the mini-client scaffolding from WI-1 is in place and compiles WHEN `ResourceReleaseDefinitionEnvironmentTemplate()` is called and its schema is inspected THEN it returns a `*schema.Resource` with `CreateContext`, `ReadContext`, `DeleteContext` (no `UpdateContext` — templates are immutable), and schema attributes for `name`, `description`, `category`, `project_id`, `environment`, `icon_task_id`, and `can_delete` (read-only)
- [x] AC2: GIVEN a `ReleaseDefinitionEnvironmentTemplate` API object with all fields populated WHEN `flattenEnvironmentTemplate` is called on it THEN the returned `map[string]interface{}` contains the same `name`, `description`, `category`, and `can_delete` values — verified by a `TestReleaseDefinitionEnvironmentTemplate_Flatten` unit test
- [x] AC3: GIVEN a Terraform resource state with all fields set WHEN `expandEnvironmentTemplate` is called on the schema.ResourceData THEN the returned `ReleaseDefinitionEnvironmentTemplate` has matching `Name`, `Description`, `Category` fields — verified by a `TestReleaseDefinitionEnvironmentTemplate_Expand` unit test
- [x] AC4: GIVEN the resource is implemented and tests pass WHEN `azuredevops/provider.go` is opened THEN `betterado_release_definition_environment_template` is listed in `ResourcesMap` pointing to `release.ResourceReleaseDefinitionEnvironmentTemplate()`
