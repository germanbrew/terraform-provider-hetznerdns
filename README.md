# Terraform Provider for Hetzner DNS

[![Terraform](https://img.shields.io/badge/Terraform-844FBA.svg?style=for-the-badge&logo=Terraform&logoColor=white)](https://registry.terraform.io/providers/germanbrew/hetznerdns/latest)
[![OpenTofu](https://img.shields.io/badge/OpenTofu-FFDA18.svg?style=for-the-badge&logo=OpenTofu&logoColor=black)](https://github.com/opentofu/registry/blob/main/providers/g/germanbrew/hetznerdns.json)
[![GitHub Release](https://img.shields.io/github/v/release/germanbrew/terraform-provider-hetznerdns?sort=date&display_name=release&style=for-the-badge&logo=github&link=https%3A%2F%2Fgithub.com%2Fgermanbrew%2Fterraform-provider-hetznerdns%2Freleases%2Flatest)](https://github.com/germanbrew/terraform-provider-hetznerdns/releases/latest)
[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/germanbrew/terraform-provider-hetznerdns/test.yaml?branch=main&style=for-the-badge&logo=github&label=Tests&link=https%3A%2F%2Fgithub.com%2Fgermanbrew%2Fterraform-provider-hetznerdns%2Factions%2Fworkflows%2Ftest.yaml)](https://github.com/germanbrew/terraform-provider-hetznerdns/actions/workflows/test.yaml)

You can find resources and data sources [documentation](https://registry.terraform.io/providers/germanbrew/hetznerdns/latest/docs) there or [here](docs).

## Requirements

-   [Terraform](https://www.terraform.io/downloads.html) > v1.0

## Installing and Using this Plugin

You most likely want to download the provider from [Terraform Registry](https://registry.terraform.io/providers/germanbrew/hetznerdns/latest/docs).
The provider is also published in the [OpenTofu Registry](https://github.com/opentofu/registry/tree/main/providers/g/germanbrew).

### Migration Guide

If you previously used the `timohirt/hetznerdns` provider, you can easily replace the provider in your terraform state
by following our [migration guide in the provider documentation](https://registry.terraform.io/providers/germanbrew/hetznerdns/latest/docs/guides/migration-from-timohirt-hetznerdns).

### Using Provider from Terraform Registry (TF >= 1.0)

This provider is published and available there. If you want to use it, just
add the following to your `terraform.tf`:

```terraform
terraform {
  required_providers {
    hetznerdns = {
      source = "germanbrew/hetznerdns"
      version = "3.0.0"  # Replace with latest version
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

Assign the API token to `HETZNER_DNS_TOKEN` env variable.

```sh
export HETZNER_DNS_TOKEN=<your api token>
```

The provider uses this token, and you don't have to enter it anymore.

## Credits

This project is a continuation of [timohirt/terraform-provider-hetznerdns](https://github.com/timohirt/terraform-provider-hetznerdns)

## Development

### Requirements

- [Go](https://golang.org/) 1.21 (to build the provider plugin)
- [golangci-lint](https://github.com/golangci/golangci-lint) (to lint code)
- [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) (to generate registry documentation)

### Install and update development tools

Run the following command

```sh
make install-devtools
```

### Makefile Commands

Check the subcommands in our [Makefile](Makefile) for useful dev tools and scripts.

### Testing the provider locally

To test the provider locally:

1. Build the provider binary with `make build`
2. Create a new file `~/.terraform.rc` and point the provider to the absolute **directory** path of the binary file:
    ```hcl
    provider_installation {
        dev_overrides {
            "germanbrew/hetznerdns" = "/path/to/your/terraform-provider-hetznerdns/bin/"
        }
        direct {}
    }
    ```
3.  - Set the variable before running terraform commands:

    ```sh
    TF_CLI_CONFIG_FILE=~/.terraform.rc terraform plan
    ```

    - Or set the env variable `TF_CLI_CONFIG_FILE` and point it to `~/.terraform.rc`: e.g.

    ```sh
    export TF_CLI_CONFIG_FILE=~/.terraform.rc`
    ```

4. Now you can just use terraform normally. A warning will appear, that notifies you that you are using an provider override
    ```
    Warning: Provider development overrides are in effect
    ...
    ```
5. Unset the env variable if you don't want to use the local provider anymore:
    ```sh
    unset TF_CLI_CONFIG_FILE
    ```
