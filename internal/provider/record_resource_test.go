package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/germanbrew/terraform-provider-hetznerdns/internal/api"
	"github.com/germanbrew/terraform-provider-hetznerdns/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRecord_Resources(t *testing.T) {
	zoneName := acctest.RandString(10) + ".online"
	aZoneTTL := 60

	value := "192.168.1.1"
	aName := acctest.RandString(10)
	aType := "A"
	ttl := aZoneTTL * 2

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", zoneName, aZoneTTL),
						testAccRecordResourceConfigWithTTL("record1", aName, aType, value, ttl),
					}, "\n",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"hetznerdns_record.record1", "id"),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "type", aType),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "name", aName),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "value", value),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "ttl", strconv.Itoa(ttl)),
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
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", zoneName, aZoneTTL),
						testAccRecordResourceConfigWithTTL("record1", aName, aType, value, ttl*2),
					}, "\n",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "ttl", strconv.Itoa(ttl*2)),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccRecord_ResourcesWithDeprecatedApiToken(t *testing.T) {
	zoneName := acctest.RandString(10) + ".online"
	aZoneTTL := 60

	value := "192.168.1.1"
	aName := acctest.RandString(10)
	aType := "A"
	ttl := aZoneTTL * 2

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				// Unset new API token and set deprecated API token instead
				PreConfig: func() {
					apiToken := os.Getenv("HETZNER_DNS_TOKEN")
					err := os.Setenv("HETZNER_DNS_API_TOKEN", apiToken)
					if err != nil {
						t.Errorf("Error while setting HETZNER_DNS_API_TOKEN: %s", err)
					}

					err = os.Unsetenv("HETZNER_DNS_TOKEN")
					if err != nil {
						t.Errorf("Error while unsetting HETZNER_DNS_TOKEN: %s", err)
					}
				},
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", zoneName, aZoneTTL),
						testAccRecordResourceConfigWithTTL("record1", aName, aType, value, ttl),
					}, "\n",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"hetznerdns_record.record1", "id"),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "type", aType),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "name", aName),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "value", value),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "ttl", strconv.Itoa(ttl)),
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
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", zoneName, aZoneTTL),
						testAccRecordResourceConfigWithTTL("record1", aName, aType, value, ttl*2),
					}, "\n",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "ttl", strconv.Itoa(ttl*2)),
				),
			},
			{
				// Undo the changes to the API token
				PreConfig: func() {
					apiToken := os.Getenv("HETZNER_DNS_API_TOKEN")
					err := os.Setenv("HETZNER_DNS_TOKEN", apiToken)
					if err != nil {
						t.Errorf("Error while setting HETZNER_DNS_TOKEN: %s", err)
					}

					err = os.Unsetenv("HETZNER_DNS_API_TOKEN")
					if err != nil {
						t.Errorf("Error while unsetting HETZNER_DNS_API_TOKEN: %s", err)
					}
				},
				RefreshState: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccRecord_Invalid(t *testing.T) {
	zoneName := acctest.RandString(10) + ".online"
	aZoneTTL := 60

	value := "-"
	aName := acctest.RandString(10)
	aType := "A"
	ttl := aZoneTTL * 2

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", zoneName, aZoneTTL),
						testAccRecordResourceConfigWithTTL("record1", aName, aType, value, ttl),
					}, "\n",
				),
				ExpectError: regexp.MustCompile("invalid A record"),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccRecord_WithDefaultTTLResources(t *testing.T) {
	// zoneName must be a valid DNS domain name with an existing TLD
	zoneName := acctest.RandString(10) + ".online"
	aZoneTTL := 3600

	value := "192.168.1.1"
	aName := acctest.RandString(10)
	aType := "A"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", zoneName, aZoneTTL),
						testAccRecordResourceConfig("record1", aName, aType, value),
					}, "\n",
				),
				PreventDiskCleanup: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hetznerdns_record.record1", "id"),
					resource.TestCheckResourceAttr("hetznerdns_record.record1", "type", aType),
					resource.TestCheckResourceAttr("hetznerdns_record.record1", "name", aName),
					resource.TestCheckResourceAttr("hetznerdns_record.record1", "value", value),
					resource.TestCheckResourceAttr("hetznerdns_record.record1", "ttl.#", "0"),
				),
			},
		},
	})
}

