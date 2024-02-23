resource "citrix_gcp_hypervisor_resource_pool" "example-gcp-rp" {
    name                = "example-gcp-rp"
    hypervisor          = citrix_gcp_hypervisor.example-gcp-hypervisor.id
    project_name = "<Project Name>"
    region              = "<VNet region>"
    subnets              = [
        "<Subnet name>"
    ]
    vpc    = "{VPC name}"
}


