# On-Premises customer provider settings
# Please comment out / remove this provider settings block if you are a Citrix Cloud customer
provider "citrix" {
  hostname                    = var.provider_hostname
  client_id                   = "${var.provider_domain_fqdn}\\${var.provider_client_id}"
  client_secret               = "${var.provider_client_secret}"
  disable_ssl_verification    = var.provider_disable_ssl_verification
}

# Citrix Cloud customer provider settings
# Please comment out / remove this provider settings block if you are an On-Premises customer
provider "citrix" {
  customer_id                 = var.provider_customer_id
  client_id                   = var.provider_client_id
  client_secret               = var.provider_client_secret
  environment                 = var.provider_environment
}
