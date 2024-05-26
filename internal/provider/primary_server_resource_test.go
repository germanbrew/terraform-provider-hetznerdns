package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOnePrimaryServersResources(t *testing.T) {
	aZoneName := acctest.RandString(10) + ".online"
	aZoneTTL := 3600

	psAddress := "1.1.0.0"
	psPort := 53

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPrimaryServerResourceConfigCreate(aZoneName, aZoneTTL, psAddress, psPort),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"hetznerdns_primary_server.test", "id"),
					resource.TestCheckResourceAttr(
						"hetznerdns_primary_server.test", "address", psAddress),
					resource.TestCheckResourceAttr(
						"hetznerdns_primary_server.test", "port", strconv.Itoa(psPort)),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hetznerdns_primary_server.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccPrimaryServerResourceConfigCreate(aZoneName, aZoneTTL, psAddress, psPort*2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetznerdns_primary_server.test", "port", strconv.Itoa(psPort*2)),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccPrimaryServerResourceConfigCreate(aZoneName string, aZoneTTL int, psAddress string, psPort int) string {
	return fmt.Sprintf(`
resource "hetznerdns_zone" "zone1" {
    name = %[1]q
    ttl  = %[2]d
}

resource "hetznerdns_primary_server" "test" {
	zone_id = hetznerdns_zone.zone1.id
	address = "%s"
	port    = %d
}
`, aZoneName, aZoneTTL, psAddress, psPort)
}

func TestAccTwoPrimaryServersResources(t *testing.T) {
	aZoneName := acctest.RandString(10) + ".online"
	aZoneTTL := 3600

	ps1Address := "1.1.0.0"
	ps1Port := 53

	ps2Address := "2.2.0.0"
	ps2Port := 53

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPrimaryServerResourceConfigCreateTwo(aZoneName, aZoneTTL, ps1Address, ps1Port, ps2Address, ps2Port),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("hetznerdns_primary_server.ps1", "id"),
					resource.TestCheckResourceAttrSet("hetznerdns_primary_server.ps2", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccPrimaryServerResourceConfigCreateTwo(aZoneName string, aTTL int, ps1Address string, ps1Port int, ps2Address string, ps2Port int) string {
	return fmt.Sprintf(`
resource "hetznerdns_zone" "zone1" {
    name = %[1]q
    ttl  = %[2]d
}

resource "hetznerdns_primary_server" "ps1" {
	zone_id = hetznerdns_zone.zone1.id
	address = "%s"
	port    = %d
}

resource "hetznerdns_primary_server" "ps2" {
	zone_id = hetznerdns_zone.zone1.id
	address = "%s"
	port    = %d
}
`, aZoneName, aTTL, ps1Address, ps1Port, ps2Address, ps2Port)
}
