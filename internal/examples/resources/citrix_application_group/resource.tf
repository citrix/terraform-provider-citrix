resource "citrix_application_group" "example-application-group" {
  name                    = "example-name"
  description             = "example-description"
  included_users = ["user@text.com"]
  delivery_groups = [citrix_delivery_group.example-delivery-group.id, citrix_delivery_group.example-delivery-group-2.id]
}


