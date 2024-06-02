package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/germanbrew/terraform-provider-hetznerdns/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &zoneDataSource{}

func NewZoneDataSource() datasource.DataSource {
	return &zoneDataSource{}
}

// zoneDataSource defines the data source implementation.
type zoneDataSource struct {
	provider *providerClient
}

// zoneDataSourceModel describes the data source data model.
type zoneDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	TTL  types.Int64  `tfsdk:"ttl"`
	NS   types.List   `tfsdk:"ns"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (d *zoneDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (d *zoneDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Provides details about a Hetzner DNS Zone",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the DNS zone to get data from",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "Time to live of this zone",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the DNS zone",
				Computed:            true,
			},
			"ns": schema.ListAttribute{
				MarkdownDescription: "Name Servers of the zone",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},

		Blocks: map[string]schema.Block{
			"timeouts": timeouts.BlockWithOpts(ctx, timeouts.Opts{
				ReadDescription: `A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes,
 such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when
 refresh is enabled. Default: 5m`,
			}),
		},
	}
}

func (d *zoneDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *zoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data zoneDataSourceModel

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
		zone    *api.Zone
		retries int64
	)

	err = retry.RetryContext(ctx, readTimeout, func() *retry.RetryError {
		retries++

		zone, err = d.provider.apiClient.GetZoneByName(ctx, data.Name.ValueString())
		if err != nil {
			if retries == d.provider.maxRetries {
				return retry.NonRetryableError(err)
			}

			return retry.RetryableError(err)
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to get zone, got error: %s", err))

		return
	}

	if zone == nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("DNS zone '%s' doesn't exist", data.Name.ValueString()))

		return
	}

	ns, diags := types.ListValueFrom(ctx, types.StringType, zone.NS)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(zone.ID)
	data.Name = types.StringValue(zone.Name)
	data.TTL = types.Int64Value(zone.TTL)
	data.NS = ns

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
