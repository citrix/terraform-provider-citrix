# Get VDA resource by machine catalog Name or Id
data "citrix_daas_vda" "vda_by_machine_catalog" {
    machine_catalog = "{MachineCatalog Name or Id}"
}

# Get VDA resource by delivery group Name or Id
data "citrix_daas_vda" "vda_by_delivery_group" {
    delivery_group = "{DeliveryGroup Name or Id}"
}