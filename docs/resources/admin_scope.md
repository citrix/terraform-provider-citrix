---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "citrix_admin_scope Resource - citrix"
subcategory: ""
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
    scoped_objects    = [
        {
            object_type = "DeliveryGroup",
            object = "<Name of existing Delivery Group to be added to the scope>"
        },
        {
            object_type = "MachineCatalog",
            object = "<Name of existing Machine Catalog to be added to the scope>"
        }
    ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the admin scope.

### Optional

- `description` (String) Description of the admin scope.
- `scoped_objects` (Attributes List) List of scoped objects to be associated with the admin scope. (see [below for nested schema](#nestedatt--scoped_objects))

### Read-Only

- `id` (String) ID of the admin scope.

<a id="nestedatt--scoped_objects"></a>
### Nested Schema for `scoped_objects`

Required:

- `object` (String) Name of an existing object under the object type to be added to the scope.
- `object_type` (String) Type of the scoped object. Allowed values are: `HypervisorConnection`, `MachineCatalog`, `DeliveryGroup`, `ApplicationGroup`, `Tag`, `PolicySet` and `Unknown`.

## Import

Import is supported using the following syntax:

```shell
# Admin Scope can be imported by specifying the GUID
terraform import citrix_admin_scope.example-admin-scope 00000000-0000-0000-0000-000000000000
```
