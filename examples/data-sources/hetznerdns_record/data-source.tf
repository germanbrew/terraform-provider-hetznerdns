data "hetznerdns_zone" "zone1" {
  name = "zone1.online"
}

data "hetznerdns_records" "zone1" {
  zone_id = data.hetznerdns_zone.zone1.id
}

output "zone1_records" {
  value = data.hetznerdns_records.zone1.records
}
