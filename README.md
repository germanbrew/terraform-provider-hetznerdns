# Terraform Provider for Hetzner DNS

[![Tests](https://github.com/germanbrew/terraform-provider-hetznerdns/actions/workflows/test.yaml/badge.svg)](https://github.com/germanbrew/terraform-provider-hetznerdns/actions/workflows/test.yaml)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/germanbrew/terraform-provider-hetznerdns)
![GitHub](https://img.shields.io/github/license/germanbrew/terraform-provider-hetznerdns)

**This provider is published on the [Terraform](https://registry.terraform.io/providers/germanbrew/hetznerdns/latest) and [OpenTofu Registry](https://github.com/opentofu/registry/tree/main/providers/g/germanbrew)**.

You can find resources and data sources [documentation](https://registry.terraform.io/providers/germanbrew/hetznerdns/latest/docs) there or [here](docs).

> This project has been forked from [timohirt/terraform-provider-hetznerdns](https://github.com/timohirt/terraform-provider-hetznerdns), which is no longer maintained.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) > v1.0
- [Go](https://golang.org/) 1.21 (to build the provider plugin)

## Installing and Using this Plugin

You most likely want to download the provider from [Terraform Registry](https://registry.terraform.io/providers/germanbrew/hetznerdns/latest/docs).
The provider is also published in the [OpenTofu Registry](https://github.com/opentofu/registry/tree/main/providers/g/germanbrew).

### Migration Guide

If you previously used the `timohirt/hetznerdns` provider, you can easily replace the provider in your terraform state by following these steps:

1. In your `terraform` -> `required_providers` config, replace the provider config:
   ```diff
   hetznerdns = {
   -  source  = "timohirt/hetznerdns"
   +  source = "germanbrew/hetznerdns"
   -  version = "2.2.0"
   +  version = "3.0.0"  # Replace with latest version
   }
   ```
2. If you have `apitoken` defined inside you provider config, replace it with `api_token`. The environment variable `HETZNER_DNS_API_TOKEN` has not changed.
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

### Using Provider from Terraform Registry (TF >= 1.0)

This provider is published and available there. If you want to use it, just
add the following to your `terraform.tf`:

```terraform
terraform {
  required_providers {
    hetznerdns = {
      source = "germanbrew/hetznerdns"
      version = "3.0.0"
    }
  }
  required_version = ">= 1.0"
}
```

Then run `terraform init` to download the provider.

## Authentication

Once installed, you have three options to provide the required API token that
is used to authenticate at the Hetzner DNS API.

### Enter API Token when needed

You can enter it every time you run `terraform`.

### Configure the Provider to take the API Token from a Variable

Add the following to your `terraform.tf`:

```terraform
variable "hetznerdns_token" {}

provider "hetznerdns" {
  api_token = var.hetznerdns_token
}
```

Now, assign your API token to `hetznerdns_token` in `terraform.tfvars`:

```terraform
hetznerdns_token = "kkd993i3kkmm4m4m4"
```

You don't have to enter the API token anymore.

### Inject the API Token via the Environment

Assign the API token to `HETZNER_DNS_API_TOKEN` env variable.

```sh
export HETZNER_DNS_API_TOKEN=<your api token>
```

The provider uses this token, and you don't have to enter it anymore.

## Development

### Testing the provider locally

To test the provider locally:

1. Build the provider binary with `make build`
2. Create a new file `~/.terraform.rc` and point the provider to the absolute **directory** path of the binary file:
    ```json
    provider_installation {
        dev_overrides {
            "germanbrew/hetznerdns" = "/path/to/your/terraform-provider-hetznerdns/bin/"
        }
        direct {}
    }
    ```
3.
   - Set the variable before running terraform commands:
    ```sh
    TF_CLI_CONFIG_FILE=~/.terraform.rc terraform plan
    ```
   - Or set the env variable `TF_CLI_CONFIG_FILE` and point it to `~/.terraform.rc`: e.g.
    ```sh
    export TF_CLI_CONFIG_FILE=~/.terraform.rc`
    ```

1. Now you can just use terraform normally. A warning will appear, that notifies you that you are using an provider override
    ```
    Warning: Provider development overrides are in effect
    ...
    ```
2. Unset the env variable if you don't want to use the local provider anymore:
    ```sh
    unset TF_CLI_CONFIG_FILE
    ```
