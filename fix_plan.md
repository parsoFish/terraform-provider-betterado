# Fix Plan — unifier sub-phase

> Initiative-level acceptance criteria. Tick each as you prove it against branch tip. Iteration 1 is initial prep; iterations 2+ react to either gate failures or send-back feedback.

- [x] AC1 (WI-1): GIVEN a clean worktree with no resource_task_group_test.go present WHEN the quality gate runs before any agent work THEN the gate exits non-zero (no TestTaskGroup_ tests exist yet)
- [x] AC2 (WI-1): GIVEN the new file azuredevops/internal/service/taskagent/resource_task_group_test.go has been created with all five TestTaskGroup_* functions WHEN go test -mod=vendor -tags all -count=1 -run TestTaskGroup ./azuredevops/internal/service/taskagent/ runs THEN all five tests show --- PASS: and the gate exits 0
- [x] AC3 (WI-1): GIVEN the new test file is written WHEN go build -mod=vendor . runs THEN the build exits 0 (no compilation errors introduced)
- [x] AC4 (WI-1): GIVEN the initiative scope fence WHEN git diff is inspected after the agent completes THEN only resource_task_group_test.go is added; resource_task_group.go and azdosdkmocks/taskagent_sdk_mock.go are unchanged
