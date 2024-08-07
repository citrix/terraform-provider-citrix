# Get details of a Citrix QuickCreate AWS Workspaces Image with image id and account id
data "citrix_quickcreate_aws_workspaces_image" "example_aws_workspaces_image" {
    id         = "00000000-0000-0000-0000-000000000000"
    account_id = "00000000-0000-0000-0000-000000000000"
}

# Get details of a Citrix QuickCreate AWS Workspaces Image with image name and account id
data "citrix_quickcreate_aws_workspaces_image" "example_aws_workspaces_image" {
    name       = "exampleImageName"
    account_id = "00000000-0000-0000-0000-000000000000"
}
