# Copyright © 2026. Citrix Systems, Inc. All Rights Reserved.
<#
Currently this script is still in TechPreview
.SYNOPSIS
    Script to onboard an existing Citrix site to Terraform.

.DESCRIPTION
    Collects the list of resources from DDC, imports them into Terraform state, and generates Terraform configuration skeletons.

    All generated output (configuration, provider files, Terraform state, and application icons) is written to a subfolder next
    to this script (default "citrix-site", configurable with -OutputFolder) so it stays self-contained and can serve as the
    customer's ongoing Terraform project. Run subsequent Terraform commands (terraform plan/apply) from inside that subfolder.

    The script is idempotent and can be re-run against an existing Terraform project directory. On re-runs, resources already
    present in the Terraform state are skipped. Newly discovered resources (added to the site since the last run, or previously
    excluded by -ResourceTypes / -NamesOrIds filters) are imported and appended to the existing .tf files.

    Do not run the script while resources are being created, modified, or deleted in the site by another process. The script
    enumerates the site and imports what it finds; concurrent changes can cause resources to be missed or imported in an
    inconsistent state. Run it against a stable site.

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
    Available resource types include: citrix_admin_folder, citrix_admin_role, citrix_admin_scope, citrix_admin_user, citrix_application, citrix_application_group, citrix_application_icon, citrix_aws_hypervisor, citrix_azure_hypervisor, citrix_cloud_admin_user, citrix_cloud_google_identity_provider, citrix_cloud_okta_identity_provider, citrix_cloud_resource_location, citrix_cloud_saml_identity_provider, citrix_delivery_group, citrix_gcp_hypervisor, citrix_image_definition, citrix_machine_catalog, citrix_nutanix_hypervisor, citrix_openshift_hypervisor, citrix_policy_set_v2, citrix_quickdeploy_catalog, citrix_scvmm_hypervisor, citrix_service_account, citrix_storefront_server, citrix_tag, citrix_vsphere_hypervisor, citrix_xenserver_hypervisor, citrix_zone
    citrix_<hypervisorType>_resource_pools are included with the citrix_<hypervisorType>_hypervisor resource.
    citrix_image_version is included with the citrix_image_definition resource.
    citrix_policy, citrix_policy_priority, citrix_policy_setting, and policy filter resources (citrix_access_control_policy_filter, citrix_branch_repeater_policy_filter, citrix_client_ip_policy_filter, citrix_client_name_policy_filter, citrix_client_platform_policy_filter, citrix_delivery_group_policy_filter, citrix_delivery_group_type_policy_filter, citrix_ou_policy_filter, citrix_tag_policy_filter, citrix_user_policy_filter) are included with the citrix_policy_set_v2 resource.
    citrix_cloud_admin_user, citrix_cloud_resource_location, citrix_cloud_saml_identity_provider, citrix_cloud_google_identity_provider, and citrix_cloud_okta_identity_provider are cloud-only resources.
    Note: Quick Deploy is cloud-only. citrix_quickdeploy_catalog is imported via terraform import. Quick Deploy template images are emitted as data sources (referenced by onboarded catalogs) rather than imported as resources. Quick Deploy on-premises network connections are not supported by this script.

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

.Parameter QuickDeployHostname
    Optional override for the Quick Deploy catalog service base (host plus the `/catalogservice` path segment, e.g. `api.dev.cloud.com/catalogservice`).
    For standard Cloud environments the host is derived from -Environment automatically; only set this for non-standard / internal environments.
    Mirrors the provider's `CITRIX_QUICK_DEPLOY_HOST_NAME` override; if this parameter is omitted, the `CITRIX_QUICK_DEPLOY_HOST_NAME` environment variable is used when present.

.Parameter OutputFolder
    Subfolder (relative to this script) where the generated Terraform project is created and where all Terraform commands should be run.
    Defaults to "citrix-site". Re-run the script with the same value to continue onboarding into an existing project.
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
    [ValidateSet("citrix_admin_folder", "citrix_admin_role", "citrix_admin_scope", "citrix_admin_user", "citrix_application", "citrix_application_group", "citrix_application_icon", "citrix_aws_hypervisor", "citrix_azure_hypervisor", "citrix_cloud_admin_user", "citrix_cloud_google_identity_provider", "citrix_cloud_okta_identity_provider", "citrix_cloud_resource_location", "citrix_cloud_saml_identity_provider", "citrix_delivery_group", "citrix_gcp_hypervisor", "citrix_image_definition", "citrix_machine_catalog", "citrix_nutanix_hypervisor", "citrix_openshift_hypervisor", "citrix_policy_set_v2", "citrix_quickdeploy_catalog", "citrix_scvmm_hypervisor", "citrix_service_account", "citrix_storefront_server", "citrix_tag", "citrix_vsphere_hypervisor", "citrix_xenserver_hypervisor", "citrix_zone")]
    [string[]] $ResourceTypes,

    [Parameter(Mandatory = $false)]
    [string[]] $NamesOrIds,

    [Parameter(Mandatory = $false)]
    [switch] $NoDependencyRelationship,

    [Parameter(Mandatory = $false)]
    [switch] $DisableSSLValidation,

    [Parameter(Mandatory=$false)]
    [switch] $ShowClientSecret,

    [Parameter(Mandatory=$false)]
    [string] $QuickDeployHostname,

    [Parameter(Mandatory=$false)]
    [string] $OutputFolder = "citrix-site"
)

### Helper Functions ###

function Get-Site {
    if ($script:onPremise) {
        $siteRequest = "https://$($script:hostname)/citrix/orchestration/api/me"
    }
    else {
        $siteRequest = "$(Get-DaasServiceBaseUrl)/cvad/manage/me"
    }

    $response = Start-GetRequest -url $siteRequest
    $script:siteId = $response.Customers[0].Sites[0].Id
}

