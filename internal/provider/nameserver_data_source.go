package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/germanbrew/terraform-provider-hetznerdns/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &nameserversDataSource{}

func NewNameserversDataSource() datasource.DataSource {
	return &nameserversDataSource{}
}

// nameserversDataSource defines the data source implementation.
type nameserversDataSource struct {
	provider *providerClient
}

type singleNameserverDataModel struct {
	Name types.String `tfsdk:"name"`
	IPV4 types.String `tfsdk:"ipv4"`
	IPV6 types.String `tfsdk:"ipv6"`
}

// nameserversDataSourceModel describes the data source data model.
type nameserversDataSourceModel struct {
	Type types.String                `tfsdk:"type"`
	NS   []singleNameserverDataModel `tfsdk:"ns"`
}

func getValidNameserverTypes() []string {
	return []string{"authoritative", "secondary", "konsoleh"}
}

func (d *nameserversDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nameservers"
}

func singleNameserverSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the name server",
			Computed:            true,
		},
		"ipv4": schema.StringAttribute{
			MarkdownDescription: "IPv4 address of the name server",
			Computed:            true,
		},
		"ipv6": schema.StringAttribute{
			MarkdownDescription: "IPv6 address of the name server",
			Computed:            true,
		},
	}
}

func (d *nameserversDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Provides details about name servers used by Hetzner DNS",

		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(
					"Type of name servers to get data from. Possible values: `%s`",
					strings.Join(getValidNameserverTypes(), "`, `")),
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(getValidNameserverTypes()...),
				},
			},
			"ns": schema.SetNestedAttribute{
				MarkdownDescription: "Name servers",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: singleNameserverSchema(),
				},
			},
		},
	}
}

func (d *nameserversDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func populateNameserverData(data *[]singleNameserverDataModel, nameservers []api.Nameserver) *[]singleNameserverDataModel {
	for _, ns := range nameservers {
		*data = append(*data, singleNameserverDataModel{
			Name: types.StringValue(ns["name"]),
			IPV4: types.StringValue(ns["ipv4"]),
			IPV6: types.StringValue(ns["ipv6"]),
		})
	}

	return data
}

func (d *nameserversDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var (
		err         error
		data        nameserversDataSourceModel
		nameservers []api.Nameserver
	)

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the nameservers based on the type
	switch data.Type.ValueString() {
	case "authoritative":
		nameservers, err = api.GetAuthoritativeNameservers(ctx)
	case "secondary":
		nameservers, err = api.GetSecondaryNameservers(ctx)
	case "konsoleh":
		nameservers, err = api.GetKonsolehNameservers(ctx)
	default:
		resp.Diagnostics.AddError(
			"Type Error",
			fmt.Sprintf("Invalid nameserver type: %s, must be one of %s",
				data.Type.String(),
				strings.Join(getValidNameserverTypes(), ", "),
			),
		)

		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Error while getting nameservers: %s", err),
		)

		return
	}

	// Populate the data model
	data.NS = make([]singleNameserverDataModel, 0, len(nameservers))
	populateNameserverData(&data.NS, nameservers)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
