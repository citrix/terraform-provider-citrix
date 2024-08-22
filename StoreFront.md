# Terraform Module for Citrix StoreFront

This Terraform module allows you to manage resources in Citrix StoreFront.

## Table of Contents
- [Terraform Module for Citrix StoreFront](#terraform-module-for-citrix-storefront)
  - [Table of Contents](#table-of-contents)
  - [Prerequisites](#prerequisites)
    - [PowerShell Remoting on StoreFront Remote Server](#powershell-remoting-on-storefront-remote-server)
      - [Enable Remoting using HTTPS](#enable-remoting-using-https)
      - [Verification of Connectivity](#verification-of-connectivity)
  - [Installation](#installation)
    - [Provider Block](#provider-block)
    - [Resource Block](#resource-block)
      - [Create a deployment](#create-a-deployment)
      - [Create an authentication service](#create-an-authentication-service)
      - [Create a store service](#create-a-store-service)
      - [Create a webreceiver service](#create-a-webreceiver-service)
  - [Manual Cleanup](#manual-cleanup)

## Prerequisites

- Terraform 0.14.x
- The provider needs to either run locally (recommended) on the StoreFront server, or have WinRM access to it. In the latter case follow the instructions in the next section to config WinRM on StoreFront Remote Server.

### PowerShell Remoting on StoreFront Remote Server
- The machine running the provider needs to be running on Windows 10+ or Server 2016+
- The machine running the provider needs WinRM access to the specified StoreFront server ([Microsoft docs on how to enable WinRM](https://learn.microsoft.com/en-us/troubleshoot/windows-server/remote/how-to-enable-windows-remote-shell))

PowerShell Remoting uses Windows Remote Management (WinRM) to allow users to run PowerShell commands on remote computers. PowerShell Remoting (and WinRM) listen on the following ports:

- HTTP: 5985
- HTTPS: 5986

#### Enable Remoting using HTTPS

1. Open PowerShell as Administrator on the storefront remote server to run the following commands.
2. Enable PowerShell Remoting (WinRM): 
    * `Enable-PSRemoting -Force`
3. Create a self signed cert on the storefront remote server
    * `$fqdn = [System.Net.Dns]::GetHostByName($env:computerName).HostName`
    * `$Cert = New-SelfSignedCertificate -CertstoreLocation Cert:\LocalMachine\My -DnsName $fqdn`
    * `Export-Certificate -Cert $Cert -FilePath 'C:\Users\Public\Desktop\exch.cer'`
4. Create a firewall rule on the storefront remote server
    * `New-Item -Path WSMan:\LocalHost\Listener -Transport HTTPS -Address * -CertificateThumbPrint $Cert.Thumbprint -Force`
    * `New-NetFirewallRule -DisplayName 'WinRM HTTPS-In' -Name 'WinRM HTTPS-In' -Profile Any -LocalPort 5986 -Protocol TCP`
5. Copy and install the new cert `exch.cer` created on the desktop on your local development server

#### Verification of Connectivity

Open PowerShell as Administrator on your local development server and run the following commands to establish a remote PS Session
   * `$securePass = ConvertTo-SecureString -Force -AsPlainText '<password>'`
   * `$credential = New-Object System.Management.Automation.PSCredential ('<domain>\<username>', $securePass)`
   * `Enter-PSSession -ConnectionUri https://<public_ip>:5986 -Credential $credential -Authentication Negotiate`
   

## Installation

### Provider Block
If running the StoreFront provider on storefront locally
~~~~
provider "citrix" {
  storefront_remote_host = {}
}
~~~~

If running the StoreFront provider on a machine other than the machine where StoreFront is installed, please provide the Active Directory Admin credentials in either environment variables or provider configuration
  - `SF_COMPUTER_NAME`: 
      - The name of the remote computer where the StoreFront server is running.
  - `SF_AD_ADMIN_USERNAME`: 
      - The Active Directory Admin username to connect to the remote PowerShell of the StoreFront Server machine.
  - `SF_AD_ADMIN_PASSWORD`: 
      - The Active Directory Admin password to connect to the remote PowerShell of the StoreFront server machine.
  - `SF_DISABLE_SSL_VERIFICATION`: 
      - An optional toggle that allows you to disable SSL verification when connecting to the target StoreFront server using Terraform. This can be useful in development or testing environments where SSL certificates may not be properly configured or trusted. However, it is recommended to keep SSL verification enabled in production environments to ensure secure communication.

~~~~
provider "citrix" {
  storefront_remote_host = {
    computer_name = "{public IP of the storefront VM}"
    ad_admin_username ="{Active Directory Admin Username}"
    ad_admin_password ="{Active Directory Admin Password}"
    disable_ssl_verification = false
  }
}
~~~~

### Resource Block
Example Usage of the StoreFront Terraform Configuration


#### Create a deployment 
~~~~
resource "citrix_stf_deployment" "example-stf-deployment" {
  site_id       = "1"
  host_base_url = "http://<localhost name>"
}
~~~~

#### Create an authentication service
~~~~
resource "citrix_stf_authentication_service" "example-stf-authentication-service" {
  site_id       = citrix_stf_deployment.example-stf-deployment.site_id
  friendly_name = "Auth"
  virtual_path  = "/Citrix/Authentication"

  depends_on = [ citrix_stf_deployment.example-stf-deployment ] //Required dependency 
}
~~~~

#### Create a store service
~~~~
resource "citrix_stf_store_service" "example-stf-store-service" {
  site_id      = citrix_stf_deployment.example-stf-deployment.site_id
  virtual_path = "/Citrix/Store"
  friendly_name = "Store"
  authentication_service_virtual_path = "${citrix_stf_authentication_service.example-stf-authentication-service.virtual_path}"
  pna = {
    enable = true
  }
  farms = [
    {
      farm_name = "Controller1"
      farm_type = "XenDesktop"
      servers = ["cvad.storefront.com"] 
      port = 80
    },
    {
      farm_name = "Controller2"
      farm_type = "XenDesktop"
      servers = ["cvad.storefront2.com"] 
      port = 443
      zones = ["Primary"]
    }
  ]
  
  // Add depends_on attribute to ensure the StoreFront Store with Authentication is created after the Authentication Service
  depends_on = [ citrix_stf_authentication_service.example-stf-authentication-service ]
}
~~~~

#### Create a webreceiver service
~~~~
resource "citrix_stf_webreceiver_service" "example-stf-webreceiver-service"{
  site_id      = citrix_stf_deployment.example-stf-deployment.site_id
  virtual_path = "/Citrix/StoreWeb"
  friendly_name = "ReceiverWeb"
  store_virtual_path = "${citrix_stf_store_service.example-stf-store-service.virtual_path}"
  depends_on = [ citrix_stf_store_service.example-stf-store-service ]
}
~~~~


## Manual Cleanup
If there's any need for manual cleanup, the following steps can be followed:

1. Either use remote PowerShell or any preferred method to access the PowerShell of the StoreFront server with Administrator privilege.
2. Enter and execute command `Clear-STFDeployment -Confirm:$false` to fully cleanup the StoreFront deployment.
3. If the command failed to cleanup the StoreFront deployment, please try to uninstall StoreFront from the server to remove all the leftover resources. Then, install a fresh new copy of StoreFront to create a fresh new environment to work with.
