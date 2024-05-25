package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/germanbrew/terraform-provider-hetznerdns/internal/api"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &primaryServerResource{}
	_ resource.ResourceWithImportState = &primaryServerResource{}
)

func NewPrimaryServerResource() resource.Resource {
	return &primaryServerResource{}
}

// primaryServerResource defines the resource implementation.
type primaryServerResource struct {
	client *api.Client
}

// primaryServerResourceModel describes the resource data model.
type primaryServerResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Address types.String `tfsdk:"address"`
	Port    types.Int64  `tfsdk:"port"`
	ZoneID  types.String `tfsdk:"zone_id"`
}

func (r *primaryServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_primary_server"
}

func (r *primaryServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Configure primary server for a domain",

		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				Description: "Address of the primary server.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Port of the primary server.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
					int64validator.AtMost(65535),
				},
			},
			"zone_id": schema.StringAttribute{
				MarkdownDescription: "Zone identifier",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Zone identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *primaryServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *primaryServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "creating primary server")

	var plan primaryServerResourceModel

	// Read Terraform plan into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.CreatePrimaryServer(ctx, api.CreatePrimaryServerRequest{
		ZoneID:  plan.ZoneID.String(),
		Address: plan.Address.String(),
		Port:    uint16(plan.Port.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("creating primary server: %s", err))

		return
	}

	plan.ID = types.StringValue(httpResp.ID)

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *primaryServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "reading primary server")

	var state primaryServerResourceModel

	// Read Terraform prior state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	record, err := r.client.GetPrimaryServer(ctx, state.ID.String())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read zene, got error: %s", err))

		return
	}

	if record == nil {
		resp.Diagnostics.AddWarning("Resource Not Found", fmt.Sprintf("Primary server with id %s doesn't exist, removing it from state", state.ID))

		return
	}

	state.ID = types.StringValue(record.ID)
	state.Address = types.StringValue(record.Address)
	state.ZoneID = types.StringValue(record.ZoneID)
	state.Port = types.Int64Value(int64(record.Port))

	// Save updated state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *primaryServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Trace(ctx, "updating primary server")

	var plan, state primaryServerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Address.Equal(state.Address) || !plan.Port.Equal(state.Port) {
		_, err := r.client.UpdatePrimaryServer(ctx, api.PrimaryServer{
			Address: plan.Address.String(),
			Port:    uint16(plan.Port.ValueInt64()),
		})
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("error primary server: %s", err))

			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *primaryServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "deleting resource record")

	var state primaryServerResourceModel

	// Read Terraform prior state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeletePrimaryServer(ctx, state.ID.String()); err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Error deleting primary server %s: %s", state.ID, err))

		return
	}
}

func (r *primaryServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
