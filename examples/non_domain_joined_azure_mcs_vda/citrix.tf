# Citrix Cloud customer provider settings
# Please comment out / remove this provider settings block if you are an On-Premises customer
provider "citrix" {
  cvad_config = {
    customer_id                 = var.provider_customer_id
    client_id                   = var.provider_client_id
    client_secret               = var.provider_client_secret
    environment                 = "Staging"
    hostname                    = "api.dev.cloud.com"
  }
}
