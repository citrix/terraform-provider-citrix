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
    The Citrix Cloud environment of the customer. Only applicable for Citrix Cloud customers. Available options: Production, Japan, Gov

.Parameter ResourceTypes
    Optional list of resource types to onboard. When specified, only those resources will be onboarded, the rest skipped.
    This helps make the onboarding process more manageable by limiting the scope.
    By default if (-NoDependencyRelationship is not specified), will resolve all dependency relationships between resources as long as the dependent resource is included.
    Available resource types include: citrix_admin_folder, citrix_admin_role, citrix_admin_scope, citrix_admin_user, citrix_application, citrix_application_group, citrix_application_icon, citrix_aws_hypervisor, citrix_azure_hypervisor, citrix_delivery_group, citrix_gcp_hypervisor, citrix_image_definition, citrix_machine_catalog, citrix_nutanix_hypervisor, citrix_openshift_hypervisor, citrix_policy_set, citrix_scvmm_hypervisor, citrix_service_account, citrix_storefront_server, citrix_tag, citrix_vsphere_hypervisor, citrix_xenserver_hypervisor, citrix_zone
    citrix_<hypervisorType>_resource_pools are included with the citrix_<hypervisorType>_hypervisor resource.
    citrix_image_version is included with the citrix_image_definition resource.

.Parameter NamesOrIds
    Optional string array parameter to filter resources by name or ID. Only resources with a Name or ID matching any of these values will be onboarded.
    This allows you to onboard multiple specific resources by name or ID in a single operation.
    By default if (-NoDependencyRelationship is not specified), will resolve all dependency relationships between resources as long as the dependent resource is included.

.Parameter NoDependencyRelationship
    Do not create dependency relationships between resources by replacing resource references with the resource IDs.

.Parameter DisableSSLValidation
    Disable SSL validation for this script. Required if DDC does not have a valid SSL certificate.

.Parameter ShowClientSecret
    Specifies whether to display the client secret value in the generated Terraform configuration file; defaults to `$false` for security.
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
    [ValidateSet("Production", "Staging", "Japan", "JapanStaging", "Gov", "GovStaging")]
    [string] $Environment = "Production",

    [Parameter(Mandatory = $false)]
    [ValidateSet("citrix_admin_folder", "citrix_admin_role", "citrix_admin_scope", "citrix_admin_user", "citrix_application", "citrix_application_group", "citrix_application_icon", "citrix_aws_hypervisor", "citrix_azure_hypervisor", "citrix_delivery_group", "citrix_gcp_hypervisor", "citrix_image_definition", "citrix_machine_catalog", "citrix_nutanix_hypervisor", "citrix_openshift_hypervisor", "citrix_policy_set", "citrix_scvmm_hypervisor", "citrix_service_account", "citrix_storefront_server", "citrix_tag", "citrix_vsphere_hypervisor", "citrix_xenserver_hypervisor", "citrix_zone")]
    [string[]] $ResourceTypes,

    [Parameter(Mandatory = $false)]
    [string[]] $NamesOrIds,

    [Parameter(Mandatory = $false)]
    [switch] $NoDependencyRelationship,

    [Parameter(Mandatory = $false)]
    [switch] $DisableSSLValidation,

    [Parameter(Mandatory=$false)]
    [switch] $ShowClientSecret
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
        [string]$ContentType = 'application/json; charset=utf-8',

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
        $jsonObj = ConvertFrom-Json $response.Content
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
        elseif ($script:environment -eq "Japan") {
            $url = "https://api.citrixcloud.jp/cctrustoauth2/$script:customerId/tokens/clients"
        }
        elseif ($script:environment -eq "JapanStaging") {
            $url = "https://api.citrixcloud-test.jp/cctrustoauth2/$script:customerId/tokens/clients"
        }
        elseif ($script:environment -eq "Gov") {
            $url = "https://api.citrixcloud.us/cctrustoauth2/$script:customerId/tokens/clients"
        }
        elseif ($script:environment -eq "GovStaging") {
            $url = "https://api.citrixcloud-test.us/cctrustoauth2/$script:customerId/tokens/clients"
        }

        $body = @{
            grant_type    = 'client_credentials'
            client_id     = $script:clientId
            client_secret = $script:clientSecret
        }
        $contentType = 'application/x-www-form-urlencoded'
        $response = Invoke-WebRequestWithRetry -Uri $url -Method 'POST' -body $body -ContentType $contentType
        $jsonObj = ConvertFrom-Json $response.Content

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
    
    $response = Invoke-WebRequestWithRetry -Uri $url -Method 'GET' -Headers $headers -ContentType $contentType
    $responseJsonString = [regex]::Replace($response.Content, '\\x([0-9A-Fa-f]{2})', { param($hex) '\u{0}' -f $hex.Groups[1].Value.PadLeft(4,'0') })
    $jsonObj = ConvertFrom-Json $responseJsonString
    return $jsonObj
}

