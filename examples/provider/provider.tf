provider "hetznerdns" {
  api_token = ""
}

data "hetznerdns_zone" "dns_zone" {
  name = "example.com"
}

data "hcloud_server" "web" {
  name = "web1"
}

resource "hetznerdns_record" "web" {
  zone_id = data.hetznerdns_zone.dns_zone.id
  name    = "www"
  value   = hcloud_server.web.ipv4_address
  type    = "A"
  ttl     = 60
}
