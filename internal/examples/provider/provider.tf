# Cloud Provider
provider "citrix" {
  customer_id   = ""
  client_id     = ""
  # secret can be specified via the CITRIX_CLIENT_SECRET environment variable
}

# On-Premises Provider
provider "citrix" {
  hostname      = "10.0.0.6"
  client_id     = "foo.local\\admin"
  # secret can be specified via the CITRIX_CLIENT_SECRET environment variable
}