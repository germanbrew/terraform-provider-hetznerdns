package provider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"testing"

	"github.com/germanbrew/terraform-provider-hetznerdns/internal/api"
	"github.com/germanbrew/terraform-provider-hetznerdns/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
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

func TestAccZone_StaleZone(t *testing.T) {
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
			// Remove zone from Hetzner DNS and check if it will be recreated by Terraform
			{
				PreConfig: func() {
					var (
						data      hetznerDNSProviderModel
						apiToken  string
						apiClient *api.Client
						err       error
					)

					apiToken = utils.ConfigureStringAttribute(data.ApiToken, "HETZNER_DNS_API_TOKEN", "")
					httpClient := logging.NewLoggingHTTPTransport(http.DefaultTransport)
					apiClient, err = api.New("https://dns.hetzner.com", apiToken, httpClient)
					if err != nil {
						t.Fatalf("Error while creating API apiClient: %s", err)
					}
					zone, err := apiClient.GetZoneByName(context.Background(), aZoneName)
					if err != nil {
						t.Fatalf("Error while fetching zone: %s", err)
					}
					err = apiClient.DeleteZone(context.Background(), zone.ID)
					if err != nil {
						t.Fatalf("Error while deleting zone: %s", err)
					}
				},
				// Check if the zone is recreated
				// ExpectNonEmptyPlan: true,
				RefreshState: true,
				ExpectError:  regexp.MustCompile("hetznerdns_zone.test will be created"),
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
