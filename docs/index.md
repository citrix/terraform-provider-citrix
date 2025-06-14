---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "Citrix Provider"
subcategory: ""
description: |-
  Manage and deploy Citrix resources easily using the Citrix Terraform provider. The provider currently supports both Citrix Virtual Apps & Desktops (CVAD 2311+) and Citrix Desktop as a Service (DaaS) solutions. You can automate creation of site setup including host connections, machine catalogs and delivery groups etc for both CVAD and Citrix DaaS. You can deploy resources in Citrix supported hypervisors and public clouds. Currently, we support deployments in Nutanix, VMware vSphere, XenServer, Microsoft Azure, AWS EC2 and Google Cloud Compute. Additionally, you can also use Manual provisioning or RemotePC to add workloads. The provider is developed and maintained by Citrix.
---

# Citrix Provider

Manage and deploy Citrix resources easily using the Citrix Terraform provider. The provider currently supports both Citrix Virtual Apps & Desktops (CVAD 2311+) and Citrix Desktop as a Service (DaaS) solutions. You can automate creation of site setup including host connections, machine catalogs and delivery groups etc for both CVAD and Citrix DaaS. You can deploy resources in Citrix supported hypervisors and public clouds. Currently, we support deployments in Nutanix, VMware vSphere, XenServer, Microsoft Azure, AWS EC2 and Google Cloud Compute. Additionally, you can also use Manual provisioning or RemotePC to add workloads. The provider is developed and maintained by Citrix.