func TestAccRecord_TwoRecordResources(t *testing.T) {
	// zoneName must be a valid DNS domain name with an existing TLD
	zoneName := acctest.RandString(10) + ".online"

	value := "192.168.1.1"
	anotherValue := "192.168.1.2"
	aName := acctest.RandString(10)
	anotherName := acctest.RandString(10)
	aType := "A"
	ttl := 60

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", zoneName, ttl),
						testAccRecordResourceConfigWithTTL("record1", aName, aType, value, ttl),
						testAccRecordResourceConfigWithTTL("record2", anotherName, aType, anotherValue, ttl),
					}, "\n",
				),
				PreventDiskCleanup: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hetznerdns_record.record1", "id"),
					resource.TestCheckResourceAttr("hetznerdns_record.record1", "type", aType),
					resource.TestCheckResourceAttr("hetznerdns_record.record1", "name", aName),
					resource.TestCheckResourceAttr("hetznerdns_record.record1", "value", value),
					resource.TestCheckResourceAttrSet("hetznerdns_record.record2", "id"),
					resource.TestCheckResourceAttr("hetznerdns_record.record2", "type", aType),
					resource.TestCheckResourceAttr("hetznerdns_record.record2", "name", anotherName),
					resource.TestCheckResourceAttr("hetznerdns_record.record2", "value", anotherValue),
				),
			},
		},
	})
}

