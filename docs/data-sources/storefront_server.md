---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "citrix_storefront_server Data Source - citrix"
subcategory: "CVAD"
description: |-
  Data source of a StoreFront server.
---

# citrix_storefront_server (Data Source)

Data source of a StoreFront server.

## Example Usage

```terraform
# Get StoreFront Server data source by name
data "citrix_storefront_server" "example_storefront_server_by_name" {
    name = "ExampleStoreFrontServer"
}

# Get StoreFront Server data source by id
data "citrix_storefront_server" "example_storefront_server_by_id" {
    id = "00000000-0000-0000-0000-000000000000"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `id` (String) GUID identifier of the StoreFront server.
- `name` (String) Name of the StoreFront server.

### Read-Only

- `description` (String) Description of the StoreFront server.
- `enabled` (Boolean) Indicates if the StoreFront server is enabled. Default is `true`.
- `url` (String) URL for connecting to the StoreFront server.