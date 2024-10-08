---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "citrix_delivery_group Data Source - citrix"
subcategory: "CVAD"
description: |-
  Read data of an existing delivery group.
---

# citrix_delivery_group (Data Source)

Read data of an existing delivery group.

## Example Usage

```terraform
data "citrix_delivery_group" "example_delivery_group" {
    name = "exampleDeliveryGroupName"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the delivery group.

### Optional

- `delivery_group_folder_path` (String) The path to the folder in which the delivery group is located.

### Read-Only

- `id` (String) GUID identifier of the delivery group.
- `tags` (Set of String) A set of identifiers of tags to associate with the delivery group.
- `tenants` (Set of String) A set of identifiers of tenants to associate with the delivery group.
- `vdas` (Attributes List) The VDAs associated with the delivery group. (see [below for nested schema](#nestedatt--vdas))

<a id="nestedatt--vdas"></a>
### Nested Schema for `vdas`

Read-Only:

- `associated_delivery_group` (String) Delivery group which the VDA is associated with.
- `associated_machine_catalog` (String) Machine catalog which the VDA is associated with.
- `hosted_machine_id` (String) Machine ID within the hypervisor hosting unit.
- `machine_name` (String) Machine name of the VDA.