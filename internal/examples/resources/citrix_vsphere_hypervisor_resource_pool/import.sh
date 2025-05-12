# Hypervisor Resource Pool can be imported with the format HypervisorId,HypervisorResourcePoolId
terraform import citrix_vsphere_hypervisor_resource_pool.example-vsphere-hypervisor-resource-pool sbf0dc45-5c42-45a0-a15d-a3df4ff5da8c,ce571dd9-1a46-4b85-891c-484423322c53

# Hypervisor Resource Pool can be imported by specifying the GUID
terraform import citrix_vsphere_hypervisor_resource_pool.example-vsphere-hypervisor-resource-pool ce571dd9-1a46-4b85-891c-484423322c53