// On-Premises customer provider settings
// Please comment out / remove this provider settings block if you are a Citrix Cloud customer
provider "citrix" {
  hostname                    = "<DDC public IP / hostname>"
  client_id                   = "<DomainFqdn>\\<Admin Username>"
  client_secret               = "<Admin Passwd>"
  disable_ssl_verification    = true # omit this field if DDC has valid SSL certificate configured 
}

// Citrix Cloud customer provider settings
// Please comment out / remove this provider settings block if you are an On-Premises customer
provider "citrix" {
  customer_id                 = "" # set your customer id
  client_id                   = ""
  client_secret               = "" # API key client id and secret are needed to interact with Citrix Cloud APIs. These can be created/found under Identity and Access Management > API Access
  environment                 = "Production" # use "Japan" for Citrix Cloud customers in Japan region
}