function New-RequiredFiles {

    Write-Verbose "Creating required files for terraform."

    # Determine the client secret value based on the ShowClientSecret flag
    $secretValue = if ($ShowClientSecret) { $script:clientSecret } else { "<Input client secret value>" }

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
        # client_secret               = "$secretValue"
        disable_ssl_verification    = $disable_ssl_verification
    }
}
"@
        Set-Content -Path ".\citrix.tf" -Value $config -Encoding utf8
    }
    else {
        $config = @"
provider "citrix" {
    cvad_config = {
        customer_id                 = "$script:customerId"
        client_id                   = "$script:clientId"
        # client_secret               = "$secretValue"
        hostname                    = "$script:hostname"
        environment                 = "$script:environment"
    }
}
"@
        Set-Content -Path ".\citrix.tf" -Value $config -Encoding utf8
    }

    # Create temporary import.tf for terraform import
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

    Write-Verbose "Required files created successfully."

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



    $resourceList = @()
    $pathMap = @{}
    foreach ($item in $items) {
        if (($NamesOrIds -and $NamesOrIds.Count -gt 0) -and # Filter by NamesOrIds if specified
            (($item.Id -or $item.Name) -and # Item has an Id or Name
            -not (($item.Id -and ($NamesOrIds -contains $item.Id)) -or ($item.Name -and ($NamesOrIds -contains $item.Name))))) { # Item's Id or Name is not in the filter list
            continue # skip
        }

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

        # Handle special case for Policy Sets
        if($requestPath -eq "gpo/policySets" -and $item.name -eq "DefaultSitePolicies")
        {
            # Skip processing for the default site policies
            continue
        }

        #Handle special case for Built-in Admin Roles
        if($requestPath -eq "Admin/Roles"){
            if($item.IsBuiltIn){
                continue;
            }
        }

        # Handle special case for Built-in Admin Scopes
        if ($requestPath -eq "Admin/Scopes") {
            if ($item.IsBuiltIn) {
                continue
            }
        }

        # Handle special case for Icons
        if ($requestPath -like "Icons*") {
            if ($item.Id -eq "0" -or $item.Id -eq "1") {
                continue
            }
        }

        # Handle special case for Policies
        if ($item.policySetGuid -and $item.policySetType -like "*Policies*") {
            $resourceList += $item.policySetGuid
        }

        # Check for id and ignore empty and default values
        if ($item.Id -and $item.Id -ne "0" -and $item.Id -ne "00000000-0000-0000-0000-000000000000") {
            $resourceList += $item.Id
        }

        # Check for ServiceAccountUid for Service Accounts
        if ($resourceProviderName -eq "service_account" -and $item.ServiceAccountUid){
            $resourceList += $item.ServiceAccountUid
        }

        # Check for Security Identifier for Admin Users
        if ($resourceProviderName -eq "admin_user" -and $item.User -and $item.User.Sid){
            $resourceList += $item.User.Sid
        }
 
        # Create a path map for ApplicationFolder paths
        if ($requestPath -eq "AdminFolders") {
            $pathMap[$item.Id] = $item.Path
        }

        # Store icons as files
        if ($requestPath -like "Icons*") {
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
            if (-not $script:parentChildMap.ContainsKey($parentId)) {
                # Initialize as a new list if not already present
                $script:parentChildMap[$parentId] = [System.Collections.Generic.List[string]]::new()
            }
            $script:parentChildMap[$parentId].Add($id)
        }
        else {
            $resourceName = "$($resourceProviderName)_$($index)"
            $resourceMapKey = $id
        }
        
        if ($resourceApi -eq "AdminFolders" -and $pathMap.Count -gt 0) {
            $script:applicationFolderPathMap[$pathMap.$id.TrimEnd('\')] = $resourceName
        }
        
        $resourceMap[$resourceMapKey] = $resourceName
        $resourceContent = "resource `"citrix_$resourceProviderName`" `"$resourceName`" {}`n"
        Add-Content -Path ".\import.tf" -Value $resourceContent
        $index += 1
    }

    return $resourceMap
}

# List all CVAD objects from existing site
function Get-ExistingCVADResources([string[]]$filter = $null) {
   
    # IMPORTANT: When adding a new resource type, the following places must be updated:
    # 1. Add the resource to the ValidateSet parameter for ResourceTypes at the top of this script
    # 2. Update the .PARAMETER ResourceTypes documentation in the script header comments
    # 3. Update the README.md in this directory to include the new resource type
    # 4. Add the resource to the $resources hashtable below without the `citrix_` prefix
    
    Write-Verbose "Get list of all existing CVAD resources from the site."
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
        "openshift_hypervisor" = @{
            "resourceApi"          = "hypervisors"
            "resourceProviderName" = "openshift_hypervisor"
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

    # If On-Prem add admin resource
    if (($script:onPremise)) {
        $resources.Add("admin_user", @{
            "resourceApi"          = "Admin/Administrators"
            "resourceProviderName" = "admin_user"
        })
    }

    $script:cvadResourcesMap = @{}

    foreach ($resource in $resources.Keys) {
        # Skip resources not in ResourceTypes if ResourceTypes parameter is specified
        if ($filter -and $filter.Count -gt 0 -and -not $filter.Contains($resource)) {
            Write-Verbose "Skipping resource type: $resource as it's not in the specified ResourceTypes parameter"
            continue
        }

        $api = $resources[$resource].resourceApi
        $resourceProviderName = $resources[$resource].resourceProviderName
        $script:cvadResourcesMap[$resource] = Get-ImportMap -resourceApi $api -resourceProviderName $resourceProviderName
        
        # Create resource pool map for each hypervisor if exists
        if ($resource -like "*hypervisor") {
            $index = 0
            foreach ($id in $script:cvadResourcesMap[$resource].Keys) {
                $resourcePoolAPI = "hypervisors/$($id)/resourcePools"
                $script:cvadResourcesMap["$($resource)_resource_pool"] += Get-ImportMap -resourceApi $resourcePoolAPI -resourceProviderName "$($resource)_resource_pool" -parentId $id -parentIndex $index
                $index += 1
            }
        }
        # Create image_version map for all image_definitions
        if ($resource -like "image_definition") {
            $index = 0
            foreach ($id in $script:cvadResourcesMap[$resource].Keys) {
                $resourcePoolAPI = "ImageDefinitions/$($id)/ImageVersions"
                $script:cvadResourcesMap["image_version"] += Get-ImportMap -resourceApi $resourcePoolAPI -resourceProviderName "image_version" -parentId $id -parentIndex $index
                $index += 1
            }
        }
    }
    Write-Verbose "Successfully retrieved all CVAD resources from the site."
}

# Function to import terraform resources into state
function Import-ResourcesToState {
    Write-Verbose "Importing terraform resources into state."
    foreach ($resource in  $script:cvadResourcesMap.Keys) {
        foreach ($id in  $script:cvadResourcesMap[$resource].Keys) {
            terraform import "citrix_$($resource).$($script:cvadResourcesMap[$resource][$id])" "$id" 
        }
    }
    Write-Verbose "Successfully imported resources into state."
}

function PostProcessProviderConfig {

    Write-Verbose "Post-processing provider config."
    # Post-process the provider config output in citrix.tf
    $content = Get-Content -Path ".\citrix.tf" -Raw -Encoding utf8

    # Uncomment field for client secret in provider config
    $content = $content -replace "# ", ""

    # Overwrite provider config with processed value
    Set-Content -Path ".\citrix.tf" -Value $content -Encoding utf8
}

function RemoveComputedPropertiesForZone {
    param(
        [parameter(Mandatory = $true)]
        [string] $content
    )

    if ($script:onPremise) {
        Write-Verbose "Removing computed properties for zone resource in on-premises."
        # Remove resource_location_id property from each zone resource for on-premises 
        $resourceLocationIdRegex = "(\s+)resource_location_id(\s+)= (\S+)"
        $content = $content -replace $resourceLocationIdRegex, ""
    }
    else {
        Write-Verbose "Removing computed properties for zone resource in cloud."
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

    Write-Verbose "Removing computed properties from terraform output."
    # Define an array of regex patterns to remove computed properties
    $regexPatterns = @(
        "(\s+)id(\s+)= (\S+)",
        '(\s+)path\s*=\s*"(.*?)"',
        "(\s+)assigned(\s+)= (\S+)",
        "(\s+)is_built_in(\s+)= (\S+)",
        "(\s+)built_in_scopes\s*=\s*\[[\s\S]*?\]",
        "(\s+)inherited_scopes\s*=\s*\[[\s\S]*?\]",
        "(\s+)total_application_groups(\s+)= (\S+)",
        "(\s+)total_applications(\s+)= (\S+)",
        "(\s+)total_delivery_groups(\s+)= (\S+)",
        "(\s+)total_machine_catalogs(\s+)= (\S+)",
        "(\s+)total_machines(\s+)= (\S+)",
        "(\s+)is_all_scope(\s+)= (\S+)",
        "(\s+)latest_version(\s+)= (\S+)",
        "(\s+)version_number(\s+)= (\S+)",
        "(\s+)associated_delivery_group_count(\s+)= (\S+)",
        "(\s+)associated_machine_catalog_count(\s+)= (\S+)",
        "(\s+)associated_machine_count(\s+)= (\S+)",
        "(\s+)associated_application_group_count(\s+)= (\S+)",
        "(\s+)associated_application_count(\s+)= (\S+)",
        "(\s+)tenant_id(\s+)= (\S+)",
        "(\s+)tenant_name(\s+)= (\S+)",
        "(\s+)tenants\s*=\s*\[[\s\S]*?\]"
    )

    # Identify the delivery_groups_priority block
    $deliveryGroupsPriorityPattern = "(\s*)delivery_groups_priority\s*=\s*\[[\s\S]*?\]"
    $deliveryGroupsPriorityMatches = [regex]::Matches($content, $deliveryGroupsPriorityPattern, [System.Text.RegularExpressions.RegexOptions]::Singleline)

    # Extract the delivery_groups_priority block and replace it with a unique placeholder
    $index = 0
    foreach ($match in $deliveryGroupsPriorityMatches) {
        $deliveryGroupsPriorityBlock = $match.Value
        $content = $content -replace [regex]::Escape($deliveryGroupsPriorityBlock), "PLACEHOLDER_DELIVERY_GROUPS_PRIORITY_$index"
        $index++
    }

    # Loop through each regex pattern and replace matches in the content
    foreach ($pattern in $regexPatterns) {
        $content = [regex]::Replace($content, $pattern, "", [System.Text.RegularExpressions.RegexOptions]::Multiline)
    }

    # Restore the delivery_groups_priority block using unique placeholders
    $index = 0
    foreach ($match in $deliveryGroupsPriorityMatches) {
        $content = $content -replace "PLACEHOLDER_DELIVERY_GROUPS_PRIORITY_$index", $match.Value
        $index++
    }

    # Remove contents for zone resource
    $content = RemoveComputedPropertiesForZone -content $content

    Write-Verbose "Computed properties removed successfully."
    return $content
}

function ReplaceDependencyRelationships {
    param(
        [parameter(Mandatory = $true)]
        [string] $content
    )

    if ($script:NoDependencyRelationship) {
        return $content
    }

    Write-Verbose "Creating dependency relationships between resources."
    # Create dependency relationships between resources with id references
    foreach ($resource in $script:cvadResourcesMap.Keys) {

        foreach ($id in $script:cvadResourcesMap[$resource].Keys) {
            if($resource -like "*_resource_pool" -or $resource -like "image_version") {
                $idArray = $id -split ","
                if($idArray.Count -gt 1) {
                    $resource_id = $idArray[1]
                    Write-Verbose "Replacing ID: $resource_id with citrix_$($resource).$($script:cvadResourcesMap[$resource][$id]).id"
                    $content = $content -replace "`"$resource_id`"", "citrix_$($resource).$($script:cvadResourcesMap[$resource][$id]).id"
                }
            }else{
                Write-Verbose "Replacing ID: $id with citrix_$($resource).$($script:cvadResourcesMap[$resource][$id]).id"
                $content = $content -replace "`"$id`"", "citrix_$($resource).$($script:cvadResourcesMap[$resource][$id]).id"
            }
        }
    }

    # Create dependency relationships between resources with path references
    foreach ( $applicationFolderPath in $script:applicationFolderPathMap.Keys) {
        $path = $applicationFolderPath.replace("\", "\\\\")
        $content = $content -replace "(\s(parent_path|application_folder_path|application_group_folder_path|delivery_group_folder_path|machine_catalog_folder_path)\s+= )(`"$path`")", "`${1}citrix_admin_folder.$($script:applicationFolderPathMap[$applicationFolderPath]).path"
    }

    return $content
}

function InjectPlaceHolderSensitiveValues {
    param(
        [parameter(Mandatory = $true)]
        [string] $content
    )

    $filteredOutput = @()
    $lines = $content -split "`r?`n"
    $iconsFolder = Join-Path -Path $PSScriptRoot -ChildPath "icons"

    $previousLine = ""
    foreach ($line in $lines) {
        if ($line -match '^\s*resource\s*"citrix_image_version"\s*') {
            $insideCitrixImageVersion = $true
        } elseif ($insideCitrixImageVersion -and $line -eq "}") {
            $insideCitrixImageVersion = $false
        }

        if ($line -match '^\s*resource\s*"citrix_image_definition"\s*') {
            $insideCitrixImageDefinition = $true
        } elseif ($insideCitrixImageDefinition -and $line -eq "}") {
            $insideCitrixImageDefinition = $false
        }

        # Skip os_type and session_support if inside citrix_image_version
        if ($insideCitrixImageVersion -and ($line -match '^\s*os_type\s*=' -or $line -match '^\s*session_support\s*=')) {
            Write-Verbose "Removing os_type or session_support from citrix_image_version."
            continue
        }

        if ($insideCitrixImageDefinition -and $line -match '^\s*hypervisor\s*=\s*"(.*)"') {
            $hypervisorId = $matches[1]
            $resourcePoolId = if ($script:parentChildMap.ContainsKey($hypervisorId) -and $script:parentChildMap[$hypervisorId].Count -gt 0) {
                $script:parentChildMap[$hypervisorId][0]
            } else {
                "<Enter hypervisor pool id>"
            }
            $filteredOutput += $line
            $filteredOutput += "hypervisor_resource_pool  = `"$resourcePoolId`""
        } elseif ($line -match 'raw_data' -and $previousLine -match 'id\s*=\s*"(.*)"') {
            $iconId = $matches[1]
            $iconFileName = "$iconsFolder\app_icon_$iconId.ico"

            # Replace backslashes with double backslashes for terraform
            $iconFileName = $iconFileName -replace '\\', '\\'

            # Replace raw_data value with icon file path using filebase64 to encode a file's content in base64 format
            $line = 'raw_data = filebase64("' + $iconFileName + '")'
            $filteredOutput += $line
        }elseif ($line -match '.*=\s*null') {
            Write-Verbose "Ignoring lines with null values."
            continue
        }
        elseif ($line -match '^\s*[^=]+\s*=\s*""') {
            Write-Verbose "Ignoring lines with empty strings."
            continue
        }elseif($previousLine -match 'citrix_service_account' -and $line -match 'account_id'){
            $filteredOutput += $line
            $filteredOutput += 'account_secret = "<input application_secret value>"'
            $filteredOutput += 'account_secret_format = "PlainText"'
        }
        elseif($line -match 'citrix_openshift_hypervisor' -and $previousLine -match 'openshift_hypervisor.openshift'){
            $filteredOutput += $line
            $filteredOutput += 'service_account_token = "<Service_Access_Token_In_Plaintext>"'
        }
        elseif ($line -match "application_id") {
            $filteredOutput += $line
            $filteredOutput += 'application_secret = "<input application_secret value>"'
        }
        elseif ($line -match "api_key") {
            $filteredOutput += $line
            $filteredOutput += 'secret_key = "<input secret_key value>"'
        }
        elseif ($line -match "username") {
            $filteredOutput += $line
            $filteredOutput += 'password = "<input password value>"'
            $filteredOutput += 'password_format = "PlainText"'
        }
        elseif (($previousLine -match "^\s*domain\s*=\s*.*$" -or $previousLine -match "^\s*domain_ou\s*=") -and $line -match "^\s*}\s*$") {
            $filteredOutput += 'service_account = "<input service_account value>"'
            $filteredOutput += 'service_account_password = "<input service_account_password value>"'
            $filteredOutput += $line
        }
        else {
            $filteredOutput += $line
        }
        $previousLine = $line
    }
    $content = $filteredOutput -join "`n"
    return $content
}

