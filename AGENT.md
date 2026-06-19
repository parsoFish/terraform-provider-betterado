# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (completed)

1. Ran `go get` for all three deps, wrote framework_provider.go + rewrote main.go, ran `go mod tidy` + `go mod vendor`, ran all builds and tests. All ACs satisfied in a single iteration.

## What worked

- **Write code first, then run `go get` + `go mod tidy` + `go mod vendor`.** If you run tidy before the imports exist, Go removes the packages from go.mod.
- **SDK upgrade was necessary.** `terraform-plugin-go@v0.31.0` added `GenerateResourceConfig` to the `tfprotov5.ProviderServer` interface. `terraform-plugin-sdk/v2@v2.38.1` didn't implement it, causing a compile error. Upgrading to `v2.40.1` fixed this.
- **Go internal package workaround.** `main.go` at the module root cannot import `azuredevops/internal/provider` (Go's internal package rule: only code with `azuredevops/` path prefix can import it). Solution: added `azuredevops/framework.go` that publicly re-exports `NewFrameworkProvider()` from the internal package. `main.go` then calls `azuredevops.NewFrameworkProvider()`.
- **`providerserver.NewProtocol6` (not `NewProtocol6WithError`).** The WI sample code referenced `NewProtocol6WithError` but the actual framework API is `NewProtocol6(provider.Provider)`. The mux pattern wraps it as `func() tfprotov6.ProviderServer`.

## What didn't work

- First attempt at `main.go` importing `azuredevops/internal/provider` directly failed: `use of internal package not allowed` — root module cannot access sub-package's internal.

## Open questions

_(things that aren't blocking but would be useful to clarify; reflector picks these up)_

- Full `golangci-lint run ./...` times out in the shell (~2+ min for whole codebase). Pre-existing issue. Lint on new files specifically with `--new-from-rev HEAD` is clean.
- `make terrafmt-check` also timed out in the bash tool — pre-existing behavior for large codebases.

## Notes for reflection

- The Go `internal/` package accessibility rule is a footgun when the provider pattern puts things in `pkg/internal/` but main.go is at the root. The thin re-export wrapper in the public package is the canonical workaround.
- When new versions of terraform-plugin-go bump protocol interfaces, dependent SDKs may lag. Always check if sdk/v2 needs upgrading when pulling in newer plugin-go.
