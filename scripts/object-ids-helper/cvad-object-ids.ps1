
# Copyright © 2025. Citrix Systems, Inc. All Rights Reserved.
<#
Currently this script is still in TechPreview
.SYNOPSIS
    Script to fetch the object ids for the CVAD Resources for On-Prem and Cloud Resources.
.DESCRIPTION 
    The script should be able to collect the list of resources from DDC, fetch the object ids, and store those in a JSON file.

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
        # Return the token if its still valid
        if ($null -eq $script:Token) {
            Write-Verbose "Requesting new token."
        }
        elseif ((Get-Date) -lt $script:TokenExpiryTime) {
            Write-Verbose "Refresh token is still valid. Returning the existing token."
            return $script:Token
        }
        else {
            Write-Verbose "Refresh token Expired. Requesting new token."
        }

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
        
        # Save the new token and calculate the expiry time of the refresh token
        $script:Token = $jsonObj.access_token
        $script:TokenExpiryTime = (Get-Date).AddSeconds([int]($jsonObj.expires_in * 0.9)) # Calculate the expiry time of the refresh token with buffer
        return $script:Token
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
            "Accept"            = "application/json, text/plain, */*"
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


# Function to get the URL for WEM objects
function Get-UrlForWemObjects {
    param(
        [parameter(Mandatory = $true)]
        [string] $requestPath
    )

    if ($script:environment -eq "Production") {
        $script:wemHostName = "api.wem.cloud.com"
    }
    else {
        $script:wemHostName = "api.wem.cloudburrito.com"
    }

    if ($requestPath -eq "sites") {
        return "https://$script:wemHostName/services/wem/sites?includeHidden=true&includeUnboundAgentsSite=true"
    }
    else {
        return "https://$script:wemHostName/services/wem/machines"
    }
}

# Function to find Catalog AD objects
function Find-CatalogADObjects {
    param(
        [parameter(Mandatory = $true)]
        [array] $items
    )

    $catalogItems = @()
    foreach ($item in $items) {
        if ($item.type -eq "Catalog") {
            $catalogItems += $item
        }
    }
    return $catalogItems
}

# Function to get list of resources for a given resource provider
function Get-ResourceList {
    param(
        [parameter(Mandatory = $true)]
        [string] $requestPath,

        [parameter(Mandatory = $true)]
        [string] $resourceProviderName
    )

    $object_list   = @()

    $url = "$script:urlBase/$requestPath"

    # Update url for WEM Objects
    if ($resourceProviderName -in "wem_configuration_set", "wem_directory_object") {
        $url = Get-UrlForWemObjects -requestPath $requestPath
    }
    
    # Check if the resource provider is supported in the current environment (eg. WEM is not supported for most environments)
    try {
        $response = Start-GetRequest -url $url
    }
    catch {
        # Ignore 503 errors for WEM objects
        if (-not($_.Exception.Response.StatusCode -eq 503)) {
            Write-Error "Failed to get $resourceProviderName. Error: $($_.Exception.Message)" -ErrorAction Continue
        }
        return @()
    }
    
    $items = $response.Items

    # WEM supports AD object type 'Catalog'. Filter out other object types
    if ($resourceProviderName -eq "wem_directory_object" -and $items.Count -gt 0) {
        $items = Find-CatalogADObjects -items $items
    }

    foreach ($item in $items) {

        # Handle special case for Policies
        if ($resourceProviderName -eq "policy_set" -and $item.policySetGuid) {
            $newObjList = [PSCustomObject]@{ Name = $item.Name; Id = $item.policySetGuid }
            $object_list += $newObjList
        }

        # Check for ServiceAccountUid for Service Accounts
        elseif ($resourceProviderName -eq "service_account" -and $item.ServiceAccountUid){
            $newObjList = [PSCustomObject]@{ Name = $item.DisplayName; Id = $item.ServiceAccountUid }
            $object_list += $newObjList
        }

        elseif ($resourceProviderName -eq "admin_folder" -and $item.Id -eq "0") {
            # Skip the root folder
            continue
        }

        # Check for Security Identifier for Admin Users
        elseif ($resourceProviderName -eq "admin_user" -and $item.User -and $item.User.Sid){
            $newObjList = [PSCustomObject]@{ Name = $item.User.DisplayName; Id = $item.User.Sid }
            $object_list += $newObjList
        }

        # Store icons as files
        elseif ($requestPath -like "Icons*") {
            if ($item.Id -eq "0" -or $item.Id -eq "1") {
                continue
            }
            $iconsFolder = Join-Path -Path $PSScriptRoot -ChildPath "icons"
            # Create the icons folder
            if (-not (Test-Path -Path $iconsFolder)) {
                New-Item -ItemType Directory -Path $iconsFolder | Out-Null
            }
            
            $iconBytes = [System.Convert]::FromBase64String($item.RawData)
            $iconFileName = "$iconsFolder\app_icon_$($item.Id).ico"

            try {
                [System.IO.File]::WriteAllBytes($iconFileName, $iconBytes)
            }
            catch {
                Write-Error "Failed to write icon file: $_"
                continue
            }
        }

        else {
            $newObjList = [PSCustomObject]@{ Name = $item.Name; Id = $item.Id }
            $object_list += $newObjList
        }
    }
    return $object_list
}

