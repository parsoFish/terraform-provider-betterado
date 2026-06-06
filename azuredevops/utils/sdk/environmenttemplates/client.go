package environmenttemplates

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
)

// locationId for the release/definitions/environmenttemplates ADO REST endpoint.
// Source: ADO REST reference — release resource area, environmenttemplates location.
// Confirmed by the WI pre-research (WI-1.md §Known context).
var environmentTemplatesLocationId, _ = uuid.Parse("6b3ad47a-2a42-4e24-9785-e3a0a8e3e64d") //nolint:errcheck

const apiVersion = "7.1-preview.1"

// Client is the interface for ADO environment-templates operations.
// Methods mirror the ADO REST surface:
//
//	GET  /{project}/_apis/release/definitions/environmenttemplates
//	GET  /{project}/_apis/release/definitions/environmenttemplates?templateId={id}
//	POST /{project}/_apis/release/definitions/environmenttemplates
//	DEL  /{project}/_apis/release/definitions/environmenttemplates?templateId={id}
type Client interface {
	// GetEnvironmentTemplates lists all environment templates for a project.
	GetEnvironmentTemplates(ctx context.Context, args *GetEnvironmentTemplatesArgs) (*[]ReleaseDefinitionEnvironmentTemplate, error)
	// GetEnvironmentTemplate retrieves a single environment template by ID.
	GetEnvironmentTemplate(ctx context.Context, args *GetEnvironmentTemplateArgs) (*ReleaseDefinitionEnvironmentTemplate, error)
	// SaveEnvironmentTemplate creates or updates an environment template.
	SaveEnvironmentTemplate(ctx context.Context, args *SaveEnvironmentTemplateArgs) (*ReleaseDefinitionEnvironmentTemplate, error)
	// DeleteEnvironmentTemplate deletes an environment template by ID.
	DeleteEnvironmentTemplate(ctx context.Context, args *DeleteEnvironmentTemplateArgs) error
}

// ClientImpl is the concrete implementation backed by azuredevops.Client.Send.
type ClientImpl struct {
	Client azuredevops.Client
}

// NewClient constructs a Client for the given connection.
func NewClient(ctx context.Context, connection *azuredevops.Connection) Client {
	client := connection.GetClientByUrl(connection.BaseUrl)
	return &ClientImpl{
		Client: *client,
	}
}

// GetEnvironmentTemplates lists all environment templates for a project.
func (c *ClientImpl) GetEnvironmentTemplates(ctx context.Context, args *GetEnvironmentTemplatesArgs) (*[]ReleaseDefinitionEnvironmentTemplate, error) {
	if args == nil || args.Project == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.Project"}
	}
	routeValues := map[string]string{
		"project": *args.Project,
	}
	queryParams := url.Values{}

	resp, err := c.Client.Send(ctx, http.MethodGet, environmentTemplatesLocationId, apiVersion, routeValues, queryParams, nil, "", "application/json", nil)
	if err != nil {
		return nil, err
	}

	var result []ReleaseDefinitionEnvironmentTemplate
	err = c.Client.UnmarshalCollectionBody(resp, &result)
	return &result, err
}

// GetEnvironmentTemplate retrieves a single environment template by ID.
func (c *ClientImpl) GetEnvironmentTemplate(ctx context.Context, args *GetEnvironmentTemplateArgs) (*ReleaseDefinitionEnvironmentTemplate, error) {
	if args == nil || args.Project == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.Project"}
	}
	routeValues := map[string]string{
		"project": *args.Project,
	}
	queryParams := url.Values{}
	if args.TemplateId != nil {
		queryParams.Set("templateId", args.TemplateId.String())
	}

	resp, err := c.Client.Send(ctx, http.MethodGet, environmentTemplatesLocationId, apiVersion, routeValues, queryParams, nil, "", "application/json", nil)
	if err != nil {
		return nil, err
	}

	var result ReleaseDefinitionEnvironmentTemplate
	err = c.Client.UnmarshalBody(resp, &result)
	return &result, err
}

// SaveEnvironmentTemplate creates or updates an environment template.
func (c *ClientImpl) SaveEnvironmentTemplate(ctx context.Context, args *SaveEnvironmentTemplateArgs) (*ReleaseDefinitionEnvironmentTemplate, error) {
	if args == nil || args.Project == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.Project"}
	}
	routeValues := map[string]string{
		"project": *args.Project,
	}

	body, err := json.Marshal(args.Template)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Send(ctx, http.MethodPost, environmentTemplatesLocationId, apiVersion, routeValues, nil, bytes.NewReader(body), "application/json", "application/json", nil)
	if err != nil {
		return nil, err
	}

	var result ReleaseDefinitionEnvironmentTemplate
	err = c.Client.UnmarshalBody(resp, &result)
	return &result, err
}

// DeleteEnvironmentTemplate deletes an environment template by ID.
func (c *ClientImpl) DeleteEnvironmentTemplate(ctx context.Context, args *DeleteEnvironmentTemplateArgs) error {
	if args == nil || args.Project == nil {
		return &azuredevops.ArgumentNilError{ArgumentName: "args.Project"}
	}
	routeValues := map[string]string{
		"project": *args.Project,
	}
	queryParams := url.Values{}
	if args.TemplateId != nil {
		queryParams.Set("templateId", args.TemplateId.String())
	}

	_, err := c.Client.Send(ctx, http.MethodDelete, environmentTemplatesLocationId, apiVersion, routeValues, queryParams, nil, "", "", nil)
	return err
}
