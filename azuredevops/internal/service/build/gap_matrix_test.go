package build

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestBuildGapMatrixExists verifies that docs/build-gap-matrix.md has been produced.
// The gate runs: go test -tags all -run TestBuildGapMatrixExists ./azuredevops/internal/service/build/...
func TestBuildGapMatrixExists(t *testing.T) {
	// Walk up from this file's location to the repo root (5 levels up from
	// azuredevops/internal/service/build/ → repo root).
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file path via runtime.Caller")
	}

	// azuredevops/internal/service/build/gap_matrix_test.go
	// Go up 4 directories to reach the repo root.
	repoRoot := filepath.Dir(filename) // .../service/build
	for i := 0; i < 4; i++ {
		repoRoot = filepath.Dir(repoRoot)
	}

	gapMatrix := filepath.Join(repoRoot, "docs", "build-gap-matrix.md")
	if _, err := os.Stat(gapMatrix); os.IsNotExist(err) {
		t.Fatalf("docs/build-gap-matrix.md does not exist at expected path %s; produce it as part of WI-1", gapMatrix)
	}

	// Sanity-check: the file must be non-empty.
	info, err := os.Stat(gapMatrix)
	if err != nil {
		t.Fatalf("stat %s: %v", gapMatrix, err)
	}
	if info.Size() == 0 {
		t.Fatalf("docs/build-gap-matrix.md exists but is empty")
	}

	t.Logf("docs/build-gap-matrix.md found at %s (%d bytes)", gapMatrix, info.Size())
}
