//go:build all

package acceptancetests

// TestAccGraphGapMatrix verifies that the Graph and Identity gap matrix documents
// exist in the repository and contain the required sections mandated by WI-1.
//
// This test is a documentation-gate test: it does NOT require TF_ACC or live ADO
// credentials. It runs every time the acceptance-test suite is invoked with
// `-tags all -run TestAccGraphGapMatrix` and confirms the deliverables of WI-1
// are present in the working tree.
//
// Quality gate: go test -tags all -run TestAccGraphGapMatrix ./azuredevops/internal/acceptancetests/

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// repoRoot returns the absolute path to the repository root by walking up from
// the test file's source location until go.mod is found.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed — cannot locate repo root")
	}
	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find go.mod walking up from test file; repo root unknown")
		}
		dir = parent
	}
}

// requiredGraphSections are the section headings that must appear in docs/graph-gap-matrix.md.
var requiredGraphSections = []string{
	"# Graph API Schema Gap Matrix",
	"## 1. GraphGroup",
	"betterado_group",
	"## 2. GraphDescriptorResult",
	"betterado_descriptor",
	"## 3. GraphStorageKeyResult",
	"betterado_storage_key",
	"## 4. GraphMembership",
	"betterado_group_membership",
	"## 5. GraphGroup list",
	"betterado_groups",
	"## 6. GraphServicePrincipal",
	"betterado_service_principal",
	"## 7. GraphUser",
	"betterado_user",
	"betterado_users",
	"## Overall Summary",
	"## Writable Gaps",
	"supported",
	"gap",
	"deferred",
}

// requiredIdentitySections are the section headings that must appear in docs/identity-gap-matrix.md.
var requiredIdentitySections = []string{
	"# Identity API Schema Gap Matrix",
	"## 1. Identity struct",
	"## 2. `betterado_identity_group`",
	"## 3. `betterado_identity_groups`",
	"## 4. `betterado_identity_user`",
	"## Overall Summary",
	"## Writable Gaps",
	"supported",
	"gap",
	"deferred",
	"implement in this initiative",
}

func TestAccGraphGapMatrix(t *testing.T) {
	root := repoRoot(t)

	t.Run("graph_gap_matrix_exists", func(t *testing.T) {
		path := filepath.Join(root, "docs", "graph-gap-matrix.md")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("docs/graph-gap-matrix.md missing or unreadable: %v\n"+
				"  Expected path: %s\n"+
				"  This file must be committed as part of WI-1 (AC1).", err, path)
		}
		if len(content) < 500 {
			t.Errorf("docs/graph-gap-matrix.md exists but appears to be a stub (only %d bytes); "+
				"it must contain the full field-by-field gap matrix", len(content))
		}
		text := string(content)
		for _, section := range requiredGraphSections {
			if !strings.Contains(text, section) {
				t.Errorf("docs/graph-gap-matrix.md is missing required content %q", section)
			}
		}
	})

	t.Run("identity_gap_matrix_exists", func(t *testing.T) {
		path := filepath.Join(root, "docs", "identity-gap-matrix.md")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("docs/identity-gap-matrix.md missing or unreadable: %v\n"+
				"  Expected path: %s\n"+
				"  This file must be committed as part of WI-1 (AC2).", err, path)
		}
		if len(content) < 500 {
			t.Errorf("docs/identity-gap-matrix.md exists but appears to be a stub (only %d bytes); "+
				"it must contain the full field-by-field gap matrix", len(content))
		}
		text := string(content)
		for _, section := range requiredIdentitySections {
			if !strings.Contains(text, section) {
				t.Errorf("docs/identity-gap-matrix.md is missing required content %q", section)
			}
		}
	})

	t.Run("graph_gap_matrix_covers_all_resources", func(t *testing.T) {
		path := filepath.Join(root, "docs", "graph-gap-matrix.md")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Skipf("docs/graph-gap-matrix.md not found (checked by previous sub-test): %v", err)
		}
		text := string(content)

		requiredResources := []string{
			"betterado_group",
			"betterado_descriptor",
			"betterado_storage_key",
			"betterado_group_membership",
			"betterado_groups",
			"betterado_service_principal",
			"betterado_user",
			"betterado_users",
		}
		for _, res := range requiredResources {
			if !strings.Contains(text, res) {
				t.Errorf("docs/graph-gap-matrix.md does not mention resource/data-source %q (required by AC1)", res)
			}
		}
	})

	t.Run("identity_gap_matrix_covers_all_data_sources", func(t *testing.T) {
		path := filepath.Join(root, "docs", "identity-gap-matrix.md")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Skipf("docs/identity-gap-matrix.md not found (checked by previous sub-test): %v", err)
		}
		text := string(content)

		requiredDataSources := []string{
			"betterado_identity_group",
			"betterado_identity_groups",
			"betterado_identity_user",
		}
		for _, ds := range requiredDataSources {
			if !strings.Contains(text, ds) {
				t.Errorf("docs/identity-gap-matrix.md does not mention data-source %q (required by AC2)", ds)
			}
		}
	})

	t.Run("writable_gaps_have_implementation_decisions", func(t *testing.T) {
		// AC3: every writable gap must be marked 'implement in this initiative' or 'deferred'
		for _, file := range []string{"docs/graph-gap-matrix.md", "docs/identity-gap-matrix.md"} {
			path := filepath.Join(root, file)
			content, err := os.ReadFile(path)
			if err != nil {
				t.Errorf("%s not found: %v", file, err)
				continue
			}
			text := string(content)
			if !strings.Contains(text, "implement in this initiative") && !strings.Contains(text, "deferred") {
				t.Errorf("%s must contain implementation decisions for writable gaps "+
					"(either 'implement in this initiative' or 'deferred' with rationale) — AC3", file)
			}
		}
	})
}
