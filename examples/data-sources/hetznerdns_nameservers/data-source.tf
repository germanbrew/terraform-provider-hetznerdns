data "hetznerdns_nameservers" "authoritative" {
  type = "authoritative"
}

# Not specifying the type will default to authoritative like above
data "hetznerdns_nameservers" "primary" {}

resource "hetznerdns_record" "mydomain_de-NS" {
  for_each = toset(data.hetznerdns_nameservers.primary.ns.*.name)

  zone_id = hetznerdns_zone.de.id
  name    = "@"
  type    = "NS"
  value   = each.value
}
