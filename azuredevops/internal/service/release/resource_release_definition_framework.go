package release

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	distributedtaskcommon "github.com/microsoft/azure-devops-go-api/azuredevops/v7/distributedtaskcommon"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Ensure the resource satisfies the framework interface.
var (
	_ resource.Resource                 = &releaseDefinitionFrameworkResource{}
	_ resource.ResourceWithImportState  = &releaseDefinitionFrameworkResource{}
	_ resource.ResourceWithUpgradeState = &releaseDefinitionFrameworkResource{}
	_ resource.ResourceWithModifyPlan   = &releaseDefinitionFrameworkResource{}
)

// releaseDefinitionFrameworkResource is the terraform-plugin-framework implementation of
// betterado_release_definition. CRUD is wired to the same release API client as the
// SDKv2 version; provider-level client injection is done in WI-5 (framework_provider.go).
type releaseDefinitionFrameworkResource struct {
	client *client.AggregatedClient
}

// NewReleaseDefinitionResource returns a new framework resource for betterado_release_definition.
func NewReleaseDefinitionResource() resource.Resource {
	return &releaseDefinitionFrameworkResource{}
}

// -------------------------------------------------------------------------
// Model types
// -------------------------------------------------------------------------

// releaseDefinitionModel is the top-level Terraform state model.
type releaseDefinitionModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	ProjectID         types.String `tfsdk:"project_id"`
	Path              types.String `tfsdk:"path"`
	Description       types.String `tfsdk:"description"`
	ReleaseNameFormat types.String `tfsdk:"release_name_format"`
	Tags              types.Set    `tfsdk:"tags"`
	Revision          types.Int64  `tfsdk:"revision"`
	Stages            types.List   `tfsdk:"stages"`
	Variables         types.Map    `tfsdk:"variables"`
	VariableGroups    types.List   `tfsdk:"variable_groups"`
	Artifact          types.List   `tfsdk:"artifact"`
	Triggers          types.List   `tfsdk:"triggers"`
}

// stageModel maps to a single ReleaseDefinitionEnvironment.
type stageModel struct {
	Name                types.String `tfsdk:"name"`
	Rank                types.Int64  `tfsdk:"rank"`
	Condition           types.List   `tfsdk:"condition"`
	EnvironmentOptions  types.List   `tfsdk:"environment_options"`
	RetentionPolicy     types.List   `tfsdk:"retention_policy"`
	PreDeployApproval   types.List   `tfsdk:"pre_deploy_approval"`
	PostDeployApproval  types.List   `tfsdk:"post_deploy_approval"`
	PreDeploymentGates  types.List   `tfsdk:"pre_deployment_gates"`
	PostDeploymentGates types.List   `tfsdk:"post_deployment_gates"`
	DeployPhase         types.List   `tfsdk:"deploy_phase"`
	Variables           types.Map    `tfsdk:"variables"`
	VariableGroups      types.List   `tfsdk:"variable_groups"`
	ExecutionPolicy     types.List   `tfsdk:"execution_policy"`
	EnvironmentTrigger  types.List   `tfsdk:"environment_trigger"`
	Schedule            types.List   `tfsdk:"schedule"`
	ProcessParameters   types.List   `tfsdk:"process_parameters"`
	Properties          types.Map    `tfsdk:"properties"`
	Owner               types.String `tfsdk:"owner"`
	StageID             types.Int64  `tfsdk:"id"`
}

// variableValueModel is the nested-attribute value object for a single variable entry.
type variableValueModel struct {
	Value         types.String `tfsdk:"value"`
	IsSecret      types.Bool   `tfsdk:"is_secret"`
	AllowOverride types.Bool   `tfsdk:"allow_override"`
}

// artifactModel maps to a single ADO artifact source.
type artifactModel struct {
	Alias               types.String `tfsdk:"alias"`
	Type                types.String `tfsdk:"type"`
	IsPrimary           types.Bool   `tfsdk:"is_primary"`
	DefinitionReference types.Map    `tfsdk:"definition_reference"`
}

// triggerModel is the list-element for the heterogeneous triggers list.
type triggerModel struct {
	CdArtifactTrigger     types.List `tfsdk:"cd_artifact_trigger"`
	ScheduleTrigger       types.List `tfsdk:"schedule_trigger"`
	SourceRepoTrigger     types.List `tfsdk:"source_repo_trigger"`
	ContainerImageTrigger types.List `tfsdk:"container_image_trigger"`
	PullRequestTrigger    types.List `tfsdk:"pull_request_trigger"`
}

// cdArtifactTriggerModel maps to the ADO "artifactSource" trigger.
type cdArtifactTriggerModel struct {
	ArtifactAlias               types.String `tfsdk:"artifact_alias"`
	BranchFilters               types.List   `tfsdk:"branch_filters"`
	TagFilter                   types.List   `tfsdk:"tag_filter"`
	UseBuildDefinitionBranch    types.Bool   `tfsdk:"use_build_definition_branch"`
	CreateReleaseOnBuildTagging types.Bool   `tfsdk:"create_release_on_build_tagging"`
}

// tagFilterModel maps to the tag_filter nested block inside cd_artifact_trigger.
type tagFilterModel struct {
	Tags types.List `tfsdk:"tags"`
}

// sourceRepoTriggerModel maps to the ADO "sourceRepo" trigger.
type sourceRepoTriggerModel struct {
	Alias         types.String `tfsdk:"alias"`
	BranchFilters types.List   `tfsdk:"branch_filters"`
}

// containerImageTriggerModel maps to the ADO "containerImage" trigger.
type containerImageTriggerModel struct {
	ArtifactAlias types.String `tfsdk:"artifact_alias"`
	Label         types.String `tfsdk:"label"`
}

// pullRequestTriggerModel maps to the ADO "pullRequest" trigger.
type pullRequestTriggerModel struct {
	ArtifactAlias        types.String `tfsdk:"artifact_alias"`
	TargetBranches       types.List   `tfsdk:"target_branches"`
	Tags                 types.List   `tfsdk:"tags"`
	UseArtifactReference types.Bool   `tfsdk:"use_artifact_reference"`
}

// scheduleTriggerModel maps to the ADO "schedule" trigger.
type scheduleTriggerModel struct {
	ScheduleOnlyWithChanges types.Bool   `tfsdk:"schedule_only_with_changes"`
	StartHours              types.Int64  `tfsdk:"start_hours"`
	StartMinutes            types.Int64  `tfsdk:"start_minutes"`
	TimeZoneID              types.String `tfsdk:"time_zone_id"`
	DaysToRelease           types.Int64  `tfsdk:"days_to_release"`
}

// deployPhaseModel maps to a single deploy phase (stored as generic map by ADO API).
type deployPhaseModel struct {
	Name            types.String `tfsdk:"name"`
	Rank            types.Int64  `tfsdk:"rank"`
	PhaseType       types.String `tfsdk:"phase_type"`
	DeploymentInput types.List   `tfsdk:"deployment_input"`
	WorkflowTask    types.List   `tfsdk:"workflow_task"`
}

// deploymentInputModel maps to the deploymentInput object inside a deploy phase.
type deploymentInputModel struct {
	QueueID                   types.Int64  `tfsdk:"queue_id"`
	Condition                 types.String `tfsdk:"condition"`
	TimeoutInMinutes          types.Int64  `tfsdk:"timeout_in_minutes"`
	JobCancelTimeoutInMinutes types.Int64  `tfsdk:"job_cancel_timeout_in_minutes"`
	SkipArtifactsDownload     types.Bool   `tfsdk:"skip_artifacts_download"`
	EnableAccessToken         types.Bool   `tfsdk:"enable_access_token"`
	AgentSpecification        types.String `tfsdk:"agent_specification"`
	OverrideInputs            types.Map    `tfsdk:"override_inputs"`
	Demands                   types.List   `tfsdk:"demands"`
	ParallelExecution         types.List   `tfsdk:"parallel_execution"`
}

// parallelExecutionModel maps to the parallelExecution object inside deployment_input.
type parallelExecutionModel struct {
	Type              types.String `tfsdk:"type"`
	MaxNumberOfAgents types.Int64  `tfsdk:"max_number_of_agents"`
	Multipliers       types.List   `tfsdk:"multipliers"`
}

// workflowTaskModel maps to a single workflow task inside a deploy phase.
type workflowTaskModel struct {
	TaskID                  types.String `tfsdk:"task_id"`
	DisplayName             types.String `tfsdk:"display_name"`
	VersionSpec             types.String `tfsdk:"version_spec"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
	AlwaysRun               types.Bool   `tfsdk:"always_run"`
	Condition               types.String `tfsdk:"condition"`
	TimeoutInMinutes        types.Int64  `tfsdk:"timeout_in_minutes"`
	RetryCountOnTaskFailure types.Int64  `tfsdk:"retry_count_on_task_failure"`
	Inputs                  types.Map    `tfsdk:"inputs"`
	DefinitionType          types.String `tfsdk:"definition_type"`
}

// conditionModel maps to releaseapi.Condition.
type conditionModel struct {
	ConditionType types.String `tfsdk:"condition_type"`
	Name          types.String `tfsdk:"name"`
	Value         types.String `tfsdk:"value"`
}

// environmentOptionsModel maps to releaseapi.EnvironmentOptions.
type environmentOptionsModel struct {
	EmailNotificationType                          types.String `tfsdk:"email_notification_type"`
	EmailRecipients                                types.String `tfsdk:"email_recipients"`
	SkipArtifactsDownload                          types.Bool   `tfsdk:"skip_artifacts_download"`
	TimeoutInMinutes                               types.Int64  `tfsdk:"timeout_in_minutes"`
	EnableAccessToken                              types.Bool   `tfsdk:"enable_access_token"`
	PublishDeploymentStatus                        types.Bool   `tfsdk:"publish_deployment_status"`
	BadgeEnabled                                   types.Bool   `tfsdk:"badge_enabled"`
	AutoLinkWorkItems                              types.Bool   `tfsdk:"auto_link_workitems"`
	PullRequestDeploymentEnabled                   types.Bool   `tfsdk:"pull_request_deployment_enabled"`
	PublishDeploymentStatusToDevopsProjectSettings types.Bool   `tfsdk:"publish_deployment_status_to_devops_project_settings"`
}

// retentionPolicyModel maps to releaseapi.EnvironmentRetentionPolicy.
type retentionPolicyModel struct {
	DaysToKeep     types.Int64 `tfsdk:"days_to_keep"`
	ReleasesToKeep types.Int64 `tfsdk:"releases_to_keep"`
	RetainBuild    types.Bool  `tfsdk:"retain_build"`
}

// executionPolicyModel maps releaseapi.EnvironmentExecutionPolicy.
type executionPolicyModel struct {
	ConcurrencyCount types.Int64 `tfsdk:"concurrency_count"`
	QueueDepthCount  types.Int64 `tfsdk:"queue_depth_count"`
}

// environmentTriggerModel maps releaseapi.EnvironmentTrigger.
type environmentTriggerModel struct {
	DefinitionEnvironmentID types.Int64  `tfsdk:"definition_environment_id"`
	TriggerType             types.String `tfsdk:"trigger_type"`
	TriggerContent          types.String `tfsdk:"trigger_content"`
}

// stageScheduleModel maps releaseapi.ReleaseSchedule at stage level.
type stageScheduleModel struct {
	DaysToRelease types.Int64  `tfsdk:"days_to_release"`
	StartHours    types.Int64  `tfsdk:"start_hours"`
	StartMinutes  types.Int64  `tfsdk:"start_minutes"`
	TimeZoneID    types.String `tfsdk:"time_zone_id"`
	JobID         types.String `tfsdk:"job_id"`
}

// processParametersModel maps distributedtaskcommon.ProcessParameters.
type processParametersModel struct {
	Input types.List `tfsdk:"input"`
}

// processParameterInputModel maps distributedtaskcommon.TaskInputDefinitionBase.
type processParameterInputModel struct {
	Name          types.String `tfsdk:"name"`
	DefaultValue  types.String `tfsdk:"default_value"`
	ParameterType types.String `tfsdk:"parameter_type"`
}

// approvalModel maps to releaseapi.ReleaseDefinitionApprovals.
type approvalModel struct {
	Approver        types.List `tfsdk:"approver"`
	ApprovalOptions types.List `tfsdk:"approval_options"`
}

// approverModel maps to releaseapi.ReleaseDefinitionApprovalStep.
type approverModel struct {
	ID          types.String `tfsdk:"id"`
	IsAutomated types.Bool   `tfsdk:"is_automated"`
	Rank        types.Int64  `tfsdk:"rank"`
}

// approvalOptionsModel maps to releaseapi.ApprovalOptions.
type approvalOptionsModel struct {
	RequiredApproverCount                                   types.Int64  `tfsdk:"required_approver_count"`
	ReleaseCreatorCanBeApprover                             types.Bool   `tfsdk:"release_creator_can_be_approver"`
	EnforceIdentityRevalidation                             types.Bool   `tfsdk:"enforce_identity_revalidation"`
	TimeoutInMinutes                                        types.Int64  `tfsdk:"timeout_in_minutes"`
	ExecutionOrder                                          types.String `tfsdk:"execution_order"`
	AutoTriggeredAndPreviousEnvironmentApprovedCanBeSkipped types.Bool   `tfsdk:"auto_triggered_and_previous_environment_approved_can_be_skipped"`
}

// deploymentGatesModel maps to releaseapi.ReleaseDefinitionGatesStep.
type deploymentGatesModel struct {
	GatesOptions types.List `tfsdk:"gates_options"`
	Gate         types.List `tfsdk:"gate"`
}

// gatesOptionsModel maps to releaseapi.ReleaseDefinitionGatesOptions.
type gatesOptionsModel struct {
	IsEnabled              types.Bool  `tfsdk:"is_enabled"`
	Timeout                types.Int64 `tfsdk:"timeout"`
	SamplingInterval       types.Int64 `tfsdk:"sampling_interval"`
	StabilizationTime      types.Int64 `tfsdk:"stabilization_time"`
	MinimumSuccessDuration types.Int64 `tfsdk:"minimum_success_duration"`
}

// gateTaskModel maps to a single gate task (releaseapi.WorkflowTask inside a gate).
// Mirrors workflowTaskModel but uses version_spec + display_name to match the
// workflow_task block, plus definition_type from the API.
type gateTaskModel struct {
	TaskID                  types.String `tfsdk:"task_id"`
	DisplayName             types.String `tfsdk:"display_name"`
	VersionSpec             types.String `tfsdk:"version_spec"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
	AlwaysRun               types.Bool   `tfsdk:"always_run"`
	Condition               types.String `tfsdk:"condition"`
	TimeoutInMinutes        types.Int64  `tfsdk:"timeout_in_minutes"`
	RetryCountOnTaskFailure types.Int64  `tfsdk:"retry_count_on_task_failure"`
	Inputs                  types.Map    `tfsdk:"inputs"`
	DefinitionType          types.String `tfsdk:"definition_type"`
}

// gateModel maps to releaseapi.ReleaseDefinitionGate.
type gateModel struct {
	Task types.List `tfsdk:"task"`
}

// -------------------------------------------------------------------------
// attr.Type helpers
// -------------------------------------------------------------------------

var conditionAttrTypes = map[string]attr.Type{
	"condition_type": types.StringType,
	"name":           types.StringType,
	"value":          types.StringType,
}

var environmentOptionsAttrTypes = map[string]attr.Type{
	"email_notification_type":                              types.StringType,
	"email_recipients":                                     types.StringType,
	"skip_artifacts_download":                              types.BoolType,
	"timeout_in_minutes":                                   types.Int64Type,
	"enable_access_token":                                  types.BoolType,
	"publish_deployment_status":                            types.BoolType,
	"badge_enabled":                                        types.BoolType,
	"auto_link_workitems":                                  types.BoolType,
	"pull_request_deployment_enabled":                      types.BoolType,
	"publish_deployment_status_to_devops_project_settings": types.BoolType,
}

var retentionPolicyAttrTypes = map[string]attr.Type{
	"days_to_keep":     types.Int64Type,
	"releases_to_keep": types.Int64Type,
	"retain_build":     types.BoolType,
}

var approverAttrTypes = map[string]attr.Type{
	"id":           types.StringType,
	"is_automated": types.BoolType,
	"rank":         types.Int64Type,
}

var approvalOptionsAttrTypes = map[string]attr.Type{
	"required_approver_count":         types.Int64Type,
	"release_creator_can_be_approver": types.BoolType,
	"enforce_identity_revalidation":   types.BoolType,
	"timeout_in_minutes":              types.Int64Type,
	"execution_order":                 types.StringType,
	"auto_triggered_and_previous_environment_approved_can_be_skipped": types.BoolType,
}

var approvalAttrTypes = map[string]attr.Type{
	"approver":         types.ListType{ElemType: types.ObjectType{AttrTypes: approverAttrTypes}},
	"approval_options": types.ListType{ElemType: types.ObjectType{AttrTypes: approvalOptionsAttrTypes}},
}

var gatesOptionsAttrTypes = map[string]attr.Type{
	"is_enabled":               types.BoolType,
	"timeout":                  types.Int64Type,
	"sampling_interval":        types.Int64Type,
	"stabilization_time":       types.Int64Type,
	"minimum_success_duration": types.Int64Type,
}

// Deploy-phase attr type maps.
var parallelExecutionAttrTypes = map[string]attr.Type{
	"type":                 types.StringType,
	"max_number_of_agents": types.Int64Type,
	"multipliers":          types.ListType{ElemType: types.StringType},
}

var deploymentInputAttrTypes = map[string]attr.Type{
	"queue_id":                      types.Int64Type,
	"condition":                     types.StringType,
	"timeout_in_minutes":            types.Int64Type,
	"job_cancel_timeout_in_minutes": types.Int64Type,
	"skip_artifacts_download":       types.BoolType,
	"enable_access_token":           types.BoolType,
	"agent_specification":           types.StringType,
	"override_inputs":               types.MapType{ElemType: types.StringType},
	"demands":                       types.ListType{ElemType: types.StringType},
	"parallel_execution":            types.ListType{ElemType: types.ObjectType{AttrTypes: parallelExecutionAttrTypes}},
}

var workflowTaskAttrTypes = map[string]attr.Type{
	"task_id":                     types.StringType,
	"display_name":                types.StringType,
	"version_spec":                types.StringType,
	"enabled":                     types.BoolType,
	"always_run":                  types.BoolType,
	"condition":                   types.StringType,
	"timeout_in_minutes":          types.Int64Type,
	"retry_count_on_task_failure": types.Int64Type,
	"inputs":                      types.MapType{ElemType: types.StringType},
	"definition_type":             types.StringType,
}

// gateTaskAttrTypes mirrors workflowTaskAttrTypes — gate tasks share the same API shape.
var gateTaskAttrTypes = map[string]attr.Type{
	"task_id":                     types.StringType,
	"display_name":                types.StringType,
	"version_spec":                types.StringType,
	"enabled":                     types.BoolType,
	"always_run":                  types.BoolType,
	"condition":                   types.StringType,
	"timeout_in_minutes":          types.Int64Type,
	"retry_count_on_task_failure": types.Int64Type,
	"inputs":                      types.MapType{ElemType: types.StringType},
	"definition_type":             types.StringType,
}

var gateAttrTypes = map[string]attr.Type{
	"task": types.ListType{ElemType: types.ObjectType{AttrTypes: gateTaskAttrTypes}},
}

var deploymentGatesAttrTypes = map[string]attr.Type{
	"gates_options": types.ListType{ElemType: types.ObjectType{AttrTypes: gatesOptionsAttrTypes}},
	"gate":          types.ListType{ElemType: types.ObjectType{AttrTypes: gateAttrTypes}},
}

var deployPhaseAttrTypes = map[string]attr.Type{
	"name":             types.StringType,
	"rank":             types.Int64Type,
	"phase_type":       types.StringType,
	"deployment_input": types.ListType{ElemType: types.ObjectType{AttrTypes: deploymentInputAttrTypes}},
	"workflow_task":    types.ListType{ElemType: types.ObjectType{AttrTypes: workflowTaskAttrTypes}},
}

