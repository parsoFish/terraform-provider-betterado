# Agent Memory — UWI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (complete)
- Added `hashicorp/terraform-plugin-framework-validators v0.19.0` to go.mod via `go get` then `go mod tidy` + `go mod vendor`
- Replaced all hand-rolled validator types in `resource_workitem_framework.go`:
  - wiIsUUIDValidator → stringvalidator.RegexMatches(uuidRegexp, ...)
  - wiNotWhitespaceValidator → stringvalidator.RegexMatches(nonWhitespaceRegexp, ...)
  - wiTagsSizeFloorValidator → setvalidator.SizeAtLeast(1)
  - wiParentIDAtLeastValidator → int64validator.AtLeast(1)
  - wiConflictingFieldsValidator → resourcevalidator.Conflicting(path.MatchRoot(...), ...)
- Replaced all hand-rolled validator types in `resource_field_framework.go`:
  - notWhitespaceValidator, lengthBetweenValidator, doesNotMatchValidator, oneOfValidator, isUUIDValidator → library equivalents
  - doesNotMatchValidator (for reference_name forbidden chars) → stringvalidator.RegexMatches with positive char-class regexp
- Replaced all hand-rolled validator types in `resource_workitemquery_framework.go` and `resource_workitemquery_folder_framework.go`:
  - wiqIsUUIDValidator, wiqNotEmptyValidator, wiqAreaValidator, wiqWiqlLengthValidator, wiqExactlyOneOfValidator, wiqfExactlyOneOfValidator → library equivalents
- Fixed `data_area_framework.go` and `data_iteration_framework.go` which used the now-deleted hand-rolled types
- Defined shared package-level regexps `uuidRegexp` and `nonWhitespaceRegexp` in resource_workitem_framework.go (accessible to all files in package)
- go build ./..., go vet ./azuredevops/internal/service/workitemtracking/..., go test ./... all pass

## What worked

- `go get github.com/hashicorp/terraform-plugin-framework-validators@v0.19.0` followed by `go mod tidy` + `go mod vendor` added the dependency and vendored only the used sub-packages (stringvalidator, int64validator, setvalidator, resourcevalidator, helpers/*, internal/*)
- Using `stringvalidator.RegexMatches` with `regexp.MustCompile(\`\S\`)` (matches at least one non-whitespace char) is the library equivalent of "non-whitespace" validation
- Converting the "doesNotMatchValidator" (forbidden chars) to a positive regex `^[^,;~:/\\*|?"&%$!+=()[\]{}<>\-]+$` works with `stringvalidator.RegexMatches`
- Placing shared regexps as package-level vars in resource_workitem_framework.go makes them accessible to all files in the same package without redeclaration

## What didn't work

- The `go mod vendor` step needed to happen AFTER the source files imported the library (otherwise only `// indirect` and not vendored)
- `data_area_framework.go` and `data_iteration_framework.go` also used the hand-rolled validators — easy to miss since they weren't in the WI's "files in scope" for validators

## Open questions

_(things that aren't blocking but would be useful to clarify; reflector picks these up)_

## Notes for reflection

_(observations the reflector should capture into the brain; the agent doesn't write them itself, but flags here)_
