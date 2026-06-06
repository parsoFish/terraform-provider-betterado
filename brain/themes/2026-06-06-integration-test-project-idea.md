---
title: Next initiative — a full shared integration-test project fixture
description: Operator idea (2026-06-06). Instead of each acc test hand-rolling a minimal release definition, stand up ONE realistic project fixture with all components wired together (identities, variable groups, repos, pipelines, releases) so new resources can be validated in full capacity. Would have pre-empted the VS402877/VS402982/wrong-key/env-template gaps.
category: idea
project: terraform-provider-betterado
created_at: 2026-06-06T00:00:00Z
updated_at: 2026-06-06T00:00:00Z
---

# Next initiative — a full shared integration-test project fixture

## The idea (operator, 2026-06-06)
Build a **full demo/integration project config** with all the components that tie
together — a real project definition, user identities, variable groups, repos,
build pipelines, release definitions, permissions — defined once and reused. With
that fixture in place, adding a new resource means exercising it *in the full
capacity it actually runs in*, not against a hand-rolled minimal stub.

## Why now (the evidence)
The INIT-1..4 close-out (2026-06-06) surfaced a cluster of gaps that a realistic
shared fixture would have caught immediately — and that the minimal,
copy-pasted-per-test HCL hid until every test was finally run live with TF_ACC:
- **VS402877** — ADO now requires BOTH pre- AND post-deploy approvals per stage;
  the minimal fixtures carried only pre. ([[2026-06-06-ado-rejects-…]] / the
  test fix commit.)
- **VS402982** — a stage with no `retention_policy` is rejected.
- **Wrong permission key** — the permissions test used `EditReleaseStage`, which
  the ReleaseManagement2 namespace doesn't define (`EditReleaseEnvironment` does).
- **Env templates** ([[2026-06-06-environment-templates-spike-findings]]) — create
  needs a full `environment` blueprint; with no real environment to template from,
  the spike couldn't be finished.

A single canonical environment/stage definition (with valid approvals, retention,
real identities, a real query for gates) reused across tests would make each of
these a one-place fix, and would give env-templates a real environment to capture.

## Shape (for the architect to refine)
- A `testutils` helper (or a fixture module) that provisions/looks-up: a project,
  a couple of real group/user identities, a variable group, a repo, a build
  definition, and a canonical multi-stage release definition — each stage valid
  against *current* ADO (pre+post approvals, stage retention).
- New-resource acc tests reference the shared fixture instead of re-declaring a
  minimal release definition inline.
- Keep it creds-gated + self-cleaning (the C7 live-acceptance discipline).

## Status
Idea only — a good roadmap-scale next initiative. Not yet specced. The release
component (INIT-1..4) is shipped + live-green; env-templates (INIT-5) is parked
pending exactly this kind of fixture.
