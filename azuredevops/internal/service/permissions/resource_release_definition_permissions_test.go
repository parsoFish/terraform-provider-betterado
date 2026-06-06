//go:build (all || permissions || resource_release_definition_permissions) && (!exclude_permissions || !resource_release_definition_permissions)
// +build all permissions resource_release_definition_permissions
// +build !exclude_permissions !resource_release_definition_permissions

package permissions

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReleaseDefinitionPermissions_TokenFormatSpike asserts that:
//  1. The confirmed ReleaseManagement2 token format constant is non-empty.
//  2. A token produced for a fake projectID + definitionID starts with the projectID
//     and ends with the definition ID (i.e. matches "projectId/releaseDefinitionId").
//  3. The format constant itself contains the expected separator ("/").
//
// No live ADO call is made — this is a pure offline unit test.
func TestReleaseDefinitionPermissions_TokenFormatSpike(t *testing.T) {
	// 1. The format constant must be non-empty.
	assert.NotEmpty(t, releaseDefinitionTokenFormat,
		"releaseDefinitionTokenFormat must be a non-empty string")

	// 2. The format constant must contain "%" (it is a format string).
	assert.Contains(t, releaseDefinitionTokenFormat, "%",
		"releaseDefinitionTokenFormat must be a fmt format string")

	// 3. Render a token with fake IDs and assert structure.
	fakeProjectID := "21cff396-a36f-4d05-bccf-91e3a2a8b4bb"
	fakeDefinitionID := 7

	token := fmt.Sprintf(releaseDefinitionTokenFormat, fakeProjectID, fakeDefinitionID)

	require.NotEmpty(t, token, "rendered token must be non-empty")

	// Must start with the project UUID.
	assert.True(t, strings.HasPrefix(token, fakeProjectID),
		"token %q must start with projectID %q", token, fakeProjectID)

	// Must contain a "/" separator (confirmed from live probe).
	assert.Contains(t, token, "/",
		"token must contain '/' separator (confirmed from live ADO probe)")

	// Must end with the definition ID.
	expectedSuffix := fmt.Sprintf("%d", fakeDefinitionID)
	assert.True(t, strings.HasSuffix(token, expectedSuffix),
		"token %q must end with definitionID %q", token, expectedSuffix)

	// Full exact match — the confirmed format is "{projectId}/{releaseDefinitionId}".
	expectedToken := fmt.Sprintf("%s/%d", fakeProjectID, fakeDefinitionID)
	assert.Equal(t, expectedToken, token,
		"token format must be '{projectId}/{releaseDefinitionId}' (confirmed via live ADO probe against namespace c788c23e-1b46-4162-8f5e-d7585343b5de)")
}
