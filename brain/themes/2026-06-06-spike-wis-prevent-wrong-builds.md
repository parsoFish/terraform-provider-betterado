---
title: Spike WIs prevent building on wrong API assumptions — zero rework cost
description: When a WI's first task is to empirically confirm an API assumption before building, incorrect assumptions are corrected before any implementation starts — rework cost is zero because no wrong code was written.
category: pattern
project: terraform-provider-betterado
created_at: 2026-06-06T06:04:52Z
updated_at: 2026-06-06T06:04:52Z
related_themes:
  - 2026-06-06-releasemgmt2-token-format-confirmed
---

# Spike WIs prevent building on wrong API assumptions

## Pattern

WI-1 of INIT-2026-06-05-release-definition-permissions was a deliberate spike: confirm the `ReleaseManagement2` token format against live ADO before writing the full permissions implementation.

The WI spec hypothesised `ReleaseManagement2/Project/{projectId}/{definitionId}`. The spike proved the correct format is `{projectId}/{releaseDefinitionId}`. Because the spike ran first, WI-2's implementation was built on confirmed ground — no wrong code was written and then deleted.

## Why it matters for this provider

The ADO security namespace API is poorly documented. Token format differences between namespaces (`ReleaseManagement` vs `ReleaseManagement2`, `Build` vs `Git`) are non-obvious from namespace names alone. Building on an unconfirmed token format would produce a resource that:
1. Compiles and passes unit tests (token is a string — any value passes offline).
2. Fails at acceptance time (wrong token → API rejects ACE writes).
3. Requires finding and fixing the format after implementation is complete.

The spike inverts this: prove the format first ($9 spike), build correctly ($2 impl). Without the spike, the fix cost after building on wrong assumptions is higher.

## Spike WI structure

A well-formed spike WI in this codebase:
1. `acceptance_criteria` describes what is to be confirmed (not what is to be built).
2. `quality_gate_cmd` runs an offline unit test that asserts the confirmed constant is non-empty and structurally valid — NOT a live ADO call. The live call happens during the agent's iteration, not as the gate.
3. The implementation file created is a documented stub (constant + token function) that WI-2 then completes.

## Operator directive alignment

The manifest explicitly called this out: *"First WI is a token-format probe spike against the live org before building; pivot the build on the confirmed pattern."* This directive was followed exactly and prevented a wrong implementation.

## Sources

- `_logs/2026-06-06T05-30-11_INIT-2026-06-05-release-definition-permissions/work-items-snapshot/WI-1.md`
- `brain/cycles/_raw/2026-06-06T05-30-11_INIT-2026-06-05-release-definition-permissions.md`
