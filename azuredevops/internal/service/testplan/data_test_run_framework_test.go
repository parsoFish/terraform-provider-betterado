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

// ── betterado_test_run unit tests ─────────────────────────────────────────────

// TestUnitTestRun_Read verifies that GetTestRunById returns a populated TestRun
// and that the state attributes are set correctly (name, state, total_tests,
// passed_tests, failed_tests).
func TestUnitTestRun_Read(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testClient := azdosdkmocks.NewMockTestClient(ctrl)
	clients := &client.AggregatedClient{
		TestClient: testClient,
		Ctx:        context.Background(),
	}

	projectID := "my-project"
	runID := 42
	runName := "Nightly Run"
	runState := "Completed"
	totalTests := 100
	passedTests := 90
	failedTests := 5

	testClient.
		EXPECT().
		GetTestRunById(clients.Ctx, test.GetTestRunByIdArgs{
			Project: &projectID,
			RunId:   &runID,
		}).
		Return(&test.TestRun{
			Id:              &runID,
			Name:            &runName,
			State:           &runState,
			TotalTests:      &totalTests,
			PassedTests:     &passedTests,
			UnanalyzedTests: &failedTests,
		}, nil).
		Times(1)

	got, err := clients.TestClient.GetTestRunById(clients.Ctx, test.GetTestRunByIdArgs{
		Project: &projectID,
		RunId:   &runID,
	})
	require.NoError(t, err)
	require.NotNil(t, got)

	model := &testRunDataModel{
		ProjectID: types.StringValue(projectID),
		RunID:     types.StringValue("42"),
	}
	flattenTestRunData(model, got)

	assert.Equal(t, "42", model.ID.ValueString())
	assert.Equal(t, "Nightly Run", model.Name.ValueString())
	assert.Equal(t, "Completed", model.State.ValueString())
	assert.Equal(t, int64(100), model.TotalTests.ValueInt64())
	assert.Equal(t, int64(90), model.PassedTests.ValueInt64())
	assert.Equal(t, int64(5), model.FailedTests.ValueInt64())
}

// TestUnitTestRun_ReadNotFound verifies that a 404 from GetTestRunById is
// identified as "not found" — data sources error on missing (unlike resources).
func TestUnitTestRun_ReadNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testClient := azdosdkmocks.NewMockTestClient(ctrl)
	clients := &client.AggregatedClient{
		TestClient: testClient,
		Ctx:        context.Background(),
	}

	projectID := "my-project"
	runID := 99

	notFoundErr := azuredevops.WrappedError{
		StatusCode: converter.Int(404),
		Message:    converter.String("run not found"),
	}

	testClient.
		EXPECT().
		GetTestRunById(clients.Ctx, test.GetTestRunByIdArgs{
			Project: &projectID,
			RunId:   &runID,
		}).
		Return(nil, notFoundErr).
		Times(1)

	got, err := clients.TestClient.GetTestRunById(clients.Ctx, test.GetTestRunByIdArgs{
		Project: &projectID,
		RunId:   &runID,
	})
	assert.Nil(t, got)
	assert.True(t, isNotFound(err), "404 error must be identified as not found")
}

// TestUnitTestRun_Schema verifies that NewTestRunDataSource() exposes the
// correct type name and required schema attributes.
func TestUnitTestRun_Schema(t *testing.T) {
	ctx := context.Background()
	ds := NewTestRunDataSource()
	require.NotNil(t, ds)

	metaResp := &datasource.MetadataResponse{}
	ds.Metadata(ctx, datasource.MetadataRequest{}, metaResp)
	assert.Equal(t, "betterado_test_run", metaResp.TypeName)

	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(),
		"Schema() must not emit diagnostics: %v", schemaResp.Diagnostics)

	for _, attr := range []string{
		"id", "project_id", "run_id", "name", "state",
		"total_tests", "passed_tests", "failed_tests",
		"started_date", "completed_date",
	} {
		_, ok := schemaResp.Schema.Attributes[attr]
		assert.True(t, ok, "attribute %q must exist in datasource schema", attr)
	}
}
