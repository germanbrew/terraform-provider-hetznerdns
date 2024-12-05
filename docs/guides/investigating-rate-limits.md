---
subcategory: ""
layout: "hetznerdns"
page_title: "Investigating Rate Limits"
description: |-
    A Guide on how to investigate and resolve rate limit issues with the Hetzner DNS API
---

# How to investigate and resolve rate limit issues with the Hetzner DNS API

Hetzner DNS API has a default rate limit of 300 requests (See [Hetzner Cloud Docs](https://docs.hetzner.cloud/#rate-limiting)) per minute. If you're getting a rate limit error, you can investigate and resolve it by following these steps:

1. If you're getting a rate limit error like below, try to increase the provider config [`max_retries`](https://registry.terraform.io/providers/germanbrew/hetznerdns/latest/docs#max_retries-1) to a higher value like `10`:
    ```bash
    Error: API Error
    read record: error getting record 3c21...75fb: API returned HTTP 429 Too Many Requests error: rate limit exceeded
    ```

2. You can view the ratelimit usage in the terraform logs by running terraform plan or apply with the `TF_LOG` environment variable set to `DEBUG`:
    ```bash
    TF_LOG=DEBUG terraform apply
    ```
    The http client will then log the entire http response including all headers. In the headers you will find rate limit details:

   | Header Name                  | Example Value |
   |------------------------------|---------------|
   | x-ratelimit-remaining-minute | 296           |
   | x-ratelimit-limit-minute     | 300           |
   | ratelimit-remaining          | 296           |
   | ratelimit-limit              | 300           |
   | ratelimit-reset              | 50 (seconds)  |

3. If you need to increase the rate limit, you can contact Hetzner Support to request a higher rate limit for your account.
