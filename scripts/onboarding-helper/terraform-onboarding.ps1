
# Copyright © 2024. Citrix Systems, Inc. All Rights Reserved.
<#
Currently this script is still in TechPreview
.SYNOPSIS
    Script to onboard an existing site to terraform. 
    
.DESCRIPTION 
    The script should be able to collect the list of resources from DDC, import into terraform, and generate the TF skeletons.

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

    [Parameter(Mandatory = $false)]
    [string] $DomainFqdn,

    [Parameter(Mandatory = $false)]
    [string] $Hostname = "api.cloud.com",

    [Parameter(Mandatory = $false)]
    [ValidateSet("Production", "Staging")]
    [string] $Environment = "Production",

    [Parameter(Mandatory = $false)]
    [switch] $SetDependencyRelationship,

    [Parameter(Mandatory = $false)]
    [switch] $DisableSSLValidation
)

### Helper Functions ###

function Get-Site {
    if ($script:onPremise) {
        $siteRequest = "https://$script:hostname/citrix/orchestration/api/me"
    }
    else {
        $siteRequest = "https://$script:hostname/cvad/manage/me"
    }

    $response = Start-GetRequest -url $siteRequest
    $script:siteId = $response.Customers[0].Sites[0].Id
}

function Get-RequestBaseUrl {
    if ($script:onPremise) {
        $url = "https://$script:hostname/citrix/orchestration/api/CitrixOnPremises/$script:siteId"
    }
    else {
        $url = "https://$script:hostname/cvad/manage"
    }

    $script:urlBase = $url
}


function Invoke-WebRequestWithRetry {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Uri,

        [Parameter(Mandatory = $true)]
        [string]$Method,

        [Parameter(Mandatory = $false)]
        [HashTable]$Headers = @{},

        [Parameter(Mandatory = $false)]
        [string]$ContentType = 'application/json',

        [Parameter(Mandatory = $false)]
        [HashTable]$Body,

        [Parameter(Mandatory = $false)]
        [int]$MaxRetries = 5,

        [Parameter(Mandatory = $false)]
        [double]$JitterFactor = 0.1
    )

    $attempt = 0
    while ($true) {
        try {
            $attempt++
            Write-Verbose "Attempting $Method $Uri..."
            if ($DisableSSLValidation -and $PSVersionTable.PSVersion.Major -ge 7) {
                $response = Invoke-WebRequest -Uri $Uri -Method $Method -Headers $Headers -ContentType $ContentType -Body $Body -SkipCertificateCheck
            }
            else {
                $response = Invoke-WebRequest -Uri $Uri -Method $Method -Headers $Headers -ContentType $ContentType -Body $Body
            }
            
            return $response
        }
        catch {
            if ($attempt -ge $MaxRetries) {
                Write-Verbose "Max retries reached. Throwing exception."
                throw
            }
            else {
                $baseDelay = [math]::Pow(2, $attempt)
                # This is a random delay that is added to the base delay to prevent a thundering herd problem where many instances of the function might be retrying at the same time. 
                # The jitter is a random number between 0 and 10% of the base delay.
                $jitter = Get-Random -Minimum 0 -Maximum ([math]::Ceiling($baseDelay * $JitterFactor))
                $delay = $baseDelay + $jitter
                Write-Verbose "Error occurred, retrying $Method $Uri after $delay seconds..."
                Start-Sleep -Seconds $delay
            }
        }
    }
}

