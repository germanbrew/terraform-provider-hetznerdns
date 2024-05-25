resource "hetznerdns_zone" "zone1" {
  name = "zone1.online"
  ttl  = 3600
}

resource "hetznerdns_primary_server" "ps1" {
  zone_id = hetznerdns_zone.zone1.id
  address = "1.1.1.1"
  port    = 53
}