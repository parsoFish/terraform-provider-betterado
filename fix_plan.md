# Fix Plan

> Checklist for WI-5. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the framework provider's Resources() and DataSources() slices WHEN WI-5 is complete THEN azuredevops/internal/provider/framework_provider.go imports the featuremanagement package and lists NewFeatureFlagResource and NewFeatureFlagDataSource in Resources() and DataSources() respectively; grep of azuredevops/provider.go confirms zero new entries for betterado_feature_flag
- [x] AC2: GIVEN make docs runs successfully WHEN docs are regenerated THEN docs/resources/feature_flag.md and docs/data-sources/feature_flag.md exist and describe every attribute; docs/guides/ is restored (git checkout -- docs/guides/); examples/resources/betterado_feature_flag/resource.tf exists with non-default field values
- [x] AC3: GIVEN provider_test.go TestProvider_HasChildResources / HasChildDataSources count assertions WHEN the counts are updated to include betterado_feature_flag THEN go test -tags all -run TestProvider ./azuredevops/ passes with the updated expected counts
- [x] AC4: GIVEN CHANGELOG.md and PROVIDER_VERSION.txt WHEN WI-5 is complete THEN CHANGELOG.md has a new entry under ## Unreleased documenting betterado_feature_flag; PROVIDER_VERSION.txt is bumped by a patch version (1.2.0 → 1.2.1)
