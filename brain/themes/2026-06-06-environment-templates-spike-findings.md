---
title: Environment templates spike — PARKED (viable but needs full environment modeling)
description: betterado_release_definition_environment_template is reachable live on the vsrm host via a raw-HTTP client, but create requires a full ReleaseDefinitionEnvironment payload (the provider's most complex type), so the resource was parked per the spike directive. This page holds everything a future build needs.
category: decision
project: terraform-provider-betterado
created_at: 2026-06-06T00:00:00Z
updated_at: 2026-06-06T00:00:00Z
---

# Environment templates spike — PARKED with findings

INIT-2026-06-05-environment-templates-spike was **parked** (a manifest-sanctioned
done state for a spike). The endpoint is viable, but completing the resource is a
substantial build, not the minor task it was scoped as. Three layers were peeled
back live; the first two are solved here, the third is the remaining work.

## 1. Host — use vsrm, not the core host (SOLVED)
`…/release/definitions/environmenttemplates` is part of the **release** REST
surface → it lives on **`vsrm.dev.azure.com`**, not `dev.azure.com`. Resolve the
client via the release resource area (`efc2f575-36ef-48e9-b672-0c6fb4a48ac5`)
exactly like the vendored `release.NewClient`, NOT `connection.GetClientByUrl(connection.BaseUrl)`
(that resolved the core host → "API resource location … is not registered on dev.azure.com").

## 2. The SDK location service does NOT register this endpoint (SOLVED)
Even on the vsrm host, `Client.Send(ctx, method, locationId, …)` fails — the
environment-templates location GUID (`6b3ad47a-2a42-4e24-9785-e3a0a8e3e64d`) is
**not registered** in the location service, so the GUID→URL-template resolution
returns "not registered on vsrm.dev.azure.com". The endpoint IS reachable by URL:

```
GET https://vsrm.dev.azure.com/{org}/{project}/_apis/release/definitions/environmenttemplates?api-version=7.1-preview.1
→ HTTP 200, {"count":34, "value":[…]}
```

**Working client pattern** (build the URL directly, reuse the SDK's auth):
- Get the vsrm-routed `*azuredevops.Client` via `GetClientByResourceAreaId(releaseArea)`.
- Derive the vsrm base by host-swapping `connection.BaseUrl` (`dev.azure.com` → `vsrm.dev.azure.com`).
- Build the URL, then `client.CreateRequestMessage(ctx, method, fullUrl, "7.1-preview.1", body, mediaType, acceptType, nil)`
  (applies PAT auth + api-version) → `client.SendRequest(req)` → `UnmarshalBody` / `UnmarshalCollectionBody`.
This bypasses `getResourceLocation` entirely. (Code lived on the abandoned branch
`forge/INIT-2026-06-05-environment-templates-spike`; recover from git if needed.)

## 3. Create requires a full `environment` blueprint (THE REMAINING BUILD)
With routing fixed, the real API rejects the create: **`Value cannot be null.
Parameter name: environment`**. The model is
`ReleaseDefinitionEnvironmentTemplate.Environment *ReleaseDefinitionEnvironment` —
i.e. a template wraps a **whole stage definition** (deploy phases, retention,
conditions, **and approvals**). The parked build only modeled `name`/`description`.

Finishing it means modeling the provider's heaviest type (the same
`environment {}` that `resource_release_definition.go` already expands, ~the bulk
of INIT-1). The test fixture's environment would also need a **VS402877-compliant
approval** (see [[2026-06-06-ado-rejects-automated-zero-guid-approval]]).

## Recommendation
Build as part of the operator's proposed **full integration-test project**
initiative (a project pre-seeded with identities, variable groups, repos,
pipelines, release definitions) — that gives a real `environment` to template
from and shared fixtures, instead of hand-rolling a minimal stage. The host +
raw-HTTP client work above is done; the remaining cost is the environment schema
+ expand/flatten + a valid live fixture.

## Sources
- Live probe (2026-06-06): `GET …/environmenttemplates` → HTTP 200 (34 templates).
- Live acc test progression: location-not-registered (core) → (vsrm) → `Value cannot be null. Parameter name: environment`.
- `vendor/.../release/models.go` `ReleaseDefinitionEnvironmentTemplate` / `ReleaseDefinitionEnvironment`.
