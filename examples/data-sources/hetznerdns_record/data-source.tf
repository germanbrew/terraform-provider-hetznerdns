data "hetznerdns_zone" "zone1" {
  name = "zone1.online"
}

data "hetznerdns_record" "record1" {
  zone_id = data.hetznerdns_zone.zone1.id
}
