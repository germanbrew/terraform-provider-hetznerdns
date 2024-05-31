package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
}

func (d *recordsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_records"
}

func (d *recordsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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

	if data.ZoneID.ValueString() == "" {
		resp.Diagnostics.AddError("Attribute Error", "no 'zone_id' set")
	}

	if resp.Diagnostics.HasError() {
		return
	}

	records, err := d.provider.apiClient.GetRecordsByZoneID(ctx, data.ZoneID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to get records from zone, got error: %s", err))

		return
	}

	// if zone == nil {
	// 	resp.Diagnostics.AddError("API Error", fmt.Sprintf("DNS zone '%s' doesn't exist", data.Name.ValueString()))

	// 	return
	// }

	// ns, diags := types.ListValueFrom(ctx, types.StringType, zone.NS)

	// resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var elements []recordDataSourceModel
	for _, record := range *records {
		elements = append(elements,
			recordDataSourceModel{
				ZoneID: types.StringValue(record.ZoneID),
				ID:     types.StringValue(record.ID),
				Type:   types.StringValue(record.Type),
				Name:   types.StringValue(record.Name),
				Value:  types.StringValue(record.Value),
				TTL:    types.Int64Value(ttl),  // FIXME memory error in this line
			},
		)
	}

	values, diags := types.ListValueFrom(ctx, types.ListType, elements)  // FIXME element type in this line

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &values)...)
}

// TODO Write tests
