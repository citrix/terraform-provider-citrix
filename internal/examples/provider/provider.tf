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

# Storefront Provider
provider "citrix" {
  storefront_remote_host = {
    computer_name = ""
    ad_admin_username =""
    ad_admin_password =""
    # secret can be specified via the CITRIX_CLIENT_SECRET environment variable
  }
}