//go:build all || resource_release_definition

package release

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProviderVersion_BumpedToV1 asserts that PROVIDER_VERSION.txt at the
// repository root contains a bare MAJOR.MINOR.PATCH semver at major version >= 1,
// confirming the v1.x release bump has been applied. (Robust across patch/minor
// releases — it does not hardcode the exact version.)
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

	version := strings.TrimSpace(string(data))
	require.Regexp(t, regexp.MustCompile(`^\d+\.\d+\.\d+$`), version,
		"PROVIDER_VERSION.txt must contain a bare MAJOR.MINOR.PATCH semver")

	major, err := strconv.Atoi(strings.SplitN(version, ".", 2)[0])
	require.NoError(t, err)
	assert.GreaterOrEqual(t, major, 1, "provider must be at major version >= 1")
}
