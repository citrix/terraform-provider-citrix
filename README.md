# Plugin for Terraform Provider for Citrix®

Citrix has developed a custom Terraform provider for automating Citrix product deployments and configurations. Using [Terraform](https://www.terraform.io) with Citrix provider, you can manage your Citrix products via Infrastructure as Code, giving you higher efficiency and consistency on infrastructure management, as well as better reusability on infrastructure configuration. The provider is developed and maintained by Citrix. Please note that this provider is still in tech preview.

## Table of Content
- [Plugin for Terraform Provider for Citrix®](#plugin-for-terraform-provider-for-citrix)
  - [Table of Content](#table-of-content)
  - [Contacting the Maintainers](#contacting-the-maintainers)
  - [Plugin for Terraform Provider for Citrix® Documentation](#plugin-for-terraform-provider-for-citrix-documentation)
    - [Navigating the repository](#navigating-the-repository)
    - [Provider Configuration](#provider-configuration)
    - [Resource Configuration](#resource-configuration)
  - [Using the Plugin for Terraform Provider for Citrix DaaS™](#using-the-plugin-for-terraform-provider-for-citrix-daas)
    - [Install Terraform](#install-terraform)
    - [(On-Premises Only) Enable Web Studio](#on-premises-only-enable-web-studio)
    - [Configure your Plugin for Terraform Provider for Citrix DaaS™](#configure-your-plugin-for-terraform-provider-for-citrix-daas)
    - [Start writing Terraform for managing your Citrix DaaS site](#start-writing-terraform-for-managing-your-citrix-daas-site)
    - [Create a Zone in Citrix DaaS as the first step](#create-a-zone-in-citrix-daas-as-the-first-step)
    - [Create a Hypervisor](#create-a-hypervisor)
    - [Create a Hypervisor Resource Pool](#create-a-hypervisor-resource-pool)
    - [Create a Machine Catalog](#create-a-machine-catalog)
    - [Create a Delivery Group](#create-a-delivery-group)
  - [Frequently Asked Questions](#frequently-asked-questions)
      - [What resource is supported for different connection types?](#what-resource-is-supported-for-different-connection-types)
      - [What provisioning types are supported for machine catalog?](#what-provisioning-types-are-supported-for-machine-catalog)
  - [Attributions](#attributions)
  - [License](#license)

## Contacting the Maintainers
This project uses GitHub to track all issues. See [CONTRIBUTING.md](CONTRIBUTING.md) for more information.

The CUGC (Citrix User Group Community) maintains a combination Slack/Discord for communication. Follow [this link](https://mycugc.org/resources/slack-discord/) to sign up. Provider discussion can be posted to the #community-support_automation channel.

## Plugin for Terraform Provider for Citrix® Documentation

### Navigating the repository

1. `internal` folder - Contains the following sub directories:
   * `provider` folder - Contains the Citrix provider implementation for Terraform
   * `daas` folder - Contains all the Citrix DaaS resources libraries that we support through Terraform.
   * `test` folder - Contains the Go tests for both `provider` and all `resources` that we have.
   * `util` folder - Contains general utility functions that can be reused.
2. `examples` folder - Contains the examples for users to use various Citrix resources e.g [zone](examples/resources/citrix_zone) folder contains the resources.tf that illustrates how citrix_zone resource can be used to create a DaaS Zone on target Citrix DaaS site. There are also examples for [Citrix provider](examples/provider) configuration for both Citrix Cloud customer and Citrix on-premises customers. Users can use the examples as a starting point to configure their own Citrix Terraform script.
3. `docs` folder - [resources](docs/resources) - contains the documentation for all resource configurations supported through Terraform. Refer this to understand the properties, accepted values, and how various properties work together for each type of resource. 

### Provider Configuration

`provider.tf` contains the information on target DaaS site where you want to apply configuration.

Depending on whether its managing a Citrix Cloud site, or a Citrix on-premises site, Citrix provider should be configured differently.

Example for on-premises site:

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

Below is a table to show the difference between on-premises and Cloud provider configuration:

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

**`citrix_zone`**

```hcl
resource "citrix_zone" "example-zone" {
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

Please refer the Plugin for Terraform Provider for Citrix DaaS™ documentation such as [docs/resources/zone.md](docs/resources/zone.md) to find out the configurable properties of each type of resources, understand what they do, and what option values are supported.

---------

## Using the Plugin for Terraform Provider for Citrix DaaS™

### Install Terraform

Refer the [Hashicorp documentation](https://learn.hashicorp.com/tutorials/terraform/install-cli) for installing Terraform CLI for your own environment.

### (On-Premises Only) Enable Web Studio

For on-premises sites with version >= 2308 are supported. Web Studio needs to be [installed and configured](https://docs.citrix.com/en-us/citrix-virtual-apps-desktops/install-configure/install-core/install-web-studio.html#install-web-studio-1) for the provider to work.

### Configure your Plugin for Terraform Provider for Citrix DaaS™

Refer section [Understanding Provider Configuration](#understanding-provider-configuration) or [Provider documentation](docs/index.md) to configure the provider for the Citrix DaaS site you want to manage with Terraform.

### Start writing Terraform for managing your Citrix DaaS site

To find all the Citrix DaaS resources manageable via Terraform, understand all the configurable properties for each resource and how they work together, refer documentations for DaaS resources that has `daas_` as resource name prefix in [Citrix Terraform resource documentation](docs/resources). To better understand how the resource is managed via Citrix DaaS Rest API, you can refer the [Citrix DaaS Rest API documentation](https://developer.cloud.com/citrixworkspace/citrix-daas/citrix-daas-rest-apis/docs/overview).

### Create a Zone in Citrix DaaS as the first step

Refer the [DaaS Zone documentation](docs/resources/zone.md) to configure a zone via terraform. 

### Create a Hypervisor

Hypervisor is needed to use your preferred public cloud provider with Citrix DaaS. Refer the [DaaS Hypervisor documentation](docs/resources/azure_hypervisor.md) to configure an Azure hypervisor in a zone via terraform.

### Create a Hypervisor Resource Pool

The hypervisor resource pool defines the network configuration for a hypervisor connection. Refer the [DaaS Hypervisor Resource Pool documentaion](docs/resources/hypervisor_resource_pool.md) to configure a hypervisr resource pool via terraform.

### Create a Machine Catalog

A machine catalog is a collection of machines managed as a single entity. Refer the [DaaS Machine Catalog documentation](docs/resources/machine_catalog.md) to configure a machine catalog via terraform.

### Create a Delivery Group
A delivery group is a collection of machines selected from one or more machine catalogs. The delivery group can also specify which users can use those machines, plus the applications and desktops available to those users. Refer the [DaaS Delivery Group documentation](docs/resources/delivery_group.md) to configure a delivery group via terraform.

## Frequently Asked Questions

#### What resource is supported for different connection types?

| Connection Type |   Hypervisor     |   Resource Pool  |   Machine Catalog   | 
|-----------------|------------------|------------------|---------------------|
| AzureRM         |:heavy_check_mark:|:heavy_check_mark:| MCS / Power Managed |
| AWS EC2         |:heavy_check_mark:|:heavy_check_mark:| MCS / Power Managed |
| GCP             |:heavy_check_mark:|:heavy_check_mark:| MCS / Power Managed |
| Vsphere         |:heavy_check_mark:|:heavy_check_mark:|    Power Managed    |
| XenServer       |:heavy_check_mark:|:heavy_check_mark:|    Power Managed    |
| Nutanix         |:heavy_check_mark:|   in progress    |    Power Managed    |


#### What provisioning types are supported for machine catalog? 
- MCS provisioning
  - Azure
  - GCP
  - AWS EC2
- Manual Power Managed
  - Azure
  - GCP
  - AWS EC2
  - Vsphere
  - XenServer
  - Nutanix
- Manual / Remote PC

## Attributions
The code in this repository makes use of the following packages:
-	Hashicorp Terraform Plugin Framework (https://github.com/hashicorp/terraform-plugin-framework)
-	Google Go Uuid (https://github.com/google/uuid)

## License 
This project is Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0 

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

<sub>Copyright © 2023. Citrix Systems, Inc.</sub>