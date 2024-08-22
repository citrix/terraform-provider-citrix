# Quick Deploy AWS WorkSpaces Image with AMI image
resource "citrix_quickcreate_aws_workspaces_image" "example_aws_workspaces_image_ami" {
  name              = "exampe-aws-workspaces-image"
  account_id        = citrix_quickcreate_aws_workspaces_account.example_aws_workspaces_account.id
  aws_image_id      = "ami-012345abcde"
  description       = "Example AWS WorkSpaces image imported with AMI id"
  session_support   = "SingleSession"
  operating_system  = "WINDOWS"
  ingestion_process = "BYOL_REGULAR_BYOP"
}

# Quick Deploy AWS WorkSpaces Image with WSI image
resource "citrix_quickcreate_aws_workspaces_image" "example_aws_workspaces_image_wsi" {
  name              = "exampe-aws-workspaces-image"
  account_id        = citrix_quickcreate_aws_workspaces_account.example_aws_workspaces_account.id
  aws_image_id      = "wsi-012345abcde"
  description       = "Example AWS WorkSpaces image imported with WSI id"
  session_support   = "SingleSession"
  operating_system  = "WINDOWS"
  ingestion_process = "BYOL_REGULAR_BYOP"
}
