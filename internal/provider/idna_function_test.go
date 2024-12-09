package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestIdnaFunction_Valid(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
                output "domain" {
                    value = provider::hetznerdns::idna("bücher.example.com")
                }

				output "emoji_domain" {
					value = provider::hetznerdns::idna("😂😂👍.com")
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("domain", knownvalue.StringExact("xn--bcher-kva.example.com")),
					statecheck.ExpectKnownOutputValue("emoji_domain", knownvalue.StringExact("xn--yp8hj1aa.com")),
				},
			},
		},
	})
}
