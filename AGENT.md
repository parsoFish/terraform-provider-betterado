# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (complete — both ACs done)

**Approach:** Python script to convert HCL array syntax to block syntax in the acceptance test file.

1. Grepped for all `keyword = [` patterns in the file — found 149 occurrences across 16 block keywords.
2. Wrote `convert_simple.py` using recursive bracket-matching + element extraction + indentation normalization.
   - `find_matching_bracket()` finds the matching `]` for each `[`
   - `convert_elements()` extracts each `{ ... }` element from the array and wraps with `keyword { ... }`
   - `normalize_indent()` re-indents content so minimum indent = target_indent + 2 spaces
   - `remove_null_lines()` strips `field = null` stub lines
   - Ran multiple passes until no more matches found (handles nesting)
3. Renamed `hclReleaseDefinitionStagesArraySyntax` → `hclReleaseDefinitionBlockSyntax` and
   `TestAccReleaseDefinition_stagesArraySyntax` → `TestAccReleaseDefinition_blockSyntax` per WI spec.
4. Ran `go test -tags all -count=1 ./azuredevops/internal/service/release/` — PASS.
5. Ran `go build -tags all ./azuredevops/internal/acceptancetests/` — PASS.

## What worked

- Python script with bracket-level tracking + per-element extraction + indent normalization worked cleanly.
- Running `go build -tags all ./azuredevops/internal/acceptancetests/` as a compile check.
- The conversion script needed multiple passes to handle nesting (inner blocks before outer).

## What didn't work

- First attempt (simple line-by-line state machine) produced correct structure but wrong indentation for closing braces.
- Second attempt (complex nested recursive state machine) was too complex.
- Third attempt (simple bracket-matching without indent fix) was correct structure but had indentation issues.
- Final approach (bracket-matching + `normalize_indent`) worked correctly.

## Open questions

- `variable = [{ ... }]` inside stages was NOT converted (not in the WI's conversion table). The WI spec says variable at definition level is checked separately; stage-level `variable` may need conversion in a future WI when live acceptance tests are run, since ConfigMode was removed from its schema.
- `workflow_task = [{ ... }]` and `task = [{ ... }]` inside gate/deploy_phase are also not in the conversion table — left as-is.

## Notes for reflection

- The Python converter script is at `/tmp/convert_simple.py` — useful template for similar HCL array→block conversions.
- The conversion adds trailing blank lines before each `}` (one empty line from the `\n` between elements). This is cosmetically suboptimal but syntactically valid HCL and the gate passed.
