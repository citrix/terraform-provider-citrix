# Terraform Module for Citrix StoreFront

This Terraform module allows you to manage resources in Citrix StoreFront.

## Table of Contents
- [Terraform Module for Citrix StoreFront](#terraform-module-for-citrix-storefront)
  - [Table of Contents](#table-of-contents)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
    - [StoreFront configuration for provider](#storefront-configuration-for-provider)
  - [Usage](#usage)
    - [Create a deployment](#create-a-deployment)
    - [Create an authentication service](#create-an-authentication-service)
    - [Create a store service](#create-a-store-service)
    - [Create a webreceiver service](#create-a-webreceiver-service)

## Prerequisites

- Terraform 0.14.x
- The machine running the provider needs to be running on Windows 10+ or Server 2016+
- The machine running the provider needs WinRM access to the specified StoreFront server ([Microsoft docs on how to enable WinRM](https://learn.microsoft.com/en-us/troubleshoot/windows-server/remote/how-to-enable-windows-remote-shell))

## Installation

If running the StoreFront provider on a machine other than the machine where StoreFront is installed, please provide the Active Directory Admin credentials in either environment variables or provider configuration
  - `SF_COMPUTER_NAME`: 
      - The name of the remote computer where the StoreFront server is running.
  - `SF_AD_ADMIN_USERNAME`: 
      - The Active Directory Admin username to connect to the remote PowerShell of the StoreFront Server machine.
  - `SF_AD_ADMIN_PASSWORD`: 
      - The Active Directory Admin password to connect to the remote PowerShell of the StoreFront server machine.


### StoreFront configuration for provider

  ```hcl
  provider "citrix" {
    hostname      = 
    customer_id   = 
    environment   = 
    client_id     = 
    client_secret = 
    disable_ssl_verification =
      storefront_remote_host = {
      computer_name = "{Name of the remote computer where the StoreFront located}"
      ad_admin_username ="{Active Directory Admin Username}"
      ad_admin_password ="{Active Directory Admin Password}"
    }
  }
  ```


## Usage
Example Usage of the StoreFront Terraform Configuration

### Create a deployment

```hcl
resource citrix_stf_deployment "testSTFDeployment" {
		site_id      = 1
		host_base_url = "https://example3.storefront.com"
}
```

### Create an authentication service
```hcl
resource "citrix_stf_authentication_service" "example-stf-authentication-service" {
  site_id       = "${citrix_stf_deployment.testSTFDeployment.site_id}"
  friendly_name = "Auth"
  virtual_path  = "/Citrix/Authentication"
}
```

### Create a store service
```hcl
resource "citrix_stf_store_service" "example-stf-store-service" {
	site_id      = "${citrix_stf_deployment.testSTFDeployment.site_id}"
	virtual_path = "/Citrix/Store"
	friendly_name = "Store"
	authentication_service = "${citrix_stf_authentication_service.example-stf-authentication-service.virtual_path}"
  farm_config = {
    farm_name = "Controller"
    farm_type = "XenDesktop"
    servers = ["cvad.storefront.com"] 
  }
}
```

### Create a webreceiver service
```hcl
resource "citrix_stf_webreceiver_service" "example-stf-webreceiver-service"{
  site_id      = "${citrix_stf_deployment.testSTFDeployment.site_id}"
	virtual_path = "/Citrix/StoreWeb"
	friendly_name = "Receiver2"
  store_service = "${citrix_stf_store_service.example-stf-store-service.virtual_path}"
  authentication_methods = [
      "ExplicitForms", 
    ]
  plugin_assistant = {
    enabled = true
    html5_single_tab_launch = true
    upgrade_at_login = true
    html5_enabled = "Off"
  }
}
```

