//go:build all || resource_test_run

package testplan

import (
	"context"
	"testing"

	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/test"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ── betterado_test_result unit tests ──────────────────────────────────────────

// TestUnitTestRun_Result_Read verifies that GetTestResultById returns a populated
// TestCaseResult and that the state attributes are set correctly
// (test_case_title, outcome, duration_in_ms).
func TestUnitTestRun_Result_Read(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testClient := azdosdkmocks.NewMockTestClient(ctrl)
	clients := &client.AggregatedClient{
		TestClient: testClient,
		Ctx:        context.Background(),
	}

	projectID := "my-project"
	runID := 10
	resultID := 7
	title := "TestLogin"
	outcome := "Passed"
	duration := 123.45
	failureType := ""
	errorMessage := ""

	testClient.
		EXPECT().
		GetTestResultById(clients.Ctx, test.GetTestResultByIdArgs{
			Project:          &projectID,
			RunId:            &runID,
			TestCaseResultId: &resultID,
		}).
		Return(&test.TestCaseResult{
			Id:            &resultID,
			TestCaseTitle: &title,
			Outcome:       &outcome,
			DurationInMs:  &duration,
			FailureType:   &failureType,
			ErrorMessage:  &errorMessage,
		}, nil).
		Times(1)

	got, err := clients.TestClient.GetTestResultById(clients.Ctx, test.GetTestResultByIdArgs{
		Project:          &projectID,
		RunId:            &runID,
		TestCaseResultId: &resultID,
	})
	require.NoError(t, err)
	require.NotNil(t, got)

	model := &testResultDataModel{
		ProjectID: types.StringValue(projectID),
		RunID:     types.StringValue("10"),
		ResultID:  types.StringValue("7"),
	}
	flattenTestResultData(model, got)

	assert.Equal(t, "7", model.ID.ValueString())
	assert.Equal(t, "TestLogin", model.TestCaseTitle.ValueString())
	assert.Equal(t, "Passed", model.Outcome.ValueString())
	assert.Equal(t, int64(123), model.DurationInMs.ValueInt64())
	assert.Equal(t, "", model.FailureType.ValueString())
	assert.Equal(t, "", model.ErrorMessage.ValueString())
}

// TestUnitTestRun_Result_ReadNotFound verifies that a 404 from GetTestResultById
// is identified as "not found" — data sources error on missing.
func TestUnitTestRun_Result_ReadNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testClient := azdosdkmocks.NewMockTestClient(ctrl)
	clients := &client.AggregatedClient{
		TestClient: testClient,
		Ctx:        context.Background(),
	}

	projectID := "my-project"
	runID := 10
	resultID := 999

	notFoundErr := azuredevops.WrappedError{
		StatusCode: converter.Int(404),
		Message:    converter.String("result not found"),
	}

	testClient.
		EXPECT().
		GetTestResultById(clients.Ctx, test.GetTestResultByIdArgs{
			Project:          &projectID,
			RunId:            &runID,
			TestCaseResultId: &resultID,
		}).
		Return(nil, notFoundErr).
		Times(1)

	got, err := clients.TestClient.GetTestResultById(clients.Ctx, test.GetTestResultByIdArgs{
		Project:          &projectID,
		RunId:            &runID,
		TestCaseResultId: &resultID,
	})
	assert.Nil(t, got)
	assert.True(t, isNotFound(err), "404 error must be identified as not found")
}

// TestUnitTestRun_Result_Schema verifies that NewTestResultDataSource() exposes
// the correct type name and required schema attributes.
func TestUnitTestRun_Result_Schema(t *testing.T) {
	ctx := context.Background()
	ds := NewTestResultDataSource()
	require.NotNil(t, ds)

	metaResp := &datasource.MetadataResponse{}
	ds.Metadata(ctx, datasource.MetadataRequest{}, metaResp)
	assert.Equal(t, "betterado_test_result", metaResp.TypeName)

	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not emit diagnostics: %v", schemaResp.Diagnostics)

	for _, attr := range []string{
		"id", "project_id", "run_id", "result_id",
		"test_case_title", "outcome", "duration_in_ms",
		"failure_type", "error_message",
	} {
		_, ok := schemaResp.Schema.Attributes[attr]
		assert.True(t, ok, "attribute %q must exist in datasource schema", attr)
	}
}
