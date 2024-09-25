resource "citrix_tag" "example_tag" {
    name = "TagName"
    description = "Example description of the tag"
    scopes = [
        citrix_admin_scope.example_admin_scope.id 
    ]
}
