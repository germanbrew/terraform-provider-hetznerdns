package main

import (
	"github.com/germanbrew/terraform-provider-hetznerdns/hetznerdns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	plugin "github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return hetznerdns.Provider()
		},
	})
}