function OrganizeTerraformResources {
    param(
        [parameter(Mandatory = $true)]
        [string] $content
    )

    Write-Verbose "Organizing terraform resources into separate files."
    # Post-process the terraform output
    $content = Get-Content -Path ".\resource.tf" -Raw -Encoding utf8

    # Regular expression to match resource blocks starting with # and ending with an empty line
    $resourcePattern = '(#\s*(\w+)\.\w+:\s*.*?)(\n\s*\n|\s*$)'

    # Find all resource blocks
    $resources = [regex]::Matches($content, $resourcePattern, [System.Text.RegularExpressions.RegexOptions]::Singleline)

    # Create a new .tf file for each resource type in its respective folder
    foreach ($resource in $resources) {
        $resourceBlock = $resource.Groups[1].Value
        $resourceType = $resource.Groups[2].Value
        $filename = "$resourceType.tf"
    
        # Append the resource block to the file
        Add-Content -Path $filename -Value $resourceBlock
        Add-Content -Path $filename -Value "`n"  # Add a newline for separation
    }

    Write-Verbose "Resource files created successfully."
}

function PostProcessTerraformOutput {

    # Post-process the terraform output
    $content = Get-Content -Path ".\resource.tf" -Raw -Encoding utf8

    # Inject placeholder for sensitive values in tf
    $content = InjectPlaceHolderSensitiveValues -content $content

    # Set dependency relationships
    $content = ReplaceDependencyRelationships -content $content

    # Remove computed properties
    $content = RemoveComputedProperties -content $content
    
    # Overwrite extracted terraform with processed value
    Set-Content -Path ".\resource.tf" -Value $content -Encoding utf8

    # Organize terraform resources into separate files
    OrganizeTerraformResources -content $content
}

