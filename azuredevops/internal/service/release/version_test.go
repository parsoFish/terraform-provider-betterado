//go:build all || resource_release_definition

package release

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProviderVersion_BumpedToV1 asserts that PROVIDER_VERSION.txt at the
// repository root contains exactly "1.0.0\n", confirming the v1.0.0 release
// bump has been applied.
func TestProviderVersion_BumpedToV1(t *testing.T) {
	// Locate the module root: this test file lives at
	//   <root>/azuredevops/internal/service/release/version_test.go
	// so the module root is four directories up.
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller must succeed")

	moduleRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")
	versionFile := filepath.Join(moduleRoot, "PROVIDER_VERSION.txt")

	data, err := os.ReadFile(versionFile)
	require.NoError(t, err, "PROVIDER_VERSION.txt must be readable at %s", versionFile)

	assert.Equal(t, "1.0.0\n", string(data), "PROVIDER_VERSION.txt must read 1.0.0")
}
