package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccExampleResource(t *testing.T) {
	aZoneName := fmt.Sprintf("%s.online", acctest.RandString(10))
	aZoneTTL := 60

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccExampleResourceConfig(aZoneName, aZoneTTL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("hetznerdns_zone.test", "id"),
					resource.TestCheckResourceAttr("hetznerdns_zone.test", "name", aZoneName),
					resource.TestCheckResourceAttr("hetznerdns_zone.test", "ttl", strconv.Itoa(aZoneTTL)),
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
				Config: testAccExampleResourceConfig(aZoneName, aZoneTTL*2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetznerdns_zone.test", "ttl", strconv.Itoa(aZoneTTL*2)),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccExampleResourceConfig(name string, ttl int) string {
	return fmt.Sprintf(`
resource "hetznerdns_zone" "test" {
  name = %[1]q
  ttl = %[2]d
}
`, name, ttl)
}