function Get-RequestBaseUrl {
    if ($script:onPremise) {
        $url = "https://$($script:hostname)/citrix/orchestration/api/CitrixOnPremises/$($script:siteId)"
    }
    else {
        $url = "$(Get-DaasServiceBaseUrl)/cvad/manage"
    }

    $script:urlBase = $url
}

# Builds the base URL for the Citrix DaaS service. Honors an explicit -Hostname override, otherwise derives from environment.
function Get-DaasServiceBaseUrl {
    if (-not [string]::IsNullOrWhiteSpace($script:hostname)) {
        return "https://$($script:hostname)"
    }
    return Get-CloudManagementBaseUrl
}

# Function to get the URL for Quick Deploy objects
function Get-UrlForQuickDeployObjects {
    param(
        [parameter(Mandatory = $true)]
        [string] $requestPath
    )

    # Allow overriding the catalog service base (host + /catalogservice path) for non-standard environments,
    # mirroring the provider's CITRIX_QUICK_DEPLOY_HOST_NAME override. When set, it is used verbatim as the base.
    if (-not [string]::IsNullOrWhiteSpace($script:quickDeployHostnameOverride)) {
        $base = "https://$($script:quickDeployHostnameOverride)"
    }
    else {
        $base = "$(Get-DaasServiceBaseUrl)/catalogservice"
    }

    return "$base/$($script:customerId)/$($script:siteId)/$requestPath"
}

# Builds the base URL for Citrix Cloud APIs based on environment. Cloud-only; never uses -Hostname.
function Get-CloudManagementBaseUrl {
    $config = $script:environmentConfig[$script:environment]
    if ($null -ne $config) {
        return $config.ApiUrl
    }
    return $script:environmentConfig["Production"].ApiUrl
}

# Builds the base URL for the CWS (Citrix Workspace Services) API based on environment. Cloud-only; never uses -Hostname.
function Get-CwsBaseUrl {
    $config = $script:environmentConfig[$script:environment]
    if ($null -ne $config) {
        return $config.CwsUrl
    }
    return $script:environmentConfig["Production"].CwsUrl
}

