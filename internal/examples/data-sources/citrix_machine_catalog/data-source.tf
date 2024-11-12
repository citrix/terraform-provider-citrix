# Get Citrix Machine Catalog data source by name
data "citrix_machine_catalog" "example_machine_catalog" {
    name = "example-catalog"
}

# Get Citrix Machine Catalog data source by ID
data "citrix_machine_catalog" "example_machine_catalog" {
    id = "00000000-0000-0000-0000-000000000000"
}
