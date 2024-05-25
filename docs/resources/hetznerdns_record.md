# hetznerdns_record Resource

Provides a Hetzner DNS Records resource to create, update and delete DNS Records.

## Example Usage

### Basic
```hcl
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
```

### TXT Records > 255 bytes

TXT Records with a length of more that 255 bytes/characters must be split, otherwise the ressource will always be recreated by the Hetzner DNS API.

```hcl
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
```

## Argument Reference

The following arguments are supported:

- `zone_id` - (Required, string) Id of the DNS zone to create
  the record in.

- `name` - (Required, string) Name of the DNS record to create.

- `value` - (Required, string) The value of the record (eg. 192.168.1.1).
  For TXT records with quoted values, the quotes have to be escaped in Terraform
  (eg. "v=spf1 include:\_spf.google.com ~all" is represented by
  "\\"v=spf1 include:\_spf.google.com ~all\\"" in Terraform).

- `type` - (Required, string) The type of the record.

- `ttl` - (Optional, int) Time to live of this record.

## Import

A Record can be imported using its `id`. Use the API to get all records of
a zone and then copy the id.

```
curl "https://dns.hetzner.com/api/v1/records" \
     -H "Auth-API-Token: $HETZNER_DNS_API_TOKEN" | jq .

{
  "records": [
    {
      "id": "3d60921a49eb384b6335766a",
      "type": "TXT",
      "name": "google._domainkey",
      "value": "\"anything:with:param\"",
      "zone_id": "rMu2waTJPbHr4",
      "created": "2020-08-18 19:11:02.237 +0000 UTC",
      "modified": "2020-08-28 19:51:41.275 +0000 UTC"
    },
    {
      "id": "ed2416cb6bc8a8055b22222",
      "type": "A",
      "name": "www",
      "value": "1.1.1.1",
      "zone_id": "rMu2waTJPbHr4",
      "created": "2020-08-27 20:55:38.745 +0000 UTC",
      "modified": "2020-08-27 20:55:38.745 +0000 UTC"
    }
  ]
}
```

The command used above was copied from Hetzer DNS API docs. `jq` is
used for formatting and is not required. Use the `id` to import a
record.

```
terraform import hetznerdns_record.dkim_1 ed2416cb6bc8a8055b22222
```
