# Terraform Module for Citrix StoreFront

This Terraform module allows you to manage resources in Citrix StoreFront.

## Table of Contents
- [Terraform Module for Citrix StoreFront](#terraform-module-for-citrix-storefront)
  - [Table of Contents](#table-of-contents)
  - [Prerequisites](#prerequisites)
  - [PowerShell Remoting on Storefront Remote Server](#powershell-remoting-on-storefront-remote-server)
    - [Enable Remoting using HTTPS (recommended)](#enable-remoting-using-https-recommended)
    - [Enable Remoting using HTTP](#enable-remoting-using-http)
    - [Verification of Connectivity](#verification-of-connectivity)
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

## PowerShell Remoting on Storefront Remote Server

PowerShell Remoting uses Windows Remote Management (WinRM) to allow users to run PowerShell commands on remote computers. PowerShell Remoting (and WinRM) listen on the following ports:

- HTTP: 5985
- HTTPS: 5986

### Enable Remoting using HTTPS (recommended)

1. Open PowerShell as Administrator on the storefront remote server to run the following commands.
2. Enable PowerShell Remoting (WinRM): 
    * `Enable-PSRemoting -Force`
3. Create a self signed cert on the storefront remote server
    * `$fqdn = [System.Net.Dns]::GetHostByName($env:computerName).HostName`
    * `$Cert = New-SelfSignedCertificate -CertstoreLocation Cert:\LocalMachine\My -DnsName $fqdn`
    * `Export-Certificate -Cert $Cert -FilePath 'C:\Users\Public\Desktop\exch.cer'`
4. Create a firewall rule he storefront remote server
    * `New-Item -Path WSMan:\LocalHost\Listener -Transport HTTPS -Address * -CertificateThumbPrint $Cert.Thumbprint -Force`
    * `New-NetFirewallRule -DisplayName 'WinRM HTTPS-In' -Name 'WinRM HTTPS-In' -Profile Any -LocalPort 5986 -Protocol TCP`
5. Copy and install the new cert `exch.cer` created on the desktop on your local development server

### Enable Remoting using HTTP

1.  Open PowerShell as Administrator on the storefront remote server to run the following commands.
2.  Enable PowerShell Remoting (WinRM): 
    * `Enable-PSRemoting -Force`
3.  Configure WinRM HTTPS Listener (Optional)
    * `New-SelfSignedCertificate -DnsName "localhost" -CertStoreLocation "Cert:\LocalMachine\My"`
    * `$thumbprint = (Get-ChildItem -Path Cert:\LocalMachine\My | Where-Object {$_.Subject -like "*localhost*"}).Thumbprint`
    * `$cmd = "winrm create winrm/config/Listener?Address=*+Transport=HTTPS '@{Hostname=""localhost""; CertificateThumbprint=""$thumbprint""}'"`
    * `Invoke-Expression $cmd`
4. Configure Firewall
    * `New-NetFirewallRule -DisplayName "WinRM HTTP" -Name "WinRM-HTTP-In-TCP" -Enabled True -Direction Inbound -Protocol TCP -LocalPort 5985 -Action Allow`
    * `New-NetFirewallRule -DisplayName "WinRM HTTPS" -Name "WinRM-HTTPS-In-TCP" -Enabled True -Direction Inbound -Protocol TCP -LocalPort 5986 -Action Allow`
5.  Now, Open PowerShell as Administrator on the local server(development machine) to run the following commands to add storefront       server to trusted host
    * `Enable-PSRemoting -Force`
    * `Set-Item WSMan:\localhost\Client\TrustedHosts -Value <Public IP of Storefront Machine>`

### Verification of Connectivity

Open PowerShell as Administrator on your local development server and run the following commands to establish a remote PS Session
   * `$securePass = ConvertTo-SecureString -Force -AsPlainText '<password>'`
   * `$credential = New-Object System.Management.Automation.PSCredential ('<domain>\<username>', $securePass)`
   * `Enter-PSSession -ConnectionUri http://<public_ip>:<5985/5986> -Credential $credential -SessionOption (New-PSSessionOption -SkipCACheck -SkipCNCheck -SkipRevocationCheck) -Authentication Negotiate`
   

## Installation

If running the StoreFront provider on a machine other than the machine where StoreFront is installed, please provide the Active Directory Admin credentials in either environment variables or provider configuration
  - `SF_COMPUTER_NAME`: 
      - The name of the remote computer where the StoreFront server is running.
  - `SF_AD_ADMIN_USERNAME`: 
      - The Active Directory Admin username to connect to the remote PowerShell of the StoreFront Server machine.
  - `SF_AD_ADMIN_PASSWORD`: 
      - The Active Directory Admin password to connect to the remote PowerShell of the StoreFront server machine.

~~~~
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
resource "citrix_stf_deployment" "example-stf-deployment" {
		site_id       = "1"
		host_base_url = "https://example3.storefront.com"
}
```

### Create an authentication service
```hcl
resource "citrix_stf_authentication_service" "example-stf-authentication-service" {
  site_id       = citrix_stf_deployment.example-stf-deployment.site_id
  friendly_name = "Auth"
  virtual_path  = "/Citrix/Authentication"
}
```

### Create a store service
```hcl
resource "citrix_stf_store_service" "example-stf-store-service" {
	site_id      = citrix_stf_deployment.example-stf-deployment.site_id
	virtual_path = "/Citrix/Store"
	friendly_name = "Store"
	authentication_service_virtual_path = "${citrix_stf_authentication_service.example-stf-authentication-service.virtual_path}"
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
  site_id      = citrix_stf_deployment.example-stf-deployment.site_id
	virtual_path = "/Citrix/StoreWeb"
	friendly_name = "Receiver2"
  store_virtual_path = "${citrix_stf_store_service.example-stf-store-service.virtual_path}"
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

