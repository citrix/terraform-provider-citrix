# Get Citrix Delivery Group data source by name
data "citrix_delivery_group" "example_delivery_group" {
    name = "exampleDeliveryGroupName"
}

# Get Citrix Delivery Group data source by ID
data "citrix_delivery_group" "example_delivery_group" {
    id = "00000000-0000-0000-0000-000000000000"
}
