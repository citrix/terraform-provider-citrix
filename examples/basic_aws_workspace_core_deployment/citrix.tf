# Citrix Cloud customer provider settings
provider "citrix" {
  cvad_config = {
    customer_id                 = var.provider_customer_id
    environment                 = var.provider_environment
    client_id                   = var.provider_client_id
    client_secret               = var.provider_client_secret
  }
}
