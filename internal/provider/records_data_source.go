package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/germanbrew/terraform-provider-hetznerdns/internal/api"
	"github.com/germanbrew/terraform-provider-hetznerdns/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &recordsDataSource{}

func NewRecordsDataSource() datasource.DataSource {
	return &recordsDataSource{}
}

// recordsDataSource defines the data source implementation.
type recordsDataSource struct {
	provider *providerClient
}

// recordDataSourceModel describes the data source data model.
type recordDataSourceModel struct {
	ZoneID types.String `tfsdk:"zone_id"`
	ID     types.String `tfsdk:"id"`
	Type   types.String `tfsdk:"type"`
	Name   types.String `tfsdk:"name"`
	Value  types.String `tfsdk:"value"`
	TTL    types.Int64  `tfsdk:"ttl"`
}

// recordsDataSourceModel describes the data source data model.
type recordsDataSourceModel struct {
	ZoneID  types.String `tfsdk:"zone_id"`
	Records types.List   `tfsdk:"records"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (d *recordsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_records"
}

func (d *recordsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Provides details about all Records of a Hetzner DNS Zone",

		Attributes: map[string]schema.Attribute{
			"records": schema.ListNestedAttribute{
				MarkdownDescription: "The DNS records of the zone",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"zone_id": schema.StringAttribute{
							MarkdownDescription: "ID of the DNS zone",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "ID of this DNS record",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of this DNS record",
							Computed:            true,
						},
						"ttl": schema.Int64Attribute{
							MarkdownDescription: "Time to live of this record",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of this DNS record",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Value of this DNS record",
							Computed:            true,
						},
					},
				},
			},
			"zone_id": schema.StringAttribute{
				MarkdownDescription: "ID of the DNS zone to get records from",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},

		Blocks: map[string]schema.Block{
			"timeouts": timeouts.BlockWithOpts(ctx, timeouts.Opts{
				ReadDescription: `[Operation Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) consisting of
numbers and unit suffixes, such as "30s" or "2h45m".\
Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Default: 5m`,
			}),
		},
	}
}

func (d *recordsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*providerClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *providerClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.provider = provider
}

func (d *recordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data recordsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := data.Timeouts.Read(ctx, 5*time.Minute)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		err     error
		records *[]api.Record
		retries int64
	)

	err = retry.RetryContext(ctx, readTimeout, func() *retry.RetryError {
		retries++

		records, err = d.provider.apiClient.GetRecordsByZoneID(ctx, data.ZoneID.ValueString())
		if err != nil {
			if retries == d.provider.maxRetries {
				return retry.NonRetryableError(err)
			}

			return retry.RetryableError(err)
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to get records from zone, got error: %s", err))

		return
	}

	elements := make([]recordDataSourceModel, 0, len(*records))

	for _, record := range *records {
		if record.Type == "TXT" && d.provider.txtFormatter {
			value := utils.TXTRecordToPlainValue(record.Value)
			if record.Value != value {
				tflog.Info(ctx, fmt.Sprintf("split TXT record value %d chunks: %q", len(value), value))
			}

			record.Value = value
		}

		elements = append(elements,
			recordDataSourceModel{
				ZoneID: types.StringValue(record.ZoneID),
				ID:     types.StringValue(record.ID),
				Type:   types.StringValue(record.Type),
				Name:   types.StringValue(record.Name),
				Value:  types.StringValue(record.Value),
				TTL:    types.Int64PointerValue(record.TTL),
			},
		)
	}

	data.Records, diags = types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"zone_id": types.StringType,
			"id":      types.StringType,
			"type":    types.StringType,
			"name":    types.StringType,
			"value":   types.StringType,
			"ttl":     types.Int64Type,
		},
	}, elements)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
