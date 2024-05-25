package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/germanbrew/terraform-provider-hetznerdns/internal/api"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &hetznerDNSProvider{}
var _ provider.ProviderWithFunctions = &hetznerDNSProvider{}

type hetznerDNSProvider struct {
	version string
}

type hetznerDNSProviderModel struct {
	ApiToken types.String `tfsdk:"apitoken"`
}

func (p *hetznerDNSProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hetznerdns"
	resp.Version = p.version
}

func (p *hetznerDNSProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"apitoken": schema.StringAttribute{
				Description: "The Hetzner DNS API token. You can pass it using the env variable `HETZNER_DNS_API_TOKEN`as well.",
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *hetznerDNSProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data hetznerDNSProviderModel

	apiToken := os.Getenv("HETZNER_DNS_API_TOKEN")

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ApiToken.ValueString() != "" {
		apiToken = data.ApiToken.ValueString()
	}

	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing API Token Configuration",
			"While configuring the provider, the API token was not found in "+
				"the HETZNER_DNS_API_TOKEN environment variable or provider "+
				"configuration block apitoken attribute.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client := api.NewClient(apiToken)

	_, err := client.GetZones()
	if err != nil {
		resp.Diagnostics.AddError(
			"API error while configuring provider",
			err.Error(),
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *hetznerDNSProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource {
			return NewZoneResource()
		},
	}
}

func (p *hetznerDNSProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource {
			return dataSourceExample{}
		},
	}
}

func (p *hetznerDNSProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &hetznerDNSProvider{
			version: version,
		}
	}
}