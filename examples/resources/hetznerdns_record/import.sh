# A Record can be imported using its `id`. Use the API to get all records of
# a zone and then copy the id.
#
# curl "https://dns.hetzner.com/api/v1/records" \
#      -H "Auth-API-Token: $HETZNER_DNS_API_TOKEN" | jq .
#
# {
#   "records": [
#     {
#       "id": "3d60921a49eb384b6335766a",
#       "type": "TXT",
#       "name": "google._domainkey",
#       "value": "\"anything:with:param\"",
#       "zone_id": "rMu2waTJPbHr4",
#       "created": "2020-08-18 19:11:02.237 +0000 UTC",
#       "modified": "2020-08-28 19:51:41.275 +0000 UTC"
#     },
#     {
#       "id": "ed2416cb6bc8a8055b22222",
#       "type": "A",
#       "name": "www",
#       "value": "1.1.1.1",
#       "zone_id": "rMu2waTJPbHr4",
#       "created": "2020-08-27 20:55:38.745 +0000 UTC",
#       "modified": "2020-08-27 20:55:38.745 +0000 UTC"
#     }
#   ]
# }

terraform import hetznerdns_record.dkim_1 ed2416cb6bc8a8055b22222