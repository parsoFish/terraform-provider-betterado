//go:build (all || resource_release_definition) && !exclude_resource_release_definition
// +build all resource_release_definition
// +build !exclude_resource_release_definition

package release

// test_fixtures_test.go — shared package-level test fixtures used across
// multiple unit test files in the release package.
// Relocated from resource_release_definition_test.go after that file was
// removed when the resource migrated to terraform-plugin-framework.

import (
	"github.com/google/uuid"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

var (
	testReleaseDefinitionProjectID      = uuid.New()
	testReleaseDefinitionID             = 42
	testReleaseDefinitionRevision       = 1
	testReleaseDefinitionEnvID          = 1
	testReleaseDefinitionEnvRank        = 1
	testReleaseDefinitionWorkflowTaskID = uuid.New()
)

var testReleaseDefinition = releaseapi.ReleaseDefinition{
	Id:                &testReleaseDefinitionID,
	Name:              converter.String("MyReleaseDefinition"),
	Path:              converter.String("\\"),
	Description:       converter.String("A test release definition"),
	ReleaseNameFormat: converter.String("Release-$(rev:r)"),
	Revision:          &testReleaseDefinitionRevision,
	Variables: &map[string]releaseapi.ConfigurationVariableValue{
		"MY_VAR": {
			Value:         converter.String("my-value"),
			IsSecret:      converter.Bool(false),
			AllowOverride: converter.Bool(false),
		},
	},
	Environments: &[]releaseapi.ReleaseDefinitionEnvironment{
		{
			Id:   &testReleaseDefinitionEnvID,
			Name: converter.String("Production"),
			Rank: &testReleaseDefinitionEnvRank,
			DeployPhases: &[]interface{}{
				map[string]interface{}{
					"name":      "Agent phase",
					"rank":      float64(1),
					"phaseType": "agentBasedDeployment",
					"workflowTasks": []interface{}{
						map[string]interface{}{
							"name":            "Run Script",
							"taskId":          testReleaseDefinitionWorkflowTaskID.String(),
							"version":         "1.*",
							"enabled":         true,
							"alwaysRun":       false,
							"continueOnError": false,
							"condition":       "succeeded()",
							"definitionType":  "task",
						},
					},
				},
			},
		},
	},
	Artifacts: &[]releaseapi.Artifact{
		{
			Alias:     converter.String("_myBuild"),
			Type:      converter.String("Build"),
			IsPrimary: converter.Bool(true),
			DefinitionReference: &map[string]releaseapi.ArtifactSourceReference{
				"definition": {Id: converter.String("1")},
				"project":    {Id: converter.String(testReleaseDefinitionProjectID.String())},
			},
		},
	},
}