Documentation regarding the [Data Sources](https://developer.hashicorp.com/terraform/language/data-sources) and [Resources](https://developer.hashicorp.com/terraform/language/resources) supported by the Citrix Provider can be found in the navigation to the left.

Check out the [release notes](https://github.com/citrix/terraform-provider-citrix/releases) to find out more about the provider's latest features and version information.

## Getting Started

New to Terraform? Click [here](https://developer.hashicorp.com/terraform) to learn more.

### Roadmap Proposal for a Smoother Onboarding Experience
To streamline your onboarding experience with the Citrix Terraform Provider, we recommend to start small by importing or creating one or two resources and build out from there.

#### Import Your Existing Resources Using the Onboarding Script
Use our [Onboarding Script](https://github.com/citrix/terraform-provider-citrix/blob/main/scripts/onboarding-helper/) with the following parameters to import existing resources and create Terraform configuration files for them.

- Use the `-ResourceTypes` parameter to specify just a few resource types (for example `citrix_zone`, `citrix_delivery_group`)
- Use the `-NamesOrIds` parameter to filter for specific resources by name or ID

Example:
```powershell
.\terraform-onboarding.ps1 CustomerId "{CustomerId}" -ClientId "{ClientId}" -ClientSecret "{ClientSecret}" -ResourceTypes "citrix_zone","citrix_delivery_group" -NamesOrIds "Primary Zone","Sales Delivery Group"
```

This incremental approach allows you to become familiar with Terraform concepts and the Citrix provider while working with a smaller, more focused set of resources.

#### Manual Configuration
Alternatively, we recommend starting by creating new `.tf` files for the core resources essential for a Citrix deployment:

- [citrix_cloud_resource_location](https://registry.terraform.io/providers/citrix/citrix/latest/docs/resources/cloud_resource_location) (for Citrix Cloud customers only)
- [citrix_zone](https://registry.terraform.io/providers/citrix/citrix/latest/docs/resources/zone)
- citrix_{hosting provider}_hypervisor
- citrix_{hosting provider}_hypervisor_resource_pool

These resources are straightforward to configure and can be created or removed quickly. Begin your Terraform journey with these resources to build confidence in managing your Citrix deployment via Terraform.

Once these resources are properly configured, the next step is to set up your [machine catalog](https://registry.terraform.io/providers/citrix/citrix/latest/docs/resources/machine_catalog) with Terraform. Managing the machine catalog with Terraform will provide a solid foundation for designing a pipeline that meets your specific use case.

### Creating Citrix resources via Terraform

Please refer to [Citrix Tech Zone](https://community.citrix.com/tech-zone/automation/) to find detailed guides on how to deploy and manage resources using the Citrix provider:
- [Installing and configuring the provider](https://community.citrix.com/tech-zone/automation/terraform-install-and-config/)
- [Daily administrative operations](https://community.citrix.com/tech-zone/automation/terraform-daily-administration/)
- [AWS EC2](https://community.citrix.com/tech-zone/build/deployment-guides/terraform-daas-aws/) via MCS
- [AWS WorkSpaces Core](https://community.citrix.com/tech-zone/learn/poc-guides/daas-and-awc-terraform)
- [Azure](https://community.citrix.com/tech-zone/build/deployment-guides/daas-azure-iac) via MCS
- [GCP](https://community.citrix.com/tech-zone/build/deployment-guides/terraform-daas-gcp/) via MCS
- [vSphere](https://community.citrix.com/tech-zone/build/deployment-guides/terraform-daas-vsphere8/) via MCS
- [XenServer](https://community.citrix.com/tech-zone/automation/citrix-terraform-xenserver) via MCS
- [Citrix policies](https://community.citrix.com/tech-zone/automation/cvad-terraform-policies/)

### Basic Examples
Basic example templates for getting started can be found in our [GitHub repository](https://github.com/citrix/terraform-provider-citrix/blob/main/examples/README.md).

### Demo video
[![alt text](https://raw.githubusercontent.com/citrix/terraform-provider-citrix/refs/heads/main/images/techzone-youtube-thumbnail.png)](https://www.youtube.com/watch?v=c33sMLaCVjY)
https://www.youtube.com/watch?v=c33sMLaCVjY

### (On-Premises Only) Enable Web Studio

For on-premises sites with version >= 2311 are supported. Web Studio needs to be [installed and configured](https://docs.citrix.com/en-us/citrix-virtual-apps-desktops/install-configure/install-core/install-web-studio.html#install-web-studio-1) for the provider to work.

### (Cloud Only) Create a Citrix Cloud Service Principal

A service principal is an API client which is not associated with an email. It can be given delegated permissions just like a regular administrator. Follow the [Citrix Cloud API Access with Service Principals](https://developer-docs.citrix.com/en-us/citrix-cloud/citrix-cloud-api-overview/get-started-with-citrix-cloud-apis#citrix-cloud-api-access-with-service-principals) guide to create a service principal for your cloud customer. When selecting the service principal's access choose an appropriate DaaS role.

## Related Citrix Automation Repositories
|            Title            |            Details            |
|-----------------------------|-------------------------------|
| [Packer Image Management Module for Citrix® Virtual Apps and Desktops](https://github.com/citrix/citrix-packer-tools) | Use Packer to create golden images with the Citrix VDA installed and using Citrix best practices. |
| [Citrix Ansible Tools](https://github.com/citrix/citrix-ansible-tools) | Playbooks to install Citrix components using automation such as the VDA. |
| [Site Deployment Module for Citrix® Virtual Apps and Desktops](https://github.com/citrix/citrix-cvad-site-deployment-module) | Uses PowerShell to drive Terraform files to create a fully functional CVAD site. |

## Frequently Asked Questions

### What resource is supported for different connection types?

| Connection Type                         |   Hypervisor       |   Resource Pool    |  MCS Power Managed | MCS Provisioning   |         PVS        | Manual/Remote PC     |
|-----------------------------------------|--------------------|--------------------|--------------------|--------------------|--------------------|----------------------|
| AzureRM                                 |:heavy_check_mark:  | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark:   |
| AWS EC2                                 |:heavy_check_mark:  | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | N/A                | :heavy_check_mark:   |
| GCP                                     |:heavy_check_mark:  | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | N/A                | :heavy_check_mark:   |
| vSphere                                 |:heavy_check_mark:  | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | N/A                | :heavy_check_mark:   |
| XenServer                               |:heavy_check_mark:  | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | N/A                | :heavy_check_mark:   |
| Nutanix                                 |:heavy_check_mark:  | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | N/A                | :heavy_check_mark:   |
| SCVMM                                   |:heavy_check_mark:  | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | N/A                | :heavy_check_mark:   |
| Red Hat OpenShift (**Techpreview**)     |:heavy_check_mark:  | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | N/A                | :heavy_check_mark:   |
| HPE Moonshot (**Techpreview**)          |:heavy_check_mark:  | N/A                | :heavy_check_mark: | N/A                | N/A                | :heavy_check_mark:   |
| Remote PC Wake On LAN (**Techpreview**) |:heavy_check_mark:  | N/A                | N/A                | N/A                | N/A                | :heavy_check_mark:   |

### What URLs should be whitelisted in order to use the Citrix Terraform provider?
- URLs of the Citrix admin consoles: please visit [this documentation](https://docs.citrix.com/en-us/citrix-cloud/overview/requirements/internet-connectivity-requirements.html) for more information.
- URL of the HashiCorp Terraform registry: https://registry.terraform.io or a private registry.

### How do I get the ID to import a DaaS resource?
The [Object IDs Helper Script](https://github.com/citrix/terraform-provider-citrix/blob/main/scripts/object-ids-helper/) will discover all resource IDs and save them to a JSON file for easy reference.

Alternatively the IDs can be found in Web Studio by looking at the network traces. Open your browser developer tools (usually F12) and navigate to the `Network` tab. Refresh Web Studio and click on the resource you want to find the ID for. There should be 2 corresponding network calls (`OPTIONS` then `GET`) for the resource which includes the ID as the last path in the url before the `?` query.

For example in this network call the delivery group ID is `9e451353-d41c-40d5-80da-37177680364b`:
```
OPTIONS https://customerId.xendesktop.net/citrix/orchestration/api/customerId/e4c48b1c-0c2c-4ede-b9a2-ec34998ab118/DeliveryGroups/9e451353-d41c-40d5-80da-37177680364b?fields=SimpleAccessPolicy%2C...
```

### Are my secrets safe in the Terraform state file?
When you use Terraform, any secret in the resource configuration will be stored in the state file. Terraform has guidance to handle the state file itself as sensitive: https://developer.hashicorp.com/terraform/language/state/sensitive-data. This can be mitigated by using a remote state file with encryption enabled. 

It is still best to avoid putting secrets in the state file, and DaaS has a few options to avoid storing secrets in the state:
#### Azure Hypervisor 
MCS offers the option to use the managed identity of the Citrix Cloud Connector to call Azure APIs instead of the application ID + secret. See the [Citrix docs](https://docs.citrix.com/en-us/citrix-daas/install-configure/connections/connection-azure-resource-manager.html#create-a-host-connection-using-azure-managed-identity) for this feature and the [provider docs](https://registry.terraform.io/providers/citrix/citrix/latest/docs/resources/service_account#authentication_mode-1)
```
resource "citrix_azure_hypervisor" "example-azure-hypervisor" {
    name                = "example-azure-hypervisor"
    zone                = "<Zone Id>"
    active_directory_id = "<Azure Tenant Id>"
    subscription_id     = "<Azure Subscription Id>"
    authentication_mode = "SystemAssignedManagedIdentity" // or "UserAssignedManagedIdentities"
    proxy_hypervisor_traffic_through_connector = true
}
```

#### Domain Password
A domain user is required for the `citrix_machine_catalog` resource to create and manage AD machine accounts for the VDAs. This can be pre-created as a Service Account in Web Studio and then imported into Terraform. The machine catalog will then use the credentials stored on the DDC to communicate with AD. See the [citrix_service_account](https://registry.terraform.io/providers/citrix/citrix/latest/docs/resources/service_account) and [citrix_machine_catalog](https://registry.terraform.io/providers/citrix/citrix/latest/docs/resources/machine_catalog#service_account_id-1) docs.
```
resource citrix_service_account "example-service-account" {
    // These values should match what was entered in Web Studio to ensure the import is successful
    display_name = "example-ad-service-account"
    identity_provider_type = "ActiveDirectory"
    identity_provider_identifier = "<DomainFQDN>"
    account_id = "<Domain>\\<Admin Username>"
    account_secret_format = "PlainText"

    // the actual secret is already in remote, putting a dummy value here and setting to ignore changes because this argument is required
    account_secret = "dummy secret for import" 
    lifecycle {
        ignore_changes = [account_secret]
    }
}

// terraform import citrix_service_account.example-service-account <service account ID>

resource "citrix_machine_catalog" "dj-test" {
    provisioning_scheme = {
        machine_domain_identity = {
            domain             = "<DomainFQDN>"
            // use the imported service account when creating this catalog
            service_account_id = citrix_service_account.cmdlab-service-account.id
    ...
```

#### DaaS, Citrix Cloud, and DaaS Quick Deploy resources
- https://api.cloud.com
- Or for Japan environment: https://api.citrixcloud.jp
- Or for Gov environment: https://[customerId].xendesktop.us and https://*.citrixworkspacesapi.us

##### Citrix Cloud Identity Providers resources
- https://cws.citrixworkspacesapi.net
- Or for Japan environment: https://cws.citrixworkspacesapi.jp
- Or for Gov environment: https://cws.citrixworkspacesapi.us
 
#### CVAD (On-premises) resources
- Hostname of the DDC

#### StoreFront resources
- Hostname of the StoreFront Server
- Hostname of the DDC

#### WEM resources
- US environment: https://api.wem.cloud.com
- EU environment: https://eu-api.wem.cloud.com
- APS environment: https://aps-api.wem.cloud.com
- Japan environment: https://jp-api.wem.citrixcloud.jp

## Example Usage

```terraform
# Cloud Provider
provider "citrix" {
    cvad_config = {
      customer_id   = ""
      client_id     = ""
      # secret can be specified via the CITRIX_CLIENT_SECRET environment variable
    }
}

# On-Premises Provider
provider "citrix" {
    cvad_config = {
      hostname      = "10.0.0.6"
      client_id     = "foo.local\\admin"
      # secret can be specified via the CITRIX_CLIENT_SECRET environment variable
    }
}

# Storefront Provider
provider "citrix" {
  storefront_remote_host = {
    computer_name = ""
    ad_admin_username =""
    ad_admin_password =""
    # secret can be specified via the CITRIX_CLIENT_SECRET environment variable
  }
}
```

### `cvad_config` options
Below is a table to show the difference between on-premises and Cloud provider configuration:

|              | Cloud                                 | On-Premises                           |
|--------------|---------------------------------------|---------------------------------------|
| environment  | `Production`, `Japan`, `Gov`          | N/A                                   |
| customerId   | Cloud Customer Id                     | N/A                                   |
| hostname     | (Optional) Cloud DDC hostname         | On-Premises DDC Hostname / IP address |
| clientId     | Citrix Cloud service principal ID     | Domain Admin Username                 |
| clientSecret | Citrix Cloud service principal secret | Domain Admin Password                 |

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `cvad_config` (Attributes) Configuration for CVAD service. (see [below for nested schema](#nestedatt--cvad_config))
- `storefront_remote_host` (Attributes) StoreFront Remote Host for Citrix DaaS service. <br />Only applicable for Citrix on-premises StoreFront. Use this to specify StoreFront Remote Host. <br /> (see [below for nested schema](#nestedatt--storefront_remote_host))
- `wem_on_prem_config` (Attributes) Configuration for WEM on-premises service. (see [below for nested schema](#nestedatt--wem_on_prem_config))

<a id="nestedatt--cvad_config"></a>
### Nested Schema for `cvad_config`

Optional:

- `client_id` (String) Client Id for Citrix DaaS service authentication. 
For Citrix On-Premises customers: Use this to specify a DDC administrator username. 
For Citrix Cloud customers: Use this to specify Cloud API Key Client Id.

-> **Note** Can be set via Environment Variable **CITRIX_CLIENT_ID**.

~> **Please Note** This parameter is required to be specified in the provider configuration or via environment variable.
- `client_secret` (String, Sensitive) Client Secret for Citrix DaaS service authentication. 
For Citrix on-premises customers: Use this to specify a DDC administrator password. 
For Citrix Cloud customers: Use this to specify Cloud API Key Client Secret.

-> **Note** Can be set via Environment Variable **CITRIX_CLIENT_SECRET**.

~> **Please Note** This parameter is required to be specified in the provider configuration or via environment variable.
- `customer_id` (String) The Citrix Cloud customer ID.

-> **Note** Can be set via Environment Variable **CITRIX_CUSTOMER_ID**.

~> **Please Note** This parameter is required for Citrix Cloud customers to be specified in the provider configuration or via environment variable.
- `disable_daas_client` (Boolean) Disable Citrix DaaS client setup. 
Set to true to skip Citrix DaaS client setup. 

-> **Note** Can be set via Environment Variable **CITRIX_DISABLE_DAAS_CLIENT**.
- `disable_ssl_verification` (Boolean) Disable SSL verification against the target DDC. 
Set to true to skip SSL verification only when the target DDC does not have a valid SSL certificate issued by a trusted CA. 
When set to true, please make sure that your provider config is set for a known DDC hostname. 

-> **Note** Can be set via Environment Variable **CITRIX_DISABLE_SSL_VERIFICATION**.

~> **Please Note** [It is recommended to configure a valid certificate for the target DDC](https://docs.citrix.com/en-us/citrix-virtual-apps-desktops/install-configure/install-core/secure-web-studio-deployment)
- `environment` (String) Citrix Cloud environment of the customer. Available options: `Production`, `Staging`, `Japan`, `JapanStaging`, `Gov`, `GovStaging`. 

-> **Note** Can be set via Environment Variable **CITRIX_ENVIRONMENT**.

~> **Please Note** Only applicable for Citrix Cloud customers.
- `hostname` (String) Host name / base URL of Citrix DaaS service. 
For Citrix on-premises customers: Use this to specify Delivery Controller hostname. 
For Citrix Cloud customers: Use this to force override the Citrix DaaS service hostname.

-> **Note** Can be set via Environment Variable **CITRIX_HOSTNAME**.

~> **Please Note** This parameter is required for on-premises customers to be specified in the provider configuration or via environment variable.
- `wem_region` (String) WEM Hosting Region of the Citrix Cloud customer. Available values are `US`, `EU`, and `APS`.

-> **Note** Can be set via Environment Variable **CITRIX_WEM_REGION**.

~> **Please Note** Only applicable for Citrix Workspace Environment Management (WEM) Cloud customers.


<a id="nestedatt--storefront_remote_host"></a>
### Nested Schema for `storefront_remote_host`

Optional:

- `ad_admin_password` (String, Sensitive) Active Directory Admin Password to connect to storefront server <br />Use this to specify AD admin password<br />Can be set via Environment Variable **SF_AD_ADMIN_PASSWORD**.<br />This parameter is **required** to be specified in the provider configuration or via environment variable.
- `ad_admin_username` (String) Active Directory Admin Username to connect to storefront server <br />Use this to specify AD admin username <br />Can be set via Environment Variable **SF_AD_ADMIN_USERNAME**.<br />This parameter is **required** to be specified in the provider configuration or via environment variable.
- `computer_name` (String) StoreFront server computer Name <br />Use this to specify StoreFront server computer name <br />Can be set via Environment Variable **SF_COMPUTER_NAME**.<br />This parameter is **required** to be specified in the provider configuration or via environment variable.
- `disable_ssl_verification` (Boolean) Disable SSL verification against the target storefront server. <br />Only applicable to customers connecting to storefront server remotely. Customers should omit this option when running storefront provider locally. Set to true to skip SSL verification only when the target DDC does not have a valid SSL certificate issued by a trusted CA. <br />When set to true, please make sure that your provider storefront_remote_host is set for a known storefront hostname. <br />Can be set via Environment Variable **SF_DISABLE_SSL_VERIFICATION**.


<a id="nestedatt--wem_on_prem_config"></a>
### Nested Schema for `wem_on_prem_config`

Optional:

- `admin_password` (String, Sensitive) WEM Admin Password to connect to WEM service <br />Use this to specify WEM admin password<br />Can be set via Environment Variable **WEM_ADMIN_PASSWORD**.<br />This parameter is **required** to be specified in the provider configuration or via environment variable.
- `admin_username` (String) WEM Admin Username to connect to WEM service <br />Use this to specify WEM admin username <br />Can be set via Environment Variable **WEM_ADMIN_USERNAME**.<br />This parameter is **required** to be specified in the provider configuration or via environment variable.
- `disable_ssl_verification` (Boolean) Disable SSL verification against the target WEM service. <br />Set to true to skip SSL verification only when the target WEM service does not have a valid SSL certificate issued by a trusted CA. <br />When set to true, please make sure that your provider config is set for a known WEM hostname. <br />Can be set via Environment Variable **WEM_DISABLE_SSL_VERIFICATION**.
- `hostname` (String) Name of server hosting Citrix WEM service. <br />Use this to specify WEM service hostname. <br />Can be set via Environment Variable **WEM_HOSTNAME**.<br />This parameter is **required** to be specified in the provider configuration or via environment variable.
