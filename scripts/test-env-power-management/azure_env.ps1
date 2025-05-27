# Copyright © 2024. Citrix Systems, Inc. All Rights Reserved.
<#
Currently this script is still in TechPreview
.SYNOPSIS
    Script to prepare an Azure test environment for test suite.
    
.DESCRIPTION 
    The script will check the power state of the VMs in the Azure environment for running Azure Test Suite. 
    It will boot up the VMs if they are powered down, and poll for orchestration service to be available.

.Parameter CustomerId
    The Citrix Cloud customer ID. Only applicable for Citrix Cloud customers. 
    When omitted, the default value is "CitrixOnPremises" for on-premises use case.

.Parameter ClientId
    The Client Id for Citrix DaaS service authentication.
    For Citrix on-premises customers: Use this to specify a DDC administrator username.
    For Citrix Cloud customers: Use this to specify Cloud API Key Client Id.
    
.Parameter ClientSecret
    The Client Secret for Citrix DaaS service authentication.
    For Citrix on-premises customers: Use this to specify a DDC administrator password.
    For Citrix Cloud customers: Use this to specify Cloud API Key Client Secret.

.Parameter DomainFqdn
    The domain FQDN of the Active Directory in test environment. Only required for on-premises customers.

.Parameter CitrixCloudHostname
    The base URL of Citrix Cloud service. Only required for Citrix Cloud customers.

.Parameter Hostname
    The Host name / base URL of Citrix DaaS service.
    For Citrix on-premises customers (Required): Use this to specify Delivery Controller hostname.
    For Citrix Cloud customers (Optional): Use this to force override the Citrix DaaS service hostname.

.Parameter AzureClientId
    The Client Id for Azure SPN used for powering on the Azure VMs for test environment.

.Parameter AzureClientSecret
    The Client Secret for Azure SPN used for powering on the Azure VMs for test environment.

.Parameter AzureTenantId
    The Tenant Id for Azure SPN used for powering on the Azure VMs for test environment.

.Parameter AzureSubscriptionId
    The Subscription Id for powering on the Azure VMs for test environment.

.Parameter AzureAdVmResourceGroupName
    The Resource Group Name for powering on the Azure Active Directory VM for test environment.

.Parameter AzureAdVmName
    The Azure VM name of the Active Directory Domain Controller.

.Parameter AzureConnectorResourceGroupName
    The Resource Group Name for powering on the Connector VMs for test environment.

.Parameter AzureConnectorVm1Name
    The Azure VM name of the Citrix Cloud Connector 1.

.Parameter AzureConnectorVm2Name
    The Azure VM name of the Citrix Cloud Connector 2.

.Parameter AzureDdcResourceGroupName
    The Resource Group Name for powering on the DDC VM for test environment.

.Parameter AzureDdcVmName
    The Azure VM name of the DDC.

.Parameter DisableSSLValidation
    Disable SSL validation for this script. Required if DDC does not have a valid SSL certificate.

.Parameter OnPremises
    Set to true if the script is used for powering on the on-premise test environment.

.Parameter OrchestrationPollingTimeout
    Timeout in seconds for polling the orchestration service to be available. Default is 600 seconds (10 minutes).

.Parameter SkipOrchestrationPolling
    Set this flag to skip polling for orchestration service to be available.

#>  

