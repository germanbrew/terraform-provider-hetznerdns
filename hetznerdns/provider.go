package hetznerdns

import (
	"context"

	"github.com/germanbrew/terraform-provider-hetznerdns/hetznerdns/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider creates and return a Terraform resource provider
// for Hetzer DNS
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"apitoken": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HETZNER_DNS_API_TOKEN", nil),
				Description: "The API access token to authenticate at Hetzner DNS API.",
			},
			"max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HETZNER_DNS_MAX_RETRIES", 10),
				Description: "The maximum number of retries to perform when an API request fails.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"hetznerdns_zone":           resourceZone(),
			"hetznerdns_record":         resourceRecord(),
			"hetznerdns_primary_server": resourcePrimaryServer(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"hetznerdns_zone": dataSourceHetznerDNSZone(),
		},
		ConfigureContextFunc: configureProvider,
	}
}

func configureProvider(c context.Context, r *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return api.NewClient(r.Get("apitoken").(string), r.Get("max_retries").(int))
}
