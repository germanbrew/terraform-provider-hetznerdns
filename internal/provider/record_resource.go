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
	"github.com/germanbrew/terraform-provider-hetznerdns/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &recordResource{}
	_ resource.ResourceWithImportState = &recordResource{}
)

func NewRecordResource() resource.Resource {
	return &recordResource{}
}

// recordResource defines the resource implementation.
type recordResource struct {
	client *api.Client
}

// recordResourceModel describes the resource data model.
type recordResourceModel struct {
	ID     types.String `tfsdk:"id"`
	ZoneID types.String `tfsdk:"zone_id"`
	Name   types.String `tfsdk:"name"`
	Type   types.String `tfsdk:"type"`
	Value  types.String `tfsdk:"value"`
	TTL    types.Int64  `tfsdk:"ttl"`
}

func (r *recordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_record"
}

func (r *recordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Provides a Hetzner DNS Zone resource to create, update and delete DNS Zones.",

		Attributes: map[string]schema.Attribute{
			"zone_id": schema.StringAttribute{
				Description: "ID of the DNS zone to create the record in.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Time to live of this record",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the DNS record to create",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"value": schema.StringAttribute{
				Description: "The value of the record (eg. 192.168.1.1)",
				MarkdownDescription: "The value of the record (eg. 192.168.1.1). For TXT records with quoted values, " +
					"the quotes have to be escaped in Terraform  (eg. \"v=spf1 include:\\_spf.google.com ~all\" is " +
					"represented by  \"\\\\\"v=spf1 include:\\_spf.google.com ~all\\\\\"\" in Terraform)",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "Time to live of this record",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
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

func (r *recordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *recordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "create resource record")

	var plan recordResourceModel

	// Read Terraform plan into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	value := plan.Value.ValueString()
	if plan.Type.ValueString() == "TXT" {
		value = utils.PlainToTXTRecordValue(value)
		if plan.Value.ValueString() != value {
			tflog.Debug(ctx, fmt.Sprintf("split TXT record value %d chunks: %q", len(value), value))
		}
	}

	httpResp, err := r.client.CreateRecord(ctx, api.CreateRecordOpts{
		ZoneID: plan.ZoneID.ValueString(),
		Name:   plan.Name.ValueString(),
		Type:   plan.Type.ValueString(),
		Value:  value,
		TTL:    plan.TTL.ValueInt64Pointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("error creating zone: %s", err))

		return
	}

	plan.ID = types.StringValue(httpResp.ID)

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *recordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "read resource record")

	var state recordResourceModel

	// Read Terraform prior state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, err := r.client.GetRecord(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read zene, got error: %s", err))

		return
	}

	if zone == nil {
		resp.Diagnostics.AddWarning("Resource Not Found", fmt.Sprintf("DNS zone with id %s doesn't exist, removing it from state", state.ID))

		return
	}

	if zone.Type == "TXT" {
		zone.Value = utils.TXTToPlainRecordValue(zone.Value)
	}

	state.Name = types.StringValue(zone.Name)
	state.TTL = types.Int64PointerValue(zone.TTL)
	state.ZoneID = types.StringValue(zone.ZoneID)
	state.Type = types.StringValue(zone.Type)
	state.Value = types.StringValue(zone.Value)
	state.ID = types.StringValue(zone.ID)

	// Save updated state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *recordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Trace(ctx, "updating resource record")

	var plan, state recordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	value := plan.Value.ValueString()
	if plan.Type.ValueString() == "TXT" {
		value = utils.PlainToTXTRecordValue(value)
	}

	if !plan.Name.Equal(state.Name) || !plan.TTL.Equal(state.TTL) || !plan.Type.Equal(state.Type) || !plan.Value.Equal(state.Value) {
		_, err := r.client.UpdateRecord(ctx, api.Record{
			ID:     state.ID.ValueString(),
			Name:   plan.Name.ValueString(),
			Type:   plan.Type.ValueString(),
			Value:  value,
			TTL:    plan.TTL.ValueInt64Pointer(),
			ZoneID: plan.ZoneID.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("error updating zone: %s", err))

			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *recordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "deleting resource record")

	var state recordResourceModel

	// Read Terraform prior state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteRecord(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("error deleting zone: %s", err))

		return
	}
}

func (r *recordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
