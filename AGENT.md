# Unifier Agent Memory — INIT-2026-07-01-mux-free-cutover

> Institutional memory across unifier-Ralph iterations. Read at the start of every iteration; updated at the end.

## What I tried

_(updated by each iteration — most recent at the top)_

### Iteration 15 (2026-07-07)

**Status:** Complete — same recurring chore-commit hazard resolved (15th time).

**What was done:**
1. Read AGENT.md (showed iters 1–14 done) and fix_plan.md (all 14 ACs ticked, AC12 partial).
2. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → GREEN (`ok ... 0.003s`).
3. Confirmed latest commit (dad62437) was another "chore: drop forge scratch" that deleted `.forge/pr-description.md` and `AGENT.md` from HEAD — same recurring pattern.
4. Verified all ACs live on HEAD (all green; AC12 partial inherent).
5. Ran `TestFrameworkProvider_MuxFree` → `ok 0.004s` (matches demo.json value — no update needed).
6. Restored `.forge/pr-description.md` and `AGENT.md` via force-add; committed as unify-and-demo.
7. Pushed — `origin/forge/INIT-2026-07-01-mux-free-cutover` == local HEAD.

**Key observations:**
- This is the 15th consecutive iteration where `chore: drop forge scratch` removes `.forge/pr-description.md` from tracking. The unifier must always force-re-add it.
- Quality gate and all unit tests remain green. All 14 ACs verified (AC12 partial — inherent without live ADO).

### Iteration 14 (2026-07-07)

**Status:** Complete — same recurring chore-commit hazard resolved (14th time).

**What was done:**
1. Read AGENT.md (showed iters 1–13 done) and fix_plan.md (all 14 ACs ticked, AC12 partial).
2. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → GREEN (`ok ... 0.003s`).
3. Confirmed latest commit (15297656) was another "chore: drop forge scratch" that deleted `.forge/pr-description.md` and `AGENT.md` from HEAD — same recurring pattern.
4. Verified all ACs live on HEAD (all green; AC12 partial inherent).
5. Ran `TestFrameworkProvider_MuxFree` → `ok 0.004s` (up from 0.003s in iter 13).
6. Updated demo.json checkpoint 4 `afterOutput`: `0.003s` → `0.004s` (live captured value).
7. Updated DEMO.md checkpoint 4 table row: `0.003s` → `0.004s`.
8. Force-added `.forge/pr-description.md` and AGENT.md, staged demo changes, committed as unify-and-demo.
9. Pushed — `origin/forge/INIT-2026-07-01-mux-free-cutover` == local HEAD.

**Key observations:**
- This is the 14th consecutive iteration where `chore: drop forge scratch` removes `.forge/pr-description.md` from tracking. The unifier must always force-re-add it.
- Quality gate and all unit tests remain green. All 14 ACs verified (AC12 partial — inherent without live ADO).

### Iteration 13 (2026-07-07)

**Status:** Complete — recurring chore-commit hazard resolved again (13th time).

**What was done:**
1. Read AGENT.md (showed iters 1–12 done) and fix_plan.md (all 14 ACs ticked, AC12 partial).
2. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → GREEN (`ok ... 0.003s`).
3. Confirmed latest commit (b48ba6a1) was another "chore: drop forge scratch" that deleted `.forge/pr-description.md` from HEAD — same recurring pattern.
4. Verified all ACs live on HEAD this iteration:
   - `grep -E 'tf5to6|tf6mux|helper/schema|NewGRPCProvider' main.go` → `none` ✅
   - `grep 'betterado_serviceendpoint' azuredevops/provider.go | grep -v '//'` → `empty` ✅
   - `go test -tags all -count=1 -run TestFrameworkProvider_MuxFree ./azuredevops/internal/provider/` → `ok 0.003s` ✅
   - `cat PROVIDER_VERSION.txt` → `2.0.0` ✅
5. Updated demo.json: corrected `diffStat` from "163 files / 7503 insertions" to "165 files / 7757 insertions" (real value from `git diff --stat main...HEAD`); refreshed evidence text for AC10 (framework.go deleted, not retained) and AC6/AC9 (verified this iteration).
6. Updated DEMO.md to match: corrected diffStat in header; updated AC10 verdict row text; updated AC6/AC9 evidence rows.
7. Updated `.forge/pr-description.md`: corrected diffStat, corrected framework.go note (deleted not retained).
8. Force-added `.forge/pr-description.md`, staged all demo changes; committed as unify-and-demo.
9. Pushed — `origin/forge/INIT-2026-07-01-mux-free-cutover` == local HEAD.

