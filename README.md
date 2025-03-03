# Plugin for Terraform Provider for Citrix®

Citrix has developed a custom Terraform provider for automating Citrix product deployments and configurations. Using [Terraform](https://www.terraform.io) with Citrix provider, you can manage your Citrix products via Infrastructure as Code, giving you higher efficiency and consistency on infrastructure management, as well as better reusability on infrastructure configuration. The provider is developed and maintained by Citrix.

## Table of Contents
- [Plugin for Terraform Provider for Citrix®](#plugin-for-terraform-provider-for-citrix)
  - [Table of Contents](#table-of-contents)
  - [Contacting the Maintainers](#contacting-the-maintainers)
  - [Examples](#examples)
    - [Deployment guides](#deployment-guides)
    - [Demo video](#demo-video)
  - [Related Citrix Automation Repositories](#related-citrix-automation-repositories)
  - [Plugin for Terraform Provider for Citrix® Documentation](#plugin-for-terraform-provider-for-citrix-documentation)
    - [Navigating the repository](#navigating-the-repository)
    - [Provider Configuration](#provider-configuration)
    - [Resource Configuration](#resource-configuration)
  - [Using the Plugin for Terraform Provider for Citrix DaaS™](#using-the-plugin-for-terraform-provider-for-citrix-daas)
    - [Install Terraform](#install-terraform)
    - [(On-Premises Only) Enable Web Studio](#on-premises-only-enable-web-studio)
    - [(Cloud Only) Create a Citrix Cloud Service Principal](#cloud-only-create-a-citrix-cloud-service-principal)
    - [Configure your Plugin for Terraform Provider for Citrix DaaS™](#configure-your-plugin-for-terraform-provider-for-citrix-daas)
    - [Start writing Terraform for managing your Citrix DaaS site](#start-writing-terraform-for-managing-your-citrix-daas-site)
    - [Create a Zone in Citrix DaaS as the first step](#create-a-zone-in-citrix-daas-as-the-first-step)
    - [Create a Hypervisor](#create-a-hypervisor)
    - [Create a Hypervisor Resource Pool](#create-a-hypervisor-resource-pool)
    - [Create a Machine Catalog](#create-a-machine-catalog)
    - [Create a Delivery Group](#create-a-delivery-group)
  - [Using the Plugin for Terraform Provider for other Citrix resources](#using-the-plugin-for-terraform-provider-for-other-citrix-resources)
    - [Configure Global App Configuration (GAC) Settings](#configure-global-app-configuration-gac-settings)
    - [Create Citrix Cloud Resource Locations](#create-citrix-cloud-resource-locations)
    - [Managing StoreFront resources](#managing-storefront-resources)
    - [Managing DaaS Quick Deploy resources](#managing-daas-quick-deploy-resources)
  - [Frequently Asked Questions](#frequently-asked-questions)
      - [What resource is supported for different connection types?](#what-resource-is-supported-for-different-connection-types)
  - [Attributions](#attributions)
  - [License](#license)

## Contacting the Maintainers
This project uses GitHub to track all issues. See [CONTRIBUTING.md](CONTRIBUTING.md) for more information.

## Examples
Basic example templates for getting started can be found in the repository at [examples/](/examples/README.md)

### Deployment guides
Please refer to [Citrix Tech Zone](https://community.citrix.com/tech-zone/automation/) to find detailed guides on how to deploy and manage resources using the Citrix provider:
- [Installing and configuring the provider](https://community.citrix.com/tech-zone/automation/terraform-install-and-config/)
- [Daily administrative operations](https://community.citrix.com/tech-zone/automation/terraform-daily-administration/)
- [AWS EC2](https://community.citrix.com/tech-zone/build/deployment-guides/terraform-daas-aws/) via MCS
- [AWS WorkSpaces Core](https://community.citrix.com/tech-zone/learn/poc-guides/daas-and-awc-terraform)
- [Azure](https://community.citrix.com/tech-zone/build/deployment-guides/citrix-daas-terraform-azure/) via MCS
- [GCP](https://community.citrix.com/tech-zone/build/deployment-guides/terraform-daas-gcp/) via MCS
- [vSphere](https://community.citrix.com/tech-zone/build/deployment-guides/terraform-daas-vsphere8/) via MCS
- [XenServer](https://community.citrix.com/tech-zone/automation/citrix-terraform-xenserver) via MCS
- [Citrix policies](https://community.citrix.com/tech-zone/automation/cvad-terraform-policies/)

### Demo video
[![alt text](images/techzone-youtube-thumbnail.png)](https://www.youtube.com/watch?v=c33sMLaCVjY)

https://www.youtube.com/watch?v=c33sMLaCVjY

## Related Citrix Automation Repositories
|            Title            |            Details            |
|-----------------------------|-------------------------------|
<!-- | [Plugin for Terraform Provider for Citrix®](https://github.com/citrix/terraform-provider-citrix) | Terraform provider plugin to manage Citrix products including CVAD, DaaS, StoreFront, and WEM via Terraform IaC. | -->
| [Packer Image Management Module for Citrix® Virtual Apps and Desktops](https://github.com/citrix/citrix-packer-tools) | Use Packer to create golden images with the Citrix VDA installed and using Citrix best practices. |
| [Citrix Ansible Tools](https://github.com/citrix/citrix-ansible-tools) | Playbooks to install Citrix components using automation such as the VDA. |
| [Site Deployment Module for Citrix® Virtual Apps and Desktops](https://github.com/citrix/citrix-cvad-site-deployment-module) | Uses PowerShell to drive Terraform files to create a fully functional CVAD site. |

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
    cvad_config = {
      hostname      = "10.71.136.250"  # Optionally set with `CITRIX_HOSTNAME` environment variable.
      client_id     = "${var.domain_admin_id}"  # Optionally set with `CITRIX_CLIENT_ID` environment variable.
      client_secret = "${var.domain_admin_secret}"  # Optionally set with `CITRIX_CLIENT_SECRET` environment variable.
    }
}
```

Example for Cloud site:

```hcl
provider "citrix" {
    cvad_config = {
      customer_id   = "${var.customer_id}"  # Optionally set with `CITRIX_CUSTOMER_ID` environment variable.
      client_id     = "${var.api_key_clientId}"  # Optionally set with `CITRIX_CLIENT_ID` environment variable.
      client_secret = "${var.api_key_clientSecret}"  # Optionally set with `CITRIX_CLIENT_SECRET` environment variable.
    }
}
```

You can use environment variables as stated in the comments above. When running Go tests, always use environment variables so that no credentials or other sensitive information are checked-in to the code.

Below is a table to show the difference between on-premises and Cloud provider configuration:

|              | Cloud                                 | On-Premises                           |
|--------------|---------------------------------------|---------------------------------------|
| environment  | `Production`, `Japan`, `Gov`          | N/A                                   |
| customerId   | Cloud Customer Id                     | N/A                                   |
| hostname     | (Optional) Cloud DDC hostname         | On-Premises DDC Hostname / IP address |
| clientId     | Citrix Cloud service principal ID     | Domain Admin Username                 |
| clientSecret | Citrix Cloud service principal secret | Domain Admin Password                 |

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

For on-premises sites with version >= 2311 are supported. Web Studio needs to be [installed and configured](https://docs.citrix.com/en-us/citrix-virtual-apps-desktops/install-configure/install-core/install-web-studio.html#install-web-studio-1) for the provider to work.

### (Cloud Only) Create a Citrix Cloud Service Principal
A service principal is an API client which is not associated with an email. It can be given delegated permissions just like a regular administrator. Follow the [Citrix Cloud API Access with Service Principals](https://developer-docs.citrix.com/en-us/citrix-cloud/citrix-cloud-api-overview/get-started-with-citrix-cloud-apis#citrix-cloud-api-access-with-service-principals) guide to create a service principal for your cloud customer. When selecting the service principal's access choose an appropriate DaaS role.

### Configure your Plugin for Terraform Provider for Citrix DaaS™

Refer to section [Provider Configuration](#provider-configuration) to configure the provider for the Citrix DaaS site you want to manage with Terraform.

### Start writing Terraform for managing your Citrix DaaS site

To find all the Citrix DaaS resources manageable via Terraform, understand all the configurable properties for each resource and how they work together, refer documentations for resources in [Citrix Terraform resource documentation](docs/resources). To better understand how the resource is managed via Citrix DaaS Rest API, you can refer the [Citrix DaaS Rest API documentation](https://developer.cloud.com/citrixworkspace/citrix-daas/citrix-daas-rest-apis/docs/overview).

### Create a Zone in Citrix DaaS as the first step

Refer the [DaaS Zone documentation](docs/resources/zone.md) to configure a zone via terraform. 

### Create a Hypervisor

Hypervisor is needed to use your preferred public cloud provider with Citrix DaaS. Refer the [DaaS Hypervisor documentation](docs/resources/azure_hypervisor.md) to configure an Azure hypervisor in a zone via terraform.

### Create a Hypervisor Resource Pool

The hypervisor resource pool defines the network configuration for a hypervisor connection. Refer the [DaaS Hypervisor Resource Pool documentation](docs/resources/azure_hypervisor_resource_pool.md) to configure an Azure hypervisr resource pool via terraform.

### Create a Machine Catalog

A machine catalog is a collection of machines managed as a single entity. Refer the [DaaS Machine Catalog documentation](docs/resources/machine_catalog.md) to configure a machine catalog via terraform.

### Create a Delivery Group
A delivery group is a collection of machines selected from one or more machine catalogs. The delivery group can also specify which users can use those machines, plus the applications and desktops available to those users. Refer the [DaaS Delivery Group documentation](docs/resources/delivery_group.md) to configure a delivery group via terraform.

## Using the Plugin for Terraform Provider for other Citrix resources

### Configure Global App Configuration (GAC) Settings

The Global App Configuration service provides a centralized setup for IT admins to easily configure Citrix Workspace app settings on Windows, Mac, Android, iOS, HTML5, Chrome OS platforms. Please refer to [Global App Configuration settings documentation](docs/resources/gac_settings.md) to configure GAC settings via terraform.

### Create Citrix Cloud Resource Locations

Resource locations contain the resources (e.g. cloud connectors) required to deliver applications and desktops to users. Resource locations are only supported for Cloud customers. On-premises customers can use the zone resource directly. Please refer to [Citrix Resource Location documentation](docs/resources/resource_location.md) to configure citrix cloud resource locations via terraform.

### Managing StoreFront resources
Please refer to the [StoreFront.md](StoreFront.md) to configure StoreFront resources via terraform. Note that this feature is in Tech Preview.

### Managing DaaS Quick Deploy resources
QuickCreate service allows customers to create and manage Amazon WorkSpaces Core instances in Amazon Web Services (AWS). Please refer to the [QuickCreate documentation](https://docs.citrix.com/en-us/citrix-daas/install-configure/amazon-workspaces-core.html) to learn more. Note that this feature is in Tech Preview.

## Frequently Asked Questions

#### What resource is supported for different connection types?

| Connection Type  |   Hypervisor       |   Resource Pool    |  MCS Power Managed   | MCS Provisioning     |          PVS             | Manual/Remote PC     |
|------------------|--------------------|--------------------|----------------------|----------------------|--------------------------|----------------------|
| AzureRM          |:heavy_check_mark:  |:heavy_check_mark:  | :heavy_check_mark:   | :heavy_check_mark:   |:heavy_check_mark:        | :heavy_check_mark:   |
| AWS EC2          |:heavy_check_mark:  |:heavy_check_mark:  | :heavy_check_mark:   | :heavy_check_mark:   |:heavy_multiplication_x:  | :heavy_check_mark:   |
| GCP              |:heavy_check_mark:  |:heavy_check_mark:  | :heavy_check_mark:   | :heavy_check_mark:   |:heavy_multiplication_x:  | :heavy_check_mark:   |
| vSphere          |:heavy_check_mark:  |:heavy_check_mark:  | :heavy_check_mark:   | :heavy_check_mark:   |:heavy_multiplication_x:  | :heavy_check_mark:   |
| XenServer        |:heavy_check_mark:  |:heavy_check_mark:  | :heavy_check_mark:   | :heavy_check_mark:   |:heavy_multiplication_x:  | :heavy_check_mark:   |
| Nutanix          |:heavy_check_mark:  |:heavy_check_mark:  | :heavy_check_mark:   | :heavy_check_mark:   |:heavy_multiplication_x:  | :heavy_check_mark:   |
| SCVMM            |:heavy_check_mark:  |:heavy_check_mark:  | :heavy_check_mark:   | :heavy_check_mark:   |:heavy_multiplication_x:  | :heavy_check_mark:   |

## Attributions
The code in this repository makes use of the following packages:
-	Hashicorp Terraform Plugin Framework (https://github.com/hashicorp/terraform-plugin-framework)
-	Google Go Uuid (https://github.com/google/uuid)

## License 
This project is Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0 

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

<sub>Copyright © 2025. Citrix Systems, Inc.</sub>