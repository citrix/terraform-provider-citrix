# Terraform Module for Citrix QuickCreate

This Terraform module allows you to manage resources in Citrix QuickCreate.

## Table of Contents
- [Terraform Module for Citrix QuickCreate](#terraform-module-for-citrix-quickcreate)
    - [Table of Contents](#table-of-contents)
    - [Prerequisites](#prerequisites)
    - [Usage](#usage)
        - [Connect an AWS Workspaces Account with AWS Role ARN](#connect-an-aws-workspaces-account-with-aws-role-arn)
        - [Connect an AWS Workspaces Account with AWS Access Key and Secret Access Key](#connect-an-aws-workspaces-account-with-aceess-key-and-secret-access-key)

## Prerequisites

- Terraform 0.14.x

## Usage

Example Usage of the QuickCreate Terraform Configuration

### Connect an AWS Workspaces Account with AWS Role ARN

```hcl
resource "citrix_quickcreate_aws_workspaces_account" "example_aws_workspaces_account_role_arn" {
    name                    = "exampe-aws-workspaces-account-role-arn"
    aws_region              = "us-east-1"
    aws_role_arn            = "<AWS Role ARN>"
}
```

### Connect an AWS Workspaces Account with Aceess Key and Secret Access Key
```hcl
resource "citrix_quickcreate_aws_workspaces_account" "example_aws_workspaces_account_access_key" {
    name                    = "exampe-aws-workspaces-account-access-key"
    aws_region              = "us-east-1"
    aws_access_key_id       = "<AWS Access Key ID>"
    aws_secret_access_key   = "<AWS Secret Access Key>"
}
```

