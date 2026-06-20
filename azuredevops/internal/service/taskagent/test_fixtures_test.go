//go:build (all || data_task_group) && !exclude_data_task_group
// +build all data_task_group
// +build !exclude_data_task_group

package taskagent

// test_fixtures_test.go — shared package-level test fixtures used across
// multiple unit test files in the taskagent package.
// Relocated from resource_task_group_test.go after that file was removed
// when the resource migrated to terraform-plugin-framework.

import (
	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

var testTaskGroupProjectID = uuid.New()
var testTaskGroupID = uuid.New()
var testTaskStepTaskID = uuid.New()

var testTaskGroupRevision = 1
var testTaskVersionMajor = 1
var testTaskVersionMinor = 0
var testTaskVersionPatch = 0
var testTaskVersionIsTest = false

var testTaskGroup = taskagent.TaskGroup{
	Id:           &testTaskGroupID,
	Name:         converter.String("MyTaskGroup"),
	FriendlyName: converter.String("My Task Group"),
	Description:  converter.String("A test task group"),
	Category:     converter.String("Build"),
	Author:       converter.String("tester"),
	Revision:     &testTaskGroupRevision,
	Version: &taskagent.TaskVersion{
		Major:  &testTaskVersionMajor,
		Minor:  &testTaskVersionMinor,
		Patch:  &testTaskVersionPatch,
		IsTest: &testTaskVersionIsTest,
	},
	Tasks: &[]taskagent.TaskGroupStep{
		{
			DisplayName:     converter.String("Run Script"),
			Enabled:         converter.Bool(true),
			AlwaysRun:       converter.Bool(false),
			ContinueOnError: converter.Bool(false),
			Condition:       converter.String("succeeded()"),
			Task: &taskagent.TaskDefinitionReference{
				Id:             &testTaskStepTaskID,
				VersionSpec:    converter.String("1.*"),
				DefinitionType: converter.String("task"),
			},
		},
	},
}
