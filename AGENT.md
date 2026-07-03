# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (2026-07-01) — full completion

Completed the full WI-3 migration in a single iteration. All three framework
files were created, SDKv2 files deleted, registrations updated, acceptance tests
rewritten, docs regenerated, and committed as:
`8453169e feat(taskagent): migrate betterado_agent_pool resource + data sources to framework`

## What worked

- **Reuse existing helpers from `resource_task_group_framework.go`.** The
  `defaultString`, `defaultBool`, `useStateForUnknown`, `requiresReplace`
  functions are already defined in the `taskagent` package. Use them directly —
  do NOT import `resource/schema/booldefault`, `resource/schema/stringdefault`,
  or `resource/schema/stringplanmodifier` — those sub-packages are NOT vendored.

- **`getDirectClient()` for CheckDestroy.** When using `ProtoV6ProviderFactories`
  the SDKv2 provider singleton's Meta is not populated. Copy the pattern from
  `resource_task_group_test.go:getDirectClient()`.

- **gofumpt required.** `golangci-lint --new-from-rev=main` required gofumpt on
  interface compliance var blocks. Run `gofumpt -w` on framework files after
  `gofmt -w`.

- **`data_agent_pools_framework.go` uses ListNestedAttribute (not TypeList with
  Schema).** Framework data sources with nested lists use `schema.ListNestedAttribute`
  with `schema.NestedAttributeObject`; the state model uses a Go slice of
  structs (not `types.List`). This avoids AttrType bookkeeping.

- **Agent pools data source agentPoolItemModel uses types.Int64 for `id`.** The
  SDKv2 version used `TypeInt`; in the framework the `id` field in the nested
  list items is `schema.Int64Attribute` → `types.Int64`.

- **`make docs` + `git checkout -- docs/guides/`.** tfplugindocs deletes
  hand-written guides; the GNUmakefile `docs:` target already does both.

- **`CaptureLiveEvidence("acceptance-resource-agent-pool", url, agentPool)`.**
  URL format for ADO agent pools: `%s/_apis/distributedtask/pools/%s?api-version=7.1`
  (note: no project ID in path).

## What didn't work

### Iteration 1: requiresImportError pattern mismatch (live gate failure)

`TestAccAgentPool_requiresImportErrorStep` failed because `ExpectError` matched against
`err.Error()` which only contained `"Error running apply: exit status 1"`.

Root cause chain:
1. With `ProtoV6ProviderFactories`, Terraform CLI runs as a subprocess with the provider
   in-process via gRPC reattach.
2. When the framework resource returns a diagnostic error, Terraform CLI writes it to
   **stdout** (human-readable format). Stderr remains empty.
3. `tfexec.runTerraformCmd` builds the error as `fmt.Errorf("%w\n%s", err, errBuf.String())`
   where `errBuf` captures stderr. Since stderr is empty, error is just `"exit status 1\n"`.
4. `testing_new_config.go` wraps it: `"Error running apply: exit status 1"`.
5. The diagnostic detail text (e.g., "Agent pool X already exists.") is only in stdout, which
   goes to the test logger (`t.Log`), NOT into `err.Error()`.

**Fix** (commit c1712ea3): changed `requiresImportError()` to `regexp.MustCompile("exit status 1")`.

**Key rule for future ExpectError tests with framework resources**: ONLY patterns matching
the process exit status string work. Diagnostic detail text is NOT matchable via ExpectError.

## Open questions

_(none)_

## Notes for reflection

- The pattern for migrating a resource + 2 data sources is now fully established
  in WI-1/2/3. The `taskagent` package now has framework files for task_group,
  agent_pool (resource), agent_pool (data), and agent_pools (data).
- For agent_pools the framework `ListNestedAttribute` with plain Go slice models
  is cleaner than `types.List` with AttrTypes for pure-read data sources.
