# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the ADO Test REST API v7.1 surfaces (_apis/test/plans, _apis/testplan/*, _apis/test/runs, _apis/test/results) WHEN the gap matrix is authored by reading azdosdkmocks/test_sdk_mock.go and the Go SDK test package THEN docs/test-gap-matrix.md exists and lists every resource type with status (implement-as-resource / implement-as-data-source / read-only / out-of-scope) plus an explicit declarative-vs-ephemeral rationale for each, including test runs and results
- [x] AC2: GIVEN the gap matrix is complete WHEN a developer reads it THEN it is clear which types the subsequent WIs implement and why test_run / test_result are data-source-only (ephemeral execution artifacts)
