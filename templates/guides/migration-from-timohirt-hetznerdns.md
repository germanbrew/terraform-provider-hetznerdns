---
subcategory: ""
layout: "hetznerdns"
page_title: "Migration from timohirt/hetznerdns"
description: |-
    A Guide on how to migrate your Terraform State from timohirt/hetznerdns to germanbrew/hetznerdns
---

# How to migrate from timohirt/hetznerdns to germanbrew/hetznerdns

If you previously used the `timohirt/hetznerdns` provider, you can easily replace the provider in your terraform state by following these steps:

~> **NOTE:** It is recommended to backup your Terraform state before migrating by running this command: `terraform state pull > terraform-state-backup.json`

## Migration Steps

1. In your `terraform` -> `required_providers` config, replace the provider config:

    ```diff
    hetznerdns = {
    -  source  = "timohirt/hetznerdns"
    +  source = "germanbrew/hetznerdns"
    -  version = "2.2.0"
    +  version = "3.0.0"  # Replace with latest version
    }
    ```

2. If you have `apitoken` defined inside you provider config, replace it with `api_token`. The environment variable is now called `HETZNER_DNS_TOKEN` instead of `HETZNER_DNS_API_TOKEN`.
   Also see our [Docs Overview](https://registry.terraform.io/providers/germanbrew/hetznerdns/latest/docs#schema), as we have more configuration options for you to choose.

    ```diff
    provider "hetznerdns" {"
    -  apitoken  = "token"
    +  api_token = "token"
    }
    ```

3. Install the new provider and replace it in the state:

    ```sh
    terraform init
    terraform state replace-provider timohirt/hetznerdns germanbrew/hetznerdns
    ```

4. Our provider automatically reformats TXT record values into the correct format ([RFC4408](https://datatracker.ietf.org/doc/html/rfc4408#section-3.1.3)).
   This means you don't need to escape the values yourself with `jsonencode()` or other functions to split the records every 255 bytes.
   You can disable this feature by specifying `enable_txt_formatter = false` in your provider config or setting the env var `HETZNER_DNS_ENABLE_TXT_FORMATTER=false`

5. Test if the migration was successful by running `terraform plan` and checking the output for any errors.
