package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/germanbrew/terraform-provider-hetznerdns/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	ApiToken             types.String `tfsdk:"apitoken"`
	MaxRetries           types.Int64  `tfsdk:"max_retries"`
	HasTxtValueFormatter types.Bool   `tfsdk:"enable_txt_formatter"`
}

type providerClient struct {
	client               *api.Client
	hasTxtValueFormatter bool
}

func (p *hetznerDNSProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hetznerdns"
	resp.Version = p.version
}

func (p *hetznerDNSProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This providers helps you automate management of DNS zones and records at Hetzner DNS.",
		Attributes: map[string]schema.Attribute{
			"apitoken": schema.StringAttribute{
				Description: "The Hetzner DNS API token. You can pass it using the env variable `HETZNER_DNS_API_TOKEN` as well.",
				Optional:    true,
				Sensitive:   true,
			},
			"max_retries": schema.Int64Attribute{
				Description: "The maximum number of retries to perform when an API request fails. " +
					"You can pass it using the env variable `HETZNER_DNS_MAX_RETRIES` as well. Default: 1",
				Optional: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"enable_txt_formatter": schema.BoolAttribute{
				Description: "Toggles the automatic formatter for TXT record values. " +
					"Values greater than 255 bytes needs to be chunked and quotes separetly. " +
					"You can pass it using the env variable `HETZNER_DNS_ENABLE_TXT_FORMATTER` as well. Default: true",
				Optional: true,
			},
		},
	}
}

// Configure configures the provider.
//
//nolint:funlen // TODO: The attributes logic should be moved to a separate function
func (p *hetznerDNSProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data hetznerDNSProviderModel

	apiToken := os.Getenv("HETZNER_DNS_API_TOKEN")

	maxRetries := int64(1)
	hasTxtValueFormatter := true

	if v, ok := os.LookupEnv("HETZNER_DNS_MAX_RETRIES"); ok {
		var err error

		maxRetries, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"max_retries must be an positive integer",
				"While configuring the provider, the max_retry option was not a positive integer",
			)
		}
	}

	if v, ok := os.LookupEnv("HETZNER_DNS_ENABLE_TXT_FORMATTER"); ok {
		var err error

		hasTxtValueFormatter, err = strconv.ParseBool(v)
		if err != nil {
			resp.Diagnostics.AddError(
				"enable_txt_formatter must be a boolean",
				"While configuring the provider, the enable_txt_formatter option was not a boolean value",
			)
		}
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ApiToken.ValueString() != "" {
		apiToken = data.ApiToken.ValueString()
	}

	if data.MaxRetries.ValueInt64Pointer() != nil {
		maxRetries = data.MaxRetries.ValueInt64()
	}

	if data.HasTxtValueFormatter.ValueBool() {
		hasTxtValueFormatter = data.HasTxtValueFormatter.ValueBool()
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

	client, err := api.New("https://dns.hetzner.com", apiToken, uint(maxRetries), http.DefaultClient)
	if err != nil {
		resp.Diagnostics.AddError(
			"API error while configuring provider",
			fmt.Sprintf("Error while creating API client: %s", err),
		)
	}

	client.SetUserAgent(fmt.Sprintf("terraform-provider-hetznerdns/%s (+https://github.com/germanbrew/terraform-provider-hetznerdns) ", p.version))

	provider := &providerClient{client: client, hasTxtValueFormatter: hasTxtValueFormatter}

	_, err = provider.client.GetZones(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"API error while configuring provider",
			fmt.Sprintf("Error while fetching zones: %s", err),
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.DataSourceData = provider
	resp.ResourceData = provider
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