func TestAccRecord_ResourcesDKIM(t *testing.T) {
	// zoneName must be a valid DNS domain name with an existing TLD
	zoneName := acctest.RandString(10) + ".online"
	ttl := 60

	value := "v=DKIM1;t=s;p=MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEArBysLW4Gqogt/VBPNpHwzuX23R54CXI0wXXjiPfeg3XHBPOtDMLlOQ3IqHu4v7PlRXwXwOfvKuo" +
		"vFjBXQtoC4KrzXacRC7fdTYOfbBz4/GEWGNL49/GVkSBCJA4hqXPKTK11pztoFkFQa7O4mpi3x11/cKDFy+FXBZvsE8QnBjyLbmSvG31/LLTmp2lzuLyN7IXEZ31g7pHm88IG0wwP84" +
		"x1PicdoTZTv1tigrDRgCMiiCC2nWQ8VMdnJu7oPuYBgvS0aE5xYckfIQWPuTZM8iDDl94sGO4ni75Ycx8vbFvy/GA9ylFF/TVwLwDhiibx6H3itywKpdaX700eYVtwjVyeqFSoUUwqf" +
		"EFkfsuKozw6vAdobAZmZjbqjjf0x04rFImytVbQCAcn1k54XJEoc6ctIt5JrNBco8O0SXg6d5QHyfpbYX/U8HLTFxFvef8Chd7+IK6N7qekj7spGnpa7HFSLpji6zMNv5PM47tMIfOd" +
		"fTNlzBetjSe/S7tO7FCL/2BuQWIQ7mHiP1AvG4XA05IAL9D81xvEMr70qmqIHS7ifRQ+DT2f/g7+u8piSzVr0JA2jy6sD0Zb9g4KyOgtXKDg1pzb78hcHjp144yHmNxIaKhtMtz00wK" +
		"Gobg5e2AKsvF+iBmWgufQYqIaKvXa4+X4H1YZjfqTgzwwBjckIN0CAwEAAQ=="
	aName := "dkim._domainkey"
	aType := "TXT"

	aRandomValue := acctest.RandString(522)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", zoneName, ttl),
						testAccRecordResourceConfigWithTTL("record1", aName, aType, value, ttl),
					}, "\n",
				),
				PreventDiskCleanup: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("hetznerdns_record.record1", "id"),
					resource.TestCheckResourceAttr("hetznerdns_record.record1", "type", aType),
					resource.TestCheckResourceAttr("hetznerdns_record.record1", "name", aName),
					resource.TestCheckResourceAttr("hetznerdns_record.record1", "value", value),
					resource.TestCheckResourceAttr("hetznerdns_record.record1", "ttl", strconv.Itoa(ttl)),
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
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", zoneName, ttl),
						testAccRecordResourceConfigWithTTL("record1", aName, aType, aRandomValue, ttl),
					}, "\n",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hetznerdns_record.record1", "value", aRandomValue),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hetznerdns_record.record1",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccRecord_StaleResources(t *testing.T) {
	zoneName := acctest.RandString(10) + ".online"
	aZoneTTL := 60

	value := "192.168.1.1"
	aName := acctest.RandString(10)
	aType := "A"
	ttl := aZoneTTL * 2

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", zoneName, aZoneTTL),
						testAccRecordResourceConfigWithTTL("record1", aName, aType, value, ttl),
					}, "\n",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"hetznerdns_record.record1", "id"),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "type", aType),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "name", aName),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "value", value),
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "ttl", strconv.Itoa(ttl)),
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
				Config: strings.Join(
					[]string{
						testAccZoneResourceConfig("test", zoneName, aZoneTTL),
						testAccRecordResourceConfigWithTTL("record1", aName, aType, value, ttl*2),
					}, "\n",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"hetznerdns_record.record1", "ttl", strconv.Itoa(ttl*2)),
				),
			},
			// Remove record from Hetzner DNS and check if it will be recreated by Terraform
			{
				PreConfig: func() {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					var (
						data      hetznerDNSProviderModel
						apiToken  string
						apiClient *api.Client
						err       error
					)

					apiToken = utils.ConfigureStringAttribute(data.ApiToken, "HETZNER_DNS_TOKEN", "")
					httpClient := logging.NewLoggingHTTPTransport(http.DefaultTransport)
					apiClient, err = api.New("https://dns.hetzner.com", apiToken, httpClient)
					if err != nil {
						t.Fatalf("Error while creating API apiClient: %s", err)
					}
					zone, err := apiClient.GetZoneByName(ctx, zoneName)
					if err != nil {
						t.Fatalf("Error while fetching zone: %s", err)
					}
					record, err := apiClient.GetRecordByName(ctx, zone.ID, aName)
					if err != nil {
						t.Fatalf("Error while fetching record: %s", err)
					}
					err = apiClient.DeleteRecord(ctx, record.ID)
					if err != nil {
						t.Fatalf("Error while deleting record: %s", err)
					}
				},
				// Check if the record is recreated
				// ExpectNonEmptyPlan: true,
				RefreshState: true,
				ExpectError:  regexp.MustCompile("hetznerdns_record.record1 will be created"),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccRecordResourceConfigWithTTL(resourceName, name, recordType, value string, ttl int) string {
	return fmt.Sprintf(`
resource "hetznerdns_record" "%s" {
	zone_id = hetznerdns_zone.test.id
	name    = %q
	type    = %q
	value   = %q
	ttl     = %d

    timeouts {
    	create = "5s"
    	delete = "5s"
    	read   = "5s"
    	update = "5s"
    }
}`, resourceName, name, recordType, value, ttl)
}

func testAccRecordResourceConfig(resourceName, name, recordType, value string) string {
	return fmt.Sprintf(`
resource "hetznerdns_record" "%s" {
	zone_id = hetznerdns_zone.test.id
	name    = %q
	type    = %q
	value   = %q

    timeouts {
    	create = "5s"
    	delete = "5s"
    	read   = "5s"
    	update = "5s"
    }
}`, resourceName, name, recordType, value)
}
