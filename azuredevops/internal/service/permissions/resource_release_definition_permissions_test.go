//go:build (all || permissions || resource_release_definition_permissions) && (!exclude_permissions || !resource_release_definition_permissions)
// +build all permissions resource_release_definition_permissions
// +build !exclude_permissions !resource_release_definition_permissions

package permissions

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

// getReleaseDefinitionPermissionsResource is a test helper that constructs a
// *schema.ResourceData for ResourceReleaseDefinitionPermissions.
// Pass definitionID < 0 to omit the release_definition_id field entirely
// (producing a project-scoped resource).
//
// NOTE: We pass values via the raw-map argument to schema.TestResourceDataRaw
// rather than via d.Set after the fact. This is important for Optional TypeInt
// fields: schema.TestResourceDataRaw builds a real Terraform diff from the raw
// config, so GetOk correctly distinguishes "field present but zero" from "field
// absent". Using d.Set on a zero-value TypeInt after construction does NOT set
// the "ok" bit.
func getReleaseDefinitionPermissionsResource(t *testing.T, projectID string, definitionID int) *schema.ResourceData {
	t.Helper()
	raw := map[string]interface{}{}
	if projectID != "" {
		raw["project_id"] = projectID
	}
	if definitionID >= 0 {
		raw["release_definition_id"] = definitionID
	}
	return schema.TestResourceDataRaw(t, ResourceReleaseDefinitionPermissions().Schema, raw)
}

// TestReleaseDefinitionPermissions_ProjectScopedToken verifies that
// createReleaseDefinitionToken returns only the projectID (no slash suffix)
// when release_definition_id is NOT set in the schema data.
//
// AC1: GIVEN no release_definition_id is set
//
//	WHEN createReleaseDefinitionToken is called
//	THEN it returns the bare projectID with no "/" suffix
func TestReleaseDefinitionPermissions_ProjectScopedToken(t *testing.T) {
	fakeProjectID := "21cff396-a36f-4d05-bccf-91e3a2a8b4bb"

	// Build resource data with only project_id set (definitionID=-1 → omitted).
	d := getReleaseDefinitionPermissionsResource(t, fakeProjectID, -1)

	token, err := createReleaseDefinitionToken(d, nil)

	require.NoError(t, err, "project-scoped token creation must not error")
	assert.Equal(t, fakeProjectID, token,
		"project-scoped token must equal projectID with no slash suffix")
	assert.NotContains(t, token, "/",
		"project-scoped token must not contain a '/' separator")
}

// TestReleaseDefinitionPermissions_TokenEdgeCases covers edge cases for
// createReleaseDefinitionToken when release_definition_id is explicitly set.
//
// AC2: GIVEN release_definition_id is set (even to 0)
//
//	WHEN createReleaseDefinitionToken is called
//	THEN the token is formatted as "projectId/definitionId"
//	AND definitionID=0 produces "projectId/0" (not a project-scoped path)
func TestReleaseDefinitionPermissions_TokenEdgeCases(t *testing.T) {
	fakeProjectID := "21cff396-a36f-4d05-bccf-91e3a2a8b4bb"

	cases := []struct {
		name         string
		definitionID int
		expected     string
	}{
		{
			name:         "definitionID=0 still produces projectId/0",
			definitionID: 0,
			expected:     fakeProjectID + "/0",
		},
		{
			name:         "definitionID=99999 produces projectId/99999",
			definitionID: 99999,
			expected:     fakeProjectID + "/99999",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			d := getReleaseDefinitionPermissionsResource(t, fakeProjectID, tc.definitionID)

			token, err := createReleaseDefinitionToken(d, nil)

			require.NoError(t, err, "token creation must not error for definitionID=%d", tc.definitionID)
			assert.Equal(t, tc.expected, token,
				"token for definitionID=%d must be %q", tc.definitionID, tc.expected)
			assert.True(t, strings.HasPrefix(token, fakeProjectID),
				"token must start with projectID")
			assert.Contains(t, token, "/",
				"definition-scoped token must contain '/' separator")
		})
	}
}
