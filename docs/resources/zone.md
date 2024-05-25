---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "hetznerdns_zone Resource - hetznerdns"
subcategory: ""
description: |-
  Provides a Hetzner DNS Zone resource to create, update and delete DNS Zones.
---

# hetznerdns_zone (Resource)

Provides a Hetzner DNS Zone resource to create, update and delete DNS Zones.

## Example Usage

```terraform
resource "hetznerdns_zone" "zone1" {
  name = "zone1.online"
  ttl  = 3600
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the DNS zone to create. Must be a valid domain with top level domain. Meaning `<domain>.de` or `<domain>.io`. Don't include sub domains on this level. So, no `sub.<domain>.io`. The Hetzner API rejects attempts to create a zone with a sub domain name.Use a record to create the sub domain.

### Optional

- `ttl` (Number) Time to live of this zone

### Read-Only

- `id` (String) Zone identifier

## Import

Import is supported using the following syntax:

```shell
terraform import hetznerdns_zone.zone1 rMu2waTJPbHr4
```