## Simple Example

resource "hetznerdns_zone" "example_com" {
  name = "example.com"
  ttl  = 3600
}

## DNS Zone Delegation

# Subdomain Zone
resource "hetznerdns_zone" "subdomain_example_com" {
  name = "subdomain.example.com"
  ttl  = 300
}

# Primary Domain Zone
resource "hetznerdns_zone" "example_com" {
  name = "example.com"
  ttl  = 300
}

# Nameserver Records for the Subdomain
## This block dynamically creates NS records in the primary domain zone to delegate authority to the subdomain.
resource "hetznerdns_record" "example_com-NS" {
  for_each = toset(hetznerdns_zone.mydomain_de.ns)

  zone_id = hetznerdns_zone.example_com.id
  name    = "@"
  type    = "NS"
  value   = each.value
}
