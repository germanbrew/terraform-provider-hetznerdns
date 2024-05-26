resource "hetznerdns_zone" "zone1" {
  name = provider::hetznerdns::idna("bücher.example.com")
  ttl  = 3600
}
