---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "citrix_daas_hypervisor_resource_pool Resource - citrix"
subcategory: ""
description: |-
  Manages a hypervisor resource pool.
---

# citrix_daas_hypervisor_resource_pool (Resource)

Manages a hypervisor resource pool.

## Example Usage

```terraform
resource "citrix_daas_hypervisor_resource_pool" "example-azure-hypervisor-resource-pool" {
    name                = "example-hypervisor-resource-pool"
    hypervisor          = citrix_daas_hypervisor.example-azure-hypervisor.id
    region              = "East US"
	virtual_network_resource_group = "{Resource Group Name}"
    virtual_network     = "{VNet name}"
    subnets     			= [
        "subnet 1",
        "subnet 2"
    ]
}

resource "citrix_daas_hypervisor_resource_pool" "example-aws-hypervisor-resource-pool" {
    name                = "example-hypervisor-resource-pool"
    hypervisor          = citrix_daas_hypervisor.example-aws-hypervisor.id
    subnets            = [
        "10.0.1.0/24",
    ]
    virtual_network   = "{VPC name}"
    availability_zone = "us-east-2a"
}

resource "citrix_daas_hypervisor_resource_pool" "example-gcp-hypervisor-resource-pool" {
    name                = "example-hypervisor-resource-pool"
    hypervisor          = citrix_daas_hypervisor.example-gcp-hypervisor.id
    region             = "us-east1"
    subnets             = [
        "us-east1",
    ]
    virtual_network    = "{VPC name}"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `hypervisor` (String) Id of the hypervisor for which the resource pool needs to be created.
- `name` (String) Name of the resource pool. Name should be unique across all hypervisors.
- `virtual_network` (String) Name of the cloud virtual network.

### Optional

- `availability_zone` (String) **[AWS: Required]** The name of the availability zone resource to use for provisioning operations in this resource pool.
- `project_name` (String) **[GCP: Required]** GCP Project name.
- `region` (String) **[Azure, GCP: Required]** Cloud Region where the virtual network sits in.
- `shared_vpc` (Boolean) **[GCP: Optional]** Indicate whether the GCP Virtual Private Cloud is a shared VPC.
- `subnets` (List of String) **[Azure, GCP: Required]** List of subnets to allocate VDAs within the virtual network.
- `virtual_network_resource_group` (String) **[Azure: Required]** The name of the resource group where the vnet resides.

### Read-Only

- `hypervisor_connection_type` (String) Connection Type of the hypervisor (AzureRM, AWS, GCP).
- `id` (String) GUID identifier of the resource pool.

## Import

Import is supported using the following syntax:

```shell
# Hypervisor Resource Pool can be imported with the format HypervisorId,HypervisorResourcePoolId
terraform import citrix_daas_hypervisor_resource_pool.example-hypervisor-resource-pool sbf0dc45-5c42-45a0-a15d-a3df4ff5da8c,ce571dd9-1a46-4b85-891c-484423322c53
```
