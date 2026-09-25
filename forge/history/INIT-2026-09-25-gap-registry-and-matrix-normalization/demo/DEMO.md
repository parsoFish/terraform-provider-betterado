# Gap registry and matrix normalization

> _Derived from `demo.json` (ADR 021). Essence:_ Gap registry and matrix normalization — 6 work items delivered against 24 acceptance criteria, 1 merge-boundary gate run.

## Summary

- WI-1 [complete] # WI-1: gap-registry.md scaffold + gate-wi-1.sh
- WI-2 [complete] # WI-2: Release + Pipeline tier — normalize 8 matrices + append registry entries
- WI-3 [complete] # WI-3: Infrastructure tier — normalize 7 matrices + append registry entries
- WI-4a [complete] # WI-4a: Identity/Security first 8 matrices — normalize + append registry entries
- WI-4b [complete] # WI-4b: Long-tail last 8 matrices — normalize + append registry entries
- WI-5 [complete] # WI-5: Synthesis — priority backlog, profile-corrections.md, api-coverage-roadmap.md update
- Commit: `f4601850b870e2658c01cf0cd711d593f6ac55f1`

## Visual Changes

### 31 files changed on this branch: docs/accounts-profile-gap-matrix.md, docs/api-coverage-roadmap.md, docs/approvalsandchecks-gap-matrix.md, docs/build-gap-matrix.md, docs/core-gap-matrix.md, docs/dashboard-gap-matrix.md, docs/extension-gap-matrix.md, docs/featuremanagement-gap-matrix.md, docs/feed-gap-matrix.md, docs/gallery-extensionmanagement-gap-matrix.md, docs/gap-registry.md, docs/git-gap-matrix.md, docs/graph-gap-matrix.md, docs/identity-gap-matrix.md, docs/memberentitlementmanagement-gap-matrix.md, docs/permissions-gap-matrix.md, docs/pipelinesapproval-gap-matrix.md, docs/policy-gap-matrix.md, docs/profile-corrections.md, docs/release-definition-gap-matrix.md, docs/release-folder-gap-matrix.md, docs/security-gap-matrix.md, docs/securityroles-gap-matrix.md, docs/serviceendpoint-gap-matrix.md, docs/servicehook-gap-matrix.md, docs/task-group-gap-matrix.md, docs/taskagent-gap-matrix.md, docs/test-gap-matrix.md, docs/wiki-gap-matrix.md, docs/workitemtracking-gap-matrix.md, docs/workitemtrackingprocess-gap-matrix.md


## Test Evidence

| test | result | delta |
|---|---|---|
| docs: forge gate docs /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/accounts-profile-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/api-coverage-roadmap.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/approvalsandchecks-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/build-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/core-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/dashboard-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/extension-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/featuremanagement-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/feed-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/gallery-extensionmanagement-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/gap-registry.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/git-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/graph-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/identity-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/memberentitlementmanagement-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/permissions-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/pipelinesapproval-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/policy-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/profile-corrections.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/release-definition-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/release-folder-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/security-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/securityroles-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/serviceendpoint-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/servicehook-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/task-group-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/taskagent-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/test-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/wiki-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/workitemtracking-gap-matrix.md /home/parso/forge-m7-a-run/_worktrees/INIT-2026-09-25-gap-registry-and-matrix-normalization/docs/workitemtrackingprocess-gap-matrix.md | pass | — |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Acceptance criteria

