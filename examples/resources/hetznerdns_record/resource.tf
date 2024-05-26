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