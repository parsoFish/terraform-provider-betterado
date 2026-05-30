# Graph Report - INIT-2026-05-31-task-group-unit-tests  (2026-05-31)

## Corpus Check
- 592 files · ~321,196 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 22 nodes · 19 edges · 3 communities
- Extraction: 100% EXTRACTED · 0% INFERRED · 0% AMBIGUOUS
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `99d67d81`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]

## God Nodes (most connected - your core abstractions)
1. `Add gomock unit tests for the betterado_task_group resource` - 4 edges
2. `Changed files` - 2 edges
3. `Five unit tests cover every CRUD path and expand/flatten symmetry of the task group resource` - 1 edges
4. `Acceptance criteria` - 1 edges
5. `code:block1 (.forge/project.json                                         )` - 1 edges
6. `essence` - 1 edges
7. `project` - 1 edges
8. `initiativeId` - 1 edges
9. `baseRef` - 1 edges
10. `changedRef` - 1 edges

## Surprising Connections (you probably didn't know these)
- None detected - all connections are within the same source files.

## Communities (3 total, 0 thin omitted)

### Community 1 - "Community 1"
Cohesion: 0.20
Nodes (9): acceptanceCriteria, baseRef, changedRef, checkpoints, diffStat, essence, initiativeId, project (+1 more)

### Community 2 - "Community 2"
Cohesion: 0.33
Nodes (5): Acceptance criteria, Add gomock unit tests for the betterado_task_group resource, Changed files, code:block1 (.forge/project.json                                         ), Five unit tests cover every CRUD path and expand/flatten symmetry of the task group resource

## Knowledge Gaps
- **12 isolated node(s):** `Five unit tests cover every CRUD path and expand/flatten symmetry of the task group resource`, `Acceptance criteria`, `code:block1 (.forge/project.json                                         )`, `title`, `essence` (+7 more)
  These have ≤1 connection - possible missing edges or undocumented components.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What connects `Five unit tests cover every CRUD path and expand/flatten symmetry of the task group resource`, `Acceptance criteria`, `code:block1 (.forge/project.json                                         )` to the rest of the system?**
  _12 weakly-connected nodes found - possible documentation gaps or missing edges._