---
title: TestProvider_HasChildResources resource-count assertion is a maintenance trap
description: The provider-level test hard-codes the expected count of registered resources; adding any new betterado_* resource silently breaks CI until the number is updated.
category: antipattern
project: terraform-provider-betterado
created_at: 2026-06-02T09:47:24Z
updated_at: 2026-06-02T09:47:24Z
---

# TestProvider_HasChildResources resource-count assertion is a maintenance trap

## Problem

`azuredevops/provider_test.go` contains `TestProvider_HasChildResources`, which asserts that the provider registers exactly N resource types. When a new `betterado_*` resource is added to `provider.go`, the test fails with a count mismatch — breaking the `unit-test.yml` CI workflow — until someone manually updates the expected number.

This was one of the three root causes of CI failure in INIT-2026-06-01-ci-green: `betterado_task_group` had been registered in `provider.go` (in a prior cycle) but the test still expected 131 resources rather than 132.

## Fix pattern

Every time a new `betterado_*` resource or data source is registered in `provider.go`, also update the expected count in `TestProvider_HasChildResources`. This is a mandatory pair: `provider.go` edit ↔ `provider_test.go` count update.

Alternatively, the test could be rewritten to use a registry snapshot rather than a raw count, but that is a larger change that requires upstream judgement. Until then, the pair-edit discipline is the mitigation.

## Recurrence risk

This will recur with every future resource addition. The next planned initiatives (release_folder, release_environment_template, data sources) will each add at least one entry to the provider registry and will each need a `provider_test.go` count update or they will break CI on merge.

## Sources

- `_logs/2026-06-02T09-28-54_INIT-2026-06-01-ci-green/events.jsonl` (Ralph iter-1 last_assistant_text: "Fixed TestProvider_HasChildResources which expected 131 resources but provider registered 132")
- `brain/cycles/_raw/2026-06-02T09-28-54_INIT-2026-06-01-ci-green.md`
