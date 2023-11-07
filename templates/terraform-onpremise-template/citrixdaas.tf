provider "citrix" {
  hostname                    = "<DDC public IP / hostname>"
  client_id                   = "<DomainFqdn>\\<Admin Username>"
  client_secret               = "<Admin Passwd>"
  disable_ssl_verification    = true # omit this field if DDC has valid SSL certificate configured 
}
