//go:build all || pipelinesapproval
// +build all pipelinesapproval

package pipelinesapproval_test

import (
	"os"
	"strings"
	"testing"
)

// TestPipelinesApprovalGapMatrix verifies that docs/pipelinesapproval-gap-matrix.md
// exists and contains the required sections documenting the ADO Pipelines Approval
// API coverage.
func TestPipelinesApprovalGapMatrix(t *testing.T) {
	const matrixPath = "../../../../docs/pipelinesapproval-gap-matrix.md"

	data, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("docs/pipelinesapproval-gap-matrix.md must exist: %v", err)
	}
	content := string(data)

	requiredStrings := []struct {
		desc    string
		contain string
	}{
		// AC1: lists every ADO Pipelines Approval API method
		{"GetApproval API method documented", "GetApproval"},
		{"QueryApprovals API method documented", "QueryApprovals"},
		{"UpdateApprovals API method documented", "UpdateApprovals"},
		// AC1: distinguishes declarative-manageable state from ephemeral-only
		{"UpdateApprovals is the write operation", "UpdateApprovals"},
		{"declarative state section present", "Declarative"},
		{"ephemeral section present", "ephemeral"},
		// AC1: notes betterado resource coverage for each field
		{"status field coverage documented", "status"},
		{"comment field coverage documented", "comment"},
		// AC2: betterado_pipeline_approval manages the approval decision
		{"betterado_pipeline_approval resource documented", "betterado_pipeline_approval"},
		// AC2: betterado_pipeline_approvals lists pending approvals
		{"betterado_pipeline_approvals data source documented", "betterado_pipeline_approvals"},
		// AC2: approval IDs are ephemeral (not importable)
		{"approval IDs are ephemeral", "ephemeral"},
		{"import not supported", "not importable"},
		{"bound to a specific pipeline run", "pipeline run"},
	}

	for _, tc := range requiredStrings {
		if !strings.Contains(content, tc.contain) {
			t.Errorf("gap matrix missing required content (%s): expected to find %q", tc.desc, tc.contain)
		}
	}
}