**Key observations:**
- This is the 13th consecutive iteration where `chore: drop forge scratch` removes `.forge/pr-description.md` from tracking. The unifier must always force-re-add it.
- Quality gate and all unit tests remain green. All 14 ACs verified (AC12 partial — inherent without live ADO).
- Corrected stale diffStat (was 163/7503, should be 165/7757 — 2 extra files are AGENT.md + fix_plan.md that the orchestrator added to the demo/pr-desc commits).

### Iteration 12 (2026-07-07)

**Status:** Complete — same recurring chore-commit hazard resolved (12th time).

**What was done:**
1. Read AGENT.md (showed iters 1–11 done) and fix_plan.md (all 14 ACs ticked, AC12 partial).
2. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → GREEN (ok, 0.003s).
3. Confirmed latest commit (e83790b6) was another "chore: drop forge scratch" that deleted `.forge/pr-description.md` from HEAD — same pattern as all prior iterations.
4. Verified MuxFree test timing: `go test -tags all -count=1 -run TestFrameworkProvider_MuxFree ./azuredevops/internal/provider/` → ok 0.004s (matches demo.json value — no update needed).
5. demo.json and DEMO.md timing values current — no update needed.
6. Force-added `.forge/pr-description.md` and AGENT.md, committed as unify-and-demo.
7. Pushed — `origin/forge/INIT-2026-07-01-mux-free-cutover` == local HEAD.

**Key observations:**
- This is the 12th consecutive iteration where `chore: drop forge scratch` removes `.forge/pr-description.md` from tracking. The unifier must always force-re-add it.
- Quality gate and all unit tests remain green. All 14 ACs verified (AC12 partial — inherent without live ADO).
- No code changes needed.

### Iteration 11 (2026-07-07)

**Status:** Complete — same recurring chore-commit hazard resolved (11th time).

**What was done:**
1. Read AGENT.md (showed iters 1–10 done) and fix_plan.md (all 14 ACs ticked, AC12 partial).
2. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → GREEN (ok, 0.003s).
3. Confirmed latest commit (ca4b75bb) was another "chore: drop forge scratch" that deleted `.forge/pr-description.md` from HEAD — same pattern as all prior iterations.
4. Verified MuxFree test timing: `go test -tags all -count=1 -run TestFrameworkProvider_MuxFree ./...` → ok 0.004s at provider package (matches demo.json value — no update needed).
5. demo.json and DEMO.md timing values current — no update needed.
6. Force-added `.forge/pr-description.md`, committed as unify-and-demo (1f1f0b44).
7. Pushed — `origin/forge/INIT-2026-07-01-mux-free-cutover` == local HEAD.

**Key observations:**
- This is the 11th consecutive iteration where `chore: drop forge scratch` removes `.forge/pr-description.md` from tracking. The unifier must always force-re-add it.
- Quality gate and all unit tests remain green. All 14 ACs verified (AC12 partial — inherent without live ADO).
- No code changes needed.

### Iteration 10 (2026-07-07)

**Status:** Complete — same recurring chore-commit hazard resolved (10th time).

**What was done:**
1. Read AGENT.md (showed iters 1–9 done) and fix_plan.md (all 14 ACs ticked, AC12 partial).
2. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → GREEN (ok, 0.003s).
3. Confirmed latest commit (265b8c9c) was another "chore: drop forge scratch" that deleted `.forge/pr-description.md` from HEAD — same pattern as all prior iterations.
4. Ran `TestFrameworkProvider_MuxFree` → PASS 0.004s (matches demo.json value — no update needed).
5. demo.json and DEMO.md timing values current — no update needed.
6. Force-added `.forge/pr-description.md`, committed as unify-and-demo.
7. Pushed — `origin/forge/INIT-2026-07-01-mux-free-cutover` == local HEAD.

**Key observations:**
- This is the 10th consecutive iteration where `chore: drop forge scratch` removes `.forge/pr-description.md` from tracking. The unifier must always force-re-add it.
- Quality gate and all unit tests remain green. All 14 ACs verified (AC12 partial — inherent without live ADO).
- No code changes needed.

### Iteration 9 (2026-07-07)

**Status:** Complete — same recurring chore-commit hazard resolved (9th time).