function Get-AuthToken {
    if ($script:onPremise) {
        $url = "https://$script:hostname/citrix/orchestration/api/tokens"
        $base64AuthInfo = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes(("{0}\{1}:{2}" -f $script:domainFqdn, $script:clientId, $script:clientSecret)))
        $basicAuth = "Basic $base64AuthInfo"
        $response = Invoke-WebRequestWithRetry -Uri $url -Method 'POST' -Headers @{Authorization = $basicAuth } 
        $jsonObj = ConvertFrom-Json $([String]::new($response.Content))
        return $jsonObj.Token
    }
    else {
        if ($script:environment -eq "Production") {
            $url = "https://api.cloud.com/cctrustoauth2/$script:customerId/tokens/clients"
        }
        elseif ($script:environment -eq "Staging") {
            $url = "https://api.cloudburrito.com/cctrustoauth2/$script:customerId/tokens/clients"
        }

        $body = @{
            grant_type    = 'client_credentials'
            client_id     = $script:clientId
            client_secret = $script:clientSecret
        }
        $contentType = 'application/x-www-form-urlencoded'
        $response = Invoke-WebRequestWithRetry -Uri $url -Method 'POST' -body $body -ContentType $contentType
        $jsonObj = ConvertFrom-Json $([String]::new($response.Content))
        return $jsonObj.access_token
    }
}

function Start-GetRequest {
    param(
        [parameter(Mandatory = $true)][string] $url
    )

    $token = Get-AuthToken
    if ($script:onPremise) {
        $headers = @{
            Authorization = "Bearer $token"
        }
    }
    else {
        $headers = @{
            "Authorization"     = "CwsAuth Bearer=$token"
            "Citrix-CustomerId" = $script:customerId
        }
        if ($null -ne $script:siteId) {
            $headers["Citrix-InstanceId"] = $script:siteId
        }
    }
    
    $contentType = 'application/json'
    $response = Invoke-WebRequestWithRetry -Uri $url -Method 'GET' -Headers $headers -ContentType $contentType
    $jsonObj = ConvertFrom-Json $([String]::new($response.Content))
    return $jsonObj
}

function New-RequiredFiles {

    # Create temporary import.tf for terraform import
    if (!(Test-Path ".\citrix.tf")) {
        New-Item -path ".\" -name "citrix.tf" -type "file" -Force
        Write-Verbose "Created new file for terraform citrix provider configuration."
    }
    if ($script:onPremise) {
        $disable_ssl_verification = $script:disable_ssl.ToString().ToLower()
        $config = @"
provider "citrix" {
    cvad_config = {
        hostname                    = "$script:hostname"
        client_id                   = "$script:domainFqdn\\$script:clientId"
        # client_secret               = "<Input client secret value>"
        disable_ssl_verification    = $disable_ssl_verification
    }
}
"@
        Set-Content -Path ".\citrix.tf" -Value $config
    }
    else {
        $config = @"
provider "citrix" {
    cvad_config = {
        customer_id                 = "$script:customerId"
        client_id                   = "$script:clientId"
        # client_secret               = "<Input client secret value>"
        hostname                    = "$script:hostname"
        environment                 = "$script:environment"
    }
}
"@
        Set-Content -Path ".\citrix.tf" -Value $config
    }

    if (!(Test-Path ".\import.tf")) {
        New-Item -path ".\" -name "import.tf" -type "file" -Force
        Write-Verbose "Created new file for terraform import."
    }
    else {
        Clear-Content -path ".\import.tf"
        Write-Verbose "Cleared content in terraform import file."
    }

    # Create resource.tf for final terraform resources
    if (!(Test-Path ".\resource.tf")) {
        New-Item -path ".\" -name "resource.tf" -type "file" -Force
        Write-Verbose "Created new file for terraform resource."
    }
    else {
        Clear-Content -path ".\resource.tf"
        Write-Verbose "Cleared content in terraform resource file."
    }

}

