package provider

import (
	"context"
	"fmt"
	"regexp"

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
	_ resource.Resource                = &zoneResource{}
	_ resource.ResourceWithImportState = &zoneResource{}
)

func NewZoneResource() resource.Resource {
	return &zoneResource{}
}

// zoneResource defines the resource implementation.
type zoneResource struct {
	client *api.Client
}

// zoneResourceModel describes the resource data model.
type zoneResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	TTL  types.Int64  `tfsdk:"ttl"`
}

func (r *zoneResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (r *zoneResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Provides a Hetzner DNS Zone resource to create, update and delete DNS Zones.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the DNS zone to create.",
				MarkdownDescription: "Name of the DNS zone to create. Must be a valid domain with top level domain. " +
					"Meaning `<domain>.de` or `<domain>.io`. Don't include sub domains on this level. So, no " +
					"`sub.<domain>.io`. The Hetzner API rejects attempts to create a zone with a sub domain name." +
					"Use a record to create the sub domain.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9-]+\.[a-z0-9-]+$`),
						"Name must be a valid domain with top level domain",
					),
				},
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "Time to live of this zone",
				Optional:            true,
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

func (r *zoneResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *zoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "create resource zone")

	var plan zoneResourceModel

	// Read Terraform plan into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.GetZoneByName(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("error creating zone: %s", err))

		return
	}

	if httpResp != nil {
		resp.Diagnostics.AddError("Error", fmt.Sprintf("zone %q already exists", plan.Name.ValueString()))

		return
	}

	httpResp, err = r.client.CreateZone(ctx, api.CreateZoneOpts{
		Name: plan.Name.ValueString(),
		TTL:  plan.TTL.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("error creating zone: %s", err))

		return
	}

	plan.ID = types.StringValue(httpResp.ID)

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *zoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Trace(ctx, "read resource zone")

	var state zoneResourceModel

	// Read Terraform prior state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, err := r.client.GetZone(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read zene, got error: %s", err))

		return
	}

	if zone == nil {
		resp.Diagnostics.AddWarning("Resource Not Found", fmt.Sprintf("DNS zone with id %s doesn't exist, removing it from state", state.ID))

		return
	}

	state.Name = types.StringValue(zone.Name)
	state.TTL = types.Int64Value(zone.TTL)
	state.ID = types.StringValue(zone.ID)

	// Save updated state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *zoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Trace(ctx, "update resource zone")

	var plan, state zoneResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.TTL.Equal(state.TTL) {
		_, err := r.client.UpdateZone(ctx, api.Zone{
			ID:   state.ID.ValueString(),
			Name: plan.Name.ValueString(),
			TTL:  plan.TTL.ValueInt64(),
		})
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("error updating zone: %s", err))

			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *zoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Trace(ctx, "deleting resource zone")

	var state zoneResourceModel

	// Read Terraform prior state into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteZone(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("error deleting zone: %s", err))

		return
	}
}

func (r *zoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
