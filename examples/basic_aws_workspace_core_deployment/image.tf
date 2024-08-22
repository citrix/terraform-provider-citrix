# Quick Deploy AWS WorkSpaces Image with AMI image
resource "citrix_quickcreate_aws_workspaces_image" "example_aws_workspaces_image" {
  name              = var.image_name
  account_id        = citrix_quickcreate_aws_workspaces_account.example_aws_workspaces_account.id
  aws_image_id      = var.image_id
  description       = var.image_description
  session_support   = var.image_session_support
  operating_system  = var.image_operating_system
  ingestion_process = var.image_ingestion_process
}