# Function to get list of resources for a given resource provider
function Get-ResourceList {
    param(
        [parameter(Mandatory = $true)]
        [string] $requestPath,

        [parameter(Mandatory = $true)]
        [string] $resourceProviderName
    )

    $url = "$script:urlBase/$requestPath"

    $response = Start-GetRequest -url $url
    $items = $response.Items
    $resourceList = @()
    $pathMap = @{}
    foreach ($item in $items) {

        # Handle special case for Machine Catalogs
        if ($requestPath -eq "machinecatalogs" -and $item.provisioningType -ne "Manual" -and $item.provisioningType -ne "MCS" -and $item.provisioningType -ne "PVSStreaming") {
            Write-Warning "Currently the citrix terraform provider only supports Manual and MCS Machine Catalogs. Ignoring the Machine Catalog with Name: $($item.name) and Type: $($item.provisioningType)"
            continue;
        }

        # Handle special case for hypervisors
        if ($requestPath -eq "hypervisors") {
            if (($item.ConnectionType -eq $script:hypervisorResourceMap.$resourceProviderName) -and ($item.ConnectionType -ne "Custom" -or $item.PluginId -eq $NUTANIX_PLUGIN_ID)) {
                $resourceList += $item.Id
            }
            # Skip other hypervisors
            continue
        }

        # Handle special case for Policies
        if ($item.policySetGuid -and $item.policySetType -like "*Policies*") {
            $resourceList += $item.policySetGuid
        }

        # Check for id and ignore empty and default values
        if ($item.Id -and $item.Id -ne "0" -and $item.Id -ne "00000000-0000-0000-0000-000000000000") {
            $resourceList += $item.Id
        }
        
        # Create a path map for ApplicationFolder paths
        if ($requestPath -eq "AdminFolders") {
            $pathMap[$item.Id] = $item.Path
        }
    }
    return $resourceList, $pathMap
}

