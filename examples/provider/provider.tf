terraform {
  required_version = ">= 1.4.0"

  required_providers {
    citrix = {
      source  = "citrix/citrix"
      version = ">=0.0.0"
    }
  }

  backend "local" {}
}

# Cloud Provider
provider "citrix" {
  customer_id   = ""
  client_id     = ""
  client_secret = ""
}

# On-Premise Provider
provider "citrix" {
  hostname      = "10.0.0.6"
  client_id     = "foo.local\\admin"
  client_secret = "foo"
}