[CmdletBinding()]
Param (
    [Parameter(Mandatory = $false)]
    [string] $CustomerId = "CitrixOnPremises",

    [Parameter(Mandatory = $true)]
    [string] $ClientId,
    
    [Parameter(Mandatory = $true)]
    [string] $ClientSecret,

    [Parameter(Mandatory = $false)]
    [string] $DomainFqdn,

    [Parameter(Mandatory = $false)]
    [string] $CitrixCloudHostname,

    [Parameter(Mandatory = $true)]
    [string] $Hostname,

    [Parameter(Mandatory = $true)]
    [string] $AzureClientId,

    [Parameter(Mandatory = $true)]
    [string] $AzureClientSecret,
    
    [Parameter(Mandatory = $true)]
    [string] $AzureTenantId,

    [Parameter(Mandatory = $true)]
    [string] $AzureSubscriptionId,

    [Parameter(Mandatory = $true)]
    [string] $AzureAdVmResourceGroupName,

    [Parameter(Mandatory = $true)]
    [string] $AzureAdVmName,

    [Parameter(Mandatory = $false)]
    [string] $AzureDdcResourceGroupName,

    [Parameter(Mandatory = $false)]
    [string] $AzureDdcVmName,

    [Parameter(Mandatory = $false)]
    [string] $AzureConnectorResourceGroupName,
    
    [Parameter(Mandatory = $false)]
    [string] $AzureConnectorVm1Name,

    [Parameter(Mandatory = $false)]
    [string] $AzureConnectorVm2Name,

    [Parameter(Mandatory = $false)]
    [bool] $DisableSSLValidation = $false,

    [Parameter(Mandatory = $false)]
    [bool] $OnPremises = $true,

    [Parameter(Mandatory = $false)]
    [int] $OrchestrationPollingTimeout = 600,

    [Parameter(Mandatory = $false)]
    [switch] $SkipOrchestrationPolling
)

function Get-CCAuthToken {
    param (
        [Parameter(Mandatory = $true)]
        [string] $ccHostname,

        [Parameter(Mandatory = $true)]
        [string] $customerId,

        [Parameter(Mandatory = $true)]
        [string] $clientId,

        [Parameter(Mandatory = $true)]
        [string] $clientSecret
    )

    $ccauth_url = "https://$ccHostname/cctrustoauth2/$customerId/tokens/clients"
    Write-Host "Requesting CC Auth Token from $ccauth_url"

    $clientSecret_encoded = [uri]::EscapeDataString($clientSecret)
    $body = @{
        grant_type = "client_credentials"
        client_id = $clientId
        client_secret = $clientSecret_encoded
    }
    $bodyString = "grant_type=client_credentials&client_id=$clientId&client_secret=$clientSecret_encoded"

    $response = Invoke-RestMethod -Uri $ccauth_url -Method POST -Body $bodyString -ContentType "application/x-www-form-urlencoded"
    $token = $response.access_token
    $authHeader = "CwsAuth Bearer=$token"

    return $authHeader
}

function Start-DaasService {
    param (
        [Parameter(Mandatory = $true)]
        [string] $ccHostname,

        [Parameter(Mandatory = $true)]
        [string] $hostname,

        [Parameter(Mandatory = $true)]
        [string] $customerId,

        [Parameter(Mandatory = $true)]
        [string] $clientId,

        [Parameter(Mandatory = $true)]
        [string] $clientSecret
    )

    $authHeader = Get-CCAuthToken -ccHostname $ccHostname -customerId $customerId -clientId $clientId -clientSecret $clientSecret
    $url = "https://$hostname/resourceprovider/$customerId/site/activation?customerId=$customerId"
    Invoke-RestMethod -Uri $url -Method POST -Headers @{ "Authorization" = "$authHeader" }
}

function Start-AzureVm {
    param (
        [Parameter(Mandatory = $true)]
        [string] $ResourceGroupName,

        [Parameter(Mandatory = $true)]
        [string] $VmName
    )

    $vm = Get-AzVM -ResourceGroupName $ResourceGroupName -Name $VmName -Status
    if ($vm.Statuses[1].Code -ne "PowerState/running") {
        Start-AzVM -ResourceGroupName $ResourceGroupName -Name $VmName
    }
}

