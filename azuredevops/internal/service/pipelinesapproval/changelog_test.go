//go:build all || pipelinesapproval
// +build all pipelinesapproval

package pipelinesapproval_test

import (
	"os"
	"strings"
	"testing"
)

// TestChangelog_HasPipelineApprovalEntry reads CHANGELOG.md from the repo root
// and asserts that the ## [Unreleased] section's ### FEATURES block contains an
// entry for the betterado_pipeline_approval resource and the
// betterado_pipeline_approvals data source.
func TestChangelog_HasPipelineApprovalEntry(t *testing.T) {
	// Path is relative to this package directory: ../../../../CHANGELOG.md
	const changelogPath = "../../../../CHANGELOG.md"

	data, err := os.ReadFile(changelogPath)
	if err != nil {
		t.Fatalf("CHANGELOG.md must exist at repo root: %v", err)
	}
	content := string(data)

	// Locate ## [Unreleased] section
	unreleasedIdx := strings.Index(content, "## [Unreleased]")
	if unreleasedIdx == -1 {
		t.Fatal("CHANGELOG.md must contain a ## [Unreleased] section")
	}

	// Find the next versioned section (## [x.y.z]) to bound the Unreleased content
	afterUnreleased := content[unreleasedIdx:]
	nextSectionIdx := strings.Index(afterUnreleased[len("## [Unreleased]"):], "\n## [")
	var unreleasedContent string
	if nextSectionIdx == -1 {
		unreleasedContent = afterUnreleased
	} else {
		unreleasedContent = afterUnreleased[:len("## [Unreleased]")+nextSectionIdx]
	}

	checks := []struct {
		desc    string
		contain string
	}{
		{
			desc:    "betterado_pipeline_approval resource entry",
			contain: "betterado_pipeline_approval",
		},
		{
			desc:    "betterado_pipeline_approvals data source entry",
			contain: "betterado_pipeline_approvals",
		},
		{
			desc:    "entry is under ### FEATURES",
			contain: "### FEATURES",
		},
	}

	for _, tc := range checks {
		if !strings.Contains(unreleasedContent, tc.contain) {
			t.Errorf("CHANGELOG.md [Unreleased] section missing required content (%s): expected to find %q", tc.desc, tc.contain)
		}
	}
}
