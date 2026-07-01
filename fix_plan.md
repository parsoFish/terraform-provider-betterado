# fix_plan — WI-4

> Checklist for WI-4: migrate betterado_release_folder permissions resources to framework.

- [x] AC1: `ProtoV6ProviderFactories` + `GetMuxProviderFactories()` in `TestAccMuxSdkv2Passthrough` — DONE (f3925364)
  - Root fix: used SharedReleaseFixture to avoid project-create (org at 1000-project cap)
  - Previous iteration had `ProtoV6ProviderFactories` but still created a new project → org cap failure
- [x] AC2: `docs/data-sources/` and `docs/resources/` files updated by `make docs`; `docs/guides/` restored — DONE (prior iterations)
- [x] AC3: `examples/resources/betterado_release_folder/resource.tf` and `examples/data-sources/betterado_release_folder/data-source.tf` exist — DONE (17bddc0f)
- [x] AC4: CHANGELOG.md entry under `## Unreleased` + PROVIDER_VERSION.txt bumped 1.0.5 → 1.1.0 — DONE (fbb7d7da)

All 4 ACs should now be satisfied. Gate re-run required to confirm AC1.