function Get-Me {
    if ($OnPremises) {
        try {
            $base64Auth = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes(("{0}:{1}" -f "$DomainFqdn\$ClientId", $ClientSecret)))
            if ($DisableSSLValidation) {
                $response = Invoke-RestMethod -Uri "https://$Hostname/citrix/orchestration/api/techpreview/tokens" -Method POST -Headers @{ "Authorization" = "Basic $base64Auth" } -SkipCertificateCheck
            } else {
                $response = Invoke-RestMethod -Uri "https://$Hostname/citrix/orchestration/api/techpreview/tokens" -Method POST -Headers @{ "Authorization" = "Basic $base64Auth" }
            }
        } catch {
            if ($_.Exception.Response.StatusCode.value__ -ne 200) {
                Write-Host "Failed to get auth token. Status Code: $($_.Exception.Response.StatusCode.value__)"
                Write-Host "Error: $($_.Exception.Message)"
                return $false
            }
        }
    
        $token = $response.Token
        try {
            
            if ($DisableSSLValidation) {
                Invoke-RestMethod -Uri "https://$Hostname/citrix/orchestration/api/techpreview/me" -Method GET -Headers @{ "Authorization" = "Bearer $token" } -SkipCertificateCheck | Out-Null
            } else {
                Invoke-RestMethod -Uri "https://$Hostname/citrix/orchestration/api/techpreview/me" -Method GET -Headers @{ "Authorization" = "Bearer $token" } | Out-Null
            }
        } catch {
            if ($_.Exception.Response.StatusCode.value__ -ne 200) {
                Write-Host "Failed to get Site"
                return $false
            }
        }
    } else {
        try {
            $authHeader = Get-CCAuthToken -cchostname $CitrixCloudHostname -customerId $CustomerId -clientId $ClientId -clientSecret $ClientSecret
        } catch {
            if ($_.Exception.Response.StatusCode.value__ -ne 200) {
                Write-Host "Failed to get auth token. Status Code: $($_.Exception.Response.StatusCode.value__)"
                Write-Host "Error: $($_.Exception.Message)"
                return $false
            }
        }
    
        try {
            Invoke-RestMethod -Uri "https://$Hostname/cvad/manage/me" -Method GET -Headers @{ "Authorization" = "$authHeader"; "Citrix-CustomerId" = "$CustomerId" } | Out-Null
        } catch {
            if ($_.Exception.Response.StatusCode.value__ -ne 200) {
                Write-Host "Failed to get Site"
                return $false
            }
        }
    }

    return $true
}

# Azure Env Booting
## Establish Azure Context
$SecureStringPwd = $AzureClientSecret | ConvertTo-SecureString -AsPlainText -Force
$pscredential = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList $AzureClientId, $SecureStringPwd
Connect-AzAccount -ServicePrincipal -Credential $pscredential -TenantId $AzureTenantId -SubscriptionId $AzureSubscriptionId

## Get the VMs and power them up if they are not running
Start-AzureVm -ResourceGroupName $AzureAdVmResourceGroupName -VmName $AzureAdVmName

if ($OnPremises -eq $true) {
    Start-AzureVm -ResourceGroupName $AzureDdcResourceGroupName -VmName $AzureDdcVmName
} else {
    if ($AzureConnectorVm1Name) {
        Start-AzureVm -ResourceGroupName $AzureConnectorResourceGroupName -VmName $AzureConnectorVm1Name
    }
    if ($AzureConnectorVm2Name) {
        Start-AzureVm -ResourceGroupName $AzureConnectorResourceGroupName -VmName $AzureConnectorVm2Name
    }
    Start-DaasService -ccHostname $CitrixCloudHostname -hostname $Hostname -customerId $CustomerId -clientId $ClientId -clientSecret $ClientSecret
}

# Skip polling for orchestration if the flag is set
if ($SkipOrchestrationPolling) {
    exit 0
}

# Poll for the orchestration service to be available
## Poll for GetMe API to return 200
$curTime = Get-Date
$endTime = $curTime.AddSeconds($OrchestrationPollingTimeout)
$success = $false
Write-Host "Start polling Orchestration service. Timeout is set to $OrchestrationPollingTimeout seconds."
while ($curTime -le $endTime) {
    $success = Get-Me
    if ($success) {
        Write-Host "Orchestration service is available."
        break
    }
    Write-Host "Orchestration service is not available. Retrying in 10 seconds..."
    Start-Sleep -Seconds 10
    $curTime = Get-Date
}

if ($success -eq $false) {
    Write-Error "Orchestration service is not available after $OrchestrationPollingTimeout seconds."
    exit 1
}

exit 0