---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "hetznerdns_record Resource - hetznerdns"
subcategory: ""
description: |-
  Provides a Hetzner DNS Zone resource to create, update and delete DNS Zones.
---

# hetznerdns_record (Resource)

Provides a Hetzner DNS Zone resource to create, update and delete DNS Zones.

## Example Usage

```terraform
# Basic Usage

data "hetznerdns_zone" "example" {
  name = "example.com"
}

# Handle root (example.com)
resource "hetznerdns_record" "example_com_root" {
  zone_id = data.hetznerdns_zone.example.id
  name    = "@"
  value   = "1.2.3.4"
  type    = "A"
  # You only need to set a TTL if it's different from the zone's TTL above
  ttl = 300
}

# Handle wildcard subdomain (*.example.com)
resource "hetznerdns_record" "all_example_com" {
  zone_id = data.hetznerdns_zone.example.id
  name    = "*"
  value   = "1.2.3.4"
  type    = "A"
}

# Handle specific subdomain (books.example.com)
resource "hetznerdns_record" "books_example_com" {
  zone_id = data.hetznerdns_zone.example.id
  name    = "books"
  value   = "1.2.3.4"
  type    = "A"
}

# Handle email (MX record with priority 10)
resource "hetznerdns_record" "example_com_email" {
  zone_id = data.hetznerdns_zone.example.id
  name    = "@"
  value   = "10 mail.example.com"
  type    = "MX"
}

# SPF record
resource "hetznerdns_record" "example_com_spf" {
  zone_id = data.hetznerdns_zone.example.id
  name    = "@"
  value   = "v=spf1 ip4:1.2.3.4 -all"
  type    = "TXT"
}

# SRV record
resource "hetznerdns_record" "example_com_srv" {
  zone_id = data.hetznerdns_zone.example.id
  name    = "_ldap._tcp"
  value   = "10 0 389 ldap.example.com."
  type    = "SRV"
  ttl     = 3600
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the DNS record to create
- `type` (String) Type of this DNS record ([See supported types](https://docs.hetzner.com/dns-console/dns/general/supported-dns-record-types/))
- `value` (String) The value of the record (e.g. `192.168.1.1`)
- `zone_id` (String) ID of the DNS zone to create the record in.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `ttl` (Number) Time to live of this record

### Read-Only

- `id` (String) Zone identifier

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) [Operation Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) consisting of
numbers and unit suffixes, such as "30s" or "2h45m".
Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Default: 5m
- `delete` (String) [Operation Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) consisting of
numbers and unit suffixes, such as "30s" or "2h45m".
Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Default: 5m
- `read` (String) [Operation Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) consisting of
numbers and unit suffixes, such as "30s" or "2h45m".
Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Default: 5m
- `update` (String) [Operation Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) consisting of
numbers and unit suffixes, such as "30s" or "2h45m".
Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Default: 5m

## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
# A Record can be imported using its `id`. Use the API to get all records of
# a zone and then copy the id.
#
# curl "https://dns.hetzner.com/api/v1/records" \
#      -H "Auth-API-Token: $HETZNER_DNS_TOKEN" | jq .
#
# {
#   "records": [
#     {
#       "id": "3d60921a49eb384b6335766a",
#       "type": "TXT",
#       "name": "google._domainkey",
#       "value": "\"anything:with:param\"",
#       "zone_id": "rMu2waTJPbHr4",
#       "created": "2020-08-18 19:11:02.237 +0000 UTC",
#       "modified": "2020-08-28 19:51:41.275 +0000 UTC"
#     },
#     {
#       "id": "ed2416cb6bc8a8055b22222",
#       "type": "A",
#       "name": "www",
#       "value": "1.1.1.1",
#       "zone_id": "rMu2waTJPbHr4",
#       "created": "2020-08-27 20:55:38.745 +0000 UTC",
#       "modified": "2020-08-27 20:55:38.745 +0000 UTC"
#     }
#   ]
# }

terraform import hetznerdns_record.dkim_google 3d60921a49eb384b6335766a
```
