package release

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestAuditGapMatrixDocExists verifies that docs/release-definition-gap-matrix.md exists in
// the repository root and contains at least 50 non-empty lines.
//
// This is a living-document gate: it fails if the file is absent or trivially empty,
// ensuring that the gap-matrix doc is always present alongside the schema code.
func TestAuditGapMatrixDocExists(t *testing.T) {
	t.Helper()

	// Locate the repository root by walking up from this test file.
	// The test lives at azuredevops/internal/service/release/, so the root is 4 levels up.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed — cannot locate test file path")
	}
	// thisFile: …/azuredevops/internal/service/release/doc_audit_test.go
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")
	repoRoot = filepath.Clean(repoRoot)

	docPath := filepath.Join(repoRoot, "docs", "release-definition-gap-matrix.md")

	info, err := os.Stat(docPath)
	if os.IsNotExist(err) {
		t.Fatalf("docs/release-definition-gap-matrix.md does not exist at expected path %s", docPath)
	}
	if err != nil {
		t.Fatalf("stat(%s): %v", docPath, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory, not a file", docPath)
	}

	// Count non-empty lines.
	f, err := os.Open(docPath)
	if err != nil {
		t.Fatalf("open(%s): %v", docPath, err)
	}
	defer f.Close()

	var lineCount int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := scanner.Text(); line != "" {
			lineCount++
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanning %s: %v", docPath, err)
	}

	const minLines = 50
	if lineCount < minLines {
		t.Fatalf("docs/release-definition-gap-matrix.md has only %d non-empty lines; expected at least %d", lineCount, minLines)
	}

	t.Logf("docs/release-definition-gap-matrix.md OK: %d non-empty lines (>= %d)", lineCount, minLines)
}

// TestAuditRoadmapDocExists verifies that docs/release-definition-roadmap.md exists in
// the repository root and contains at least 20 non-empty lines.
//
// This is a living-document gate: it fails if the file is absent or trivially empty,
// ensuring that the roadmap doc is always present alongside the schema code.
func TestAuditRoadmapDocExists(t *testing.T) {
	t.Helper()

	// Locate the repository root by walking up from this test file.
	// The test lives at azuredevops/internal/service/release/, so the root is 4 levels up.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed — cannot locate test file path")
	}
	// thisFile: …/azuredevops/internal/service/release/doc_audit_test.go
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")
	repoRoot = filepath.Clean(repoRoot)

	docPath := filepath.Join(repoRoot, "docs", "release-definition-roadmap.md")

	info, err := os.Stat(docPath)
	if os.IsNotExist(err) {
		t.Fatalf("docs/release-definition-roadmap.md does not exist at expected path %s", docPath)
	}
	if err != nil {
		t.Fatalf("stat(%s): %v", docPath, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory, not a file", docPath)
	}

	// Count non-empty lines.
	f, err := os.Open(docPath)
	if err != nil {
		t.Fatalf("open(%s): %v", docPath, err)
	}
	defer f.Close()

	var lineCount int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := scanner.Text(); line != "" {
			lineCount++
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanning %s: %v", docPath, err)
	}

	const minLines = 20
	if lineCount < minLines {
		t.Fatalf("docs/release-definition-roadmap.md has only %d non-empty lines; expected at least %d", lineCount, minLines)
	}

	t.Logf("docs/release-definition-roadmap.md OK: %d non-empty lines (>= %d)", lineCount, minLines)
}