# Function to get import map for each resource
function Get-ImportMap {
    param(
        [parameter(Mandatory = $true)]
        [string] $resourceApi,

        [parameter(Mandatory = $true)]
        [string] $resourceProviderName,

        [parameter(Mandatory = $false)]
        [string] $parentId = "",

        [parameter(Mandatory = $false)]
        [int] $parentIndex = 0
    )

    $list, $pathMap = Get-ResourceList -requestPath $resourceApi -resourceProviderName $resourceProviderName
    $resourceMap = @{}
    $index = 0
    foreach ($id in $list) {
        if ($parentId -ne "") {
            $resourceName = "$($resourceProviderName)_$($parentIndex)_$($index)"
            $resourceMapKey = "$($parentId),$($id)"
        }
        else {
            $resourceName = "$($resourceProviderName)_$($index)"
            $resourceMapKey = $id
        }
        
        if ($resourceApi -eq "AdminFolders" -and $pathMap.Count -gt 0) {
            $applicationFolderPathMap[$pathMap.$id] = $resourceName
        }
        
        $resourceMap[$resourceMapKey] = $resourceName
        $resourceContent = "resource `"citrix_$resourceProviderName`" `"$resourceName`" {}`n"
        Add-Content -Path ".\import.tf" -Value $resourceContent
        $index += 1
    }

    return $resourceMap
}

# List all CVAD objects from existing site
function Get-ExistingCVADResources {
   
    $resources = @{
        "zone"                 = @{
            "resourceApi"          = "zones"
            "resourceProviderName" = "zone"
        }
        "azure_hypervisor"     = @{
            "resourceApi"          = "hypervisors"
            "resourceProviderName" = "azure_hypervisor"
        }
        "aws_hypervisor"       = @{
            "resourceApi"          = "hypervisors"
            "resourceProviderName" = "aws_hypervisor"
        }
        "gcp_hypervisor"       = @{
            "resourceApi"          = "hypervisors"
            "resourceProviderName" = "gcp_hypervisor"
        }
        "scvmm_hypervisor"     = @{
            "resourceApi"          = "hypervisors"
            "resourceProviderName" = "scvmm_hypervisor"
        }
        "xenserver_hypervisor" = @{
            "resourceApi"          = "hypervisors"
            "resourceProviderName" = "xenserver_hypervisor"
        }
        "vsphere_hypervisor"   = @{
            "resourceApi"          = "hypervisors"
            "resourceProviderName" = "vsphere_hypervisor"
        }
        "nutanix_hypervisor"   = @{
            "resourceApi"          = "hypervisors"
            "resourceProviderName" = "nutanix_hypervisor"
        }
        "machine_catalog"      = @{
            "resourceApi"          = "machinecatalogs"
            "resourceProviderName" = "machine_catalog"
        }
        "delivery_group"       = @{
            "resourceApi"          = "deliverygroups"
            "resourceProviderName" = "delivery_group"
        }
        "admin_scope"          = @{
            "resourceApi"          = "Admin/Scopes"
            "resourceProviderName" = "admin_scope"
        }
        "admin_role"           = @{
            "resourceApi"          = "Admin/Roles"
            "resourceProviderName" = "admin_role"
        }
        "policy_set"           = @{
            "resourceApi"          = "/gpo/policySets"
            "resourceProviderName" = "policy_set"
        }
        "application"      = @{
            "resourceApi"          = "Applications"
            "resourceProviderName" = "application"
        }
        "admin_folder"         = @{
            "resourceApi"          = "AdminFolders"
            "resourceProviderName" = "admin_folder"
        }
        "application_group"    = @{
            "resourceApi"          = "ApplicationGroups"
            "resourceProviderName" = "application_group"
        }
        "application_icon" = @{
            "resourceApi"          = "Icons"
            "resourceProviderName" = "application_icon"
        }
    }

    $script:cvadResourcesMap = @{}

    foreach ($resource in $resources.Keys) {
        $api = $resources[$resource].resourceApi
        $resourceProviderName = $resources[$resource].resourceProviderName
        $script:cvadResourcesMap[$resource] = Get-ImportMap -resourceApi $api -resourceProviderName $resourceProviderName
        
        # Create resource pool map for each hypervisor if exists
        if ($resource -like "*hypervisor") {
            $index = 0
            foreach ($id in $script:cvadResourcesMap[$resource].Keys) {
                $resourcePoolAPI = "hypervisors/$($id)/resourcePools"
                $script:cvadResourcesMap["$($resource)_resource_pool"] = Get-ImportMap -resourceApi $resourcePoolAPI -resourceProviderName "$($resource)_resource_pool" -parentId $id -parentIndex $index
                $index += 1
            }
        }
    }
}

# Function to import terraform resources into state
function Import-ResourcesToState {
    foreach ($resource in  $script:cvadResourcesMap.Keys) {
        foreach ($id in  $script:cvadResourcesMap[$resource].Keys) {
            terraform import "citrix_$($resource).$($script:cvadResourcesMap[$resource][$id])" "$id" 
        }
    }
}

function InjectSecretValues {
    param(
        [parameter(Mandatory = $true)]
        [string] $targetProperty,

        [parameter(Mandatory = $true)]
        [string] $newProperty,

        [parameter(Mandatory = $true)]
        [string] $content
    )

    $regex = "(\s+)$targetProperty(\s+)= (\S+)"
    if ($content -match $regex) {
        $target = $Matches[0]
        $newContent = $target -replace $targetProperty, $newProperty
        $newContent = $newContent -replace "`"\S+`"", "`"<input $($newProperty) value>`""
        if ("username" -eq $targetProperty) {
            # In this case, it would be on-premises hypervisor. We need to have password format.
            $format = $target -replace $targetProperty, "password_format"
            $format = $format -replace "`"\S+`"", "`"PlainText`""
            $content = $content -replace $regex, "$($target)$($newContent)$($format)"
        }
        else {
            $content = $content -replace $regex, "$($target)$($newContent)"
        }
    }

    return $content
}

function RemoveComputedPropertiesForZone {
    param(
        [parameter(Mandatory = $true)]
        [string] $content
    )

    if ($script:onPremise) {
        # Remove resource_location_id property from each zone resource for on-premises 
        $resourceLocationIdRegex = "(\s+)resource_location_id(\s+)= (\S+)"
        $content = $content -replace $resourceLocationIdRegex, ""
    }
    else {
        # Remove name property from each zone resource in cloud
        $filteredOutput = @()
        $lines = $content -split "`r?`n"
        foreach ($line in $lines) {
            if ($line -like 'resource "citrix_zone"*') {
                $insideCitrixZone = $true
            }
        
            if ($insideCitrixZone -and $line -like '*name*') {
                continue
            }
        
            if ($insideCitrixZone -and $line -like '}*') {
                $insideCitrixZone = $false
            }
            $filteredOutput += $line
        }
        $content = $filteredOutput -join "`n"
    }
    return $content
}

