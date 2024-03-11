---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "citrix_vda Data Source - citrix"
subcategory: ""
description: |-
  Data source for the list of VDAs that belong to either a machine catalog or a delivery group. Machine catalog and delivery group cannot be specified at the same time.
---

# citrix_vda (Data Source)

Data source for the list of VDAs that belong to either a machine catalog or a delivery group. Machine catalog and delivery group cannot be specified at the same time.

## Example Usage

```terraform
# Get VDA resource by machine catalog Name or Id
data "citrix_vda" "vda_by_machine_catalog" {
    machine_catalog = "{MachineCatalog Name or Id}"
}

# Get VDA resource by delivery group Name or Id
data "citrix_vda" "vda_by_delivery_group" {
    delivery_group = "{DeliveryGroup Name or Id}"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `delivery_group` (String) The delivery group which the VDAs are associated with.
- `machine_catalog` (String) The machine catalog which the VDAs are associated with.

### Read-Only

- `vdas` (Attributes List) The VDAs associated with the specified machine catalog or delivery group. (see [below for nested schema](#nestedatt--vdas))

<a id="nestedatt--vdas"></a>
### Nested Schema for `vdas`

Read-Only:

- `associated_delivery_group` (String) Delivery group which the VDA is associated with.
- `associated_machine_catalog` (String) Machine catalog which the VDA is associated with.
- `hosted_machine_id` (String) Machine ID within the hypervisor hosting unit.
- `machine_name` (String) Machine name of the VDA.

