package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = (*ClientConfigDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ClientConfigDataSource)(nil)
)

// ClientConfigDataSource is the terraform-plugin-framework implementation of
// the betterado_client_config data source.
type ClientConfigDataSource struct {
	client *client.AggregatedClient
}

// NewClientConfigDataSource returns a new datasource.DataSource.
func NewClientConfigDataSource() datasource.DataSource {
	return &ClientConfigDataSource{}
}

// clientConfigModel is the Terraform state model for betterado_client_config.
type clientConfigModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	OrganizationURL types.String `tfsdk:"organization_url"`
	OwnerID         types.String `tfsdk:"owner_id"`
	Status          types.String `tfsdk:"status"`
	TenantID        types.String `tfsdk:"tenant_id"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *ClientConfigDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "betterado_client_config"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *ClientConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to access the configuration of the AzureDevOps provider.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the Azure DevOps organization.",
			},
			"organization_url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL of the Azure DevOps organization.",
			},
			"owner_id": schema.StringAttribute{
				Computed:    true,
				Description: "The owner ID of the Azure DevOps organization.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the Azure DevOps organization.",
			},
			"tenant_id": schema.StringAttribute{
				Computed:    true,
				Description: "The tenant ID of the Azure DevOps organization.",
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *ClientConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	agg, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *client.AggregatedClient, got: %T", req.ProviderData),
		)
		return
	}
	d.client = agg
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (d *ClientConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model clientConfigModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.Split(d.client.OrganizationURL, "/")
	if len(parts) < 4 {
		resp.Diagnostics.AddError("invalid organization URL", fmt.Sprintf("unexpected format: %s", d.client.OrganizationURL))
		return
	}

	orgMeta, err := d.client.OrganizationClient.GetOrganization(d.client.Ctx, parts[3])
	if err != nil {
		resp.Diagnostics.AddError("reading organization metadata", err.Error())
		return
	}

	model.ID = types.StringValue(*orgMeta.Id)
	model.OrganizationURL = types.StringValue(d.client.OrganizationURL)
	model.Name = types.StringValue(*orgMeta.Name)
	if orgMeta.Status != nil {
		model.Status = types.StringValue(*orgMeta.Status)
	} else {
		model.Status = types.StringValue("")
	}
	if orgMeta.TenantId != nil {
		model.TenantID = types.StringValue(*orgMeta.TenantId)
	} else {
		model.TenantID = types.StringValue("")
	}
	if orgMeta.Owner != nil {
		model.OwnerID = types.StringValue(*orgMeta.Owner)
	} else {
		model.OwnerID = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