# Function to enumerate Quick Deploy template images and build data source map
function Get-QuickDeployTemplateImageDataSources {
    Write-Verbose "Enumerating Quick Deploy template images for data source generation."

    $url = Get-UrlForQuickDeployObjects -requestPath "images"

    try {
        $response = Start-GetRequest -url $url
    }
    catch {
        Write-Verbose "Failed to enumerate Quick Deploy template images. Error: $($_.Exception.Message)"
        return @{}
    }

    $items = $response.items
    if (-not $items -or $items.Count -eq 0) {
        Write-Verbose "No Quick Deploy template images found."
        return @{}
    }

    # Build an id -> name lookup for every available template image. Data sources are NOT emitted for all of
    # these; only the ones actually referenced by an onboarded catalog are emitted (see ReplaceDependencyRelationships
    # / WriteQuickDeployTemplateImageDataSources), so a customer's onboarded TF stays scoped to what it actually uses.
    $templateImageMap = @{}
    $nameCountMap = @{}

    foreach ($item in $items) {
        if (-not $item.id -or -not $item.name) {
            continue
        }

        # Track duplicate names
        if ($nameCountMap.ContainsKey($item.name)) {
            $nameCountMap[$item.name] += 1
            Write-Warning "Duplicate Quick Deploy template image name detected: '$($item.name)'. Data source matching by name may be ambiguous."
        }
        else {
            $nameCountMap[$item.name] = 1
        }

        $templateImageMap[$item.id] = @{
            Name = $item.name
        }
    }

    Write-Verbose "Found $($templateImageMap.Count) Quick Deploy template images available for lookup."
    return $templateImageMap
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
        $url = "https://$($script:hostname)/citrix/orchestration/api/tokens"
        $base64AuthInfo = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes(("{0}\{1}:{2}" -f $($script:domainFqdn), $($script:clientId), $($script:clientSecret))))
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

        # Token endpoint is always environment-derived; the -Hostname override applies only to the DaaS/Quick Deploy service calls, not Citrix Cloud auth.
        $authConfig = $script:environmentConfig[$script:environment]
        if ($null -eq $authConfig) {
            $authConfig = $script:environmentConfig["Production"]
        }
        $url = $authConfig.AuthUrl -f $script:customerId

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
            "Citrix-CustomerId" = $($script:customerId)
            "Accept"            = "application/json"
        }
        if ($null -ne $script:siteId) {
            $headers["Citrix-InstanceId"] = $($script:siteId)
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
    $clientSecretPrefix = if ($ShowClientSecret) { "" } else { "# " }

    if (!(Test-Path ".\citrix.tf")) {
        New-Item -path ".\" -name "citrix.tf" -type "file" -Force
        Write-Verbose "Created new file for terraform citrix provider configuration."
    }
    if ($script:onPremise) {
        $disable_ssl_verification = $script:disable_ssl.ToString().ToLower()
        $config = @"
provider "citrix" {
    cvad_config = {
        hostname                    = "$($script:hostname)"
        client_id                   = "$($script:domainFqdn)\\$($script:clientId)"
        ${clientSecretPrefix}client_secret               = "$secretValue"
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
        customer_id                 = "$($script:customerId)"
        client_id                   = "$($script:clientId)"
        ${clientSecretPrefix}client_secret               = "$secretValue"
        hostname                    = "$($script:hostname)"
        environment                 = "$($script:environment)"
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
        [parameter(Mandatory = $false)]
        [string] $requestPath = "",

        [parameter(Mandatory = $true)]
        [string] $resourceProviderName,

        [parameter(Mandatory = $false)]
        [string] $overrideUrl = ""
    )

    if ($overrideUrl) {
        $url = $overrideUrl
    }
    elseif ($resourceProviderName -eq "quickdeploy_catalog") {
        $url = Get-UrlForQuickDeployObjects -requestPath $requestPath
    }
    else {
        $url = "$($script:urlBase)/$requestPath"
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

    # Quick Deploy catalogs endpoint returns response.items[] (canonical) or response.catalogs[] (backward-compat alias)
    if ($resourceProviderName -eq "quickdeploy_catalog") {
        $items = if ($null -ne $response.items) { $response.items } else { $response.catalogs }
    }

    # Cloud resource locations endpoint wraps results under "locations" rather than "Items"
    if ($resourceProviderName -eq "cloud_resource_location") {
        $items = if ($null -ne $response.locations) { $response.locations } else { $response.Items }
    }



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
            # Ids <= 1 are built-in/sentinel icons (including negative placeholder ids) and are not importable
            if ([int]$item.Id -le 1) {
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

        # Cloud Admin Users import by userId (accepted invitation) or ucOid; skip uninvited users
        if ($resourceProviderName -eq "cloud_admin_user") {
            if ($item.userId) {
                $resourceList += $item.userId
            }
            elseif ($item.ucOid) {
                $resourceList += $item.ucOid
            }
            continue
        }

        # Identity providers use idpInstanceId/idpNickname rather than Id/Name
        if ($resourceProviderName -in @("cloud_saml_identity_provider", "cloud_google_identity_provider", "cloud_okta_identity_provider")) {
            if (($NamesOrIds -and $NamesOrIds.Count -gt 0) -and
                -not (($item.idpInstanceId -and ($NamesOrIds -contains $item.idpInstanceId)) -or
                      ($item.idpNickname -and ($NamesOrIds -contains $item.idpNickname)))) {
                continue
            }
            if ($item.idpInstanceId) {
                $resourceList += $item.idpInstanceId
            }
            continue
        }
 
        # Create a path map for ApplicationFolder paths
        if ($requestPath -eq "AdminFolders") {
            $pathMap[$item.Id] = $item.Path
        }

        # Store icons as files
        if ($requestPath -like "Icons*") {
            $iconsFolder = Join-Path -Path $script:siteFolder -ChildPath "icons"
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

# Reads the current terraform state into lookup maps used to keep re-runs idempotent:
#   existingStateEntries - addresses already managed, so their stubs and imports are skipped
#   existingNameById     - resource identity -> name, so a resource keeps its original name across runs
#   usedNamesByType      - names already taken per type, so new resources never collide with an existing name
function Read-ExistingState {
    $script:existingStateEntries = @{}
    $script:existingNameById = @{}
    $script:usedNamesByType = @{}

    # Most resources expose their import identity as the "id" attribute, but a few import through a different attribute
    # (the provider uses ImportStatePassthroughID to a non-id path). Those attributes hold the value the script imports with,
    # so re-runs must match on them instead of "id" to recognize resources already in state.
    $identityAttributeByType = @{
        "citrix_cloud_admin_user" = "admin_id"
        "citrix_policy_priority"  = "policy_set_id"
    }

    $showJson = terraform show -json 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Verbose "No readable terraform state found; treating this as a first run."
        return
    }

    try {
        $parsed = ($showJson | Out-String) | ConvertFrom-Json
    }
    catch {
        Write-Warning "Failed to parse terraform state JSON; treating this as a first run. Error: $($_.Exception.Message)"
        return
    }

    if (-not ($parsed.values -and $parsed.values.root_module -and $parsed.values.root_module.resources)) {
        return
    }

    foreach ($res in $parsed.values.root_module.resources) {
        if ($res.mode -ne "managed") {
            continue
        }

        $script:existingStateEntries[$res.address] = $true

        if (-not $script:usedNamesByType.ContainsKey($res.type)) {
            $script:usedNamesByType[$res.type] = [System.Collections.Generic.HashSet[string]]::new()
        }
        [void]$script:usedNamesByType[$res.type].Add($res.name)

        $idAttr = if ($identityAttributeByType.ContainsKey($res.type)) { $identityAttributeByType[$res.type] } else { "id" }
        $resId = $res.values.$idAttr
        if ($null -ne $resId -and $resId -ne "") {
            $script:existingNameById["$($res.type)|$resId"] = $res.name
        }
    }
}

# Returns a stable terraform resource name for a discovered resource. A resource already in state keeps the
# name it was first imported under (matched by its own id, not its position), and a newly discovered resource
# gets the next free index so it can never collide with or overwrite an existing name when the remote set changes.
function Resolve-ResourceName {
    param(
        [parameter(Mandatory = $true)][string] $resourceProviderName,
        [parameter(Mandatory = $true)][string] $importKey,
        [parameter(Mandatory = $true)][string] $namePrefix,
        [parameter(Mandatory = $false)][int] $startIndex = 0
    )

    $type = "citrix_$resourceProviderName"
    $resourceId = ($importKey -split ',')[-1] # child import keys are "parentId,childId"; identity is the child's own id
    $idKey = "$type|$resourceId"

    if ($script:existingNameById.ContainsKey($idKey)) {
        return $script:existingNameById[$idKey]
    }

    if (-not $script:usedNamesByType.ContainsKey($type)) {
        $script:usedNamesByType[$type] = [System.Collections.Generic.HashSet[string]]::new()
    }
    $used = $script:usedNamesByType[$type]

    $index = $startIndex
    $name = "$namePrefix$index"
    while ($used.Contains($name)) {
        $index++
        $name = "$namePrefix$index"
    }

    [void]$used.Add($name)
    $script:existingNameById[$idKey] = $name # keep repeated references to the same id stable within this run
    return $name
}

# Function to get import map for each resource
function Get-ImportMap {
    param(
        [parameter(Mandatory = $false)]
        [string] $resourceApi = "",

        [parameter(Mandatory = $true)]
        [string] $resourceProviderName,

        [parameter(Mandatory = $false)]
        [string] $parentId = "",

        [parameter(Mandatory = $false)]
        [int] $parentIndex = 0,

        [parameter(Mandatory = $false)]
        [string] $overrideUrl = ""
    )

    $list, $pathMap = Get-ResourceList -requestPath $resourceApi -resourceProviderName $resourceProviderName -overrideUrl $overrideUrl
    $resourceMap = @{}
    $index = 0
    foreach ($id in $list) {
        if ($parentId -ne "") {
            $resourceMapKey = "$($parentId),$($id)"
            $resourceName = Resolve-ResourceName -resourceProviderName $resourceProviderName -importKey $resourceMapKey -namePrefix "$($resourceProviderName)_$($parentIndex)_" -startIndex $index
            if (-not $script:parentChildMap.ContainsKey($parentId)) {
                # Initialize as a new list if not already present
                $script:parentChildMap[$parentId] = [System.Collections.Generic.List[string]]::new()
            }
            $script:parentChildMap[$parentId].Add($id)
        }
        else {
            $resourceMapKey = $id
            $resourceName = Resolve-ResourceName -resourceProviderName $resourceProviderName -importKey $resourceMapKey -namePrefix "$($resourceProviderName)_" -startIndex $index
        }
        
        if ($resourceApi -eq "AdminFolders" -and $pathMap.Count -gt 0) {
            $script:applicationFolderPathMap[$pathMap.$id.TrimEnd('\')] = $resourceName
        }
        
        $resourceMap[$resourceMapKey] = $resourceName
        # Only write a stub if this resource is not already declared in an existing .tf file from a prior run
        if (-not $script:existingStateEntries.ContainsKey("citrix_$($resourceProviderName).$resourceName")) {
            Add-Content -Path ".\import.tf" -Value "resource `"citrix_$resourceProviderName`" `"$resourceName`" {}`n" -Encoding utf8
        }
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
        "policy_set_v2"        = @{
            "resourceApi"          = "gpo/policySets"
            "resourceProviderName" = "policy_set_v2"
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

    # Add Quick Deploy and Citrix Cloud resources for cloud customers (Quick Deploy is cloud-only)
    if (-not($script:onPremise)) {
        $resources.Add("quickdeploy_catalog", @{
            "resourceApi"          = "catalogs"
            "resourceProviderName" = "quickdeploy_catalog"
        })
        $cloudMgmtBase = Get-CloudManagementBaseUrl
        $resources.Add("cloud_resource_location", @{
            "resourceApi"          = ""
            "resourceProviderName" = "cloud_resource_location"
            "overrideUrl"          = "$cloudMgmtBase/resourcelocations/"
        })
        $resources.Add("cloud_admin_user", @{
            "resourceApi"          = ""
            "resourceProviderName" = "cloud_admin_user"
            "overrideUrl"          = "$cloudMgmtBase/administrators/"
        })
        $cwsBase = Get-CwsBaseUrl
        $resources.Add("cloud_saml_identity_provider", @{
            "resourceApi"          = ""
            "resourceProviderName" = "cloud_saml_identity_provider"
            "overrideUrl"          = "$cwsBase/$($script:customerId)/identityProviders/saml"
        })
        $resources.Add("cloud_google_identity_provider", @{
            "resourceApi"          = ""
            "resourceProviderName" = "cloud_google_identity_provider"
            "overrideUrl"          = "$cwsBase/$($script:customerId)/identityProviders/google"
        })
        $resources.Add("cloud_okta_identity_provider", @{
            "resourceApi"          = ""
            "resourceProviderName" = "cloud_okta_identity_provider"
            "overrideUrl"          = "$cwsBase/$($script:customerId)/identityProviders/okta"
        })
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
        $overrideUrl = $resources[$resource].overrideUrl
        $script:cvadResourcesMap[$resource] = Get-ImportMap -resourceApi $api -resourceProviderName $resourceProviderName -overrideUrl $overrideUrl
        Write-Verbose "Collected $($script:cvadResourcesMap[$resource].Count) resource(s) for type '$resource'"

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
        # Create policy, policy_priority, policy_setting, and filter resources for each policy_set_v2
        if ($resource -eq "policy_set_v2") {
            $policyIndex = 0
            $priorityIndex = 0
            $settingIndex = 0
            $filterIndexes = @{}

            foreach ($policySetId in $script:cvadResourcesMap[$resource].Keys) {
                # One citrix_policy_priority per policy set; import key is the policy set GUID
                if (-not $script:cvadResourcesMap.ContainsKey("policy_priority")) {
                    $script:cvadResourcesMap["policy_priority"] = @{}
                }
                $priorityName = Resolve-ResourceName -resourceProviderName "policy_priority" -importKey $policySetId -namePrefix "policy_priority_" -startIndex $priorityIndex
                $script:cvadResourcesMap["policy_priority"][$policySetId] = $priorityName
                if (-not $script:existingStateEntries.ContainsKey("citrix_policy_priority.$priorityName")) {
                    Add-Content -Path ".\import.tf" -Value "resource `"citrix_policy_priority`" `"$priorityName`" {}`n" -Encoding utf8
                }
                $priorityIndex++

                try {
                    $policiesResponse = Start-GetRequest -url "$($script:urlBase)/gpo/policies?policySetGuid=$policySetId"
                    $policyItems = if ($policiesResponse.Items) { $policiesResponse.Items } else { @() }

                    foreach ($policyItem in $policyItems) {
                        $policyId = $policyItem.policyGuid
                        if (-not $policyId) { continue }

                        if (-not $script:cvadResourcesMap.ContainsKey("policy")) {
                            $script:cvadResourcesMap["policy"] = @{}
                        }
                        $policyName = Resolve-ResourceName -resourceProviderName "policy" -importKey $policyId -namePrefix "policy_" -startIndex $policyIndex
                        $script:cvadResourcesMap["policy"][$policyId] = $policyName
                        if (-not $script:existingStateEntries.ContainsKey("citrix_policy.$policyName")) {
                            Add-Content -Path ".\import.tf" -Value "resource `"citrix_policy`" `"$policyName`" {}`n" -Encoding utf8
                        }
                        $policyIndex++

                        try {
                            $settingsResponse = Start-GetRequest -url "$($script:urlBase)/gpo/settings?policyGuid=$policyId"
                            $settingItems = if ($settingsResponse.Items) { $settingsResponse.Items } else { @() }

                            foreach ($settingItem in $settingItems) {
                                $settingId = $settingItem.settingGuid
                                if (-not $settingId) { continue }

                                if (-not $script:cvadResourcesMap.ContainsKey("policy_setting")) {
                                    $script:cvadResourcesMap["policy_setting"] = @{}
                                }
                                $settingName = Resolve-ResourceName -resourceProviderName "policy_setting" -importKey $settingId -namePrefix "policy_setting_" -startIndex $settingIndex
                                $script:cvadResourcesMap["policy_setting"][$settingId] = $settingName
                                if (-not $script:existingStateEntries.ContainsKey("citrix_policy_setting.$settingName")) {
                                    Add-Content -Path ".\import.tf" -Value "resource `"citrix_policy_setting`" `"$settingName`" {}`n" -Encoding utf8
                                }
                                $settingIndex++
                            }
                        }
                        catch {
                            Write-Warning "Failed to get settings for policy $policyId. Error: $($_.Exception.Message)"
                        }

                        try {
                            $filtersResponse = Start-GetRequest -url "$($script:urlBase)/gpo/filters?policyGuid=$policyId"
                            $filterItems = if ($filtersResponse.Items) { $filtersResponse.Items } else { @() }

                            foreach ($filterItem in $filterItems) {
                                $filterId = $filterItem.filterGuid
                                $filterType = $filterItem.filterType
                                if (-not $filterId -or -not $filterType) { continue }

                                $filterResourceType = $script:policyFilterTypeMap[$filterType]
                                if (-not $filterResourceType) {
                                    Write-Warning "Unknown policy filter type '$filterType'. Skipping filter $filterId."
                                    continue
                                }

                                if (-not $filterIndexes.ContainsKey($filterResourceType)) {
                                    $filterIndexes[$filterResourceType] = 0
                                }
                                $filterIdx = $filterIndexes[$filterResourceType]
                                $filterResourceName = Resolve-ResourceName -resourceProviderName $filterResourceType -importKey $filterId -namePrefix "${filterResourceType}_" -startIndex $filterIdx
                                $filterIndexes[$filterResourceType]++

                                if (-not $script:cvadResourcesMap.ContainsKey($filterResourceType)) {
                                    $script:cvadResourcesMap[$filterResourceType] = @{}
                                }
                                $script:cvadResourcesMap[$filterResourceType][$filterId] = $filterResourceName
                                if (-not $script:existingStateEntries.ContainsKey("citrix_$filterResourceType.$filterResourceName")) {
                                    Add-Content -Path ".\import.tf" -Value "resource `"citrix_$filterResourceType`" `"$filterResourceName`" {}`n" -Encoding utf8
                                }
                            }
                        }
                        catch {
                            Write-Warning "Failed to get filters for policy $policyId. Error: $($_.Exception.Message)"
                        }
                    }
                }
                catch {
                    Write-Warning "Failed to get policies for policy set $policySetId. Error: $($_.Exception.Message)"
                }
            }
        }
    }
    Write-Verbose "Successfully retrieved all CVAD resources from the site."
}

# Runs a single terraform import, retrying with backoff. Each import is a separate provider process; publishing the
# access token (cloud only) lets them all reuse one sign-in. Backoff still guards transient errors.
function Invoke-TerraformImportWithRetry {
    param(
        [parameter(Mandatory = $true)][string] $ResourcePath,
        [parameter(Mandatory = $true)][string] $Id,
        [parameter(Mandatory = $false)][int] $MaxRetries = 6
    )

    $attempt = 0
    while ($true) {
        $attempt++
        # Publish a fresh token before every spawn so even a retry that waited out an expiry hands the child a valid one.
        if (-not $script:onPremise) {
            $env:CITRIX_ACCESS_TOKEN = Get-AuthToken
        }
        $output = terraform import "$ResourcePath" "$Id" 2>&1
        $output | ForEach-Object { Write-Host $_ }

        if ($LASTEXITCODE -eq 0) {
            return $true
        }

        $outputText = $output | Out-String

        # The provider's post-import read could not retrieve this object by id; retrying will not help, so stop immediately.
        if ($outputText -match "Cannot import non-existent remote object") {
            Write-Warning "Skipping $ResourcePath (id '$Id'): the provider could not read this object by id after import. It was listed during enumeration but is not retrievable for import in this context."
            return $false
        }

        $isRateLimited = $outputText -match "Too many requests"

        if ($attempt -ge $MaxRetries) {
            Write-Warning "Import FAILED for $ResourcePath (id '$Id') after $attempt attempts. Dependent resources may reference an undeclared resource."
            return $false
        }

        # Rate limiting resets on a longer horizon than transient errors, so back off harder (capped) for it.
        $delay = if ($isRateLimited) { [math]::Min(90, [math]::Pow(2, $attempt) * 5) } else { [math]::Min(30, [math]::Pow(2, $attempt)) }
        Write-Warning "Import attempt $attempt for $ResourcePath failed (exit $LASTEXITCODE). Retrying in $delay seconds..."
        Start-Sleep -Seconds $delay
    }
}

# Function to import terraform resources into state
function Import-ResourcesToState {
    $script:newlyImportedResources = [System.Collections.Generic.List[string]]::new()

    foreach ($resource in $script:cvadResourcesMap.Keys) {
        foreach ($id in $script:cvadResourcesMap[$resource].Keys) {
            $resourcePath = "citrix_$($resource).$($script:cvadResourcesMap[$resource][$id])"
            if ($script:existingStateEntries.ContainsKey($resourcePath)) {
                Write-Verbose "Skipping $resourcePath (already in state)"
                continue
            }
            Write-Verbose "Importing $resourcePath with id '$id'"
            $imported = Invoke-TerraformImportWithRetry -ResourcePath $resourcePath -Id $id
            if ($imported) {
                $script:newlyImportedResources.Add($resourcePath)
            }
        }
    }
    Write-Verbose "Imported $($script:newlyImportedResources.Count) new resource(s) into state."
}

# On re-run, appends newly imported resources to their existing per-type .tf files using terraform state show.
function Add-NewResourcesToExistingTfFiles {
    Write-Verbose "Appending $($script:newlyImportedResources.Count) newly imported resource(s) to existing .tf files."

    foreach ($resourcePath in $script:newlyImportedResources) {
        # Derive filename from the type segment: "citrix_machine_catalog.machine_catalog_0" -> "citrix_machine_catalog.tf"
        $resourceTypeFull = ($resourcePath -split '\.')[0]
        $filename = Resolve-TfFileName -resourceType $resourceTypeFull

        $showOutput = terraform state show -no-color $resourcePath 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Warning "Failed to get state for $resourcePath. Skipping."
            continue
        }

        $content = $showOutput | Out-String
        $content = InjectPlaceHolderSensitiveValues -content $content
        $content = RemoveApplicationSecretForManagedIdentity -content $content
        $content = ReplaceDependencyRelationships -content $content
        $content = RemoveComputedProperties -content $content

        Add-Content -Path ".\$filename" -Value "`n$content" -Encoding utf8
        Write-Verbose "Appended $resourcePath to $filename"
    }

    WriteQuickDeployTemplateImageDataSources
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
        "(\s+)tenants\s*=\s*\[[\s\S]*?\]",
        "(\s+)max_number_of_users(\s+)= (\S+)",
        "(\s+)admin_id(\s+)= (\S+)",
        "(\s+)policy_set_name(\s+)= (\S+)",
        "(\s+)policy_names\s*=\s*\[[\s\S]*?\]",
        # citrix_cloud_saml_identity_provider computed fields
        '(\s+)cert_common_name\s*=\s*"[^"]*"',
        '(\s+)cert_expiration\s*=\s*"[^"]*"',
        "(\s+)scoped_entity_id_suffix(\s+)= (\S+)",
        # citrix_cloud_google_identity_provider computed fields
        "(\s+)google_customer_id(\s+)= (\S+)",
        '(\s+)google_domain\s*=\s*"[^"]*"'
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
        $content = $content -replace "PLACEHOLDER_DELIVERY_GROUPS_PRIORITY_$index\s", "$($match.Value)`n"
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

        # policy_priority is imported by the policy set GUID, colliding with policy_set_v2's key. It is never referenced
        # by id, so skip it here; the shared GUID then always resolves to the policy set that policy_set_id expects.
        if ($resource -eq "policy_priority") {
            continue
        }

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

    # Replace Quick Deploy template image ID references with data source references.
    # Only template images actually referenced by an onboarded catalog get a data source (and a stable, reference-
    # ordered resource name), so the generated TF is not polluted with a data block for every available image.
    $script:referencedQuickDeployTemplateImages = @()
    if ($script:quickDeployTemplateImageMap.Count -gt 0) {
        $referencedImageIds = @()
        foreach ($imageId in $script:quickDeployTemplateImageMap.Keys) {
            if ($content -match "template_image_id\s*=\s*`"$([regex]::Escape($imageId))`"") {
                $referencedImageIds += $imageId
            }
        }
        $index = 0
        foreach ($imageId in ($referencedImageIds | Sort-Object)) {
            $resourceName = "quickdeploy_template_image_$index"
            $imageName = $script:quickDeployTemplateImageMap[$imageId].Name
            Write-Verbose "Replacing Quick Deploy template image ID: $imageId with data.citrix_quickdeploy_template_image.$resourceName.id"
            $content = $content -replace "(template_image_id\s*=\s*)`"$([regex]::Escape($imageId))`"", "`${1}data.citrix_quickdeploy_template_image.$resourceName.id"
            $script:referencedQuickDeployTemplateImages += [PSCustomObject]@{ Id = $imageId; Name = $imageName; ResourceName = $resourceName }
            $index++
        }
    }

    return $content
}

function WriteQuickDeployTemplateImageDataSources {
    if (-not $script:referencedQuickDeployTemplateImages -or $script:referencedQuickDeployTemplateImages.Count -eq 0) {
        Write-Verbose "No referenced Quick Deploy template images to generate data sources for."
        return
    }

    Write-Verbose "Writing Quick Deploy template image data source blocks to dedicated file."
    $dataSourceBlocks = ""

    foreach ($image in $script:referencedQuickDeployTemplateImages) {
        $resourceName = $image.ResourceName
        $imageName = $image.Name

        $dataSourceBlocks += @"
data "citrix_quickdeploy_template_image" "$resourceName" {
  name = "$imageName"
}


"@
    }

    # Write data source blocks to dedicated file
    $filename = "quickdeploy_template_image.tf"
    Add-Content -Path ".\$filename" -Value $dataSourceBlocks -Encoding utf8
    Write-Verbose "Wrote $($script:referencedQuickDeployTemplateImages.Count) Quick Deploy template image data source blocks to $filename."
}

function InjectPlaceHolderSensitiveValues {
    param(
        [parameter(Mandatory = $true)]
        [string] $content
    )

    $filteredOutput = @()
    $lines = $content -split "`r?`n"
    $iconsFolder = Join-Path -Path $script:siteFolder -ChildPath "icons"

    $previousLine = ""
    $currentIdpType = $null  # "saml", "google", or "okta" when inside an identity provider block
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

        if ($line -match '^\s*resource\s*"citrix_cloud_saml_identity_provider"') {
            $currentIdpType = "saml"
        } elseif ($line -match '^\s*resource\s*"citrix_cloud_google_identity_provider"') {
            $currentIdpType = "google"
        } elseif ($line -match '^\s*resource\s*"citrix_cloud_okta_identity_provider"') {
            $currentIdpType = "okta"
        } elseif ($null -ne $currentIdpType -and $line -eq "}") {
            # Inject required fields that are null after import before the closing brace
            switch ($currentIdpType) {
                "saml" {
                    $filteredOutput += '  cert_file_path = "<path to certificate file (.pem, .crt, or .cer)>"'
                }
                "google" {
                    $filteredOutput += '  private_key = "<input Google service account private key>"'
                    $filteredOutput += '  impersonated_user = "<input impersonated user email>"'
                }
                "okta" {
                    $filteredOutput += '  okta_client_secret = "<input Okta client secret>"'
                    $filteredOutput += '  okta_api_token = "<input Okta API token>"'
                }
            }
            $currentIdpType = $null
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

function RemoveApplicationSecretForManagedIdentity {
    param(
        [parameter(Mandatory = $true)]
        [string] $content
    )

    # Azure hypervisor blocks are flat (no nested braces), so [^}]* safely spans a single resource block.
    $azureBlockPattern = 'resource\s+"citrix_azure_hypervisor"\s+"[^"]+"\s*\{[^}]*\}'
    $evaluator = {
        param($match)
        $block = $match.Value
        if ($block -match 'authentication_mode\s*=\s*"(UserAssignedManagedIdentity|SystemAssignedManagedIdentity)"') {
            $block = $block -replace '(?m)^\s*application_secret\s*=.*\r?\n', ''
        }
        return $block
    }
    return [regex]::Replace($content, $azureBlockPattern, $evaluator)
}

# Maps a resource type name to the .tf filename it should be written to.
# All policy-related types (policy sets, policies, settings, and all filter types) are grouped into a single file.
function Resolve-TfFileName {
    param([parameter(Mandatory = $true)][string] $resourceType)
    if ($resourceType -match '^citrix_policy' -or $resourceType -match '_policy_filter$') {
        return "citrix_policy.tf"
    }
    return "$resourceType.tf"
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
        $filename = Resolve-TfFileName -resourceType $resourceType

        # Append the resource block to the file
        Add-Content -Path $filename -Value $resourceBlock -Encoding utf8
        Add-Content -Path $filename -Value "`n" -Encoding utf8  # Add a newline for separation
    }

    Write-Verbose "Resource files created successfully."
}

function PostProcessTerraformOutput {

    # Post-process the terraform output
    $content = Get-Content -Path ".\resource.tf" -Raw -Encoding utf8

    # Inject placeholder for sensitive values in tf
    $content = InjectPlaceHolderSensitiveValues -content $content

    # Remove application_secret from Azure hypervisors using a managed identity (the provider rejects it there)
    $content = RemoveApplicationSecretForManagedIdentity -content $content

    # Set dependency relationships
    $content = ReplaceDependencyRelationships -content $content

    # Remove computed properties
    $content = RemoveComputedProperties -content $content

    # Overwrite extracted terraform with processed value
    Set-Content -Path ".\resource.tf" -Value $content -Encoding utf8

    # Write Quick Deploy template image data sources to dedicated file
    WriteQuickDeployTemplateImageDataSources

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
# Optional override for the Quick Deploy catalog service base; falls back to the provider's CITRIX_QUICK_DEPLOY_HOST_NAME env var.
$script:quickDeployHostnameOverride = if (-not [string]::IsNullOrWhiteSpace($QuickDeployHostname)) { $QuickDeployHostname } else { $env:CITRIX_QUICK_DEPLOY_HOST_NAME }
# Single source of truth for Citrix Cloud endpoints per environment. ApiUrl backs the management APIs; AuthUrl ({0} = customer id) backs the OAuth token request. Values mirror the provider's environment mapping.
$script:environmentConfig = @{
    "Production"   = @{ ApiUrl = "https://api.cloud.com";             AuthUrl = "https://api.cloud.com/cctrustoauth2/{0}/tokens/clients";             CwsUrl = "https://cws.citrixworkspacesapi.net" }
    "Staging"      = @{ ApiUrl = "https://api.cloudburrito.com";      AuthUrl = "https://api.cloudburrito.com/cctrustoauth2/{0}/tokens/clients";      CwsUrl = "https://cws.ctxwsstgapi.net" }
    "Japan"        = @{ ApiUrl = "https://api.citrixcloud.jp";        AuthUrl = "https://api.citrixcloud.jp/cctrustoauth2/{0}/tokens/clients";        CwsUrl = "https://cws.citrixworkspacesapi.jp" }
    "JapanStaging" = @{ ApiUrl = "https://api.citrixcloudstaging.jp"; AuthUrl = "https://api.citrixcloudstaging.jp/cctrustoauth2/{0}/tokens/clients"; CwsUrl = "https://cws.citrixstagingapi.jp" }
    "Gov"          = @{ ApiUrl = "https://api.cloud.us";              AuthUrl = "https://trust.citrixworkspacesapi.us/{0}/tokens/clients";             CwsUrl = "https://cws.citrixworkspacesapi.us" }
    "GovStaging"   = @{ ApiUrl = "https://api.cloudstaging.us";       AuthUrl = "https://trust.ctxwsstgapi.us/{0}/tokens/clients";                    CwsUrl = "https://cws.ctxwsstgapi.us" }
}
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
$script:quickDeployTemplateImageMap = @{} # Map of Quick Deploy template image ID to @{Name} for data source name lookup
$script:referencedQuickDeployTemplateImages = @() # Ordered list of template images actually referenced by onboarded catalogs (Id/Name/ResourceName)
$script:policyFilterTypeMap = @{
    "AccessControl"  = "access_control_policy_filter"
    "BranchRepeater" = "branch_repeater_policy_filter"
    "ClientIP"       = "client_ip_policy_filter"
    "ClientName"     = "client_name_policy_filter"
    "ClientPlatform" = "client_platform_policy_filter"
    "DesktopGroup"   = "delivery_group_policy_filter"
    "DesktopKind"    = "delivery_group_type_policy_filter"
    "OU"             = "ou_policy_filter"
    "DesktopTag"     = "tag_policy_filter"
    "User"           = "user_policy_filter"
}

$script:TokenExpiryTime = (Get-Date).AddMinutes(-1) # Initialize the expiry time of the refresh token to an earlier time
$script:existingStateEntries = @{} # Populated after terraform init; controls which import stubs are written and which imports are skipped
$script:existingNameById = @{}     # "citrix_<type>|<resourceId>" -> terraform resource name already in state, so re-runs reuse names by identity instead of by unstable position
$script:usedNamesByType = @{}      # "citrix_<type>" -> HashSet of names already taken, so newly discovered resources never collide with an existing name
$script:siteFolder = $null         # Subfolder holding the generated Terraform project (config, state, icons)
$script:pushedSiteLocation = $false

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

    # Keep the generated configuration together in a dedicated subfolder so it can serve as the customer's ongoing
    # Terraform project instead of scattering .tf files, state, and icons across the folder holding the script.
    $script:siteFolder = Join-Path -Path $PSScriptRoot -ChildPath $OutputFolder
    if (-not (Test-Path -Path $script:siteFolder)) {
        New-Item -ItemType Directory -Path $script:siteFolder | Out-Null
        Write-Verbose "Created site configuration folder: $script:siteFolder"
    }

    # terraform.tf pins the provider; seed it into the site folder on first run without clobbering later customer edits.
    $rootTerraformTf = Join-Path -Path $PSScriptRoot -ChildPath "terraform.tf"
    $siteTerraformTf = Join-Path -Path $script:siteFolder -ChildPath "terraform.tf"
    if ((Test-Path -Path $rootTerraformTf) -and -not (Test-Path -Path $siteTerraformTf)) {
        Copy-Item -Path $rootTerraformTf -Destination $siteTerraformTf
        Write-Verbose "Copied terraform.tf into the site configuration folder."
    }

    # Run the rest of the flow inside the site folder so all relative paths and terraform state live there.
    Push-Location -Path $script:siteFolder
    $script:pushedSiteLocation = $true

    New-RequiredFiles

    # Initialize terraform before enumerating resources so import.tf is still empty and cannot conflict
    # with resource blocks already declared in per-type .tf files from a prior run.
    terraform init

    # Read existing state so re-runs reuse names by resource identity and skip resources already imported.
    Read-ExistingState
    $script:preExistingStateCount = $script:existingStateEntries.Count
    Write-Verbose "Found $($script:preExistingStateCount) resource(s) already in state."

    # Get CVAD resources from existing site (only writes import stubs for resources not already in state)
    Get-ExistingCVADResources $resouceTypesWithoutCitrixPrefix

    # Enumerate Quick Deploy template images for data source generation (cloud-only)
    if (-not($script:onPremise) -and ($null -eq $resouceTypesWithoutCitrixPrefix -or $resouceTypesWithoutCitrixPrefix -contains "quickdeploy_catalog")) {
        $script:quickDeployTemplateImageMap = Get-QuickDeployTemplateImageDataSources
    }

    # Import terraform resources into state, skipping any already present
    Import-ResourcesToState

    if ($script:preExistingStateCount -eq 0) {
        # First run: generate .tf files from the full state dump
        $prev = [Console]::OutputEncoding
        [Console]::OutputEncoding = [System.Text.UTF8Encoding]::new()
        terraform show -no-color >> ".\resource.tf"
        [Console]::OutputEncoding = $prev

        PostProcessTerraformOutput

        Remove-Item ".\import.tf"
        Remove-Item ".\resource.tf"
    } else {
        # Re-run: append only newly imported resources to their existing per-type .tf files
        if ($script:newlyImportedResources.Count -gt 0) {
            Add-NewResourcesToExistingTfFiles
        } else {
            Write-Host "All resources already in state. No new resources to add."
        }

        Remove-Item ".\import.tf" -ErrorAction SilentlyContinue
        Remove-Item ".\resource.tf" -ErrorAction SilentlyContinue
    }

    # Post-process citrix.tf (uncomments client_secret placeholder)
    PostProcessProviderConfig

    # Format terraform files
    terraform fmt

    Write-Host ""
    Write-Host "Onboarding complete. The generated Terraform project is in: $script:siteFolder" -ForegroundColor Green
    Write-Host "Run all Terraform commands (terraform plan, terraform apply) from that folder:" -ForegroundColor Green
    Write-Host "    cd `"$script:siteFolder`"; terraform plan" -ForegroundColor Green
}
finally {
    # Return to the original directory if we switched into the site folder
    if ($script:pushedSiteLocation) {
        Pop-Location
    }

    # Clean up environment variables for client secret and the shared import access token
    $env:CITRIX_CLIENT_SECRET = ''
    $env:CITRIX_ACCESS_TOKEN = ''
}