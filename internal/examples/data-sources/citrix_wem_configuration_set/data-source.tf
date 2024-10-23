# Get WEM configuration set by id
data "citrix_wem_configuration_set" "test_wem_configuration_set_by_id" {
    id = "1"
}

# Get WEM configuration set by name
data "citrix_wem_configuration_set" "test_wem_configuration_set_by_name" {
    name = "Default Site"
}