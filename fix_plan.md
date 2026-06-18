# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN all HCL string fixtures in resource_release_definition_test.go (acceptance tests) WHEN grep is run for 'stages = [' array syntax THEN zero matches — all fixtures use block syntax (stages { … })
- [x] AC2: GIVEN the offline unit test suite for the release package WHEN go test -tags all -count=1 ./azuredevops/internal/service/release/ is run THEN all existing tests pass
