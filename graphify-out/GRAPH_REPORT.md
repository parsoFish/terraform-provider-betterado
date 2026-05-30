# Graph Report - INIT-2026-05-31-task-group-unit-tests  (2026-05-31)

## Corpus Check
- 591 files · ~320,053 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 16 nodes · 14 edges · 2 communities
- Extraction: 100% EXTRACTED · 0% INFERRED · 0% AMBIGUOUS
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `100df800`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 1|Community 1]]

## God Nodes (most connected - your core abstractions)
1. `essence` - 1 edges
2. `project` - 1 edges
3. `initiativeId` - 1 edges
4. `baseRef` - 1 edges
5. `changedRef` - 1 edges
6. `diffStat` - 1 edges
7. `checkpoints` - 1 edges
8. `acceptanceCriteria` - 1 edges

## Surprising Connections (you probably didn't know these)
- None detected - all connections are within the same source files.

## Communities (2 total, 0 thin omitted)

### Community 1 - "Community 1"
Cohesion: 0.20
Nodes (9): acceptanceCriteria, baseRef, changedRef, checkpoints, diffStat, essence, initiativeId, project (+1 more)

## Knowledge Gaps
- **9 isolated node(s):** `title`, `essence`, `project`, `initiativeId`, `baseRef` (+4 more)
  These have ≤1 connection - possible missing edges or undocumented components.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What connects `title`, `essence`, `project` to the rest of the system?**
  _9 weakly-connected nodes found - possible documentation gaps or missing edges._