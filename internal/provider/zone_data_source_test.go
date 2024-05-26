package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccZoneDataSource(t *testing.T) {
	aZoneName := acctest.RandString(10) + ".online"
	aZoneTTL := 60

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccZoneDataSourceConfig(aZoneName, aZoneTTL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hetznerdns_zone.zone1", "id"),
					resource.TestCheckResourceAttr("data.hetznerdns_zone.zone1", "name", aZoneName),
					resource.TestCheckResourceAttr("data.hetznerdns_zone.zone1", "ttl", strconv.Itoa(aZoneTTL)),
				),
			},
		},
	})
}

func testAccZoneDataSourceConfig(name string, ttl int) string {
	return fmt.Sprintf(`
resource "hetznerdns_zone" "zone1" {
    name = %[1]q
    ttl  = %[2]d
}

data "hetznerdns_zone" "zone1" {
	name = hetznerdns_zone.zone1.name
}
`, name, ttl)
}
