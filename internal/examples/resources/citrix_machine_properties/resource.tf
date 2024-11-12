resource "citrix_machine_properties" "example_machine_properties" {
    name = "domain\\machine-name" // For workgroup machines, use machine-name only
    machine_catalog_id = "00000000-0000-0000-0000-000000000000" // Id of the machine catalog the machine belongs to
    tags = [ "11111111-1111-1111-1111-111111111111" ] // Tags to be assigned to the machine
}
