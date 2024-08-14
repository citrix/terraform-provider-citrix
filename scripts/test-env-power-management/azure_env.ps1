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

.Parameter Hostname
    The Host name / base URL of Citrix DaaS service.
    For Citrix on-premises customers (Required): Use this to specify Delivery Controller hostname.
    For Citrix Cloud customers (Optional): Use this to force override the Citrix DaaS service hostname.

.Parameter Environment
    The Citrix Cloud environment of the customer. Only applicable for Citrix Cloud customers. Available options: Production, Staging

.Parameter SetDependencyRelationship
    Create dependency relationships between resources by replacing resource IDs with resource references.

.Parameter DisableSSLValidation
    Disable SSL validation for this script. Required if DDC does not have a valid SSL certificate.

#>  

[CmdletBinding()]
Param (
    [Parameter(Mandatory = $false)]
    [string] $CustomerId = "CitrixOnPremises",

    [Parameter(Mandatory = $true)]
    [string] $ClientId,
    
    [Parameter(Mandatory = $true)]
    [string] $ClientSecret,

    [Parameter(Mandatory = $true)]
    [string] $DomainFqdn,

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
    [string] $AzureResourceGroupName,

    [Parameter(Mandatory = $true)]
    [string] $AzureDdcVmName,

    [Parameter(Mandatory = $true)]
    [string] $AzureAdVmName,

    [Parameter(Mandatory = $false)]
    [bool] $DisableSSLValidation = $false
)

function Get-Me {
    $base64Auth = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes(("{0}:{1}" -f "$DomainFqdn\$ClientId", $ClientSecret)))
    try {
        $response = Invoke-RestMethod -Uri "https://$Hostname/citrix/orchestration/api/techpreview/tokens" -Method POST -Headers @{ "Authorization" = "Basic $base64Auth" }
    } catch {
        if ($_.Exception.Response.StatusCode.value__ -ne 200) {
            Write-Host "Failed to get auth token. Status Code: $($_.Exception.Response.StatusCode.value__)"
            Write-Host "Error: $($_.Exception.Message)"
            return $false
        }
    }

    $token = $response.Token
    try {
        Invoke-RestMethod -Uri "https://$Hostname/citrix/orchestration/api/techpreview/me" -Method GET -Headers @{ "Authorization" = "Bearer $token" }
    } catch {
        if ($_.Exception.Response.StatusCode.value__ -ne 200) {
            Write-Host "Failed to get Site"
            return $false
        }
    }

    return $true
}

# Azure Env Booting
## Establish Azure Context
$SecureStringPwd = $AzureClientSecret | ConvertTo-SecureString -AsPlainText -Force
$pscredential = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList $AzureClientId, $SecureStringPwd
Connect-AzAccount -ServicePrincipal -Credential $pscredential -TenantId $AzureTenantId -SubscriptionId $AzureSubscriptionId

## Get the VMs
$ddcVm = Get-AzVM -ResourceGroupName $AzureResourceGroupName -Name $AzureDdcVmName -Status
$adVm = Get-AzVM -ResourceGroupName $AzureResourceGroupName -Name $AzureAdVmName -Status

## Check the power state of the VMs
if ($ddcVm.Statuses[1].Code -ne "PowerState/running") {
    Start-AzVM -ResourceGroupName $AzureResourceGroupName -Name $AzureDdcVmName
}

if ($adVm.Statuses[1].Code -ne "PowerState/running") {
    Start-AzVM -ResourceGroupName $AzureResourceGroupName -Name $AzureAdVmName
}

# Poll for the orchestration service to be available
## Disable SSL validation for test env
if ($DisableSSLValidation) {
    Write-Host "Disabling SSL..."
    if (-not("dummy" -as [type])) {
        add-type -TypeDefinition @"
using System;
using System.Net;
using System.Net.Security;
using System.Security.Cryptography.X509Certificates;

public static class Dummy {
    public static bool ReturnTrue(object sender,
        X509Certificate certificate,
        X509Chain chain,
        SslPolicyErrors sslPolicyErrors) { return true; }

    public static RemoteCertificateValidationCallback GetDelegate() {
        return new RemoteCertificateValidationCallback(Dummy.ReturnTrue);
    }
}
"@
    }
    
    [System.Net.ServicePointManager]::ServerCertificateValidationCallback = [dummy]::GetDelegate()
    [Net.ServicePointManager]::SecurityProtocol = "tls12, tls11, tls"
}

## Poll for GetMe API to return 200
$timeout = 300
$curTime = Get-Date
$endTime = $curTime.AddSeconds($timeout)
$success = $false
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
    Write-Host "Orchestration service is not available after $timeout seconds."
    exit 1
}

exit 0