// variableValueAttrTypes is the attr.Type map for the nested object inside a variables MapNestedAttribute.
var variableValueAttrTypes = map[string]attr.Type{
	"value":          types.StringType,
	"is_secret":      types.BoolType,
	"allow_override": types.BoolType,
}

var executionPolicyAttrTypes = map[string]attr.Type{
	"concurrency_count": types.Int64Type,
	"queue_depth_count": types.Int64Type,
}

var environmentTriggerAttrTypes = map[string]attr.Type{
	"definition_environment_id": types.Int64Type,
	"trigger_type":              types.StringType,
	"trigger_content":           types.StringType,
}

var stageScheduleAttrTypes = map[string]attr.Type{
	"days_to_release": types.Int64Type,
	"start_hours":     types.Int64Type,
	"start_minutes":   types.Int64Type,
	"time_zone_id":    types.StringType,
	"job_id":          types.StringType,
}

var processParameterInputAttrTypes = map[string]attr.Type{
	"name":           types.StringType,
	"default_value":  types.StringType,
	"parameter_type": types.StringType,
}

var processParametersAttrTypes = map[string]attr.Type{
	"input": types.ListType{ElemType: types.ObjectType{AttrTypes: processParameterInputAttrTypes}},
}

// artifactAttrTypes is the attr.Type map for a single artifact list element.
var artifactAttrTypes = map[string]attr.Type{
	"alias":                types.StringType,
	"type":                 types.StringType,
	"is_primary":           types.BoolType,
	"definition_reference": types.MapType{ElemType: types.StringType},
}

// tagFilterAttrTypes is the attr.Type map for a tag_filter element.
var tagFilterAttrTypes = map[string]attr.Type{
	"tags": types.ListType{ElemType: types.StringType},
}

// cdArtifactTriggerAttrTypes is the attr.Type map for a cd_artifact_trigger element.
var cdArtifactTriggerAttrTypes = map[string]attr.Type{
	"artifact_alias":                  types.StringType,
	"branch_filters":                  types.ListType{ElemType: types.StringType},
	"tag_filter":                      types.ListType{ElemType: types.ObjectType{AttrTypes: tagFilterAttrTypes}},
	"use_build_definition_branch":     types.BoolType,
	"create_release_on_build_tagging": types.BoolType,
}

// scheduleTriggerAttrTypes is the attr.Type map for a schedule_trigger element.
var scheduleTriggerAttrTypes = map[string]attr.Type{
	"schedule_only_with_changes": types.BoolType,
	"start_hours":                types.Int64Type,
	"start_minutes":              types.Int64Type,
	"time_zone_id":               types.StringType,
	"days_to_release":            types.Int64Type,
}

// sourceRepoTriggerAttrTypes is the attr.Type map for a source_repo_trigger element.
var sourceRepoTriggerAttrTypes = map[string]attr.Type{
	"alias":          types.StringType,
	"branch_filters": types.ListType{ElemType: types.StringType},
}

// containerImageTriggerAttrTypes is the attr.Type map for a container_image_trigger element.
var containerImageTriggerAttrTypes = map[string]attr.Type{
	"artifact_alias": types.StringType,
	"label":          types.StringType,
}

// pullRequestTriggerAttrTypes is the attr.Type map for a pull_request_trigger element.
var pullRequestTriggerAttrTypes = map[string]attr.Type{
	"artifact_alias":         types.StringType,
	"target_branches":        types.ListType{ElemType: types.StringType},
	"tags":                   types.ListType{ElemType: types.StringType},
	"use_artifact_reference": types.BoolType,
}

// triggerAttrTypes is the attr.Type map for a single triggers list element.
var triggerAttrTypes = map[string]attr.Type{
	"cd_artifact_trigger":     types.ListType{ElemType: types.ObjectType{AttrTypes: cdArtifactTriggerAttrTypes}},
	"schedule_trigger":        types.ListType{ElemType: types.ObjectType{AttrTypes: scheduleTriggerAttrTypes}},
	"source_repo_trigger":     types.ListType{ElemType: types.ObjectType{AttrTypes: sourceRepoTriggerAttrTypes}},
	"container_image_trigger": types.ListType{ElemType: types.ObjectType{AttrTypes: containerImageTriggerAttrTypes}},
	"pull_request_trigger":    types.ListType{ElemType: types.ObjectType{AttrTypes: pullRequestTriggerAttrTypes}},
}

var stageAttrTypes = map[string]attr.Type{
	"name":                  types.StringType,
	"rank":                  types.Int64Type,
	"condition":             types.ListType{ElemType: types.ObjectType{AttrTypes: conditionAttrTypes}},
	"environment_options":   types.ListType{ElemType: types.ObjectType{AttrTypes: environmentOptionsAttrTypes}},
	"retention_policy":      types.ListType{ElemType: types.ObjectType{AttrTypes: retentionPolicyAttrTypes}},
	"pre_deploy_approval":   types.ListType{ElemType: types.ObjectType{AttrTypes: approvalAttrTypes}},
	"post_deploy_approval":  types.ListType{ElemType: types.ObjectType{AttrTypes: approvalAttrTypes}},
	"pre_deployment_gates":  types.ListType{ElemType: types.ObjectType{AttrTypes: deploymentGatesAttrTypes}},
	"post_deployment_gates": types.ListType{ElemType: types.ObjectType{AttrTypes: deploymentGatesAttrTypes}},
	"deploy_phase":          types.ListType{ElemType: types.ObjectType{AttrTypes: deployPhaseAttrTypes}},
	"variables":             types.MapType{ElemType: types.ObjectType{AttrTypes: variableValueAttrTypes}},
	"variable_groups":       types.ListType{ElemType: types.Int64Type},
	"execution_policy":      types.ListType{ElemType: types.ObjectType{AttrTypes: executionPolicyAttrTypes}},
	"environment_trigger":   types.ListType{ElemType: types.ObjectType{AttrTypes: environmentTriggerAttrTypes}},
	"schedule":              types.ListType{ElemType: types.ObjectType{AttrTypes: stageScheduleAttrTypes}},
	"process_parameters":    types.ListType{ElemType: types.ObjectType{AttrTypes: processParametersAttrTypes}},
	"properties":            types.MapType{ElemType: types.StringType},
	"owner":                 types.StringType,
	"id":                    types.Int64Type,
}

// -------------------------------------------------------------------------
// Metadata / Schema
// -------------------------------------------------------------------------

func (r *releaseDefinitionFrameworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_release_definition"
}

// ModifyPlan pins the computed `revision` to its prior-state value on a no-op
// plan, while leaving it unknown ("known after apply") on a genuine change.
//
// Why this is needed: `revision` is Computed with no plan modifier. The
// framework's plan gate (MarkComputedNilsAsUnknown) flips every config-null
// Computed attribute to unknown whenever the proposed new state differs from
// prior state — which happens on EVERY plan, because Terraform core nulls the
// config-omitted Optional+Computed nested blocks under the Required `stages`
// list. Those nested blocks are restored by their useStateForUnknown* plan
// modifiers, but `revision` had none, so it surfaced as a perpetual
// `~ revision = N -> (known after apply)` diff (the idempotency failure).
//
// A plain UseStateForUnknown is wrong for `revision`: ADO increments the
// revision on every update, so pinning it to prior on a real update produces a
// "Provider produced inconsistent result after apply". Instead this runs at
// resource scope — after all attribute plan modifiers have restored the nested
// blocks — and pins `revision` only when the rest of the planned state is
// byte-identical to prior state (a true no-op, where ADO will NOT bump the
// revision). On any real change it leaves `revision` unknown so the
// server-incremented value is accepted at apply.
func (r *releaseDefinitionFrameworkResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Create (no prior state) or destroy (null plan): nothing to pin.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	// The framework's plan gate (MarkComputedNilsAsUnknown) flips every
	// config-null Computed attribute to unknown whenever the proposed new state
	// differs from prior state — which happens on EVERY plan, because Terraform
	// core nulls the config-omitted Optional+Computed nested blocks under the
	// Required `stages` list. On a true no-op this leaves a spray of
	// `(known after apply)` diffs (revision, stage ids, ADO-assigned environment
	// ids, …) and, worse, "inconsistent result after apply" when ADO re-assigns
	// those server-managed values on the next write.
	//
	// If the ONLY differences between the proposed plan and prior state are
	// attributes the gate marked unknown, this is a genuine no-op: snap the whole
	// plan back to prior state so the plan is empty and nothing is sent to ADO.
	// On any real change, leave the plan untouched so server-(re)assigned
	// computed values are accepted at apply.
	if planIsNoOp(req.Plan.Raw, req.State.Raw) {
		resp.Plan.Raw = req.State.Raw.Copy()
	}
}

// planIsNoOp reports whether `plan` equals `state` everywhere `plan` carries a
// known value — i.e. every difference is an unknown introduced by the framework's
// computed-nil gate. Such a plan represents no real change. A known value that
// differs from prior state (or a changed collection length) is a real change.
func planIsNoOp(plan, state tftypes.Value) bool {
	if !plan.IsKnown() {
		return true // unknown subtree — would be snapped back to prior state
	}
	if plan.IsNull() || state.IsNull() {
		return plan.Equal(state)
	}
	ty := plan.Type()
	switch {
	case ty.Is(tftypes.Object{}), ty.Is(tftypes.Map{}):
		var pm, sm map[string]tftypes.Value
		if err := plan.As(&pm); err != nil {
			return plan.Equal(state)
		}
		if err := state.As(&sm); err != nil {
			return false
		}
		if len(pm) != len(sm) {
			return false
		}
		for k, pv := range pm {
			sv, ok := sm[k]
			if !ok || !planIsNoOp(pv, sv) {
				return false
			}
		}
		return true
	case ty.Is(tftypes.List{}), ty.Is(tftypes.Set{}), ty.Is(tftypes.Tuple{}):
		var pl, sl []tftypes.Value
		if err := plan.As(&pl); err != nil {
			return plan.Equal(state)
		}
		if err := state.As(&sl); err != nil {
			return false
		}
		if len(pl) != len(sl) {
			return false
		}
		for i := range pl {
			if !planIsNoOp(pl[i], sl[i]) {
				return false
			}
		}
		return true
	default:
		return plan.Equal(state)
	}
}

// mergePlanComputed returns a value that equals `plan` for every KNOWN attribute
// and takes `api` only where the plan value is UNKNOWN. It is the core of the
// plan-faithful write path: Create/Update build the result from the plan (so every
// configured attribute — including the Sensitive variable `value`s — is exactly
// what the practitioner wrote) and fill only the server-assigned computed values
// (resource id, revision, stage ids, schedule job_ids, owner, …) that the plan
// left unknown from the concrete API-derived value. This makes the apply result
// consistent with the plan for every configured attribute BY CONSTRUCTION: ADO
// normalising a value it was sent can no longer produce "inconsistent values for
// sensitive attribute" — a genuinely divergent stored value instead surfaces as
// ordinary drift on the next read/plan.
//
// The top-level `stages` list is matched by stage `name` (ADO returns
// environments in a rank-sorted order that need not match the configured order);
// all other lists are matched positionally, which only affects how computed
// (unknown) fills are aligned — never a configured value, since those always come
// from the plan.
func mergePlanComputed(plan, api tftypes.Value) tftypes.Value {
	// Unknown in the plan → take the concrete server value.
	if !plan.IsKnown() {
		return api
	}
	// Explicit null (or any value once we cannot recurse) → the plan wins.
	if plan.IsNull() {
		return plan
	}

	ty := plan.Type()
	switch {
	case ty.Is(tftypes.Object{}):
		var pm map[string]tftypes.Value
		if err := plan.As(&pm); err != nil {
			return plan
		}
		am := map[string]tftypes.Value{}
		if api.IsKnown() && !api.IsNull() {
			if err := api.As(&am); err != nil {
				return api
			}
		}
		out := make(map[string]tftypes.Value, len(pm))
		for k, pv := range pm {
			av, ok := am[k]
			if !ok {
				out[k] = pv
				continue
			}
			if k == "stages" {
				out[k] = mergeStages(pv, av)
			} else {
				out[k] = mergePlanComputed(pv, av)
			}
		}
		return tftypes.NewValue(plan.Type(), out)
	case ty.Is(tftypes.Map{}):
		var pm map[string]tftypes.Value
		if err := plan.As(&pm); err != nil {
			return plan
		}
		am := map[string]tftypes.Value{}
		if api.IsKnown() && !api.IsNull() {
			if err := api.As(&am); err != nil {
				return api
			}
		}
		out := make(map[string]tftypes.Value, len(pm))
		for k, pv := range pm {
			if av, ok := am[k]; ok {
				out[k] = mergePlanComputed(pv, av)
			} else {
				out[k] = pv
			}
		}
		return tftypes.NewValue(plan.Type(), out)
	case ty.Is(tftypes.List{}), ty.Is(tftypes.Set{}), ty.Is(tftypes.Tuple{}):
		var pl []tftypes.Value
		if err := plan.As(&pl); err != nil {
			return plan
		}
		var al []tftypes.Value
		if api.IsKnown() && !api.IsNull() {
			if err := api.As(&al); err != nil {
				return api
			}
		}
		out := make([]tftypes.Value, len(pl))
		for i, pv := range pl {
			if i < len(al) {
				out[i] = mergePlanComputed(pv, al[i])
			} else {
				out[i] = pv
			}
		}
		return tftypes.NewValue(plan.Type(), out)
	default:
		return plan
	}
}

// mergeStages merges the plan and API stage lists element-by-element, matched by
// the stage `name` attribute and emitted in the PLAN's order. A per-name queue
// keeps duplicate names stable. A plan stage with no API counterpart keeps its
// plan value (should not occur after a successful create/update, since ADO
// returns every stage that was sent).
func mergeStages(plan, api tftypes.Value) tftypes.Value {
	if !plan.IsKnown() {
		return api
	}
	if plan.IsNull() {
		return plan
	}
	var pl []tftypes.Value
	if err := plan.As(&pl); err != nil {
		return plan
	}

	queues := map[string][]tftypes.Value{}
	if api.IsKnown() && !api.IsNull() {
		var al []tftypes.Value
		if err := api.As(&al); err != nil {
			return api
		}
		for _, av := range al {
			if n, ok := stageNameFromRaw(av); ok {
				queues[n] = append(queues[n], av)
			}
		}
	}

	out := make([]tftypes.Value, len(pl))
	for i, pv := range pl {
		if n, ok := stageNameFromRaw(pv); ok {
			if q := queues[n]; len(q) > 0 {
				out[i] = mergePlanComputed(pv, q[0])
				queues[n] = q[1:]
				continue
			}
		}
		out[i] = pv
	}
	return tftypes.NewValue(plan.Type(), out)
}

// stripUnknowns recursively replaces any remaining unknown value with a typed
// null. After mergePlanComputed every plan-unknown that had an API counterpart is
// already filled with a concrete value; the only unknowns that can survive are in
// the "no API counterpart" fallbacks (a plan list element or stage with no matching
// API element, where the plan element still carries unknown computed children).
// Terraform core rejects an apply result that contains any unknown value, but a
// typed null is consistency-legal for a Computed attribute — so this final pass
// guarantees the create/update result is fully known.
func stripUnknowns(v tftypes.Value) tftypes.Value {
	if !v.IsKnown() {
		return tftypes.NewValue(v.Type(), nil) // typed null
	}
	if v.IsNull() {
		return v
	}
	ty := v.Type()
	switch {
	case ty.Is(tftypes.Object{}), ty.Is(tftypes.Map{}):
		var m map[string]tftypes.Value
		if err := v.As(&m); err != nil {
			return v
		}
		out := make(map[string]tftypes.Value, len(m))
		for k, e := range m {
			out[k] = stripUnknowns(e)
		}
		return tftypes.NewValue(v.Type(), out)
	case ty.Is(tftypes.List{}), ty.Is(tftypes.Set{}), ty.Is(tftypes.Tuple{}):
		var l []tftypes.Value
		if err := v.As(&l); err != nil {
			return v
		}
		out := make([]tftypes.Value, len(l))
		for i, e := range l {
			out[i] = stripUnknowns(e)
		}
		return tftypes.NewValue(v.Type(), out)
	default:
		return v
	}
}

// stageNameFromRaw extracts a stage object's `name` attribute as a string.
func stageNameFromRaw(v tftypes.Value) (string, bool) {
	if !v.IsKnown() || v.IsNull() {
		return "", false
	}
	var m map[string]tftypes.Value
	if err := v.As(&m); err != nil {
		return "", false
	}
	nv, ok := m["name"]
	if !ok || !nv.IsKnown() || nv.IsNull() {
		return "", false
	}
	var s string
	if err := nv.As(&s); err != nil {
		return "", false
	}
	return s, true
}

