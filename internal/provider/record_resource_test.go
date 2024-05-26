package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRecordResources(t *testing.T) {
	aZoneName := acctest.RandString(10) + ".online"
	aZoneTTL := 60

	aValue := "192.168.1.1"
	aName := acctest.RandString(10)
	aType := "A"
	aTTL := aZoneTTL * 2

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRecordResourceConfigCreate(aZoneName, aZoneTTL, aName, aType, aValue, aTTL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"hetznerdns_record.record1", "id"),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "type", aType),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "name", aName),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "value", aValue),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "ttl", strconv.Itoa(aTTL)),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hetznerdns_record.record1",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccRecordResourceConfigCreate(aZoneName, aZoneTTL, aName, aType, aValue, aTTL*2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "ttl", strconv.Itoa(aTTL*2)),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccRecordResourceConfigCreate(aZoneName string, aZoneTTL int, aName string, aType string, aValue string, aTTL int) string {
	return fmt.Sprintf(`
resource "hetznerdns_zone" "zone1" {
    name = %[1]q
    ttl  = %[2]d
}

resource "hetznerdns_record" "record1" {
	zone_id = hetznerdns_zone.zone1.id
	type    = "%s"
	name    = "%s"
	value   = "%s"
	ttl     = %d
}
`, aZoneName, aZoneTTL, aType, aName, aValue, aTTL)
}

func TestAccRecordWithDefaultTTLResources(t *testing.T) {
	// aZoneName must be a valid DNS domain name with an existing TLD
	aZoneName := acctest.RandString(10) + ".online"
	aZoneTTL := 3600

	aValue := "192.168.1.1"
	aName := acctest.RandString(10)
	aType := "A"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccRecordResourceConfigCreateWithDefaultTTL(aZoneName, aZoneTTL, aName, aType, aValue),
				PreventDiskCleanup: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"hetznerdns_record.record1", "id"),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "type", aType),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "name", aName),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "value", aValue),
					resource.TestCheckResourceAttr("hetznerdns_record.record1", "ttl.#", "0"),
				),
			},
		},
	})
}

func testAccRecordResourceConfigCreateWithDefaultTTL(aZoneName string, aZoneTTL int, aName string, aType string, aValue string) string {
	return fmt.Sprintf(`
resource "hetznerdns_zone" "zone1" {
    name = %[1]q
    ttl  = %[2]d
}

resource "hetznerdns_record" "record1" {
	zone_id = hetznerdns_zone.zone1.id
	type    = "%s"
	name    = "%s"
	value   = "%s"
}
`, aZoneName, aZoneTTL, aType, aName, aValue)
}

func TestAccTwoRecordResources(t *testing.T) {
	// aZoneName must be a valid DNS domain name with an existing TLD
	aZoneName := acctest.RandString(10) + ".online"

	aValue := "192.168.1.1"
	anotherValue := "192.168.1.2"
	aName := acctest.RandString(10)
	anotherName := acctest.RandString(10)
	aType := "A"
	aTTL := 60

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccRecordResourceConfigCreateTwo(aZoneName, aName, anotherName, aType, aValue, anotherValue, aTTL),
				PreventDiskCleanup: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"hetznerdns_record.record1", "id"),
					resource.TestCheckResourceAttrSet(
						"hetznerdns_record.record2", "id"),
				),
			},
		},
	})
}

func testAccRecordResourceConfigCreateTwo(aZoneName string, aName string, anotherName string, aType string, aValue string, anotherValue string, aTTL int) string {
	return fmt.Sprintf(`
resource "hetznerdns_zone" "zone1" {
    name = %[1]q
    ttl  = %[2]d
}

resource "hetznerdns_record" "record1" {
	zone_id = hetznerdns_zone.zone1.id
	type    = "%s"
	name    = "%s"
	value   = "%s"
	ttl     = %d
}

resource "hetznerdns_record" "record2" {
	zone_id = hetznerdns_zone.zone1.id
	type    = "%s"
	name    = "%s"
	value   = "%s"
	ttl     = %d
}
`, aZoneName, aTTL, aType, aName, aValue, aTTL, aType, anotherName, anotherValue, aTTL)
}

func TestAccRecordResourcesDKIM(t *testing.T) {
	// aZoneName must be a valid DNS domain name with an existing TLD
	aZoneName := fmt.Sprintf("%s.online", acctest.RandString(10))
	aTTL := 60

	aValue := "v=DKIM1;t=s;p=MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEArBysLW4Gqogt/VBPNpHwzuX23R54CXI0wXXjiPfeg3XHBPOtDMLlOQ3IqHu4v7PlRXwXwOfvKuovFjBXQtoC4KrzXacRC7fdTYOfbBz4/GEWGNL49/GVkSBCJA4hqXPKTK11pztoFkFQa7O4mpi3x11/cKDFy+FXBZvsE8QnBjyLbmSvG31/LLTmp2lzuLyN7IXEZ31g7pHm88IG0wwP84x1PicdoTZTv1tigrDRgCMiiCC2nWQ8VMdnJu7oPuYBgvS0aE5xYckfIQWPuTZM8iDDl94sGO4ni75Ycx8vbFvy/GA9ylFF/TVwLwDhiibx6H3itywKpdaX700eYVtwjVyeqFSoUUwqfEFkfsuKozw6vAdobAZmZjbqjjf0x04rFImytVbQCAcn1k54XJEoc6ctIt5JrNBco8O0SXg6d5QHyfpbYX/U8HLTFxFvef8Chd7+IK6N7qekj7spGnpa7HFSLpji6zMNv5PM47tMIfOdfTNlzBetjSe/S7tO7FCL/2BuQWIQ7mHiP1AvG4XA05IAL9D81xvEMr70qmqIHS7ifRQ+DT2f/g7+u8piSzVr0JA2jy6sD0Zb9g4KyOgtXKDg1pzb78hcHjp144yHmNxIaKhtMtz00wKGobg5e2AKsvF+iBmWgufQYqIaKvXa4+X4H1YZjfqTgzwwBjckIN0CAwEAAQ=="
	aName := "dkim._domainkey"
	aType := "TXT"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccRecordResourceConfigCreateDKIM(aZoneName, aTTL, aName, aValue),
				PreventDiskCleanup: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"hetznerdns_record.record1", "id"),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "type", aType),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "name", aName),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "value", aValue),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "ttl", strconv.Itoa(aTTL)),
				),
			},
		},
	})
}

func testAccRecordResourceConfigCreateDKIM(aZoneName string, aTTL int, aName string, aValue string) string {
	return fmt.Sprintf(`
resource "hetznerdns_zone" "zone1" {
    name = %[1]q
    ttl  = %[2]d
}

resource "hetznerdns_record" "record1" {
	zone_id = hetznerdns_zone.zone1.id
	type 	= "TXT"
	name 	= "%s"
	value 	= "%s"
	ttl 	= %d
}
`, aZoneName, aTTL, aName, aValue, aTTL)
}
