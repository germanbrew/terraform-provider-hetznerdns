package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/germanbrew/terraform-provider-hetznerdns/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
	provider *providerClient
}

// primaryServerResourceModel describes the resource data model.
type primaryServerResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Address types.String `tfsdk:"address"`
	Port    types.Int64  `tfsdk:"port"`
	ZoneID  types.String `tfsdk:"zone_id"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *primaryServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_primary_server"
}

func (r *primaryServerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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

		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,

				CreateDescription: `[Operation Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) consisting of numbers and unit suffixes, such as "30s" or "2h45m".
Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Default: 5m`,
				DeleteDescription: `[Operation Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) consisting of numbers and unit suffixes, such as "30s" or "2h45m".
Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Default: 5m`,
				ReadDescription: `A[Operation Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) consisting of numbers and unit suffixes, such as "30s" or "2h45m".
Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Default: 5m`,
				UpdateDescription: `[Operation Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) consisting of numbers and unit suffixes, such as "30s" or "2h45m".
Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Default: 5m`,
			}),
		},
	}
}

func (r *primaryServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *primaryServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "creating primary server")

	var plan primaryServerResourceModel

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

	var (
		err     error
		server  *api.PrimaryServer
		retries int64
	)

	serverRequest := api.CreatePrimaryServerRequest{
		ZoneID:  plan.ZoneID.ValueString(),
		Address: plan.Address.ValueString(),
		Port:    uint16(plan.Port.ValueInt64()),
	}

	err = retry.RetryContext(ctx, createTimeout, func() *retry.RetryError {
		retries++

		server, err = r.provider.apiClient.CreatePrimaryServer(ctx, serverRequest)
		if err != nil {
			if retries == r.provider.maxRetries {
				return retry.NonRetryableError(err)
			}

			return retry.RetryableError(err)
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("creating primary server: %s", err))

		return
	}

	plan.ID = types.StringValue(server.ID)

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

	readTimeout, diags := state.Timeouts.Read(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		err     error
		server  *api.PrimaryServer
		retries int64
	)

	err = retry.RetryContext(ctx, readTimeout, func() *retry.RetryError {
		retries++

		server, err = r.provider.apiClient.GetPrimaryServer(ctx, state.ID.ValueString())
		if err != nil {
			if retries == r.provider.maxRetries {
				return retry.NonRetryableError(err)
			}

			return retry.RetryableError(err)
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("read primary server: %s", err))

		return
	}

	if server == nil {
		resp.Diagnostics.AddWarning("Resource Not Found", fmt.Sprintf("Primary server with id %s doesn't exist, removing it from state", state.ID))

		return
	}

	state.ID = types.StringValue(server.ID)
	state.Address = types.StringValue(server.Address)
	state.ZoneID = types.StringValue(server.ZoneID)
	state.Port = types.Int64Value(int64(server.Port))

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
		updateTimeout, diags := plan.Timeouts.Update(ctx, 5*time.Minute)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		var (
			err     error
			retries int64
		)

		server := api.PrimaryServer{
			ID:      state.ID.ValueString(),
			Address: plan.Address.ValueString(),
			Port:    uint16(plan.Port.ValueInt64()),
			ZoneID:  plan.ZoneID.ValueString(),
		}

		err = retry.RetryContext(ctx, updateTimeout, func() *retry.RetryError {
			retries++

			_, err = r.provider.apiClient.UpdatePrimaryServer(ctx, server)
			if err != nil {
				if retries == r.provider.maxRetries {
					return retry.NonRetryableError(err)
				}

				return retry.RetryableError(err)
			}

			return nil
		})
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("update primary server %s: %s", state.ID, err))

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

		err = r.provider.apiClient.DeletePrimaryServer(ctx, state.ID.ValueString())
		if err != nil {
			if retries == r.provider.maxRetries {
				return retry.NonRetryableError(err)
			}

			return retry.RetryableError(err)
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("deleting primary server %s: %s", state.ID, err))

		return
	}
}

func (r *primaryServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
