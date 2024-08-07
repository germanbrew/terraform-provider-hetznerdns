---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "hetznerdns_records Data Source - hetznerdns"
subcategory: ""
description: |-
  Provides details about all Records of a Hetzner DNS Zone
---

# hetznerdns_records (Data Source)

Provides details about all Records of a Hetzner DNS Zone



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `zone_id` (String) ID of the DNS zone to get records from

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `records` (Attributes List) The DNS records of the zone (see [below for nested schema](#nestedatt--records))

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) [Operation Timeouts](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts) consisting of
numbers and unit suffixes, such as "30s" or "2h45m".
Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Default: 5m


<a id="nestedatt--records"></a>
### Nested Schema for `records`

Read-Only:

- `id` (String) ID of this DNS record
- `name` (String) Name of this DNS record
- `ttl` (Number) Time to live of this record
- `type` (String) Type of this DNS record
- `value` (String) Value of this DNS record
- `zone_id` (String) ID of the DNS zone
