---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "citrix_machine_catalog Data Source - citrix"
subcategory: "CVAD"
description: |-
  Read data of an existing machine catalog.
---

# citrix_machine_catalog (Data Source)

Read data of an existing machine catalog.

## Example Usage

```terraform
# Get Citrix Machine Catalog data source by name
data "citrix_machine_catalog" "example_machine_catalog" {
    name = "example-catalog"
}

# Get Citrix Machine Catalog data source by ID
data "citrix_machine_catalog" "example_machine_catalog" {
    id = "00000000-0000-0000-0000-000000000000"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `id` (String) GUID identifier of the machine catalog.
- `machine_catalog_folder_path` (String) The path to the folder in which the machine catalog is located.
- `name` (String) Name of the machine catalog.

### Read-Only

- `tags` (Set of String) A set of identifiers of tags to associate with the machine catalog.
- `tenants` (Set of String) A set of identifiers of tenants to associate with the machine catalog.
- `vdas` (Attributes List) The VDAs associated with the machine catalog. (see [below for nested schema](#nestedatt--vdas))

<a id="nestedatt--vdas"></a>
### Nested Schema for `vdas`

Read-Only:

- `associated_delivery_group` (String) Delivery group which the VDA is associated with.
- `associated_machine_catalog` (String) Machine catalog which the VDA is associated with.
- `hosted_machine_id` (String) Machine ID within the hypervisor hosting unit.
- `id` (String) Id of the VDA.
- `machine_name` (String) Machine name of the VDA.