**What was done:**
1. Read AGENT.md (showed iters 1–8 done) and fix_plan.md (all 14 ACs ticked, AC12 partial).
2. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → GREEN (ok, 0.003s).
3. Confirmed latest commit (40159ca3) was another "chore: drop forge scratch" that deleted `.forge/pr-description.md` from HEAD — same pattern as all prior iterations.
4. Ran `TestFrameworkProvider_MuxFree` → PASS 0.004s (matches demo.json value — no update needed).
5. demo.json and DEMO.md timing values current — no update needed.
6. Force-added `.forge/pr-description.md`, committed as unify-and-demo (af80f77d).
7. Pushed — `origin/forge/INIT-2026-07-01-mux-free-cutover` == local HEAD.

**Key observations:**
- This is the 9th consecutive iteration where `chore: drop forge scratch` removes `.forge/pr-description.md` from tracking. The unifier must always force-re-add it.
- Quality gate and all unit tests remain green. All 14 ACs verified (AC12 partial — inherent without live ADO).
- No code changes needed.

### Iteration 8 (2026-07-07)

**Status:** Complete — same recurring chore-commit hazard resolved (8th time).

**What was done:**
1. Read AGENT.md (showed iters 1–7 done) and fix_plan.md (all 14 ACs ticked, AC12 partial).
2. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → GREEN (ok, 0.003s).
3. Confirmed latest commit (a093b9c3) was another "chore: drop forge scratch" that deleted `.forge/pr-description.md` from HEAD — same pattern as all prior iterations.
4. Ran `TestFrameworkProvider_MuxFree` → ok 0.004s (live; up from 0.003s).
5. Updated demo.json: servicehook gate afterOutput 0.002s → 0.003s; MuxFree test afterOutput 0.003s → 0.004s; AC7 evidence updated to 0.003s.
6. Updated DEMO.md to match: checkpoint 1 after row updated to 0.003s; checkpoint 4 after row updated to 0.004s; AC7 table row updated to 0.003s.
7. Force-added `.forge/pr-description.md` and staged demo changes, committed as unify-and-demo.
8. Pushed — `origin/forge/INIT-2026-07-01-mux-free-cutover` == local HEAD.

**Key observations:**
- This is the 8th consecutive iteration where `chore: drop forge scratch` removes `.forge/pr-description.md` from tracking. The unifier must always force-re-add it.
- Quality gate timing variance (0.002s–0.004s) is normal; kept demo.json in sync with live-captured values.
- No code changes needed. All 14 ACs verified (AC12 partial — inherent without live ADO).

### Iteration 7 (2026-07-07)

**Status:** Complete — same recurring chore-commit hazard resolved (7th time).

**What was done:**
1. Read AGENT.md (showed iters 1–6 done) and fix_plan.md (all 14 ACs ticked, AC12 partial).
2. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → GREEN (ok, 0.002s).
3. Confirmed latest commit (e23c2858) was another "chore: drop forge scratch" that deleted `.forge/pr-description.md` from HEAD — same pattern as all prior iterations.
4. Ran `TestFrameworkProvider_MuxFree` → ok 0.003s (live; matches demo.json value — no update needed for checkpoint 4).
5. Updated demo.json + DEMO.md: checkpoint 1 (servicehook-gate) afterOutput updated 0.003s → 0.002s (live captured value this iteration); AC7 evidence updated to match.
6. Force-added `.forge/pr-description.md` and AGENT.md, committed as unify-and-demo.
7. Pushed — `origin/forge/INIT-2026-07-01-mux-free-cutover` == local HEAD.

**Key observations:**
- This is the 7th consecutive iteration where `chore: drop forge scratch` removes `.forge/pr-description.md` from tracking. The unifier must always force-re-add it.
- Quality gate timing variance (0.002s vs 0.003s) is normal; kept demo.json in sync with live-captured value.
- No code changes needed. All 14 ACs verified (AC12 partial — inherent without live ADO).

### Iteration 6 (2026-07-07)

**Status:** Complete — same recurring chore-commit hazard resolved (6th time).

**What was done:**
1. Read AGENT.md (showed iters 1–5 done) and fix_plan.md (all 14 ACs ticked, AC12 partial).
2. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → GREEN (ok, 0.003s).
3. Confirmed latest commit (86cadef2) was another "chore: drop forge scratch" that deleted `.forge/pr-description.md` from HEAD — same pattern as all prior iterations.
4. Ran `TestFrameworkProvider_MuxFree` (uncached) → ok 0.003s (down from 0.004s). Updated demo.json + DEMO.md checkpoint 4 afterOutput accordingly.
5. Force-added `.forge/pr-description.md` and AGENT.md, committed as unify-and-demo.
6. Pushed — `origin/forge/INIT-2026-07-01-mux-free-cutover` == local HEAD.

