package provider

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRecordsDataSource(t *testing.T) {
	aZoneName := acctest.RandString(10) + ".online"
	aZoneTTL := 60

	aValue := "192.168.1.1"
	aName := acctest.RandString(10)
	aType := "A"
	annotherValue := "Hello World"
	annotherName := acctest.RandString(10)
	annotherType := "TXT"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig(aZoneName, aZoneTTL),
						testAccRecordResourceConfig("record1", aName, aType, aValue),
						testAccRecordResourceConfig("record2", annotherName, annotherType, annotherValue),
						testAccRecordsDataSourceConfig(aZoneName),
					}, "\n",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.hetznerdns_records.records1.0", "id"),
					resource.TestCheckResourceAttr("data.hetznerdns_records.records1.0", "name", aZoneName),
					resource.TestCheckResourceAttrSet("data.hetznerdns_records.records1.0", "value"),
					resource.TestCheckResourceAttr("data.hetznerdns_records.records1.0", "ttl", strconv.Itoa(aZoneTTL)),
				),
			},
		},
	})
}

func testAccRecordsDataSourceConfig(zoneName string) string {
	return fmt.Sprintf(`data "hetznerdns_records" "records1" {
	zone_id = hetznerdns_zone.%s.id
}`, zoneName)
}
