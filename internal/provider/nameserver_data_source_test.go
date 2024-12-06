package provider

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/germanbrew/terraform-provider-hetznerdns/internal/api"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

// prepareKnownValues prepares a list of known values for the given nameserver list.
func prepareKnownValues(nameserver []string) []knownvalue.Check {
	knownValues := make([]knownvalue.Check, len(nameserver))
	// Sort the list to ensure the order is consistent and matches the output of the data source.
	sort.Strings(nameserver)

	for i, ip := range nameserver {
		knownValues[i] = knownvalue.StringExact(ip)
	}

	return knownValues
}

func TestAccNameservers_DataSource(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	authorizedNameservers, err := api.GetAuthoritativeNameservers(ctx)
	if err != nil {
		t.Fatalf("error fetching authoritative nameservers: %s", err)
	}

	nsNames := make([]knownvalue.Check, len(authorizedNameservers))

	for i, ns := range authorizedNameservers {
		nsNames[i] = knownvalue.StringExact(ns["name"])
	}

	// The IPv4 addresses of the authoritative nameservers to check against.
	// https://docs.hetzner.com/dns-console/dns/general/authoritative-name-servers/#new-name-servers-for-robot-and-cloud-console-customers
	authorizedNsIPv4 := []string{
		"193.47.99.5",
		"213.133.100.98",
		"88.198.229.192",
	}

	// The IPv6 addresses of the secondary nameservers to check against.
	// https://docs.hetzner.com/dns-console/dns/general/authoritative-name-servers/#secondary-dns-servers-old-name-servers-for-robot-customers
	secondaryNsIPv6 := []string{
		"2a01:4f8:0:a101::a:1",
		"2a01:4f8:0:1::5ddc:2",
		"2001:67c:192c::add:a3",
	}

	authorizedIPv4 := prepareKnownValues(authorizedNsIPv4)
	secondaryIPv6 := prepareKnownValues(secondaryNsIPv6)

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
					statecheck.ExpectKnownOutputValue("primary_ipv4s", knownvalue.ListExact(authorizedIPv4)),
					statecheck.ExpectKnownOutputValue("secondary_ipv6s", knownvalue.ListExact(secondaryIPv6)),
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

output "primary_ipv4s" {
	value = data.hetznerdns_nameservers.primary.ns.*.ipv4
}

output "secondary_ipv6s" {
	value = data.hetznerdns_nameservers.secondary.ns.*.ipv6
}
`
}