func (r *releaseDefinitionFrameworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: "Manages an Azure DevOps release definition (pipeline).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					useStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the release definition.",
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the project that owns this release definition.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             staticString(`\`),
				MarkdownDescription: "Folder path within the release definitions hierarchy.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             staticString(""),
				MarkdownDescription: "Optional description for this release definition.",
			},
			"release_name_format": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             staticString("Release-$(rev:r)"),
				MarkdownDescription: "Format string for auto-generated release names.",
			},
			"tags": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Definition-level tags applied to the release definition.",
				PlanModifiers:       []planmodifier.Set{useStateForUnknownSetModifier()},
			},
			"revision": schema.Int64Attribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{useStateForUnknownInt64Modifier()},
				MarkdownDescription: "Current revision number (managed by ADO; used for optimistic concurrency).",
			},
			"stages": schema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: "List of deployment stages (mapped from ADO 'environments').",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Name of the stage.",
						},
						"rank": schema.Int64Attribute{
							Optional:            true,
							Computed:            true,
							Default:             staticInt64(1),
							MarkdownDescription: "Execution order rank of the stage (1-based).",
						},
						"condition": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "Conditions that must be met before the stage runs.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"condition_type": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Type of condition: event, environmentState, artifact, undefined.",
									},
									"name": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Condition name.",
									},
									"value": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticString(""),
										MarkdownDescription: "Condition value expression.",
									},
								},
							},
						},
						"environment_options": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "Stage-level execution options (max 1 block).",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"email_notification_type": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticString("OnlyOnFailure"),
										MarkdownDescription: "When to send email notifications: OnlyOnFailure, Always, Never.",
									},
									"email_recipients": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										MarkdownDescription: "(deprecated) Email recipients for deployment notifications. ADO defaults to 'release.environment.owner;release.creator'.",
									},
									"skip_artifacts_download": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticBool(false),
										MarkdownDescription: "(deprecated) Whether to skip artifact download for this stage. Prefer deployment_input.skip_artifacts_download.",
									},
									"timeout_in_minutes": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             staticInt64(0),
										MarkdownDescription: "(deprecated) Timeout in minutes for this stage. Prefer deployment_input.timeout_in_minutes.",
									},
									"enable_access_token": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticBool(false),
										MarkdownDescription: "(deprecated) Whether to enable OAuth access token. Prefer deployment_input.enable_access_token.",
									},
									"publish_deployment_status": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticBool(false),
										MarkdownDescription: "Whether to publish deployment status to the ADO build.",
									},
									"badge_enabled": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticBool(false),
										MarkdownDescription: "Whether to show a badge for this stage.",
									},
									"auto_link_workitems": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticBool(false),
										MarkdownDescription: "Whether to auto-link work items to the deployment.",
									},
									"pull_request_deployment_enabled": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticBool(false),
										MarkdownDescription: "Whether pull request deployments are enabled for this stage.",
									},
									"publish_deployment_status_to_devops_project_settings": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticBool(false),
										MarkdownDescription: "Whether to publish deployment status to the project settings page.",
									},
								},
							},
						},
						"retention_policy": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "Retention policy for this stage's releases (max 1 block).",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"days_to_keep": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             staticInt64(30),
										MarkdownDescription: "Number of days to keep releases.",
									},
									"releases_to_keep": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             staticInt64(3),
										MarkdownDescription: "Number of releases to retain.",
									},
									"retain_build": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticBool(true),
										MarkdownDescription: "Whether to retain the triggering build.",
									},
								},
							},
						},
						"pre_deploy_approval": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "Pre-deployment approval configuration (max 1 block).",
							NestedObject: schema.NestedAttributeObject{
								Attributes: approvalNestedAttributes(),
							},
						},
						"post_deploy_approval": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "Post-deployment approval configuration (max 1 block).",
							NestedObject: schema.NestedAttributeObject{
								Attributes: approvalNestedAttributes(),
							},
						},
						"pre_deployment_gates": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "Pre-deployment gates configuration (max 1 block).",
							NestedObject: schema.NestedAttributeObject{
								Attributes: deploymentGatesNestedAttributes(),
							},
						},
						"post_deployment_gates": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "Post-deployment gates configuration (max 1 block).",
							NestedObject: schema.NestedAttributeObject{
								Attributes: deploymentGatesNestedAttributes(),
							},
						},
						"deploy_phase": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "Deploy phases for this stage.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: deployPhaseNestedAttributes(),
							},
						},
						"variables": schema.MapNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.Map{useStateForUnknownMapModifier()},
							MarkdownDescription: "Stage-level variables keyed by variable name.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: variableValueNestedAttributes(),
							},
						},
						"variable_groups": schema.ListAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							ElementType:         types.Int64Type,
							MarkdownDescription: "IDs of variable groups linked to this stage.",
						},
						"execution_policy": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "Execution policy for concurrent deployments (max 1 block).",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"concurrency_count": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             staticInt64(0),
										MarkdownDescription: "Max concurrent deployments (0 = unlimited).",
									},
									"queue_depth_count": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             staticInt64(0),
										MarkdownDescription: "Max queue depth (0 = unlimited).",
									},
								},
							},
						},
						"environment_trigger": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "Environment-level triggers (e.g. rollback/redeploy).",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"definition_environment_id": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										MarkdownDescription: "ID of the source environment for this trigger (ADO-assigned when not set).",
									},
									"trigger_type": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticString(""),
										MarkdownDescription: "Trigger type: rollbackRedeploy, deploymentGroupRedeploy, undefined.",
									},
									"trigger_content": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										MarkdownDescription: "Serialised trigger content (ADO internal).",
									},
								},
							},
						},
						"schedule": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "Stage-level scheduled deployment triggers.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"days_to_release": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             staticInt64(127),
										MarkdownDescription: "Bitmask of days to release (127 = all days).",
									},
									"start_hours": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             staticInt64(0),
										MarkdownDescription: "Hour of day to start (0-23).",
									},
									"start_minutes": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             staticInt64(0),
										MarkdownDescription: "Minute of hour to start (0-59).",
									},
									"time_zone_id": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticString("UTC"),
										MarkdownDescription: "Time zone identifier (e.g. 'UTC', 'Eastern Standard Time').",
									},
									"job_id": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "ADO-assigned job UUID for the schedule.",
									},
								},
							},
						},
						"process_parameters": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "Process parameters (max 1 block).",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"input": schema.ListNestedAttribute{
										Optional:            true,
										Computed:            true,
										PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
										MarkdownDescription: "Input parameter definitions.",
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													Optional:            true,
													Computed:            true,
													Default:             staticString(""),
													MarkdownDescription: "Parameter name.",
												},
												"default_value": schema.StringAttribute{
													Optional:            true,
													Computed:            true,
													Default:             staticString(""),
													MarkdownDescription: "Default value.",
												},
												"parameter_type": schema.StringAttribute{
													Optional:            true,
													Computed:            true,
													Default:             staticString(""),
													MarkdownDescription: "Parameter type string.",
												},
											},
										},
									},
								},
							},
						},
						"properties": schema.MapAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.Map{useStateForUnknownMapModifier()},
							ElementType:         types.StringType,
							MarkdownDescription: "Stage-level properties (key/value string pairs). ADO wraps values in {\"$type\":\"System.String\",\"$value\":\"...\"} on the wire; this is handled automatically.",
						},
						"owner": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.String{useStateForUnknown()},
							MarkdownDescription: "Identity UUID of the stage owner (ADO-assigned when not set).",
						},
						"id": schema.Int64Attribute{
							Computed:            true,
							PlanModifiers:       []planmodifier.Int64{useStateForUnknownInt64Modifier()},
							MarkdownDescription: "ADO-assigned ID of this stage.",
						},
					},
				},
			},
			"variables": schema.MapNestedAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Map{useStateForUnknownMapModifier()},
				MarkdownDescription: "Definition-level variables keyed by variable name.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: variableValueNestedAttributes(),
				},
			},
			"variable_groups": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
				ElementType:         types.Int64Type,
				MarkdownDescription: "IDs of variable groups linked to this release definition.",
			},
			"artifact": schema.ListNestedAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
				MarkdownDescription: "Artifact sources linked to this release definition.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"alias": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Artifact source alias (e.g. '_build').",
						},
						"type": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Default:             staticString("Build"),
							MarkdownDescription: "Artifact type (e.g. 'Build', 'Git').",
						},
						"is_primary": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             staticBool(false),
							MarkdownDescription: "Whether this is the primary artifact.",
						},
						"definition_reference": schema.MapAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.Map{useStateForUnknownMapModifier()},
							ElementType:         types.StringType,
							MarkdownDescription: "Key/value pairs describing the artifact source (user-set keys round-trip; API-computed keys are stripped).",
						},
					},
				},
			},
			"triggers": schema.ListNestedAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
				MarkdownDescription: "Release triggers (CD artifact, schedule, source repo, container image, pull request).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cd_artifact_trigger": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "CD artifact trigger (max 1 per triggers block).",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"artifact_alias": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticString(""),
										MarkdownDescription: "Alias of the artifact that triggers this release.",
									},
									"branch_filters": schema.ListAttribute{
										Optional:            true,
										Computed:            true,
										ElementType:         types.StringType,
										MarkdownDescription: "Branch filter expressions for this CD trigger.",
									},
									"tag_filter": schema.ListNestedAttribute{
										Optional:            true,
										Computed:            true,
										PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
										MarkdownDescription: "Build-tag filter (ADO 7.1 does not persist tagFilter.pattern — only the tags list round-trips).",
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"tags": schema.ListAttribute{
													Optional:            true,
													Computed:            true,
													ElementType:         types.StringType,
													MarkdownDescription: "Build tags that must be present to trigger.",
												},
											},
										},
									},
									"use_build_definition_branch": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticBool(false),
										MarkdownDescription: "Use the build definition's default branch as the trigger source branch.",
									},
									"create_release_on_build_tagging": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticBool(false),
										MarkdownDescription: "Create a release when the build is tagged.",
									},
								},
							},
						},
						"schedule_trigger": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "Schedule trigger (max 1 per triggers block).",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"schedule_only_with_changes": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticBool(false),
										MarkdownDescription: "Only trigger when there are new changes.",
									},
									"start_hours": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             staticInt64(0),
										MarkdownDescription: "Hour of the day to start (0-23).",
									},
									"start_minutes": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             staticInt64(0),
										MarkdownDescription: "Minute of the hour to start (0-59).",
									},
									"time_zone_id": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticString("UTC"),
										MarkdownDescription: "Time zone identifier (e.g. 'UTC', 'Eastern Standard Time').",
									},
									"days_to_release": schema.Int64Attribute{
										Optional:            true,
										Computed:            true,
										Default:             staticInt64(0),
										MarkdownDescription: "Bitmask of days (0=all) on which to trigger.",
									},
								},
							},
						},
						"source_repo_trigger": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "Source repository trigger (triggerType=sourceRepo).",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"alias": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticString(""),
										MarkdownDescription: "Alias of the source repository artifact.",
									},
									"branch_filters": schema.ListAttribute{
										Optional:            true,
										Computed:            true,
										ElementType:         types.StringType,
										MarkdownDescription: "Branch filter expressions for this source-repo trigger.",
									},
								},
							},
						},
						"container_image_trigger": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "Container image trigger (triggerType=containerImage).",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"artifact_alias": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticString(""),
										MarkdownDescription: "Alias of the container-image artifact this trigger monitors.",
									},
									"label": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticString(""),
										MarkdownDescription: "Image tag/label to filter on (empty = any tag).",
									},
								},
							},
						},
						"pull_request_trigger": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
							MarkdownDescription: "Pull request trigger (triggerType=pullRequest).",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"artifact_alias": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticString(""),
										MarkdownDescription: "Alias of the artifact associated with the pull request trigger.",
									},
									"target_branches": schema.ListAttribute{
										Optional:            true,
										Computed:            true,
										ElementType:         types.StringType,
										MarkdownDescription: "Target branches that trigger on pull request (e.g. refs/heads/main).",
									},
									"tags": schema.ListAttribute{
										Optional:            true,
										Computed:            true,
										ElementType:         types.StringType,
										MarkdownDescription: "Pull request tags that must be present to trigger.",
									},
									"use_artifact_reference": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										Default:             staticBool(true),
										MarkdownDescription: "When true (default), ADO derives the pull request code repository reference from the linked source-based artifact. Required by ADO for artifact-linked pull request triggers.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// variableValueNestedAttributes returns the schema attributes for a single variable value nested object.
func variableValueNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"value": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Sensitive:           true,
			Default:             staticString(""),
			MarkdownDescription: "Variable value. Marked sensitive because secret variables are write-only (ADO never returns secret values on read).",
		},
		"is_secret": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticBool(false),
			MarkdownDescription: "Whether the variable is a secret.",
		},
		"allow_override": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticBool(false),
			MarkdownDescription: "Whether the variable can be overridden at release time.",
		},
	}
}

// approvalNestedAttributes returns the shared schema attributes for pre/post approval blocks.
func approvalNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"approver": schema.ListNestedAttribute{
			Optional:            true,
			MarkdownDescription: "List of approvers.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "UUID of the approver identity.",
					},
					"is_automated": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             staticBool(false),
						MarkdownDescription: "Whether this approval step is automated.",
					},
					"rank": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						Default:             staticInt64(1),
						MarkdownDescription: "Execution order of this approver.",
					},
				},
			},
		},
		"approval_options": schema.ListNestedAttribute{
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
			MarkdownDescription: "Approval policy options (max 1 block).",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"required_approver_count": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						Default:             staticInt64(0),
						MarkdownDescription: "Number of approvals required (0 = all).",
					},
					"release_creator_can_be_approver": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             staticBool(false),
						MarkdownDescription: "Whether the release creator can approve their own deployment.",
					},
					"enforce_identity_revalidation": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             staticBool(false),
						MarkdownDescription: "Whether to revalidate approver identity before completing.",
					},
					"timeout_in_minutes": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						Default:             staticInt64(0),
						MarkdownDescription: "Timeout in minutes (0 = no timeout).",
					},
					"execution_order": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             staticString("beforeGates"),
						MarkdownDescription: "Order of approval execution relative to gates: beforeGates, afterSuccessfulGates, afterGatesAlways.",
					},
					"auto_triggered_and_previous_environment_approved_can_be_skipped": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             staticBool(false),
						MarkdownDescription: "Whether auto-triggered deployments can skip approval if previous stage was approved.",
					},
				},
			},
		},
	}
}

// deploymentGatesNestedAttributes returns the schema attributes for pre/post deployment gates.
func deploymentGatesNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"gates_options": schema.ListNestedAttribute{
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
			MarkdownDescription: "Gate evaluation options (max 1 block).",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"is_enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             staticBool(false),
						MarkdownDescription: "Whether gates are enabled for this stage.",
					},
					"timeout": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						Default:             staticInt64(0),
						MarkdownDescription: "Timeout in minutes after which gates fail.",
					},
					"sampling_interval": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						Default:             staticInt64(0),
						MarkdownDescription: "Interval in minutes between gate evaluations.",
					},
					"stabilization_time": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						Default:             staticInt64(0),
						MarkdownDescription: "Delay in minutes before the first gate evaluation.",
					},
					"minimum_success_duration": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						Default:             staticInt64(0),
						MarkdownDescription: "Minimum duration in minutes for gates to stay green before passing.",
					},
				},
			},
		},
		"gate": schema.ListNestedAttribute{
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
			MarkdownDescription: "Individual gate definitions.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"task": schema.ListNestedAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
						MarkdownDescription: "Workflow tasks within this gate.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: gateTaskNestedAttributes(),
						},
					},
				},
			},
		},
	}
}

// deployPhaseNestedAttributes returns the schema attributes for a deploy_phase block.
func deployPhaseNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticString(""),
			MarkdownDescription: "Name of the deploy phase.",
			PlanModifiers: []planmodifier.String{
				// When the HCL omits `name` the Default("") fires, giving plan=""
				// while state holds the ADO-generated name (e.g. "Agent job").
				// useStateForEmpty() substitutes the state value so the plan
				// matches state and no perpetual diff is produced.
				useStateForEmpty(),
			},
		},
		"rank": schema.Int64Attribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Rank of the deploy phase.",
		},
		"phase_type": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticString("agentBasedDeployment"),
			MarkdownDescription: "Type of the deploy phase: agentBasedDeployment, runOnServer, machineGroupBasedDeployment.",
		},
		"deployment_input": schema.ListNestedAttribute{
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
			MarkdownDescription: "Deployment input configuration (max 1 block).",
			NestedObject: schema.NestedAttributeObject{
				Attributes: deploymentInputNestedAttributes(),
			},
		},
		"workflow_task": schema.ListNestedAttribute{
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.List{useStateForUnknownListModifier()},
			MarkdownDescription: "Workflow tasks for this deploy phase.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: workflowTaskNestedAttributes(),
			},
		},
	}
}

// deploymentInputNestedAttributes returns schema attributes for the deployment_input block.
func deploymentInputNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"queue_id": schema.Int64Attribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "ID of the agent queue to use for this phase.",
		},
		"condition": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticString("succeeded()"),
			MarkdownDescription: "Condition that must be satisfied before the phase runs.",
		},
		"timeout_in_minutes": schema.Int64Attribute{
			Optional:            true,
			Computed:            true,
			Default:             staticInt64(0),
			MarkdownDescription: "Timeout in minutes for the phase (0 = no timeout).",
		},
		"job_cancel_timeout_in_minutes": schema.Int64Attribute{
			Optional:            true,
			Computed:            true,
			Default:             staticInt64(1),
			MarkdownDescription: "Timeout in minutes for job cancellation (default 1).",
		},
		"skip_artifacts_download": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticBool(false),
			MarkdownDescription: "Whether to skip downloading artifacts for this phase.",
		},
		"enable_access_token": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticBool(false),
			MarkdownDescription: "Whether to enable OAuth access token for scripts in this phase.",
		},
		"agent_specification": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticString(""),
			MarkdownDescription: "Agent specification identifier (e.g. ubuntu-latest).",
		},
		"override_inputs": schema.MapAttribute{
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
			PlanModifiers:       []planmodifier.Map{useStateForUnknownMapModifier()},
			MarkdownDescription: "Phase-level task input overrides (e.g. for parameterised task groups).",
		},
		"demands": schema.ListAttribute{
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "List of demands for agent selection.",
		},
		"parallel_execution": schema.ListNestedAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "Parallel execution configuration (max 1 block).",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Parallel execution type: noPar, multiConfiguration, multiMachine.",
					},
					"max_number_of_agents": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						Default:             staticInt64(1),
						MarkdownDescription: "Maximum number of agents to use for parallel execution.",
					},
					"multipliers": schema.ListAttribute{
						Optional:            true,
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Variable names used as multipliers for multi-configuration.",
					},
				},
			},
		},
	}
}

// workflowTaskNestedAttributes returns schema attributes for a workflow_task block.
func workflowTaskNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"task_id": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "UUID of the task definition.",
		},
		"display_name": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticString(""),
			MarkdownDescription: "Display name for the task step.",
		},
		"version_spec": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticString("0.*"),
			MarkdownDescription: "Version spec for the task (e.g. '1.*', '0.*').",
		},
		"enabled": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticBool(true),
			MarkdownDescription: "Whether the task is enabled.",
		},
		"always_run": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticBool(false),
			MarkdownDescription: "Whether the task runs even if a previous step failed.",
		},
		"condition": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticString("succeeded()"),
			MarkdownDescription: "Run condition expression for the task.",
		},
		"timeout_in_minutes": schema.Int64Attribute{
			Optional:            true,
			Computed:            true,
			Default:             staticInt64(0),
			MarkdownDescription: "Timeout in minutes for the task (0 = no timeout).",
		},
		"retry_count_on_task_failure": schema.Int64Attribute{
			Optional:            true,
			Computed:            true,
			Default:             staticInt64(0),
			MarkdownDescription: "Number of retries on task failure.",
		},
		"inputs": schema.MapAttribute{
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "Task-specific input values as key/value pairs.",
		},
		"definition_type": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticString("task"),
			MarkdownDescription: "Task definition type (e.g. 'task', 'metaTask'). Defaults to 'task'.",
		},
	}
}

// gateTaskNestedAttributes returns schema attributes for a task within a gate block.
// Mirrors workflowTaskNestedAttributes but kept separate for gate-task customisation.
func gateTaskNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"task_id": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "UUID of the gate task definition.",
		},
		"display_name": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticString(""),
			MarkdownDescription: "Display name for the gate task step.",
		},
		"version_spec": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticString("0.*"),
			MarkdownDescription: "Version spec for the gate task (e.g. '0.*', '1.*').",
		},
		"enabled": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticBool(true),
			MarkdownDescription: "Whether the gate task is enabled.",
		},
		"always_run": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticBool(false),
			MarkdownDescription: "Whether the gate task always runs.",
		},
		"condition": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticString("succeeded()"),
			MarkdownDescription: "Run condition expression for the gate task.",
		},
		"timeout_in_minutes": schema.Int64Attribute{
			Optional:            true,
			Computed:            true,
			Default:             staticInt64(0),
			MarkdownDescription: "Timeout in minutes for the gate task (0 = no timeout).",
		},
		"retry_count_on_task_failure": schema.Int64Attribute{
			Optional:            true,
			Computed:            true,
			Default:             staticInt64(0),
			MarkdownDescription: "Number of retries on gate task failure.",
		},
		"inputs": schema.MapAttribute{
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "Gate task-specific input values as key/value pairs.",
		},
		"definition_type": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             staticString("task"),
			MarkdownDescription: "Gate task definition type (e.g. 'task', 'metaTask'). Defaults to 'task'.",
		},
	}
}

// -------------------------------------------------------------------------
// CRUD methods
// -------------------------------------------------------------------------

// Configure captures the provider-configured AggregatedClient.
// WI-5 wires the client into framework_provider.go; until then this is a no-op.
func (r *releaseDefinitionFrameworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.AggregatedClient, got %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *releaseDefinitionFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_release_definition Create: provider client not yet wired (see WI-5)")
		return
	}

	var model releaseDefinitionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	def, projectID, err := expandReleaseDefinitionFramework(ctx, &model)
	if err != nil {
		resp.Diagnostics.AddError("Expand error", err.Error())
		return
	}

	created, err := r.client.ReleaseClient.CreateReleaseDefinition(r.client.Ctx, releaseapi.CreateReleaseDefinitionArgs{
		ReleaseDefinition: def,
		Project:           &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("creating release definition: %s", err))
		return
	}

	model.ID = types.StringValue(strconv.Itoa(*created.Id))
	resp.Diagnostics.Append(flattenReleaseDefinitionFramework(ctx, created, projectID, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Plan-faithful result: every configured (plan-known) attribute is kept exactly
	// as planned; only server-assigned computed values (unknown in the plan) come
	// from the flattened API response. See mergePlanComputed.
	resp.State.Raw = stripUnknowns(mergePlanComputed(req.Plan.Raw, resp.State.Raw))
}

func (r *releaseDefinitionFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_release_definition Read: provider client not yet wired (see WI-5)")
		return
	}

	var model releaseDefinitionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	defID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("cannot parse definition ID %q: %s", model.ID.ValueString(), err))
		return
	}
	projectID := model.ProjectID.ValueString()

	def, err := r.client.ReleaseClient.GetReleaseDefinition(r.client.Ctx, releaseapi.GetReleaseDefinitionArgs{
		Project:      &projectID,
		DefinitionId: &defID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading release definition (ID: %d): %s", defID, err))
		return
	}

	resp.Diagnostics.Append(flattenReleaseDefinitionFramework(ctx, def, projectID, &model)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *releaseDefinitionFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_release_definition Update: provider client not yet wired (see WI-5)")
		return
	}

	var model releaseDefinitionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state releaseDefinitionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.ID = state.ID

	def, projectID, err := expandReleaseDefinitionFramework(ctx, &model)
	if err != nil {
		resp.Diagnostics.AddError("Expand error", err.Error())
		return
	}

	defID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("cannot parse definition ID %q: %s", model.ID.ValueString(), err))
		return
	}
	def.Id = &defID
	if !state.Revision.IsNull() && !state.Revision.IsUnknown() {
		rev := int(state.Revision.ValueInt64())
		def.Revision = &rev
	}

	updated, err := retryOnStaleRevision(r.client.Ctx, r.client.ReleaseClient, def, projectID, defID)
	if err != nil {
		resp.Diagnostics.AddError("Update error", fmt.Sprintf("updating release definition (ID: %d): %s", defID, err))
		return
	}

	resp.Diagnostics.Append(flattenReleaseDefinitionFramework(ctx, updated, projectID, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Plan-faithful result: keep every configured (plan-known) attribute as planned;
	// fill only server-assigned computed values from the API response. See
	// mergePlanComputed.
	resp.State.Raw = stripUnknowns(mergePlanComputed(req.Plan.Raw, resp.State.Raw))
}

func (r *releaseDefinitionFrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_release_definition Delete: provider client not yet wired (see WI-5)")
		return
	}

	var model releaseDefinitionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	defID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("cannot parse definition ID %q: %s", model.ID.ValueString(), err))
		return
	}
	projectID := model.ProjectID.ValueString()

	if err := r.client.ReleaseClient.DeleteReleaseDefinition(r.client.Ctx, releaseapi.DeleteReleaseDefinitionArgs{
		Project:      &projectID,
		DefinitionId: &defID,
	}); err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("deleting release definition (ID: %d): %s", defID, err))
	}
}

func (r *releaseDefinitionFrameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "<project_id>/<definition_id>"
	idx := rdIndexOf(req.ID, "/")
	if idx < 0 || idx == 0 || idx == len(req.ID)-1 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format: <project_id>/<definition_id>, got: %q", req.ID),
		)
		return
	}

	// Populate a minimal model with typed null/empty values for every list and
	// map attribute. The framework requires that all types.List / types.Map
	// fields carry element-type information when written to state — otherwise
	// the type-system sees tftypes.DynamicPseudoType and panics with
	// "MISSING TYPE" during the post-import Read. The Read call that follows
	// ImportState will overwrite these placeholders with real values.
	model := releaseDefinitionModel{
		ProjectID:         types.StringValue(req.ID[:idx]),
		ID:                types.StringValue(req.ID[idx+1:]),
		Name:              types.StringValue(""),
		Path:              types.StringValue(`\`),
		Description:       types.StringValue(""),
		ReleaseNameFormat: types.StringValue(""),
		Revision:          types.Int64Value(0),
		Tags:              types.SetNull(types.StringType),
		Stages:            types.ListNull(types.ObjectType{AttrTypes: stageAttrTypes}),
		Variables:         types.MapNull(types.ObjectType{AttrTypes: variableValueAttrTypes}),
		VariableGroups:    types.ListNull(types.Int64Type),
		Artifact:          types.ListNull(types.ObjectType{AttrTypes: artifactAttrTypes}),
		Triggers:          types.ListNull(types.ObjectType{AttrTypes: triggerAttrTypes}),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// retryOnStaleRevision calls UpdateReleaseDefinition and, if the API returns a stale-revision
// error (HTTP 400 with "old copy" in the message), re-reads the current revision via
// GetReleaseDefinition and retries the update once. It returns the updated definition or an error.
func retryOnStaleRevision(
	ctx context.Context,
	rc releaseapi.Client,
	def *releaseapi.ReleaseDefinition,
	projectID string,
	defID int,
) (*releaseapi.ReleaseDefinition, error) {
	updated, err := rc.UpdateReleaseDefinition(ctx, releaseapi.UpdateReleaseDefinitionArgs{
		ReleaseDefinition: def,
		Project:           &projectID,
	})
	if err != nil && strings.Contains(err.Error(), "old copy") {
		// Stale-revision conflict: re-read the live revision and retry once.
		freshDef, readErr := rc.GetReleaseDefinition(ctx, releaseapi.GetReleaseDefinitionArgs{
			Project:      &projectID,
			DefinitionId: &defID,
		})
		if readErr != nil {
			return nil, fmt.Errorf("re-reading release definition after revision conflict (ID: %d): %w", defID, readErr)
		}
		def.Revision = freshDef.Revision
		updated, err = rc.UpdateReleaseDefinition(ctx, releaseapi.UpdateReleaseDefinitionArgs{
			ReleaseDefinition: def,
			Project:           &projectID,
		})
	}
	return updated, err
}

// rdIndexOf returns the index of the first occurrence of sub in s, or -1.
func rdIndexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// -------------------------------------------------------------------------
// Expand: Terraform model → API objects
// -------------------------------------------------------------------------

// expandReleaseDefinitionFramework converts the top-level model + stages into
// a releaseapi.ReleaseDefinition ready to POST/PUT.
func expandReleaseDefinitionFramework(ctx context.Context, model *releaseDefinitionModel) (*releaseapi.ReleaseDefinition, string, error) {
	projectID := model.ProjectID.ValueString()

	def := &releaseapi.ReleaseDefinition{
		Name:              converter.String(model.Name.ValueString()),
		Path:              converter.String(model.Path.ValueString()),
		Description:       converter.String(model.Description.ValueString()),
		ReleaseNameFormat: converter.String(model.ReleaseNameFormat.ValueString()),
	}

	if !model.Stages.IsNull() && !model.Stages.IsUnknown() {
		stages, err := expandStagesFramework(ctx, model.Stages)
		if err != nil {
			return nil, projectID, fmt.Errorf("expanding stages: %w", err)
		}
		def.Environments = &stages
	}

	// Definition-level variables
	if !model.Variables.IsNull() && !model.Variables.IsUnknown() {
		vars, err := expandVariablesFramework(ctx, model.Variables)
		if err != nil {
			return nil, projectID, fmt.Errorf("expanding variables: %w", err)
		}
		def.Variables = vars
	}

	// Definition-level variable_groups
	if !model.VariableGroups.IsNull() && !model.VariableGroups.IsUnknown() {
		groups, err := expandVariableGroupsFramework(ctx, model.VariableGroups)
		if err != nil {
			return nil, projectID, fmt.Errorf("expanding variable_groups: %w", err)
		}
		def.VariableGroups = &groups
	}

	// Artifacts
	if !model.Artifact.IsNull() && !model.Artifact.IsUnknown() {
		artifacts, err := expandArtifactsFramework(ctx, model.Artifact)
		if err != nil {
			return nil, projectID, fmt.Errorf("expanding artifact: %w", err)
		}
		def.Artifacts = &artifacts
	}

	// Triggers
	if !model.Triggers.IsNull() && !model.Triggers.IsUnknown() {
		triggers, err := expandTriggersFramework(ctx, model.Triggers)
		if err != nil {
			return nil, projectID, fmt.Errorf("expanding triggers: %w", err)
		}
		def.Triggers = &triggers
	}

	// Definition-level tags
	if !model.Tags.IsNull() && !model.Tags.IsUnknown() {
		var tags []string
		if d := model.Tags.ElementsAs(ctx, &tags, false); d.HasError() {
			return nil, projectID, fmt.Errorf("expanding tags: %s", d)
		}
		def.Tags = &tags
	}

	return def, projectID, nil
}

// expandStagesFramework converts the stages list into []releaseapi.ReleaseDefinitionEnvironment.
func expandStagesFramework(ctx context.Context, stagesList types.List) ([]releaseapi.ReleaseDefinitionEnvironment, error) {
	var stageModels []stageModel
	if d := stagesList.ElementsAs(ctx, &stageModels, false); d.HasError() {
		return nil, fmt.Errorf("decoding stages: %s", d)
	}

	envs := make([]releaseapi.ReleaseDefinitionEnvironment, 0, len(stageModels))
	for _, s := range stageModels {
		env := releaseapi.ReleaseDefinitionEnvironment{
			Name: converter.String(s.Name.ValueString()),
		}
		if !s.Rank.IsNull() && !s.Rank.IsUnknown() {
			rank := int(s.Rank.ValueInt64())
			env.Rank = &rank
		}

		// Conditions
		if !s.Condition.IsNull() && !s.Condition.IsUnknown() {
			conditions, err := expandConditionsFramework(ctx, s.Condition)
			if err != nil {
				return nil, err
			}
			env.Conditions = &conditions
		}

		// EnvironmentOptions
		if !s.EnvironmentOptions.IsNull() && !s.EnvironmentOptions.IsUnknown() {
			opts, err := expandEnvironmentOptionsFramework(ctx, s.EnvironmentOptions)
			if err != nil {
				return nil, err
			}
			env.EnvironmentOptions = opts
		}

		// RetentionPolicy
		if !s.RetentionPolicy.IsNull() && !s.RetentionPolicy.IsUnknown() {
			rp, err := expandRetentionPolicyFramework(ctx, s.RetentionPolicy)
			if err != nil {
				return nil, err
			}
			env.RetentionPolicy = rp
		}

		// Pre/Post deploy approvals
		if !s.PreDeployApproval.IsNull() && !s.PreDeployApproval.IsUnknown() {
			approvals, err := expandApprovalsFramework(ctx, s.PreDeployApproval)
			if err != nil {
				return nil, fmt.Errorf("pre_deploy_approval: %w", err)
			}
			env.PreDeployApprovals = approvals
		}
		if !s.PostDeployApproval.IsNull() && !s.PostDeployApproval.IsUnknown() {
			approvals, err := expandApprovalsFramework(ctx, s.PostDeployApproval)
			if err != nil {
				return nil, fmt.Errorf("post_deploy_approval: %w", err)
			}
			env.PostDeployApprovals = approvals
		}

		// Pre/Post deployment gates
		if !s.PreDeploymentGates.IsNull() && !s.PreDeploymentGates.IsUnknown() {
			gates, err := expandDeploymentGatesFramework(ctx, s.PreDeploymentGates)
			if err != nil {
				return nil, fmt.Errorf("pre_deployment_gates: %w", err)
			}
			env.PreDeploymentGates = gates
		}
		if !s.PostDeploymentGates.IsNull() && !s.PostDeploymentGates.IsUnknown() {
			gates, err := expandDeploymentGatesFramework(ctx, s.PostDeploymentGates)
			if err != nil {
				return nil, fmt.Errorf("post_deployment_gates: %w", err)
			}
			env.PostDeploymentGates = gates
		}

		// Deploy phases
		if !s.DeployPhase.IsNull() && !s.DeployPhase.IsUnknown() {
			phases, err := expandDeployPhasesFramework(ctx, s.DeployPhase)
			if err != nil {
				return nil, fmt.Errorf("deploy_phase: %w", err)
			}
			env.DeployPhases = &phases
		}

		// Stage-level variables
		if !s.Variables.IsNull() && !s.Variables.IsUnknown() {
			vars, err := expandVariablesFramework(ctx, s.Variables)
			if err != nil {
				return nil, fmt.Errorf("stage variables: %w", err)
			}
			env.Variables = vars
		}

		// Stage-level variable_groups
		if !s.VariableGroups.IsNull() && !s.VariableGroups.IsUnknown() {
			groups, err := expandVariableGroupsFramework(ctx, s.VariableGroups)
			if err != nil {
				return nil, fmt.Errorf("stage variable_groups: %w", err)
			}
			env.VariableGroups = &groups
		}

		// Execution policy
		if !s.ExecutionPolicy.IsNull() && !s.ExecutionPolicy.IsUnknown() {
			var epModels []executionPolicyModel
			if d := s.ExecutionPolicy.ElementsAs(ctx, &epModels, false); !d.HasError() && len(epModels) > 0 {
				ep := epModels[0]
				cc := int(ep.ConcurrencyCount.ValueInt64())
				qd := int(ep.QueueDepthCount.ValueInt64())
				env.ExecutionPolicy = &releaseapi.EnvironmentExecutionPolicy{
					ConcurrencyCount: &cc,
					QueueDepthCount:  &qd,
				}
			}
		}

		// Environment triggers
		if !s.EnvironmentTrigger.IsNull() && !s.EnvironmentTrigger.IsUnknown() {
			var trigModels []environmentTriggerModel
			if d := s.EnvironmentTrigger.ElementsAs(ctx, &trigModels, false); !d.HasError() && len(trigModels) > 0 {
				triggers := make([]releaseapi.EnvironmentTrigger, 0, len(trigModels))
				for _, tm := range trigModels {
					t := releaseapi.EnvironmentTrigger{}
					if !tm.DefinitionEnvironmentID.IsNull() && !tm.DefinitionEnvironmentID.IsUnknown() {
						id := int(tm.DefinitionEnvironmentID.ValueInt64())
						t.DefinitionEnvironmentId = &id
					}
					if !tm.TriggerType.IsNull() && !tm.TriggerType.IsUnknown() && tm.TriggerType.ValueString() != "" {
						ttVal := releaseapi.EnvironmentTriggerType(tm.TriggerType.ValueString())
						t.TriggerType = &ttVal
					}
					if !tm.TriggerContent.IsNull() && !tm.TriggerContent.IsUnknown() && tm.TriggerContent.ValueString() != "" {
						tc := tm.TriggerContent.ValueString()
						t.TriggerContent = &tc
					}
					triggers = append(triggers, t)
				}
				env.EnvironmentTriggers = &triggers
			}
		}

		// Stage-level schedules
		if !s.Schedule.IsNull() && !s.Schedule.IsUnknown() {
			var schedModels []stageScheduleModel
			if d := s.Schedule.ElementsAs(ctx, &schedModels, false); !d.HasError() && len(schedModels) > 0 {
				scheds := make([]releaseapi.ReleaseSchedule, 0, len(schedModels))
				for _, sm := range schedModels {
					rs := releaseapi.ReleaseSchedule{}
					if !sm.DaysToRelease.IsNull() && !sm.DaysToRelease.IsUnknown() {
						sd := releaseapi.ScheduleDays(strconv.Itoa(int(sm.DaysToRelease.ValueInt64())))
						rs.DaysToRelease = &sd
					}
					if !sm.StartHours.IsNull() && !sm.StartHours.IsUnknown() {
						h := int(sm.StartHours.ValueInt64())
						rs.StartHours = &h
					}
					if !sm.StartMinutes.IsNull() && !sm.StartMinutes.IsUnknown() {
						m := int(sm.StartMinutes.ValueInt64())
						rs.StartMinutes = &m
					}
					if !sm.TimeZoneID.IsNull() && !sm.TimeZoneID.IsUnknown() && sm.TimeZoneID.ValueString() != "" {
						tz := sm.TimeZoneID.ValueString()
						rs.TimeZoneId = &tz
					}
					if !sm.JobID.IsNull() && !sm.JobID.IsUnknown() && sm.JobID.ValueString() != "" {
						if parsed, err := uuid.Parse(sm.JobID.ValueString()); err == nil {
							rs.JobId = &parsed
						}
					}
					scheds = append(scheds, rs)
				}
				env.Schedules = &scheds
			}
		}

		// Process parameters
		if !s.ProcessParameters.IsNull() && !s.ProcessParameters.IsUnknown() {
			var ppModels []processParametersModel
			if d := s.ProcessParameters.ElementsAs(ctx, &ppModels, false); !d.HasError() && len(ppModels) > 0 {
				pm := ppModels[0]
				pp := &distributedtaskcommon.ProcessParameters{}
				if !pm.Input.IsNull() && !pm.Input.IsUnknown() {
					var inputModels []processParameterInputModel
					if d2 := pm.Input.ElementsAs(ctx, &inputModels, false); !d2.HasError() && len(inputModels) > 0 {
						inputs := make([]distributedtaskcommon.TaskInputDefinitionBase, 0, len(inputModels))
						for _, im := range inputModels {
							inp := distributedtaskcommon.TaskInputDefinitionBase{}
							if !im.Name.IsNull() && !im.Name.IsUnknown() {
								inp.Name = converter.String(im.Name.ValueString())
							}
							if !im.DefaultValue.IsNull() && !im.DefaultValue.IsUnknown() {
								inp.DefaultValue = converter.String(im.DefaultValue.ValueString())
							}
							if !im.ParameterType.IsNull() && !im.ParameterType.IsUnknown() {
								inp.Type = converter.String(im.ParameterType.ValueString())
							}
							inputs = append(inputs, inp)
						}
						pp.Inputs = &inputs
					}
				}
				env.ProcessParameters = pp
			}
		}

		// Properties — wrap each value as {"$type":"System.String","$value":"..."}
		if !s.Properties.IsNull() && !s.Properties.IsUnknown() {
			var propsMap map[string]string
			if d := s.Properties.ElementsAs(ctx, &propsMap, false); !d.HasError() && len(propsMap) > 0 {
				wrapped := make(map[string]interface{}, len(propsMap))
				for k, v := range propsMap {
					wrapped[k] = map[string]interface{}{
						"$type":  "System.String",
						"$value": v,
					}
				}
				env.Properties = wrapped
			}
		}

		// Owner
		if !s.Owner.IsNull() && !s.Owner.IsUnknown() && s.Owner.ValueString() != "" {
			ownerID := s.Owner.ValueString()
			env.Owner = &webapi.IdentityRef{Id: &ownerID}
		}

		envs = append(envs, env)
	}
	return envs, nil
}

// expandDeployPhasesFramework converts the deploy_phase list into []interface{} for the ADO API.
// The ADO API expects deploy phases as raw maps (not typed structs).
func expandDeployPhasesFramework(ctx context.Context, list types.List) ([]interface{}, error) {
	var phaseModels []deployPhaseModel
	if d := list.ElementsAs(ctx, &phaseModels, false); d.HasError() {
		return nil, fmt.Errorf("decoding deploy_phase: %s", d)
	}

	phases := make([]interface{}, 0, len(phaseModels))
	for _, p := range phaseModels {
		// ADO rejects empty deploy-phase names (VS402859). Default to a
		// conventional name so that omitting the optional `name` attribute in
		// HCL still produces a valid API payload.
		phaseName := p.Name.ValueString()
		if phaseName == "" {
			switch p.PhaseType.ValueString() {
			case "runOnServer":
				phaseName = "Agentless job"
			default:
				phaseName = "Agent job"
			}
		}
		phaseRaw := map[string]interface{}{
			"name":      phaseName,
			"rank":      int(p.Rank.ValueInt64()),
			"phaseType": p.PhaseType.ValueString(),
		}

		// deployment_input
		if !p.DeploymentInput.IsNull() && !p.DeploymentInput.IsUnknown() {
			var diModels []deploymentInputModel
			if d := p.DeploymentInput.ElementsAs(ctx, &diModels, false); d.HasError() {
				return nil, fmt.Errorf("decoding deployment_input: %s", d)
			}
			if len(diModels) > 0 {
				di := expandDeploymentInputFramework(ctx, diModels[0], p.PhaseType.ValueString())
				phaseRaw["deploymentInput"] = di
			}
		}

		// workflow_task
		if !p.WorkflowTask.IsNull() && !p.WorkflowTask.IsUnknown() {
			var taskModels []workflowTaskModel
			if d := p.WorkflowTask.ElementsAs(ctx, &taskModels, false); d.HasError() {
				return nil, fmt.Errorf("decoding workflow_task: %s", d)
			}
			if len(taskModels) > 0 {
				tasks := expandWorkflowTasksFramework(ctx, taskModels)
				phaseRaw["workflowTasks"] = tasks
			}
		}

		phases = append(phases, phaseRaw)
	}
	return phases, nil
}

// expandDeploymentInputFramework converts a deploymentInputModel to the ADO API map shape.
func expandDeploymentInputFramework(ctx context.Context, m deploymentInputModel, phaseType string) map[string]interface{} {
	di := map[string]interface{}{
		"timeoutInMinutes":          int(m.TimeoutInMinutes.ValueInt64()),
		"jobCancelTimeoutInMinutes": int(m.JobCancelTimeoutInMinutes.ValueInt64()),
		"condition":                 m.Condition.ValueString(),
		"skipArtifactsDownload":     m.SkipArtifactsDownload.ValueBool(),
		"enableAccessToken":         m.EnableAccessToken.ValueBool(),
	}

	// Only include queueId for agent-based phases.
	queueID := int(m.QueueID.ValueInt64())
	if phaseType != "runOnServer" && queueID != 0 {
		di["queueId"] = queueID
	}

	// Agent specification
	agentSpec := m.AgentSpecification.ValueString()
	if agentSpec != "" {
		di["agentSpecification"] = map[string]interface{}{"identifier": agentSpec}
	}

	// Override inputs
	if !m.OverrideInputs.IsNull() && !m.OverrideInputs.IsUnknown() {
		var oiMap map[string]string
		if d := m.OverrideInputs.ElementsAs(ctx, &oiMap, false); !d.HasError() && len(oiMap) > 0 {
			di["overrideInputs"] = oiMap
		}
	}

	// Demands
	if !m.Demands.IsNull() && !m.Demands.IsUnknown() {
		var demandStrs []string
		if d := m.Demands.ElementsAs(ctx, &demandStrs, false); !d.HasError() && len(demandStrs) > 0 {
			di["demands"] = demandStrs
		}
	}

	// Parallel execution
	if !m.ParallelExecution.IsNull() && !m.ParallelExecution.IsUnknown() {
		var peModels []parallelExecutionModel
		if d := m.ParallelExecution.ElementsAs(ctx, &peModels, false); !d.HasError() && len(peModels) > 0 {
			di["parallelExecution"] = expandParallelExecutionFramework(ctx, peModels[0])
		}
	}

	return di
}

// expandParallelExecutionFramework converts a parallelExecutionModel to the ADO API map.
func expandParallelExecutionFramework(ctx context.Context, m parallelExecutionModel) map[string]interface{} {
	peType := m.Type.ValueString()
	if peType == "" {
		peType = "none"
	}
	result := map[string]interface{}{
		"parallelExecutionType": peType,
	}
	maxAgents := int(m.MaxNumberOfAgents.ValueInt64())
	if maxAgents > 0 {
		result["maxNumberOfAgents"] = maxAgents
	}
	if !m.Multipliers.IsNull() && !m.Multipliers.IsUnknown() {
		var multipliers []string
		if d := m.Multipliers.ElementsAs(ctx, &multipliers, false); !d.HasError() && len(multipliers) > 0 {
			// ADO stores multipliers as comma-joined string.
			joined := ""
			for i, mult := range multipliers {
				if i > 0 {
					joined += ","
				}
				joined += mult
			}
			result["multipliers"] = joined
		}
	}
	return result
}

// expandWorkflowTasksFramework converts []workflowTaskModel to []releaseapi.WorkflowTask.
func expandWorkflowTasksFramework(ctx context.Context, models []workflowTaskModel) []releaseapi.WorkflowTask {
	tasks := make([]releaseapi.WorkflowTask, 0, len(models))
	for _, m := range models {
		timeout := int(m.TimeoutInMinutes.ValueInt64())
		retryCount := int(m.RetryCountOnTaskFailure.ValueInt64())
		task := releaseapi.WorkflowTask{
			Name:                    converter.String(m.DisplayName.ValueString()),
			Version:                 converter.String(m.VersionSpec.ValueString()),
			Enabled:                 converter.Bool(m.Enabled.ValueBool()),
			AlwaysRun:               converter.Bool(m.AlwaysRun.ValueBool()),
			Condition:               converter.String(m.Condition.ValueString()),
			TimeoutInMinutes:        &timeout,
			RetryCountOnTaskFailure: &retryCount,
		}

		if taskIDStr := m.TaskID.ValueString(); taskIDStr != "" {
			if parsed, err := uuid.Parse(taskIDStr); err == nil {
				task.TaskId = &parsed
			}
		}

		if !m.DefinitionType.IsNull() && !m.DefinitionType.IsUnknown() && m.DefinitionType.ValueString() != "" {
			task.DefinitionType = converter.String(m.DefinitionType.ValueString())
		}

		if !m.Inputs.IsNull() && !m.Inputs.IsUnknown() {
			var inputMap map[string]string
			if d := m.Inputs.ElementsAs(ctx, &inputMap, false); !d.HasError() && len(inputMap) > 0 {
				task.Inputs = &inputMap
			}
		}

		tasks = append(tasks, task)
	}
	return tasks
}

// expandGateTasksFramework converts []gateTaskModel to []releaseapi.WorkflowTask for a gate.
func expandGateTasksFramework(ctx context.Context, models []gateTaskModel) []releaseapi.WorkflowTask {
	tasks := make([]releaseapi.WorkflowTask, 0, len(models))
	for _, m := range models {
		timeout := int(m.TimeoutInMinutes.ValueInt64())
		retryCount := int(m.RetryCountOnTaskFailure.ValueInt64())
		task := releaseapi.WorkflowTask{
			Name:                    converter.String(m.DisplayName.ValueString()),
			Version:                 converter.String(m.VersionSpec.ValueString()),
			Enabled:                 converter.Bool(m.Enabled.ValueBool()),
			AlwaysRun:               converter.Bool(m.AlwaysRun.ValueBool()),
			Condition:               converter.String(m.Condition.ValueString()),
			TimeoutInMinutes:        &timeout,
			RetryCountOnTaskFailure: &retryCount,
		}
		if taskIDStr := m.TaskID.ValueString(); taskIDStr != "" {
			if parsed, err := uuid.Parse(taskIDStr); err == nil {
				task.TaskId = &parsed
			}
		}
		if !m.DefinitionType.IsNull() && !m.DefinitionType.IsUnknown() && m.DefinitionType.ValueString() != "" {
			task.DefinitionType = converter.String(m.DefinitionType.ValueString())
		}
		if !m.Inputs.IsNull() && !m.Inputs.IsUnknown() {
			var inputMap map[string]string
			if d := m.Inputs.ElementsAs(ctx, &inputMap, false); !d.HasError() && len(inputMap) > 0 {
				task.Inputs = &inputMap
			}
		}
		tasks = append(tasks, task)
	}
	return tasks
}

func expandConditionsFramework(ctx context.Context, list types.List) ([]releaseapi.Condition, error) {
	var condModels []conditionModel
	if d := list.ElementsAs(ctx, &condModels, false); d.HasError() {
		return nil, fmt.Errorf("decoding conditions: %s", d)
	}
	conditions := make([]releaseapi.Condition, 0, len(condModels))
	for _, c := range condModels {
		ct := releaseapi.ConditionType(c.ConditionType.ValueString())
		conditions = append(conditions, releaseapi.Condition{
			ConditionType: &ct,
			Name:          converter.String(c.Name.ValueString()),
			Value:         converter.String(c.Value.ValueString()),
		})
	}
	return conditions, nil
}

func expandEnvironmentOptionsFramework(ctx context.Context, list types.List) (*releaseapi.EnvironmentOptions, error) {
	var models []environmentOptionsModel
	if d := list.ElementsAs(ctx, &models, false); d.HasError() {
		return nil, fmt.Errorf("decoding environment_options: %s", d)
	}
	if len(models) == 0 {
		return nil, nil
	}
	m := models[0]
	//nolint:staticcheck // deprecated ADO fields preserved for SDK v2 parity
	opts := &releaseapi.EnvironmentOptions{
		EmailNotificationType:        converter.String(m.EmailNotificationType.ValueString()),
		SkipArtifactsDownload:        converter.Bool(m.SkipArtifactsDownload.ValueBool()),
		EnableAccessToken:            converter.Bool(m.EnableAccessToken.ValueBool()),
		PublishDeploymentStatus:      converter.Bool(m.PublishDeploymentStatus.ValueBool()),
		BadgeEnabled:                 converter.Bool(m.BadgeEnabled.ValueBool()),
		AutoLinkWorkItems:            converter.Bool(m.AutoLinkWorkItems.ValueBool()),
		PullRequestDeploymentEnabled: converter.Bool(m.PullRequestDeploymentEnabled.ValueBool()),
	}
	if !m.TimeoutInMinutes.IsNull() && !m.TimeoutInMinutes.IsUnknown() {
		v := int(m.TimeoutInMinutes.ValueInt64())
		opts.TimeoutInMinutes = &v //nolint:staticcheck
	}
	// API rejects empty string EmailRecipients; omit blank so ADO uses its default.
	if !m.EmailRecipients.IsNull() && !m.EmailRecipients.IsUnknown() && m.EmailRecipients.ValueString() != "" {
		opts.EmailRecipients = converter.String(m.EmailRecipients.ValueString()) //nolint:staticcheck
	}
	return opts, nil
}

func expandRetentionPolicyFramework(ctx context.Context, list types.List) (*releaseapi.EnvironmentRetentionPolicy, error) {
	var models []retentionPolicyModel
	if d := list.ElementsAs(ctx, &models, false); d.HasError() {
		return nil, fmt.Errorf("decoding retention_policy: %s", d)
	}
	if len(models) == 0 {
		return nil, nil
	}
	m := models[0]
	dtk := int(m.DaysToKeep.ValueInt64())
	rtk := int(m.ReleasesToKeep.ValueInt64())
	return &releaseapi.EnvironmentRetentionPolicy{
		DaysToKeep:     &dtk,
		ReleasesToKeep: &rtk,
		RetainBuild:    converter.Bool(m.RetainBuild.ValueBool()),
	}, nil
}

func expandApprovalsFramework(ctx context.Context, list types.List) (*releaseapi.ReleaseDefinitionApprovals, error) {
	var models []approvalModel
	if d := list.ElementsAs(ctx, &models, false); d.HasError() {
		return nil, fmt.Errorf("decoding approvals: %s", d)
	}
	if len(models) == 0 {
		return nil, nil
	}
	m := models[0]
	approvals := &releaseapi.ReleaseDefinitionApprovals{}

	// Approvers
	if !m.Approver.IsNull() && !m.Approver.IsUnknown() {
		var approverModels []approverModel
		if d := m.Approver.ElementsAs(ctx, &approverModels, false); d.HasError() {
			return nil, fmt.Errorf("decoding approvers: %s", d)
		}
		steps := make([]releaseapi.ReleaseDefinitionApprovalStep, 0, len(approverModels))
		for _, a := range approverModels {
			rank := int(a.Rank.ValueInt64())
			step := releaseapi.ReleaseDefinitionApprovalStep{
				IsAutomated: converter.Bool(a.IsAutomated.ValueBool()),
				Rank:        &rank,
			}
			if !a.ID.IsNull() && !a.ID.IsUnknown() && a.ID.ValueString() != "" {
				id := a.ID.ValueString()
				step.Approver = &webapi.IdentityRef{Id: &id}
			}
			steps = append(steps, step)
		}
		approvals.Approvals = &steps
	}

	// ApprovalOptions
	if !m.ApprovalOptions.IsNull() && !m.ApprovalOptions.IsUnknown() {
		var optModels []approvalOptionsModel
		if d := m.ApprovalOptions.ElementsAs(ctx, &optModels, false); d.HasError() {
			return nil, fmt.Errorf("decoding approval_options: %s", d)
		}
		if len(optModels) > 0 {
			o := optModels[0]
			reqCount := int(o.RequiredApproverCount.ValueInt64())
			timeoutMin := int(o.TimeoutInMinutes.ValueInt64())
			execOrder := releaseapi.ApprovalExecutionOrder(o.ExecutionOrder.ValueString())
			approvals.ApprovalOptions = &releaseapi.ApprovalOptions{
				RequiredApproverCount:       &reqCount,
				ReleaseCreatorCanBeApprover: converter.Bool(o.ReleaseCreatorCanBeApprover.ValueBool()),
				EnforceIdentityRevalidation: converter.Bool(o.EnforceIdentityRevalidation.ValueBool()),
				TimeoutInMinutes:            &timeoutMin,
				ExecutionOrder:              &execOrder,
				AutoTriggeredAndPreviousEnvironmentApprovedCanBeSkipped: converter.Bool(o.AutoTriggeredAndPreviousEnvironmentApprovedCanBeSkipped.ValueBool()),
			}
		}
	}

	return approvals, nil
}

func expandDeploymentGatesFramework(ctx context.Context, list types.List) (*releaseapi.ReleaseDefinitionGatesStep, error) {
	var models []deploymentGatesModel
	if d := list.ElementsAs(ctx, &models, false); d.HasError() {
		return nil, fmt.Errorf("decoding deployment gates: %s", d)
	}
	if len(models) == 0 {
		return nil, nil
	}
	m := models[0]
	step := &releaseapi.ReleaseDefinitionGatesStep{}

	if !m.GatesOptions.IsNull() && !m.GatesOptions.IsUnknown() {
		var optModels []gatesOptionsModel
		if d := m.GatesOptions.ElementsAs(ctx, &optModels, false); d.HasError() {
			return nil, fmt.Errorf("decoding gates_options: %s", d)
		}
		if len(optModels) > 0 {
			o := optModels[0]
			timeout := int(o.Timeout.ValueInt64())
			sampling := int(o.SamplingInterval.ValueInt64())
			stabilization := int(o.StabilizationTime.ValueInt64())
			minSuccess := int(o.MinimumSuccessDuration.ValueInt64())
			step.GatesOptions = &releaseapi.ReleaseDefinitionGatesOptions{
				IsEnabled:              converter.Bool(o.IsEnabled.ValueBool()),
				Timeout:                &timeout,
				SamplingInterval:       &sampling,
				StabilizationTime:      &stabilization,
				MinimumSuccessDuration: &minSuccess,
			}
		}
	}

	// Gates
	if !m.Gate.IsNull() && !m.Gate.IsUnknown() {
		var gateModels []gateModel
		if d := m.Gate.ElementsAs(ctx, &gateModels, false); d.HasError() {
			return nil, fmt.Errorf("decoding gate: %s", d)
		}
		gates := make([]releaseapi.ReleaseDefinitionGate, 0, len(gateModels))
		for _, g := range gateModels {
			gate := releaseapi.ReleaseDefinitionGate{}
			if !g.Task.IsNull() && !g.Task.IsUnknown() {
				var taskModels []gateTaskModel
				if d := g.Task.ElementsAs(ctx, &taskModels, false); !d.HasError() && len(taskModels) > 0 {
					wfTasks := expandGateTasksFramework(ctx, taskModels)
					gate.Tasks = &wfTasks
				}
			}
			gates = append(gates, gate)
		}
		step.Gates = &gates
	} else {
		step.Gates = &[]releaseapi.ReleaseDefinitionGate{}
	}

	return step, nil
}

// -------------------------------------------------------------------------
// Flatten: API objects → Terraform model
// -------------------------------------------------------------------------

// flattenReleaseDefinitionFramework writes API response fields into the model.
func flattenReleaseDefinitionFramework(ctx context.Context, def *releaseapi.ReleaseDefinition, projectID string, model *releaseDefinitionModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if def.Id != nil {
		model.ID = types.StringValue(strconv.Itoa(*def.Id))
	}
	model.ProjectID = types.StringValue(projectID)
	if def.Name != nil {
		model.Name = types.StringValue(*def.Name)
	}
	if def.Path != nil {
		model.Path = types.StringValue(*def.Path)
	}
	if def.Description != nil {
		model.Description = types.StringValue(*def.Description)
	} else {
		model.Description = types.StringValue("")
	}
	if def.ReleaseNameFormat != nil {
		model.ReleaseNameFormat = types.StringValue(*def.ReleaseNameFormat)
	}
	if def.Revision != nil {
		model.Revision = types.Int64Value(int64(*def.Revision))
	}

	// Capture existing plan/prior values before overwriting — needed for
	// secret variable preservation (ADO never returns secret values on read).
	existingStages := model.Stages
	existingVars := model.Variables

	// Stages
	if def.Environments != nil {
		stagesList, d := flattenStagesFramework(ctx, *def.Environments, existingStages)
		diags.Append(d...)
		model.Stages = stagesList
	} else {
		model.Stages = types.ListValueMust(types.ObjectType{AttrTypes: stageAttrTypes}, []attr.Value{})
	}

	// Definition-level variables
	defVars, d := flattenVariablesFramework(ctx, def.Variables, existingVars)
	diags.Append(d...)
	model.Variables = defVars

	// Definition-level variable_groups
	defVarGroups, d := flattenVariableGroupsFramework(def.VariableGroups)
	diags.Append(d...)
	model.VariableGroups = defVarGroups

	// Artifacts
	defArtifacts, d := flattenArtifactsFramework(ctx, def.Artifacts)
	diags.Append(d...)
	model.Artifact = defArtifacts

	// Triggers
	defTriggers, d := flattenTriggersFramework(ctx, def.Triggers)
	diags.Append(d...)
	model.Triggers = defTriggers

	// Definition-level tags. ADO's CreateReleaseDefinition response does not echo
	// tags even though it persists them, so when the API returns none, preserve
	// the planned/prior value (the model already carries it on create and read)
	// rather than clobbering to empty — only fall back to an empty set when the
	// model has no known value (e.g. import).
	if def.Tags != nil && len(*def.Tags) > 0 {
		tagVals := make([]attr.Value, len(*def.Tags))
		for i, t := range *def.Tags {
			tagVals[i] = types.StringValue(t)
		}
		tagsSet, d := types.SetValue(types.StringType, tagVals)
		diags.Append(d...)
		model.Tags = tagsSet
	} else if model.Tags.IsNull() || model.Tags.IsUnknown() {
		model.Tags = types.SetValueMust(types.StringType, []attr.Value{})
	}

	return diags
}

func flattenStagesFramework(ctx context.Context, envs []releaseapi.ReleaseDefinitionEnvironment, existingStages types.List) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	stageObjects := make([]attr.Value, 0, len(envs))

	// Build a name→existing-variables lookup from the prior/plan state so we
	// can preserve secret variable values that ADO never returns on read.
	existingStageVarsByName := map[string]types.Map{}
	if !existingStages.IsNull() && !existingStages.IsUnknown() {
		for _, stageVal := range existingStages.Elements() {
			stageObj, ok := stageVal.(types.Object)
			if !ok {
				continue
			}
			attrs := stageObj.Attributes()
			nameAttr, ok := attrs["name"].(types.String)
			if !ok || nameAttr.IsNull() || nameAttr.IsUnknown() {
				continue
			}
			if varsAttr, ok := attrs["variables"].(types.Map); ok {
				existingStageVarsByName[nameAttr.ValueString()] = varsAttr
			}
		}
	}

	for _, env := range envs {
		// conditions
		conditions, d := flattenConditionsFramework(env.Conditions)
		diags.Append(d...)

		// environment_options
		envOpts, d := flattenEnvironmentOptionsFramework(env.EnvironmentOptions)
		diags.Append(d...)

		// retention_policy
		retPolicy, d := flattenRetentionPolicyFramework(env.RetentionPolicy)
		diags.Append(d...)

		// pre/post deploy approvals
		preApproval, d := flattenApprovalsFramework(env.PreDeployApprovals)
		diags.Append(d...)

		postApproval, d := flattenApprovalsFramework(env.PostDeployApprovals)
		diags.Append(d...)

		// pre/post deployment gates
		preGates, d := flattenDeploymentGatesFramework(env.PreDeploymentGates)
		diags.Append(d...)

		postGates, d := flattenDeploymentGatesFramework(env.PostDeploymentGates)
		diags.Append(d...)

		// deploy phases
		deployPhases, d := flattenDeployPhasesFramework(ctx, env.DeployPhases)
		diags.Append(d...)

		// stage-level variables — pass existing stage vars for secret preservation
		stageName := ""
		if env.Name != nil {
			stageName = *env.Name
		}
		existingStageVars := existingStageVarsByName[stageName] // zero value (null map) if not found
		stageVars, d := flattenVariablesFramework(ctx, env.Variables, existingStageVars)
		diags.Append(d...)

		// stage-level variable_groups
		stageVarGroups, d := flattenVariableGroupsFramework(env.VariableGroups)
		diags.Append(d...)

		// Execution policy
		execPolicy, d := flattenExecutionPolicyFramework(env.ExecutionPolicy)
		diags.Append(d...)

		// Environment triggers
		envTriggers, d := flattenEnvironmentTriggersFramework(env.EnvironmentTriggers)
		diags.Append(d...)

		// Stage-level schedules
		stageSchedules, d := flattenStageSchedulesFramework(env.Schedules)
		diags.Append(d...)

		// Process parameters
		processParams, d := flattenProcessParametersFramework(env.ProcessParameters)
		diags.Append(d...)

		// Properties — unwrap ADO's {"$type":"System.String","$value":"..."} wrapper
		stageProps, d := flattenStagePropertiesFramework(env.Properties)
		diags.Append(d...)

		// Owner
		ownerStr := ""
		if env.Owner != nil && env.Owner.Id != nil {
			ownerStr = *env.Owner.Id
		}

		// Stage ID
		stageID := int64(0)
		if env.Id != nil {
			stageID = int64(*env.Id)
		}

		rank := int64(0)
		if env.Rank != nil {
			rank = int64(*env.Rank)
		}
		name := ""
		if env.Name != nil {
			name = *env.Name
		}

		stageObj, d := types.ObjectValue(stageAttrTypes, map[string]attr.Value{
			"name":                  types.StringValue(name),
			"rank":                  types.Int64Value(rank),
			"condition":             conditions,
			"environment_options":   envOpts,
			"retention_policy":      retPolicy,
			"pre_deploy_approval":   preApproval,
			"post_deploy_approval":  postApproval,
			"pre_deployment_gates":  preGates,
			"post_deployment_gates": postGates,
			"deploy_phase":          deployPhases,
			"variables":             stageVars,
			"variable_groups":       stageVarGroups,
			"execution_policy":      execPolicy,
			"environment_trigger":   envTriggers,
			"schedule":              stageSchedules,
			"process_parameters":    processParams,
			"properties":            stageProps,
			"owner":                 types.StringValue(ownerStr),
			"id":                    types.Int64Value(stageID),
		})
		diags.Append(d...)
		stageObjects = append(stageObjects, stageObj)
	}

	// Reorder the flattened stages to match the plan/prior stage order (by name).
	// ADO's create/update response — and read — can return environments in a
	// different order than the configured `stages` list (it stores by rank, and the
	// response order is not guaranteed to equal the config order). The framework
	// compares list elements positionally, and the variable `value` attribute is
	// Sensitive, so an order mismatch surfaces as
	// "Provider produced inconsistent result after apply: .stages: inconsistent
	// values for sensitive attribute". Aligning to the plan order keeps every
	// per-stage configured attribute (including sensitive values) index-matched.
	stageObjects = reorderStagesByName(stageObjects, existingStages)

	stageElemType := types.ObjectType{AttrTypes: stageAttrTypes}
	list, d := types.ListValue(stageElemType, stageObjects)
	diags.Append(d...)
	return list, diags
}

// reorderStagesByName returns the flattened stage objects reordered to match the
// order of stages in existingStages (the plan on create/update, the prior state
// on read), matched by the stage `name` attribute. A per-name queue of original
// indices preserves stable ordering across duplicate names. Stages whose names
// are not present in existingStages — and the whole list when existingStages is
// absent (e.g. import) — are kept in their original flatten order, so a stage is
// never dropped or duplicated.
func reorderStagesByName(stageObjects []attr.Value, existingStages types.List) []attr.Value {
	if existingStages.IsNull() || existingStages.IsUnknown() || len(stageObjects) <= 1 {
		return stageObjects
	}

	nameOf := func(v attr.Value) (string, bool) {
		obj, ok := v.(types.Object)
		if !ok {
			return "", false
		}
		nameAttr, ok := obj.Attributes()["name"].(types.String)
		if !ok || nameAttr.IsNull() || nameAttr.IsUnknown() {
			return "", false
		}
		return nameAttr.ValueString(), true
	}

	idxByName := map[string][]int{}
	for i, v := range stageObjects {
		if n, ok := nameOf(v); ok {
			idxByName[n] = append(idxByName[n], i)
		}
	}

	consumed := make([]bool, len(stageObjects))
	ordered := make([]attr.Value, 0, len(stageObjects))
	for _, stageVal := range existingStages.Elements() {
		obj, ok := stageVal.(types.Object)
		if !ok {
			continue
		}
		nameAttr, ok := obj.Attributes()["name"].(types.String)
		if !ok || nameAttr.IsNull() || nameAttr.IsUnknown() {
			continue
		}
		q := idxByName[nameAttr.ValueString()]
		if len(q) == 0 {
			continue
		}
		i := q[0]
		idxByName[nameAttr.ValueString()] = q[1:]
		ordered = append(ordered, stageObjects[i])
		consumed[i] = true
	}

	// Append any stages the plan order did not match, in their original order.
	for i, v := range stageObjects {
		if !consumed[i] {
			ordered = append(ordered, v)
		}
	}
	return ordered
}

func flattenConditionsFramework(conditions *[]releaseapi.Condition) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: conditionAttrTypes}
	if conditions == nil || len(*conditions) == 0 {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}
	var diags diag.Diagnostics
	objs := make([]attr.Value, 0, len(*conditions))
	for _, c := range *conditions {
		ct := ""
		if c.ConditionType != nil {
			ct = string(*c.ConditionType)
		}
		name := ""
		if c.Name != nil {
			name = *c.Name
		}
		val := ""
		if c.Value != nil {
			val = *c.Value
		}
		obj, d := types.ObjectValue(conditionAttrTypes, map[string]attr.Value{
			"condition_type": types.StringValue(ct),
			"name":           types.StringValue(name),
			"value":          types.StringValue(val),
		})
		diags.Append(d...)
		objs = append(objs, obj)
	}
	list, d := types.ListValue(elemType, objs)
	diags.Append(d...)
	return list, diags
}

func flattenEnvironmentOptionsFramework(opts *releaseapi.EnvironmentOptions) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: environmentOptionsAttrTypes}
	if opts == nil {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}
	//nolint:staticcheck // deprecated ADO fields preserved for SDK v2 parity
	emailType := ""
	if opts.EmailNotificationType != nil { //nolint:staticcheck
		emailType = *opts.EmailNotificationType //nolint:staticcheck
	}
	emailRecipients := ""
	if opts.EmailRecipients != nil { //nolint:staticcheck
		emailRecipients = *opts.EmailRecipients //nolint:staticcheck
	}
	timeoutMins := int64(0)
	if opts.TimeoutInMinutes != nil { //nolint:staticcheck
		timeoutMins = int64(*opts.TimeoutInMinutes) //nolint:staticcheck
	}
	obj, diags := types.ObjectValue(environmentOptionsAttrTypes, map[string]attr.Value{
		"email_notification_type":                              types.StringValue(emailType),
		"email_recipients":                                     types.StringValue(emailRecipients),
		"skip_artifacts_download":                              types.BoolValue(converter.ToBool(opts.SkipArtifactsDownload, false)), //nolint:staticcheck
		"timeout_in_minutes":                                   types.Int64Value(timeoutMins),
		"enable_access_token":                                  types.BoolValue(converter.ToBool(opts.EnableAccessToken, false)), //nolint:staticcheck
		"publish_deployment_status":                            types.BoolValue(converter.ToBool(opts.PublishDeploymentStatus, false)),
		"badge_enabled":                                        types.BoolValue(converter.ToBool(opts.BadgeEnabled, false)),
		"auto_link_workitems":                                  types.BoolValue(converter.ToBool(opts.AutoLinkWorkItems, false)),
		"pull_request_deployment_enabled":                      types.BoolValue(converter.ToBool(opts.PullRequestDeploymentEnabled, false)),
		"publish_deployment_status_to_devops_project_settings": types.BoolValue(false),
	})
	if diags.HasError() {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}
	list, d := types.ListValue(elemType, []attr.Value{obj})
	diags.Append(d...)
	return list, diags
}

func flattenRetentionPolicyFramework(rp *releaseapi.EnvironmentRetentionPolicy) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: retentionPolicyAttrTypes}
	if rp == nil {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}
	dtk := int64(30)
	if rp.DaysToKeep != nil {
		dtk = int64(*rp.DaysToKeep)
	}
	rtk := int64(3)
	if rp.ReleasesToKeep != nil {
		rtk = int64(*rp.ReleasesToKeep)
	}
	obj, diags := types.ObjectValue(retentionPolicyAttrTypes, map[string]attr.Value{
		"days_to_keep":     types.Int64Value(dtk),
		"releases_to_keep": types.Int64Value(rtk),
		"retain_build":     types.BoolValue(converter.ToBool(rp.RetainBuild, true)),
	})
	if diags.HasError() {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}
	list, d := types.ListValue(elemType, []attr.Value{obj})
	diags.Append(d...)
	return list, diags
}

func flattenApprovalsFramework(approvals *releaseapi.ReleaseDefinitionApprovals) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: approvalAttrTypes}
	if approvals == nil {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}

	var diags diag.Diagnostics

	// Approvers
	var approverObjs []attr.Value
	if approvals.Approvals != nil {
		for _, step := range *approvals.Approvals {
			approverID := ""
			if step.Approver != nil && step.Approver.Id != nil {
				approverID = *step.Approver.Id
			}
			// ADO omits the Approver identity for automated approval steps.
			// Represent them consistently with the null UUID so that the plan
			// config value ("00000000-0000-0000-0000-000000000000") round-trips
			// without a perpetual diff.
			if approverID == "" && converter.ToBool(step.IsAutomated, false) {
				approverID = "00000000-0000-0000-0000-000000000000"
			}
			rank := int64(1)
			if step.Rank != nil {
				rank = int64(*step.Rank)
			}
			obj, d := types.ObjectValue(approverAttrTypes, map[string]attr.Value{
				"id":           types.StringValue(approverID),
				"is_automated": types.BoolValue(converter.ToBool(step.IsAutomated, false)),
				"rank":         types.Int64Value(rank),
			})
			diags.Append(d...)
			approverObjs = append(approverObjs, obj)
		}
	}
	approverList, d := types.ListValue(types.ObjectType{AttrTypes: approverAttrTypes}, approverObjs)
	diags.Append(d...)

	// ApprovalOptions
	var optObjs []attr.Value
	if approvals.ApprovalOptions != nil {
		o := approvals.ApprovalOptions
		reqCount := int64(0)
		if o.RequiredApproverCount != nil {
			reqCount = int64(*o.RequiredApproverCount)
		}
		timeoutMin := int64(0)
		if o.TimeoutInMinutes != nil {
			timeoutMin = int64(*o.TimeoutInMinutes)
		}
		execOrder := "beforeGates"
		if o.ExecutionOrder != nil {
			execOrder = string(*o.ExecutionOrder)
		}
		obj, d := types.ObjectValue(approvalOptionsAttrTypes, map[string]attr.Value{
			"required_approver_count":         types.Int64Value(reqCount),
			"release_creator_can_be_approver": types.BoolValue(converter.ToBool(o.ReleaseCreatorCanBeApprover, false)),
			"enforce_identity_revalidation":   types.BoolValue(converter.ToBool(o.EnforceIdentityRevalidation, false)),
			"timeout_in_minutes":              types.Int64Value(timeoutMin),
			"execution_order":                 types.StringValue(execOrder),
			"auto_triggered_and_previous_environment_approved_can_be_skipped": types.BoolValue(converter.ToBool(o.AutoTriggeredAndPreviousEnvironmentApprovedCanBeSkipped, false)),
		})
		diags.Append(d...)
		optObjs = append(optObjs, obj)
	}
	optList, d := types.ListValue(types.ObjectType{AttrTypes: approvalOptionsAttrTypes}, optObjs)
	diags.Append(d...)

	obj, d := types.ObjectValue(approvalAttrTypes, map[string]attr.Value{
		"approver":         approverList,
		"approval_options": optList,
	})
	diags.Append(d...)

	list, d := types.ListValue(elemType, []attr.Value{obj})
	diags.Append(d...)
	return list, diags
}

func flattenDeploymentGatesFramework(step *releaseapi.ReleaseDefinitionGatesStep) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: deploymentGatesAttrTypes}
	if step == nil {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}

	var diags diag.Diagnostics

	// GatesOptions
	gatesOptElemType := types.ObjectType{AttrTypes: gatesOptionsAttrTypes}
	var optObjs []attr.Value
	if step.GatesOptions != nil {
		o := step.GatesOptions
		timeout := int64(0)
		if o.Timeout != nil {
			timeout = int64(*o.Timeout)
		}
		sampling := int64(0)
		if o.SamplingInterval != nil {
			sampling = int64(*o.SamplingInterval)
		}
		stabilization := int64(0)
		if o.StabilizationTime != nil {
			stabilization = int64(*o.StabilizationTime)
		}
		minSuccess := int64(0)
		if o.MinimumSuccessDuration != nil {
			minSuccess = int64(*o.MinimumSuccessDuration)
		}
		obj, d := types.ObjectValue(gatesOptionsAttrTypes, map[string]attr.Value{
			"is_enabled":               types.BoolValue(converter.ToBool(o.IsEnabled, false)),
			"timeout":                  types.Int64Value(timeout),
			"sampling_interval":        types.Int64Value(sampling),
			"stabilization_time":       types.Int64Value(stabilization),
			"minimum_success_duration": types.Int64Value(minSuccess),
		})
		diags.Append(d...)
		optObjs = append(optObjs, obj)
	}
	optList, d := types.ListValue(gatesOptElemType, optObjs)
	diags.Append(d...)

	// Gates
	gateElemType := types.ObjectType{AttrTypes: gateAttrTypes}
	var gateObjs []attr.Value
	if step.Gates != nil {
		for _, g := range *step.Gates {
			gateTaskList, d := flattenGateTasksFramework(g.Tasks)
			diags.Append(d...)
			gateObj, d := types.ObjectValue(gateAttrTypes, map[string]attr.Value{
				"task": gateTaskList,
			})
			diags.Append(d...)
			gateObjs = append(gateObjs, gateObj)
		}
	}
	gateList, d := types.ListValue(gateElemType, gateObjs)
	diags.Append(d...)

	obj, d := types.ObjectValue(deploymentGatesAttrTypes, map[string]attr.Value{
		"gates_options": optList,
		"gate":          gateList,
	})
	diags.Append(d...)

	list, d := types.ListValue(elemType, []attr.Value{obj})
	diags.Append(d...)
	return list, diags
}

// flattenGateTasksFramework converts *[]releaseapi.WorkflowTask (from a gate) to types.List.
func flattenGateTasksFramework(tasks *[]releaseapi.WorkflowTask) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: gateTaskAttrTypes}
	if tasks == nil || len(*tasks) == 0 {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}
	var diags diag.Diagnostics
	taskObjs := make([]attr.Value, 0, len(*tasks))
	for _, t := range *tasks {
		taskID := ""
		if t.TaskId != nil {
			taskID = t.TaskId.String()
		}
		displayName := ""
		if t.Name != nil {
			displayName = *t.Name
		}
		versionSpec := "0.*"
		if t.Version != nil {
			versionSpec = *t.Version
		}
		enabled := true
		if t.Enabled != nil {
			enabled = *t.Enabled
		}
		alwaysRun := false
		if t.AlwaysRun != nil {
			alwaysRun = *t.AlwaysRun
		}
		condition := "succeeded()"
		if t.Condition != nil {
			condition = *t.Condition
		}
		timeout := int64(0)
		if t.TimeoutInMinutes != nil {
			timeout = int64(*t.TimeoutInMinutes)
		}
		retryCount := int64(0)
		if t.RetryCountOnTaskFailure != nil {
			retryCount = int64(*t.RetryCountOnTaskFailure)
		}
		definitionType := "task"
		if t.DefinitionType != nil && *t.DefinitionType != "" {
			definitionType = *t.DefinitionType
		}
		inputsMap := map[string]attr.Value{}
		if t.Inputs != nil {
			for k, v := range *t.Inputs {
				inputsMap[k] = types.StringValue(v)
			}
		}
		inputsTF, d := types.MapValue(types.StringType, inputsMap)
		diags.Append(d...)

		taskObj, d := types.ObjectValue(gateTaskAttrTypes, map[string]attr.Value{
			"task_id":                     types.StringValue(taskID),
			"display_name":                types.StringValue(displayName),
			"version_spec":                types.StringValue(versionSpec),
			"enabled":                     types.BoolValue(enabled),
			"always_run":                  types.BoolValue(alwaysRun),
			"condition":                   types.StringValue(condition),
			"timeout_in_minutes":          types.Int64Value(timeout),
			"retry_count_on_task_failure": types.Int64Value(retryCount),
			"inputs":                      inputsTF,
			"definition_type":             types.StringValue(definitionType),
		})
		diags.Append(d...)
		taskObjs = append(taskObjs, taskObj)
	}
	taskList, d := types.ListValue(elemType, taskObjs)
	diags.Append(d...)
	return taskList, diags
}

// flattenExecutionPolicyFramework converts EnvironmentExecutionPolicy to a types.List.
func flattenExecutionPolicyFramework(ep *releaseapi.EnvironmentExecutionPolicy) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: executionPolicyAttrTypes}
	if ep == nil {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}
	cc := int64(0)
	if ep.ConcurrencyCount != nil {
		cc = int64(*ep.ConcurrencyCount)
	}
	qd := int64(0)
	if ep.QueueDepthCount != nil {
		qd = int64(*ep.QueueDepthCount)
	}
	obj, diags := types.ObjectValue(executionPolicyAttrTypes, map[string]attr.Value{
		"concurrency_count": types.Int64Value(cc),
		"queue_depth_count": types.Int64Value(qd),
	})
	if diags.HasError() {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}
	list, d := types.ListValue(elemType, []attr.Value{obj})
	diags.Append(d...)
	return list, diags
}

// flattenEnvironmentTriggersFramework converts []EnvironmentTrigger to a types.List.
func flattenEnvironmentTriggersFramework(triggers *[]releaseapi.EnvironmentTrigger) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: environmentTriggerAttrTypes}
	if triggers == nil || len(*triggers) == 0 {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}
	var diags diag.Diagnostics
	objs := make([]attr.Value, 0, len(*triggers))
	for _, t := range *triggers {
		defEnvID := int64(0)
		if t.DefinitionEnvironmentId != nil {
			defEnvID = int64(*t.DefinitionEnvironmentId)
		}
		trigType := ""
		if t.TriggerType != nil {
			trigType = string(*t.TriggerType)
		}
		trigContent := ""
		if t.TriggerContent != nil {
			trigContent = *t.TriggerContent
		}
		obj, d := types.ObjectValue(environmentTriggerAttrTypes, map[string]attr.Value{
			"definition_environment_id": types.Int64Value(defEnvID),
			"trigger_type":              types.StringValue(trigType),
			"trigger_content":           types.StringValue(trigContent),
		})
		diags.Append(d...)
		objs = append(objs, obj)
	}
	list, d := types.ListValue(elemType, objs)
	diags.Append(d...)
	return list, diags
}

// flattenStageSchedulesFramework converts []ReleaseSchedule to a types.List.
func flattenStageSchedulesFramework(scheds *[]releaseapi.ReleaseSchedule) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: stageScheduleAttrTypes}
	if scheds == nil || len(*scheds) == 0 {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}
	var diags diag.Diagnostics
	objs := make([]attr.Value, 0, len(*scheds))
	for _, s := range *scheds {
		days := int64(127)
		if s.DaysToRelease != nil {
			if d, err := strconv.Atoi(string(*s.DaysToRelease)); err == nil {
				days = int64(d)
			}
		}
		startH := int64(0)
		if s.StartHours != nil {
			startH = int64(*s.StartHours)
		}
		startM := int64(0)
		if s.StartMinutes != nil {
			startM = int64(*s.StartMinutes)
		}
		tz := "UTC"
		if s.TimeZoneId != nil {
			tz = *s.TimeZoneId
		}
		jobID := ""
		if s.JobId != nil {
			jobID = s.JobId.String()
		}
		obj, d := types.ObjectValue(stageScheduleAttrTypes, map[string]attr.Value{
			"days_to_release": types.Int64Value(days),
			"start_hours":     types.Int64Value(startH),
			"start_minutes":   types.Int64Value(startM),
			"time_zone_id":    types.StringValue(tz),
			"job_id":          types.StringValue(jobID),
		})
		diags.Append(d...)
		objs = append(objs, obj)
	}
	list, d := types.ListValue(elemType, objs)
	diags.Append(d...)
	return list, diags
}

// flattenProcessParametersFramework converts ProcessParameters to a types.List.
func flattenProcessParametersFramework(pp *distributedtaskcommon.ProcessParameters) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: processParametersAttrTypes}
	if pp == nil {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}
	var diags diag.Diagnostics
	inputElemType := types.ObjectType{AttrTypes: processParameterInputAttrTypes}
	var inputObjs []attr.Value
	if pp.Inputs != nil {
		inputObjs = make([]attr.Value, 0, len(*pp.Inputs))
		for _, inp := range *pp.Inputs {
			name := ""
			if inp.Name != nil {
				name = *inp.Name
			}
			dv := ""
			if inp.DefaultValue != nil {
				dv = *inp.DefaultValue
			}
			pt := ""
			if inp.Type != nil {
				pt = *inp.Type
			}
			iObj, d := types.ObjectValue(processParameterInputAttrTypes, map[string]attr.Value{
				"name":           types.StringValue(name),
				"default_value":  types.StringValue(dv),
				"parameter_type": types.StringValue(pt),
			})
			diags.Append(d...)
			inputObjs = append(inputObjs, iObj)
		}
	}
	inputList, d := types.ListValue(inputElemType, inputObjs)
	diags.Append(d...)

	obj, d := types.ObjectValue(processParametersAttrTypes, map[string]attr.Value{
		"input": inputList,
	})
	diags.Append(d...)
	list, d := types.ListValue(elemType, []attr.Value{obj})
	diags.Append(d...)
	return list, diags
}

// flattenStagePropertiesFramework unwraps ADO's {"$type":"System.String","$value":"..."}
// wrapper from stage properties and returns a types.Map of string values.
func flattenStagePropertiesFramework(properties interface{}) (types.Map, diag.Diagnostics) {
	if properties == nil {
		return types.MapValueMust(types.StringType, map[string]attr.Value{}), nil
	}
	props, ok := properties.(map[string]interface{})
	if !ok || len(props) == 0 {
		return types.MapValueMust(types.StringType, map[string]attr.Value{}), nil
	}
	attrs := make(map[string]attr.Value, len(props))
	for k, v := range props {
		switch cv := v.(type) {
		case string:
			attrs[k] = types.StringValue(cv)
		case map[string]interface{}:
			// Unwrap ADO's {"$type":"System.String","$value":"..."} wrapper.
			if val, ok := cv["$value"].(string); ok {
				attrs[k] = types.StringValue(val)
			} else {
				attrs[k] = types.StringValue(fmt.Sprintf("%v", v))
			}
		default:
			attrs[k] = types.StringValue(fmt.Sprintf("%v", v))
		}
	}
	return types.MapValue(types.StringType, attrs)
}

// -------------------------------------------------------------------------
// Flatten: deploy_phase
// -------------------------------------------------------------------------

// flattenDeployPhasesFramework converts the ADO API deploy phases ([]interface{}) into
// a framework types.List of deploy_phase objects.
func flattenDeployPhasesFramework(ctx context.Context, phases *[]interface{}) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: deployPhaseAttrTypes}
	if phases == nil || len(*phases) == 0 {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}

	var diags diag.Diagnostics
	phaseObjs := make([]attr.Value, 0, len(*phases))

	for _, phase := range *phases {
		// ADO API returns deploy phases as generic interface{} — marshal/unmarshal to normalise.
		phaseData, err := json.Marshal(phase)
		if err != nil {
			diags.AddError("Deploy phase marshal error", err.Error())
			continue
		}
		var phaseMap map[string]interface{}
		if err := json.Unmarshal(phaseData, &phaseMap); err != nil {
			diags.AddError("Deploy phase unmarshal error", err.Error())
			continue
		}

		name := ""
		if v, ok := phaseMap["name"].(string); ok {
			name = v
		}
		rank := int64(1)
		if v, ok := phaseMap["rank"].(float64); ok {
			rank = int64(v)
		}
		phaseType := "agentBasedDeployment"
		if v, ok := phaseMap["phaseType"].(string); ok {
			phaseType = v
		}

		// deployment_input
		deploymentInput, d := flattenDeploymentInputFramework(ctx, phaseMap["deploymentInput"])
		diags.Append(d...)

		// workflow_task (workflowTasks in ADO API)
		workflowTasks, d := flattenWorkflowTasksFramework(ctx, phaseMap["workflowTasks"])
		diags.Append(d...)

		phaseObj, d := types.ObjectValue(deployPhaseAttrTypes, map[string]attr.Value{
			"name":             types.StringValue(name),
			"rank":             types.Int64Value(rank),
			"phase_type":       types.StringValue(phaseType),
			"deployment_input": deploymentInput,
			"workflow_task":    workflowTasks,
		})
		diags.Append(d...)
		phaseObjs = append(phaseObjs, phaseObj)
	}

	list, d := types.ListValue(elemType, phaseObjs)
	diags.Append(d...)
	return list, diags
}

// flattenDeploymentInputFramework converts the ADO API deploymentInput map to a framework types.List.
func flattenDeploymentInputFramework(_ context.Context, raw interface{}) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: deploymentInputAttrTypes}
	if raw == nil {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}

	diMap, ok := raw.(map[string]interface{})
	if !ok {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}

	queueID := int64(0)
	if v, ok := diMap["queueId"].(float64); ok {
		queueID = int64(v)
	}
	condition := "succeeded()"
	if v, ok := diMap["condition"].(string); ok {
		condition = v
	}
	timeout := int64(0)
	if v, ok := diMap["timeoutInMinutes"].(float64); ok {
		timeout = int64(v)
	}
	jobCancelTimeout := int64(1)
	if v, ok := diMap["jobCancelTimeoutInMinutes"].(float64); ok {
		jobCancelTimeout = int64(v)
	}
	skipDownload := false
	if v, ok := diMap["skipArtifactsDownload"].(bool); ok {
		skipDownload = v
	}
	enableToken := false
	if v, ok := diMap["enableAccessToken"].(bool); ok {
		enableToken = v
	}
	agentSpec := ""
	if v, ok := diMap["agentSpecification"].(map[string]interface{}); ok {
		if id, ok := v["identifier"].(string); ok {
			agentSpec = id
		}
	}

	// override_inputs
	var diags diag.Diagnostics
	overrideInputsMap := map[string]attr.Value{}
	if oi, ok := diMap["overrideInputs"].(map[string]interface{}); ok {
		for k, v := range oi {
			if s, ok := v.(string); ok {
				overrideInputsMap[k] = types.StringValue(s)
			}
		}
	}
	overrideInputsTF, d := types.MapValue(types.StringType, overrideInputsMap)
	diags.Append(d...)

	// demands
	var demandStrs []attr.Value
	if demands, ok := diMap["demands"].([]interface{}); ok {
		for _, d := range demands {
			if s, ok := d.(string); ok {
				demandStrs = append(demandStrs, types.StringValue(s))
			}
		}
	}
	demandsList, d := types.ListValue(types.StringType, demandStrs)
	diags.Append(d...)

	// parallel_execution
	parallelExec, d := flattenParallelExecutionFramework(diMap["parallelExecution"])
	diags.Append(d...)

	diObj, d := types.ObjectValue(deploymentInputAttrTypes, map[string]attr.Value{
		"queue_id":                      types.Int64Value(queueID),
		"condition":                     types.StringValue(condition),
		"timeout_in_minutes":            types.Int64Value(timeout),
		"job_cancel_timeout_in_minutes": types.Int64Value(jobCancelTimeout),
		"skip_artifacts_download":       types.BoolValue(skipDownload),
		"enable_access_token":           types.BoolValue(enableToken),
		"agent_specification":           types.StringValue(agentSpec),
		"override_inputs":               overrideInputsTF,
		"demands":                       demandsList,
		"parallel_execution":            parallelExec,
	})
	diags.Append(d...)

	diList, d := types.ListValue(elemType, []attr.Value{diObj})
	diags.Append(d...)
	return diList, diags
}

// flattenParallelExecutionFramework converts the ADO parallelExecution map to a framework types.List.
func flattenParallelExecutionFramework(raw interface{}) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: parallelExecutionAttrTypes}
	var diags diag.Diagnostics

	if raw == nil {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}
	peMap, ok := raw.(map[string]interface{})
	if !ok {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}

	peType := ""
	if v, ok := peMap["parallelExecutionType"].(string); ok {
		peType = v
	}
	// Suppress "none" type — emit empty list so state matches HCL's absence.
	if peType == "none" || peType == "" {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}

	maxAgents := int64(1)
	if v, ok := peMap["maxNumberOfAgents"].(float64); ok {
		maxAgents = int64(v)
	}

	// ADO stores multipliers as comma-joined string.
	var multiplierVals []attr.Value
	switch v := peMap["multipliers"].(type) {
	case string:
		for _, part := range rdSplitComma(v) {
			if part != "" {
				multiplierVals = append(multiplierVals, types.StringValue(part))
			}
		}
	case []interface{}:
		for _, m := range v {
			if s, ok := m.(string); ok {
				multiplierVals = append(multiplierVals, types.StringValue(s))
			}
		}
	}
	multipliersList, d := types.ListValue(types.StringType, multiplierVals)
	diags.Append(d...)

	peObj, d := types.ObjectValue(parallelExecutionAttrTypes, map[string]attr.Value{
		"type":                 types.StringValue(peType),
		"max_number_of_agents": types.Int64Value(maxAgents),
		"multipliers":          multipliersList,
	})
	diags.Append(d...)

	peList, d := types.ListValue(elemType, []attr.Value{peObj})
	diags.Append(d...)
	return peList, diags
}

// flattenWorkflowTasksFramework converts ADO workflowTasks (raw interface{} slice) to a framework types.List.
func flattenWorkflowTasksFramework(_ context.Context, raw interface{}) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: workflowTaskAttrTypes}
	var diags diag.Diagnostics

	if raw == nil {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}

	tasksRaw, ok := raw.([]interface{})
	if !ok {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}

	taskObjs := make([]attr.Value, 0, len(tasksRaw))
	for _, t := range tasksRaw {
		taskMap, ok := t.(map[string]interface{})
		if !ok {
			continue
		}

		taskID := ""
		if v, ok := taskMap["taskId"].(string); ok {
			taskID = v
		}
		// Also check nested task.id form (some ADO versions).
		if taskID == "" {
			if taskRef, ok := taskMap["task"].(map[string]interface{}); ok {
				if id, ok := taskRef["id"].(string); ok {
					taskID = id
				}
			}
		}

		displayName := ""
		if v, ok := taskMap["name"].(string); ok {
			displayName = v
		}
		versionSpec := "0.*"
		if v, ok := taskMap["version"].(string); ok {
			versionSpec = v
		}
		enabled := true
		if v, ok := taskMap["enabled"].(bool); ok {
			enabled = v
		}
		alwaysRun := false
		if v, ok := taskMap["alwaysRun"].(bool); ok {
			alwaysRun = v
		}
		condition := "succeeded()"
		if v, ok := taskMap["condition"].(string); ok {
			condition = v
		}
		timeout := int64(0)
		if v, ok := taskMap["timeoutInMinutes"].(float64); ok {
			timeout = int64(v)
		}
		retryCount := int64(0)
		if v, ok := taskMap["retryCountOnTaskFailure"].(float64); ok {
			retryCount = int64(v)
		}

		// definition_type
		definitionType := "task"
		if v, ok := taskMap["definitionType"].(string); ok && v != "" {
			definitionType = v
		}

		// inputs map
		inputsMap := map[string]attr.Value{}
		if inputs, ok := taskMap["inputs"].(map[string]interface{}); ok {
			for k, v := range inputs {
				if s, ok := v.(string); ok {
					inputsMap[k] = types.StringValue(s)
				}
			}
		}
		inputsTF, d := types.MapValue(types.StringType, inputsMap)
		diags.Append(d...)

		taskObj, d := types.ObjectValue(workflowTaskAttrTypes, map[string]attr.Value{
			"task_id":                     types.StringValue(taskID),
			"display_name":                types.StringValue(displayName),
			"version_spec":                types.StringValue(versionSpec),
			"enabled":                     types.BoolValue(enabled),
			"always_run":                  types.BoolValue(alwaysRun),
			"condition":                   types.StringValue(condition),
			"timeout_in_minutes":          types.Int64Value(timeout),
			"retry_count_on_task_failure": types.Int64Value(retryCount),
			"inputs":                      inputsTF,
			"definition_type":             types.StringValue(definitionType),
		})
		diags.Append(d...)
		taskObjs = append(taskObjs, taskObj)
	}

	taskList, d := types.ListValue(elemType, taskObjs)
	diags.Append(d...)
	return taskList, diags
}

// rdSplitComma splits a comma-separated string into trimmed parts.
func rdSplitComma(s string) []string {
	if s == "" {
		return nil
	}
	var parts []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			part := s[start:i]
			// trim spaces
			for len(part) > 0 && part[0] == ' ' {
				part = part[1:]
			}
			for len(part) > 0 && part[len(part)-1] == ' ' {
				part = part[:len(part)-1]
			}
			parts = append(parts, part)
			start = i + 1
		}
	}
	return parts
}

// -------------------------------------------------------------------------
// Variables expand/flatten
// -------------------------------------------------------------------------

// expandVariablesFramework converts a types.Map (MapNestedAttribute) of variableValueModel into
// *map[string]releaseapi.ConfigurationVariableValue.
func expandVariablesFramework(ctx context.Context, m types.Map) (*map[string]releaseapi.ConfigurationVariableValue, error) {
	if m.IsNull() || m.IsUnknown() {
		return nil, nil
	}
	var models map[string]variableValueModel
	if d := m.ElementsAs(ctx, &models, false); d.HasError() {
		return nil, fmt.Errorf("decoding variables: %s", d)
	}
	result := make(map[string]releaseapi.ConfigurationVariableValue, len(models))
	for name, v := range models {
		result[name] = releaseapi.ConfigurationVariableValue{
			Value:         converter.String(v.Value.ValueString()),
			IsSecret:      converter.Bool(v.IsSecret.ValueBool()),
			AllowOverride: converter.Bool(v.AllowOverride.ValueBool()),
		}
	}
	return &result, nil
}

// flattenVariablesFramework converts *map[string]releaseapi.ConfigurationVariableValue into a
// types.Map (MapNestedAttribute).
//
// existingVars is the prior/plan-time map (before being overwritten by this call).
// ADO never returns secret variable values; when is_secret=true and the API returns
// an empty value, we preserve the existing value from existingVars so the provider
// does not produce an inconsistent result (was <secret>, now "").
func flattenVariablesFramework(ctx context.Context, vars *map[string]releaseapi.ConfigurationVariableValue, existingVars types.Map) (types.Map, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: variableValueAttrTypes}
	if vars == nil || len(*vars) == 0 {
		empty, d := types.MapValue(elemType, map[string]attr.Value{})
		return empty, d
	}

	// Build a lookup of existing variable values keyed by variable name.
	existingValsByName := map[string]string{}
	if !existingVars.IsNull() && !existingVars.IsUnknown() {
		for k, v := range existingVars.Elements() {
			obj, ok := v.(types.Object)
			if !ok {
				continue
			}
			if valAttr, ok := obj.Attributes()["value"].(types.String); ok && !valAttr.IsNull() && !valAttr.IsUnknown() {
				existingValsByName[k] = valAttr.ValueString()
			}
		}
	}

	var diags diag.Diagnostics
	objs := make(map[string]attr.Value, len(*vars))
	for name, v := range *vars {
		apiVal := ""
		if v.Value != nil {
			apiVal = *v.Value
		}
		isSecret := false
		if v.IsSecret != nil {
			isSecret = *v.IsSecret
		}
		allowOverride := false
		if v.AllowOverride != nil {
			allowOverride = *v.AllowOverride
		}
		// The variable `value` attribute is unconditionally Sensitive (write-only):
		// the configured/planned value is the source of truth and the framework
		// requires the post-apply value to match the plan EXACTLY. ADO does not
		// faithfully echo variable values (it returns an empty value for secrets,
		// and may normalise non-secret values), so always preserve the prior/plan
		// value when this variable was present in config/state. Fall back to the
		// API value only when there is no prior value (e.g. import).
		val := apiVal
		if existingVal, ok := existingValsByName[name]; ok {
			val = existingVal
		}
		obj, d := types.ObjectValue(variableValueAttrTypes, map[string]attr.Value{
			"value":          types.StringValue(val),
			"is_secret":      types.BoolValue(isSecret),
			"allow_override": types.BoolValue(allowOverride),
		})
		diags.Append(d...)
		objs[name] = obj
	}
	result, d := types.MapValue(elemType, objs)
	diags.Append(d...)
	return result, diags
}

// expandVariableGroupsFramework converts types.List of Int64 into []int.
func expandVariableGroupsFramework(ctx context.Context, list types.List) ([]int, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var vals []types.Int64
	if d := list.ElementsAs(ctx, &vals, false); d.HasError() {
		return nil, fmt.Errorf("decoding variable_groups: %s", d)
	}
	result := make([]int, len(vals))
	for i, v := range vals {
		result[i] = int(v.ValueInt64())
	}
	return result, nil
}

// flattenVariableGroupsFramework converts *[]int into types.List.
func flattenVariableGroupsFramework(groups *[]int) (types.List, diag.Diagnostics) {
	if groups == nil || len(*groups) == 0 {
		return types.ListValueMust(types.Int64Type, []attr.Value{}), nil
	}
	vals := make([]attr.Value, len(*groups))
	for i, g := range *groups {
		vals[i] = types.Int64Value(int64(g))
	}
	return types.ListValue(types.Int64Type, vals)
}

// -------------------------------------------------------------------------
// Artifacts expand/flatten
// -------------------------------------------------------------------------

// apiComputedArtifactKeys lists keys the ADO API injects that should be stripped on read-back.
var apiComputedArtifactKeys = map[string]bool{
	"artifactSourceDefinitionUrl": true,
	"defaultVersionType":          true,
	"defaultVersionBranch":        true,
	"defaultVersionSpecific":      true,
	"defaultVersionTags":          true,
}

// expandArtifactsFramework converts types.List (ListNestedAttribute of artifactModel) into
// []releaseapi.Artifact.
func expandArtifactsFramework(ctx context.Context, list types.List) ([]releaseapi.Artifact, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []artifactModel
	if d := list.ElementsAs(ctx, &models, false); d.HasError() {
		return nil, fmt.Errorf("decoding artifacts: %s", d)
	}
	result := make([]releaseapi.Artifact, 0, len(models))
	for _, m := range models {
		a := releaseapi.Artifact{
			Alias:     converter.String(m.Alias.ValueString()),
			Type:      converter.String(m.Type.ValueString()),
			IsPrimary: converter.Bool(m.IsPrimary.ValueBool()),
		}
		if !m.DefinitionReference.IsNull() && !m.DefinitionReference.IsUnknown() {
			var defRefMap map[string]string
			if d := m.DefinitionReference.ElementsAs(ctx, &defRefMap, false); d.HasError() {
				return nil, fmt.Errorf("decoding definition_reference: %s", d)
			}
			refs := make(map[string]releaseapi.ArtifactSourceReference, len(defRefMap))
			for k, v := range defRefMap {
				vCopy := v
				refs[k] = releaseapi.ArtifactSourceReference{Id: &vCopy}
			}
			a.DefinitionReference = &refs
		}
		result = append(result, a)
	}
	return result, nil
}

// flattenArtifactsFramework converts *[]releaseapi.Artifact into types.List, filtering API-computed
// definition_reference keys. The user-configured keys are preserved as-is.
func flattenArtifactsFramework(ctx context.Context, artifacts *[]releaseapi.Artifact) (types.List, diag.Diagnostics) {
	_ = ctx
	elemType := types.ObjectType{AttrTypes: artifactAttrTypes}
	if artifacts == nil || len(*artifacts) == 0 {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}
	var diags diag.Diagnostics
	objs := make([]attr.Value, 0, len(*artifacts))
	for _, a := range *artifacts {
		alias := ""
		if a.Alias != nil {
			alias = *a.Alias
		}
		artType := "Build"
		if a.Type != nil {
			artType = *a.Type
		}
		isPrimary := false
		if a.IsPrimary != nil {
			isPrimary = *a.IsPrimary
		}
		defRefVals := map[string]attr.Value{}
		if a.DefinitionReference != nil {
			for k, v := range *a.DefinitionReference {
				if apiComputedArtifactKeys[k] {
					continue
				}
				id := ""
				if v.Id != nil {
					id = *v.Id
				}
				defRefVals[k] = types.StringValue(id)
			}
		}
		defRef, d := types.MapValue(types.StringType, defRefVals)
		diags.Append(d...)

		obj, d := types.ObjectValue(artifactAttrTypes, map[string]attr.Value{
			"alias":                types.StringValue(alias),
			"type":                 types.StringValue(artType),
			"is_primary":           types.BoolValue(isPrimary),
			"definition_reference": defRef,
		})
		diags.Append(d...)
		objs = append(objs, obj)
	}
	list, d := types.ListValue(elemType, objs)
	diags.Append(d...)
	return list, diags
}

// -------------------------------------------------------------------------
// Triggers expand/flatten
// -------------------------------------------------------------------------

// expandTriggersFramework converts types.List (ListNestedAttribute of triggerModel) into
// []interface{} that ADO's polymorphic Triggers field accepts.
func expandTriggersFramework(ctx context.Context, list types.List) ([]interface{}, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []triggerModel
	if d := list.ElementsAs(ctx, &models, false); d.HasError() {
		return nil, fmt.Errorf("decoding triggers: %s", d)
	}
	result := make([]interface{}, 0)
	for _, t := range models {
		// CD artifact triggers
		if !t.CdArtifactTrigger.IsNull() && !t.CdArtifactTrigger.IsUnknown() {
			var cdModels []cdArtifactTriggerModel
			if d := t.CdArtifactTrigger.ElementsAs(ctx, &cdModels, false); d.HasError() {
				return nil, fmt.Errorf("decoding cd_artifact_trigger: %s", d)
			}
			for _, cd := range cdModels {
				trigEntry := map[string]interface{}{
					"triggerType":   "artifactSource",
					"artifactAlias": cd.ArtifactAlias.ValueString(),
				}
				trigEntry["triggerConditions"] = expandCdArtifactTriggerConditionsFramework(ctx, cd)
				result = append(result, trigEntry)
			}
		}
		// Schedule triggers
		if !t.ScheduleTrigger.IsNull() && !t.ScheduleTrigger.IsUnknown() {
			var schModels []scheduleTriggerModel
			if d := t.ScheduleTrigger.ElementsAs(ctx, &schModels, false); d.HasError() {
				return nil, fmt.Errorf("decoding schedule_trigger: %s", d)
			}
			for _, sch := range schModels {
				schedule := map[string]interface{}{
					"scheduleOnlyWithChanges": sch.ScheduleOnlyWithChanges.ValueBool(),
					"startHours":              int(sch.StartHours.ValueInt64()),
					"startMinutes":            int(sch.StartMinutes.ValueInt64()),
					"timeZoneId":              sch.TimeZoneID.ValueString(),
					"daysToRelease":           int(sch.DaysToRelease.ValueInt64()),
				}
				result = append(result, map[string]interface{}{
					"triggerType": "schedule",
					"schedule":    schedule,
				})
			}
		}
		// Source repo triggers
		if !t.SourceRepoTrigger.IsNull() && !t.SourceRepoTrigger.IsUnknown() {
			var srtModels []sourceRepoTriggerModel
			if d := t.SourceRepoTrigger.ElementsAs(ctx, &srtModels, false); d.HasError() {
				return nil, fmt.Errorf("decoding source_repo_trigger: %s", d)
			}
			for _, srt := range srtModels {
				trigEntry := map[string]interface{}{
					"triggerType": "sourceRepo",
					"alias":       srt.Alias.ValueString(),
				}
				if !srt.BranchFilters.IsNull() && !srt.BranchFilters.IsUnknown() {
					var bfs []string
					if d := srt.BranchFilters.ElementsAs(ctx, &bfs, false); d.HasError() {
						return nil, fmt.Errorf("decoding source_repo_trigger branch_filters: %s", d)
					}
					if len(bfs) > 0 {
						trigEntry["branchFilters"] = bfs
					}
				}
				result = append(result, trigEntry)
			}
		}
		// Container image triggers
		if !t.ContainerImageTrigger.IsNull() && !t.ContainerImageTrigger.IsUnknown() {
			var ciModels []containerImageTriggerModel
			if d := t.ContainerImageTrigger.ElementsAs(ctx, &ciModels, false); d.HasError() {
				return nil, fmt.Errorf("decoding container_image_trigger: %s", d)
			}
			for _, ci := range ciModels {
				trigEntry := map[string]interface{}{
					"triggerType": "containerImage",
					"alias":       ci.ArtifactAlias.ValueString(),
				}
				if label := ci.Label.ValueString(); label != "" {
					trigEntry["tagFilters"] = []interface{}{
						map[string]interface{}{"pattern": label},
					}
				}
				result = append(result, trigEntry)
			}
		}
		// Pull request triggers
		if !t.PullRequestTrigger.IsNull() && !t.PullRequestTrigger.IsUnknown() {
			var prModels []pullRequestTriggerModel
			if d := t.PullRequestTrigger.ElementsAs(ctx, &prModels, false); d.HasError() {
				return nil, fmt.Errorf("decoding pull_request_trigger: %s", d)
			}
			for _, pr := range prModels {
				trigEntry := map[string]interface{}{
					"triggerType":   "pullRequest",
					"artifactAlias": pr.ArtifactAlias.ValueString(),
				}
				// Each (target_branch, tags) combination maps to one PullRequestFilter condition.
				// We build a single condition per trigger entry carrying both target_branches and tags.
				var targetBranches []string
				if !pr.TargetBranches.IsNull() && !pr.TargetBranches.IsUnknown() {
					if d := pr.TargetBranches.ElementsAs(ctx, &targetBranches, false); d.HasError() {
						return nil, fmt.Errorf("decoding pull_request_trigger target_branches: %s", d)
					}
				}
				var prTags []string
				if !pr.Tags.IsNull() && !pr.Tags.IsUnknown() {
					if d := pr.Tags.ElementsAs(ctx, &prTags, false); d.HasError() {
						return nil, fmt.Errorf("decoding pull_request_trigger tags: %s", d)
					}
				}
				if len(targetBranches) > 0 || len(prTags) > 0 {
					conditions := make([]map[string]interface{}, 0)
					for _, tb := range targetBranches {
						cond := map[string]interface{}{"targetBranch": tb}
						if len(prTags) > 0 {
							cond["tags"] = prTags
						}
						conditions = append(conditions, cond)
					}
					if len(conditions) == 0 && len(prTags) > 0 {
						// tags without target_branches: emit one condition with empty targetBranch
						conditions = append(conditions, map[string]interface{}{
							"targetBranch": "",
							"tags":         prTags,
						})
					}
					trigEntry["triggerConditions"] = conditions
				}
				// ADO requires a non-null repository reference for a pull request
				// trigger. For source-based (Build) artifacts, useArtifactReference
				// tells ADO to derive it from the linked artifact.
				useArtifactRef := true
				if !pr.UseArtifactReference.IsNull() && !pr.UseArtifactReference.IsUnknown() {
					useArtifactRef = pr.UseArtifactReference.ValueBool()
				}
				trigEntry["pullRequestConfiguration"] = map[string]interface{}{
					"useArtifactReference": useArtifactRef,
				}
				result = append(result, trigEntry)
			}
		}
	}
	return result, nil
}

// expandCdArtifactTriggerConditionsFramework builds the triggerConditions slice for a
// cd_artifact_trigger, incorporating branch_filters, tag_filter, use_build_definition_branch,
// and create_release_on_build_tagging. Mirrors expandArtifactTriggerConditions from the SDK v2 file.
func expandCdArtifactTriggerConditionsFramework(ctx context.Context, cd cdArtifactTriggerModel) []map[string]interface{} {
	var conditions []map[string]interface{}
	if !cd.BranchFilters.IsNull() && !cd.BranchFilters.IsUnknown() {
		var bfs []string
		if d := cd.BranchFilters.ElementsAs(ctx, &bfs, false); !d.HasError() {
			for _, bf := range bfs {
				conditions = append(conditions, map[string]interface{}{"sourceBranch": bf})
			}
		}
	}

	// Gather tag_filter tags, use_build_definition_branch, create_release_on_build_tagging.
	var tags []string
	if !cd.TagFilter.IsNull() && !cd.TagFilter.IsUnknown() {
		var tfModels []tagFilterModel
		if d := cd.TagFilter.ElementsAs(ctx, &tfModels, false); !d.HasError() && len(tfModels) > 0 {
			tf := tfModels[0]
			if !tf.Tags.IsNull() && !tf.Tags.IsUnknown() {
				_ = tf.Tags.ElementsAs(ctx, &tags, false)
			}
		}
	}
	useBuildBranch := cd.UseBuildDefinitionBranch.ValueBool()
	createOnTagging := cd.CreateReleaseOnBuildTagging.ValueBool()

	hasExtras := len(tags) > 0 || useBuildBranch || createOnTagging
	if !hasExtras {
		return conditions
	}

	// Emit extras on the first condition. If no branch conditions exist, create
	// a synthetic entry with empty sourceBranch as the container.
	if len(conditions) == 0 {
		conditions = []map[string]interface{}{{"sourceBranch": ""}}
	}
	if len(tags) > 0 {
		conditions[0]["tags"] = tags
	}
	if useBuildBranch {
		conditions[0]["useBuildDefinitionBranch"] = true
	}
	if createOnTagging {
		conditions[0]["createReleaseOnBuildTagging"] = true
	}
	return conditions
}

// flattenTriggersFramework converts the ADO polymorphic *[]interface{} into types.List
// (ListNestedAttribute of triggerModel). All trigger variants from the API are grouped
// into a single triggerModel list element (matches the typical ADO shape).
func flattenTriggersFramework(ctx context.Context, triggers *[]interface{}) (types.List, diag.Diagnostics) {
	_ = ctx
	elemType := types.ObjectType{AttrTypes: triggerAttrTypes}
	cdElemType := types.ObjectType{AttrTypes: cdArtifactTriggerAttrTypes}
	schElemType := types.ObjectType{AttrTypes: scheduleTriggerAttrTypes}
	srtElemType := types.ObjectType{AttrTypes: sourceRepoTriggerAttrTypes}
	ciElemType := types.ObjectType{AttrTypes: containerImageTriggerAttrTypes}
	prElemType := types.ObjectType{AttrTypes: pullRequestTriggerAttrTypes}
	tagFilterElemType := types.ObjectType{AttrTypes: tagFilterAttrTypes}

	if triggers == nil || len(*triggers) == 0 {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}

	var diags diag.Diagnostics
	var cdObjs []attr.Value
	var schObjs []attr.Value
	var srtObjs []attr.Value
	var ciObjs []attr.Value
	var prObjs []attr.Value

	for _, raw := range *triggers {
		data, err := json.Marshal(raw)
		if err != nil {
			continue
		}
		var trigMap map[string]interface{}
		if err := json.Unmarshal(data, &trigMap); err != nil {
			continue
		}
		trigType, _ := trigMap["triggerType"].(string)
		switch trigType {
		case "artifactSource":
			alias := ""
			if v, ok := trigMap["artifactAlias"].(string); ok {
				alias = v
			}
			var bfVals []attr.Value
			var tagVals []attr.Value
			useBuildBranch := false
			createOnTagging := false
			if conds, ok := trigMap["triggerConditions"].([]interface{}); ok {
				for i, c := range conds {
					if cm, ok := c.(map[string]interface{}); ok {
						if sb, ok := cm["sourceBranch"].(string); ok && sb != "" {
							bfVals = append(bfVals, types.StringValue(sb))
						}
						// Read extras from the first condition entry only.
						if i == 0 {
							if v, ok := cm["tags"].([]interface{}); ok {
								for _, t := range v {
									if s, ok := t.(string); ok {
										tagVals = append(tagVals, types.StringValue(s))
									}
								}
							}
							if v, ok := cm["useBuildDefinitionBranch"].(bool); ok {
								useBuildBranch = v
							}
							if v, ok := cm["createReleaseOnBuildTagging"].(bool); ok {
								createOnTagging = v
							}
						}
					}
				}
			}
			bfList, d := types.ListValue(types.StringType, bfVals)
			diags.Append(d...)

			// Build tag_filter list: one element if any tags, empty list otherwise.
			var tagFilterObjs []attr.Value
			if len(tagVals) > 0 {
				tagList, d2 := types.ListValue(types.StringType, tagVals)
				diags.Append(d2...)
				tfObj, d2 := types.ObjectValue(tagFilterAttrTypes, map[string]attr.Value{
					"tags": tagList,
				})
				diags.Append(d2...)
				tagFilterObjs = append(tagFilterObjs, tfObj)
			}
			tagFilterList, d := types.ListValue(tagFilterElemType, tagFilterObjs)
			diags.Append(d...)

			obj, d := types.ObjectValue(cdArtifactTriggerAttrTypes, map[string]attr.Value{
				"artifact_alias":                  types.StringValue(alias),
				"branch_filters":                  bfList,
				"tag_filter":                      tagFilterList,
				"use_build_definition_branch":     types.BoolValue(useBuildBranch),
				"create_release_on_build_tagging": types.BoolValue(createOnTagging),
			})
			diags.Append(d...)
			cdObjs = append(cdObjs, obj)

		case "schedule":
			onlyChanges := false
			startHours := int64(0)
			startMinutes := int64(0)
			tzID := "UTC"
			daysToRelease := int64(0)
			if sched, ok := trigMap["schedule"].(map[string]interface{}); ok {
				if v, ok := sched["scheduleOnlyWithChanges"].(bool); ok {
					onlyChanges = v
				}
				if v, ok := sched["startHours"].(float64); ok {
					startHours = int64(v)
				}
				if v, ok := sched["startMinutes"].(float64); ok {
					startMinutes = int64(v)
				}
				if v, ok := sched["timeZoneId"].(string); ok {
					tzID = v
				}
				if v, ok := sched["daysToRelease"].(float64); ok {
					daysToRelease = int64(v)
				}
			}
			obj, d := types.ObjectValue(scheduleTriggerAttrTypes, map[string]attr.Value{
				"schedule_only_with_changes": types.BoolValue(onlyChanges),
				"start_hours":                types.Int64Value(startHours),
				"start_minutes":              types.Int64Value(startMinutes),
				"time_zone_id":               types.StringValue(tzID),
				"days_to_release":            types.Int64Value(daysToRelease),
			})
			diags.Append(d...)
			schObjs = append(schObjs, obj)

		case "sourceRepo":
			alias := ""
			if v, ok := trigMap["alias"].(string); ok {
				alias = v
			}
			var bfVals []attr.Value
			if bfs, ok := trigMap["branchFilters"].([]interface{}); ok {
				for _, bf := range bfs {
					if s, ok := bf.(string); ok {
						bfVals = append(bfVals, types.StringValue(s))
					}
				}
			}
			bfList, d := types.ListValue(types.StringType, bfVals)
			diags.Append(d...)
			obj, d := types.ObjectValue(sourceRepoTriggerAttrTypes, map[string]attr.Value{
				"alias":          types.StringValue(alias),
				"branch_filters": bfList,
			})
			diags.Append(d...)
			srtObjs = append(srtObjs, obj)

		case "containerImage":
			alias := ""
			if v, ok := trigMap["alias"].(string); ok {
				alias = v
			}
			label := ""
			if tfs, ok := trigMap["tagFilters"].([]interface{}); ok && len(tfs) > 0 {
				if tf, ok := tfs[0].(map[string]interface{}); ok {
					label, _ = tf["pattern"].(string)
				}
			}
			obj, d := types.ObjectValue(containerImageTriggerAttrTypes, map[string]attr.Value{
				"artifact_alias": types.StringValue(alias),
				"label":          types.StringValue(label),
			})
			diags.Append(d...)
			ciObjs = append(ciObjs, obj)

		case "pullRequest":
			artifactAlias := ""
			if v, ok := trigMap["artifactAlias"].(string); ok {
				artifactAlias = v
			}
			var targetBranchVals []attr.Value
			var prTagVals []attr.Value
			if conds, ok := trigMap["triggerConditions"].([]interface{}); ok {
				for _, c := range conds {
					if cm, ok := c.(map[string]interface{}); ok {
						if tb, ok := cm["targetBranch"].(string); ok && tb != "" {
							targetBranchVals = append(targetBranchVals, types.StringValue(tb))
						}
						// Collect tags from the first condition that has them.
						if len(prTagVals) == 0 {
							if v, ok := cm["tags"].([]interface{}); ok {
								for _, t := range v {
									if s, ok := t.(string); ok {
										prTagVals = append(prTagVals, types.StringValue(s))
									}
								}
							}
						}
					}
				}
			}
			tbList, d := types.ListValue(types.StringType, targetBranchVals)
			diags.Append(d...)
			prTagList, d := types.ListValue(types.StringType, prTagVals)
			diags.Append(d...)
			useArtifactRef := true
			if pc, ok := trigMap["pullRequestConfiguration"].(map[string]interface{}); ok {
				if v, ok := pc["useArtifactReference"].(bool); ok {
					useArtifactRef = v
				}
			}
			obj, d := types.ObjectValue(pullRequestTriggerAttrTypes, map[string]attr.Value{
				"artifact_alias":         types.StringValue(artifactAlias),
				"target_branches":        tbList,
				"tags":                   prTagList,
				"use_artifact_reference": types.BoolValue(useArtifactRef),
			})
			diags.Append(d...)
			prObjs = append(prObjs, obj)
		}
	}

	if len(cdObjs) == 0 && len(schObjs) == 0 && len(srtObjs) == 0 && len(ciObjs) == 0 && len(prObjs) == 0 {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	cdList, d := types.ListValue(cdElemType, cdObjs)
	diags.Append(d...)
	schList, d := types.ListValue(schElemType, schObjs)
	diags.Append(d...)
	srtList, d := types.ListValue(srtElemType, srtObjs)
	diags.Append(d...)
	ciList, d := types.ListValue(ciElemType, ciObjs)
	diags.Append(d...)
	prList, d := types.ListValue(prElemType, prObjs)
	diags.Append(d...)

	trigObj, d := types.ObjectValue(triggerAttrTypes, map[string]attr.Value{
		"cd_artifact_trigger":     cdList,
		"schedule_trigger":        schList,
		"source_repo_trigger":     srtList,
		"container_image_trigger": ciList,
		"pull_request_trigger":    prList,
	})
	diags.Append(d...)

	list, d := types.ListValue(elemType, []attr.Value{trigObj})
	diags.Append(d...)
	return list, diags
}

// ── State upgraders ───────────────────────────────────────────────────────────

// UpgradeState implements resource.ResourceWithUpgradeState.
// Version 0 → 1: renames the flat environment list to stages (see state_upgrade_v0.go).
func (r *releaseDefinitionFrameworkResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: releaseDefinitionStateUpgraderV0(),
	}
}
