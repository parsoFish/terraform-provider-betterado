package release

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// countNonEmptyLines opens the file at path and counts non-empty (non-whitespace-only) lines.
func countNonEmptyLines(t *testing.T, path string) int {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open(%s): %v", path, err)
	}
	defer f.Close()

	var count int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanning %s: %v", path, err)
	}
	return count
}

// TestDataSourceDocPagesExist verifies that the documentation pages for the two new
// data sources exist under docs/data-sources/ and each contains at least 10 non-empty lines.
func TestDataSourceDocPagesExist(t *testing.T) {
	t.Helper()

	// Locate the repository root by walking up from this test file.
	// The test lives at azuredevops/internal/service/release/, so the root is 4 levels up.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed — cannot locate test file path")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", ".."))

	const minLines = 10

	docFiles := []string{
		filepath.Join(repoRoot, "docs", "data-sources", "release_definition_revision.md"),
		filepath.Join(repoRoot, "docs", "data-sources", "release_definition_history.md"),
	}

	for _, docPath := range docFiles {
		rel, _ := filepath.Rel(repoRoot, docPath)

		info, err := os.Stat(docPath)
		if os.IsNotExist(err) {
			t.Fatalf("%s does not exist at expected path %s", rel, docPath)
		}
		if err != nil {
			t.Fatalf("stat(%s): %v", docPath, err)
		}
		if info.IsDir() {
			t.Fatalf("%s is a directory, not a file", docPath)
		}

		n := countNonEmptyLines(t, docPath)
		if n < minLines {
			t.Fatalf("%s has only %d non-empty lines; expected at least %d", rel, n, minLines)
		}

		t.Logf("%s OK: %d non-empty lines (>= %d)", rel, n, minLines)
	}
}

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
