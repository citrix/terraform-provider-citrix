resource "citrix_zone" "example-zone" {
    resource_location_id = citrix_cloud_resource_location.example-resource-location.id
}
