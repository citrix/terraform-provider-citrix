terraform {
  required_version = ">= 1.4.0"

  required_providers {
    citrix = {
      source  = "citrix/citrix"
      version = ">=0.5.3"
    }
  }

  backend "local" {}
}
