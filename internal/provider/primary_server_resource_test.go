package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPrimaryServer_OnePrimaryServersResources(t *testing.T) {
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
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", aZoneName, aZoneTTL),
						testAccPrimaryServerResourceConfigCreate("test", psAddress, psPort),
					}, "\n",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("hetznerdns_primary_server.test", "id"),
					resource.TestCheckResourceAttr("hetznerdns_primary_server.test", "address", psAddress),
					resource.TestCheckResourceAttr("hetznerdns_primary_server.test", "port", strconv.Itoa(psPort)),
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
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", aZoneName, aZoneTTL),
						testAccPrimaryServerResourceConfigCreate("test", psAddress, psPort*2),
					}, "\n",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetznerdns_primary_server.test", "port", strconv.Itoa(psPort*2)),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func TestAccPrimaryServer_Invalid(t *testing.T) {
	aZoneName := acctest.RandString(10) + ".online"
	aZoneTTL := 3600

	psAddress := "-"
	psPort := 53

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", aZoneName, aZoneTTL),
						testAccPrimaryServerResourceConfigCreate("test", psAddress, psPort),
					}, "\n",
				),
				ExpectError: regexp.MustCompile("422 Unprocessable Entity"),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccPrimaryServer_TwoPrimaryServersResources(t *testing.T) {
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
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", aZoneName, aZoneTTL),
						testAccPrimaryServerResourceConfigCreate("ps1", ps1Address, ps1Port),
						testAccPrimaryServerResourceConfigCreate("ps2", ps2Address, ps2Port),
					}, "\n",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("hetznerdns_primary_server.ps1", "id"),
					resource.TestCheckResourceAttrSet("hetznerdns_primary_server.ps2", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccPrimaryServerResourceConfigCreate(resourceName, psAddress string, psPort int) string {
	return fmt.Sprintf(`
resource "hetznerdns_primary_server" "%s" {
	zone_id = hetznerdns_zone.test.id
	address = %q
	port    = %d
}
`, resourceName, psAddress, psPort)
}
