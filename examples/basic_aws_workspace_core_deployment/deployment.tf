# User coupled workspaces with MANUAL running mode
resource "citrix_quickcreate_aws_workspaces_deployment" "example_aws_workspaces_deployment" {
    name                    = var.deployment_name
    account_id              = citrix_quickcreate_aws_workspaces_account.example_aws_workspaces_account.id
    directory_connection_id = citrix_quickcreate_aws_workspaces_directory_connection.example_aws_workspaces_directory_connection.id
    image_id                = citrix_quickcreate_aws_workspaces_image.example_aws_workspaces_image.id
    performance             = "STANDARD"
    root_volume_size        = "80"
    user_volume_size        = "50"
    volumes_encrypted       = true
    volumes_encryption_key  = "alias/aws/workspaces"

    running_mode = "ALWAYS_ON"

    user_decoupled_workspaces = false
    dynamic "workspaces" {
        for_each = var.deployment_usernames
        content {
            username         = workspaces.value
            root_volume_size = 80
            user_volume_size = 50
            maintenance_mode = false
        }
    }
}
