//go:build all || wiki

package wiki

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestWikiGapMatrix_FileExists verifies that docs/wiki-gap-matrix.md exists in
// the repository root and contains at least 20 non-empty lines.
//
// This is the quality-gate sentinel for WI-1: it fails on a clean tree (file
// absent) and passes once the gap-matrix document has been written.
func TestWikiGapMatrix_FileExists(t *testing.T) {
	t.Helper()

	// Locate the repository root by walking up from this test file.
	// The test lives at azuredevops/internal/service/wiki/, so the root is 4 levels up.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed — cannot locate test file path")
	}
	// thisFile: …/azuredevops/internal/service/wiki/gap_matrix_test.go
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", ".."))

	docPath := filepath.Join(repoRoot, "docs", "wiki-gap-matrix.md")

	info, err := os.Stat(docPath)
	if os.IsNotExist(err) {
		t.Fatalf("docs/wiki-gap-matrix.md does not exist at expected path %s", docPath)
	}
	if err != nil {
		t.Fatalf("stat(%s): %v", docPath, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory, not a file", docPath)
	}

	// Count non-empty lines to ensure the file is substantive.
	f, err := os.Open(docPath)
	if err != nil {
		t.Fatalf("open(%s): %v", docPath, err)
	}
	defer f.Close()

	var lineCount int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			lineCount++
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanning %s: %v", docPath, err)
	}

	const minLines = 20
	if lineCount < minLines {
		t.Fatalf("docs/wiki-gap-matrix.md has only %d non-empty lines; expected at least %d", lineCount, minLines)
	}

	t.Logf("docs/wiki-gap-matrix.md OK: %d non-empty lines (>= %d)", lineCount, minLines)
}