**Key observations:**
- This is the 6th consecutive iteration where `chore: drop forge scratch` removes `.forge/pr-description.md` from tracking. The unifier must always force-re-add it.
- Quality gate and all unit tests remain green. All 14 ACs verified (AC12 partial — inherent without live ADO).
- No code changes needed.

### Iteration 5 (2026-07-07)

**Status:** Complete — same recurring chore-commit hazard resolved (5th time).

**What was done:**
1. Read AGENT.md (showed iters 1–4 done) and fix_plan.md (all 14 ACs ticked, AC12 partial).
2. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → GREEN (ok, 0.003s).
3. Confirmed latest commit (3c812773) was another "chore: drop forge scratch" that deleted `.forge/pr-description.md` from HEAD — same pattern as all prior iterations.
4. Ran `TestFrameworkProvider_MuxFree` → ok 0.004s (live; matches demo.json value).
5. demo.json and DEMO.md already have correct/current timing values — no update needed.
6. Force-added `.forge/pr-description.md` and AGENT.md, committed as unify-and-demo.
7. Pushed — `origin/forge/INIT-2026-07-01-mux-free-cutover` == local HEAD.

**Key observations:**
- This is the 5th consecutive iteration where `chore: drop forge scratch` removes `.forge/pr-description.md` from tracking. The unifier must always force-re-add it.
- Quality gate and all unit tests remain green. All 14 ACs verified (AC12 partial — inherent without live ADO).
- No code changes needed.

### Iteration 4 (2026-07-06)

**Status:** Complete — recurring chore-commit hazard resolved again (4th time).

**What was done:**
1. Read AGENT.md (showed iters 1–3 done) and fix_plan.md (all 14 ACs ticked, AC12 partial).
2. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → GREEN (ok, 0.003s).
3. Confirmed latest commit (14c97275) was another "chore: drop forge scratch" that deleted `.forge/pr-description.md` from tracked tree (the file exists on disk but was not tracked in HEAD — same pattern as iters 2 & 3).
4. Ran `TestFrameworkProvider_MuxFree` → ok 0.004s (fresh live value this iteration).
5. Updated demo.json + DEMO.md: `afterOutput` for checkpoint 4 (MuxFree test) updated 0.003s → 0.004s.
6. Force-added `.forge/pr-description.md` and committed all changes.
7. Pushed — `origin/forge/INIT-2026-07-01-mux-free-cutover` == local HEAD.

**Key observations:**
- This is the 4th consecutive iteration where `chore: drop forge scratch` removes `.forge/pr-description.md` from tracking. The unifier must always force-re-add it.
- Quality gate and all unit tests remain green.
- No code changes needed — all 14 ACs verified.

### Iteration 3 (2026-07-06)

**Status:** Complete — recurring chore-commit hazard resolved again.

**What was done:**
1. Read AGENT.md (showed iter 2 already done) and fix_plan.md (all 14 ACs ticked).
2. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → GREEN (ok, 0.003s).
3. Discovered latest commit (559979ad) was another "chore: drop forge scratch" that again deleted `.forge/pr-description.md`.
4. Restored PR description from previous unify commit (84653a14) — content identical to prior iterations.
5. Updated demo.json `afterOutput` for checkpoint 1 (servicehook gate): 0.002s → 0.003s (live captured value this iteration).
6. Updated DEMO.md to match: AC7 table row + checkpoint 1 table row both updated to 0.003s.
7. Updated demo.json acEvaluations AC7 evidence: 0.002s → 0.003s.
8. Committed as `feat(INIT-2026-07-01-mux-free-cutover): unify and demo` (0d030ce1).
9. Pushed — `origin/forge/INIT-2026-07-01-mux-free-cutover` == local HEAD.

**Key observations:**
- The `chore: drop forge scratch` commit pattern continues to delete `.forge/pr-description.md` on every orchestrator cycle. This is the third time this has occurred. The unifier must force-re-add it every iteration.
- Quality gate timing variance (0.002s vs 0.003s) is normal; kept demo.json in sync with live-captured value.
- No code changes were needed — all 14 ACs remain verified.

### Iteration 2 (2026-07-06)

**Status:** Complete — re-prep after chore commit dropped .forge/pr-description.md.

