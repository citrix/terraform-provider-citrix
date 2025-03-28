# Plugin for Terraform Provider for Citrix® Examples

This folder contains examples of how to configure Citrix environments in various ways using this provider.

## Table of Contents
- [Plugin for Terraform Provider for Citrix® Examples](#plugin-for-terraform-provider-for-citrix-examples)
  - [Table of Contents](#table-of-contents)
  - [Deployment guides](#deployment-guides)
  - [How to use the examples in this repository](#how-to-use-the-examples-in-this-repository)
    - [Specifying variables](#specifying-variables)
    - [Provider settings](#provider-settings)
  - [Examples](#examples)
    - [Basic MCS](#basic-mcs)
    - [Non-domain joined MCS](#non-domain-joined-mcs)
    - [Quick Deploy](#quick-deploy)

## Deployment guides
Complete steps from start to finish for a variety of senarios from [Citrix Tech Zone](https://community.citrix.com/tech-zone/automation/):
- [Installing and configuring the provider](https://community.citrix.com/tech-zone/automation/terraform-install-and-config/)
- [AWS EC2](https://community.citrix.com/tech-zone/build/deployment-guides/terraform-daas-aws/) via MCS
- [Azure](https://community.citrix.com/tech-zone/build/deployment-guides/daas-azure-iac) via MCS
- [GCP](https://community.citrix.com/tech-zone/build/deployment-guides/terraform-daas-gcp/) via MCS
- [vSphere](https://community.citrix.com/tech-zone/build/deployment-guides/terraform-daas-vsphere8/) via MCS
- [XenServer](https://community.citrix.com/tech-zone/automation/citrix-terraform-xenserver) via MCS
- [Citrix policies](https://community.citrix.com/tech-zone/automation/cvad-terraform-policies/)

## How to use the examples in this repository
Clone this repository and then navigate to the given example directory, [specify the variables](#specifying-variables), and run terraform there:
```shell
> git clone https://github.com/citrix/terraform-provider-citrix.git
> cd terraform-provider-citrix/examples/basic_azure_mcs_vda
> cp terraform.template.tfvars terraform.tfvars
> code terraform.tfvars # open the file and specify variables
> terraform init
> terraform plan
> terraform apply
```

### Specifying variables
Each example contains a [variables.tf](basic_azure_mcs_vda/variables.tf) file which needs to be specified. There are also some default values and configuration options in the other `.tf` files in the directory. Review these options and adjust depending on your use case. 

The variables can be specified by copying the [terraform.template.tfvars](basic_azure_mcs_vda/terraform.template.tfvars) file to `terraform.tfvars` and then filling it out with your values.

Another option is to pass them into the terraform command one by one:
```shell
terraform apply -var="customer_id=<customerId>" -var="client_id=<clientId>" -var=...
```

This method has the benefit of being able to fetch secrets from secure locations and pass them via the commandline in the form of shell or environment variables.

### Provider settings
Each example contains a `citrix.tf` file with the Citrix provider configuration. In this file select between on-premises and Cloud and fill in the required credentials, then delete the other one so there is only one Citrix provider configuration.


## Examples

### Basic MCS
Each of the following examples creates a single multi-session domain joined VDA in the given hypervisor using machine creation services. The VDA is power managed and uses autoscale to stay powered on between 9am and 5pm. [machine_catalogs.tf](basic_azure_mcs_vda/machine_catalogs.tf) can be modified depending how the master image is stored. *Note for Cloud customers*, the zone specified by `var.zone_name` needs to already exist and have [Cloud Connectors](https://docs.citrix.com/en-us/citrix-cloud/citrix-cloud-resource-locations/citrix-cloud-connector.html) configured.
* [AWS EC2](basic_aws_mcs_vda/)
* [Microsoft Azure](basic_azure_mcs_vda/)
  * If using an Azure image gallery, uncomment the `gallery_image` in [machine_catalogs.tf](basic_azure_mcs_vda/machine_catalogs.tf) and remove the VHD parameters.
* [Google Cloud Compute](basic_gcp_mcs_vda/)
* [Nutanix](basic_nutanix_mcs_vda/)
* [VMware vSphere](basic_vsphere_mcs_vda/)
* [XenServer](basic_xenserver_mcs_vda/)

### Non-domain joined MCS
A single example of how to create a non-domain joined VDA, based on the [basic_azure_mcs_vda](basic_aws_mcs_vda/) example.
* [Microsoft Azure](non_domain_joined_azure_mcs_vda/)
  * If using an Azure image gallery, uncomment the `gallery_image` in [machine_catalogs.tf](basic_azure_mcs_vda/machine_catalogs.tf) and remove the VHD parameters.

The difference is in [machine_catalogs.tf](non_domain_joined_azure_mcs_vda/machine_catalogs.tf), `provisioning_scheme.identity_type = "Workgroup"` and the addition of a Citrix Group Policy in [policy_sets.tf](non_domain_joined_azure_mcs_vda/policy_sets.tf).

### Quick Deploy
Examples of creating resources using DaaS Quick Deploy. This includes the account, directory or resource connection, image, and finally the deployment which is assigned to a list of usernames provided in the variables.
* [AWS WorkSpaces Core](basic_aws_workspace_core_deployment)
