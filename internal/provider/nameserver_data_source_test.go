package provider

import (
	"strings"
	"testing"

	"github.com/germanbrew/terraform-provider-hetznerdns/internal/api"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

func TestAccNameservers_DataSource(t *testing.T) {
	authorizedNameservers := api.GetAuthoritativeNameservers()
	nsNames := make([]knownvalue.Check, len(authorizedNameservers))

	for i, ns := range authorizedNameservers {
		nsNames[i] = knownvalue.StringExact(ns["name"])
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: strings.Join(
					[]string{
						testAccNameserversDataSourceConfig(),
					}, "\n",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetznerdns_nameservers.primary", "ns.0.name", authorizedNameservers[0]["name"]),
					resource.TestCheckResourceAttrSet("data.hetznerdns_nameservers.primary", "ns.#"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("primary_names", knownvalue.ListExact(nsNames)),
				},
			},
		},
	})
}

func testAccNameserversDataSourceConfig() string {
	return `
data "hetznerdns_nameservers" "primary" {
	type = "authoritative"
}

data "hetznerdns_nameservers" "secondary" {
	type = "secondary"
}

data "hetznerdns_nameservers" "konsoleh" {
	type = "konsoleh"
}

output "primary_names" {
	value = data.hetznerdns_nameservers.primary.ns.*.name
}
`
}
