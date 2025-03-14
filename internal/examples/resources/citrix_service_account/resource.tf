resource citrix_service_account "example-azuread-service-account" {
    display_name = "example-azuread-service-account"
    description = "created with terraform"
    identity_provider_type = "AzureAD"
    identity_provider_identifier = "<Azure-Tenant-ID>"
    account_id = "<Application-ID>"
    account_secret = "<Application-Secret>"
    account_secret_format = "PlainText"
    enable_intune_enrolled_device_management = true
    secret_expiry_time = "2099-12-31"
}

resource citrix_service_account "example-ad-service-account" {
    display_name = "example-ad-service-account"
    description = "created with terraform"
    identity_provider_type = "ActiveDirectory"
    identity_provider_identifier = "domain.com" # Domain name
    account_id = "domain\\admin" # Admin user name
    account_secret = "admin-secret" # Admin password
    account_secret_format = "PlainText"
}