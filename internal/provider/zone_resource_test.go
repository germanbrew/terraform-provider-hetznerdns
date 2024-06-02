package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccZone_Resource(t *testing.T) {
	aZoneName := acctest.RandString(10) + ".online"
	aZoneTTL := 60

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccZoneResourceConfig("test", aZoneName, aZoneTTL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("hetznerdns_zone.test", "id"),
					resource.TestCheckResourceAttr("hetznerdns_zone.test", "name", aZoneName),
					resource.TestCheckResourceAttr("hetznerdns_zone.test", "ttl", strconv.Itoa(aZoneTTL)),
					resource.TestCheckResourceAttrSet("hetznerdns_zone.test", "ns.#"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hetznerdns_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccZoneResourceConfig("test", aZoneName, aZoneTTL*2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetznerdns_zone.test", "ttl", strconv.Itoa(aZoneTTL*2)),
					resource.TestCheckResourceAttrSet("hetznerdns_zone.test", "ns.#"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccZone_Invalid(t *testing.T) {
	aZoneName := "-.de"
	aZoneTTL := 60

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config:      testAccZoneResourceConfig("test", aZoneName, aZoneTTL),
				ExpectError: regexp.MustCompile("422 Unprocessable Entity: invalid label"),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccZone_NoTLD(t *testing.T) {
	aZoneName := "de"
	aZoneTTL := 60

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config:      testAccZoneResourceConfig("test", aZoneName, aZoneTTL),
				ExpectError: regexp.MustCompile("Attribute name Name must be a valid domain with top level domain, got: de"),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccZone_ZoneExists(t *testing.T) {
	aZoneName := acctest.RandString(10) + ".online"
	aZoneTTL := 60

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccZoneResourceConfig("test", aZoneName, aZoneTTL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("hetznerdns_zone.test", "id"),
					resource.TestCheckResourceAttr("hetznerdns_zone.test", "name", aZoneName),
					resource.TestCheckResourceAttr("hetznerdns_zone.test", "ttl", strconv.Itoa(aZoneTTL)),
					resource.TestCheckResourceAttrSet("hetznerdns_zone.test", "ns.#"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hetznerdns_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Create same zone again
			{
				Config:      testAccZoneResourceConfig("test2", aZoneName, aZoneTTL),
				ExpectError: regexp.MustCompile(fmt.Sprintf("zone %q already exists", aZoneName)),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccZoneResourceConfig(resourceName string, name string, ttl int) string {
	return fmt.Sprintf(`
resource "hetznerdns_zone" "%[1]s" {
    name = %[2]q
    ttl  = %[3]d

    timeouts {
    	create = "5s"
    	delete = "5s"
    	read   = "5s"
    	update = "5s"
    }
}`, resourceName, name, ttl)
}
