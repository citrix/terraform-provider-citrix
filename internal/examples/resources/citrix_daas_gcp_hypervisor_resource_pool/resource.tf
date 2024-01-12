resource "citrix_daas_gcp_hypervisor_resource_pool" "example-gcp-hypervisor-resource-pool" {
    name                = "example-gcp-hypervisor-resource-pool"
    hypervisor          = citrix_daas_gcp_hypervisor.example-gcp-hypervisor.id
    project_name       = "10000-example-gcp-project"
    region             = "us-east1"
    subnets             = [
        "us-east1",
    ]
    vpc    = "{VPC name}"
}