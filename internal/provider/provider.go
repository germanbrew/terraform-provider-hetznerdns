package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/germanbrew/terraform-provider-hetznerdns/internal/api"
	"github.com/germanbrew/terraform-provider-hetznerdns/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var (
	_ provider.Provider              = &hetznerDNSProvider{}
	_ provider.ProviderWithFunctions = &hetznerDNSProvider{}
)

type hetznerDNSProvider struct {
	version string
}

type hetznerDNSProviderModel struct {
	ApiToken           types.String `tfsdk:"api_token"`
	MaxRetries         types.Int64  `tfsdk:"max_retries"`
	EnableTxtFormatter types.Bool   `tfsdk:"enable_txt_formatter"`
}

type providerClient struct {
	apiClient    *api.Client
	maxRetries   int64
	txtFormatter bool
}

func (p *hetznerDNSProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hetznerdns"
	resp.Version = p.version
}

func (p *hetznerDNSProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This providers helps you automate management of DNS zones and records at Hetzner DNS.",
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				Description: "The Hetzner DNS API token. You can pass it using the env variable `HETZNER_DNS_API_TOKEN` as well.",
				Optional:    true,
				Sensitive:   true,
			},
			"max_retries": schema.Int64Attribute{
				Description: "`Default: 1` The maximum number of retries to perform when an API request fails. " +
					"You can pass it using the env variable `HETZNER_DNS_MAX_RETRIES` as well.",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"enable_txt_formatter": schema.BoolAttribute{
				Description: "`Default: true` Toggles the automatic formatter for TXT record values. " +
					"Values greater than 255 bytes get split into multiple quoted chunks " +
					"([RFC4408](https://datatracker.ietf.org/doc/html/rfc4408#section-3.1.3)). " +
					"You can pass it using the env variable `HETZNER_DNS_ENABLE_TXT_FORMATTER` as well.",
				Optional: true,
			},
		},
	}
}

// Configure configures the provider.
func (p *hetznerDNSProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var (
		data     hetznerDNSProviderModel
		apiToken string
		err      error
	)

	client := &providerClient{}

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiToken = utils.ConfigureStringAttribute(data.ApiToken, "HETZNER_DNS_API_TOKEN", "")
	if apiToken == "" {
		resp.Diagnostics.AddError("Missing API Token Configuration",
			"While configuring the client, the API token was not found in the HETZNER_DNS_API_TOKEN environment variable or client configuration block"+
				"api_token attribute.",
		)
	}

	client.maxRetries, err = utils.ConfigureInt64Attribute(data.MaxRetries, "HETZNER_DNS_MAX_RETRIES", 1)
	if err != nil {
		resp.Diagnostics.AddError("max_retries must be an integer", err.Error())
	}

	client.txtFormatter, err = utils.ConfigureBoolAttribute(data.EnableTxtFormatter, "HETZNER_DNS_MAX_RETRIES", true)
	if err != nil {
		resp.Diagnostics.AddError("enable_txt_formatter must be a boolean", err.Error())
	}

	if resp.Diagnostics.HasError() {
		return
	}

	httpClient := logging.NewLoggingHTTPTransport(http.DefaultTransport)

	client.apiClient, err = api.New("https://dns.hetzner.com", apiToken, httpClient)
	if err != nil {
		resp.Diagnostics.AddError("API error while configuring client", fmt.Sprintf("Error while creating API apiClient: %s", err))

		return
	}

	client.apiClient.SetUserAgent(fmt.Sprintf("terraform-client-hetznerdns/%s (+https://github.com/germanbrew/terraform-client-hetznerdns) ", p.version))

	if _, err = client.apiClient.GetZones(ctx); err != nil {
		resp.Diagnostics.AddError("API error while configuring client", fmt.Sprintf("Error while fetching zones: %s", err))

		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *hetznerDNSProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPrimaryServerResource,
		NewRecordResource,
		NewZoneResource,
	}
}

func (p *hetznerDNSProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewZoneDataSource,
		NewRecordsDataSource,
	}
}

func (p *hetznerDNSProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{
		NewIdnaFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &hetznerDNSProvider{
			version: version,
		}
	}
}
