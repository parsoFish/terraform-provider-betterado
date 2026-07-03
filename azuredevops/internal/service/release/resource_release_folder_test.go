//go:build (all || resource_release_folder) && !exclude_resource_release_folder
// +build all resource_release_folder
// +build !exclude_resource_release_folder

package release

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ── Package-level fixtures ─────────────────────────────────────────────────

var (
	testReleaseFolderProjectID   = uuid.New()
	testReleaseFolderPath        = `\MyFolder`
	testReleaseFolderDescription = "Test folder description"
)

var testReleaseFolder = releaseapi.Folder{
	Path:        converter.String(testReleaseFolderPath),
	Description: converter.String(testReleaseFolderDescription),
}

// ── 1. Expand / Flatten roundtrip ─────────────────────────────────────────

// TestReleaseFolderExpand verifies that expand → flatten preserves all schema fields.
func TestReleaseFolderExpand(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseFolder().Schema, map[string]interface{}{
		"project_id":  testReleaseFolderProjectID.String(),
		"path":        testReleaseFolderPath,
		"description": testReleaseFolderDescription,
	})

	folder, projectID := expandReleaseFolder(resourceData)

	require.Equal(t, testReleaseFolderProjectID.String(), projectID)
	require.Equal(t, testReleaseFolderPath, *folder.Path)
	require.Equal(t, testReleaseFolderDescription, *folder.Description)

	// Flatten back and verify fields round-trip correctly.
	flattenReleaseFolder(resourceData, folder, projectID)

	require.Equal(t, testReleaseFolderPath, resourceData.Id())
	require.Equal(t, testReleaseFolderProjectID.String(), resourceData.Get("project_id"))
	require.Equal(t, testReleaseFolderPath, resourceData.Get("path"))
	require.Equal(t, testReleaseFolderDescription, resourceData.Get("description"))
}

// ── 2. Create error ────────────────────────────────────────────────────────

// TestReleaseFolderCreate_Error verifies that a CreateFolder API error surfaces as a diag error.
func TestReleaseFolderCreate_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRelease := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: mockRelease,
		Ctx:           context.Background(),
	}

	expectedErr := errors.New("create folder API error")
	projectID := testReleaseFolderProjectID.String()

	mockRelease.EXPECT().
		CreateFolder(gomock.Any(), gomock.Any()).
		Return(nil, expectedErr)

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseFolder().Schema, map[string]interface{}{
		"project_id":  projectID,
		"path":        testReleaseFolderPath,
		"description": testReleaseFolderDescription,
	})

	diags := resourceReleaseFolderCreate(context.Background(), resourceData, clients)
	require.NotEmpty(t, diags)
	require.Contains(t, diags[0].Summary, "creating release folder")
}

// ── 3. Read 404 clears ID ──────────────────────────────────────────────────

// TestReleaseFolderRead_404ClearsID verifies that an empty GetFolders response clears the resource ID.
func TestReleaseFolderRead_404ClearsID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRelease := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: mockRelease,
		Ctx:           context.Background(),
	}

	projectID := testReleaseFolderProjectID.String()
	emptyFolders := &[]releaseapi.Folder{}

	mockRelease.EXPECT().
		GetFolders(gomock.Any(), gomock.Any()).
		Return(emptyFolders, nil)

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseFolder().Schema, map[string]interface{}{
		"project_id":  projectID,
		"path":        testReleaseFolderPath,
		"description": "",
	})
	resourceData.SetId(testReleaseFolderPath)

	diags := resourceReleaseFolderRead(context.Background(), resourceData, clients)
	require.Empty(t, diags)
	require.Equal(t, "", resourceData.Id(), "ID should be cleared when folder is not found")
}

// ── 4. Update args ─────────────────────────────────────────────────────────

// TestReleaseFolderUpdate_Args verifies that UpdateFolder is called with the correct path and description.
func TestReleaseFolderUpdate_Args(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRelease := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: mockRelease,
		Ctx:           context.Background(),
	}

	projectID := testReleaseFolderProjectID.String()
	updatedDescription := "Updated description"
	updatedFolder := releaseapi.Folder{
		Path:        converter.String(testReleaseFolderPath),
		Description: converter.String(updatedDescription),
	}
	folders := &[]releaseapi.Folder{updatedFolder}

	mockRelease.EXPECT().
		UpdateFolder(gomock.Any(), gomock.Eq(releaseapi.UpdateFolderArgs{
			Folder: &releaseapi.Folder{
				Path:        converter.String(testReleaseFolderPath),
				Description: converter.String(updatedDescription),
			},
			Project: &projectID,
			Path:    converter.String(testReleaseFolderPath),
		})).
		Return(&updatedFolder, nil)

	mockRelease.EXPECT().
		GetFolders(gomock.Any(), gomock.Any()).
		Return(folders, nil)

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseFolder().Schema, map[string]interface{}{
		"project_id":  projectID,
		"path":        testReleaseFolderPath,
		"description": updatedDescription,
	})
	resourceData.SetId(testReleaseFolderPath)

	diags := resourceReleaseFolderUpdate(context.Background(), resourceData, clients)
	require.Empty(t, diags)
}

// ── 5. Delete error ────────────────────────────────────────────────────────

// TestReleaseFolderDelete_Error verifies that a DeleteFolder API error surfaces as a diag error.
func TestReleaseFolderDelete_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRelease := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: mockRelease,
		Ctx:           context.Background(),
	}

	projectID := testReleaseFolderProjectID.String()
	expectedErr := errors.New("delete folder API error")

	mockRelease.EXPECT().
		DeleteFolder(gomock.Any(), gomock.Any()).
		Return(expectedErr)

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseFolder().Schema, map[string]interface{}{
		"project_id":  projectID,
		"path":        testReleaseFolderPath,
		"description": "",
	})
	resourceData.SetId(testReleaseFolderPath)

	diags := resourceReleaseFolderDelete(context.Background(), resourceData, clients)
	require.NotEmpty(t, diags)
	require.Contains(t, diags[0].Summary, "deleting release folder")
}
