resource "citrix_wem_directory_object" "example-directory-object" {
    configuration_set_id = citrix_wem_configuration_set.example-config-set.id
    machine_catalog_id = citrix_machine_catalog.example-machine-catalog.id
    enabled = true
}
