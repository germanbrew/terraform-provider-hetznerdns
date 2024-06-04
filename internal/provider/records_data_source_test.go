package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRecords_DataSource(t *testing.T) {
	aZoneName := acctest.RandString(10) + ".online"
	aZoneTTL := 60

	aValue := "192.168.1.1"
	aName := acctest.RandString(10)
	aType := "A"
	anotherValue := acctest.RandString(200)
	anotherName := acctest.RandString(10)
	anotherType := "TXT"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", aZoneName, aZoneTTL),
						testAccRecordResourceConfig("record1", aName, aType, aValue),
						testAccRecordResourceConfig("record2", anotherName, anotherType, anotherValue),
						testAccRecords_DataSourceConfig(),
					}, "\n",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hetznerdns_records.test", "records.#", "3"),
					resource.TestMatchTypeSetElemNestedAttrs("data.hetznerdns_records.test", "records.*", map[string]*regexp.Regexp{
						"zone_id": regexp.MustCompile(`^\S+$`),
						"id":      regexp.MustCompile(`^\S+$`),
						"name":    regexp.MustCompile("@"),
						"value":   regexp.MustCompile(`[a-z.]+ [a-z.]+ [\d ]+`),
						"type":    regexp.MustCompile("SOA"),
					}),
					resource.TestMatchTypeSetElemNestedAttrs("data.hetznerdns_records.test", "records.*", map[string]*regexp.Regexp{
						"zone_id": regexp.MustCompile(`^\S+$`),
						"id":      regexp.MustCompile(`^\S+$`),
						"name":    regexp.MustCompile(aName),
						"value":   regexp.MustCompile(aValue),
						"type":    regexp.MustCompile(aType),
						"fqdn":    regexp.MustCompile(fmt.Sprintf("^%s.%s$", aName, aZoneName)),
					}),
					resource.TestMatchTypeSetElemNestedAttrs("data.hetznerdns_records.test", "records.*", map[string]*regexp.Regexp{
						"zone_id": regexp.MustCompile(`^\S+$`),
						"id":      regexp.MustCompile(`^\S+$`),
						"name":    regexp.MustCompile(anotherName),
						"value":   regexp.MustCompile(anotherValue),
						"type":    regexp.MustCompile(anotherType),
						"fqdn":    regexp.MustCompile(fmt.Sprintf("^%s.%s$", anotherName, aZoneName)),
					}),
				),
			},
		},
	})
}

func testAccRecords_DataSourceConfig() string {
	return `data "hetznerdns_records" "test" {
	zone_id = hetznerdns_zone.test.id

	depends_on = [hetznerdns_record.record1, hetznerdns_record.record2]
}`
}
