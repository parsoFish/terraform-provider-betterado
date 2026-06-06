# Fix Plan

> Checklist for WI-4. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN both data sources are registered in the provider WHEN docs/resources/release_definition.md.tmpl (or equivalent docs file) and examples/data-sources/betterado_release_definition/main.tf are created THEN a valid minimal Terraform example exists showing data.betterado_release_definition lookup by id and by name
- [x] AC2: GIVEN docs/resources/release_definitions.md.tmpl (or equivalent) and examples/data-sources/betterado_release_definitions/main.tf are created WHEN a user reads the example THEN the example shows a valid list lookup with project_id and an optional path filter outputting release_definitions
