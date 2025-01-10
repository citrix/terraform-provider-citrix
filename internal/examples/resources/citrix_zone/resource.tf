# Example for On-Premises Zone
resource "citrix_zone" "example-onpremises-zone" {
    name                = "example-zone"
    description         = "zone example"
    metadata            = [
        {
            name    = "key1"
            value   = "value1"
        }
    ]
}

# Example for Cloud Zone
resource "citrix_cloud_resource_location" "example-resource-location" {
    name = "example-resource-location"
}

resource "citrix_zone" "example-cloud-zone" {
    resource_location_id = citrix_cloud_resource_location.example-resource-location.id
}