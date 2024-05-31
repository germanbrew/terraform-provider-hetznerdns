package provider

import (
	"regexp"
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
						testAccRecordsDataSourceConfig(),
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
					}),
					resource.TestMatchTypeSetElemNestedAttrs("data.hetznerdns_records.test", "records.*", map[string]*regexp.Regexp{
						"zone_id": regexp.MustCompile(`^\S+$`),
						"id":      regexp.MustCompile(`^\S+$`),
						"name":    regexp.MustCompile(annotherName),
						"value":   regexp.MustCompile(annotherValue),
						"type":    regexp.MustCompile(annotherType),
					}),
				),
			},
		},
	})
}

func testAccRecordsDataSourceConfig() string {
	return `data "hetznerdns_records" "test" {
	zone_id = hetznerdns_zone.test.id

	depends_on = [hetznerdns_record.record1, hetznerdns_record.record2]
}`
}
