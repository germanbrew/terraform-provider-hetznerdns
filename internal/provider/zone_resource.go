package provider

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/germanbrew/terraform-provider-hetznerdns/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
	provider *providerClient
}

// zoneResourceModel describes the resource data model.
type zoneResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	TTL  types.Int64  `tfsdk:"ttl"`
	NS   types.List   `tfsdk:"ns"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *zoneResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (r *zoneResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"ns": schema.ListAttribute{
				Computed:            true,
				MarkdownDescription: "Name Servers of the zone",
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},

		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,

				CreateDescription: `[Operation Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) consisting of
numbers and unit suffixes, such as "30s" or "2h45m".\
Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Default: 5m`,
				DeleteDescription: `[Operation Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) consisting of
numbers and unit suffixes, such as "30s" or "2h45m".\
Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Default: 5m`,
				ReadDescription: `[Operation Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) consisting of
numbers and unit suffixes, such as "30s" or "2h45m".\
Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Default: 5m`,
				UpdateDescription: `[Operation Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) consisting of
numbers and unit suffixes, such as "30s" or "2h45m".\
Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Default: 5m`,
			}),
		},
	}
}

func (r *zoneResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*providerClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *providerClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.provider = provider
}

func (r *zoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "create resource zone")

	var plan zoneResourceModel

	// Read Terraform plan into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, err := r.provider.apiClient.GetZoneByName(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("error read zone: %s", err))

		return
	} else if zone != nil {
		resp.Diagnostics.AddError("Error", fmt.Sprintf("zone %q already exists", plan.Name.ValueString()))

		return
	}

	var retries int64

	zoneRequest := api.CreateZoneOpts{
		Name: plan.Name.ValueString(),
		TTL:  plan.TTL.ValueInt64(),
	}

	err = retry.RetryContext(ctx, createTimeout, func() *retry.RetryError {
		retries++

		zone, err = r.provider.apiClient.CreateZone(ctx, zoneRequest)
		if err != nil {
			if retries == r.provider.maxRetries {
				return retry.NonRetryableError(err)
			}

			return retry.RetryableError(err)
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("creating zone: %s", err))

		return
	}

	ns, diags := types.ListValueFrom(ctx, types.StringType, zone.NS)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(zone.ID)
	plan.NS = ns

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

	readTimeout, diags := state.Timeouts.Read(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		err     error
		zone    *api.Zone
		retries int64
	)

	err = retry.RetryContext(ctx, readTimeout, func() *retry.RetryError {
		retries++

		zone, err = r.provider.apiClient.GetZone(ctx, state.ID.ValueString())
		if err != nil {
			if retries == r.provider.maxRetries {
				return retry.NonRetryableError(err)
			}

			return retry.RetryableError(err)
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("read zone: %s", err))

		return
	}

	if zone == nil {
		resp.Diagnostics.AddWarning("Resource Not Found", fmt.Sprintf("DNS zone with id %s doesn't exist, removing it from state", state.ID))

		return
	}

	ns, diags := types.ListValueFrom(ctx, types.StringType, zone.NS)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	state.Name = types.StringValue(zone.Name)
	state.TTL = types.Int64Value(zone.TTL)
	state.ID = types.StringValue(zone.ID)
	state.NS = ns

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
		updateTimeout, diags := plan.Timeouts.Update(ctx, 5*time.Minute)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		var (
			err     error
			retries int64
		)

		zone := api.Zone{
			ID:   state.ID.ValueString(),
			Name: plan.Name.ValueString(),
			TTL:  plan.TTL.ValueInt64(),
		}

		err = retry.RetryContext(ctx, updateTimeout, func() *retry.RetryError {
			retries++

			_, err = r.provider.apiClient.UpdateZone(ctx, zone)
			if err != nil {
				if retries == r.provider.maxRetries {
					return retry.NonRetryableError(err)
				}

				return retry.RetryableError(err)
			}

			return nil
		})
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("update zone: %s", err))

			return
		}

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
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

	deleteTimeout, diags := state.Timeouts.Delete(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		err     error
		retries int64
	)

	err = retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
		retries++

		err = r.provider.apiClient.DeleteZone(ctx, state.ID.ValueString())
		if err != nil {
			if retries == r.provider.maxRetries {
				return retry.NonRetryableError(err)
			}

			return retry.RetryableError(err)
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("deleting zone %s: %s", state.ID, err))

		return
	}
}

func (r *zoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
