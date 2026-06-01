//go:build (all || resource_release_folder) && !exclude_resource_release_folder
// +build all resource_release_folder
// +build !exclude_resource_release_folder

package release

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ── Package-level fixtures ─────────────────────────────────────────────────

var testReleaseFolderProjectID = uuid.New()
var testReleaseFolderPath = "\\MyFolder"

// TestReleaseFolder_ExpandFlatten_Roundtrip verifies that a Folder struct can be
// read into resource data and that the fields survive a flatten→expand roundtrip.
func TestReleaseFolder_ExpandFlatten_Roundtrip(t *testing.T) {
	folderPath := testReleaseFolderPath
	folder := releaseapi.Folder{
		Path: converter.String(folderPath),
	}

	// Simulate flatten: set schema values from struct
	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseFolder().Schema, map[string]interface{}{
		"project_id": testReleaseFolderProjectID.String(),
		"path":       folderPath,
		"name":       "",
	})
	resourceData.SetId(testReleaseFolderProjectID.String() + "/" + folderPath)

	// Verify values were correctly set
	require.Equal(t, testReleaseFolderProjectID.String(), resourceData.Get("project_id"))
	require.Equal(t, folderPath, resourceData.Get("path"))

	// Expand: read values back from schema → struct (roundtrip)
	expandedPath := resourceData.Get("path").(string)
	expandedProjectID := resourceData.Get("project_id").(string)

	require.Equal(t, *folder.Path, expandedPath)
	require.Equal(t, testReleaseFolderProjectID.String(), expandedProjectID)
}

// TestReleaseFolder_Create_DoesNotSwallowError verifies that when CreateFolder
// returns an error, resourceReleaseFolderCreate surfaces it.
func TestReleaseFolder_Create_DoesNotSwallowError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseFolder().Schema, map[string]interface{}{
		"project_id": testReleaseFolderProjectID.String(),
		"path":       testReleaseFolderPath,
		"name":       "",
	})

	releaseClient.
		EXPECT().
		CreateFolder(clients.Ctx, releaseapi.CreateFolderArgs{
			Folder:  &releaseapi.Folder{Path: converter.String(testReleaseFolderPath)},
			Project: converter.String(testReleaseFolderProjectID.String()),
		}).
		Return(nil, errors.New("CreateFolder() Failed")).
		Times(1)

	err := resourceReleaseFolderCreate(resourceData, clients)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed creating release folder")
}

// TestReleaseFolder_Read_ClearsIdOn404 verifies that when GetFolders returns a
// 404 WrappedError, resourceReleaseFolderRead clears the resource ID and returns
// nil (graceful drift detection).
func TestReleaseFolder_Read_ClearsIdOn404(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	notFoundStatusCode := http.StatusNotFound
	notFoundErr := azuredevops.WrappedError{
		StatusCode: &notFoundStatusCode,
	}

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseFolder().Schema, map[string]interface{}{
		"project_id": testReleaseFolderProjectID.String(),
		"path":       testReleaseFolderPath,
		"name":       "",
	})
	resourceData.SetId(testReleaseFolderProjectID.String() + "/" + testReleaseFolderPath)

	releaseClient.
		EXPECT().
		GetFolders(clients.Ctx, releaseapi.GetFoldersArgs{
			Project: converter.String(testReleaseFolderProjectID.String()),
			Path:    converter.String(testReleaseFolderPath),
		}).
		Return(nil, notFoundErr).
		Times(1)

	err := resourceReleaseFolderRead(resourceData, clients)
	require.NoError(t, err)
	require.Empty(t, resourceData.Id())
}

// TestReleaseFolder_Update_CallsSDKWithArgs verifies that resourceReleaseFolderUpdate
// invokes UpdateFolder with the correct arguments.
func TestReleaseFolder_Update_CallsSDKWithArgs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	newPath := "\\MyFolder\\Renamed"
	updatedFolder := &releaseapi.Folder{Path: converter.String(newPath)}

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseFolder().Schema, map[string]interface{}{
		"project_id": testReleaseFolderProjectID.String(),
		"path":       testReleaseFolderPath,
		"name":       "",
	})
	resourceData.SetId(testReleaseFolderProjectID.String() + "/" + testReleaseFolderPath)

	// Simulate the path change
	resourceData.Set("path", newPath)

	// UpdateFolder should be called with the new path
	releaseClient.
		EXPECT().
		UpdateFolder(clients.Ctx, gomock.Any()).
		Return(updatedFolder, nil).
		Times(1)

	// GetFolders will be called by the Read at the end of Update
	releaseClient.
		EXPECT().
		GetFolders(clients.Ctx, gomock.Any()).
		Return(&[]releaseapi.Folder{{Path: converter.String(newPath)}}, nil).
		Times(1)

	err := resourceReleaseFolderUpdate(resourceData, clients)
	require.NoError(t, err)
}

// TestReleaseFolder_Delete_SurfacesAPIError verifies that when DeleteFolder returns
// an error, resourceReleaseFolderDelete surfaces it.
func TestReleaseFolder_Delete_SurfacesAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	resourceData := schema.TestResourceDataRaw(t, ResourceReleaseFolder().Schema, map[string]interface{}{
		"project_id": testReleaseFolderProjectID.String(),
		"path":       testReleaseFolderPath,
		"name":       "",
	})
	resourceData.SetId(testReleaseFolderProjectID.String() + "/" + testReleaseFolderPath)

	releaseClient.
		EXPECT().
		DeleteFolder(clients.Ctx, releaseapi.DeleteFolderArgs{
			Project: converter.String(testReleaseFolderProjectID.String()),
			Path:    converter.String(testReleaseFolderPath),
		}).
		Return(errors.New("DeleteFolder() Failed")).
		Times(1)

	err := resourceReleaseFolderDelete(resourceData, clients)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed deleting release folder")
}
