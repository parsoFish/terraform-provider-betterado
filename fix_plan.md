# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN docs/resources/task_group.md documents the schema before the gap-field additions WHEN the Schema section and example usage block are updated to include icon_url, input.visible_rule, input.properties, and input.aliases THEN the docs file reflects all four new attributes in both the example and the nested schema tables; the example uses non-default values matching the acceptance test fixture
- [x] AC2: GIVEN examples/ has no task_group example file WHEN examples/resources/betterado_task_group/resource.tf is created THEN the file contains a complete HCL example that demonstrates icon_url and all new input fields, compiles without error under terrafmt
- [x] AC3: GIVEN the CI gate includes make terrafmt-check WHEN make terrafmt-check is run after this WI THEN no HCL formatting errors are reported for the new or updated example files