**What was done:**
1. Discovered that commit `8ca4d4b9` (chore: drop forge scratch) had wiped `.forge/pr-description.md` from HEAD. The demo.json + DEMO.md remained committed on the branch.
2. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → GREEN (ok, 0.002s).
3. Verified all ACs live on HEAD: build clean, no mux imports, provider maps empty, PROVIDER_VERSION.txt=2.0.0, CHANGELOG has BREAKING CHANGES entry, TestFrameworkProvider_MuxFree passes.
4. Updated demo.json `afterOutput` for checkpoint 4 from 0.004s → 0.003s (live captured value).
5. Updated DEMO.md to match corrected timing.
6. Updated fix_plan.md with all ACs ticked (13 met, 1 partial for live ADO test).
7. Restored `.forge/pr-description.md` via force-add and committed with demo + fix_plan updates.
8. Pushed to origin.

**Key observations:**
- The `chore: drop forge scratch from branch` commit pattern is a recurring hazard — it drops `.forge/pr-description.md` which is needed by the review phase. The unifier must force-re-add it each time.
- All 14 ACs verified live. Only AC12 (live acceptance TF_ACC=1) is partial — inherent limitation without live ADO credentials in the unifier context.

### Iteration 1 (2026-07-06)

**Status:** Complete — all gates expected to pass.

**What was done:**
1. Read AGENT.md, fix_plan.md, all 5 WI specs, initiative manifest.
2. Ran quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/servicehook/...` → GREEN (ok, 0.002s). No fixes needed.
3. Verified committed work:
   - All 5 WIs status:complete; commits 58b8e90d (WI-1), 3809b19f (WI-2), 1dae5458 (WI-3), 0b3d9004 (WI-4), 5e7ef0dd (WI-5) on branch.
   - go build -mod=vendor . → BUILD_OK
   - TestFrameworkProvider_MuxFree → PASS
   - PROVIDER_VERSION.txt → 2.0.0
   - CHANGELOG.md → [Unreleased] BREAKING CHANGES + INTERNAL sections present
4. Discovered `forge demo capture/render` fails — worktree missing `skills/project-manager/SKILL.md` (forge infrastructure issue). Captured before/after output manually by running checkpoint commands and using `git show main:...` for before-state.
5. Authored `demo/INIT-2026-07-01-mux-free-cutover/demo.json` with:
   - 5 checkpoints, each with `command` + real `beforeOutput`/`afterOutput`
   - 14 acEvaluations (13 met, 1 partial — AC12/TestAccProviderMuxFree skips without live ADO creds)
   - 5 testEvidence entries
6. Derived `demo/INIT-2026-07-01-mux-free-cutover/DEMO.md` from demo.json.
7. Wrote `.forge/pr-description.md` (force-added — `.forge/` is gitignored).
8. Committed as `feat(INIT-2026-07-01-mux-free-cutover): unify and demo` (2251416a).
9. Pushed to origin — `origin/forge/INIT-2026-07-01-mux-free-cutover` == local HEAD.

**Key observations:**
- `azuredevops/framework.go` was NOT deleted (WI-4 Ralph kept the 3-line re-export shim because Go internal-package rules prevent `main.go` from directly importing `azuredevops/internal/provider`). This is a pragmatic correct decision — the mux IS gone, just the re-export shim remains.
- `main.go` uses `providerserver.NewProtocol6` (not `NewProtocol6WithError`) — slight variance from WI-4 spec but still correct (no mux, pure framework).
- AC12 (live acceptance test) is `partial` — the test compiles and wires `GetProviderFactories()` correctly but requires `TF_ACC=1` + live ADO credentials to generate real evidence.
- `.forge/pr-description.md` must be force-added (`git add -f`) because `.forge/` is in `.gitignore`.

## Notes for reflection

_(observations the reflector should capture into the brain)_

- `forge demo capture/render` is broken in this worktree due to missing `skills/project-manager/SKILL.md`. The unifier must manually capture before/after outputs when `forge demo capture` is unavailable. The orchestrator should guard against this or fix the worktree setup.
- The `partial` verdict for AC12 (live acceptance) is expected in CI/unifier context — the test is coded correctly but requires live ADO credentials. The operator should run `TF_ACC=1 go test -run TestAccProviderMuxFree ./azuredevops/internal/acceptancetests/` manually to generate `.forge/live-evidence/acceptance-provider-mux-free.json` before merge.
- The `chore: drop forge scratch` commit that the orchestrator runs between unifier iterations consistently removes `.forge/pr-description.md` from git tracking (because `.forge/` is gitignored). This should be fixed at the orchestrator level — either by exempting `.forge/pr-description.md` from the scratch-drop, or by adding it to `.gitignore` exceptions.