# List all CVAD objects from existing site
function Get-ObjectIdsFromCVADResources {
   
    Write-Verbose "Get list of all existing CVAD resources from the site."
    $resources = @{
        "zone"                 = @{
            "resourceApi"          = "zones"
            "resourceProviderName" = "zone"
        }
        "hypervisor"     = @{
            "resourceApi"          = "hypervisors"
            "resourceProviderName" = "hypervisor"
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
            "resourceApi"          = "gpo/policySets"
            "resourceProviderName" = "policy_set"
        }
        "application"          = @{
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
        "application_icon"     = @{
            "resourceApi"          = "Icons?builtIn=false"
            "resourceProviderName" = "application_icon"
        }
        "service_account" = @{
            "resourceApi"          = "Identity/ServiceAccounts"
            "resourceProviderName" = "service_account"
        }
        "image_definition" = @{
            "resourceApi"          = "ImageDefinitions"
            "resourceProviderName" = "image_definition"
        }
        "storefront_server" = @{
            "resourceApi"          = "StoreFrontServers"
            "resourceProviderName" = "storefront_server"
        }
        "tag" = @{
            "resourceApi"          = "Tags"
            "resourceProviderName" = "tag"
        }
    }

    $script:myObject = [PSCustomObject]@{
        cvad_objects = @()
    }

    # Add WEM resources for cloud customer environment
    if (-not($script:onPremise)) {
        $wemResources = @{
            "wem_configuration_set" = @{
                "resourceApi"          = "sites"
                "resourceProviderName" = "wem_configuration_set"
            }
            "wem_directory_object"  = @{
                "resourceApi"          = "ad_objects"
                "resourceProviderName" = "wem_directory_object"
            }
        }
        $resources += $wemResources
    }else {
    # If On-Prem add admin resource
        $resources.Add("admin_user", @{
            "resourceApi"          = "Admin/Administrators"
            "resourceProviderName" = "admin_user"
        })
    }

    foreach ($resource in $resources.Keys) {
        $api = $resources[$resource].resourceApi
        $resourceProviderName = $resources[$resource].resourceProviderName

        $objects = Get-ResourceList -requestPath $api -resourceProviderName $resourceProviderName
        if ($objects.Count -eq 0 -or $resource -eq "application_icon") {
            continue
        }

        $myObject.cvad_objects += [PSCustomObject]@{
            object_type = $resource
            object_list   = $objects
        }

    }
    
    # Save the $myObject to a JSON file
    $json = $myObject | ConvertTo-Json -Depth 10
    $json | Out-File -FilePath ".\object_ids.json" -Encoding utf8 -Force

    Write-Verbose "Successfully retrieved all the CVAD Object IDs from the site."
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
$script:applicationFolderPathMap = @{}
$script:parentChildMap = @{} # Initialize the parent-child map for hypervisors and image_definitions

$script:TokenExpiryTime = (Get-Date).AddMinutes(-1) # Initialize the expiry time of the refresh token to an earlier time

# Set environment variables for client secret
$env:CITRIX_CLIENT_SECRET = $ClientSecret

try {
    Get-Site
    Get-RequestBaseUrl

    # Get Object IDs from CVAD resources
    # This will create a JSON file with the object IDs for all resources in the site
    Get-ObjectIdsFromCVADResources
    
}
finally {
    # Clean up environment variables for client secret
    $env:CITRIX_CLIENT_SECRET = ''
}