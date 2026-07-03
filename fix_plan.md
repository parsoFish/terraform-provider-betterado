# Fix Plan

> Checklist for WI-5. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN all three entitlement resources migrated to framework WHEN make docs runs and docs are committed THEN docs/resources/user_entitlement.md, docs/resources/group_entitlement.md, docs/resources/service_principal_entitlement.md reflect current framework schemas; docs/guides/ is restored
- [x] AC2: GIVEN live evidence captured in WI-2, WI-3, WI-4 acceptance tests WHEN demo.json is inspected THEN a checkpoint exists with liveEvidence.url pointing to a real REST GET URL
- [x] AC3: GIVEN all three resources migrated WHEN CHANGELOG.md is read THEN an Unreleased entry notes the framework migration of the three entitlement resources
- [x] AC4: GIVEN a user-visible schema change has shipped WHEN PROVIDER_VERSION.txt is read THEN the semver has been bumped (patch or minor) relative to its pre-cycle value
