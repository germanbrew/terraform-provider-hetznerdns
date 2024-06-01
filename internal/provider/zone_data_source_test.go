package provider

import (
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccZone_DataSource(t *testing.T) {
	aZoneName := acctest.RandString(10) + ".online"
	aZoneTTL := 60

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", aZoneName, aZoneTTL),
						testAccZoneDataSourceConfig(),
					}, "\n",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hetznerdns_zone.zone1", "id"),
					resource.TestCheckResourceAttr("data.hetznerdns_zone.zone1", "name", aZoneName),
					resource.TestCheckResourceAttr("data.hetznerdns_zone.zone1", "ttl", strconv.Itoa(aZoneTTL)),
					resource.TestCheckResourceAttrSet("data.hetznerdns_zone.zone1", "ns.#"),
				),
			},
		},
	})
}

func testAccZoneDataSourceConfig() string {
	return `data "hetznerdns_zone" "zone1" {
	name = hetznerdns_zone.test.name
}`
}
