//go:build (all || resource_release_definition) && !exclude_resource_release_definition
// +build all resource_release_definition
// +build !exclude_resource_release_definition

package release

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	webapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestDataReleaseDefinitionHistory_Read_Populates verifies that when
// GetReleaseDefinitionHistory returns two revisions the data source sets
// the revisions list with the correct values.
func TestDataReleaseDefinitionHistory_Read_Populates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	projectID := testReleaseDefinitionProjectID.String()
	defID := 42

	rev1 := 1
	rev2 := 2
	displayName1 := "Alice"
	displayName2 := "Bob"
	changeType1 := releaseapi.AuditActionValues.Add
	changeType2 := releaseapi.AuditActionValues.Update
	comment1 := "Initial creation"
	comment2 := "Updated trigger"
	changedTime1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	changedTime2 := time.Date(2024, 2, 20, 12, 30, 0, 0, time.UTC)

	releaseClient.
		EXPECT().
		GetReleaseDefinitionHistory(clients.Ctx, releaseapi.GetReleaseDefinitionHistoryArgs{
			Project:      &projectID,
			DefinitionId: &defID,
		}).
		Return(&[]releaseapi.ReleaseDefinitionRevision{
			{
				Revision:    &rev1,
				ChangedBy:   &webapi.IdentityRef{DisplayName: &displayName1},
				ChangedDate: &azuredevops.Time{Time: changedTime1},
				ChangeType:  &changeType1,
				Comment:     &comment1,
			},
			{
				Revision:    &rev2,
				ChangedBy:   &webapi.IdentityRef{DisplayName: &displayName2},
				ChangedDate: &azuredevops.Time{Time: changedTime2},
				ChangeType:  &changeType2,
				Comment:     &comment2,
			},
		}, nil).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, DataReleaseDefinitionHistory().Schema, map[string]interface{}{
		"project_id":            projectID,
		"release_definition_id": defID,
	})

	diags := dataReleaseDefinitionHistoryRead(context.Background(), resourceData, clients)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	rawList := resourceData.Get("revisions").([]interface{})
	require.Len(t, rawList, 2)

	first := rawList[0].(map[string]interface{})
	require.Equal(t, rev1, first["revision"])
	require.Equal(t, displayName1, first["changed_by"])
	require.Equal(t, changedTime1.Format(time.RFC3339), first["changed_date"])
	require.Equal(t, string(changeType1), first["change_type"])
	require.Equal(t, comment1, first["comment"])

	second := rawList[1].(map[string]interface{})
	require.Equal(t, rev2, second["revision"])
	require.Equal(t, displayName2, second["changed_by"])
}

// TestDataReleaseDefinitionHistory_Read_Empty verifies that when
// GetReleaseDefinitionHistory returns an empty slice the revisions list is
// empty and no error is returned.
func TestDataReleaseDefinitionHistory_Read_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	projectID := testReleaseDefinitionProjectID.String()
	defID := 99

	releaseClient.
		EXPECT().
		GetReleaseDefinitionHistory(clients.Ctx, releaseapi.GetReleaseDefinitionHistoryArgs{
			Project:      &projectID,
			DefinitionId: &defID,
		}).
		Return(&[]releaseapi.ReleaseDefinitionRevision{}, nil).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, DataReleaseDefinitionHistory().Schema, map[string]interface{}{
		"project_id":            projectID,
		"release_definition_id": defID,
	})

	diags := dataReleaseDefinitionHistoryRead(context.Background(), resourceData, clients)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	rawList := resourceData.Get("revisions").([]interface{})
	require.Len(t, rawList, 0)
}

// TestDataReleaseDefinitionHistory_Read_Error verifies that when
// GetReleaseDefinitionHistory returns an error the function returns a
// non-empty Diagnostics.
func TestDataReleaseDefinitionHistory_Read_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	releaseClient := azdosdkmocks.NewMockReleaseClient(ctrl)
	clients := &client.AggregatedClient{
		ReleaseClient: releaseClient,
		Ctx:           context.Background(),
	}

	projectID := testReleaseDefinitionProjectID.String()
	defID := converter.Int(7)

	releaseClient.
		EXPECT().
		GetReleaseDefinitionHistory(clients.Ctx, releaseapi.GetReleaseDefinitionHistoryArgs{
			Project:      &projectID,
			DefinitionId: defID,
		}).
		Return(nil, errors.New("api error")).
		Times(1)

	resourceData := schema.TestResourceDataRaw(t, DataReleaseDefinitionHistory().Schema, map[string]interface{}{
		"project_id":            projectID,
		"release_definition_id": *defID,
	})

	diags := dataReleaseDefinitionHistoryRead(context.Background(), resourceData, clients)
	require.True(t, diags.HasError(), "expected diagnostics to have an error")
}
