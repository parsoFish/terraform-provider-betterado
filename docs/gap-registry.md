# Gap registry — terraform-provider-betterado

> Canonical cross-area coverage registry. Generated from the 31 per-area gap matrices in `docs/`.
> Vocabulary defined below; all matrices use ONLY these four tokens.

## Vocabulary

| Token | Meaning |
|---|---|
| `covered` | Field/resource is implemented and acceptance-tested |
| `gap-open` | Field/resource is missing and should be implemented |
| `gap-deferred` | Intentionally skipped; reason documented in the per-area matrix |
| `out-of-scope` | Non-declarative (imperative/runtime-only); will not be implemented |

**Forbidden tokens** (must not appear in any matrix): `mapped`, `supported`, `implemented`, `partial`,
`missing`, `present`, `gap-resolved`, `✅`, `⚠️`, `🚫`, `read-only`, `gap`, `breaking-deferral`,
`missing-writable-gap`, `missing-computed-gap`, and any variant not in the four above.

### Token mapping rules

- `implemented` / `mapped` / `present` / `✅ Implemented` → `covered`
- `gap` / `missing-writable-gap` / `⚠️ Gap` → `gap-open`
- `deferred` / `breaking-deferral` (when intentionally skipped) → `gap-deferred`
- `read-only` (server-assigned, never user-configurable) → `out-of-scope`
- `missing-computed-gap` → `out-of-scope` (default); `gap-deferred` only if field could usefully be Computed
- `🚫 Out of scope` / `out-of-scope` → `out-of-scope`

## Coverage index

| Area | Classification | gap-open count | gap-deferred count | v7.1→v7.2 delta |
|---|---|---|---|---|
| _(populated by WI-2 through WI-4b)_ | | | | |

## Priority backlog

_(Populated by WI-5 after all tier normalization completes.)_

### Tier 1 — betterado-net-new resources

### Tier 2 — high-value upstream gaps

### Tier 3 — low-value computed-field gaps
