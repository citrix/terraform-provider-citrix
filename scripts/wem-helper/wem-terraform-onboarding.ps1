# Copyright © 2025. Citrix Systems, Inc. All Rights Reserved.
<#
Currently this script is still in TechPreview
.SYNOPSIS
    Script to onboard existing WEM Resources to terraform. 
    
.DESCRIPTION 
    The script should be able to collect the list of WEM resources from DDC, import into terraform, and generate the corresponding TF skeletons.

.Parameter CustomerId
    The Citrix Cloud customer ID. Only applicable for Citrix Cloud customers.

.Parameter ClientId
    The Client Id for Citrix DaaS service authentication.
    For Citrix Cloud customers: Use this to specify Cloud API Key Client Id.
    
.Parameter ClientSecret
    The Client Secret for Citrix DaaS service authentication.
    For Citrix Cloud customers: Use this to specify Cloud API Key Client Secret.

.Parameter Hostname
    The Host name / base URL of Citrix DaaS service.
    For Citrix Cloud customers (Optional): Use this to force override the Citrix DaaS service hostname.

.Parameter Environment
    The Citrix Cloud environment of the customer. Only applicable for Citrix Cloud customers. Available options: Production, Staging.

.Parameter ResourceTypes
    Optional list of resource types to onboard. When specified, only those resources will be onboarded, the rest skipped.
    This helps make the onboarding process more manageable by limiting the scope.
    Available resource types include: citrix_wem_configuration_set, citrix_wem_directory_object

.Parameter NamesOrIds
    Optional string array parameter to filter resources by name or ID. Only resources with a Name or ID matching any of these values will be onboarded.
    This allows you to onboard multiple specific resources by name or ID in a single operation.

.Parameter DisableSSLValidation
    Disable SSL validation for this script. Required if DDC does not have a valid SSL certificate.

.Parameter ShowClientSecret
    Specifies whether to display the client secret value in the generated Terraform configuration file; defaults to `$false` for security.
#>  

[CmdletBinding()]
Param (
    [Parameter(Mandatory = $false)]
    [string] $CustomerId,

    [Parameter(Mandatory = $true)]
    [string] $ClientId,
    
    [Parameter(Mandatory = $true)]
    [string] $ClientSecret,

    [Parameter(Mandatory = $false)]
    [string] $Hostname = "api.cloud.com",

    [Parameter(Mandatory = $false)]
    [ValidateSet("Production", "Staging")]
    [string] $Environment = "Production",

    [Parameter(Mandatory = $false)]
    [ValidateSet("citrix_wem_configuration_set", "citrix_wem_directory_object")]
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
    $siteRequest = "https://$script:hostname/cvad/manage/me"
    $response = Start-GetRequest -url $siteRequest
    $script:siteId = $response.Customers[0].Sites[0].Id
}

function Get-RequestBaseUrl {
    $url = "https://$script:hostname/cvad/manage"
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

function Start-GetRequest {
    param(
        [parameter(Mandatory = $true)][string] $url
    )

    $token = Get-AuthToken
    
        $headers = @{
            "Authorization"     = "CwsAuth Bearer=$token"
            "Citrix-CustomerId" = $script:customerId
            "Accept"            = "application/json, text/plain, */*"
        }
        if ($null -ne $script:siteId) {
            $headers["Citrix-InstanceId"] = $script:siteId
        }
    
    
    $contentType = 'application/json'
    $response = Invoke-WebRequestWithRetry -Uri $url -Method 'GET' -Headers $headers -ContentType $contentType
    $jsonObj = ConvertFrom-Json $([String]::new($response.Content))
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
        Set-Content -Path ".\citrix.tf" -Value $config
    

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

# Function to get the URL for WEM objects
function Get-UrlForWemObjects {
    param(
        [parameter(Mandatory = $true)]
        [string] $requestPath
    )

    if ($script:environment -eq "Production") {
        $script:wemHostName = "api.wem.cloud.com"
    }
    elseif ($script:environment -eq "Staging") {
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

    $resourceList = @()
    foreach ($item in $items) {
        if (($NamesOrIds -and $NamesOrIds.Count -gt 0) -and # Filter by NamesOrIds if specified
            (($item.Id -or $item.Name) -and # Item has an Id or Name
            -not (($item.Id -and ($NamesOrIds -contains $item.Id)) -or ($item.Name -and ($NamesOrIds -contains $item.Name))))) { # Item's Id or Name is not in the filter list
            continue # skip
        }

        # Check for id and ignore empty and default values
        if ($item.Id -and $item.Id -ne "0" -and $item.Id -ne "00000000-0000-0000-0000-000000000000") {
            $resourceList += $item.Id
        }

    }
    return $resourceList
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

    $list = Get-ResourceList -requestPath $resourceApi -resourceProviderName $resourceProviderName
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
            "wem_configuration_set" = @{
                "resourceApi"          = "sites"
                "resourceProviderName" = "wem_configuration_set"
            }
            "wem_directory_object"  = @{
                "resourceApi"          = "ad_objects"
                "resourceProviderName" = "wem_directory_object"
            }
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

function RemoveComputedProperties {
    param(
        [parameter(Mandatory = $true)]
        [string] $content
    )

    Write-Verbose "Removing computed properties from terraform output."
    # Define an array of regex patterns to remove computed properties
    $regexPatterns = @(
        "(\s+)id(\s+)= (\S+)"
    )

    # Loop through each regex pattern and replace matches in the content
    foreach ($pattern in $regexPatterns) {
        $content = [regex]::Replace($content, $pattern, "", [System.Text.RegularExpressions.RegexOptions]::Multiline)
    }

    Write-Verbose "Computed properties removed successfully."
    return $content
}


function PostProcessProviderConfig {

    Write-Verbose "Post-processing provider config."
    # Post-process the provider config output in citrix.tf
    $content = Get-Content -Path ".\citrix.tf" -Raw

    # Uncomment field for client secret in provider config
    $content = $content -replace "# ", ""

    # Overwrite provider config with processed value
    Set-Content -Path ".\citrix.tf" -Value $content
}


function FilterOutput {
    param(
        [parameter(Mandatory = $true)]
        [string] $content
    )

    $filteredOutput = @()
    $lines = $content -split "`r?`n"

    foreach ($line in $lines) {    
        $filteredOutput += $line
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
    $content = Get-Content -Path ".\resource.tf" -Raw

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
    $content = Get-Content -Path ".\resource.tf" -Raw

    # Inject placeholder for sensitive values in tf
    $content = FilterOutput -content $content

    # Remove computed properties
    $content = RemoveComputedProperties -content $content

    # Overwrite extracted terraform with processed value
    Set-Content -Path ".\resource.tf" -Value $content

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
$script:customerId = $CustomerId
$script:clientId = $ClientId
$script:clientSecret = $ClientSecret
$script:hostname = $Hostname
$script:environment = $Environment
$script:parentChildMap = @{} # Initialize the parent-child map for hypervisors and image_definitions

$script:TokenExpiryTime = (Get-Date).AddMinutes(-1) # Initialize the expiry time of the refresh token to an earlier time

# Set environment variables for client secret
$env:CITRIX_CLIENT_SECRET = $ClientSecret

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