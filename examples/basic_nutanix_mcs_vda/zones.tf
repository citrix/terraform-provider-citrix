# CVAD zone configuration
# Please comment out / remove this zone resource block if you are a Citrix Cloud customer
resource "citrix_zone" "example-zone" {
    name        = var.zone_name
    description = "description for example zone"
}

# DaaS zone configuration
# Please uncomment this zone resource block if you are a Citrix Cloud customer
# resource "citrix_zone" "example-zone" {
#     resource_location_id = var.nutanix_resource_location_id
# }