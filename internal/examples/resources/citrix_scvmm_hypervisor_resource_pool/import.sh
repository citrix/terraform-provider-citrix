# Hypervisor Resource Pool can be imported with the format HypervisorId,HypervisorResourcePoolId
terraform import citrix_scvmm_hypervisor_resource_pool.example-scvmm-hypervisor-resource-pool b2338edf-281b-436e-9c3a-54c546c3526e,ce571dd9-1a46-4b85-891c-484423322c53

# Hypervisor Resource Pool can be imported by specifying the GUID
terraform import citrix_scvmm_hypervisor_resource_pool.example-scvmm-hypervisor-resource-pool ce571dd9-1a46-4b85-891c-484423322c53