### End Helper Functions ###
$ErrorActionPreference = "Stop"

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
    "openshift_hypervisor" = "OpenShift"
}
$NUTANIX_PLUGIN_ID = "AcropolisFactory"
$script:applicationFolderPathMap = @{}
$script:parentChildMap = @{} # Initialize the parent-child map for hypervisors and image_definitions

$script:TokenExpiryTime = (Get-Date).AddMinutes(-1) # Initialize the expiry time of the refresh token to an earlier time

# Set environment variables for client secret
$env:CITRIX_CLIENT_SECRET = $ClientSecret

# Strip citrix_ prefix from ResourceTypes if present
if ($ResourceTypes) {
    $resouceTypesWithoutCitrixPrefix = $ResourceTypes | ForEach-Object { $_ -replace '^citrix_', '' }
} else {
    $resouceTypesWithoutCitrixPrefix = $null # import all resources
}

try {
    Get-Site
    Get-RequestBaseUrl
    New-RequiredFiles

    # Get CVAD resources from existing site
    Get-ExistingCVADResources $resouceTypesWithoutCitrixPrefix

    # Initialize terraform
    terraform init

    # Import terraform resources into state
    Import-ResourcesToState

    # Save the current console output encoding
    $prev = [Console]::OutputEncoding

    # Set the console output encoding to UTF-8
    [Console]::OutputEncoding = [System.Text.UTF8Encoding]::new()

    # Run terraform show command and output to resource.tf file. Add -no-color to disable output with coloring
    terraform show -no-color >> ".\resource.tf"

    # Restore the previous console output encoding
    [Console]::OutputEncoding = $prev

    # Post-process citrix.tf output
    PostProcessProviderConfig

    # Post-process terraform output
    PostProcessTerraformOutput

    # Remove temporary files
    Remove-Item ".\import.tf"
    Remove-Item ".\resource.tf"

    # Format terraform files
    terraform fmt
}
finally {
    # Clean up environment variables for client secret
    $env:CITRIX_CLIENT_SECRET = ''
}