package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/germanbrew/terraform-provider-hetznerdns/internal/api"
	"github.com/germanbrew/terraform-provider-hetznerdns/internal/utils"
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
	_ resource.Resource                = &recordResource{}
	_ resource.ResourceWithImportState = &recordResource{}
)

func NewRecordResource() resource.Resource {
	return &recordResource{}
}

// recordResource defines the resource implementation.
type recordResource struct {
	provider *providerClient
}

// recordResourceModel describes the resource data model.
type recordResourceModel struct {
	ID     types.String `tfsdk:"id"`
	ZoneID types.String `tfsdk:"zone_id"`
	Name   types.String `tfsdk:"name"`
	Type   types.String `tfsdk:"type"`
	Value  types.String `tfsdk:"value"`
	TTL    types.Int64  `tfsdk:"ttl"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *recordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_record"
}

func (r *recordResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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

func (r *recordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *recordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Trace(ctx, "create resource record")

	var plan recordResourceModel

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

	value := plan.Value.ValueString()
	if plan.Type.ValueString() == "TXT" && r.provider.txtFormatter {
		value = utils.PlainToTXTRecordValue(value)
		if plan.Value.ValueString() != value {
			tflog.Debug(ctx, fmt.Sprintf("split TXT record value %d chunks: %q", len(value), value))
		}
	}

	var (
		err     error
		record  *api.Record
		retries int64
	)

	recordRequest := api.CreateRecordOpts{
		ZoneID: plan.ZoneID.ValueString(),
		Name:   plan.Name.ValueString(),
		Type:   plan.Type.ValueString(),
		Value:  value,
		TTL:    plan.TTL.ValueInt64Pointer(),
	}

	err = retry.RetryContext(ctx, createTimeout, func() *retry.RetryError {
		retries++

		record, err = r.provider.apiClient.CreateRecord(ctx, recordRequest)
		if err != nil {
			if retries == r.provider.maxRetries {
				return retry.NonRetryableError(err)
			}

			return retry.RetryableError(err)
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("creating record: %s", err))

		return
	}

	plan.ID = types.StringValue(record.ID)

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

	readTimeout, diags := state.Timeouts.Read(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		err     error
		record  *api.Record
		retries int64
	)

	err = retry.RetryContext(ctx, readTimeout, func() *retry.RetryError {
		retries++

		record, err = r.provider.apiClient.GetRecord(ctx, state.ID.ValueString())
		if err != nil {
			if retries == r.provider.maxRetries {
				return retry.NonRetryableError(err)
			}

			return retry.RetryableError(err)
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("read record: %s", err))

		return
	}

	if record == nil {
		resp.Diagnostics.AddWarning("Resource Not Found", fmt.Sprintf("DNS record with id %s doesn't exist, removing it from state", state.ID))
		resp.State.RemoveResource(ctx)

		return
	}

	if record.Type == "TXT" && r.provider.txtFormatter {
		record.Value = utils.TXTRecordToPlainValue(record.Value)
	}

	state.Name = types.StringValue(record.Name)
	state.TTL = types.Int64PointerValue(record.TTL)
	state.ZoneID = types.StringValue(record.ZoneID)
	state.Type = types.StringValue(record.Type)
	state.Value = types.StringValue(record.Value)
	state.ID = types.StringValue(record.ID)

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
	if plan.Type.ValueString() == "TXT" && r.provider.txtFormatter {
		value = utils.PlainToTXTRecordValue(value)
	}

	if !plan.Name.Equal(state.Name) || !plan.TTL.Equal(state.TTL) || !plan.Type.Equal(state.Type) || !plan.Value.Equal(state.Value) {
		updateTimeout, diags := plan.Timeouts.Update(ctx, 5*time.Minute)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		var (
			err     error
			retries int64
		)

		record := api.Record{
			ID:     state.ID.ValueString(),
			Name:   plan.Name.ValueString(),
			Type:   plan.Type.ValueString(),
			Value:  value,
			TTL:    plan.TTL.ValueInt64Pointer(),
			ZoneID: plan.ZoneID.ValueString(),
		}

		err = retry.RetryContext(ctx, updateTimeout, func() *retry.RetryError {
			retries++

			_, err = r.provider.apiClient.UpdateRecord(ctx, record)
			if err != nil {
				if retries == r.provider.maxRetries {
					return retry.NonRetryableError(err)
				}

				return retry.RetryableError(err)
			}

			return nil
		})
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("update record: %s", err))

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

		err = r.provider.apiClient.DeleteRecord(ctx, state.ID.ValueString())
		if err != nil {
			if retries == r.provider.maxRetries {
				return retry.NonRetryableError(err)
			}

			return retry.RetryableError(err)
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("deleting record %s: %s", state.ID, err))

		return
	}
}

func (r *recordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
