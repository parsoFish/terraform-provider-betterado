# Graph Report - INIT-2026-05-31-task-group-unit-tests  (2026-05-31)

## Corpus Check
- 592 files · ~321,679 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 32 nodes · 27 edges · 5 communities (4 shown, 1 thin omitted)
- Extraction: 100% EXTRACTED · 0% INFERRED · 0% AMBIGUOUS
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `b2efbde4`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]

## God Nodes (most connected - your core abstractions)
1. `What I tried` - 5 edges
2. `Add gomock unit tests for the betterado_task_group resource` - 4 edges
3. `Unifier Agent Memory — INIT-2026-05-31-task-group-unit-tests` - 3 edges
4. `Changed files` - 2 edges
5. `Iteration 4 (unifier)` - 1 edges
6. `Iteration 3 (unifier)` - 1 edges
7. `Iteration 2 (unifier)` - 1 edges
8. `Iteration 1 (unifier)` - 1 edges
9. `Notes for reflection` - 1 edges
10. `Fix Plan — unifier sub-phase` - 1 edges

## Surprising Connections (you probably didn't know these)
- None detected - all connections are within the same source files.

## Communities (5 total, 1 thin omitted)

### Community 1 - "Community 1"
Cohesion: 0.20
Nodes (9): acceptanceCriteria, baseRef, changedRef, checkpoints, diffStat, essence, initiativeId, project (+1 more)

### Community 2 - "Community 2"
Cohesion: 0.33
Nodes (5): Acceptance criteria, Add gomock unit tests for the betterado_task_group resource, Changed files, code:block1 (.forge/project.json                                         ), Five unit tests cover every CRUD path and expand/flatten symmetry of the task group resource

### Community 3 - "Community 3"
Cohesion: 0.25
Nodes (7): Iteration 1 (unifier), Iteration 2 (unifier), Iteration 3 (unifier), Iteration 4 (unifier), Notes for reflection, Unifier Agent Memory — INIT-2026-05-31-task-group-unit-tests, What I tried

## Knowledge Gaps
- **18 isolated node(s):** `Iteration 4 (unifier)`, `Iteration 3 (unifier)`, `Iteration 2 (unifier)`, `Iteration 1 (unifier)`, `Notes for reflection` (+13 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **1 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What connects `Iteration 4 (unifier)`, `Iteration 3 (unifier)`, `Iteration 2 (unifier)` to the rest of the system?**
  _18 weakly-connected nodes found - possible documentation gaps or missing edges._