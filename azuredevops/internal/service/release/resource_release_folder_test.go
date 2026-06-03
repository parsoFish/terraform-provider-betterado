//go:build (all || resource_release_folder) && !exclude_resource_release_folder
// +build all resource_release_folder
// +build !exclude_resource_release_folder

package release

import (
	"testing"

	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/stretchr/testify/require"
)

// TestReleaseFolder_ExpandFlatten_Roundtrip verifies that expandReleaseFolder and
// flattenReleaseFolder preserve path, description, and project_id without data loss.
func TestReleaseFolder_ExpandFlatten_Roundtrip(t *testing.T) {
	t.Parallel()

	projectID := "00000000-0000-0000-0000-000000000001"
	path := `\Production\Web`
	description := "Production web deployments"

	// Build a ResourceData using the schema from ResourceReleaseFolder.
	r := ResourceReleaseFolder()
	d := r.TestResourceData()
	d.Set("project_id", projectID)   //nolint:errcheck
	d.Set("path", path)              //nolint:errcheck
	d.Set("description", description) //nolint:errcheck

	// Expand to SDK struct.
	folder := expandReleaseFolder(d)
	require.NotNil(t, folder)
	require.Equal(t, path, *folder.Path)
	require.Equal(t, description, *folder.Description)

	// Simulate an API response by adding project_id back (the API does not echo it).
	apiResponse := &releaseapi.Folder{
		Path:        converter.String(path),
		Description: converter.String(description),
	}

	// Flatten back into a fresh ResourceData.
	d2 := r.TestResourceData()
	d2.Set("project_id", projectID) //nolint:errcheck
	flattenReleaseFolder(d2, apiResponse)

	// Assert round-trip fidelity.
	require.Equal(t, projectID, d2.Get("project_id"))
	require.Equal(t, path, d2.Get("path"))
	require.Equal(t, description, d2.Get("description"))
}
