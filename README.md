# Terraform `Citrix` Provider

Citrix has developed a custom Terraform provider for automating Citrix product deployments and configurations. Using [Terraform](https://www.terraform.io) with Citrix provider, you can manage your Citrix products via Infrastructure as Code, giving you higher efficiency and consistency on infrastructure management, as well as better reusability on infrastructure configuration.

## Table of Content
- [Terraform `Citrix` Provider](#terraform-citrix-provider)
  - [Table of Content](#table-of-content)
  - [Citrix Terraform Provider Documentation](#citrix-terraform-provider-documentation)
    - [Navigating the repository](#navigating-the-repository)
    - [Provider Configuration](#provider-configuration)
    - [Resource Configuration](#resource-configuration)
  - [Using Citrix Terraform provider for Citrix DaaS](#using-citrix-terraform-provider-for-citrix-daas)
    - [Install Terraform](#install-terraform)
    - [Configure your Citrix Terraform Provider](#configure-your-citrix-terraform-provider)
    - [Start writing Terraform for managing your Citrix DaaS site](#start-writing-terraform-for-managing-your-citrix-daas-site)
    - [Create a Zone in Citrix DaaS as the first step](#create-a-zone-in-citrix-daas-as-the-first-step)
  - [Frequently Asked Questions](#frequently-asked-questions)
  - [Attributions](#attributions)
  - [License](#license)

## Citrix Terraform Provider Documentation

### Navigating the repository

1. `internal` folder - Contains the following sub directories:
   * `provider` folder - Contains the Citrix provider implementation for Terraform
   * `daas` folder - Contains all the Citrix DaaS resources libraries that we support through Terraform.
   * `test` folder - Contains the Go tests for both `provider` and all `resources` that we have.
   * `util` folder - Contains general utility functions that can be reused.
2. `examples` folder - Contains the examples for users to use various Citrix resources e.g [zone](examples/resources/citrix_daas_zone) folder contains the resources.tf that illustrates how citrix_daas_zone resource can be used to create a DaaS Zone on target Citrix DaaS site. There are also examples for [Citrix provider](examples/provider) configuration for both Citrix Cloud customer and Citrix On-Premises customers. Users can use the examples as a starting point to configure their own Citrix Terraform script.
3. `docs` folder - [resources](docs/resources) - contains the documentation for all resource configurations supported through Terraform. Refer this to understand the properties, accepted values, and how various properties work together for each type of resource. 

### Provider Configuration

`provider.tf` contains the information on target DaaS site where you want to apply configuration.

Depending on whether its managing a Citrix Cloud site, or a Citrix On-Premises site, Citrix provider should be configured differently.

Example for On-Premises site:

```hcl
provider "citrix" {
    hostname      = "10.71.136.250"  # Optionally set with `CITRIX_HOSTNAME` environment variable.
    client_id     = "${var.domain_admin_id}"  # Optionally set with `CITRIX_CLIENT_ID` environment variable.
    client_secret = "${var.domain_admin_secret}"  # Optionally set with `CITRIX_CLIENT_SECRET` environment variable.
}
```

Example for Cloud site:

```hcl
provider "citrix" {
    region        = "US"  # Optionally set with `CITRIX_REGION` environment variable.
    environment   = "Production"  # Optionally set with `CITRIX_ENVIRONMENT` environment variable.
    customer_id   = "${var.customer_id}"  # Optionally set with `CITRIX_CUSTOMER_ID` environment variable.
    client_id     = "${var.api_key_clientId}"  # Optionally set with `CITRIX_CLIENT_ID` environment variable.
    client_secret = "${var.api_key_clientSecret}"  # Optionally set with `CITRIX_CLIENT_SECRET` environment variable.
}
```

You can also set `hostname` for cloud site to force override the Citrix DaaS service URL for a cloud customer. 

You can use environment variables as stated in the comments above. When running Go tests, always use environment variables so that no credentials or other sensitive information are checked-in to the code.

Below is a table to show the difference between On-Premises and Cloud provider configuration:

|              | Cloud                             | On-Premises                           |
|--------------|-----------------------------------|---------------------------------------|
| region       | `US` / `EU` / `AP-S` / `JP`       | :x:                                   |
| environment  | `Production` / `Staging`          | :x:                                   |
| customerId   | Cloud Customer Id                 | :x:                                   |
| hostname     | (Optional) Cloud DDC hostname     | On-Premises DDC Hostname / IP address |
| clientId     | Citrix Cloud API Key clientId     | Domain Admin Username                 |
| clientSecret | Citrix Cloud API Key clientSecret | Domain Admin Password                 |

### Resource Configuration

Resources.tf can be used to configure the desired state of the resources that you want to create and manage in your Citrix Services. The example below shows how you can configure a Citrix DaaS Zone in Citrix DaaS service in resource.tf.

**`citrix_daas_zone`**

```hcl
resource "citrix_daas_zone" "example-zone" {
    name                = "example-zone"
    description         = "zone example"
    metadata            = [
        {
            name    = "key1"
            value   = "value1"
        }
    ]
}
```

Please refer the Citrix Terraform Provider documentation such as [docs/resources/daas_zone.md](docs/resources/daas_zone.md) to find out the configurable properties of each type of resources, understand what they do, and what option values are supported.

---------

## Using Citrix Terraform provider for Citrix DaaS

### Install Terraform

Refer the [Hashicorp documentation](https://learn.hashicorp.com/tutorials/terraform/install-cli) for installing Terraform CLI for your own environment.

### Configure your Citrix Terraform Provider

Refer section [Understanding Provider Configuration](#understanding-provider-configuration) or [Provider documentation](docs/index.md) to configure Citrix Terraform Provider for the Citrix DaaS site you want to manage with Terraform.

### Start writing Terraform for managing your Citrix DaaS site

To find all the Citrix DaaS resources manageable via Terraform, understand all the configurable properties for each resource and how they work together, refer documentations for DaaS resources that has `daas_` as resource name prefix in [Citrix Terraform resource documentation](docs/resources). To better understand how the resource is managed via Citrix DaaS Rest API, you can refer the [Citrix DaaS Rest API documentation](https://developer.cloud.com/citrixworkspace/citrix-daas/citrix-daas-rest-apis/docs/overview).

### Create a Zone in Citrix DaaS as the first step

Refer the [DaaS Zone documentation](docs/resources/daas_zone.md) to configure a zone via terraform. 

## Frequently Asked Questions


## Attributions
The code in this repository makes use of the following packages:
-	Hashicorp Terraform Plugin Framework (https://github.com/hashicorp/terraform-plugin-framework)
-	Google Go Uuid (https://github.com/google/uuid)

## License 
This project is Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0 

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

<sub>Copyright © 2023. Citrix Systems, Inc. All Rights Reserved.</sub>