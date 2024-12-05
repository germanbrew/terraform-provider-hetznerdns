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
var _ datasource.DataSource = &nameserverDataSource{}

func NewNameserverDataSource() datasource.DataSource {
	return &nameserverDataSource{}
}

// nameserverDataSource defines the data source implementation.
type nameserverDataSource struct {
	provider *providerClient
}

type singleNameserverDataModel struct {
	Name types.String `tfsdk:"name"`
	IPV4 types.String `tfsdk:"ipv4"`
	IPV6 types.String `tfsdk:"ipv6"`
}

// nameserverDataSourceModel describes the data source data model.
type nameserverDataSourceModel struct {
	Type types.String                `tfsdk:"type"`
	NS   []singleNameserverDataModel `tfsdk:"ns"`
}

func getValidNameserverTypes() []string {
	return []string{"authoritative", "secondary", "konsoleh"}
}

func (d *nameserverDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nameserver"
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

func (d *nameserverDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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

func (d *nameserverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *nameserverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var (
		data        nameserverDataSourceModel
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
		nameservers = api.GetAuthoritativeNameservers()
	case "secondary":
		nameservers = api.GetSecondaryNameservers()
	case "konsoleh":
		nameservers = api.GetKonsolehNameservers()
	default:
		resp.Diagnostics.AddError(
			"Invalid nameserver type",
			fmt.Sprintf("Invalid nameserver type: %s, must be one of %s",
				data.Type.String(),
				strings.Join(getValidNameserverTypes(), ", "),
			),
		)
	}

	// Populate the data model
	data.NS = make([]singleNameserverDataModel, 0, len(nameservers))
	populateNameserverData(&data.NS, nameservers)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
