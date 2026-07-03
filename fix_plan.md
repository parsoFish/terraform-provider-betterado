# Fix Plan

> Checklist for WI-6. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN make docs is run (tfplugindocs) after both servicehook resources are registered as framework resources WHEN docs/ is regenerated THEN docs/resources/betterado_servicehook_storage_queue_pipelines.md and docs/resources/betterado_servicehook_webhook_tfs.md exist and describe every attribute; git checkout -- docs/guides/ restores hand-written guides
- [x] AC2: GIVEN examples/resources/betterado_servicehook_storage_queue_pipelines/resource.tf and examples/resources/betterado_servicehook_webhook_tfs/resource.tf WHEN they exist with valid HCL THEN make terrafmt-check passes (HCL is formatted)
- [x] AC3: GIVEN CHANGELOG.md WHEN the migration is complete THEN CHANGELOG.md has a new entry under ## Unreleased describing the framework migration for both servicehook resources
- [x] AC4: GIVEN PROVIDER_VERSION.txt WHEN a user-visible change (resource schema parity maintained, internal migration) is delivered THEN PROVIDER_VERSION.txt is bumped to the next patch semver
