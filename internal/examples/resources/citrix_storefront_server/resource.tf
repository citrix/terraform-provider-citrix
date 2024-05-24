resource "citrix_storefront_server" "example-storefront_server" {
    name                = "example-storefront-server"
    description         = "StoreFront server example"
    url                 = "https://storefront.example.com/citrix/store"
    enabled             = true
}