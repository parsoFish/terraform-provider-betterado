# Releasing `betterado` to the Terraform Registry

This provider publishes to the public Terraform Registry under the
`parsoFish/betterado` namespace. Releases are cut by pushing a semver tag; the
[`release.yml`](../.github/workflows/release.yml) GitHub Actions workflow runs
[GoReleaser](../.goreleaser.yml) to build, checksum, GPG-sign, and publish the
provider archives to a GitHub Release, which the registry then ingests.

The repo-side rig (GoReleaser config, release workflow, registry manifest,
version wiring, docs tooling) is already in place. The steps below are the
**one-time setup** and the **per-release** procedure — they require human action
(key material, account access, pushing a tag) and are not automated.

## One-time setup

### 1. Generate a GPG signing key

The registry verifies the `SHA256SUMS` signature against a public key registered
under your namespace.

```bash
gpg --full-generate-key      # RSA, 4096 bits, no expiry (or your policy)
gpg --list-secret-keys --keyid-format LONG   # note the key ID / fingerprint

# Export the PUBLIC key (added to the registry, below):
gpg --armor --export <KEY_ID> > betterado-public.gpg

# Export the PRIVATE key (added to GitHub secrets, below):
gpg --armor --export-secret-keys <KEY_ID> > betterado-private.gpg
```

### 2. Add GitHub Actions secrets

In the GitHub repo: **Settings → Secrets and variables → Actions → New
repository secret**:

| Secret name       | Value                                              |
|-------------------|----------------------------------------------------|
| `GPG_PRIVATE_KEY` | full contents of `betterado-private.gpg`           |
| `PASSPHRASE`      | the passphrase you set when generating the key     |

`GITHUB_TOKEN` is provided automatically by Actions — do not add it.

> Delete the exported `*.gpg` files from disk once the secrets are stored.

### 3. Register on the Terraform Registry

1. Sign in to <https://registry.terraform.io> with the GitHub account that owns
   (or can publish) `parsoFish/terraform-provider-betterado`.
2. **Publish → Provider**, select the `terraform-provider-betterado` repo, and
   confirm the `parsoFish` namespace.
3. Under the namespace **Settings → GPG Keys**, paste the contents of
   `betterado-public.gpg`.

The registry is now ready to ingest releases.

## Per-release procedure

1. **Generate / refresh docs** (first release, or after schema changes):

   ```bash
   make docs            # runs tfplugindocs generate into docs/
   ```

   On the very first run, migrate the inherited legacy `website/docs/` content
   into the modern `docs/` layout, then generate:

   ```bash
   go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.20.0
   tfplugindocs migrate --provider-name betterado   # website/ -> docs/ (one time)
   make docs
   ```

   Review the diff, commit `docs/`.

2. **Update the changelog** — move the `## 0.1.0 (Unreleased)` heading in
   [`CHANGELOG.md`](../CHANGELOG.md) to the release version + date, and add a new
   Unreleased section above it.

3. **Tag and push** — the tag drives the GoReleaser version:

   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

4. **Watch the release** — the `release` workflow runs GoReleaser. On success the
   GitHub Release for `v0.1.0` contains:
   - `terraform-provider-betterado_0.1.0_<os>_<arch>.zip` (per platform)
   - `terraform-provider-betterado_0.1.0_SHA256SUMS`
   - `terraform-provider-betterado_0.1.0_SHA256SUMS.sig`
   - `terraform-provider-betterado_0.1.0_manifest.json`

5. **Verify ingestion** — within a few minutes the version appears at
   <https://registry.terraform.io/providers/parsoFish/betterado>. Confirm it is
   installable:

   ```hcl
   terraform {
     required_providers {
       betterado = {
         source  = "parsoFish/betterado"
         version = "~> 0.1.0"
       }
     }
   }
   ```

   ```bash
   terraform init     # downloads + signature-verifies the provider
   ```

## Local validation before tagging

```bash
go build -mod=vendor .                 # entry point compiles
goreleaser check                       # validates .goreleaser.yml
goreleaser release --snapshot --clean  # dry-run build of all artifacts (no publish)
```

`--snapshot` produces the full archive + checksum set under `dist/` without
needing a tag or pushing anything — use it to confirm the matrix and naming
before cutting the real tag.
