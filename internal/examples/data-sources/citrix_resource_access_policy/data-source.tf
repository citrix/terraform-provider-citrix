# Get Resource Access Policy detail by name
data "citrix_resource_access_policy" "example_policy_by_name" {
    name = "Example Resource Access Policy"
}

# Get Resource Access Policy detail by id
data "citrix_resource_access_policy" "example_policy_by_id" {
    id = "1"
}
