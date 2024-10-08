---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "citrix_admin_scope Resource - citrix"
subcategory: "CVAD"
description: |-
  Manages an administrator scope.
---

# citrix_admin_scope (Resource)

Manages an administrator scope.

## Example Usage

```terraform
resource "citrix_admin_scope" "example-admin-scope" {
    name        = "example-admin-scope"
    description = "Example admin scope for delivery group and machine catalog"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the admin scope.

### Optional

- `description` (String) Description of the admin scope.
- `is_tenant_scope` (Boolean) Indicates whether the admin scope is a tenant scope. Defaults to `false`.

### Read-Only

- `id` (String) ID of the admin scope.

## Import

Import is supported using the following syntax:

```shell
# Admin Scope can be imported by specifying the GUID
terraform import citrix_admin_scope.example-admin-scope 00000000-0000-0000-0000-000000000000
```