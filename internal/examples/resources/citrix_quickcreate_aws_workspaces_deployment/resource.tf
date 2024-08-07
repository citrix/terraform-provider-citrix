# User coupled workspaces with MANUAL running mode
resource "citrix_quickcreate_aws_workspaces_deployment" "example_qcs_deployment" {
    name                    = "example-aws-workspaces-deployment"
    account_id              = citrix_quickcreate_aws_workspaces_account.example_aws_workspaces_account.id
    directory_connection_id = citrix_quickcreate_aws_workspaces_directory_connection.example_directory_connection.id
    image_id                = citrix_quickcreate_aws_workspaces_image.example_image.id
    performance             = "STANDARD"
    root_volume_size        = "80"
    user_volume_size        = "50"
    volumes_encrypted       = true
    volumes_encryption_key  = "alias/aws/workspaces"

    running_mode = "MANUAL"
    scale_settings = {
        disconnect_session_idle_timeout = 5
        shutdown_disconnect_timeout = 15
        shutdown_log_off_timeout = 15
        buffer_capacity_size_percentage = 0
    }

    user_decoupled_workspaces = false
    workspaces = [
        {
            username = "user0001"
            root_volume_size = 80
            user_volume_size = 50
            maintenance_mode = false
        },
        {
            username = "user0002"
            root_volume_size = 80
            user_volume_size = 50
            maintenance_mode = true
        },
    ]
}

# User coupled workspaces with ALWAYS_ON running mode
resource "citrix_quickcreate_aws_workspaces_deployment" "example_qcs_deployment" {
    name                    = "example-aws-workspaces-deployment"
    account_id              = citrix_quickcreate_aws_workspaces_account.example_aws_workspaces_account.id
    directory_connection_id = citrix_quickcreate_aws_workspaces_directory_connection.example_directory_connection.id
    image_id                = citrix_quickcreate_aws_workspaces_image.example_image.id
    performance             = "STANDARD"
    root_volume_size        = "80"
    user_volume_size        = "50"
    volumes_encrypted       = true
    volumes_encryption_key  = "alias/aws/workspaces"

    running_mode = "ALWAYS_ON"

    user_decoupled_workspaces = false
    workspaces = [
        {
            username = "user0001"
            root_volume_size = 80
            user_volume_size = 50
            maintenance_mode = false
        },
        {
            username = "user0002"
            root_volume_size = 80
            user_volume_size = 50
            maintenance_mode = true
        },
    ]
}

# User decoupled workspaces with MANUAL running mode
resource "citrix_quickcreate_aws_workspaces_deployment" "example_qcs_deployment" {
    name                    = "example-aws-workspaces-deployment"
    account_id              = citrix_quickcreate_aws_workspaces_account.example_aws_workspaces_account.id
    directory_connection_id = citrix_quickcreate_aws_workspaces_directory_connection.example_directory_connection.id
    image_id                = citrix_quickcreate_aws_workspaces_image.example_image.id
    performance             = "STANDARD"
    root_volume_size        = "80"
    user_volume_size        = "50"
    volumes_encrypted       = true
    volumes_encryption_key  = "alias/aws/workspaces"

    running_mode            = "MANUAL"
    scale_settings = {
        disconnect_session_idle_timeout = 5
        shutdown_disconnect_timeout = 15
        shutdown_log_off_timeout = 15
        buffer_capacity_size_percentage = 0
    }

    user_decoupled_workspaces = true
    workspaces = [
        {
            root_volume_size = 80
            user_volume_size = 50
            maintenance_mode = false
        },
        {
            root_volume_size = 80
            user_volume_size = 50
            maintenance_mode = true
        },
    ]
}

# User decoupled workspaces with ALWAYS_ON running mode
resource "citrix_quickcreate_aws_workspaces_deployment" "example_qcs_deployment" {
    name                    = "example-aws-workspaces-deployment"
    account_id              = citrix_quickcreate_aws_workspaces_account.example_aws_workspaces_account.id
    directory_connection_id = citrix_quickcreate_aws_workspaces_directory_connection.example_directory_connection.id
    image_id                = citrix_quickcreate_aws_workspaces_image.example_image.id
    performance             = "STANDARD"
    root_volume_size        = "80"
    user_volume_size        = "50"
    volumes_encrypted       = true
    volumes_encryption_key  = "alias/aws/workspaces"

    running_mode = "ALWAYS_ON"

    user_decoupled_workspaces = true
    workspaces = [
        {
            root_volume_size = 80
            user_volume_size = 50
            maintenance_mode = false
        },
        {
            root_volume_size = 80
            user_volume_size = 50
            maintenance_mode = true
        },
    ]
}
