resource "citrix_daas_admin_scope" "example-admin-scope" {
    name        = "example-admin-scope"
    description = "Example admin scope for delivery group and machine catalog"
    scoped_objects    = [
        {
            object_type = "DeliveryGroup",
            object = "<Name of existing Delivery Group to be added to the scope>"
        },
        {
            object_type = "MachineCatalog",
            object = "<Name of existing Machine Catalog to be added to the scope>"
        }
    ]
}