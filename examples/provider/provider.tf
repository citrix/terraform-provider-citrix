# Cloud Provider
provider "citrix" {
  customer_id   = ""
  client_id     = ""
  client_secret = ""
}

# On-Premises Provider
provider "citrix" {
  hostname      = "10.0.0.6"
  client_id     = "foo.local\\admin"
  client_secret = "foo"
}