- (WI-1) GIVEN no docs/gap-registry.md exists on the clean tree WHEN ralph runs bash .forge/scripts/gate-wi-1.sh at iter-0 THEN the gate fails (test -f fails); after ralph creates the scaffold, the gate exits 0 with 'WI-1 PASSED'
- (WI-1) GIVEN docs/gap-registry.md is created WHEN the gate script checks section headers THEN grep -c finds at least 4 of: '## Vocabulary', '## Classification', '## Area index', '## Priority backlog'
- (WI-1) GIVEN docs/gap-registry.md ## Vocabulary section exists WHEN the gate checks canonical tokens THEN each of 'covered', 'gap-open', 'gap-deferred', 'out-of-scope' appears at least once in the file
- (WI-1) GIVEN .forge/scripts/gate-wi-1.sh is the deliverable WHEN the gate asserts test -f .forge/scripts/gate-wi-1.sh THEN the script file exists and is itself the artifact that passed
- (WI-2) GIVEN .forge/scripts/gate-wi-2.sh does not exist on the clean tree WHEN bash .forge/scripts/gate-wi-2.sh runs at iter-0 THEN bash fails with 'No such file or directory'; gate fails on clean tree
- (WI-2) GIVEN all 8 Release + Pipeline tier matrix files have been normalized WHEN gate-wi-2.sh checks each file for forbidden tokens in table rows THEN grep '^|' file | grep -cE 'mapped|supported|implemented|partial|missing|present|gap-resolved' returns 0 for every file
- (WI-2) GIVEN 8 registry entries have been appended to docs/gap-registry.md WHEN gate-wi-2.sh counts v7.2 delta lines in the registry THEN grep -c 'v7.2 delta' docs/gap-registry.md returns >= 8
- (WI-2) GIVEN each registry entry has gap-open count WHEN the count in the registry is compared to grep-count of gap-open rows in each matrix THEN grep -c '^| .*gap-open' in each matrix file matches the count stated in that area's registry entry
- (WI-3) GIVEN .forge/scripts/gate-wi-3.sh does not exist on the clean tree WHEN bash .forge/scripts/gate-wi-3.sh runs at iter-0 THEN bash fails with 'No such file or directory'; gate fails on clean tree
- (WI-3) GIVEN all 7 Infrastructure tier matrix files have been normalized WHEN gate-wi-3.sh checks each file for forbidden tokens in table rows THEN grep '^|' file | grep -cE 'mapped|supported|implemented|partial|missing|present|gap-resolved' returns 0 for every file
- (WI-3) GIVEN 7 registry entries have been appended to docs/gap-registry.md WHEN gate-wi-3.sh counts cumulative v7.2 delta lines THEN grep -c 'v7.2 delta' docs/gap-registry.md returns >= 15 (8 from WI-2 + 7 from WI-3)
- (WI-3) GIVEN each registry entry states a gap-open count WHEN the count is compared to grep-count of gap-open rows in each matrix THEN grep -c '^| .*gap-open' in each matrix file matches the count in that area's registry entry
- (WI-4a) GIVEN .forge/scripts/gate-wi-4a.sh does not exist on the clean tree WHEN bash .forge/scripts/gate-wi-4a.sh runs at iter-0 THEN bash fails with 'No such file or directory'; gate fails on clean tree
- (WI-4a) GIVEN all 8 Identity/Security first matrix files have been normalized WHEN gate-wi-4a.sh checks each file for forbidden tokens in table rows THEN grep '^|' file | grep -cE 'mapped|supported|implemented|partial|missing|present|gap-resolved' returns 0 for all 8 files
- (WI-4a) GIVEN 8 registry entries have been appended to docs/gap-registry.md WHEN gate-wi-4a.sh counts cumulative v7.2 delta lines THEN grep -c 'v7.2 delta' docs/gap-registry.md returns >= 23 (15 from prior WIs + 8 from WI-4a)
- (WI-4a) GIVEN each registry entry states a gap-open count WHEN count compared to grep-count in the matrix THEN grep -c '^| .*gap-open' in each matrix matches the stated count in the registry
- (WI-4b) GIVEN .forge/scripts/gate-wi-4b.sh does not exist on the clean tree WHEN bash .forge/scripts/gate-wi-4b.sh runs at iter-0 THEN bash fails with 'No such file or directory'; gate fails on clean tree
- (WI-4b) GIVEN all 8 long-tail matrix files have been normalized WHEN gate-wi-4b.sh checks each file for forbidden tokens in table rows THEN grep '^|' file | grep -cE 'mapped|supported|implemented|partial|missing|present|gap-resolved' returns 0 for all 8 files
- (WI-4b) GIVEN 8 registry entries have been appended to docs/gap-registry.md (all 31 areas now covered) WHEN gate-wi-4b.sh counts cumulative v7.2 delta lines THEN grep -c 'v7.2 delta' docs/gap-registry.md returns >= 31 (all 31 areas)
- (WI-4b) GIVEN each registry entry states a gap-open count WHEN count compared to grep-count in each matrix THEN grep -c '^| .*gap-open' in each matrix matches the count in that area's registry entry
- (WI-5) GIVEN .forge/scripts/gate-wi-5.sh does not exist on the clean tree WHEN bash .forge/scripts/gate-wi-5.sh runs at iter-0 THEN bash fails with 'No such file or directory'; gate fails on clean tree
- (WI-5) GIVEN docs/gap-registry.md ## Priority backlog section is complete WHEN gate-wi-5.sh checks for Tier labels THEN grep -qE 'Tier 1|Tier 2|Tier 3' docs/gap-registry.md returns 0 (match found)
- (WI-5) GIVEN docs/profile-corrections.md exists with >= 2 fenced replacement pairs WHEN gate-wi-5.sh counts fence markers THEN grep -c '```' docs/profile-corrections.md returns >= 4
- (WI-5) GIVEN docs/api-coverage-roadmap.md has been updated with FEAT-1 through FEAT-4 statuses WHEN gate-wi-5.sh checks each FEAT row THEN grep -qE 'FEAT-N.*(complete|partial|not-started|shipped)' passes for N=1,2,3,4; 'parked plan' does not appear

## Files Changed

- `docs/accounts-profile-gap-matrix.md`
- `docs/api-coverage-roadmap.md`
- `docs/approvalsandchecks-gap-matrix.md`
- `docs/build-gap-matrix.md`
- `docs/core-gap-matrix.md`
- `docs/dashboard-gap-matrix.md`
- `docs/extension-gap-matrix.md`
- `docs/featuremanagement-gap-matrix.md`
- `docs/feed-gap-matrix.md`
- `docs/gallery-extensionmanagement-gap-matrix.md`
- `docs/gap-registry.md`
- `docs/git-gap-matrix.md`
- `docs/graph-gap-matrix.md`
- `docs/identity-gap-matrix.md`
- `docs/memberentitlementmanagement-gap-matrix.md`
- `docs/permissions-gap-matrix.md`
- `docs/pipelinesapproval-gap-matrix.md`
- `docs/policy-gap-matrix.md`
- `docs/profile-corrections.md`
- `docs/release-definition-gap-matrix.md`
- `docs/release-folder-gap-matrix.md`
- `docs/security-gap-matrix.md`
- `docs/securityroles-gap-matrix.md`
- `docs/serviceendpoint-gap-matrix.md`
- `docs/servicehook-gap-matrix.md`
- `docs/task-group-gap-matrix.md`
- `docs/taskagent-gap-matrix.md`
- `docs/test-gap-matrix.md`
- `docs/wiki-gap-matrix.md`
- `docs/workitemtracking-gap-matrix.md`
- `docs/workitemtrackingprocess-gap-matrix.md`

```
31 files changed, 1826 insertions(+), 1219 deletions(-)
```
