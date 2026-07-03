//go:build all || gallery

package gallery_test

import (
	"os"
	"strings"
	"testing"
)

// TestGalleryGapMatrix asserts that the Gallery / ExtensionManagement gap matrix
// document exists and contains the minimum required sections.
func TestGalleryGapMatrix(t *testing.T) {
	const docPath = "../../../../docs/gallery-extensionmanagement-gap-matrix.md"

	data, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("gap matrix doc not found at %s: %v", docPath, err)
	}

	content := string(data)

	requiredStrings := []string{
		"betterado_extension_install",
		"betterado_extension_settings",
		"betterado_marketplace_extension",
		"betterado_extension",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(content, required) {
			t.Errorf("gap matrix missing required string %q", required)
		}
	}
}
