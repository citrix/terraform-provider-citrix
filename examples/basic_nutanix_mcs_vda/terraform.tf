terraform {
  required_version = ">= 1.4.0"

  required_providers {
    citrix = {
      source  = "citrix/citrix"
      version = ">=1.0.14"
    }
  }

  backend "local" {}
}
