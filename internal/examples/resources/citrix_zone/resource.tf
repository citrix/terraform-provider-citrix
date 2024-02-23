resource "citrix_zone" "example-zone" {
    name                = "example-zone"
    description         = "zone example"
    metadata            = [
        {
            name    = "key1"
            value   = "value1"
        }
    ]
}