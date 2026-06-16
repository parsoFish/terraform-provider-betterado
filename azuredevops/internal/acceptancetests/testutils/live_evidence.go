package testutils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// CaptureLiveEvidence persists a real API GET response as forge demo
// "live-evidence": proof the demo can SHOW the actual resource (the request URL +
// the response body), not merely name the tests that touched it. It writes
// <repo-root>/.forge/live-evidence/<label>.json which forge's demo phase reads
// and back-fills into demo.json (checkpoint.liveEvidence) — the artifact the
// betterado live-evidence gate requires.
//
// Call it from a live acceptance test's Check step, while the resource still
// exists (BEFORE destroy), passing the response object you already fetched for
// the read-back assertion. Best-effort by contract: it returns an error for the
// caller to ignore — capturing demo evidence must never fail a real assertion.
func CaptureLiveEvidence(label, url string, response interface{}) error {
	root := repoRoot()
	if root == "" {
		return nil // not in a repo checkout — nothing to capture into
	}
	dir := filepath.Join(root, ".forge", "live-evidence")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	var respStr string
	switch v := response.(type) {
	case string:
		respStr = v
	case nil:
		respStr = ""
	default:
		if b, err := json.MarshalIndent(response, "", "  "); err == nil {
			respStr = string(b)
		}
	}

	payload := map[string]interface{}{
		"label":      label,
		"url":        url,
		"capturedAt": time.Now().UTC().Format(time.RFC3339),
		"response":   respStr,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, label+".json"), b, 0o644)
}

// repoRoot walks up from the working directory to the first dir containing go.mod
// (the repo root, where forge's .forge/ also lives). Returns "" if none is found.
func repoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
