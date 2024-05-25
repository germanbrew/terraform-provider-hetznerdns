# Basic Usage

data "hetznerdns_zone" "some_domain" {
  name = "some-domain.com"
}

resource "hetznerdns_record" "www" {
  zone_id = data.hetznerdns_zone.some_domain.id
  name    = "www"
  type    = "A"
  value   = "192.168.1.1"
  ttl     = 60
}


# TXT Records > 255 bytes

# TXT Records with a length of more that 255 bytes/characters must be split, otherwise the resource will always be
# recreated by the Hetzner DNS API.

locals {
  example_dkim = "v=DKIM1;h=sha256;k=rsa;s=email;p=MIIBIjAN..."
}

resource "hetznerdns_record" "dkim" {
  zone_id = data.hetznerdns_zone.some_domain.id
  name    = "example._domainkey"
  type    = "TXT"
  value   = join("\"", ["", replace(local.example_dkim, "/(.{255})/", "$1\" \""), " "])
  ttl     = 60
}