# Get StoreFront Server data source by name
data "citrix_storefront_server" "example_storefront_server_by_name" {
    name = "ExampleStoreFrontServer"
}

# Get StoreFront Server data source by id
data "citrix_storefront_server" "example_storefront_server_by_id" {
    id = "00000000-0000-0000-0000-000000000000"
}