function RemoveComputedProperties {
    param(
        [parameter(Mandatory = $true)]
        [string] $content
    )

    # Remove Id property from each resource since they are computed
    $idRegex = "(\s+)id(\s+)= (\S+)"
    $content = $content -replace $idRegex, ""

    # Remove total_machines property from machine_catalog since it is computed
    $totalMachineRegex = "(\s+)total_machines(\s+)= (\S+)"
    $content = $content -replace $totalMachineRegex, ""

    # Remove path property from application_folder since it is computed
    $pathRegex = '(\s+)path\s*=\s*".*\\\\.*"'
    $content = $content -replace $pathRegex, ""

    # Remove assigned property from application since it is computed
    $isAssignedRegex = "(\s+)assigned(\s+)= (\S+)"
    $content = $content -replace $isAssignedRegex, ""

    # Remove is_built_in property from admin_role since it is computed
    $isBuiltInRegex = "(\s+)is_built_in(\s+)= (\S+)"
    $content = $content -replace $isBuiltInRegex, ""

    # Remove contents for zone respource
    $content = RemoveComputedPropertiesForZone -content $content
    
    return $content
}

function ReplaceDependencyRelationships {
    param(
        [parameter(Mandatory = $true)]
        [string] $content
    )

    if (-not $script:SetDependencyRelationship) {
        return $content
    }

    # Create dependency relationships between resources with id references
    foreach ($resource in $script:cvadResourcesMap.Keys) {
        foreach ($id in $script:cvadResourcesMap[$resource].Keys) {
            $content = $content -replace "`"$id`"", "citrix_$($resource).$($script:cvadResourcesMap[$resource][$id]).id"
        }
    }

    # Create dependency relationships between resources with path references
    foreach ( $applicationFolderPath in $script:applicationFolderPathMap.Keys) {
        $path = $applicationFolderPath.replace("\", "\\\\")
        $content = $content -replace "`"$path`"", "citrix_admin_folder.$($script:applicationFolderPathMap[$applicationFolderPath]).path"
    }

    return $content
}

function InjectPlaceHolderSensitiveValues {
    param(
        [parameter(Mandatory = $true)]
        [string] $content
    )

    ### hypervisor secrets ###
    ######   Azure   ######
    $content = InjectSecretValues -targetProperty "application_id" -newProperty "application_secret" -content $content
    ######    AWS    ######
    $content = InjectSecretValues -targetProperty "api_key" -newProperty "secret_key" -content $content
    ######    GCP    ######
    $content = InjectSecretValues -targetProperty "service_account_id" -newProperty "service_account_credentials" -content $content
    ###### XenServer / vSphere ######
    $content = InjectSecretValues -targetProperty "username" -newProperty "password" -content $content

    ### machine catalog service accounts ###
    $content = InjectSecretValues -targetProperty "domain" -newProperty "service_account" -content $content
    $content = InjectSecretValues -targetProperty "domain" -newProperty "service_account_password" -content $content

    return $content
}
function ExtractAndSaveApplicationIcons {
    param(
        [parameter(Mandatory = $true)]
        [string] $content
    )

    # Check if application icon exists; if not, then exit
    if ($content -notmatch 'citrix_application_icon') {
        return
    }

    $filteredOutput = @()
    $lines = $content -split "`r?`n"
    $iconCounter = 0

    # Create the icons folder
    $iconsFolder = Join-Path -Path $PSScriptRoot -ChildPath "icons"
    if (-not (Test-Path -Path $iconsFolder)) {
        New-Item -ItemType Directory -Path $iconsFolder | Out-Null
    }

    foreach ($line in $lines) {
        if ($line -match 'raw_data\s*=\s*"([^"]+)"') {
            $rawDataValue = $matches[1]
            $iconBytes = [System.Convert]::FromBase64String($rawDataValue)
            $iconFileName = "$iconsFolder\app_icon_$iconCounter.ico"

            try {
                [System.IO.File]::WriteAllBytes($iconFileName, $iconBytes)
            }
            catch {
                Write-Error "Failed to write icon file: $_"
                continue
            }

            # Replace backslashes with double backslashes for terraform
            $iconFileName = $iconFileName -replace '\\', '\\'

            $iconCounter++
            # Replace raw_data value with icon file path using filebase64 to encode a file's content in base64 format
            $line = 'raw_data = filebase64("' + $iconFileName + '")'
        }
        $filteredOutput += $line
    }

    $content = $filteredOutput -join "`n"
    return $content
}

function PostProcessTerraformOutput {

    # Post-process the terraform output
    $content = Get-Content -Path ".\resource.tf" -Raw

    # Remove computed properties
    $content = RemoveComputedProperties -content $content

    # Set dependency relationships
    $content = ReplaceDependencyRelationships -content $content

    # Inject placeholder for sensitive values in tf
    $content = InjectPlaceHolderSensitiveValues -content $content
    
    # Extract and save citrix application icons
    $content = ExtractAndSaveApplicationIcons -content $content

    # Overwrite extracted terraform with processed value
    Set-Content -Path ".\resource.tf" -Value $content

}

function PostProcessProviderConfig {

    # Post-process the provider config output in citrix.tf
    $content = Get-Content -Path ".\citrix.tf" -Raw

    # Uncomment field for client secret in provider config
    $content = $content -replace "# ", ""

    # Overwrite provider config with processed value
    Set-Content -Path ".\citrix.tf" -Value $content
}

if ($DisableSSLValidation -and $PSVersionTable.PSVersion.Major -lt 7) {
    $code = @"
using System.Net;
using System.Security.Cryptography.X509Certificates;
public class TrustAllCertsPolicy : ICertificatePolicy {
    public bool CheckValidationResult(ServicePoint srvPoint, X509Certificate certificate, WebRequest request, int certificateProblem) {
        return true;
    }
}
"@
    Add-Type -TypeDefinition $code -Language CSharp
    [System.Net.ServicePointManager]::CertificatePolicy = New-Object TrustAllCertsPolicy
}

# Initialize script variables
$script:onPremise = ($CustomerId -eq "CitrixOnPremises")
$script:customerId = $CustomerId
$script:clientId = $ClientId
$script:clientSecret = $ClientSecret
$script:domainFqdn = $DomainFqdn
$script:hostname = $Hostname
$script:environment = $Environment
$script:disable_ssl = $DisableSSLValidation
$script:hypervisorResourceMap = @{
    "azure_hypervisor"     = "AzureRM"
    "aws_hypervisor"       = "AWS"
    "gcp_hypervisor"       = "GoogleCloudPlatform"
    "scvmm_hypervisor"     = "SCVMM"
    "xenserver_hypervisor" = "XenServer"
    "vsphere_hypervisor"   = "VCenter"
    "nutanix_hypervisor"   = "Custom"
}
$NUTANIX_PLUGIN_ID = "AcropolisFactory"
$script:applicationFolderPathMap = @{}

# Set environment variables for client secret
$env:CITRIX_CLIENT_SECRET = $ClientSecret

try {
    Get-Site
    Get-RequestBaseUrl
    New-RequiredFiles

    # Get CVAD resources from existing site
    Get-ExistingCVADResources

    # Initialize terraform
    terraform init

    # Import terraform resources into state
    Import-ResourcesToState

    # Export terraform resources
    terraform show >> ".\resource.tf"

    # Post-process citrix.tf output
    PostProcessProviderConfig

    # Post-process terraform output
    PostProcessTerraformOutput

    # Remove temporary TF file
    Remove-Item ".\import.tf"

    # Format terraform files
    terraform fmt
}
finally {
    # Clean up environment variables for client secret
    $env:CITRIX_CLIENT_SECRET = ''
}