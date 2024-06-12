resource "citrix_gcp_hypervisor_resource_pool" "example-gcp-rp" {
    name                = var.resource_pool_name
    hypervisor          = citrix_gcp_hypervisor.example-gcp-hypervisor.id
    project_name        = var.gcp_project_name
    region              = var.gcp_vpc_region
    subnets             = var.gcp_subnets
    vpc                 = var.gcp_vpc
}


