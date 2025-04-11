# Policy Set Resource

## Enable Policy Set Management

Please navigate to `Setting` node in WebStudio. Enable `Policy sets` setting to show `policy sets` tab in the `Policies` node. This feature is in **PREVIEW**.

## Policy Set Resource Example
```
// Using citrix_policy_set
resource "citrix_policy_set" "example-policy-set" {
    name = "Policy Set Name"
    description = "Policy Set Description"
    type = "DeliveryGroupPolicies"
    scopes = [ "citrix_admin_scope.example-admin-scope.id" ]
    policies = [
        {
            name = "Name of the Policy with Priority 0"
            description = "Policy in the example policy set with priority 0"
            enabled = true
            policy_settings = [
                {
                    name = "AdvanceWarningPeriod"
                    value = "13:00:00"
                    use_default = false
                },
            ]
            delivery_group_filters = [
                {
                    delivery_group_id   = citrix_delivery_group.example-delivery-group.id
                    enabled = true
                    allowed = true
                },
            ]
        },
        {
            name = "Name of the Policy with Priority 1"
            description = "Policy in the example policy set with priority 1"
            enabled = false
            policy_settings = []
        }
    ]
}

// Using citrix_policy_set_v2
resource "citrix_policy_set_v2" "policy_set_v2" {
    name        = "Example Policy Set V2"
    description = "Policy Set V2 Description"
    scopes      = [ citrix_admin_scope.example-admin-scope.id ]
    delivery_groups = [ citrix_delivery_group.example-delivery-group.id ]
}

resource "citrix_policy" "first_policy" {
    policy_set_id   = citrix_policy_set_v2.policy_set_v2.id
    name            = "Name of the Policy with Priority 0"
    description     = "Policy in the example policy set v2 with priority 0"
    enabled         = true
}

resource "citrix_policy_setting" "advance_warning_period" {
    policy_id   = citrix_policy.first_policy.id
    name        = "AdvanceWarningPeriod"
    use_default = false
    value       = "13:00:00"
}

resource "citrix_delivery_group_policy_filter" "delivery_group_filter" {
    policy_id          = citrix_policy.first_policy.id
    enabled            = true
    allowed            = true
    delivery_group_id  = citrix_delivery_group.example-delivery-group.id
}

resource "citrix_policy" "second_policy" {
    policy_set_id   = citrix_policy_set_v2.policy_set_v2.id
    name            = "Name of the Policy with Priority 1"
    description     = "Policy in the example policy set v2 with priority 1"
    enabled         = false
}

resource "citrix_policy_priority" "policy_priority" {
    policy_set_id    = citrix_policy_set_v2.policy_set_v2.id
    policy_priority  = [
        citrix_policy.first_policy.id,
        citrix_policy.second_policy.id
    ]
}


```

* Please refer to [Available Policy Filters](#available-policy-filters) for more details on policy filters. 

* Please refer to [Available Policy Settings](#available-policy-settings) for more details on policy settings

## Available Policy Filters

### Access Control
Filter Type: `AccessControl`

Example: 
```
# With Citrix Gateway
access_control_filters = [
    {
        enabled    = true
        allowed    = true
        connection = "WithAccessGateway"
        condition  = {Access Condition} // Wildcard `*` is allowed
        gateway    = {Gateway farm name} // Wildcard `*` is allowed
    }
]

# Without Citrix Gateway
access_control_filters = [
    {
        enabled    = true
        allowed    = true
        connection = "WithoutAccessGateway"
        condition  = "*"
        gateway    = "*"
    }
]
```

### Citrix SD-WAN
Filter Type: `BranchRepeater`

Filter Data should not be specified

When `allowed` is set to `true`, this means policy is applied to `Connections with Citrix SD-WAN`. When it is set to `false`, this means policy is applied to `Connections without Citrix SD-WAN`.

Example: 
```
branch_repeater_filter = {
    allowed    = true
}
```

```
branch_repeater_filter = {
    allowed    = false
}
```


### Client IP Address
Filter Type: `ClientIP`

Example: 
```
client_ip_filters = [
    {
        enabled    = true
        allowed    = true
        ip_address = "{IP address to be filtered}"
    }
]
```

### Client Name
Filter Type: `Client Name`

Example: 
```
client_name_filters = [
    {
        enabled     = true
        allowed     = true
        client_name = "{Name of the client to be filtered}"
    }
]
```

### Delivery Group
Filter Type: `DesktopGroup`

Example: 
```
delivery_group_filters = [
    {
        enabled           = true
        allowed           = true
        delivery_group_id = "{ID of the delivery group to be filtered}"
    }
]
```

### Delivery Group Type
Filter Type: `DesktopKind`

Filter Data:
Delivery Group Type | Filter Data
-- | --
Private Desktop | `Private`
Private Application | `PrivateApp`
Shared Desktop | `Shared`
Shared Application | `SharedApp`

Example: 
```
delivery_group_type_filters = [
    {
        enabled             = true
        allowed             = true
        delivery_group_type = "{Type of the delivery group to be filtered}"
    }
]
```


### Organizational Unit (OU)
Filter Type: `OU`

Example: 
```
ou_filters = [
    {
        enabled = true
        allowed = true
        ou      = "{Path of the organizational unit to be filtered}"
    }
]
```

### User or Group
Filter Type: `User`

Filter Data: Sid of the user or group

Filter Data Example: `S-1-5-21-4235287923-3346439331-1564732298-1103`

Example: 
```
user_filters = [
    {
        enabled = true
        allowed = true
        sid     = "{SID of the user or user group to be filtered}"
    }
]
```

### Tag
Filter Type: `DesktopTag`

Example: 
```
tag_filters = [
    {
        enabled = true
        allowed = true
        tag     = "{ID of the tag to be filtered}"
    }
]
```

## Available Policy Settings

### Accelerate folder mirroring
Description: 
```
With both this policy and the Folders to mirror policy enabled, Profile Management stores mirrored folders on a VHDX-based virtual disk. It attaches the virtual disk during logons and detaches it during logoffs. Enabling this policy eliminates the need to copy the folders between the user store and local profiles and accelerates folder mirroring.
```

Setting Name: `AccelerateFolderMirroring`

Setting Value:
```
{
    enabled = true | false
}
```


### Active Directory actions
Description: 
```
Detailed log settings.

Define events or actions that Profile management logs in depth.

If this setting is not configured here, Profile management uses the settings from the .ini file.

If this setting is not configured here or in the .ini file, errors and general information are logged.
```

Setting Name: `LogLevel_ActiveDirectoryActions`

Setting Value:
```
{
    enabled = true | false
}
```

### Active write back
Description: 
```
With this setting, files and folders (but not Registry entries) that are modified can be synchronized to the user store in the middle of a session, before logoff.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, active write back is disabled.
```

Setting Name: `PSMidSessionWriteBack`

Setting Value:
```
{
    enabled = true | false
}
```

### Active write back on session lock and disconnection
Description: 
```
Use this policy as an extension to the "Active write back" and "Active write back registry" policies.

With the "Active write back" policy enabled, by default, files and folders are written back from the local computer to the user store every five minutes. With both this new policy and the "Active write back" policy enabled, files and folders are written back only when a session is locked or disconnected.

With both the "Active write back" and "Active write back registry" policies enabled, by default, registry entries are written back from the local computer to the user store every five minutes. With this new policy and the "Active write back" and "Active write back registry" policies enabled, registry entries are written back only when a session is locked or disconnected.

If this setting is not configured here, the value from the .ini file is used.

If this setting is configured neither here nor in the .ini file, this feature is disabled.
```

Setting Name: `PSMidSessionWriteBackSessionLock`

Setting Value:
```
{
    enabled = true | false
}
```

### Active write back registry
Description: 
```
Leverage this policy with "Active write back" enabled.

Registry entries that are modified can be synchronized to the user store in the middle of a session.

If you do not configure this setting here, the value from the .ini file is used.

If you do not configure this setting here or in the .ini file, active write back registry is disabled.
```

Setting Name: `PSMidSessionWriteBackReg`

Setting Value:
```
{
    enabled = true | false
}
```

### Adaptive audio
Description: 
```
Adaptive audio policy settings dynamically adjust to provide an optimal user experience
```

Setting Name: `EnableAdaptiveAudio`

Setting Value:
```
{
    enabled = true | false
}
```

### Advance warning frequency interval
Description: 
```
For Manual and PVS Server OS catalogs, the Connector Agent service for Configuration Manager 2012 alerts users in advance when there are pending application installs and/or software updates that will take place in the near future so that they can plan their use of the system accordingly. For MCS catalogs, use Studio to send messages to users.

This setting determines how often advance warnings should popup.

A time span string consists of the following:

ddd . Days (from 0 to 999) [optional]
hh : Hours (from 0 to 23)
mm : Minutes (from 0 to 59)
ss Seconds (from 0 to 59)
```

Setting Name: `AdvanceWarningFrequency`

Setting Value: `{Time span in format of ddd.hh:mm:ss where ddd. is optional}`

Example Setting Value: `01:00:00`

### Advance warning message box body text
Description: 
```
Text displayed in the message box that alerts users in advance of an upcoming maintenance
```

Setting Name: `AdvanceWarningMessageBody`

Setting Value: `{Message to display in advance of an upcoming maintenance}`

Example Setting Value: `{TIMESTAMP}Please save your work. The system will go offline for maintenance in {TIMELEFT}`

### Advance warning message box title
Description: 
```
Window caption of the message box that alerts users in advance of an upcoming maintenance
```

Setting Name: `AdvanceWarningMessageTitle`

Setting Value: `{Title to display in advance of an upcoming maintenance}`

Example Setting Value: `Upcoming Maintenance`

Constraint:
```
The text of the message body must be supplied and must be 3072 characters or less.
```

### Advance warning time period
Description: 
```
For Manual and PVS Server OS catalogs, the Connector Agent service for Configuration Manager 2012 alerts users in advance when there are pending application installs and/or software updates that will take place in the near future so that they can plan their use of the system accordingly. For MCS catalogs, use Studio to send messages to users.

This setting determines how far in advance users are to be notified.

A time span string consists of the following:

ddd . Days (from 0 to 999) [optional]
hh : Hours (from 0 to 23)
mm : Minutes (from 0 to 59)
ss Seconds (from 0 to 59)
```

Setting Name: `AdvanceWarningPeriod`

Setting Value: `{Time span in format of ddd.hh:mm:ss where ddd. is optional}`

Example Setting Value: `16:00:00`

### Allow applications to use the physical location of the client device
Description: 
```
Enables or disables the ability for applications to use the physical location of the client device. By default, the ability to use the physical location of the client device is disabled.
```

Setting Name: `AllowLocationServices`

Setting Value:
```
{
    enabled = true | false
}
```

### Allow bidirectional content redirection
Description: 
```
When enabled, allows URLs to be redirected from client to server (or vice versa) based on the rules specified in 'Allowed URLs to be redirected to VDA' and 'Allowed URLs to be redirected to Client'.
```

Related Settings:
```
Allowed URLs to be redirected to VDA

Allowed URLs to be redirected to Client
```

Setting Name: `AllowBidirectionalContentRedirection`

Setting Value:
```
{
    enabled = true | false
}
```

### Allow existing USB devices to be automatically connected
Description: 
```
When enabled, permits eligible USB devices connected to the endpoint at the start of a session to be automatically redirected to the remote session.
```

Setting Name: `UsbConnectExistingDevices`

Setting Value: `Never`, `Ask`, `Always`

Setting Value Explanation :
Setting Value | Explanation
-- | --
`Never` | Do not automatically redirect available USB devices.
`Ask` | Ask before redirecting available USB devices.
`Always` | Automatically redirect available USB devices.

### Allow local app access
Description: 
```
Local App Access enables the integration of users' locally-installed applications and hosted applications within a hosted desktop environment. This allows users to access in one place all of the applications they need.

When a user then launches the locally-installed application using the shortcut, the application appears to be running on their virtual desktop, even though it is running on the local device.
```

Related Settings:
```
URL redirection black list

URL redirection white list
```

Setting Name: `AllowLocalAppAccess`

Setting Value:
```
{
    enabled = true | false
}
```

### Allow newly arrived USB devices to be automatically connected
Description: 
```
When enabled, permits eligible USB devices inserted at the endpoint during a session to be automatically redirected to the remote session.
```

Setting Name: `UsbConnectNewDevices`

Setting Value: `Never`, `Ask`, `Always`

Setting Value Explanation:
Setting Value | Explanation
-- | --
`Never` | Do not automatically redirect available USB devices.
`Ask` | Ask before redirecting available USB devices.
`Always` | Automatically redirect available USB devices.

### Allow visually lossless compression
Description: 
```
This setting allows visually lossless compression to be used instead of true lossless compression for graphics. Visually lossless improves performance over true lossless, but has minor loss that's unnoticeable by sight. This setting changes the way the values of the 'Visual quality' setting are used. For more information on how this setting affects the use of 'Visual quality', see the help for that setting.
```

Related Settings:
```
Video quality
```

Setting Name: `AllowVisuallyLosslessCompression`

Setting Value:
```
{
    enabled = true | false
}
```

### Allowed URLs to be redirected to Client
Description: 
```
Specifies the list of URLs to open on Client.
A semi-colon (;) is the delimter. Double quotes (") can be used. An asterisk (*) can be used as a wild card. For example *.xyz.com.
Note: To configure 'Allowed URLs to be redirected to Client', you must first enable 'Allow Bidirectional Content Redirection'.
```

Related Settings:
```
Allow bidirectional content redirection, Allowed URLs to be redirected to VDA
```

Setting Name: `ClientURLs`

Setting Value: 
```
jsonencode([
    {Client URL1},
    {Client URL2},
    ...
])
```

### Allowed URLs to be redirected to VDA
Description: 
```
Specifies the list of URLs to open on the VDA.
A semi-colon (;) is the delimiter. Double quotes (") can be used. An asterisk (*) can be used as a wild card. For example *.xyz.com.
Note: To configure 'Allowed URLs to be redirected to VDA', you must first enable 'Allow Bidirectional Content Redirection'.
```

Related settings:
```
Allow bidirectional content redirection, Allowed URLs to be redirected to Client
```

Setting Name: `VDAURLs`

Setting Value: 
```
jsonencode([
    {VDA URL1},
    {VDA URL2},
    ...
])
```

### Always cache
Description: 
```
Optionally, to enhance the user experience, use this setting with the Profile streaming setting, which imposes a lower limit on the size of files that are streamed. Any files this size or larger are cached as soon as possible after logon. To use the cache entire profile feature, set this limit to zero (which caches all of the profile contents).
```

Setting Name: `PSAlwaysCache`

Setting Value:
```
{
    enabled = true | false
}
```

### Always cache size
Description: 
```
Optionally, to enhance the user experience, use this setting with the Profile streaming setting, which imposes a lower limit on the size of files that are streamed. Any files this size or larger are cached as soon as possible after logon. To use the cache entire profile feature, set this limit to zero (which caches all of the profile contents).
```

Setting Name: `PSAlwaysCache_Part`

Setting Value: `{Limit in MB}`

### App access control
Description: 
```
Controls user access to applications based on your rules. Enter the rules created using the PowerShell script CPM_App_Access_Control_Config.ps1.

The script is available with the Profile Management installation package. For more information, see the Profile Management documentation.
```

Setting Name: `AppAccessControl_Part`

Setting Value: `{Name of the Rule}`

### AppData(Roaming) path
Description: 
```
Lets you specify how to redirect the AppData(Roaming) folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRAppDataPath_Part`

Setting Value: `{Path to the AppData Roaming directory}`

Example Setting Value: `C:\\Users\\User1\\AppData\\Roaming`

### Application launch wait timeout
Description: 
```
A session can end before the first application starts. The issue might occur when a one minute time-out is exceeded. For example, when the profile share is located across a WAN link rather than on a local share.
Use this setting to specify the timeout (in milliseconds) for the session to wait before ending.
```

Setting Name: `ApplicationLaunchWaitTimeout`

Setting Value: `{Timeout in milliseconds}`

### Audio over UDP
Description: 
```
Allow Audio over UDP on the server. The server will open a UDP port, to support all connections which are configured to use Audio over UDP Real-Time Transport. When set to Prohibited, the UDP port will not be opened.
```

Related settings:
```
Audio UDP port range
Audio over UDP real-time transport
```

Setting Name: `UDPAudioOnServer`

Setting Value:
```
{
    enabled = true | false
}
```

### Audio over UDP real-time transport
Description: 
```
Allows transmission of audio between host and client over Real-time Transport Protocol (RTP) using the User Datagram Protocol (UDP). UDP is best suited for real-time Voice over Internet Protocol activities. UDP transport avoids the lag that can occur with TCP when there is network congestion or packet loss providing real time delivery of the audio data. To use UDP, the audio quality must be set to 'Medium - optimized for speech.'
```

Setting Name: `AllowRtpAudio`

Setting Value:
```
{
    enabled = true | false
}
```

### Audio Plug N Play
Description: 
```
Allows the use of multiple audio devices.
```

Setting Name: `AudioPlugNPlay`

Setting Value:
```
{
    enabled = true | false
}
```

### Audio quality
Description: 
```
Specify the sound quality as low, medium, or high.

Choose 'Low - for low-speed connections' for low-bandwidth connections. Sounds sent to the client are compressed up to 16Kbps. This compression results in a significant decrease in the quality of the sound but allows reasonable performance for a low-bandwidth connection.

Choose 'Medium - optimized for speech' for delivering Voice over IP applications. Audio sent to the client is compressed up to 64Kbps. This compression results in a moderate decrease in the quality of the audio played on the client device, but provides low latency and consumes very low bandwidth. Real-time Transport (RTP) over User Datagram Protocol (UDP) is only supported when this audio quality is selected. Use this audio quality even for delivering media applications over challenging network connections, like very low (less than 512Kbps) lines and when there is congestion and packet loss in the network.

Choose 'High - high definition audio' when delivering media applications. This setting provides high fidelity stereo audio but consumes more bandwidth than the Medium quality setting. Use this setting when network bandwidth is plentiful and sound quality is important.

Bandwidth is consumed only when audio is recording or playing. When both occur at the same time, the bandwidth consumption doubles.
```

Related Settings:
```
Audio redirection bandwidth limit

Audio redirection bandwidth limit percent
```

Setting Name: `AudioQuality`

Setting Value: `High`, `Medium`, `Low`

Setting Value Explanation:
Setting Value | Explanation
-- | --
`High` | High Definition Audio
`Medium` | Optimized for Speech
`Low` | For Low-Speed Connections

### Audio redirection bandwidth limit
Description: 
```
Specifies the maximum allowed bandwidth in kilobits per second for playing or recording audio in a client session.

If you enter a value for this setting and a value for the 'Audio redirection bandwidth limit percent' setting, the most restrictive setting (with the lower value) is applied.
```

Related Settings:
```
Audio redirection bandwidth limit percent

Client audio redirection
```

Setting Name: `AudioBandwidthLimit`

Setting Value: `{Bandwidth Limit in Kbps}`

### Audio redirection bandwidth limit percent
Description: 
```
Specifies the maximum allowed bandwidth limit for playing or recording audio as a percent of the total session bandwidth.

If you enter a value for this setting and a value for the 'Audio redirection bandwidth limit' setting, the most restrictive setting (with the lower value) is applied.

If you configure this setting, you must also configure the 'Overall session bandwidth limit' setting which specifies the total amount of bandwidth available for client sessions.
```

Related settings:
```
Audio redirection bandwidth limit

Overall session bandwidth limit

Client audio redirection
```

Setting Name: `AudioBandwidthPercent`

Setting Value: `{Percent number of Overall session bandwidth limit allowed for audio}`

### Audio UDP port range
Description: 
```
Range of UDP ports that will be used by the host to exchange audio packet data over RTP with the client. The host will attempt to use a UDP port pair starting from the lowest number port first, incrementing by 2 until it attempts the highest port. Each port will handle both inbound and outbound traffic.

Enter a range in the format (Low port),(High port).

Use client GPO to configure client ports.
```

Setting Name: `RtpAudioPortRange`

Setting Value: `{Port Range in format (Low port),(High port)}`

Example Setting Value: `16500, 16509`

### Auto client reconnect
Description: 
```
Allows or prevents automatic reconnection by the same client after a connection has been interrupted.

Allowing automatic reconnection allows users to resume working where they were interrupted when a connection was broken. Automatic reconnection detects broken connections and then reconnects the users to their sessions.

However, automatic reconnection can result in a new session being launched (instead of reconnecting to an existing session) if a plug-in's cookie, containing the key to the session ID and credentials, is not used. The cookie is not used if it has expired, for example, because of a delay in reconnection, or if credentials must be reentered. Auto Client Reconnect is not triggered if users intentionally disconnect.
```

Related Settings:
```
Auto client reconnect authentication
```

Setting Name: `AutoClientReconnect`

Setting Value:
```
{
    enabled = true | false
}
```

### Auto client reconnect authentication
Description: 
```
Requires authentication for automatic client reconnections.

When a user initially logs on to a farm, XenDesktop encrypts and stores the user credentials in memory and creates a cookie containing the encryption key which is sent to the plug-in. When this setting is added, cookies are not used. Instead, XenDesktop displays a dialog box to users requesting credentials when the plug-in attempts to reconnect automatically.
```

Related Settings:
```
Auto client reconnect
```

Setting Name: `AutoClientReconnectAuthenticationRequired`

Setting Value: `RequireAuthentication`, `DoNotRequireAuthentication`

### Auto client reconnect logging
Description: 
```
Records or prevents recording auto client reconnections in the event log. By default, logging is disabled.

When logging is enabled, the server's System log captures information about successful and failed automatic reconnection events. The server farm does not provide a combined log of reconnection events for all servers.
```

Setting Name: `AutoClientReconnectLogging`

Setting Value: `DoNotLogAutoReconnectEvents`, `LogAutoReconnectEvents`

### Auto client reconnect timeout
Description: 
```
Auto client reconnect timeout allows admins to configure the time for which the Receiver will attempt the reconnection through ACR. If Session reliability is also configured with a certain timeout, then the ACR will trigger after the session reliability timeout is over. The maximum allowed ACR Timeout value is 5 minutes.
```

Related settings:
```
Auto client reconnect
```

Setting Name: `ACRTimeout`

Setting Value: `{Timeout Value in Seconds}`

### Auto connect client COM ports
Description: 
```
For VDA versions that do not support this setting, follow CTX139345 to enable redirection using the registry.

When enabled, COM ports from the client are automatically connected. When disabled, COM ports from the client are not automatically connected. By default, COM ports are not automatically connected.
```

Related Settings:
```
Client COM port redirection
```

Setting Name: `ClientComPortsAutoConnection`

Setting Value:
```
{
    enabled = true | false
}
```

### Auto connect client drives
Description: 
```
Allows or prevents automatic connection of client drives when users log on. By default, automatic connection is allowed. When allowing this setting, make sure to enable the settings for the drive types you want automatically connected. For example, configure the 'Client optical drives' setting to allow automatic connection of CD-ROM drives on the client device.
```

Related Settings:
```
Client drive redirection

Client floppy drives

Client optical drives

Client fixed drives

Client network drives

Client removable drives
```

Setting Name: `AutoConnectDrives`

Setting Value:
```
{
    enabled = true | false
}
```

### Auto connect client LPT ports
Description: 
```
For VDA versions that do not support this setting, follow CTX139345 to enable redirection using the registry.

When enabled, LPT ports from the client are automatically connected. When disabled, LPT ports from the client are not automatically connected. By default, LPT ports are not automatically connected.
```

Related Settings:
```
Client LPT port redirection
```

Setting Name: `ClientLptPortsAutoConnection`

Setting Value:
```
{
    enabled = true | false
}
```

### Auto-create client printers
Description: 
```
Specifies which client printers are auto-created. This setting overrides default client printer auto-creation settings. By default, all client printers are auto-created.

This setting applies only if the 'Client printer redirection' setting is enabled.
'Auto-create all client printers' creates all printers on the client device.
'Do not auto-create client printers' turns off printer auto-creation when users log on.
'Auto-create the client's default printer only' automatically creates the printer selected as the client's default printer.
'Auto-create local (non-network) client printers only' automatically creates only printers directly connected to the client device through LPT, COM, USB, or other local port.
```

Setting Name: `ClientPrinterAutoCreation`

Setting Value: `DoNotAutoCreate`, `DefaultPrinterOnly`, `LocalPrintersOnly`, `AllPrinters`

### Auto-create generic universal printer
Description: 
```
Enables or disables auto-creation of the Citrix UNIVERSAL Printer generic printing object for sessions with a UPD capable client. By default, the generic universal printer is not auto-created.
```

Setting Name: `GenericUniversalPrinterAutoCreation`

Setting Value:
```
{
    enabled = true | false
}
```

### Auto-create PDF Universal Printer
Description: 
```
Enables or disables auto-creation of the Citrix PDF Printer for sessions using Citrix Receiver for Windows (starting from VDA 7.19), Citrix Receiver for HTML5 or Citrix Receiver for Chrome. By default, the Citrix PDF Printer is not auto-created.

When the Auto-Create Generic Universal Printer policy setting is enabled and the session is using Citrix Receiver for Windows, the Generic Universal Printer will be created in the session but the Citrix PDF Printer will not be created in the session.

The Citrix PDF Printing Feature Pack is required for Citrix Receiver for HTML5 and Citrix Receiver for Chrome.
```

Setting Name: `AutoCreatePDFPrinter`

Setting Value:
```
{
    enabled = true | false
}
```

### Automatic keyboard display
Description: 
```
Enables or disables the automatic display of the soft keyboard on mobile devices. By default, the automatic keyboard display is disabled.
```

Setting Name: `AutoKeyboardPopUp`

Setting Value:
```
{
    enabled = true | false
}
```

### Automatic migration of existing application profiles
Description: 
```
This setting enables or disables the automatic migration of existing application profiles across different operating systems. The application profiles include both the application data in the AppData folder and the registry entries under HKEY_CURRENT_USER\SOFTWARE. This setting can be useful in cases where you want to migrate your application profiles across different operating systems.

For example, suppose you upgrade your operating system (OS) from Windows 10 version 1803 to Windows 10 version 1809. If this setting is enabled, Profile Management automatically migrates the existing application settings to Windows 10 version 1809 the first time each user logs on. As a result, the application data in the AppData folder and the registry entries under HKEY_CURRENT_USER\SOFTWARE are migrated.

If there are multiple existing application profiles, Profile Management performs the migration in the following order of priority:
1. Migrates profiles of the same OS type (Desktop OS to Desktop OS and Server OS to Server OS).
2. Migrates profiles of the same Windows OS family; for example, Windows 8 to Windows 8.1, or Windows 10 version 1803 to Windows 10 version 1809.
3. Migrates profiles of an earlier version of the OS; for example, Windows 7 to Windows 10, or Windows Server 2012 to Windows Server 2016.
4. Migrates from the profiles of the closest OS.

Note:You must specify the short name of the operating system by including the variable "!CTX_OSNAME!" in the user store path so that Profile Management can locate the existing application profiles.

If this setting is not configured here, the setting from the .ini file is used.
If this setting is not configured here or in the .ini file, it is disabled by default.
```

Setting Name: `ApplicationProfilesAutoMigration`

Setting Value:
```
{
    enabled = true | false
}
```

### Automatically reattach VHDX disks in sessions
Description: 
```
This policy applies only to VHDX disks that Profile Management uses.

With the policy enabled, when a VHDX disk is detached in a session, Profile Management can detect the event and then reattach the disk automatically.
```

Setting Name: `EnableVolumeReattach`

Setting Value:
```
{
    enabled = true | false
}
```

### Browser content redirection
Description: 
```
Controls and optimizes the way Citrix Workspace app delivers web browser content (like HTML5) to users. Only the visible area of the browser where content is displayed (also known as the viewport) is redirected.

By default, this setting is Allowed, and Citrix Workspace app attempts client fetch and client render. If client fetch and client render fails, server-side rendering is attempted. If the browser content redirection proxy configuration setting is also enabled, Citrix Workspace app attempts only server fetch and client render.

System Requirements:
Citrix Receiver for Windows 4.10 minimum to support Internet Explorer 11 viewport redirection.
Citrix Workspace app 1809 for Windows minimum to support Internet Explorer 11 viewport redirection and Google Chrome viewport redirection.
Citrix HDXJsInjector add-on must be enabled on Internet Explorer 11 on the VDA.

The browser content redirection extension (from the Chrome Web Store) must be added and enabled on Google Chrome on the VDA.
```

Related Settings:
```
Browser content redirection ACL configuration

Browser content redirection proxy configuration

Browser content redirection block list configuration

Browser content redirection authentication sites
```

Setting Name: `WebBrowserRedirection`

Setting Value:
```
{
    enabled = true | false
}
```

### Browser content redirection ACL configuration
Description: 
```
This setting allows you to configure an Access Control List (ACL) of URLs that can use browser content redirection.

Authorized URLs: Specifies the allowed URLs whose content will be redirected to the client.
Wildcard '*' is permitted, however wildcard '*' is not permitted within the protocol of the URL.
For example, http://www.xyz.com/index.html, https://www.xyz.com/*, http://www.xyz.com/*videos*, https://*.xyz.com/*, and https://*.*.xyz.com/* are allowed.
However, http*://www.xyz.com/ is not allowed.
Wildcard '*' after the domain of the URL matches all characters.
Wildcard '*' in the domain of the URL only matches a single domain.
For example, http://*.xyz.com/* matches http://www.xyz.com/index.html but not http://www.abc.xyz.com/index.html or http://www.othersite.com/index.html?abc=www.xyz.com/index.html

Better granularity can be achieved by specifying paths in the URL, for example https://www.xyz.com/sports/index.html. In this case, only index.html page will be redirected.

By default, this setting is set to "https://www.youtube.com/*"
```

Related Settings:
```
Browser content redirection

Browser content redirection proxy configuration

Browser content redirection block list configuration

Browser content redirection authentication sites
```

Setting Name: `WebBrowserRedirectionAcl`

Setting Value: 
```
jsonencode([
    {URL Rule1},
    {URL Rule2},
    ...
])
```

Example Setting Value:
```
jsonencode([
    "https://www.youtube.com/*",
    "https://some.website.com/index.html",
    ...
])
```

### Browser content redirection authentication sites
Description: 
```
This setting allows you to configure a list of URLs that sites redirected via browser content redirection can use to authenticate a user.

Specifies the URLs for which browser content redirection will remain active (redirected) when navigating from an allowed URL.
Wildcard '*' is permitted, however wildcard '*' is not permitted within the protocol of the URL.
For example, http://www.xyz.com/index.html, https://www.xyz.com/*, http://www.xyz.com/*videos*, https://*.xyz.com/*, and https://*.*.xyz.com/* are allowed.
However, http*://www.xyz.com/ is not allowed.
Wildcard '*' after the domain of the URL matches all characters.
Wildcard '*' in the domain of the URL only matches a single domain.
For example, http://*.xyz.com/* matches http://www.xyz.com/index.html but not http://www.abc.xyz.com/index.html or http://www.othersite.com/index.html?abc=www.xyz.com/index.html

By default, this list is empty.
```

Related Settings:
```
Browser content redirection

Browser content redirection proxy configuration

Browser content redirection ACL configuration

Browser content redirection block list configuration
```

Setting Name: `WebBrowserRedirectionAuthenticationSites`

Setting Value: 
```
jsonencode([
    {URL Rule1},
    {URL Rule2},
    ...
])
```

Example Setting Value:
```
jsonencode([
    "https://www.youtube.com/*",
    "https://some.website.com/index.html",
    ...
])
```

### Browser content redirection block list configuration
Description: 
```
This setting allows you to configure a block list of URLs that cannot use the browser content redirection feature. If a URL is configured via this setting, then the browser content of that URL will not be redirected, but rendered on the server.

This setting works in conjunction with the browser content redirection ACL configuration setting. If URLs are present in both browser content redirection ACL configuration setting and the above block list configuration setting, the block list configuration takes precedence and the browser content of the URL will not be redirected.

Unauthorized URLs: Specifies the blocked URLs whose browser content will not be redirected to the client, but rendered on the server.
Wildcard '*' is permitted, however wildcard '*' is not permitted within the protocol of the URL.
For example, http://www.xyz.com/index.html, https://www.xyz.com/*, http://www.xyz.com/*videos*, https://*.xyz.com/*, and https://*.*.xyz.com/* are allowed.
However, http*://www.xyz.com/ is not allowed.
Wildcard '*' after the domain of the URL matches all characters.
Wildcard '*' in the domain of the URL only matches a single domain.
For example, http://*.xyz.com/* matches http://www.xyz.com/index.html but not http://www.abc.xyz.com/index.html or http://www.othersite.com/index.html?abc=www.xyz.com/index.html

By default, this list is empty.
```

Related Settings:
```
Browser content redirection

Browser content redirection proxy configuration

Browser content redirection ACL configuration

Browser content redirection authentication sites
```

Setting Name: `WebBrowserRedirectionBlacklist`

Setting Value: 
```
jsonencode([
    {URL Rule1},
    {URL Rule2},
    ...
])
```

Example Setting Value:
```
jsonencode([
    "https://www.youtube.com/*",
    "https://some.website.com/index.html",
    ...
])
```

### Browser content redirection integrated Windows authentication support
Description: 
```
Controls browser content redirection (BCR) overlay to use the Negotiate Auth scheme for authentication by using the VDA user's credentials.

When Allowed, the BCR components on the client and on the VDA enable the Overlay to request and obtain a Negotiate ticket from the VDA, which the Overlay provides in response to the authentication challenge.

When Prohibited, the BCR Overlay does not obtain a Negotiate ticket from the VDA.

By default, this setting is Prohibited.

Configuration Requirement:

The Kerberos infrastructure must be configured to issue Kerberos tickets for SPNs (Service Principal Names) constructed from that hostname (e.g. "HTTP/serverhostname.com").

Server-Fetch: When BCR is configured in server-fetch mode, an administrator must ensure that DNS is configured properly on the VDA.

Client-Fetch: When BCR is configured in client-fetch mode, the BCR overlay (running on the client) relies on DNS to resolve hostnames into an IP addresses reachable by the overlay.
```

Related Settings:
```
Browser content redirection

Browser content redirection ACL configuration

Browser content redirection proxy configuration

Browser content redirection server fetch proxy auth

Browser content redirection block list configuration

Browser content redirection authentication sites
```

Setting Name: `WebBrowserRedirectionIwaSupport`

Setting Value:
```
{
    enabled = true | false
}
```

### Browser content redirection proxy configuration
Description: 
```
This setting provides configuration options for proxy settings on the VDA for the browser content redirection feature.
If enabled with a valid proxy address:port number, PAC file wpad URL, or Direct/Transparent setting, only server fetch client render is attempted.

If disabled or left unconfigured with default value, client fetch client rendering is attempted.

To configure Server Fetch Client Render, three choices are available once Enabled is selected:

1. Direct or Transparent
Browser content redirection traffic will be routed through the VDA and forwarded directly to the web server hosting the content (e.g. www.youtube.com). The text box should be filled with the keyword "DIRECT" if direct forwarding to a web server is desired.

This setting is designed for scenarios where a Transparent proxy is configured in your environment, or you simply do not have any proxy at all.

2. Explicit Proxy
Browser content redirection traffic will be routed through the VDA and forwarded to the web proxy URL that has been specified in the text box.

Allowed pattern: http://<hostname/ip address>:<port>
For example, http://proxy.example.citrix.com:80 or http://10.0.0.1:8080

3. PAC Files
Browser content redirection traffic will be routed through the VDA and forwarded to the web proxy determined by evaluating the PAC file provided at the URL that has been specified in the text box.
Allowed pattern: http://<hostname/ip address>:<port>/<path>/<Proxy.pac>
For example, http://proxy.example.citrix.com:80/configuration/pac/Proxy.pac
Allowed pattern: http://<hostname/ip address>:<port>/<path>/<wpad.dat>
For example, http://proxy.example.citrix.com:80/configuration/pac/wpad.dat

By default, this setting is prohibited.
```

Related Settings:
```
Browser content redirection

Browser content redirection ACL configuration

Browser content redirection block list configuration

Browser content redirection authentication sites
```

Setting Name: `WebBrowserRedirectionProxy`

Setting Value: `{Proxy Server Address}`

Example Values:
```
http://proxy.example.citrix.com:80/configuration/pac/Proxy.pac
```

```
DIRECT
```

```
http://proxy.example.citrix.com:80
```

### Browser content redirection server fetch proxy auth
Description: 
```
Controls browser content redirection(BCR) so that HTTP traffic originating at the overlay can be routed through web proxies configured to authorize HTTP traffic authenticated with the VDA user's domain credentials via the HTTP “Negotiate” auth scheme.

BCR must be configured for server-fetch in PAC File mode through the browser content redirection proxy configuration setting. The PAC script must provide instructions to route the overlay's traffic through a downstream web proxy. The downstream web proxy must be configured to require authentication of VDA users via the “Negotiate” auth scheme.

When Allowed, if the web proxy responds with a 407 Negotiate challenge (i.e. contains a "Proxy-Authenticate: Negotiate" header), then BCR obtains a Kerberos service ticket by using the VDA user's domain credentials, and includes it in the subsequent request to the web proxy.

When Prohibited, BCR proxies all TCP traffic between the overlay and the web proxy without interfering. Thus, the overlay uses whatever credentials are available to it in order to authenticate to the web proxy (e.g. including basic auth credentials).

By default, this setting is Prohibited.
```

Related Settings:
```
Browser content redirection

Browser content redirection ACL configuration

Browser content redirection proxy configuration

Browser content redirection block list configuration

Browser content redirection authentication sites
```

Setting Name: `WebBrowserRedirectionProxyAuth`

Setting Value:
```
{
    enabled = true | false
}
```

### Citrix Cloud Connectors
Description: 
```
This setting lets you configure Citrix Cloud Connector machines to which the Workspace Environment Management agent can connect. To ensure high availability, we recommend at least two Cloud Connectors in each resource location. Make sure that Cloud Connector machines and agent machines reside in the same AD domain.
```

Setting Name: `WemCloudConnectorList`

Setting Value: 
```
jsonencode([
    {Connector 1 URL},
    {Connector 2 URL},
    ...
])
```

Example Setting Value:
```
jsonencode([
    "https://connector1.citrix.com",
    "https://connector2.citrix.com"
])
```

### Client audio redirection
Description: 
```
Allows or prevents applications hosted on the server to play sounds through a sound device installed on the client computer. Also allows or prevents users to record audio input.

After allowing this setting, you can limit the bandwidth consumed by playing or recording audio. Limiting the amount of bandwidth consumed by audio can improve application performance but may also degrade audio quality. Bandwidth is consumed only while audio is recording or playing. If both occur at the same time, the bandwidth consumption doubles.

To specify the maximum amount of bandwidth, configure the 'Audio redirection bandwidth limit' setting or the 'Audio redirection bandwidth limit percent' setting.
```

Setting Name: `ClientAudioRedirection`

Related Settings:
```
Audio redirection bandwidth limit

Audio redirection bandwidth limit percent

Client microphone redirection
```

Setting Value:
```
{
    enabled = true | false
}
```

### Client clipboard redirection
Description: 
```
Allow or prevent the clipboard on the client device to be mapped to the clipboard on the server. By default, clipboard redirection is allowed.

To prevent cut-and-paste data transfer between a session and the local clipboard, select 'Prohibited'. Users can still cut and paste data between applications running in a session.

After allowing this setting, configure the maximum allowed bandwidth the clipboard can consume in a client connection using the Clipboard redirection bandwidth limit setting or the Bandwidth limit for clipboard redirection channel as percent of total session bandwidth setting.
```

Setting Name: `ClipboardRedirection`

Related Settings:
```
Clipboard redirection bandwidth limit

Clipboard redirection bandwidth limit percent
```

Setting Value:
```
{
    enabled = true | false
}
```

### Client clipboard write allowed formats
Description: 
```
This setting doesn't apply if 'Client clipboard redirection' is set to Prohibited or 'Restrict client clipboard write' is not Enabled.

When the setting 'Restrict client clipboard write' is set to Enabled, host clipboard data cannot be shared with client endpoint but this setting can be used to selectively allow specific data formats to be shared with client endpoint clipboard. Administrator can enable this setting and add specific formats to be allowed.

The following are system defined clipboard formats.
CF_TEXT
CF_BITMAP
CF_METAFILEPICT
CF_SYLK
CF_DIF
CF_TIFF
CF_OEMTEXT
CF_DIB
CF_PALETTE
CF_PENDATA
CF_RIFF
CF_WAVE
CF_UNICODETEXT
CF_LOCALE
CF_DIBV5
CF_DSPTEXT
CF_DSPBITMAP
CF_DSPMETAFILEPICT
CF_DSPENHMETAFILE
CF_HTML

The followings are custom formats predefined in XenApp and XenDesktop.
CFX_RICHTEXT
CFX_OfficeDrawingShape
CFX_BIFF8
CFX_FILE

Additional custom formats can be added. Actual name for custom formats must match the formats to be registered with system. Format names are case insensitive.
```

Setting Name: `ClientClipboardWriteAllowedFormats`

Related Settings:
```
Client clipboard redirection

Restrict client clipboard write

Restrict session clipboard write

Session clipboard write allowed formats
```

Setting Value: 
```
jsonencode([
    {FORMAT_1},
    {FORMAT_2},
    ...
])
```

Example Setting Value: 
```
jsonencode([
    "CF_TEXT",
    "CFX_FILE",
    ...
])
```

### Client COM port redirection
Description: 
```
For VDA versions that do not support this setting, follow CTX139345 to enable redirection using the registry.

When enabled, COM port redirection to and from the client is allowed. When disabled, COM port redirection to and from the client is not allowed. By default COM port redirection is disabled.
```

Setting Name: `ClientComPortRedirection`

Related Settings: 
```
Auto connect client COM ports
```

Setting Value:
```
{
    enabled = true | false
}
```

### Client drive redirection
Description: 
```
Enables or disables file (drive) redirection to and from the client. When enabled, users can save files to all their client drives. When disabled, all file redirection is prevented, regardless of the state of the individual file redirection settings such as 'Client floppy drives' and 'Client network drives.' By default, file redirection is enabled.
```

Setting Name: `ClientDriveRedirection`

Related Settings:
```
Client floppy drives

Client optical drives

Client fixed drives

Client network drives

Client removable drives
```

Setting Value:
```
{
    enabled = true | false
}
```

### Client fixed drives
Description: 
```
Allows or prevents users from accessing or saving files to fixed drives on the user device. By default, accessing client fixed drives is allowed.

When enabling this setting, make sure the 'Client drive redirection' setting is enabled. Additionally, to ensure fixed drives are automatically connected, configure the 'Auto connect client drives' setting. If these settings are disabled, client fixed drives are not mapped and users cannot access these drives manually, regardless of the state of the 'Client fixed drives' setting.
```

Setting Name: `ClientFixedDrives`

Related Settings:
```
Client drive redirection

Auto connect client drives
```

Setting Value:
```
{
    enabled = true | false
}
```

### Client floppy drives
Description: 
```
Allows or prevents users from accessing or saving files to floppy drives on the client device. By default, accessing client floppy drives is allowed.

When enabling this setting, make sure the 'Client drive redirection' setting is enabled. Additionally, to ensure floppy drives are automatically connected, configure the 'Auto connect client drives' setting. If these settings are disabled, client floppy drives are not mapped and users cannot access these drives manually, regardless of the state of the 'Client floppy drives' setting.
```

Setting Name: `ClientFloppyDrives`

Related Settings:
```
Client drive redirection

Auto connect client drives
```

Setting Value:
```
{
    enabled = true | false
}
```

### Client keyboard layout synchronization and IME improvement
Description: 
```
Allows a user to change the client keyboard layout and synchronize it to the VDA side dynamically in a session without relogging on or reconnecting. For Chinese, Korean, and Japanese users, who require client keyboard synchronization to use IME improvement, allows them to use client IME improvement for best user experience.
For the Windows VDA, setting “Support dynamic client keyboard synchronization” equals “Support dynamic client keyboard synchronization and IME improvement.” For Linux VDA, they are two different settings.
When not configured, the default is “Disabled” in Windows Server 2016 and Windows Server 2019. The default is “Support dynamic client keyboard synchronization and IME improvement” in Windows Server 2012 and Window 10 for consistency with the previous release.
Note:
For non-Windows Citrix Workspace app, such as Citrix Workspace app for Mac, enable Unicode keyboard layout mapping to ensure correct key mapping.
We are phasing out the registry setting in the Windows VDA -
HKEY_LOCAL_MACHINE\Software\Citrix\ICA\lcalme\DisableKeyboardSync value = DWORD 0
```

Setting Name: `ClientKeyboardLayoutSyncAndIME`

Related Settings:
```
Enable Unicode keyboard layout mapping

Hide keyboard layout switch pop-up message box
```

Setting Value: `NotSupport`, `ClientKeyboardLayoutSync`, `ClientKeyboardLayoutSyncAndIME`

Setting Value Explanation:
Setting Value | Explanation
--|--
`NotSupport` | Disabled
`ClientKeyboardLayoutSync` | Support dynamic client keyboard layout synchronization
`ClientKeyboardLayoutSyncAndIME` | Support dynamic client keyboard layout synchronization and IME improvement

### Client LPT port redirection
Description: 
```
For VDA versions that do not support this setting, follow CTX139345 to enable redirection using the registry.

When enabled, LPT port redirection to the client is allowed. When disabled, LPT port redirection to the client is not allowed. By default, LPT port redirection is disabled.

This rule is necessary only for servers that host legacy applications that print to LPT ports; most applications today can send print jobs to printer objects.
```

Setting Name: `ClientLptPortRedirection`

Setting Value:
```
{
    enabled = true | false
}
```

### Client microphone redirection
Description: 
```
Enables or disables client microphone redirection. When enabled, clients can use microphones to record audio input.

For security, users are alerted when servers that are not trusted by their client devices try to access microphones. Users can choose to accept or not accept access. Users can disable the alert on Citrix Receiver.

If audio is disabled on the client device, this rule has no effect.
```

Setting Name: `MicrophoneRedirection`

Related Settings:
```
Audio redirection bandwidth limit

Audio redirection bandwidth limit percent
```

Setting Value:
```
{
    enabled = true | false
}
```

### Client network drives
Description: 
```
Allows or prevents users from accessing and saving files to client network (remote) drives. By default, accessing client network drives is allowed.

When allowing this setting, make sure the 'Client drive redirection' setting is enabled. Additionally, to ensure network drives are automatically connected, configure the 'Auto connect client drives' setting. If these settings are disabled, client network drives are not mapped and users cannot access these drives manually, regardless of the state of the 'Client network drives' setting.
```

Setting Name: `ClientNetworkDrives`

Related Settings:
```
Client drive redirection

Auto connect client drives
```

Setting Value:
```
{
    enabled = true | false
}
```

### Client optical drives
Description: 
```
Allows or prevents users from accessing or saving files to CD-ROM, DVD-ROM, and BD-ROM drives on the client device. By default, accessing client optical drives is allowed.

When enabling this setting, make sure the 'Client drive redirection' setting is enabled. Additionally, to ensure CD-ROM drives are automatically connected, configure the 'Auto connect client drives' setting. If these settings are disabled, client optical drives are not mapped and users cannot access these drives manually, regardless of the state of the 'Client optical drives' setting.
```

Setting Name: `ClientOpticalDrives`

Related Settings:
```
Client drive redirection

Auto connect client drives
```

Setting Value:
```
{
    enabled = true | false
}
```

### Client printer names
Description: 
```
Selects the naming convention for auto-created client printers. By default, standard printer names are used.

For most configurations, select 'Standard printer names' which are similar to those created by native Terminal Services, such as 'HPLaserJet 4 from clientname in session 3'.

Select 'Legacy printer names' to use old-style client printer names. An example of a legacy printer name is 'Client/clientname#/HPLaserJet 4'.
```

Setting Name: `ClientPrinterNames`

Setting Value: `StandardPrinterNames`, `LegacyPrinterNames`

### Client printer redirection
Description: 
```
Allows or prevents client printers to be mapped to a server when a user logs on to a session. By default, client printer mapping is allowed.
```

Setting Name: `ClientPrinterRedirection`

Related Settings:
```
Auto-create client printers
```

Setting Value:
```
{
    enabled = true | false
}
```

### Client removable drives
Description: 
```
Allows or prevents users from accessing or saving files to removable drives on the user device. By default, accessing client removable drives is allowed.

When enabling this setting, make sure the 'Client drive redirection' setting is enabled. Additionally, to ensure removable drives are automatically connected, configure the 'Auto connect client drives' setting. If these settings are disabled, client removable drives are not mapped and users cannot access these drives manually, regardless of the state of the 'Client removable drives' setting.
```

Setting Name: `ClientRemoveableDrives`

Related Settings:
```
Client drive redirection

Auto connect client drives
```

Setting Value:
```
{
    enabled = true | false
}
```

### Client TWAIN device redirection
Description: 
```
Allows or prevents users to access TWAIN devices, such as digital cameras or scanners, on the client device from published image processing applications. By default, TWAIN device redirection is allowed.
```

Setting Name: `TwainRedirection`

Related Settings:
```
TWAIN compression level

TWAIN device redirection bandwidth limit

TWAIN device redirection bandwidth limit percent
```

Setting Value:
```
{
    enabled = true | false
}
```

### Client USB device optimization rules
Description: 
```
This setting is available only on VDA versions XenApp and XenDesktop 7.6 Feature Pack 3 and later.

When a user plugins a USB input device, the host checks if the device is allowed by policy. If device is allowed by policy, host further checks the optimization rules for the device. If no rule is specified the device is considered operating in 'Interactive' mode.

The policy rules can be applied to devices to disable optimization or change the optimization mode. The policy rules take the form of 'tag' ='expressions' separated by whitespaces. The following tags are supported:

Mode - The device optimization mode. Refer optimization mode below.
VID - Vendor ID from the device descriptor
PID - Product ID from the device descriptor
REL - Release ID from the device descriptor
Class - Class from either the device descriptor or an interface descriptor
Subclass - Subclass from either the device descriptor or an interface descriptor
Prot - Protocol from either the device descriptor or an interface descriptor

When creating new policy rules, be aware of the following:

* Rules are case-insensitive.
* Values on the right side of the = sign are hex numbers, which may or may not have the leading 0x.
* Rules may have an optional comment at the end, introduced by #.
* Blank and pure comment lines are ignored.
* Tags must use the matching operator =. For example Mode=00000001 VID=1230.
* Refer to the USB class codes available from the USB implementers Forum Inc. Web site.
* The optimization mode is supported for input devices for class=03. Current supported modes are as:

* No Optimization - 00000001
* Interactive Mode - 00000002
* Capture Mode - 00000004

Examples of administrator-defined USB policy rules:

Mode=00000004 VID=1230 PID=1230 class=03 #Input device operating in capture mode
Mode=00000002 VID=1230 PID=1230 class=03 #Input device operating in interactive mode
Mode=00000001 VID=1230 PID=1230 class=03 #Input device operating without any optimization
```

Setting Name: `ClientUsbDeviceOptimizationRules`

Related Settings:
```
Client USB device redirection

Client USB device redirection rules
```

Setting Value:
```
jsonencode([
    {RULE_1},
    {RULE_2},
    ...
])
```

### Client USB device redirection
Description: 
```
Enables or disables redirection of USB devices to and from the client (workstation hosts only).
```

Setting Name: `UsbDeviceRedirection`

Related Settings:
```
Client USB device redirection rules
```

Setting Value:
```
{
    enabled = true | false
}
```

### Client USB device redirection bandwidth limit
Description: 
```
Specifies the maximum allowed bandwidth in kilobits per second for redirection of USB devices to and from the client (workstation hosts only).

If you enter a value for this setting and a value for the 'Client USB device redirection bandwidth limit Percent' setting, the most restrictive setting (with the lower value) is applied.
```

Setting Name: `USBBandwidthLimit`

Related Settings:
```
Client USB device redirection

Client USB device redirection bandwidth limit percent

Overall session bandwidth limit
```

Setting Value: `{Bandwidth limit in Kbps}`

### Client USB device redirection bandwidth limit percent
Description: 
```
Specifies the maximum allowed bandwidth for redirection of USB devices to and from the client (workstation hosts only).

If you enter a value for this setting and a value for the 'Client USB device redirection bandwidth limit' setting, the most restrictive setting (with the lower value) is applied.

If you configure this setting, you must also configure the 'Overall session bandwidth limit' setting which specifies the total amount of bandwidth available for client sessions.
```

Setting Name: `USBBandwidthPercent`

Related Settings:
```
Client USB device redirection

Client USB device redirection bandwidth limit

Overall session bandwidth limit
```

Setting Value: `{Bandwidth Limit Percentage}`

### Client USB device redirection rules
Description: 
```
Lists redirection rules for USB devices.

When a user plugs in a USB device, the host device checks it against each policy rule in turn until a match is found. The first match for any device is considered definitive. When the first match is an Allow rule, the device is remoted to the virtual desktop. When the first match is a Deny rule, the device is available only to the local desktop. When no match is found, default rules are used. For more information about the default policy configuration for USB devices, see CTX119722, 'Creating USB Policy Rules,' in the Citrix Knowledge Center.

Policy rules take the format (Allow:|Deny:) followed by a set of tag= value expressions separated by whitespace. The following tags are supported:
* VID - Vendor ID from the device descriptor
* PID - Product ID from the device descriptor
* REL - Release ID from the device descriptor
* Class - class from either the device descriptor or an interface descriptor
* SubClass - Subclass from either the device descriptor or an interface descriptor
* Prot - Protocol from either the device descriptor or an interface descriptor

When creating new policy rules, be aware of the following:
* Rules are case-insensitive.
* Rules can have an optional comment at the end, introduced by #.
* Blank and pure comment lines are ignored.
* Tags must use the matching operator =. For example, VID=1230.
* Each rule must start on a new line or form part of a semicolon-separated list.
* See the USB class codes available from the USB Implementers Forum, Inc. website.

Examples of administrator-defined USB policy rules:
Allow: VID=1230 PID=0007 # Another Industries, Another Flash Drive
Deny: Class=08 subclass=05 # Mass Storage
To create a rule that denies all USB devices, use 'DENY:' with no other tags.
```

Setting Name: `UsbDeviceRedirectionRules`

Related Settings:
```
Client USB device redirection

Client USB device redirection rules (Version 2)
```

Setting Value:
```
jsonencode([
    {RULE_1},
    {RULE_2},
    ...
])
```

### Client USB device redirection rules (Version 2) 
Description: 
```
Specifies rules for filtering, splitting and auto-connecting USB devices to a remote session.

When this setting is selected, the host overrides the setting - 'Client USB device redirection rules' with device rules configured in this setting.

The device rules are an ordered list of case insensitive rules terminated by newlines or semicolon, where:
- anything after "#" is line comment
- each rule is:
(CONNECT | ALLOW | DENY | FORCEDENY): (filters)* (split/intf) (attributes)*

- (filters)* are a series of zero or more USB device selector key/value pairs:
(vid | pid | rel) = (xxxx | *) or (class | subclass | prot) = (xx | *), where
xxxx is a 4-digit hex number. xx is a 2-digit hex number. * matches any value.
if unspecified, default is * [match any]

- (split/intf) is an optional directive to split and select interfaces of a composite USB device:
split=(0 | 00 | f | false | 1 | 01 | t | true)
intf=* or intf=(xx(,xx)*)
if unspecified, default is "split=0 intf=*"
Example: split=01 intf=00,02,03 # split and apply to three interfaces

- (attributes)* is an optional list of attributes key/value pairs
Example: Mode=1 ConnectNew=1 ConnectExisiting=0

The rules configured in this setting override the client-side GPO and the defaults. The sole exception to this is the FORCEDENY rule, which can be used in the client-side to override an ALLOW/CONNECT rule configured in this DDC setting.

The default value for the device rules is:
# Block some devices we never want to see
DENY: vid=17e9 # All DisplayLink USB displays
DENY:vid=045e pid=079A # Microsoft Surface Pro 1 Touch Cover
DENY:vid=045e pid=079c # Microsoft Surface Pro 1 Type Cover
DENY:vid=045e pid=07dc # Microsoft Surface Pro 3 Type Cover
DENY:vid=045e pid=07dd # Microsoft Surface Pro JP 3 Type Cover
DENY:vid=045e pid=07de # Microsoft Surface Pro 3_2 Type Cover
DENY:vid=045e pid=07e2 # Microsoft Surface Pro 3 Type Cover
DENY:vid=045e pid=07e4 # Microsoft Surface Pro 4 Type Cover with fingerprint reader
DENY:vid=045e pid=07e8 # Microsoft Surface Pro 4_2 Type Cover
DENY:vid=03eb pid=8209 # Surface Pro Atmel maXTouch Digitizer
#
# Special devices - Bloomberg 5 keyboard
#CONNECT: vid=1188 pid=A101 # Bloomberg 5 Biometric module
#DENY: vid=1188 pid=A001 split=01 intf=00 # Bloomberg 5 Primary keyboard
#CONNECT: vid=1188 pid=A001 split=01 intf=01 # Bloomberg 5 Keyboard HID
#DENY: vid=1188 pid=A301 split=01 intf=02 # Bloomberg 5 Keyboard Audio Channel
#CONNECT: vid=1188 pid=A301 split=01 intf=00,01 # Bloomberg 5 Keyboard Audio HID
#
# Block device classes that are not useful/dangerous
DENY: class=03 subclass=01 prot=01 # HID Boot keyboards
DENY: class=03 subclass=01 prot=02 # HID Boot mice
DENY: class=09 # Hub devices
DENY: class=11 # Billboard device
DENY: class=12 # Type C bridge device
DENY: class=3c # Diagnostic device
DENY: class=e0 # Wireless controller
DENY: class=ef subclass=04 # Miscellaneous network device
#
# Some additional classes we usually don't want to redirect but an admin might want to change
DENY: class=02 # Other Communications and CDC-Control devices
DENY: class=0a # CDC-Data
DENY: class=0b # Smartcard
```

Setting Name: `USBDeviceRulesV2`

Related Settings:
```
Client USB device redirection

Client USB device redirection rules
```

Setting Value:
```
jsonencode([
    {RULE_1},
    {RULE_2},
    ...
])
```

### Client USB Plug and Play device redirection
Description: 
```
Allows or prevents plug-n-play devices such as cameras or point-of-sale (POS) devices to be used in a client session. By default, plug-n-play device redirection is allowed.

When set to 'Allowed', all plug-n-play devices for a specific user or group are redirected. When set to 'Prohibited', no devices are redirected.
```

Setting Name: `UsbPlugAndPlayRedirection`

Setting Value:
```
{
    enabled = true | false
}
```

### Clipboard place metadata collection for security monitoring
Description: 
```
Clipboard place metadata collection by Broker service for security monitoring, auditing, and compliance
```

Setting Name: `EnableClipboardMetadataCollection`

Setting Value:
```
{
    enabled = true | false
}
```

### Clipboard redirection bandwidth limit
Description: 
```
Specifies the maximum allowed bandwidth in kilobits per second for data transfer between a session and the local clipboard.

If you enter a value for this setting and a value for the 'Clipboard redirection bandwidth limit percent' setting, the most restrictive setting (with the lower value) is applied.
```

Setting Name: `ClipboardBandwidthLimit`

Related Settings:
```
Clipboard redirection bandwidth limit percent

Client clipboard redirection
```

Setting Value: `{Bandwidth Limit in Kbps}`

### Clipboard redirection bandwidth limit percent
Description: 
```
Specifies the maximum allowed bandwidth for data transfer between a session and the local clipboard as a percent of the total session bandwidth.

If you enter a value for this setting and a value for the 'Clipboard redirection bandwidth limit' setting, the most restrictive setting (with the lower value) is applied.

If you configure this setting, you must also configure the 'Overall session bandwidth limit' setting which specifies the total amount of bandwidth available for client sessions.
```

Setting Name: `ClipboardBandwidthPercent`

Related Settings:
```
Clipboard redirection bandwidth limit

Overall session bandwidth limit
```

Setting Value: `{Bandwidth Limit Percentage}`

### Clipboard selection update mode
Description: 
```
This setting is supported only by Linux VDA version 1.4 onwards.

This setting controls whether CLIPBOARD selection changes on the Linux VDA are updated on the client's clipboard (and vice versa). It can include one of the following selection changes:

AllUpdatesDenied: Selection changes are not updated on the client or the host. CLIPBOARD selection changes do not update a client's clipboard. Client clipboard changes do not update CLIPBOARD selection.

UpdateToClientDenied: Host selection changes are not updated on the client. CLIPBOARD selection changes do not update a client's clipboard. Client clipboard changes update the CLIPBOARD selection.

UpdateToHostDenied: Client selection changes are not updated on the host. CLIPBOARD selection changes update the client's clipboard. Client clipboard changes do not update the CLIPBOARD selection.

AllUpdatesAllowed: Selection changes are updated on both the client and host. CLIPBOARD selection change updates the client's clipboard. Client clipboard changes update the CLIPBOARD selection.
```

Setting Name: `ClipboardSelectionUpdateMode`

Related Settings:
```
Primary selection update mode
```

Setting Value: `AllUpdatesDenied`, `UpdateToClientDenied`, `UpdateToHostDenied`, `AllUpdatesAllowed`

### COM port redirection bandwidth limit
Description: 
```
For VDA versions that do not support this setting, follow CTX139345 to enable redirection using the registry.

Specifies the maximum allowed bandwidth in kilobits per second for accessing a COM port in a client connection.

If you enter a value for this setting and a value for the 'COM port redirection bandwidth limit percent' setting, the most restrictive setting (with the lower value) is applied.
```

Setting Name: `ComPortBandwidthLimit`

Related Settings:
```
COM port redirection bandwidth limit percent

Client COM port redirection
```

Setting Value: `{Bandwidth Limit in Kbps}`

### COM port redirection bandwidth limit percent
Description: 
```
For VDA versions that do not support this setting, follow CTX139345 to enable redirection using the registry.

Specifies the maximum allowed bandwidth for accessing COM ports in a client connection as a percent of the total session bandwidth.

If you enter a value for this setting and a value for the 'COM port redirection bandwidth limit (Kbps)' setting, the most restrictive setting (with the lower value) is applied.

If you configure this setting, you must also configure the 'Overall session bandwidth limit' setting which specifies the total amount of bandwidth available for client sessions.
```

Setting Name: `ComPortBandwidthPercent`

Related Settings:
```
COM port redirection bandwidth limit

Overall session bandwidth limit

Client COM port redirection
```

Setting Value: `{Bandwidth Limit Percentage}`

### Common information
Description: 
```
Applies to: both file-based and container-based profile solutions

Detailed log settings.

Define events or actions that Profile management logs in depth.

If this setting is not configured here, Profile management uses the settings from the .ini file.

If this setting is not configured here or in the .ini file, errors and general information are logged.
```

Setting Name: `LogLevel_Information`

Setting Value:
```
{
    enabled = true | false
}
```

### Common warnings
Description: 
```
Applies to: both file-based and container-based profile solutions

Detailed log settings.

Define events or actions that Profile management logs in depth.

If this setting is not configured here, Profile management uses the settings from the .ini file.

If this setting is not configured here or in the .ini file, errors and general information are logged.
```

Setting Name: `LogLevel_Warnings`

Setting Value:
```
{
    enabled = true | false
}
```

### Concurrent logons tolerance
Description: 
```
Define the expected number of concurrent logons for a server.
```

Setting Name: `ConcurrentLogonsTolerance`

Related Settings:
```
Maximum number of sessions

CPU usage

Disk usage

Memory usage

Memory usage base load
```

Setting Value: `{Expected number off concurrent logons for a server}`

### Contacts path
Description: 
```
Lets you specify how to redirect the Contacts folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRContactsPath_Part`

Setting Value: `{Redirected Path for Contacts Folder}`

### CPU usage
Description: 
```
Defines the CPU usage percentage value at which the server reports full load.
```

Setting Name: `CPUUsage`

Related Settings:
```
Maximum number of sessions

Concurrent logons tolerance

Disk usage

Memory usage

Memory usage base load
```

Setting Value: `{CPU Usage Percentage Value to Report Full Load}`

### CPU usage excluded process priority
Description: 
```
It is quite common for background processes to run when a server is deemed idle and to consume all CPU that is not required by other active processes on the server. One way of achieving this is for a background process to run at a process priority that is below Normal, i.e. 'Below Normal' or 'Low'. If a new user session were to be established on a server when a background process was consuming CPU, the Windows Operating System scheduler would attempt to withdraw as much CPU from background process as necessary to fulfill the demands of the processes from the new user session. This setting allows the CPU usage associated with background processes to be excluded from the CPU Usage load calculation.

If the CPU Usage setting is disabled, then this setting will be ignored irrespective of its configuration.
```

Setting Name: `CPUUsageExcludedProcessPriority`

Related Settings:
```
Maximum number of sessions

CPU usage

Disk usage

Memory usage

Memory usage base load
```


Setting Value: `Disabled`, `Low`, `BelowNormalOrLow`

### Cross-platform settings user groups
Description: 
```
Applies to: both file-based and container-based profile solutions

Enter one or more Windows user groups.

If this setting is configured, the cross-platform settings feature of Profile management processes only members of these user groups. If this setting is disabled, all user groups are processed.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, all user groups are processed.
```

Setting Name: `CPUserGroups_Part`

```
jsonencode([
    {User_Group_1},
    {User_Group_2},
    ...
])
```

### Customer Experience Improvement Program
Description: 
```
Applies to: both file-based and container-based profile solutions

By default, the Customer Experience Improvement Program is enabled to help improve the quality and performance of Citrix products by sending anonymous statistics and usage information.

If this setting is not configured here, the value from the .ini file is used.
```

Setting Name: `CEIPEnabled`

Setting Value:
```
{
    enabled = true | false
}
```

### Customize storage path for VHDX files
Description: 
```
Applies to: both file-based and container-based profile solutions

By default, Profile Management stores VHDX files in the user store. With this policy enabled, you can specify a separate path to store VHDX files.

Policies that use VHDX files include the following: Profile container, Search index roaming for Outlook, and Accelerate folder mirroring. VHDX files of different policies are stored in different folders under the storage path.
```

Setting Name: `VhdStorePath_Part`

Setting Value: `{Path to store VHDX files}`

### Customized User Layer Size in GB
Description: 
```
The customized size (in GB) of each new user layer disk. The value must be between 10GB and 2040GB.
```

Setting Name: `UplCustomizedUserLayerSizeInGb`

Setting Value: `{User Layer Size in GB}`

### Default capacity of VHD containers
Description: 
```
Applies to: both file-based and container-based profile solutions

Specifies the default storage capacity (in GB) of VHD containers.

If this policy is not configured here, the value from the .ini file is used.

If this policy is not configured either here or in the .ini file, the default is 50 (GB).
```

Setting Name: `VhdContainerCapacity_Part`

Setting Value: `{Default storage capacity (in GB) of VHD containers}`

### Default printer
Description: 
```
Specifies how the client's default printer is established in an ICA session. By default, the client's current printer is used as the default printer for the session.

ClientDefault: 'Set default printer to the client's main printer' allows the client's current default printer to be used as the default printer for the session.

GenericUniversalPrinter: 'Set default printer to the Generic Universal printer' allows the Generic Universal printer to be used as the default printer for the session.

PDFPrinter: 'Set default printer to the PDF printer' allows the PDF printer to be used as the default printer for the sessions using the the Citrix Workspace App for HTML5 or Chrome.

DoNotAdjust: 'Do not adjust the user's default printer' uses the current Terminal Services or Windows user profile setting for the default printer. If you choose this option, the default printer is not saved in the profile and it does not change according to other session or client properties. The default printer in a session will be the first printer auto-created in the session, which is either:

* The first printer added locally to the Windows server in Control Panel > Printers
* The first auto-created printer, if there are no printers added locally to the server

You can use this option to present users with the nearest printer through profile settings (known as Proximity Printing).

When policy settings are resolved, the default printer applied by this setting is overridden by a default printer applied by the 'Printer assignments' setting.

Use individual 'Default printers' policies to set default behaviors for a farm, site, large group, or OU. Use the 'Printer assignments' setting to assign a large group of printers to multiple users.
```

Setting Name: `DefaultClientPrinter`

Setting Value: `ClientDefault`, `DoNotAdjust`, `GenericUniversalPrinter`, `PDFPrinter`

### Delay before deleting cached profiles
Description: 
```
Applies to: file-based profile solution only

Works only if 'Delete locally cached profiles on logoff' is enabled. Sets an optional extension to the delay before locally cached profiles are deleted at logoff. A value of 0 deletes the profiles immediately during logoff processing. Checks take place every minute, so a value of 60 ensures that profiles are deleted between one and two minutes after users have logged off (depending on when the last check took place). Extending the delay is useful if you know that a process keeps files or the user registry hive open during logoff. With large profiles, this can also speed up logoff.

If this policy is not configured here, the value from the .ini file is used.

If this policy is not configured here or in the .ini file, profiles are deleted immediately.
```

Setting Name: `ProfileDeleteDelay_Part`

Setting Value: `{Delay time in seconds}` 

### Delete locally cached profiles on logoff
Description: 
```
Applies to: file-based profile solution only

Specifies whether locally cached profiles are deleted after logoff.

If this setting is enabled, a user's local profile cache is deleted after they have logged off. This is recommended for terminal servers.

If this setting is disabled cached profiles are not deleted.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, cached profiles are not deleted.
```

Setting Name: `DeleteCachedProfilesOnLogoff`

Setting Value:
```
{
    enabled = true | false
}
```

### Desktop launches
Description: 
```
Allows or prevents non-administrative users to connect to a desktop session on the server.

When allowed, non-administrative users can connect. By default, non-administrative users cannot connect to desktop sessions.
```

Setting Name: `DesktopLaunchForNonAdmins`

Setting Value:
```
{
    enabled = true | false
}
```

### Desktop path
Description: 
```
Applies to: file-based profile solution only

Lets you specify how to redirect the Desktop folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRDesktopPath_Part`

Setting Value: `{Path to the Redirected Desktop Folder}`

### Desktop wallpaper
Description: 
```
Enables or disables the desktop wallpaper in user sessions. By default, desktop wallpaper is allowed.
```

Setting Name: `DesktopWallpaper`

Setting Value:
```
{
    enabled = true | false
}
```

### Diagnostic data collection for performance monitoring
Description: 
```
The monitoring service gathers diagnostic data such as session information, UPM/EUEM service states, Microsoft Teams optimisation, and connection protocols.
```

Setting Name: `EnableVdaDiagnosticsCollection`

Setting Value:
```
{
    enabled = true | false
}
```


### Direct connections to print servers
Description: 
```
Enables or disables direct connections from the host to a print server for client printers hosted on an accessible network share. By default, direct connections are enabled.

Allow direct connections if the network print server is not across a WAN from the host. Direct communication results in faster printing if the network print server and host server are on the same LAN.
```

Setting Name: `DirectConnectionsToPrintServers`

Setting Value:
```
{
    enabled = true | false
}
```

### Directories to synchronize
Description: 
```
Applies to: file-based profile solution only

Profile management synchronizes each user's entire profile between the system it is installed on and the user store.

It allows you to include subfolders of excluded folders.

Paths on this list should be relative to the user profile.

Wildcards are supported but they are not applied recursively.

Examples:
Desktop\exclude\include specifies the include subfolder of the Desktop\exclude folder.

Desktop\exclude\* specifies all immediate subfolders of the Desktop\exclude folder.

Disabling this setting has the same effect as enabling it and configuring an empty list.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, only non-excluded folders in the user profile are synchronized.
```

Setting Name: `SyncDirList_Part`

Setting Value: 
```
jsonencode([
    {Path 1 to Sync},
    {Path 2 to Sync},
    ...
])
```

### Disable automatic configuration
Description: 
```
Applies to: both file-based and container-based profile solutions

Profile management 5.x examines its environment and configures itself accordingly. To disable this for troubleshooting or to retain the settings you used in an earlier version, enable this setting.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, automatic configuration is turned on so Profile management settings might change if the environment changes.
```

Setting Name: `DisableDynamicConfig`

Setting Value:
```
{
    enabled = true | false
}
```

### Disable defragmentation for VHD disk compaction
Description: 
```
Applies to: both file-based and container-based profile solutions

Lets you specify whether to disable file defragmentation for VHD disk compaction.

When VHD disk compaction is enabled, the VHD disk file is first automatically defragmented using the Windows built-in 'defrag' tool, and then compacted. VHD disk defragmentation produces better compaction results while disabling it can save system resources.

If this policy is not configured here, the value from the .ini file is used.

If this policy is not configured here or in the .ini file, the defragmentation is enabled by default.
```

Setting Name: `NDefrag4Compaction`

Setting Value:
```
{
    enabled = true | false
}
```

### Disconnected session timer
Description: 
```
Enables or disables a timer to determine how long a disconnected, locked workstation can remain locked before the session is logged off. By default, this timer is disabled, and disconnected sessions are not logged off.
```

Setting Name: `SessionDisconnectTimer`

Related Settings:
```
Disconnected session timer interval
```

Setting Value:
```
{
    enabled = true | false
}
```

### Disconnected session timer - Multi-session
Description: 
```
Enables or disables a timer to determine how long a disconnected RDS session, can persist before the session is logged off. By default, this timer is disabled, and disconnected sessions are not logged off.
```

Setting Name: `EnableServerDisconnectionTimer`

Related Settings:
```
Disconnected session timer interval - Multi-session
```

Setting Value:
```
{
    enabled = true | false
}
```

### Disconnected session timer for Remote PC access
Description: 
```
Enables or disables a timer to determine how long a disconnected session on a locked workstation can remain in a disconnected and locked state before the session is logged off. By default, this timer is disabled, and disconnected sessions are not logged off.
```

Setting Name: `EnableRemotePCDisconnectTimer`

Related Settings:
```
Disconnected session timer interval
```

Setting Value:
```
{
    enabled = true | false
}
```

### Disconnected session timer interval
Description: 
```
Determines how long, in minutes, a disconnected, locked workstation can remain locked before the session is logged off. By default, the time period is 1440 minutes (24 hours).
```

Setting Name: `SessionDisconnectTimerInterval`

Setting Value: `{Timer Interval in Minutes}`

### Disconnected session timer interval - Multi-session
Description: 
```
Determines how long, in minutes, a disconnected RDS session can persist before the session is logged off. By default, the time period is 1440 minutes (24 hours).
```

Setting Name: `ServerDisconnectionTimerInterval`

Setting Value: `{Timer Interval in Minutes}`

### Disk usage
Description: 
```
Defines the disk queue length value at which the server reports 75% load.

Zero load is automatically determined to be one third of the 75% load disk queue length value. It is a common rule of thumb that when disk queue length exceeds twice the number of spindles, a disk bottleneck is likely developing. This should be considered in conjunction with the disk topology of the server when setting the '75% disk queue length' value.
```

Setting Name: `DiskUsage`

Related Settings:
```
Maximum number of sessions

Concurrent logons tolerance

CPU usage

Memory usage

Memory usage base load
```

Setting Value: `{Disk Queue Length to report 75% load}`

### Display memory limit
Description: 
```
This setting specifies the maximum video buffer size in kilobytes for the session.

For VDA versions 2308 and prior, specify an amount in kilobytes from 128 to 4,194,303. When not specified, the setting defaults to 65536 kilobytes.
For VDA versions 2311 and later, specify any amount in kilobytes. When not specified, or when specified as 0, the setting defaults to no maximum.

Using more color depth and higher resolution for connections requires more memory. In legacy graphics mode, if the memory limit is reached, the display degrades according to the 'Display mode degrade preference' setting.
```

Setting Name: `DisplayMemoryLimit`

Setting Value: `{Memory Limit in KB}`

### Documents path
Description: 
```
Applies to: file-based profile solution only

Lets you specify how to redirect the Documents folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRDocumentsPath_Part`

Setting Value: `{Path to the Redirected Documents Folder}`

### Download file for Citrix Workspace app for Chrome OS/HTML5
Description: 
```
This setting allows or prevents Citrix Workspace app for Chrome OS/HTML5 users from downloading files from a virtual desktop to the client device. By default, downloading files from virtual desktops is enabled.

When adding this setting to a policy, ensure that the 'File transfer for Citrix Workspace app for Chrome OS/HTML5' setting is present and set to Allowed.
```

Setting Name: `AllowFileDownload`

Related Settings:
```
File transfer for Citrix Workspace app for Chrome OS/HTML5

Upload file for Citrix Workspace app for Chrome OS/HTML5
```

Setting Value:
```
{
    enabled = true | false
}
```

### Downloads path
Description: 
```
Applies to: file-based profile solution only

Lets you specify how to redirect the Downloads folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRDownloadsPath_Part`

Setting Value: `{Path to the Redirected Downloads Folder}`

### Drag and drop
Description: 
```
When enabled, files can be dragged between client and remote applications and desktops.
```

Setting Name: `DragDrop`

Setting Value:
```
{
    enabled = true | false
}
```

### Dynamic windows preview
Description: 
```
Dynamic windows preview enables the state of seamless windows to be seen on the various windows previews (Flip, Flip 3D, Taskbar Preview, and Peek). By default, dynamic windows preview is enabled.
```

Setting Name: `DynamicPreview`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable asynchronous processing for user Group Policy on logon
Description: 
```
Applies to: both file-based and container-based profile solutions

Windows provides two processing modes for user Group Policy: synchronous and asynchronous. Windows uses a registry value to determine the processing mode for the next user logon. If the registry value doesn't exist, synchronous mode is applied. The registry value is a machine-level setting and doesn't roam with users. Thus, asynchronous mode will not be applied as expected if users:
Log on to different machines.
Log on to the same machine where the Delete locally cached profiles on logoff policy is enabled.

With this policy enabled, Citrix Profile Management lets the registry value roam with users. As a result, processing mode is applied each time users log on.

For asynchronous mode to take effect on Windows Server machines, make sure that you install the Remote Desktop Session Host role and set the Group Policies as follows:
Computer Config | Admin Templates | System | Logon | Always wait for the network at computer startup and logon: Disabled
Computer Config | Admin Templates | System | Group Policy | Allow asynchronous user Group Policy processing when logging on through Remote Desktop Services: Enabled
```

Setting Name: `SyncGpoStateEnabled`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable auto update of Controllers
Description: 
```
Enable this setting to apply the list of DDCs to the initial bootstrap connection. Disable this setting to manage the list of DDCs manually.
```

Setting Name: `EnableAutoUpdateOfControllers`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable Citrix Virtual Apps Optimization
Description: 
```
Applies to: file-based profile solution only

When you enable this feature, only the settings specific to the published applications a user launches or exits are synchronized.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, no Citrix Virtual Apps optimization settings will be applied.
```

Setting Name: `XenAppOptimizationEnabled`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable concurrent session support for Outlook search data roaming
Description: 
```
Applies to: both file-based and container-based profile solutions

Provides native Outlook search experience in concurrent sessions. Use this policy with the Search index roaming for Outlook policy.

With this policy enabled, each concurrent session uses a separate Outlook OST file.

By default, only two VHDX disks can be used to store Outlook OST files (one file per disk). If more sessions start, their Outlook OST files are stored in the local user profiles. You can specify the maximum number of VHDX disks for storing Outlook OST files.

Example:
You set the number to 3. As a result, Profile Management stores Outlook OST files on the VHDX disks for the first, second, and third concurrent sessions. It stores the OST file for the fourth concurrent session in the local profile instead.
```

Setting Name: `OutlookSearchRoamingConcurrentSession`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable credential-based access to user stores
Description: 
```
Applies to: both file-based and container-based profile solutions

By default, Profile Management impersonates the current user to access user stores. Therefore, it requires the current user to have permission to directly access the user stores. Enable this feature if you do not want Profile Management to impersonate the current user when accessing user stores. You can put user stores in storage repositories (for example, Azure Files) that the current user has no permission to access.

To ensure that Profile Management can access user stores, save the profile storage server's credentials in Workspace Environment Management or Windows Credential Manager. We recommend that you use Workspace Environment Management to eliminate the need to configure the same credentials for each machine where Profile Management runs. If you use Windows Credential Manager, use the Local System account to securely save the credentials.

Note: Starting with Profile Management version 2212, this feature is available for both file-based and VHDX-based Profile Management solutions. For Profile Management versions earlier than 2212, this feature is available only for the VHDX-based Profile Management solution.

If this setting is not configured here, the value from the .ini file is used. If this setting is configured neither here nor in the .ini file, it is disabled by default.
```

Setting Name: `CredBasedAccessEnabled`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable cross-platform settings
Description: 
```
Applies to: both file-based and container-based profile solutions

By default, to facilitate deployment, cross-platform settings is disabled.

Turn on processing by enabling this setting.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, no cross-platform settings will be applied.
```

Setting Name: `CPEnable`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable default exclusion list
Description: 
```
Applies to: file-based profile solution only

Default list of registry keys in the HKCU hive that are not synchronized to the user's profile.

If you disable this setting, registry keys are not excluded by default.

If you do not configure this setting here, the value from the .ini file is used.

If you do not configure this setting here or in the .ini file, registry keys are not excluded by default.

Note: Software\Miscrosoft\Speech_OneCore is deprecated in Profile Management 5.8 and later.
```

Setting Name: `DefaultExclusionList`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable default exclusion list - directories
Description: 
```
Applies to: file-based profile solution only

Default list of directories ignored during synchronization.

If you disable this setting, folders are not excluded by default.

If you do not configure this setting here, the value from the .ini file is used.

If you do not configure this setting here or in the .ini file, folders are not excluded by default.
```

Setting Name: `DefaultExclusionListSyncDir`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable exclusive access to VHD containers - OneDrive container
Description: 
```
Applies to: both file-based and container-based profile solutions

By default, VHD containers allow concurrent access. With this setting enabled, they allow only one access at a time.

Note: Enabling this setting for profile containers in a container-based profile solution automatically disables the "Enable multi-session write-back for profile containers" setting.

If this policy is not configured here, the value from the .ini file is used.

If this policy is not configured either here or in the .ini file, the setting is disabled.
```

Setting Name: `DisableConcurrentAccessToOneDriveContainer`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable exclusive access to VHD containers - Profile container
Description:
```
Applies to: both file-based and container-based profile solutions

By default, VHD containers allow concurrent access. With this setting enabled, they allow only one access at a time.

Note: Enabling this setting for profile containers in a container-based profile solution automatically disables the "Enable multi-session write-back for profile containers" setting.

If this policy is not configured here, the value from the .ini file is used.

If this policy is not configured either here or in the .ini file, the setting is disabled.
```

Setting Name: `DisableConcurrentAccessToProfileContainer`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable local caching for profile containers
Description: 
```
Applies to: container-based profile solution only

With local caching for profile containers enabled, each local profile serves as a local cache of its profile container. If profile streaming is in use, locally cached files are created on demand. Otherwise, they are created during user logons.
```

Setting Name: `ProfileContainerLocalCache`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable logging
Description: 
```
Applies to: both file-based and container-based profile solutions

Activation of this setting enables debug mode (verbose logging). In debug mode, extensive status information is logged in the log files in "%SystemRoot%\System32\Logfiles\UserProfileManager" or in the location specified by the "Path to log file" policy setting.

If this setting is disabled only errors are logged.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, only errors are logged.
```

Setting Name: `DebugMode`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable monitoring of application failures
Description: 
```
Use this setting to configure monitoring of application failures.

* You can monitor either application errors or faults (crashes and unhandled exceptions), or both.
* You can disable monitoring of application failures by setting Value to None.

To specify a list of processes not to be monitored, set the 'List of applications excluded from failure monitoring' setting

By default, failures from applications hosted only on the Server OS VDAs are monitored. To monitor Desktop OS VDAs, set the 'Enable monitoring of application failures on Desktop OS VDAs' setting.
```

Setting Name: `SelectedFailureLevel`

Related Settings:
```
List of applications excluded from failure monitoring

Enable monitoring of application failures on Desktop OS VDAs
```

Setting Value: `None`, `Error`, `Fault`, `Both`
Setting Value Explanation:
Setting Value | Explanation
-- | --
`None` | None
`Error` | Application Errors Only
`Fault` | Application Faults Only
`Both` | Both Application Errors and Faults

### Enable monitoring of application failures on Desktop OS VDAs
Description: 
```
Use this setting to enable monitoring of application failures on Desktop OS VDAs. By default, application failures on the Desktop OS VDAs are not monitored. If you have applications hosted on Desktop OS VDAs, enable and scope this policy to monitor specific Delivery Groups.
```

Setting Name: `EnableWorkstationVDAFaultMonitoring`

Related Settings:
```
Enable monitoring of application failures
```

Setting Value:
```
{
    enabled = true | false
}
```

### Enable multi-session write-back for profile containers
Description: 
```
Applies to: both file-based and container-based profile solutions

Enables write-back for profile containers in multi-session scenarios. If enabled, changes in all sessions are written back to profile containers. Otherwise, only changes in the first session are saved because only the first session is in read/write mode in profile containers. Citrix Profile Management profile containers are supported starting with Citrix Profile Management 2103. FSLogix Profile Container is supported starting with Citrix Profile Management 2003.

To use this policy for FSLogix Profile Container, ensure that the following prerequisites are met:
1.The FSLogix Profile Container feature is installed and enabled.
2.Profile type is set to "Try for read-write profile and fallback to read-only" in FSLogix.
```

Setting Name: `FSLogixProfileContainerSupport`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable process monitoring
Description: 
```
Enable this setting to allow monitoring of processes running on machines with VDAs. Statistics (such as CPU and memory use) are sent to the Monitor Service. The statistics are used for real-time notifications and historical reporting in Director.

Note: Process monitoring requires large amounts of storage. For example, one year of data for a deployment with 50,000 VDAs might require approximately 500 GB of SQL Server storage.
```

Setting Name: `EnableProcessMonitoring`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable Profile management
Description: 
```
Applies to: both file-based and container-based profile solutions

By default, to facilitate deployment, Profile management does not process logons or logoffs.

Turn on processing by enabling this setting.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, Profile management does not process Windows user profiles in any way.
```

Setting Name: `ServiceActive`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable profile streaming for folders
Description: 
```
Applies to: file-based profile solution only

With profile streaming for folders enabled, folders are fetched only when they are being accessed. This approach eliminates the need to traverse all folders during user logons.
```

Setting Name: `PSForFoldersEnabled`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable profile streaming for pending area
Description: 
```
Applies to: file-based profile solution only

With this policy enabled, files in the pending area are fetched to the local profile only when they are requested. Use the policy with the Profile streaming policy to ensure optimum logon experience in concurrent session scenarios.

The policy applies to folders in the pending area when the Enable profile streaming for folders policy is enabled.

By default, this policy is disabled. All files and folders in the pending area are fetched to the local profile during logon.

The pending area is used to ensure profile consistency while profile streaming is enabled. It temporarily stores profile files and folders changed in concurrent sessions.
```

Setting Name: `PSForPendingAreaEnabled`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable resource monitoring
Description: 
```
Enable this setting to allow monitoring of critical performance counters on machines with VDAs. Statistics (such as CPU, memory , IOPS and disk latency data) are sent to the monitoring Service. The statistics are used for real-time notifications and historical reporting in Director.
```

Setting Name: `EnableResourceMonitoring`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable search index roaming for Outlook
Description: 
```
Applies to: both file-based and container-based profile solutions

Allow user-based Outlook search experience by automatically roaming Outlook search data along with user profile. This requires extra spaces in the user store to store search index for Outlook.

You must log off and then log on again for this policy to take effect.
```

Setting Name: `OutlookSearchRoamingEnabled`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable session watermark
Description: 
```
Controls the way in which ICA session content is displayed.

When this setting is enabled, session display will be watermarked by an extra layer of semi-transparent text labels containing session-specific information (logon username, VDA host name etc).

Session watermark serves as a deterrent method in scenarios where a malicious session end user might leak sensitive information by taking pictures or screenshots of the session display. This policy can be enabled if some level of deterrence is prefered while sensitive information are displayed in the ICA session, and the administrator should be fully aware of the impact to user experience.

By default, session watermark is disabled.

Note:
Session watermark should not be regarded as a complete security protection mechanism by itself, combine it with other security solutions whenever it is possible.

Measures should be taken to ensure the integrity of Citrix components on VDA and client machines, otherwise session watermark may not function properly.

Effectiveness of watermark protection varies according to session conditions. It might not provide effective protection under certain usage scenario.

Watermark labels on pictures and screenshots taken from a session could be removed, distorted, modified or falsified using a variety of methods.

Watermark information are for reference only and should not be used for any legal purpose, nor should it be viewed as a reliable source for tracking data leakage.

Enabling session watermark can bring negative effect to the end user experience of the session. Adminstrator can adjust the watermark style related settings to achieve a balance between deterrence effect and user experience.

For a complete list of recommendations on using this feature, refer to Citrix Knowledge Base.
```

Setting Name: `EnableSessionWatermark`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable Unicode keyboard layout mapping
Description: 
```
Setting provides VDA-side support for Unicode keyboard layout mapping. This setting resolves incorrect input key mapping, which might occur in a non-Windows Citrix Workspace app. Incorrect input key mapping occurs because the mechanism of the non-Windows client keyboard layout is different with VDAs.
Note:
We are phasing out these registry settings in the Windows VDA -
HKEY_LOCAL_MACHINE\SOFTWARE\Citrix\CtxKlMap\EnableKlMap value = DWORD 1
HKEY_LOCAL_MACHINE\SOFTWARE\Citrix\CtxKlMap\DisableWindowHook value = DWORD 1
```

Setting Name: `EnableUnicodeKeyboardLayoutMapping`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable user layer compaction
Description: 
```
Enable user layer compaction to reclaim wasted disk space when user logs off.
```

Setting Name: `UserLayerCompactionEnabled`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable user-level policy settings
Description: 
```
Applies to: both file-based and container-based profile solutions

By default, Profile Management policies work at a machine level. With this policy enabled, Profile Management policies can work at a user level - it can be for an individual user or a user group.

When a session starts, Profile Management determines the policy settings to apply by prioritizing user settings over user group settings, and user group settings over machine settings.

If this setting is not configured here, the value from the .ini file is used.

If this setting is configured neither here nor in the .ini file, it is disabled.
```

Setting Name: `UserGroupLevelConfigEnabled`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable VHD auto-expansion for profile container
Description: 
```
Applies to: both file-based and container-based profile solutions

Specifies whether to enable VHD auto-expansion for the profile container. When enabled, all profile container auto-expansion settings apply to the profile container.

If this policy is not configured here, the value from the .ini file is used.

If this policy is not configured either here or in the .ini file, it is disabled.
```

Setting Name: `EnableVHDAutoExtend`

Setting Value:
```
{
    enabled = true | false
}
```

### Enable VHD disk compaction
Description: 
```
Applies to: both file-based and container-based profile solutions

Lets you specify whether to compact VHD disk files on user logoff when certain conditions are met.
Disk compaction can save space for central or cloud storage.

Depending on your needs and the resources available, you can adjust the default compaction settings and behavior using the following policies in Advanced settings:

- Free space ratio to trigger VHD disk compaction
- Number of logoffs to trigger VHD disk compaction
- Disable defragmentation for VHD disk compaction

If this policy is not configured here, the value from the .ini file is used.

If this policy is not configured here or in the .ini file, VHD disk compaction is not enabled.
```

Setting Name: `EnableVHDDiskCompaction`

Setting Value:
```
{
    enabled = true | false
}
```

### Enhanced Desktop Experience
Description: 
```
The Enhanced Desktop Experience feature configures Server OSs to deliver remote desktops which look as close as possible to Client OSs.

Note the following caveats:

If a user profile with Windows Classic theme already exists, then enabling this policy does not provide Enhanced Desktop Experience for that user.

If users with a user profile of Windows 7 theme logon to Windows Server 2012, for which this setting is either not configured or disabled, an error message is shown to the user. This error message indicates a failure to apply the theme.

In both above cases, resetting the user profile resolves the issue.

If this setting switches from enabled state to disabled state on a Server OS that has active user sessions, the look and feel for those sessions is inconsistent with Windows 7 and Windows Classic desktop experience.

After switching from enabled state to disabled state on a TS VDA, reboot the VDA. It is mandatory to delete the roaming profiles after such a switch. If other user profiles exist on the VDA, it is advisable to delete the same to avoid inconsistencies.

If roaming profiles are being used, all the Server OSs should have the Enhanced Desktop Experience feature either enabled or disabled site-wide.

The sharing of roaming profiles between Server OSs and Client OSs is discouraged as it causes inconsistency in the profile properties.
```

Setting Name: `EnhancedDesktopExperience`

Setting Value:
```
{
    enabled = true | false
}
```

### Enhanced domain passthrough for single sign on
Description: 
```
This feature allows leveraging delegated non-exportable Active Directory credentials for single sign-on into the virtual session.

The Windows policy setting "Remote host allows delegation of non-exportable credentials" must also be enabled on the VDA through Group Policy or Local Policy to use this feature. This setting is located under Computer Configuration > Administrative Templates > System > Credentials Delegation.

This feature can only be used with domain joined Windows endpoint devices.

By default, this setting is disabled.
```

Setting Name: `RemoteCredentialGuard`

Setting Value:
```
{
    enabled = true | false
}
```

### Estimate local time for legacy clients
Description: 
```
Enables or disables estimating the local time zone of client devices that send inaccurate time zone information to the server. By default, the server estimates the local time zone when necessary.
```

Setting Name: `LocalTimeEstimation`

Related Settings:
```
Use local time of client
```

Setting Value:
```
{
    enabled = true | false
}
```

### Excluded groups
Description: 
```
Applies to: both file-based and container-based profile solutions

You can use computer local groups and domain groups (local, global, and universal) to prevent particular user profiles from being processed. Specify domain groups in the form <DOMAIN NAME>\<GROUP NAME>.

If this setting is configured here, Profile management excludes members of these user groups.

If this setting is disabled, Profile management does not exclude any users.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, no members of any groups are excluded.
```

Setting Name: `ExcludedGroups_Part`

Setting Value:
```
jsonencode([
    {Group 1},
    {Group 2},
    ...
])
```

### Exclusion list
Description: 
```
Applies to: file-based profile solution only

List of registry keys in the HKCU hive that are ignored during logoff.

Example:
Software\Policies.

If this setting is disabled, no registry keys are excluded.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, no registry keys are excluded.
```

Setting Name: `ExclusionList_Part`

Setting Value:
```
jsonencode([
    {Registry Key 1},
    {Registry Key 2},
    ...
])
```

Example Setting Value:
```
jsonencode([
    "Software\Policies"
])
```

### Exclusion list - directories
Description: 
```
Applies to: file-based profile solution only

List of directories that are ignored during synchronization.

Folder names should be specified as paths relative to the user profile.

Wildcards are supported but they are not applied recursively.

Examples:
Desktop ignores the Desktop folder in the user profile.

Downloads\* ignores all immediate subfolders of the Downloads folder.

If this setting is disabled, no folders are excluded.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, no folders are excluded.
```

Setting Name: `ExclusionListSyncDir_Part`

Setting Value:
```
jsonencode([
    {Folder Path 1},
    {Folder Path 2},
    ...
])
```

### Exclusion list - files
Description: 
```
Applies to: file-based profile solution only

List of files that are ignored during synchronization.

File names should be specified as paths relative to the user profile. Wildcards are allowed. Wildcards in file names are applied recursively while wildcards in folder names are not.

Note: As of Profile Management 7.15, you can use the vertical bar '|' for applying a policy to only the current folder without propagating it to the subfolders.

Examples:
Desktop\Desktop.ini ignores the Desktop.ini file in the Desktop folder.

AppData\*.tmp ignores all files with the .tmp extension in the AppData folder and its subfolders.

AppData\*.tmp| ignores all files with the .tmp extension in the AppData folder.

Downloads\*\a.txt ignores a.txt in any immediate subfolder of the Downloads folder. Note that wildcards in folder names are not applied recursively.

If this setting is disabled, no files are excluded.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, no files are excluded.
```

Setting Name: `ExclusionListSyncFiles_Part`

Setting Value:
```
jsonencode([
    {File Path 1},
    {File Path 2},
    ...
])
```

### Extra color compression
Description: 
```
Extra color compression improves responsiveness over low bandwidth connections, by reducing the quality of displayed images.

For VDA versions 7.0 through 7.6 Feature Pack 2, this setting applies only when legacy graphics mode is enabled. For later VDA versions, this setting applies when legacy graphics mode is enabled, or when the legacy graphics mode is disabled and a video codec is not used to compress graphics.
```

Setting Name: `ExtraColorCompression`

Related Settings:
```
Use video codec for compression
```

Setting Value:
```
{
    enabled = true | false
}
```

### Favorites path
Description: 
```
Applies to: file-based profile solution only

Lets you specify how to redirect the Favorites folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRFavoritesPath_Part`

Setting Value: `{Path to the Redirected Favorite Folder}`

### FIDO2 Redirection
Description: 
```
Enables or disables FIDO2 Redirection. When enabled, users can perform FIDO2 Authentication using the local endpoint capabilities. By default, FIDO2 Authentication is enabled.
```

Setting Name: `AllowFidoRedirection`

Setting Value:
```
{
    enabled = true | false
}
```

### File redirection bandwidth limit
Description: 
```
Specifies the maximum allowed bandwidth in kilobits per second for accessing a client drive in a client connection.

If you enter a value for this setting and a value for the 'File redirection bandwidth limit percent' setting, the most restrictive setting (with the lower value) is applied.
```

Setting Name: `FileRedirectionBandwidthLimit`

Related Settings:
```
File redirection bandwidth limit percent

Client drive redirection
```

Setting Value: `{Bandwidth limit in Kbps}`

### File redirection bandwidth limit percent
Description: 
```
Specifies the maximum allowed bandwidth limit for accessing client drives as a percent of the total session bandwidth.

If you enter a value for this setting and a value for the 'File redirection bandwidth limit (Kbps)' setting, the most restrictive setting (with the lower value) is applied.

If you configure this setting, you must also configure the 'Overall session bandwidth limit' setting which specifies the total amount of bandwidth available for client sessions.
```

Setting Name: `FileRedirectionBandwidthPercent`

Related Settings:
```
File redirection bandwidth limit

Overall session bandwidth limit
```

Setting Value: `{Bandwidth Limit Percentage Value}`

### File system actions
```
Applies to: both file-based and container-based profile solutions

Detailed log settings.

Define events or actions that Profile management logs in depth.

If this setting is not configured here, Profile management uses the settings from the .ini file.

If this setting is not configured here or in the .ini file, errors and general information are logged.
```

Setting Name: `LogLevel_FileSystemActions`

Setting Value:
```
{
    enabled = true | false
}
```

### File system notifications
Description: 
```
Applies to: both file-based and container-based profile solutions

Detailed log settings.

Define events or actions that Profile management logs in depth.

If this setting is not configured here, Profile management uses the settings from the .ini file.

If this setting is not configured here or in the .ini file, errors and general information are logged.
```

Setting Name: `LogLevel_FileSystemNotification`

Setting Value:
```
{
    enabled = true | false
}
```

### File transfer for Citrix Workspace app for Chrome OS/HTML5
Description: 
```
This setting allows or prevents Citrix Workspace app for Chrome OS/HTML5 users from transferring files between a virtual desktop and the client device using the File Transfer virtual channel. By default, file transfers are allowed.

When adding this setting to a policy, ensure that you also enable the settings for uploading and downloading files for Citrix Workspace app for Chrome OS/HTML5. These settings are enabled by default. The maximum size of files that can be transferred is constrained by the amount of memory available to the browser.
```

Setting Name: `AllowFileTransfer`

Related Settings:
```
Upload file for Citrix Workspace app for Chrome OS/HTML5

Download file for Citrix Workspace app for Chrome OS/HTML5
```

Setting Value:
```
{
    enabled = true | false
}
```

### Files to exclude from the shared store
Description: 
```
Applies to: both file-based and container-based profile solutions

Lets you specify files to exclude from the shared store. Use this policy along with the "Files to include in the shared store for file deduplication" policy.

Starting with Profile Management version 2311, this policy applies to profile containers.

Specify the file names with paths relative to the user profile. Consider the following:

- Wildcards are supported.

- Wildcards in file names are applied recursively. To restrict them only to the current folder, use the vertical bar (|).

- Wildcards in folder names are not applied recursively.

Examples:

- Downloads\profilemgt_x64.msi specifies the profilemgt_x64.msi file in the Downloads folder.

- *.tmp specifies files with the .tmp extension in the user profile folder and its subfolders.

- AppData\*.tmp specifies files with the .tmp extension in the AppData folder and its subfolders.

- AppData\*.tmp| specifies files with the .tmp extension only in the AppData folder.

- Downloads\*\a.txt specifies the a.txt file in any immediate subfolder of the Downloads folder.

If this setting is disabled, no files are excluded.
If this setting is not configured here, the value from the .ini file is used.
If this setting is configured neither here nor in the .ini file, no files are excluded.
```

Setting Name: `SharedStoreFileExclusionList_Part`

Setting Value:
```
jsonencode([
    {File Path 1},
    {File Path 2},
    ...
])
```

### Files to exclude in profile container
Description: 
```
Applies to: both file-based and container-based profile solutions

List of files to exclude from the profile container.

Specify files to exclude in the form of relative paths to the user profile. You can use wildcards. Wildcards in file names are applied recursively while wildcards in folder names are not. You can use the vertical bar (|) to restrict the policy only to the current folder so that the policy does not apply to the subfolders.

If this setting is disabled, no files are excluded.

If this setting is not configured here, the value from the .ini file is used.

If this setting is configured neither here nor in the .ini file, no files are excluded.

Examples:
Desktop\Desktop.ini excludes the file Desktop\Desktop.ini from the profile container.

AppData\*.tmp excludes all files with the extension .tmp in the directory AppData and its subfolders.

AppData\*.tmp| excludes all files with the extension .tmp in the directory AppData.

Downloads\*\a.txt excludes a.txt in any immediate subfolder of the Downloads folder. Note that wildcards in folder names are not applied recursively.
```

Setting Name: `ProfileContainerExclusionListFile_Part`

Setting Value:
```
jsonencode([
    {File Path 1},
    {File Path 2},
    ...
])
```

### Files to include in profile container
Description: 
```
Applies to: both file-based and container-based profile solutions

List of files to include in the profile container when their parent folders are excluded from it.

Specify files to include in the form of relative paths to the user profile. You can use wildcards. Wildcards in file names are applied recursively while wildcards in folder names are not. You can use the vertical bar (|) to restrict the policy only to the current folder so that the policy does not apply to the subfolders.

Disabling this setting has the same effect as enabling it and configuring an empty list.

If this setting is not configured here, the value from the .ini file is used.

If this setting is configured neither here nor in the .ini file, folders not on the exclusion list are included in the profile container.

Examples:
AppData\Local\Microsoft\Office\Access.qat includes a file under a folder excluded in the default configuration.

AppData\Local\MyApp\*.cfg includes all files with the extension .cfg in the profile folder AppData\Local\MyApp and its subfolders.

AppData\Local\MyApp\*.cfg| includes all files with the extension .cfg in the profile folder AppData\Local\MyApp.

Downloads\*\b.txt includes b.txt in any immediate subfolder of the Downloads folder. Note that wildcards in folder names are not applied recursively.
```

Setting Name: `ProfileContainerInclusionListFile_Part`

Setting Value:
```
jsonencode([
    {File Path 1},
    {File Path 2},
    ...
])
```

### Files to include in the shared store for deduplication
Description: 
```
Applies to: both file-based and container-based profile solutions

Identical files can exist in various user profiles. Separating those files from the user store and storing them in a central location can save storage space. Doing so avoids duplicates. This policy lets you specify files that you want to include in the shared store on the server hosting the user store. With the policy enabled, Profile Management generates the shared store automatically. It then centrally stores the specified files in the shared store rather than in each user profile in the user store.

Starting with Profile Management version 2311, this policy applies to profile containers.

Specify the file names with paths relative to the user profile. Consider the following:

- Wildcards are supported.

- Wildcards in file names are applied recursively. To restrict them only to the current folder, use the vertical bar (|).

- Wildcards in folder names are not applied recursively.

Examples:

- Downloads\profilemgt_x64.msi specifies the profilemgt_x64.msi file in the Downloads folder.

- *.cfg specifies files with the .cfg extension in the user profile folder and its subfolders.

- Music\* specifies files in the Music folder and its subfolders.

- Downloads\*.iso specifies files with the .iso extension in the Downloads folder and its subfolders.

- Downloads\*.iso| specifies files with the .iso extension only in the Downloads folder.

- AppData\Local\Microsoft\OneDrive\*\*.dll specifies files with the .dll extension in any immediate subfolder of the AppData\Local\Microsoft\OneDrive folder.

If this setting is disabled, the shared store is disabled.
If this setting is not configured here, the value from the .ini file is used.
If this setting is configured neither here nor in the .ini file, the shared store is disabled.
```

Setting Name: `SharedStoreFileInclusionList_Part`

Setting Value:
```
jsonencode([
    {File Path 1},
    {File Path 2},
    ...
])
```

### Files to synchronize
Description: 
```
Applies to: file-based profile solution only

Profile management synchronizes each user's entire profile between the system it is installed on and the user store.

This setting allows for the inclusion of files below excluded folders.

Paths on this list should be relative to the user profile. Wildcards are allowed. Wildcards in file names are applied recursively while wildcards in folder names are not.

Note: As of Profile Management 7.15, you can use the vertical bar '|' for applying a policy to only the current folder without propagating it to the subfolders.

Examples:
AppData\Local\Microsoft\Office\Access.qat specifies a file below a folder excluded in the default configuration.

AppData\Local\MyApp\*.cfg specifies all files with the extension .cfg in the profile folder AppData\Local\MyApp and its subfolders.

AppData\Local\MyApp\*.cfg| specifies all files with the extension .cfg in the profile folder AppData\Local\MyApp.

Downloads\*\b.txt specifies b.txt in any immediate subfolder of the Downloads folder. Note that wildcards in folder names are not applied recursively.

Disabling this setting has the same effect as enabling it and configuring an empty list.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, only non-excluded files in the user profile are synchronized.
```

Setting Name: `SyncFileList_Part`

Setting Value:
```
jsonencode([
    {File Path 1},
    {File Path 2},
    ...
])
```

### Final force logoff message box body text
Description: 
```
Text displayed in the message box that alerts users when a forced logoff is already in progress
The text of the message body must be supplied and must be 3072 characters or less.
```

Setting Name: `FinalForceLogoffMessageBody`

Setting Value: `{Text Message to Display}`

### Final force logoff message box title
Description: 
```
Window caption of the message box that alerts users of an impending forced logoff
The title must be supplied and must be 80 characters or less.
```

Setting Name: `FinalForceLogoffMessageTitle`

Setting Value: `{Title to Display}`

### Folders to exclude in profile container
Description: 
```
Applies to: both file-based and container-based profile solutions

List of folders to exclude from the profile container.

Enter folders to exclude as relative paths to the user profile.

Wildcards are supported but they are not applied recursively.

Examples:
Desktop specifies the Desktop folder.

Downloads\* specifies all immediate subfolders of the Downloads folder.

If this setting is disabled, no folder is excluded.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, no folder is excluded.
```

Setting Name: `ProfileContainerExclusionListDir_Part`

Setting Value:
```
jsonencode([
    {Folder Path 1},
    {Folder Path 2},
    ...
])
```

### Folders to include in profile container
Description: 
```
Applies to: both file-based and container-based profile solutions

List of folders to keep in the profile container when their parent folders are excluded.

Folders on this list must be subfolders of the excluded folders. Otherwise, this setting does not work.

Wildcards are supported but they are not applied recursively.

Disabling this setting has the same effect as enabling it and configuring an empty list.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, folders not on the exclusion list are included in the profile container.
```

Setting Name: `ProfileContainerInclusionListDir_Part`

Setting Value:
```
jsonencode([
    {Folder Path 1},
    {Folder Path 2},
    ...
])
```

### Folders to mirror
Description: 
```
Applies to: file-based profile solution only

Profile management can mirror a folder relative to the profile's root folder. Use this setting for files whose contents index data and where separate instances of the data are likely to exist. For example, you can mirror the Internet Explorer cookies folder so that index.dat is synchronized with the cookies that it indexes. Be aware that, in these situations the "last write wins" so files in mirrored folders that have been modified in more than one session are overwritten by the last update, resulting in loss of profile changes.
```

Setting Name: `MirrorFoldersList_Part`

Setting Value:
```
jsonencode([
    {Folder Path 1},
    {Folder Path 2},
    ...
])
```

### Force logoff grace period
Description: 
```
For Manual and PVS Server OS catalogs, The Connector Agent service for Configuration Manager 2012 will eventually force all outstanding user sessions to logoff during a maintenance window when there is at least one pending application install and/or software update whose deadline has expired.

This setting specifies how much time user/s should be given to save their work upon being notified (via message box) of an impending forced logoff of all sessions.

A time span string consists of the following:

ddd . Days (from 0 to 999) [optional]
hh : Hours (from 0 to 23)
mm : Minutes (from 0 to 59)
ss Seconds (from 0 to 59)
```

Setting Name: `ForceLogoffGracePeriod`

Setting Value: `{ddd.hh:mm:ss where ddd. is optional}`

Example Setting Value:
```
00:05:00
```

### Force logoff message box body text
Description: 
```
Text displayed in the message box that alerts users when a forced logoff is already in progress
```

Setting Name: `ForceLogoffMessageBody`

Setting Value: `{Message to Display}`

Example Message:
```
{TIMESTAMP}Please save your work. The system will go offline for maintenance in {TIMELEFT}
```

### Force logoff message box title
Description: 
```
Window caption of the message box that alerts users of an impending forced logoff
The title must be supplied and must be 80 characters or less.
```

Setting Name: `ForceLogoffMessageTitle`

Setting Value: `{Title to Display}`

Example Setting Value:
```
Notification From IT Staff
```

### Free space ratio to trigger VHD disk compaction
Description: 
```
Applies to: both file-based and container-based profile solutions

Lets you specify the free space ratio to trigger VHD disk compaction. When the free space ratio exceeds the specified value on user logoff, disk compaction is triggered.

Free space ratio = (current VHD file size - required minimum VHD file size*) / current VHD file size

* Obtained using the GetSupportedSize method of the MSFT_Partition class from the Microsoft Windows operating system.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, the default value 20 (%) is used.
```

Setting Name: `FreeRatio4Compaction_Part`

Setting Value: `{Free Space Ratio}`

### Grant administrator access
Description: 
```
Applies to: file-based profile solution only

By default, users are granted exclusive access to the contents of their redirected folders.

Enabling this option allows administrator access to the contents of the users redirected folders.
```

Setting Name: `FRAdminAccess_Part`

Setting Value:
```
{
    enabled = true | false
}
```

### Graphic status indicator
Description: 
```
This setting configures the graphics status indicator to run in the user session. This tool lets the user see information about the active graphics mode, including details about video codec, hardware encoding, image quality, and monitors in use for the session. With the graphics status indicator, the user can also enable or disable pixel perfect mode.

Releases of CVAD 2103 and later include an image quality slider to help the user find the right balance between image quality and interactivity.

Releases of CVAD 2109 and later include functionality to configure a virtual display layout through a user interface launched using the graphics status indicator context menu option "Configure virtual displays".

The graphics status indicator replaces the lossless indicator tool from previous versions. This policy enables the lossless indicator for versions 7.16 to 1809.
```

Setting Name: `DisplayLosslessIndicator`

Related Settings:
```
Visual quality

Allow visually lossless compression.
```

Setting Value:
```
{
    enabled = true | false
}
```

### Groups using customized user layer size
Description: 
```
Active Directory groups that will use Customized User Layer Size (in GB). You can have multiple groups. Groups should have the format of <Domain>\<Group>
```

Setting Name: `UplGroupsUsingCustomizedUserLayerSize`

Setting Value:
```
jsonencode([
    {Group 1},
    {Group 2},
    ...
])
```

### HDX adaptive transport
Description: 
```
Adaptive transport is a network-aware data transport engine that provides efficient, reliable, and consistent congestion and flow control.

By default, adaptive transport is set to Preferred, data transport takes place over a proprietary transport protocol, Enlightened Data Transport (EDT), that is built on top of UDP, with automatic fallback to TCP. Additional configuration is not required to optimize for LAN, WAN, or Internet conditions. Citrix's transport protocol responds to changing conditions.

When set to Off, adaptive transport is disabled and TCP is used. Recommended when using SD-WAN WAN optimization, which provides cross-session tokenized compression, since WAN optimization has its own congestion and flow control.

Setting Diagnostic mode forces EDT on and disables fallback to TCP. Recommended for testing purposes only.

None of these settings affects other services that depend on UDP transport, such as UDP Audio and Framehawk.
```

Setting Name: `HDXAdaptiveTransport`

Setting Value: `Off`, `Preferred`, `DiagnosticMode`

### HDX Direct
Description: 
```
HDX Direct allows the client to automatically establish a direct connection with the session host when direct communication is available. Connections are established securely using network-level encryption.
Please refer to the documentation for system requirements and additional details.
```

Setting Name: `HDXDirect`

Setting Value:
```
{
    enabled = true | false
}
```

### HDX Direct mode
Description: 
```
HDX Direct can be used to establish direct connections with session hosts for internal and external clients. This setting determines if HDX Direct is available for internal clients only or for both internal and external clients.

When set to Internal only, HDX Direct will attempt to establish direct connections for clients in the internal network only.

When set to Internal and external, HDX Direct will attempt to establish direct connections for internal and external clients.

By default, HDX Direct is set for internal clients only.
```

Setting Name: `HDXDirectMode`

Related Settings:
```
HDX Direct
```

Setting Value: `InternalOnly`, `InternalAndExternal`

### HDX Direct port range
Description: 
```
Range of ports that will be used by HDX Direct for connections from external users.
Enter a range in the format (low port),(high port).
By default, HDX Direct will use the port range 55000,55250.
```

Setting Name: `HDXDirectPortRange`

Setting Value: `{Low Port},{High Port}`

### HDX MediaStream Multimedia Acceleration bandwidth limit
Description: 
```
Specifies the maximum allowed bandwidth in kilobits per second to deliver streaming audio and video to users using HDX MediaStream Multimedia Acceleration.

If you enter a value for this setting and a value for the 'HDX MediaStream Multimedia Acceleration bandwidth limit percent' setting, the most restrictive setting (with the lower value) is applied.
```

Setting Name: `HDXMultimediaBandwidthLimit`

Related Settings:
```
HDX MediaStream Multimedia Acceleration bandwidth limit percent

Overall session bandwidth limit
```

Setting Value: `{Bandwidth Limit in Kbps}`

### HDX MediaStream Multimedia Acceleration bandwidth limit percent
Description: 
```
Specifies the maximum allowed bandwidth, as a percent of the total session bandwidth, to deliver streaming audio and video to users using HDX MediaStream Multimedia Acceleration.

If you enter a value for this setting and a value for the 'HDX MediaStream Multimedia Acceleration bandwidth limit' setting, the most restrictive setting (with the lower value) is applied.

If you configure this setting, you must also configure the 'Overall session bandwidth limit' setting which specifies the total amount of bandwidth available for client sessions.
```

Setting Name: `HDXMultimediaBandwidthPercent`

Related Settings:
```
HDX MediaStream Multimedia Acceleration bandwidth limit

Overall session bandwidth limit
```

Setting Value: `{Bandwidth Percentage Value}`

### Hide keyboard layout switch pop-up message box
Description: 
```
A message box pops up to notify the user that the keyboard layout is synchronizing when the user changes the client keyboard layout. This message occurs when the “Client keyboard layout sync and IME improvement” group policy enables client keyboard synchronization. This setting hides the pop-up message box.
Note:
We are phasing out the registry setting in the Windows VDA -
HKEY_LOCAL_MACHINE\Software\Citrix\lcalme\HideNotificationWindow value = DWORD 1
```

Setting Name: `HideKeyboardLayoutSwitchPopupMessageBox`

Related Settings:
```
Client keyboard layout synchronization and IME improvement
```

Setting Value:
```
{
    enabled = true | false
}
```

### Host to client redirection
Description: 
```
Enables or disables file type associations for URLs and some media content to be opened on the client device. By default, file type association is disabled.

These URL types are opened locally when you enable this setting:
* Hypertext Transfer Protocol (HTTP)
* Secure Hypertext Transfer Protocol (HTTPS)
* Real Player and QuickTime (RTSP)
* Real Player and QuickTime (RTSPU)
* Legacy Real Player (PNM)
* Microsoft's Media Format (MMS)
```

Setting Name: `HostToClientRedirection`

Setting Value:
```
{
    enabled = true | false
}
```

### HTML5 video redirection
Description: 
```
Controls and optimizes the way Virtual Apps and Virtual Desktops servers deliver HTML5 multimedia web content to users. By default, this setting is prohibited.

To use HTML5 video redirection, make sure the Windows Media Redirection setting is enabled.
```

Setting Name: `HTML5VideoRedirection`

Related Settings:
```
Windows Media Redirection
```

Setting Value:
```
{
    enabled = true | false
}
```

### ICA keep alive timeout
Description: 
```
Seconds between successive ICA keep-alive messages. By default, the interval between keep-alive messages is 60 seconds.

Specify an interval between 1-3600 seconds in which to send ICA keep-alive messages. Do not configure this setting if your network monitoring software is responsible for closing inactive connections. If using Citrix NetScaler Gateway, set keep-alive intervals on the NetScaler Gateway to match the keep-alive intervals on XenApp.
```

Setting Name: `IcaKeepAliveTimeout`

Setting Value: `{Timeout in Seconds}`

### ICA keep alives
Description: 
```
Sends or prevents sending ICA keep-alive messages periodically. By default, keep-alive messages are not sent.

Enabling this setting prevents broken connections from being disconnected. If XenApp detects no activity, this setting prevents Terminal Services from disconnecting the session. XenApp sends keep-alive messages every few seconds to detect if the session is active. If the session is no longer active, XenApp marks the session as disconnected.

ICA Keep-Alive does not work if you are using Session Reliability. Configure ICA Keep-Alive only for connections that are not using Session Reliability.
```

Setting Name: `IcaKeepAlives`

Setting Value: `DoNotSendKeepAlives`, `SendKeepAlives`

### ICA listener connection timeout
Description: 
```
Maximum wait time for a connection using the ICA protocol to be completed. By default, the maximum wait time is 120000 milliseconds, or two minutes.
```

Setting Name: `IcaListenerTimeout`

Setting Value: `{Timeout in Milliseconds}`

### ICA listener port number
Description: 
```
The TCP/IP port number used by the ICA protocol on the server.

The default port number is 1494. The port number must be in the range of 0-65535 and must not conflict with other well-known port numbers.

If you change the port number, restart the server for the new value to take effect. If you change the port number on the server, you must also change it on every plug-in that connects to the server.
```

Setting Name: `IcaListenerPortNumber`

Setting Value: `{Port Number}`

### ICA round trip calculation
Description: 
```
Enables or disables the calculation of ICA round trip measurements. By default, ICA round trip calculations are allowed.

The ICA round trip is the time interval, measured at the client device, between the first action a user takes (such as typing the letter 'A') and the result of that action (such as the graphical display of the letter 'A').

By default, each ICA roundtrip measurement initiation is delayed until some traffic occurs that indicates user interaction. This delay can be indefinite in length and is designed to prevent the ICA roundtrip measurement being the sole reason for ICA traffic. To perform round trip calculations regardless of user activity level, configure the 'ICA round trip calculations for idle connections' setting.

To set the interval at which calculations occur, configure the 'ICA round trip calculation interval' setting.
```

Setting Name: `IcaRoundTripCalculation`

Setting Value:
```
{
    enabled = true | false
}
```

### ICA round trip calculation interval
Description: 
```
The frequency, in seconds, at which ICA round trip calculations are performed. By default, ICA round trip is calculated every 15 seconds.
```

Setting Name: `IcaRoundTripCalculationInterval`

Setting Value: `{Interval in Seconds}`


### ICA round trip calculations for idle connections
Description: 
```
Determines whether ICA round trip calculations are performed for idle connections. By default, calculations are not performed for idle connections.

By default, each ICA roundtrip measurement initiation is delayed until some traffic occurs that indicates user interaction. This delay can be indefinite in length and is designed to prevent the ICA roundtrip measurement being the sole reason for ICA traffic.
```

Setting Name: `IcaRoundTripCalculationWhenIdle`

Setting Value:
```
{
    enabled = true | false
}
```

### Image-managed mode
Description: 
```
The Connector Agent will automatically detect if it's running on a PVS or MCS image-managed clone. The Agent blocks Configuration Manager updates on image-managed clones and automatically installs the updates on the master image of the catalog. After the master image is updated, use Studio to orchestrate the reboot of MCS catalog clones. The Connector Agent will automatically orchestrate the reboot of PVS catalog clones during Configuration Manager maintenance windows. By setting this to Disabled, you can override this behavior so that software is installed on catalog clones by Configuration Manager.
```

Setting Name: `ImageProviderIntegrationEnabled`

Setting Value:
```
{
    enabled = true | false
}
```

### Include client IP address
Description: 
```
Controls the content of session watermark text.

This policy is only effective when session watermark is enabled.

When this policy is enabled, the client IP address of the current ICA session will be displayed in the watermark text labels.

By default, this setting is disabled.

Note: To achieve better user experience, it is suggested that no more than 2 watermark text items are selected at the same time.
```

Setting Name: `WatermarkIncludeClientIPAddress`

Setting Value:
```
{
    enabled = true | false
}
```

### Include connection time
Description: 
```
Controls the content of session watermark text labels.

This policy is only effective when session watermark is enabled.

When this policy is enabled, connect time will be displayed in the watermark text labels in the format of yyyy/mm/dd hh:mm. The time displayed is based on the client system clock and time zone.

By default, this setting is disabled.

Note: To achieve better user experience, it is suggested that no more than 2 watermark text items are selected at the same time.
```

Setting Name: `WatermarkIncludeConnectTime`

Setting Value:
```
{
    enabled = true | false
}
```

### Include domain name
Description: 
```
Applies to: file-based profile solution only

Enabling this option will include the %userdomain% environment variable as part of the UNC path.
```

Setting Name: `FRIncDomainName_Part`

Setting Value:
```
{
    enabled = true | false
}
```

### Include logon user name
Description: 
```
Controls the content of session watermark text labels.

This policy is only effective when session watermark is enabled.

When this policy is enabled, the logon user name of the current ICA session will be displayed in the watermark text labels in the USERNAME@DOMAINNAME format.

By default, this setting is enabled.

Note:

It is strongly suggested that the logon user name should contain no more than 20 characters; otherwise, excessively small character fonts or truncation might occur and watermark effectiveness is therefore affected.

To achieve better user experience, it is suggested that no more than 2 watermark text items are selected at the same time.
```

Setting Name: `WatermarkIncludeLogonUsername`

Setting Value:
```
{
    enabled = true | false
}
```

### Include VDA host name
Description: 
```
Controls the content of session watermark text labels.

This policy is only effective when session watermark is enabled.

When this policy is enabled, the VDA hostname of the current ICA session will be displayed in the watermark text labels.

By default, this setting is enabled.

Note:

To achieve better user experience, it is suggested that no more than 2 watermark text items are selected at the same time.
```

Setting Name: `WatermarkIncludeVDAHostName`

Setting Value:
```
{
    enabled = true | false
}
```

### Include VDA IP address
Description: 
```
Controls the content of session watermark text.

This policy is only effective when session watermark is enabled.

When this policy is enabled, the VDA IP address of the current ICA session will be displayed in the watermark text labels.

By default, this setting is disabled.

Note: To achieve better user experience, it is suggested that no more than 2 watermark text items are selected at the same time.
```

Setting Name: `WatermarkIncludeVDAIPAddress`

Setting Value:
```
{
    enabled = true | false
}
```

### Inclusion list
Description: 
```
Applies to: file-based profile solution only

List of registry keys in the HKCU hive that are processed during logoff. Example: Software\Adobe.

If this setting is enabled, only keys on this list are processed.

If this setting is disabled, the complete HKCU hive is processed.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, all of HKCU is processed.
```

Setting Name: `IncludeListRegistry_Part`

Setting Value:
```
jsonencode([
    {Registry Key 1},
    {Registry Key 2},
    ...
])
```

Example Setting Value:
```
jsonencode([
    "Software\Adobe"
])
```
### Large File Handling List - Files to be created as symbolic links
Description: 
```
Applies to: file-based profile solution only

To improve logon performance and to process large size files, a symbolic link is created instead of copying files in this list.

You can use wildcards in policies that refer to files. For example, !ctx_localappdata!\Microsoft\Outlook\*.OST.

To process the offline folder file (*.ost) of Microsoft Outlook, make sure that the Outlook folder is not excluded for Citrix Profile Management.

Note that those files cannot be accessed in multiple sessions simultaneously.
```

Setting Name: `LargeFileHandlingList_Part`

Setting Value:
```
jsonencode([
    {File Path 1},
    {File Path 2},
    ...
])
```

### Launch touch-optimized desktop
Description: 
```
Enables or disables the launching of a touch-optimized desktop for mobile clients. By default, launching a touch-optimized desktop is enabled, except for Windows 10 and Windows Server 2016 machines. This setting is disabled and deprecated for Windows 10 and Windows Server 2016 machines.
```

Setting Name: `MobileDesktop`

Setting Value:
```
{
    enabled = true | false
}
```

### Launching of non-published programs during client connection
Description: 
```
Specifies whether to launch initial applications or published applications through ICA or RDP on the server. By default, only published applications are allowed to launch.
```

Setting Name: `NonPublishedProgramLaunching`

Setting Value:
```
{
    enabled = true | false
}
```

### Limit clipboard client to session transfer size
Description: 
```
The value set will be the maximum size (kilobytes) that can be transferred into a session during a single copy/paste operation for non-file types (i.e. raw text, bitmaps, etc.).
```

Setting Name: `LimitClipboardTransferC2H`

Related Settings:
```
Limit clipboard session to client transfer size
```

Setting Value: `{Size Limit in Kilobytes}`

### Limit clipboard session to client transfer size
Description: 
```
The value set will be the maximum size (kilobytes) that can be transferred out of a session during a single copy/paste operation for non-file types (i.e. raw text, bitmaps, etc.).
```

Setting Name: `LimitClipboardTransferH2C`

Related Settings:
```
Limit clipboard client to session transfer size
```

Setting Value: `{Size Limit in Kilobytes}`

### Limit video quality
Description: 
```
Limits video quality for HDX connection to specified value
```

Setting Name: `VideoQuality`

Related Settings:
```
Windows Media redirection

Optimization for Windows Media multimedia redirection over WAN
```

Setting Value: `Unconfigured`, `P1080`, `P720`, `P480`, `P280`, `P240`

Setting Value Explanation:
Setting Value | Explanation
-- | --
`Unconfigured` | Not Configured
`P1080` | Maximum Video Quality 1080p/8.5Mbps
`P720` | Maximum Video Quality 720p/4.0Mbps
`P480` | Maximum Video Quality 480p/720Kbps
`P280` | Maximum Video Quality 380p/400Kbps
`P240` | Maximum Video Quality 240p/200Kbps

### Links path
Description: 
```
Applies to: file-based profile solution only

Lets you specify how to redirect the Links folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRLinksPath_Part`

Setting Value: `{Path to the Redirected Links Folder}`

### List of applications excluded from failure monitoring
Description: 
```
Use this setting to configure a comma-separated list of processes whose faults are to be ignored. This helps filter out applications that are not managed or those with known issues.
```

Setting Name: `AppFailureExclusionList`

Setting Value: `{Process 1},{Process 2},...`

### Local profile conflict handling
Description: 
```
Applies to: file-based profile solution only

This setting configures what Profile management does if both a profile in the user store and a local Windows user profile (not a Citrix user profile) exist.

`Use`: If this setting is disabled or set to the default value of "Use local profile", Profile management uses the local profile, but does not change it in any way.

`Delete`: If this setting is set to "Delete local profile", Profile management deletes the local Windows user profile, and then imports the Citrix user profile from the user store.

`Rename`: If this setting is set to "Rename local profile", Profile management renames the local Windows user profile (in order to back it up) and then imports the profile from the user store.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, existing local profiles are used.
```

Setting Name: `LocalProfileConflictHandling_Part`

Setting Value: `Use`, `Delete`, `Rename`

### Log off user if a problem is encountered
Description: 
```
Applies to: both file-based and container-based profile solutions

If a problem is encountered when the user logs on (for example, if the user store is unavailable) a temporary profile is provided.

Alternatively, enabling this setting displays an error message and logs the user off.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, a temporary profile is provided.
```

Setting Name: `LogoffRatherThanTempProfile`

Setting Value:
```
{
    enabled = true | false
}
```

# Log off users when profile container is not available during logon
Description: 
```
Applies to: container-based profile solution only

Specifies whether to log off users when the profile container is not available during logon. By default, users can log on using temporary profiles when the profile container is not available. Alternatively, enabling this setting displays an error message to users and logs them off after they click OK. Enter the error message to display. Leaving it empty will display a default message.

If this policy is not configured here, the setting from the .ini file is used.

If this policy is not configured either here or in the .ini file, the setting is disabled.
```

Setting Name: `PreventLoginWhenMountFailed_Part`

Setting Value: `{Error Message to Display}`

### Logoff
Description: 
```
Applies to: both file-based and container-based profile solutions

Detailed log settings.

Define events or actions that Profile management logs in depth.

If this setting is not configured here, Profile management uses the settings from the .ini file.

If this setting is not configured here or in the .ini file, errors and general information are logged.
```

Setting Name: `LogLevel_Logoff`

Setting Value:
```
{
    enabled = true | false
}
```

### Logoff checker startup delay
Description: 
```
This setting specifies the duration to delay the logoff checker startup.
Use this setting to set the length of time(in seconds) a client session waits before disconnecting the session.

Note: Setting this value also increases the time it takes for a user to log off the server.
```

Setting Name: `LogoffCheckerStartupDelay`

Setting Value: `{Delay in Seconds}`

### Logon
Description: 
```
Applies to: both file-based and container-based profile solutions

Detailed log settings.

Define events or actions that Profile management logs in depth.

If this setting is not configured here, Profile management uses the settings from the .ini file.

If this setting is not configured here or in the .ini file, errors and general information are logged.
```

Setting Name: `LogLevel_Logon`

Setting Value:
```
{
    enabled = true | false
}
```

### Logon Exclusion Check
Description: 
```
Applies to: file-based profile solution only

This setting configures what Profile Management does if a profile in the user store contains excluded files or folders.

`Disable`: If this setting is disabled or set to the default value of "Synchronize excluded files or folders on logon," Profile Management will synchronize these excluded files or folders from the user store to local profile when a user logs on.

`Ignore`: If this setting is set to "Ignore excluded files or folders on logon," Profile Management ignores the excluded files or folders in the user store when a user logs on.

`Delete`: If this setting is set to "Delete excluded files or folder on logon," Profile Management deletes the excluded files or folders in the user store when a user logs on.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, the excluded files or folders are synchronized from the user store to local profile when a user logs on.
```

Setting Name: `LogonExclusionCheck_Part`

Setting Value: `Disable`, `Ignore`, `Delete`

### Loss tolerant mode
Description: 
```
Loss tolerant mode is designed to maintain usable session interactivity over a network connection with high packet loss and latency.

To maintain an interactive session, it trades off both image quality and bandwidth usage.

When available, the mode is entered only when packet loss and latency are above a threshold, which by default is set high enough to only be triggered when incurring the overhead is necessary to maintain interactivity. The default threshold can be overridden by using loss tolerant mode thresholds.
```

Setting Name: `LossTolerantModeAvailable`

Setting Value:
```
{
    enabled = true | false
}
```

### Loss tolerant mode for audio
Description: 
```
Loss tolerant mode for audio is optimized for real-time communication and performs better over networks with packet loss and high latency.

The HDX adaptive transport setting must be set to Preferred for this setting to be applied.
```

Setting Name: `LossTolerantAudio`

Setting Value:
```
{
    enabled = true | false
}
```

### Loss tolerant mode thresholds
Description: 
```
Specifies the network metrics thresholds at which the session switches to loss tolerant mode if it is available. The session will go into loss tolerant mode and remain there as long as each of the thresholds is exceeded.

The default thresholds are as follows:
• Packet loss threshold: 5 %
• Latency threshold: 300 ms rtt
```

Setting Name: `LossTolerantThresholds`

Setting Value: `loss,{Loss Threshold};latency,{Latency Threshold};`

### LPT port redirection bandwidth limit
Description: 
```
For VDA versions that do not support this setting, follow CTX139345 to enable redirection using the registry.

Specifies the maximum allowed bandwidth in kilobits per second for print jobs using an LPT port in a single client session.

If you enter a value for this setting and a value for the 'LPT port redirection bandwidth limit percent' setting, the most restrictive setting (with the lower value) is applied.
```

Setting Name: `LptBandwidthLimit`

Related Settings:
```
LPT port redirection bandwidth limit percent

Client LPT port redirection
```

Setting Value: `{Bandwidth Limit in Kbps}`

### LPT port redirection bandwidth limit percent
Description: 
```
For VDA versions that do not support this setting, follow CTX139345 to enable redirection using the registry.

Specifies the bandwidth limit for print jobs using an LPT port in a single client session as a percent of the total session bandwidth.

If you enter a value for this setting and a value for the 'LPT port redirection bandwidth limit (Kbps)' setting, the most restrictive setting (with the lower value) is applied.

If you configure this setting, you must also configure the 'Overall session bandwidth limit' setting which specifies the total amount of bandwidth available for client sessions.
```

Setting Name: `LptBandwidthLimitPercent`

Related Settings:
```
LPT port redirection bandwidth limit

Overall session bandwidth limit

Client LPT port redirection
```

Setting Value: `{Bandwidth Limit Percentage Value}`

### Max Speex quality
Description: 
```
This setting is supported only by Linux Virtual Desktop 1.4 and onwards.

Audio redirection encodes audio data with Speex codec when the Audio quality setting is set to medium or low. Speex is a lossy codec, which means that it achieves compression at the expense of fidelity of the input speech signal. Unlike some other speech codecs, it is possible to control the tradeoff made between quality and bit-rate. The Speex encoding process is controlled most of the time by a quality parameter that ranges from 0 to 10. The higher the quality, the higher the bit-rate.

This setting sets the best Speex quality to encode audio data specified by the Audio redirection bandwidth limit setting. If the Audio quality setting is set to medium, the encoder is in wide band mode, which means higher sampling rate. If the Audio quality setting is set to low, the encoder is in narrow band mode, which means lower sampling rate. The same Speex quality has different bit-rate in different mode. The best Speex quality is the largest value that meets the following conditions:

* It is equal to or less than the max Speex quality
* Its bit-rate is equal to or less than the bandwidth limit
```

Setting Name: `MaxSpeexQuality`

Setting Value: `{Max Speex Quality}`

### Maximum allowed color depth
Description: 
```
This setting applies only when legacy graphics mode is enabled.

Specifies the maximum color depth allowed for a session. By default, the maximum allowed color depth is 32 bits per pixel.
```

Setting Name: `MaximumColorDepth`

Setting Value: `BitsPerPixel8`, `BitsPerPixel15`, `BitsPerPixel8`, `BitsPerPixel16`, `BitsPerPixel24`, `BitsPerPixel32`

### Maximum number of sessions
Description: 
```
Specifies the maximum number of sessions a server is allowed to host.
```

Setting Name: `MaximumNumberOfSessions`

Setting Value: `{Maximum Number of Sessions Allowed}`

### Maximum number of VHDX disks for storing Outlook OST files
Description: 
```
Applies to: both file-based and container-based profile solutions

Provides native Outlook search experience in concurrent sessions. Use this policy with the Search index roaming for Outlook policy.

With this policy enabled, each concurrent session uses a separate Outlook OST file.

By default, only two VHDX disks can be used to store Outlook OST files (one file per disk). If more sessions start, their Outlook OST files are stored in the local user profiles. You can specify the maximum number of VHDX disks for storing Outlook OST files.

Example:
You set the number to 3. As a result, Profile Management stores Outlook OST files on the VHDX disks for the first, second, and third concurrent sessions. It stores the OST file for the fourth concurrent session in the local profile instead.
```

Setting Name: `OutlookSearchRoamingConcurrentSession_Part`

Setting Value: `{Maximum VHDX Disks Count}`

### Maximum size of the log file
Description: 
```
Applies to: both file-based and container-based profile solutions

Sets the maximum size of the log file in bytes. If the log file grows beyond this size an existing backup of the file (.bak) is deleted, the log file is renamed to .bak, and a new log file is created.

The log file is created in "%SystemRoot%\System32\Logfiles\UserProfileManager" or in the location specified by the "Path to log file" policy setting.

If this setting is disabled, the default value of 10 MB is used.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, the default of 10 MB is used.
```

Setting Name: `MaxLogSize_Part`

Setting Value: `{Maximum Log File Size in Bytes}`

### Memory usage
Description: 
```
Defines the memory usage percentage value at which the server reports full load.
```

Setting Name: `MemoryUsage`

Setting Value: `{Memory Usage Percentage Value to Report Full Load}`

### Memory usage base load
Description: 
```
A significant portion of memory may be required for base operating system functions, i.e. before any sessions have started. This setting is an approximation of the base operating system's memory usage and defines the memory usage in MBs below which a server is considered to have zero load.

If the Memory Usage setting is disabled, then this setting will be ignored irrespective of its configuration.

In most installations the default value should be adequate. However, if the server has limited memory, for example 2GBs in a Proof of Concept environment, the value may need to be tuned to more closely reflect the memory usage of the operating system.
```

Setting Name: `MemoryUsageBaseLoad`

Setting Value: `{Memory Usage Value in MBs to Report Zero Load}`

### Menu animation
Description: 
```
Allows or prevents menu animation. By default, menu animation is allowed.

Menu animation is a Microsoft personal preference setting that causes a menu to appear after a short delay, either by scrolling or fading in. When allowed, an arrow icon appears at the bottom of the menu; the menu appears when you mouse over that arrow.
```

Setting Name: `MenuAnimation`

Setting Value:
```
{
    enabled = true | false
}
```

### Microsoft Teams redirection
Description: 
```
Controls and optimizes the way Citrix Virtual Apps and Desktops servers deliver Microsoft Teams multimedia content to users.

Only multimedia content is redirected to the user's client machine, where it is decoded locally, effectively offloading all CPU, RAM, GPU, I/O, and bandwidth processing from the VDA to the endpoint.

In addition to this policy, the appropriate version of Citrix Workspace app is required for Microsoft Teams redirection to occur.

For more information and troubleshooting, see Knowledge Center article CTX253754.

When clicking on Help / Report a Problem in Microsoft Teams, logs will be automatically shared between Citrix and Microsoft in order to resolve technical issues.
```

Setting Name: `MSTeamsRedirection`

Setting Value:
```
{
    enabled = true | false
}
```

### Migrate user store
Description: 
```
Applies to: file-based profile solution only

Specifies the path to the folder where the user settings (registry changes and synchronized files) were previously saved (the user store path that you previously used).

If this setting is configured, the user settings stored in the previous user store are migrated to the current user store specified in the "Path to user store" policy setting.

The path can be an absolute UNC path or a path relative to the home directory.

In both cases, you can use the following types of variables: system environment variables enclosed in percent signs and attributes of the Active Directory user object enclosed in hash signs.

Examples:
The folder Windows\%ProfileVer% stores the user settings in the subfolder called Windows\W2k3 of the user store (if %ProfileVer% is a system environment variable resolving to W2k3).

\\server\share\#SAMAccountName# stores the user settings to the UNC path \\server\share\JohnSmith (if #SAMAccountName# resolves to JohnSmith for the current user).

You can use user environment variables except %username% and %userdomain%.

If this setting is disabled, the user settings are still saved in the current user store.

If this setting is not configured here, the setting from the .ini file is used.

If this setting is not configured here or in the .ini file, the user settings are still saved in the current user store.

If the path you configured here points to the same location as the "Path to user store" policy setting, this policy setting does not take effect.
```

Setting Name: `MigrateUserStore_Part`

Setting Value: `{Folder Path Where User Settings Were Saved}`

### Migration of existing profiles
Description: 
```
Applies to: file-based profile solution only

Profile management can migrate existing profiles "on the fly" during logon if the user has no profile in the user store.

The following event takes place during logon: if an existing Windows profile is found and the user does not yet have a Citrix user profile in the user store, the Windows profile is migrated (copied) to the user store on the fly. After this process, the user store profile is used by Profile management in the current and any other session configured with the path to the same user store.

If this setting is enabled, profile migration can be activated for roaming and local profiles (the default), roaming profiles only, local profiles only, or profile migration can be disabled altogether.

If profile migration is disabled and no Citrix user profile exists in the user store, the existing Windows mechanism for creating new profiles is used as in a setup without Profile management.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, all types of existing profiles are migrated.
```

Setting Name: `MigrateWindowsProfilesToUserStore_Part`

Setting Value: `All`, `Local`, `Roaming`, `None`

### Minimum image quality
Description: 
```
Applies to: both file-based and container-based profile solutions

Specifies the minimum size (MB) of files to deduplicate from profile containers. The value must be 256 or greater.
If this setting is not configured here, the value from the .ini file is used.
If this setting is not configured either here or in the .ini file, the value is 256.
```

Setting Name: `SharedStoreProfileContainerFileSizeLimit_Part`

Setting Value: `{Minimum Size in MBs}`

### Multi-Port Policy

Description: 
```
Port number range is 0 to 65535. This port number cannot be the same as the default Citrix Group Policy port number.
```

Setting Name: `MultiPortPolicy`

Setting Value: `{Port1},{Priority Value};{Port2},{Priority Value};{Port3},{Priority Value};`

Priority Mapping:
Priority Name | Priority Value
-- | --
Very High | 0
Medium | 2
Low | 3

Constraints:
```
Port number range is 0 to 65535. This port number cannot be the same as the default Citrix Group Policy port number.
```

### Multi-Stream computer setting
Description: 
```
This setting determines whether Multi-Stream can be used on the session host.

If disabled, the use of Multi-Stream is prohibited. By default, this setting is disabled.

If enabled, the use of Multi-Stream is allowed. Note that for sessions to use Multi-Stream, you must also enable the "Multi-Stream user setting" setting.

When configuring this setting, you must restart the session host for changes to take effect.
```

Setting Name: `MultiStreamPolicy`

Setting Value:
```
{
    enabled = true | false
}
```

### Multi-Stream user setting
Description: 
```
This setting enables or disables Multi-Stream in HDX sessions.

If disabled, HDX sessions will not use Multi-Stream. By default, this setting is disabled.

If enabled, HDX sessions will use Multi-Stream. Note that you must allow the use of Multi-Stream by enabling "Multi-Stream computer setting" for this setting to take effect.
```

Setting Name: `MultiStream`
Related Settings:
```
Multi-Stream computer setting
```

Setting Value:
```
{
    enabled = true | false
}
```

### Multi-Stream virtual channel stream assignment
Description: 
```
Specifies which ICA stream the virtual channels will be assigned to when Multi-Stream is in use. If not configured, virtual channels will be kept in their default stream. To assign a virtual channel to an ICA stream select the desired stream number (0, 1, 2, 3) from the Stream Number drop-down list next to the virtual channel name.

If there is a custom virtual channel in use in the environment, click on the Add button, enter the virtual channel name in the text box under Virtual Channels, and select the desired stream number from the Stream Number drop-down list next to it. Please note that the name entered must be the actual virtual channel name and not a friendly name. e.g. CTXSBR instead of Citrix Browser Acceleration.

When configuring this setting, ensure that the Multi-Stream computer setting is enabled. Otherwise, this setting has no effect.

For more information on virtual channel assignments and priorities, please refer to CTX131001.
```

Setting Name: `MultiStreamAssignment`

Setting Value: `{Channel Name 1},{Assignment 1};{Channel Name 2},{Assignment 2};...`

Setting Value Explanation:
Channel Name | Description | Default Stream Assignment
--|--|--
CTXCAM | Audio | 0
CTXEUEM | End User Experience Monitoring | 1
CTXCTL | ICA Control | 1
CTXIME | Input Method Editor | 1
CTXLIC | License Management | 1
CTXMTOP | Microsoft Teams / WebRTC Redirection | 1
CTXMOB | Mobile Receiver | 1
CTXMTCH | MultiTouch | 1
CTXTWI | Seamless (Transparent Window Integration) | 1
CTXSENS | Sensor and Location | 1
CTXSCRD | Smart Card | 1
CTXTW | Thinwire Graphics | 1
CTXDND** | Drag & Drop | 1
CTXNSAP** | App Flow | 2
CTXCSB | Browser Content Redirection | 2
CTXCDM | Client Drive Mapping | 2
CTXCLIP | Clipboard | 2
CTXFILE | File Transfer (HTML5) | 2
CTXFLS2 | Flash v2 | 2
CTXFLSH | Flash | 2
CTXGDT | Generic Data Transfer | 2
CTXPFWD | Port Forwarding | 2
CTXMM | Remote Audio and Video Extensions (RAVE) | 2
CTXTUI | Transparent UI Integration / UI Status | 2
CTXTWN | TWAIN Redirection | 2
CTXGUSB | USB | 2
CTXZLC | Zero Latency Data Channel | 2
CTXZLFK | Zero Latency Font and Keyboard | 2
CTXCCM | Client COM Port Mapping | 3
CTXCPM | Client Printer Mapping | 3
CTXCOM1 | Legacy Client Printer Mapping (COM1) | 3
CTXCOM2 | Legacy Client Printer Mapping (COM2) | 3
CTXLPT1 | Legacy Client Printer Mapping (LPT1) | 3
CTXLPT2 | Legacy Client Printer Mapping (LPT2) | 3

### Multimedia conferencing
Description: 
```
Allows or prevents support for video conferencing applications. By default, video conferencing support is enabled.

When using multimedia conferencing, make sure the following items are present:
* Manufacturer-supplied drivers for the web cam used for multimedia conferencing must be installed.
* The web cam must be connected to the client device before initiating a video conferencing session. XenApp uses only one installed web cam at any given time. If multiple web cams are installed on the client device, XenApp attempts to use each web cam in succession until a video conferencing session is created successfully.
* An Office Communicator server must be present in your farm environment.
* The Office Communicator client software must be published on the server.
```

Setting Name: `MultimediaConferencing`

Related Settings:
```
Windows Media redirection
```

Setting Value:
```
{
    enabled = true | false
}
```

### Music path
Description: 
```
Applies to: file-based profile solution only

Lets you specify how to redirect the Music folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRMusicPath_Part`

Setting Value: `{Path of the Redirected Music Folder}`

### NTUSER.DAT backup
Description: 
```
Applies to: file-based profile solution only

By default, the NTUSER.DAT backup policy is enabled to back up the last good load NTUSER.DAT and roll back to it if NTUSER.DAT is corrupt.

If you do not configure this setting here, Profile Management uses the settings from the .ini file.

If you do not configure this setting here or in the .ini file, Profile Management does not backup the NTUSER.DAT.
```

Setting Name: `LastKnownGoodRegistry`

Setting Value:
```
{
    enabled = true | false
}
```

### Number of logoffs to trigger VHD disk compaction
Description: 
```
Applies to: both file-based and container-based profile solutions

Lets you specify the number of user logoffs to trigger VHD disk compaction.

When the number of logoffs since the last compaction reaches the specified value, disk compaction is triggered again.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, the default value 5 is used.
```

Setting Name: `NLogoffs4Compaction_Part`

Setting Value: `{Number of User Logoffs to Trigger VHD Disk Compaction}`

### Number of retries when accessing locked files
Description: 
```
Applies to: both file-based and container-based profile solutions

Sets the number of retries when accessing locked files.

If this setting is disabled the default value of five retries is used.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, the default value of five retries is used.
```

Setting Name: `LoadRetries_Part`

Setting Value: `{Number of Retries When Accessing Locked Files}`

### Offline profile support
Description:
```
Applies to: file-based profile solution only

Enables the offline profiles feature. This is intended for computers that are commonly removed from networks, typically laptops or mobile devices not servers or desktops.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, offline profile support is disabled.
```

Setting Name: `OfflineSupport`

Setting Value:
```
{
    enabled = true | false
}
```

### OneDrive container - List of OneDrive folders
Description:
```
Applies to: both file-based and container-based profile solutions

With this policy enabled, Profile Management stores OneDrive folders on a VHDX disk. The disk is attached during logons and detached during logoffs. The policy lets OneDrive folders roam with users.

Specify OneDrive folders as paths relative to the user profile. For example, if a OneDrive folder is located at "%userprofile%\OneDrive - Citrix", add "OneDrive - Citrix" to the list.

Starting with Profile Management version 2311, this policy supports OneDrive synchronization between concurrent sessions.
```

Setting Name: `OneDriveContainer_Part`

Setting Value:
```
jsonencode([
    {File Path 1},
    {File Path 2},
    ...
])
```

### Optimization for Windows Media multimedia redirection over WAN
Description:
```
Compresses Windows Media over WAN to ensure smooth playback
```

Setting Name: `MultimediaOptimization`

Setting Value:
```
{
    enabled = true | false
}
```

### Optimize for 3D graphics workload
Description:
```
This setting will configure appropriate default settings best suited for graphically intense workloads. Enable this setting for users whose workload will focus on graphically intense applications. This policy should only be applied in cases where a GPU is available to the session.
```

Setting Name: `OptimizeFor3dWorkload`

Related Settings:
```
Use video codec for compression

Graphic status indicator
```

Setting Value:
```
{
    enabled = true | false
}
```

### Outlook search index database - backup and restore
Description:
```
Applies to: both file-based and container-based profile solutions

This setting configures what Profile Management does during logon when search index roaming for Outlook is enabled.

If this setting is enabled, Profile Management saves a backup of the search index database each time the database is mounted successfully on logon. Profile Management treats the backup as the good copy of the search index database. When an attempt to mount the search index database fails because the database becomes corrupted, Profile Management automatically reverts the search index database to the last known good copy.

Note: Profile Management deletes the previously saved backup after a new backup is saved successfully. The backup consumes the available storage space of the VHDX files.
```

Setting Name: `OutlookEdbBackupEnabled`

Setting Value:
```
{
    enabled = true | false
}
```

### Specifies the total amount of bandwidth available for client sessions.
Description:
```
Specifies the total amount of bandwidth available for client sessions.
```

Setting Name: `OverallBandwidthLimit`

Related Settings:
```
Related settings
Printer redirection bandwidth limit percent

Audio redirection bandwidth limit percent

TWAIN device redirection bandwidth limit percent

File redirection bandwidth limit percent

Clipboard redirection bandwidth limit percent

LPT port redirection bandwidth limit percent

COM port redirection bandwidth limit percent

Client USB device redirection bandwidth limit percent

HDX MediaStream Multimedia Acceleration bandwidth limit percent
```

Setting Value: `{Bandwidth Limit in Kbps}`


### Path to Citrix Virtual Apps optimization definitions:
Description:
```
Applies to: file-based profile solution only

Specify a folder where the Citrix Virtual Apps optimization definition files are located.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, no Citrix Virtual Apps optimization settings will be applied.

Note:
* The folder can be local or it can reside on an SMB file share.
```

Setting Name: `XenAppOptimizationDefinitionPathData`

Setting Value: `{Path to the XenApp optimization Definition}`


### Path to cross-platform definitions
Description:
```
Applies to: both file-based and container-based profile solutions

Specify a folder where the definition files are located.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, no cross-platform settings will be applied.

Note:
* The folder can be local or it can reside on an SMB file share.
```

Setting Name: `CPSchemaPathData`

Setting Value: `{Path to the Cross-Platform Definition}`


### Path to cross-platform settings store
Description:
```
Applies to: both file-based and container-based profile solutions

Sets the path to the directory in which users' cross-platform settings are saved (cross-platform settings store).

The path can be an absolute UNC path or a path relative to the home directory.

In both cases, the following types of variables can be used: system environment variables enclosed in percent signs and attributes of the Active Directory user object enclosed in hashes.

User environment variables cannot be used, except for %username% and %userdomain%.

If this setting is disabled, the path Windows\PM_CP is used.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, the default value Windows\PM_CP is used.
```

Setting Name: `CPPathData`

Setting Value:  `{Path to the Cross-Platform Settings Store}`


### Path to log file
Description:
```
Sets an alternative path to which the log files are saved.

The path can point to a local drive or a remote, network-based one (a UNC path). Remote paths can be useful in large, distributed environments but they can create significant network traffic, which may be inappropriate for log files. For provisioned, virtual machines with a persistent hard drive, set a local path to that drive. This ensures log files are preserved when the machine restarts. For virtual machines without a persistent hard drive, setting a UNC path allows you to retain the log files but the system account for the machines must have write access to the UNC share. Use a local path for any laptops managed by the offline profiles feature.

Examples:
D:\LogFiles\ProfileManagement.
\\servername\LogFiles\ProfileManagement

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, the default location "%SystemRoot%\System32\Logfiles\UserProfileManager" is used.

If a UNC path is used for log files then Citrix recommends that an appropriate access control list is applied to the log file folder to ensure that only authorized user or computer accounts can access the stored files.
```

Setting Name: `DebugFilePath_Part`

Setting Value: `{Path to Log Files}`


### Path to the template profile:
Description:
```
By default, new user profiles are created from the default user profile on the computer where a user first logs on. Profile management can alternatively use a centrally stored template when creating new user profiles. Template profiles are identical to normal profiles in that they reside in any file share on the network. Use UNC notation to specifying paths to templates. Users need read access to a template profile.

If this setting is disabled, templates are not used.

If this setting is enabled, Profile management uses the template instead of the local default profile when creating new user profiles.

If a user has no Citrix user profile, but a local Windows user profile exists, by default the local profile is used (and migrated to the user store, if this is not disabled). This can be changed by enabling the setting "Template profile overrides local profile".

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, no template is used.
```

Setting Name: `TemplateProfilePath`

Setting Value: `{Path to Template Profile}`


### Path to user store
Description:
```
Sets the path to the directory in which the user settings (registry changes and synchronized files) are saved (user store).

The path can be an absolute UNC path or a path relative to the home directory.

In both cases, the following types of variables can be used: system environment variables enclosed in percent signs and attributes of the Active Directory user object enclosed in hashes.

Examples:
The folder Windows\%ProfileVer% stores the user settings in the subfolder called Windows\W2k3 of the user store (if %ProfileVer% is a system environment variable resolving to W2k3).

\\server\share\#SAMAccountName# stores the user settings to the UNC path \\server\share\JohnSmith (if #SAMAccountName# resolves to JohnSmith for the current user).

User environment variables cannot be used, except for %username% and %userdomain%.

If this setting is disabled, the user settings are saved in the Windows subdirectory of the home directory.

If this setting is not configured here, the setting from the .ini file is used.

If this setting is not configured here or in the .ini file, the Windows directory on the home drive is used.
```

Setting Name: `DATPath_Part`

Setting Value: `{Path to the User Store}`

### Personalized user information
Description:
```
Detailed log settings.

Define events or actions that Profile management logs in depth.

If this setting is not configured here, Profile management uses the settings from the .ini file.

If this setting is not configured here or in the .ini file, errors and general information are logged.
```

Setting Name: `LogLevel_UserName`

Setting Value:
```
{
    enabled = true | false
}
```

### Pictures path
Description:
```
Lets you specify how to redirect the Pictures folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRPicturesPath_Part`

Setting Value: `{Path to the Redirected Pictures Path}`


### Policy values at logon and logoff
Description:
```
Detailed log settings.

Define events or actions that Profile management logs in depth.

If this setting is not configured here, Profile management uses the settings from the .ini file.

If this setting is not configured here or in the .ini file, errors and general information are logged.
```

Setting Name: `LogLevel_PolicyUserLogon`

Setting Value:
```
{
    enabled = true | false
}
```

### Posture check for Citrix Workspace App
Description:
```
App Protection Posture Check

This allows you to block access to resources protected by App Protection unless they are on versions of Citrix Workspace App where the specific App Protection controls can be enforced.

Note: If this feature is applied, users on the Workspace app versions that do not support App Protection Posture Check will also be blocked from accessing protected sessions.
For more details on prerequisites and configuration refer to https://docs.citrix.com/en-us/citrix-workspace-app/app-protection/features.html#posture-check

Important considerations while creating new policy:
- Each entry should have only one capability.
- No space is allowed in the name of capability.
- Ensure the values are spelt correctly. Incorrectly spelt values will cause session disconnects.
```

Setting Name: `AppProtectionPostureCheck`

Setting Value:
```
jsonencode([
    {Posture Check 1},
    {Posture Check 2},
    ...
])
```


### Preferred color depth for simple graphics
Description:
```
This setting is available only on VDA versions XenApp and XenDesktop 7.6 Feature Pack 3 and later.

Allows lowering of the color depth at which simple graphics are sent to 8 or 16 bits per pixel, potentially improving responsiveness over low bandwidth connections, at the cost of a slight degradation of image quality. This option is supported only when a video codec is not used to compress graphics.

8-bit is only available for 7.12 VDAs and later. Setting 8-bit for 7.11 and earlier VDAs will result in those VDAs falling back to 24-bit.
```

Setting Name: `PreferredColorDepthForSimpleGraphics`

Setting Value: `ColorDepth8Bit`, `ColorDepth16Bit`, `ColorDepth24Bit`


### Preserve client drive letters
Description:
```
Enables or disables preservation of client drive letters. When enabled, and client drive mapping is enabled, client drives are mapped to the same drive letter in the session, where possible. By default, client drive letters are not preserved.
```

Setting Name: `ClientDriveLetterPreservation`

Related Settings:
```
Client drive redirection
```

Setting Value:
```
{
    enabled = true | false
}
```


### Primary selection update mode
Description:
```
This setting is supported only by Linux VDA version 1.4 onwards.

PRIMARY selection is used for explicit copy/paste actions such as mouse selection and middle mouse button paste. This setting controls whether PRIMARY selection changes on the Linux VDA can be updated on the client's clipboard (and vice versa). It can include one of the following selection changes:

Selection changes are not updated on the client or the host. PRIMARY selection changes do not update a client's clipboard. Client clipboard changes do not update PRIMARY selection.

Host selection changes are not updated on the client. PRIMARY selection changes do not update a client's clipboard. Client clipboard changes update the PRIMARY selection.

Client selection changes are not updated on the host. PRIMARY selection changes update the client's clipboard. Client clipboard changes do not update the PRIMARY selection.

Selection changes are updated on both the client and host. PRIMARY selection change updates the client’s clipboard. Client clipboard changes update the PRIMARY selection.
```

Setting Name: `PrimarySelectionUpdateMode`

Setting Value: `AllUpdatesAllowed`, `UpdateToHostDenied`, `UpdateToClientDenied`, `AllUpdatesDenied`


### Printer assignments
Description:
```
Contains one or more entries, each specifying a default printer value, or session printers value, or both, for one or more clients. The default printer value specifies how the clients' default printer is established in an ICA session. The session printer value lists the network printers to be auto-created in the clients' ICA sessions.

When creating filters for a policy that contains this setting, include all clients specified in all entries in the setting.

By default, the client's current printer is used as the default printer for the session.

When setting the default printer value:
'Client main printer' allows the client's current default printer to be used as the default printer for the session.

'Generic Universal Printer' allows the Generic Universal printer to be used as the default printer for the session.

'PDF Printer' allows the PDF printer to be used as the default printer for the sessions using the the Citrix HTML or Chrome Receiver.

'Do not adjust' uses the current Terminal Services or Windows user profile setting for the default printer. If you choose this option, the default printer is not saved in the profile and it does not change according to other session or client properties. The default printer in a session will be the first printer auto-created in the session, which is either:

* The first printer added locally to the Windows server in Control Panel > Printers
* The first auto-created printer, if there are no printers added locally to the server
You can use this option to present users with the nearest printer through profile settings (known as Proximity Printing).

'<Not Set>' means that you are not changing the value set by the system or other policies. It equates to a 'Not Configured' state for that entry.

When policy settings are resolved, default printers applied by this setting take precedence over default printers applied by the 'Default printer' setting.

When setting the session printer value:

You can add printers to the list, edit the settings of a list entry, or remove printers from the list. You can apply customized settings for the current session at every logon. When printing polices are applied, session printers are merged into a single list: all printers applied in all policies that include a 'Printing Assignments' setting; and, all printers applied in policies that include a 'Session printer' setting.

Use individual 'Default printers' and 'Session printers' settings to set default behaviors for a farm, site, large group, or OU. Use the 'Printer assignments' setting to assign a large group of printers to multiple users.
```

Setting Name: `PrinterAssignments`

### Printer auto-creation event log preference
Description:
```
Specifies which events are logged during the printer auto-creation process. You can choose to log no errors or warnings, only errors, or errors and warnings. By default, errors and warnings are logged.
An example of a warning is an event in which a printer's native driver could not be installed and the universal printer driver is installed instead. To allow universal printer drivers to be used in this scenario, set the 'Universal print driver usage' setting to 'Use universal printing only' or 'Use universal printing only if requested driver is unavailable'.
```

Setting Name: `AutoCreationEventLogPreference`

Setting Value: `DoNotLog`, `LogErrorsOnly`, `LogErrorsAndWarnings`


### Printer driver mapping and compatibility
Description:
```
Lists driver substitution rules for auto-created client printers. When you define these rules, you can allow or prevent printers to be created with the specified driver. Additionally, you can allow created printers to use only universal printer drivers.
Driver substitution overrides (or maps) printer driver names the client provides, substituting an equivalent driver on the server. This gives server applications access to client printers that have the same drivers as the server but different driver names.

You can add a driver mapping, edit an existing mapping, override custom settings for a mapping, remove a mapping, or change the order of driver entries in the list. When adding a mapping, enter the client printer driver name and then select the server driver you want to substitute.
```

Setting Name: `PrinterDriverMappings`

Setting Value:
```
jsonencode([
    "{Driver Name 1},{Action 1}",
    "{Driver Name 2},{Action 2}",
])
""

Setting Value:
```
jsonencode([
    "Microsoft XPS Document Writer,Deny,Deny",
    "Send to Microsoft OneNote *,UPD_Only",
    "Printer Driver,Replace=Printer Driver Replaced",
    "Printer Driver Allowed,Allow",
])
""

Action Value Mapping:
Action Value | Comment
-- | --
`Deny,Deny` | Do not create
`UPD_Only` | Create with universal driver
`Replace={Replaced Driver Name}` | Replace with `Replaced Driver Name`
`Allow` | Allow

### Printer properties retention
Description:
```
Specifies whether and where to store printer properties. By default, the system determines whether printer properties are stored on the client device, if available, or in the user profile.

Select 'Held in profile only if not saved on client' to allow the system to determine where printer properties are stored. Printer properties are stored either on the client device, if available, or in the user profile. Although this option is the most flexible, it can also slow logon time and use extra bandwidth for system-checking.

Select 'Saved on the client device only' if your system has a mandatory or roaming profile that you do not save. Choose this option only if all the servers in your farm are running XenApp 5 and your users are using Citrix XenApp Plugin versions 9.x or later.

Select 'Retained in user profile only' if your system is constrained by bandwidth (this option reduces network traffic) and logon speed or your users use legacy plug-ins. This option stores printer properties in the user profile on the server and prevents any properties exchange with the client device. Use this option with MetaFrame Presentation Server 3.0 or earlier and MetaFrame Presentation Server Client 8.x or earlier. Note that this is applicable only if a Terminal Services roaming profile is used.
```

Setting Name: `PrinterPropertiesRetention`

Setting Value: `DoNotRetain`, `FallbackToProfile`, `RetainedInUserProfile`, `SavedOnClientDevice`


### Printer redirection bandwidth limit
Description:
```
Specifies the maximum allowed bandwidth in kilobits per second for accessing client printers in a client session.

If you enter a value for this setting and a value for the 'Printer redirection bandwidth limit percent' setting, the most restrictive setting (with the lower value) is applied.
```

Setting Name: `PrinterBandwidthLimit`

Related Settings:
```
Printer redirection bandwidth limit percent
```

Setting Value: `{Bandwidth Limit in Kbps}`

### Printer redirection bandwidth limit percent
Description:
```
Specifies the maximum allowed bandwidth for accessing client printers as a percent of the total session bandwidth.

If you enter a value for this setting and a value for the 'Printer redirection bandwidth limit (Kbps)' setting, the most restrictive setting (with the lower value) is applied.

If you configure this setting, you must also configure the 'Overall session bandwidth limit' setting which specifies the total amount of bandwidth available for client sessions.
```

Setting Name: `PrinterBandwidthPercent`

Related Settings:
```
Printer redirection bandwidth limit

Overall session bandwidth limit

Client printer redirection
```

Setting Value: `{Bandwidth Limit Percentage Value}`


### Process Internet cookie files on logoff
Description:
```
Some deployments leave extra Internet cookies that are not referenced by the file index.dat. The extra cookies left in the file system after sustained browsing can lead to profile bloat. Enable this setting to force processing of index.dat and remove the extra cookies. The setting might slightly increase logoff time. Enable this setting only when you experience the issue.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, no processing of index.dat takes place.
```

Setting Name: `ProcessCookieFiles`

Setting Value:
```
{
    enabled = true | false
}
```


### Process logons of local administrators
Description:
```
Specifies whether logons of members of the local group "Administrators" are processed by Profile management.

If this setting is disabled, logons by local administrators are not processed by Profile management.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, administrators will not be processed.
```

Setting Name: `ProcessAdmins`

Setting Value:
```
{
    enabled = true | false
}
```


### Processed groups
Description:
```
Both computer local groups and domain groups (local, global and universal) can be used. Domain groups should be specified in the format: <DOMAIN NAME>\<GROUP NAME>.

If this setting is configured here, Profile management processes only members of these user groups.

If this setting is disabled, Profile management processes all users.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, members of all user groups are processed.
```

Setting Name: `ProcessedGroups_Part`

Setting Value:
```
jsonencode([
    {Group 1},
    {Group 2}
])
```

Example Setting Value:
```
jsonencode([
    "DOMAIN\GROUPNAM1",
    "DOMAIN\GROUPNAM2",
    ...
])
```


### Profile container
Description:
```
By default, VHD containers allow concurrent access. With this setting enabled, they allow only one access at a time.

Note: Enabling this setting for profile containers in a container-based profile solution automatically disables the "Enable multi-session write-back for profile containers" setting.

If this policy is not configured here, the value from the .ini file is used.

If this policy is not configured either here or in the .ini file, the setting is disabled.
```

Setting Name: `DisableConcurrentAccessToProfileContainer`

Setting Value:
```
{
    enabled = true | false
}
```

### Profile container - List of folders to be contained in profile disk
Description:
```
A profile container is a VHDX based profile solution that lets you specify the folders to contain on the profile disk. The profile container attaches the profile disk containing those folders, thus eliminating the need to save a copy of the folders to the local profile. Doing so decreases logon times.

To use a profile container, enable this policy and add the relative paths of the folders to the list. Citrix recommends that you include the folders containing large cache files in the list. For example, add the Citrix Files content cache folder to the list: AppData\Local\Citrix\Citrix Files\PartCache

To enable a profile container for the entire user profile, add an asterisk (*) to the list. The following user profiles (if any) are automatically migrated to the container upon its first use:

- Local Windows user profile
- User profiles from the Citrix file-based profile solution

If this setting is not configured here, the value from the .ini file is used. If this setting is configured neither here nor in the .ini file, it is disabled by default.
```

Setting Name: `ProfileContainer_Part`

Setting Value:
```
jsonencode([
    {Path 1},
    {Path 2}
])
```


### Profile container auto-expansion increment
Description:
```
Specifies the amount of storage capacity (in GB) by which profile containers automatically expand when auto-expansion is triggered.

If this policy is not configured here, the value from the .ini file is used.

If this policy is not configured either here or in the .ini file, the default is 10 (GB).
```

Setting Name: `VhdAutoExpansionIncrement_Part`

Setting Value: `{Expansion Increment in GB}`

### Profile container auto-expansion limit
Description:
```
Specifies the maximum storage capacity (in GB) to which profile containers can automatically expand when auto-expansion is triggered.

If this policy is not configured here, the value from the .ini file is used.

If this policy is not configured either here or in the .ini file, the default is 80 (GB).
```

Setting Name: `VhdAutoExpansionLimit_Part`

Setting Value: `{Maximum Storage Capacity in GB}`

### Profile container auto-expansion threshold
Description:
```
Specifies the utilization percentage of storage capacity at which profile containers trigger auto-expansion.

If this policy is not configured here, the value from the .ini file is used.

If this policy is not configured here or in the .ini file, the default is 90 (%) of storage capacity.
```

Setting Name: `VhdAutoExpansionThreshold_Part`

Setting Value: `{Utilization Percentage Value}`


### Profile streaming
Description:
```
With profile streaming, users' profiles are synchronized on the local computer only when they are needed. Registry entries are cached immediately, but files and folders are only cached when accessed by users.
```

Setting Name: `PSEnabled`

Setting Value:
```
{
    enabled = true | false
}
```


### Profile streaming exclusion list - directories
Description:
```
List of directories that are ignored by Profile Streaming.
Folder names must be specified as paths relative to the user profile.

Examples:
Entering "Desktop" (without quotes) ignores the Desktop directory in the user profile.
If this setting is disabled, no folders are excluded.
If this setting is not configured here, the value from the .ini file is used.
If this setting is not configured here or in the .ini file, no folders are excluded.

Note:
Profile Streaming exclusions do not indicate that the configured folders are excluded from profile handling. They are still processed by Citrix Profile Management.
```

Setting Name: `StreamingExclusionList_Part`

Setting Value:
```
jsonencode([
    {Folder Path 1},
    {Folder Path 2}
])
```


### Read-only client drive access
Description:
```
When enabled, files/folders on mapped client drives can only be accessed in read-only mode.
When disabled, files/folders on mapped client drives can be accessed in regular read/write mode.

When enabling/disabling this setting, make sure the 'Client drive redirection' setting is enabled.
```

Setting Name: `ReadOnlyMappedDrive`

Related Settings:
```
Client drive redirection
```

Setting Value:
```
{
    enabled = true | false
}
```


### Reboot message box body text
Description:
```
Text displayed in the message box that alerts users when a machine restart is in progress. The text of the message body must be supplied and must be 3072 characters or less.
```

Setting Name: `RebootMessageBody`

Setting Value: `{Reboot Message}`


### Reconnection UI transparency level
Description:
```
This setting controls transparency level applied to XenApp/XenDesktop session window during ACR and SR reconnection period

0% - XenApp/XenDesktop session window will be turned to black window
100% - No transparency layer will be applied (frozen screen)
```

Setting Name: `ReconnectionUiTransparencyLevel`

Setting Value: `{Transparency Level Percentage Value}`


### Redirection settings for AppData(Roaming)
Description:
```
Lets you specify how to redirect the AppData(Roaming) folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRAppData_Part`

Setting Value: `{Path to the Redirected AppData Roaming Folder}`


### Redirection settings for Contacts
Description:
```
Lets you specify how to redirect the Contacts folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRContacts_Part`

Setting Value: `{Path to the Redirected Contacts Folder}`


### Redirection settings for Desktop
Description:
```
Lets you specify how to redirect the Desktop folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRDesktop_Part`

Setting Value: `{Path to the Redirected Desktop Folder}`


### Redirection settings for Documents
Description:
```
Lets you specify how to redirect the Documents folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRDocuments_Part`

Setting Value: `{Path to the Redirected Documents Folder}`


### Redirection settings for Downloads
Description:
```
Lets you specify how to redirect the Downloads folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRDownloads_Part`

Setting Value: `{Path to the Redirected Downloads Folder}`


### Redirection settings for Favorites
Description:
```
Lets you specify how to redirect the Favorites folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRFavorites_Part`

Setting Value: `{Path to the Redirected Favorites Folder}`


### Redirection settings for Links
Description:
```
Lets you specify how to redirect the Links folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRLinks_Part`

Setting Value: `{Path to the Redirected Links Folder}`


### Redirection settings for Music
Description:
```
Lets you specify how to redirect the Music folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRMusic_Part`

Setting Value: `{Path to the Redirected Music Folder}`


### Redirection settings for Pictures
Description:
```
Lets you specify how to redirect the Pictures folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRPictures_Part`

Setting Value: `{Path to the Redirected Pictures Folder}`


### Redirection settings for Saved Games
Description:
```
Lets you specify how to redirect the Saved Games folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRSavedGames_Part`

Setting Value: `{Path to the Redirected SavedGames Folder}`


### Redirection settings for Searches
Description:
```
Lets you specify how to redirect the Searches folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRSearches_Part`

Setting Value: `{Path to the Redirected Searches Folder}`


### Redirection settings for Start Menu
Description:
```
Lets you specify how to redirect the Start Menu folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRStartMenu_Part`

Setting Value: `{Path to the Redirected Start Menu Folder}`


### Redirection settings for Videos
Description:
```
Lets you specify how to redirect the Videos folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRVideos_Part`

Setting Value: `{Path to the Redirected Videos Folder}`


### Registry actions
Description:
```
Detailed log settings.

Define events or actions that Profile management logs in depth.

If this setting is not configured here, Profile management uses the settings from the .ini file.

If this setting is not configured here or in the .ini file, errors and general information are logged.
```

Setting Name: `LogLevel_RegistryActions`

Setting Value:
```
{
    enabled = true | false
}
```


### Registry differences at logoff
Description:
```
Detailed log settings.

Define events or actions that Profile management logs in depth.

If this setting is not configured here, Profile management uses the settings from the .ini file.

If this setting is not configured here or in the .ini file, errors and general information are logged.
```

Setting Name: `LogLevel_RegistryDifference`

Setting Value:
```
{
    enabled = true | false
}
```


### Regular time interval at which the agent task is to run
Description:
```
The Connector Agent service for System Center Configuration Manager 2012 runs the AgentTask at the time interval specified by this setting.

A time span string consists of the following:

ddd . Days (from 0 to 999) [optional]
hh : Hours (from 0 to 23)
mm : Minutes (from 0 to 59)
ss Seconds (from 0 to 59)
```

Setting Name: `AgentTaskInterval`

Setting Value: `{ddd.hh.mm.ss where ddd. is optional}`

Example Setting Value: `00:05:00`

### Remote Credential Guard Mode
Description:
```
This feature allows leveraging delegated non-exportable Active Directory credentials for single sign-on into the virtual session.

The Windows policy setting "Remote host allows delegation of non-exportable credentials" must also be enabled on the VDA through Group Policy or Local Policy to use this feature. This setting is located under Computer Configuration > Administrative Templates > System > Credentials Delegation.

This feature can only be used with domain joined Windows endpoint devices.

By default, this setting is disabled.
```

Setting Name: `RemoteCredentialGuard`

Setting Value:
```
{
    enabled = true | false
}
```

### Remote the combo box
Description:
```
Enables or disables the remoting of the combo box on mobile devices. By default, the remoting of the combo box is disabled.
```

Setting Name: `ComboboxRemoting`

Setting Value:
```
{
    enabled = true | false
}
```


### Rendezvous Protocol
Description:
```
This setting is only applicable to HDX session established through Citrix Cloud.

When this setting is set to enabled: VDA establishes an outbound connection to NetScaler Gateway Service rendezvous endpoint on Port 443 bypassing the Citrix Cloud Connector.
When this setting is set to disabled: Citrix Cloud Connector proxies the VDA connection to NetScaler Gateway Service.
```

Setting Name: `RendezvousProtocol`

Related Settings:
```
Rendezvous proxy configuration
```

Setting Value:
```
{
    enabled = true | false
}
```


### Rendezvous proxy configuration
Description:
```
This setting allows configuring an explicit proxy for use with the Rendezvous protocol. If using a transparent proxy, this setting does not need to be enabled.

By default, this setting is disabled.

When disabled, the VDA will not route outbound traffic through any non-transparent proxies when trying to establish a Rendezvous connection with the Gateway Service.

When enabled, the VDA will attempt to establish a Rendezvous connection with the Gateway Service through the proxy defined in this setting.

The VDA supports using HTTP and SOCKS5 proxies for Rendezvous connections. To configure the VDA to use a proxy for the Rendezvous connection, you must enable this setting and specify either the proxy's address or the path to the PAC file.

For example:
Proxy address: "http://<URL or IP>:<port>" or "socks5://<URL or IP>:<port>"
PAC file: “http://<URL or IP>/<path>/<fileName>.pac”

NOTES:
1. EDT can only be proxied through SOCKS5 proxies. If using an HTTP proxy, you must use ICA over TCP.
2. Please refer to the product documentation for guidance on PAC file schema for SOCKS5 proxies.
3. VDA version 2103 is the minimum required for proxy configuration via PAC file.
```

Setting Name: `RendezvousProxy`

Setting Value: `{Path to Proxy Configuration File}`


### Replicate user stores - Paths to replicate a user store
Description:
```
Lets you replicate a user store to multiple paths in addition to the path that the Path to user store policy specifies, on each logon and logoff. To synchronize to the user store files and folders modified during a session, enable active write back.

Starting with Profile Management version 2209, this feature is available for the full profile container. You can replicate the container to multiple paths. Replicated containers provide profile redundancy for user logon but not for in-session failover.

Note: Enabling the policy can increase system I/O and might prolong logoffs.

For Profile Management version 2112 or later, you can separate VHDX files from the replicated user store and store them to different paths. To do so, add the path for VHDX files after the path to the replicated user store, and separate the two paths with a vertical bar (|).
Example: \\path_a|\\path_b indicates that the user store (with VHDX files excluded) is stored in \\path_a and VHDX files are stored in \\path_b.
```

Setting Name: `MultiSiteReplication_Part`

Setting Value:
```
jsonencode([
    {Path 1},
    {Path 2},
    ...
])
```


### Restore Desktop OS time zone on session disconnect or logoff
Description:
```
Determines whether the time zone setting for a single session Desktop OS VDA is restored to the machine's original time zone when the user disconnects or logs off.

When disabled, the VDA keeps the machine's time zone set to the client's time zone when the user disconnects or logs off.

When enabled, the VDA restores the machine's time zone to its original setting when the user disconnects or logs off.

This setting has effect only when the "Use local time of client" setting is set to "Use client time zone."
```

Related Settings:
```
Use local time of client
```

Setting Name: `RestoreServerTime`

Setting Value:
```
{
    enabled = true | false
}
```


### Restrict client clipboard write
Description:
```
If this setting is set to Enabled, host clipboard data cannot be shared with client endpoint. Administrators can selectively allow specific formats by enabling setting 'Client clipboard write allowed formats'.
```

Setting Name: `RestrictClientClipboardWrite`

Setting Value:
```
{
    enabled = true | false
}
```


### Restrict session clipboard write
Description:
```
When this setting is set to Enabled, client clipboard data cannot be shared within the user session. Administrators could selectively allow some formats by enabling the setting 'Session clipboard write allowed formats'.
```

Setting Name: `RestrictSessionClipboardWrite`

Setting Value:
```
{
    enabled = true | false
}
```

### Saved Games path
Description:
```
Lets you specify how to redirect the Saved Games folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRSavedGamesPath_Part`

Setting Value: `{Path to the Redirected Saved Games Folder}`


### Screen sharing
Description:
```
This setting enables users to share their sessions, including screen contents, keyboards, and mice, with other users.

The VDA attempts to use ports from the TCP port range to exchange data, starting with the lowest port and incrementing on each subsequent connection. The port handles both inbound and outbound traffic.

By default, the TCP port range is set to 52525-52625.
```

Setting Name: `ScreenSharing`

Setting Value:
```
{
    enabled = true | false
}
```


### Searches path
Description:
```
Lets you specify how to redirect the Searches folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRSearchesPath_Part`

Setting Value: `{Path to the Redirected Searches Folder}`


### Secure HDX
Description:
```
Secure HDX uses application-level encryption to secure the ICA protocol and provide true End-to-End encryption for HDX sessions.

When disabled, Secure HDX is not used.

When enabled, Secure HDX is enforced, and connections to clients that do not support Secure HDX are blocked. Also, no network components in the path can parse HDX traffic, including NetScaler Gateway (HDX Insight).

When enabled for direct connections only, Secure HDX encrypts the ICA protocol for sessions where the client connects directly to the session host but is not used for sessions connected through a Gateway.
```

Setting Name: `SecureHDX`

Setting Value: `Disabled`, `Enabled`, `DirectConnectionsOnly`


### SecureICA minimum encryption level
Description:
```
Specifies the minimum level at which to encrypt session data sent between the server and a client device.

'Basic' encrypts the client connection using a non-RC5 algorithm. It protects the data stream from being read directly, but it can be decrypted. By default, the server uses Basic encryption for client-server traffic.

`LogOn`: 'RC5 (128 bit) logon only' encrypts the logon data with RC5 128-bit encryption and the client connection using Basic encryption.

`Bits40`: 'RC5 (40 bit)' encrypts the client connection with RC5 40-bit encryption.

`Bits56`: 'RC5 (56 bit)' encrypts the client connection with RC5 56-bit encryption.

`Bits128`: 'RC5 (128 bit)' encrypts the client connection with RC5 128-bit encryption.

The settings you specify for client-server encryption can interact with any other encryption settings in XenApp and your Windows operating system. If a higher priority encryption level is set on either a server or client device, settings you specify for published resources can be overridden.

You can raise encryption levels to further secure communications and message integrity for certain users. If a policy requires a higher encryption level, plug-ins using a lower encryption level are denied connection.

SecureICA does not perform authentication or check data integrity. To provide end-to-end encryption for your server farm, use SecureICA with TLS encryption.

SecureICA does not use FIPS-compliant algorithms. If this is an issue, configure the server and plug-ins to avoid using SecureICA.
```

Setting Name: `MinimumEncryptionLevel`

Setting Value: `Basic`, `LogOn`, `Bits40`, `Bits56`, `Bits128`


### Server idle timer interval
Description:
```
Determines, in milliseconds, how long an uninterrupted user session will be maintained if there is no input from the user. By default, idle connections are not disconnected (Server idle timer interval = 0).
```

Setting Name: `IdleTimerInterval`

Setting Value: `{Idle Timer Interval in milliseconds}` 


### Session clipboard write allowed formats
Description:
```
This setting doesn't apply if 'Client clipboard redirection' is set to Prohibited or 'Restrict session clipboard write' is not set to Enabled.

When the setting 'Restrict session clipboard write' is set to Enabled, client clipboard data cannot be shared with session applications but this setting can be used to selectively allow specific data formats to be shared with session clipboard. Administrators can enable this setting and specify the formats to be allowed.

The followings are system defined clipboard formats.
CF_TEXT
CF_BITMAP
CF_METAFILEPICT
CF_SYLK
CF_DIF
CF_TIFF
CF_OEMTEXT
CF_DIB
CF_PALETTE
CF_PENDATA
CF_RIFF
CF_WAVE
CF_UNICODETEXT
CF_LOCALE
CF_DIBV5
CF_DSPTEXT
CF_DSPBITMAP
CF_DSPMETAFILEPICT
CF_DSPENHMETAFILE
CF_HTML

The followings are custom formats predefined in XenApp and XenDesktop.
CFX_RICHTEXT
CFX_OfficeDrawingShape
CFX_BIFF8
CFX_FILE

Additional custom formats can be added. Actual name for custom formats must match the formats to be registered with system. Format names are case insensitive.
```

Setting Name: `SessionClipboardWriteAllowedFormats`

Setting Value:
```
jsonencode([
    {Format 1},
    {Format 2},
    ...
])
```

### Session connection timer
Description:
```
Enables or disables a timer to determine the maximum duration of an uninterrupted connection between a user device and a workstation. By default, this timer is disabled.
```

Setting Name: `SessionConnectionTimer`

Setting Value:
```
{
    enabled = true | false
}
```


### Session connection timer - Multi-session
Description:
```
Enables or disables a timer to determine the maximum duration of an uninterrupted connection between a user device and a terminal server. By default, this timer is disabled.
```

Setting Name: `EnableServerConnectionTimer`

Setting Value:
```
{
    enabled = true | false
}
```


### Session connection timer interval
Description:
```
Determines, in minutes, the maximum duration of an uninterrupted connection between a user device and a workstation. By default, the maximum duration is 1440 minutes (24 hours).
```

Setting Name: `SessionConnectionTimerInterval`

Setting Value: `{Idle Timer Inverval in Minutes}`


### Session connection timer interval - Multi-session
Description:
```
Determines, in minutes, the maximum duration of an uninterrupted connection between a user device and an RDS session. By default, the maximum duration is 1440 minutes (24 hours).
```

Setting Name: `ServerConnectionTimerInterval`

Setting Value: `{Idle Timer Inverval in Minutes}`


### Session idle timer
Description:
```
Enables or disables a timer to determine how long an uninterrupted user device connection to a workstation will be maintained if there is no input from the user. By default, this timer is enabled.
```

Setting Name: `SessionIdleTimer`

Setting Value:
```
{
    enabled = true | false
}
```


### Session idle timer - Multi-session
Description:
```
Enables or disables a timer to determine the maximum duration of an idle connection between a user device and a terminal server. By default, this timer is disabled.
```

Setting Name: `EnableServerIdleTimer`

Setting Value:
```
{
    enabled = true | false
}
```


### Session idle timer interval
Description:
```
Determines, in minutes, how long an uninterrupted user device connection to a workstation will be maintained if there is no input from the user. By default, idle connections are maintained for 1440 minutes (24 hours).
```

Setting Name: `SessionIdleTimerInterval`

Setting Value: `{Idle Timer Inverval in Minutes}`


### Session idle timer interval - Multi-session
Description:
```
Determines, in minutes, the maximum duration of an idle connection between a user device and an RDS session. By default, the maximum duration is 1440 minutes (24 hours).
```

Setting Name: `ServerIdleTimerInterval`

Setting Value: `{Idle Timer Inverval in Minutes}`


### Session printers
Description:
```
Lists the network printers to be auto-created in an ICA session. You can add printers to the list, edit the settings of a list entry, or remove printers from the list. You can apply customized settings for the current session at every logon.

The printers are merged with any other 'Session printers' settings applied in other policies, and also with any 'Printer assignments' settings that apply session printers.

Use individual 'Session printers' policies to set default behaviors for a farm, site, large group, or OU. Use the 'Printer assignments' policy to assign a large group of printers to multiple users.
```

Setting Name: `SessionPrinters`

Setting Value:
```
jsonencode([
    "{Printer Address 1},model={Shared Name 1},location={Location 1}",
    "{Printer Address 2},model={Shared Name 2},location={Location 2}",
    ...
])
```


### Session reliability connections
Description:
```
Allow or prevent session reliability connections.

Session Reliability keeps sessions active when network connectivity is interrupted. Users continue to see the application they are using until network connectivity resumes.

When connectivity is momentarily lost, the session remains active on the server. The user's display freezes and the cursor changes to a spinning hourglass until connectivity resumes. The user continues to access the display during the interruption and can resume interacting with the application when the network connection is restored. Session Reliability reconnects users without reauthentication prompts.

If you do not want users to be able to reconnect to interrupted sessions without having to reauthenticate, configure the Auto client reconnect authentication setting to require authentication. Users are then prompted to reauthenticate when reconnecting to interrupted sessions.

If you use both Session Reliability and Auto Client Reconnect, the two features work in sequence. Session Reliability closes, or disconnects, the user session after the amount of time you specify in the Session reliability timeout setting. After that, the settings you configure for Auto Client Reconnect take effect, attempting to reconnect the user to the disconnected session.
```

Setting Name: `SessionReliabilityConnections`

Setting Value:
```
{
    enabled = true | false
}
```


### Session reliability port number
Description:
```
TCP port number for incoming session reliability connections.
```

Setting Name: `SessionReliabilityPort`

Setting Value: `{TCP Port Number}`


### Session reliability timeout
Description:
```
The length of time in seconds the session reliability proxy waits for a client to reconnect before allowing the session to be disconnected.

The default length of time is 180 seconds, or three minutes. Though you can extend the amount of time a session is kept open, this feature is designed to be convenient to the user and it does not prompt the user for reauthentication. If you extend the amount of time a session is kept open indiscriminately, chances increase that a user may get distracted and walk away from the client device, potentially leaving the session accessible to unauthorized users.

If you do not want users to be able to reconnect to interrupted sessions without having to reauthenticate, configure the Auto client reconnect authentication setting to require authentication. Users are then prompted to reauthenticate when reconnecting to interrupted sessions.

If you use both Session Reliability and Auto Client Reconnect, the two features work in sequence. Session Reliability closes, or disconnects, the user session after the amount of time you specify in the Session reliability timeout setting. After that, the settings you configure for Auto Client Reconnect take effect, attempting to reconnect the user to the disconnected session.
```

Setting Name: `SessionReliabilityTimeout`

Setting Value: `{Timeout in Seconds}`


### Session watermark style
Description:
```
Controls the style of session watermark text labels.

This policy is only effective when session watermark is enabled.

With Multiple configuration, five watermark labels will be displayed in the session, with one in the center and four in the corners.

With Single configuration, a single watermark text label will displayed at the center of the session.
```

Setting Name: `WatermarkStyle`

Setting Value: `StyleSingle`, `StyleMultiple`


### Set priority order for user groups
Description:
```
Specify the priority order for user groups. The order determines which group takes precedence when a user belongs to multiple groups with different policy settings. In the Priority order for user groups field, enter the Security Identifiers (SIDs) or domain names of the groups in descending order of priority, separated by semicolons (;).

Example:
ctxxa.local\groupb;S-1-5-21-674278408-26188528-2146851469-1174;ctxxa.local\groupc;

When a user belongs to multiple groups with different policy settings, consider the following:

- If the user belongs to one or more groups defined in this policy, the group with the highest priority takes precedence.

- If the user doesn't belong to any of the groups defined in this policy, the group with the SID listed earliest in alphabetical order takes precedence.

If this setting is not configured here, the value from the .ini file is used.

If this setting is configured neither here nor in the .ini file, no priority order is specified.
```

Setting Name: `OrderedGroups_Part`

Setting Value: `{UserGroup1};{UserGroup2};...`


### Source for creating cross-platform settings
Description:
```
Default: `false`

Definition files contain a set of definitions that configure an application. If the cross-platform settings store contains a definition with no data, or the cached data in the single-platform profile is newer than the definition's data in the store, Profile management only migrates the data from the single-platform profile to the store if you enable this setting.
```

Setting Name: `CPMigrationFromBaseProfileToCPStore`

Setting Value:
```
{
    enabled = true | false
}
```


### Special folder redirection
Description:
```
Allows or prevents Citrix Receiver and Web Interface users to see their local special folders, such as Documents and Desktop, from a session. By default, Special Folder Redirection is allowed.

This setting prevents any objects filtered through a policy from having Special Folder Redirection, regardless of settings that exist elsewhere. When this setting is set to 'Prohibited', any related settings specified for Web Interface or Citrix Receiver are ignored.

To define which users can have special folder redirection, select 'Allowed' and include this setting in a policy filtered on the users you want to have this feature. This setting overrides all other special folder redirection settings throughout XenApp.

Because Special Folder Redirection must interact with the client device, settings that prevent users from accessing or saving files to their local hard drives also prevent Special Folder Redirection from working.

For seamless applications and seamless and published desktops, Special Folder Redirection works for Documents and Desktops folders. Citrix does not recommend using Special Folder Redirection with published Windows Explorer.
```

Setting Name: `SpecialFolderRedirection`

Setting Value:
```
{
    enabled = true | false
}
```


### SSL cipher suite
Description:
```
Specifies the SSL cipher suites used by the SSL cryptographic module in the Universal Print Client.

Select "ALL" to use All SSL cipher suites.

Select "COM" to use only Commercial SSL cipher suites.

Select "GOV" to use only Government SSL cipher suites.
```

Setting Name: `UpcSslCipherSuite`

Setting Value: `ALL`, `COM`, `GOV`


### SSL compliance mode
Description:
```
Controls whether the SSL cryptographic module in the Universal Print Client complies with NIST SP 800-52.

Select "OPEN" to specify that the SSL cryptographic module will operate without complying with NIST SP 800-52.

Select "SP800_52" to specify that the SSL cryptographic module should comply with NIST SP 800-52.
```

Setting Name: `UpcSslComplianceMode`

Setting Value: `OPEN`, `SP800_52`


### SSL enabled
Description:
```
Specifies whether the Universal Print Client will use SSL to connect to the Universal Print Server.
```

Setting Name: `UpcSslEnable`

Setting Value:
```
{
    enabled = true | false
}
```

### SSL FIPS mode
Description:
```
If true, the SSL cryptographic module in the Universal Print Client will operate in the FIPS 140 approved mode of operation.
```

Setting Name: `UpcSslFips`

Setting Value:
```
{
    enabled = true | false
}
```

### SSL protocol version
Description:
```
Specifies the SSL protocol version used by the SSL cryptographic module in the Universal Print Client.

Select "ALL" to use all supported TLS protocol versions.

Select "TLS1" to use TLS version 1.0.

Select "TLS11" to use TLS version 1.1.

Select "TLS12" to use TLS version 1.2. 
```

Setting Name: `UpcSslProtocolVersion`

Setting Value: `ALL`, `TLS1`, `TLS11`, `TLS12`


### SSL Universal Print Server encrypted print data stream (CGP) port
Description:
```
Applies to Universal Print Server

Specifies the TCP port number used by the Universal Print Client when connecting to the Universal Print Server's encrypted print data stream (CGP) listener.
The Universal Print Server is an optional component that enables the use of Citrix's universal print drivers for network printing scenarios. When Universal Print Server is used, printing commands are sent from XenApp and XenDesktop hosts to the Universal Print Server via SOAP over HTTP (or HTTPS). However bulk print data streams are delivered to the print server on separate TCP connections using Common Gateway Protocol (CGP). This setting modifies the TCP port to which the Universal Print Client sends encrypted CGP connections (i.e., outbound print jobs).
```

Setting Name: `UpcSslCgpPort`

Setting Value: `{Port Number}`


### SSL Universal Print Server encrypted web service (HTTPS/SOAP) port
Description:
```
Specifies the TCP port number used by the Universal Print Client when connecting to the Universal Print Server's encrypted web service (HTTPS/SOAP) listener. The Universal Print Server is an optional component that enables the use of Citrix universal print drivers for network printing scenarios. When the Universal Print Server is used, printing commands are sent from XenApp and XenDesktop hosts to the Universal Print Server via SOAP over HTTPS. This setting modifies the TCP port to which the Universal Print Client will send HTTPS/SOAP requests. The default value is 8443.

You must configure both host and print server HTTPS port identically . If you do not configure the ports identically, the host software will not connect to the Universal Print Server. This setting changes the VDA on XenApp and XenDesktop. In addition, you must change the default port on the Universal Print Server. For more information, refer to the product documentation.
```

Setting Name: `UpcSslHttpsPort`

Setting Value: `{Port Number}`


### Start Menu path
Description:
```
Lets you specify how to redirect the Start Menu folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRStartMenuPath_Part`

Setting Value: `{Path to the Redirected Start Menu Folder}`


### Streamed user profile groups
Description:
```
Enter one or more Windows user groups.

If this setting is enabled, only the profiles of those groups' members are streamed. If this setting is disabled, all user groups are processed.

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, all users are processed.
```

Setting Name: `PSUserGroups_Part`

Setting Value:
```
jsonencode([
    {Domain 1}\{Group 1},
    {Domain 2}\{Group 2},
    ...
])
```


### Tablet mode toggle
Description:
```
The tablet mode toggle optimizes (on the VDA) the look and behavior of Store apps, Win32 apps and the Windows shell by automatically toggling the virtual desktop to tablet mode when connecting from small form factor devices like phones and tablets (or any touch enabled device).

If you set this setting to disabled, the user can interact with the virtual desktop in the same way as interacting with a PC using input devices like keyboard and mouse (also known as desktop mode).

For a VDA hosted on a virtual machine, the XenServer hypervisor must support Slate mode in the ACPI Table of the Virtualized BIOS/UEFI.

The feature is supported only on Windows 10 VDA and a minimum version of XenServer 7.2.
```

Setting Name: `TabletModeToggle`

Setting Value:
```
{
    enabled = true | false
}
```

### Target frame rate
Description:
```
Sets the maximum number of frames per second that the virtual desktop will send to the client.

If you want to improve the user experience you can increase the maximum FPS to 30. This will consume more resources and bandwidth, but will provide a better user experience.

On the other hand, if you want to maximize server scalability and reduce bandwidth usage at the expense of user experience, you can set the value somewhere between 10 or 15.
```

Setting Name: `FramesPerSecond`

Setting Value: `{Target FPS Value}`


### Target minimum frame rate
Description:
```
Selects the frame rate the system attempts to maintain when adapting to low bandwidth conditions.

For VDA versions 7.0 through 7.6 Feature Pack 2, this setting applies only when legacy graphics mode is enabled. For later VDA versions, this setting applies when the legacy graphics mode is disabled or enabled.
```

Setting Name: `TargetedMinimumFramesPerSecond`

Setting Value: `{Target Minimum FPS Value}`


### Template profile overrides local profile
Description:
```
By default, new user profiles are created from the default user profile on the computer where a user first logs on. Profile management can alternatively use a centrally stored template when creating new user profiles. Template profiles are identical to normal profiles in that they reside in any file share on the network. Use UNC notation to specifying paths to templates. Users need read access to a template profile.

If this setting is disabled, templates are not used.

If this setting is enabled, Profile management uses the template instead of the local default profile when creating new user profiles.

If a user has no Citrix user profile, but a local Windows user profile exists, by default the local profile is used (and migrated to the user store, if this is not disabled). This can be changed by enabling the setting "Template profile overrides local profile".

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, no template is used.
```

Setting Name: `TemplateProfileOverridesLocalProfile`

Setting Value:
```
{
    enabled = true | false
}
```


### Template profile overrides roaming profile
Description:
```
By default, new user profiles are created from the default user profile on the computer where a user first logs on. Profile management can alternatively use a centrally stored template when creating new user profiles. Template profiles are identical to normal profiles in that they reside in any file share on the network. Use UNC notation to specifying paths to templates. Users need read access to a template profile.

If this setting is disabled, templates are not used.

If this setting is enabled, Profile management uses the template instead of the local default profile when creating new user profiles.

If a user has no Citrix user profile, but a local Windows user profile exists, by default the local profile is used (and migrated to the user store, if this is not disabled). This can be changed by enabling the setting "Template profile overrides local profile".

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, no template is used.
```

Setting Name: `TemplateProfileOverridesRoamingProfile`

Setting Value:
```
{
    enabled = true | false
}
```


### Template profile used as a Citrix mandatory profile for all logons
Description:
```
By default, new user profiles are created from the default user profile on the computer where a user first logs on. Profile management can alternatively use a centrally stored template when creating new user profiles. Template profiles are identical to normal profiles in that they reside in any file share on the network. Use UNC notation to specifying paths to templates. Users need read access to a template profile.

If this setting is disabled, templates are not used.

If this setting is enabled, Profile management uses the template instead of the local default profile when creating new user profiles.

If a user has no Citrix user profile, but a local Windows user profile exists, by default the local profile is used (and migrated to the user store, if this is not disabled). This can be changed by enabling the setting "Template profile overrides local profile".

If this setting is not configured here, the value from the .ini file is used.

If this setting is not configured here or in the .ini file, no template is used.
```

Setting Name: `TemplateProfileIsMandatory`

Setting Value:
```
{
    enabled = true | false
}
```


### Timeout for pending area lock files (days)
Description:
```
You can set a timeout period that frees up users' files so they are written back to the user store from the pending area in the event that the user store remains locked when a server becomes unresponsive (for example, when it goes down). Use this setting to prevent bloat in the pending area and to ensure the user store always contains the most up-to-date files.
```

Setting Name: `PSPendingLockTimeout`

Setting Value: `{Timeout in Days}`


### TWAIN compression level
Description:
```
Specifies the level of compression of image transfers from client to server. Use 'Low' for best image quality, 'Medium' for good image quality, or 'High' for low image quality. By default, medium compression is applied.
```

Setting Name: `TwainCompressionLevel`

Setting Value: `None`, `Low`, `Medium`, `High`


### TWAIN device redirection bandwidth limit
Description:
```
Specifies the maximum allowed bandwidth in kilobits per second for controlling TWAIN imaging devices from published applications.

If you enter a value for this setting and a value for the 'TWAIN device redirection bandwidth limit percent' setting, the most restrictive setting (with the lower value) is applied.
```

Setting Name: `TwainBandwidthLimit`

Setting Value: `{Bandwidth Limit in Kbps}`

### TWAIN device redirection bandwidth limit percent
Description:
```
Specifies the maximum allowed bandwidth for controlling TWAIN imaging devices from published applications as a percent of the total session bandwidth.

If you enter a value for this setting and a value for the 'TWAIN device redirection bandwidth limit' setting, the most restrictive setting (with the lower value) is applied.

If you configure this setting, you must also configure the 'Overall session bandwidth limit' setting which specifies the total amount of bandwidth available for client sessions.
```

Setting Name: `TwainBandwidthPercent`

Setting Value: `{Bandwidth Limit Percentage Value}`

### Universal driver preference
Description:
```
Specifies the order in which XenApp attempts to use Universal Printer drivers, beginning with the first entry in the list. You can add, edit, or remove drivers, and change the order of drivers in the list.
```

Setting Name: `UniversalDriverPriority`

Setting Value:
```
jsonencode([
    {Driver 1},
    {Driver 2},
    ...
])
```

### Universal print driver usage
Description:
```
Specifies when to use universal printing. Universal printing employs generic printer drivers instead of standard model-specific drivers potentially simplifying burden of driver management on host machines. The availability of a universal print driver depends upon capabilities of client, host, and print server software. In some certain configurations, universal printing may not be available. By default, universal printing is used only if the requested driver is unavailable.

`FallbackToSpecific`: 'Use universal printing only if requested driver is unavailable' uses standard model specific drivers for printer creation if they are available. If a requested driver model is not available, the system will attempt to create the printer with a universal driver.

`SpecificOnly`: 'Use only printer model specific drivers' specifies that the printers should not be created with a universal driver. If the requested driver model is unavailable, the printer will not be created.

`UpdOnly`: 'Use universal printing only' specifies that only universal drivers should be used to create printers. If a suitable universal driver universal printing is not available, then the printer will not be created.

`FallbackToUpd`: 'Use printer model specific drivers only if universal printing is unavailable' employs a universal printer driver if it is available. If a suitable universal driver is not available, the printer may still be created with requested driver model if it is available.
```

Setting Name: `UniversalPrintDriverUsage`

Setting Value: `UpdOnly`, `FallbackToUpd`, `SpecificOnly`, `FallbackToSpecific`


### Universal Print Server enable
Description:
```
Enables (disables) use of Universal Print Server on a XenApp or XenDesktop host. Also controls Universal Print Server interactions with Windows' native remote printing. By default, Universal Print Server is Disabled.

`UpsDisabled`: 'Disabled' System does not attempt to connect with Universal Print Server when connecting to a network printer with a UNC name. Connections to remote printers continue to use Windows' native remote printing facility.

`UpsEnabledWithFallback`: 'Enabled with fallback to Windows' native remote printing' Network printer connections are serviced by Universal Print Server if possible. If the Universal Printer Server is unavailable, then the system falls back to Windows' native remote printing facility.

`UpsOnlyEnabled`: 'Enabled with no fallback to Windows' native remote printing' Network printer connections are serviced by Universal Print Server exclusively. If the Universal Printer Server is unavailable, then the network printer connection fails.
```

Setting Name: `UpsEnable`

Setting Value: `UpsDisabled`, `UpsEnabledWithFallback`, `UpsOnlyEnabled`


### Universal Print Server print data stream (CGP) port
Description:
```
Applies to Universal Print Server

Specifies TCP port number used by Universal Print Server's print data stream (CGP) listener
The Universal Print Server is an optional component that enables the use of Citrix's universal print drivers for network printing scenarios. When Universal Print Server is used, printing commands are sent from XenApp and XenDesktop hosts to the Universal Print Server via SOAP over HTTP. However bulk print data streams are delivered to the print server on separate TCP connections using Common Gateway Protocol (CGP). This setting modifies the default TCP port on which the Universal Print Server listens for incoming CGP connections (e.g. inbound print jobs).
This policy must only be applied to OUs housing the Universal Print Server computers.
```

Setting Name: `UpsCgpPort`

Setting Value: `{Port Number}`


### Universal Print Server print stream input bandwidth limit (Kbps)
Description:
```
Applies to XenApp 6.5 or later with the Universal Print Client and XenDesktop 5.5 or later with the Universal Print Client

Specifies a upper bound in kilobits-per-second for the transfer rate of print data delivered from a XenApp or XenDesktop host session to a Universal Print Server via the CGP protocol. Default value is unlimited (0).
```

Setting Name: `UpsPrintStreamInputBandwidthLimit`

Setting Value: `{Bandwidth Limit in Kbps}`


### Universal Print Server web service (HTTP/SOAP) connect timeout
Description:
```
Specifies the number of seconds that the Universal Print Client should wait until a Universal Print Server web service connect() operation times out.
```

Setting Name: `UpcHttpConnectTimeout`

Setting Value: `{Timeout in Seconds}`


### Universal Print Server web service (HTTP/SOAP) port
Description:
```
Specifies the TCP port number used by the Universal Print Server's web service (HTTP/SOAP) listener. The Universal Print Server is an optional component that enables the use of Citrix universal print drivers for network printing scenarios. When the Universal Print Server is used, printing commands are sent from XenApp and XenDesktop hosts to the Universal Print Server via SOAP over HTTP. This setting modifies the default TCP port on which the Universal Print Server listens for incoming HTTP/SOAP requests. The default value is 8080.

You must configure both host and print server HTTP port identically . If you do not configure the ports identically, the host software will not connect to the Universal Print Server. This setting changes the VDA on XenApp and XenDesktop. In addition, you must change the default port on the Universal Print Server. For more information, refer to the product documentation.
```

Setting Name: `UpsHttpPort`

Setting Value: `{Port Number}`


### Universal Print Server web service (HTTP/SOAP) receive timeout
Description:
```
Specifies the number of seconds that the Universal Print Client should wait until a Universal Print Server web service recv() operation times out.
```

Setting Name: `UpcHttpReceiveTimeout`

Setting Value: `{Timeout in Seconds}`


### Universal Print Server web service (HTTP/SOAP) send timeout
Description:
```
Specifies the number of seconds that the Universal Print Client should wait until a Universal Print Server web service send() operation times out.
```

Setting Name: `UpcHttpSendTimeout`

Setting Value: `{Timeout in Seconds}`


### Universal Print Servers for load balancing
Description:
```
Lists the Universal Print Servers to be used to load balance printer connections established at session launch, after evaluating other Citrix printing policy settings. To optimize printer creation time, Citrix recommends that all print servers have the same set of the shared printers.

This setting also implements print server failover detection and printer connection recovery. The print servers are checked periodically for availability. If a server failure is detected, that server is removed from the load balancing scheme, and printer connections on that server are redistributed among other available print servers. When the failed print server recovers, it is returned to the load balancing scheme.
```

Setting Name: `LoadBalancedPrintServers`

Setting Value:
```
jsonencode([
    {Server 1},
    {Server 2},
    ...
])
```

### Universal Print Servers out-of-service threshold
Description:
```
Specifies the number of seconds the load balancer waits for an unavailable Universal Print Server to recover before it assumes the server is offline. That server's load is then redistributed among other available print servers.
```

Setting Name: `PrintServersOutOfServiceThreshold`

Setting Value: `{Threshold in Seconds}`


### Universal printing EMF processing mode
Description:
```
Controls the method of processing the EMF spool file on the Windows client machine. When using the Citrix Universal Printer driver, the system spools the EMF records on the host and delivers the EMF spool file to the Windows client for processing. Typically, these EMF spool files are injected directly to the client's spool queue. For printers and drivers that are compatible with the EMF format, this is the fastest printing method. However, some drivers may require the EMF spool file to be reprocessed and sent through the GDI subsystem on the client. Normally, this alternate print path is selected automatically. However, since it is not always possible to detect such printers and drivers automatically, this setting can be used to force the EMF spool file to be reprocessed and sent through the GDI subsystem on the client.
```

Setting Name: `EMFProcessingMode`

Setting Value: `ReprocessEMFsForPrinter`, `SpoolDirectlyToPrinter`


### Universal printing image compression limit
Description:
```
Defines the maximum quality and the minimum compression level available for images printed with the Universal Printer driver. By default, the image compression limit is set to 'Best quality (lossless compression).' If 'No compression' is selected, compression is disabled for EMF printing only. Compression is not disabled for XPS printing.

When adding this setting to a policy that includes the 'Universal printing optimization defaults' setting, be aware of the following items:

* If the compression level in the 'Universal printing compression limit' setting is lower than the level defined in 'Universal printing optimization defaults' setting, images are compressed at the level defined in the 'Universal printing compression limits' setting.
* If compression is disabled, the 'Desired image quality' and 'Enable heavyweight compression' options of the 'Universal printing optimization defaults' setting have no effect in the policy.
```

Setting Name: `ImageCompressionLimit`

Setting Value: `NoCompression`, `LosslessCompression`, `MinimumCompression`, `MediumCompression`, `MaximumCompression`

### Universal printing optimization defaults
Description:
```
Specifies the default settings when the Universal Printer is created for a session.

* Desired image quality - Sets the image compression level (EMF, XPS, PDF).
* Enable heavyweight compression - Reduces bandwidth beyond the 'Desired image quality' level without losing image quality (EMF only).
* Allow caching of embedded images - (EMF only).
* Allow caching of embedded fonts - (EMF and PDF). - For the PDF driver, this enables standard and licensed font embedding by default.
* Allow non-administrators to modify these settings - Allows non-administrators to modify these options through the driver's advanced print settings (EMF and PDF).

The value is a set of a key-value pairs separated by commas. The default value of this setting is 'ImageCompression=StandardQuality, HeavyweightCompression=False, ImageCaching=True, FontCaching=True, AllowNonAdminsToModify=False'

When entering the value on a command line, follow the format as specified in the default value above.

Please consult the product documentation for more information about this setting.
```

Setting Name: `UPDCompressionDefaults`

Setting Value: `ImageCompression={ReducedQuality | StandardQuality | HighQuality | BestQuality}, HeavyweightCompression={True | False}, ImageCaching={True | False}, FontCaching={True | False}, AllowNonAdminsToModify={True | False}`


### Universal printing preview preference
Description:
```
Specifies whether to use the print preview function for auto-created or generic universal printers. By default, print preview is not used for auto-created or generic universal printers.
```

Setting Name: `UniversalPrintingPreviewPreference`

Setting Value: `NoPrintPreview`, `AutoCreatedOnly`, `GenericOnly`, `AutoCreatedAndGeneric`


### Universal printing print quality limit
Description:
```
Specifies the maximum dots per inch (dpi) available for generating printed output in the session. By default, no limit is specified.
`Draft`: Maximum 150 DPI
`LowResolution`: Maximum 300 DPI
`MediumResolution`: Maximum 600 DPI
`HighResolution`: Maximum 1200 DPI
```

Setting Name: `DPILimit`

Setting Value: `Draft`, `LowResolution`, `MediumResolution`, `HighResolution`, `Unlimited`


### Upload file for Citrix Workspace app for Chrome OS/HTML5
Description:
```
This setting allows or prevents Citrix Workspace app for Chrome OS/HTML5 users from uploading files from the client device to a virtual desktop. By default, uploading files to virtual desktop is enabled.

When adding this setting to a policy, ensure that the 'File transfer for Citrix Workspace app for Chrome OS/HTML5' setting is present and set to True.
```

Setting Name: `AllowFileUpload`

Setting Value:
```
{
    enabled = true | false
}
```

### URL redirection black list
Description:
```
The URL blacklist specifies URLs that are redirected to the locally-running default web browser.
```

Setting Name: `URLRedirectionBlackList`

Setting Value:
```
jsonencode([
    {URL 1},
    {URL 2},
    ...
])
```

### URL redirection white list
Description:
```
The URL whitelist specifies URLs that are viewed on the default web browser of the environment in which they are launched.
```

Setting Name: `URLRedirectionWhiteList`

Setting Value:
```
jsonencode([
    {URL 1},
    {URL 2},
    ...
])
```

### Use asynchronous writes
Description:
```
Enables or disables asynchronous disk writes. By default, asynchronous writes are disabled.

Asynchronous disk writes can improve the speed of file transfers and writing to client disks over WANs, which are typically characterized by relatively high bandwidth and high latency. However, if there is a connection or disk fault, the client file or files being written may end in an undefined state. If this happens, a pop-up window informs the user of the files affected. The user can then take remedial action, such as restarting an interrupted file transfer on reconnection or when the disk fault is corrected.

Citrix recommends enabling asynchronous disk writes only for users who need remote connectivity with good file access speed and who can easily recover files or data lost in the event of connection or disk failure. When enabling this setting, make sure that the 'Client drive redirection' setting is enabled. If this setting is disabled, asynchronous writes will not occur.
```

Setting Name: `AsynchronousWrites`

Setting Value:
```
{
    enabled = true | false
}
```

### Use GPU for optimizing Windows Media multimedia redirection over WAN
Description:
```
Uses GPU for optimization Windows Media over WAN. Currently Only NVIDIA GPUs are supported
```

Setting Name: `UseGPUForMultimediaOptimization`

Setting Value:
```
{
    enabled = true | false
}
```

### Use hardware encoding for video codec
Description:
```
Enables use of hardware encoding. For Intel Iris Pro graphics processors, hardware encoding is supported with XenApp and XenDesktop VDAs in standard or HDX 3D Pro mode. For NVIDIA GRID GPUs, hardware encoding is supported with XenDesktop VDAs in HDX 3D Pro mode.
```

Setting Name: `UseHardwareEncodingForVideoCodec`

Setting Value:
```
{
    enabled = true | false
}
```

### Use local time of client
Description:
```
Determines the time zone setting of the user session. When enabled, the administrator can choose to default the user session’s time zone settings to that of the user's time zone settings.

On XenDesktop: By default, the user session's time zone is used for the session. 'Use Client Time Zone' will be in effect if this policy is not active or 'Use default value' is checked.

On XenApp or XenDesktop VDA Server OS: By default, the server's time zone is used for the session. For this setting to be enforced, enable the 'Allow time zone redirection' setting in the Remote Desktop Services node of the Group Policy Management Console.
```

Setting Name: `SessionTimeZone`

Setting Value: `UseServerTimeZone`, `UseClientTimeZone`

### Use video codec for compression
Description:
```
This setting is available only on VDA versions XenApp and XenDesktop 7.6 Feature Pack 3 and later.

Allows use of a video codec to compress graphics when video decoding is available on the endpoint. When 'For the entire screen' is chosen the video codec will be applied as the default codec for all content (some small images and text will be optimized and sent losslessly). When 'For actively changing regions' is selected the video codec will be used for areas where there is constant change on the screen, other data will use still image compression and bitmap caching. When video decoding is not available on the endpoint, or when you specify 'Do not use', a combination of still image compression and bitmap caching is used. When 'Use when preferred' is selected, the system chooses, based on various factors. The results may vary between versions as the selection method is enhanced.

Select 'DoNotUseVideoCodec' to optimize for server CPU load and for cases that do not have a lot of server-rendered video or other graphically intense applications.

Select 'UseVideoCodecIfAvailable' to optimize for cases with heavy use of server-rendered video and 3D graphics, especially in low bandwidth.

Select 'ActivelyChangingRegions' to optimize for improved video performance, especially in low bandwidth, while maintaining scalability for static and slowly changing content.

Select 'UseVideoCodecIfPreferred' to allow the system to make its best effort to choose appropriate settings for the current scenario.
```

Setting Name: `UseVideoCodecForCompression`

Setting Value: `DoNotUseVideoCodec`, `UseVideoCodecIfAvailable`, `ActivelyChangingRegions`, `UseVideoCodecIfPreferred`

### User layer exclusions
Description:
```
Excludes a list of files and directories so that they don't persist in the user layer.

Directories are excluded if there is a \ at the end of the path.
Example: C:\Program Files\AntiVirusHome\.

Files are excluded if there is no \ at the end of the path.
Example: C:\ProgramData\AntiVirus\virusdefs.db.

There is no limit to the number of exclusion rules that you can add. You can also use a * as a wildcard in a path. For example, C:\Users\*\AppData\Local\Temp excludes the Temp directory for all users. There is only one * allowed in the rule, and that * only matches one level of directories.
```

Setting Name: `UplUserExclusions`

Setting Value:
```
jsonencode([
    {Path 1},
    {Path 2},
    ...
])
```

### User layer repository path
Description:
```
The SMB directory path where user layer VHDs are located. Format:'\\\\server\share\path'
```

Setting Name: `UplRepositoryPath`

Setting Value: `{SMB Directory Path}`

### User layer size in GB
Description:
```
The size (in GB) of each new user layer disk. The value must be between 10GB and 2040GB. Defaults to 10GB
```

Setting Name: `UplUserLayerSizeInGb`

Setting Value: `{User Layer Size in GB}`

### UWP app roaming
Description:
```
Lets UWP (Universal Windows Platform) apps roam with users. As a result, users can access the same UWP apps from different devices.

With this policy enabled, Profile Management lets UWP apps roam with users by storing the apps on separate VHDX disks. Those disks are attached during user logons and detached during user logoffs.

If this setting is not configured here, the value from the .ini file is used.

If this setting is configured neither here nor in the .ini file, this feature is disabled.
```

Setting Name: `UwpAppsRoaming`

Setting Value:
```
{
    enabled = true | false
}
```

### VDA data collection for Analytics
Description:
```
VDA Data Collection by Monitor Service for Analytics. Changing this setting to disabled stops VDA data flow to Analytics. Metrics like Network Latency, Input and Output Bandwidth details will not be available.
```

Setting Name: `VdcPolicyEnable`

Setting Value:
```
{
    enabled = true | false
}
```

### Videos path
Description:
```
Lets you specify how to redirect the Videos folder. To do so, select Enabled and then type the redirected path.
Caution: Potential data loss might occur. See below for details.
You might want to modify the path after the policy takes effect. However, consider potential data loss before you do so. The data contained in the redirected folder might be deleted if the modified path points to the same location as the previous path.
For example, suppose you specify the Contacts path as path1. Later, you change path1 to path2. If path1 and path2 point to the same location, all data contained in the redirected folder is deleted after the policy takes effect.
To avoid potential data loss, complete the following steps:
1. Apply Microsoft policy to machines where Profile Management is running through Active Directory Group Policy Objects. To do so, open the Group Policy Management Console, navigate to Computer Configuration > Administrative Templates > Windows Components > File Explorer, and then enable Verify old and new Folder Redirection targets point to the same share before redirecting.
2. If applicable, apply hotfixes to machines where Profile Management is running. For details, see https://support.microsoft.com/en-us/help/977229 and https://support.microsoft.com/en-us/help/2799904.
```

Setting Name: `FRVideosPath_Part`

Setting Value: `{Path to the Redirected Video Folder}`

### View window contents while dragging
Description:
```
Controls the display of window content when dragging a window across the screen.

When allowed, the entire window appears to move when you drag it. When prohibited, only the window outline appears to move until you drop it. By default, viewing window contents is allowed.
```

Setting Name: `WindowContentsVisibleWhileDragging`

Setting Value:
```
{
    enabled = true | false
}
```

### Virtual channel allow list
Description:
```
Enables the use of an allow list that specifies which virtual channels are allowed to be opened in an ICA session.

When disabled, all virtual channels are allowed.

When enabled, only Citrix virtual channels are allowed. To use custom or third-party virtual channels with this feature, the virtual channels must be added to the list. To add a virtual channel to the list, enter the virtual channel name, followed by a comma, and the path to the process that accesses the virtual channel. Additional executable paths can be listed and the paths are separated by commas.

Examples:
CTXCVC1,C:\VC1\vchost.exe
CTXCVC2,C:\VC2\vchost.exe,C:\Program Files\Third Party\vcaccess.exe
```

Setting Name: `VirtualChannelWhiteList`

Setting Value:
```
jsonencode([
    "{Virtual Channel Name 1},{App Path 1}",
    "{Virtual Channel Name 2},{App Path 2}",
    ...
])
```

Setting Value for disable:
```
jsonencode([
    "=disabled="
])
```

### Virtual channel allow list log throttling
Description:
```
This setting determines the throttle time for logging events for a session. Events for each virtual channel configured in virtual channel allow list are throttled for the duration specified in this setting. The throttle timer will be reset upon disconnect/logoff.
```

Setting Name: `VirtualChannelWhiteListLogThrottling`

Setting Value: `{Throttle Time in Hours}`

### Virtual channel allow list logging
Description:
```
Disabled: disables logging events to event viewer for custom virtual channels
Log warnings only: logs warning events to event viewer for custom virtual channels that are not in the virtual channel allow list
Log all events: logs all events to event viewer for custom virtual channels
```

Setting Name: `VirtualChannelWhiteListLogging`

Setting Value: `Disabled`, `LogWarningsOnly` `LogAllEvents`

### Virtual IP loopback support
Description:
```
Allows each session to have its own virtual loopback address for communication. Configure list of programs via the 'Virtual IP virtual loopback programs list' policy.
```

Setting Name: `VirtualLoopbackSupport`

Setting Value:
```
{
    enabled = true | false
}
```

### Virtual IP virtual loopback programs list
Description:
```
Allows each session to have its own virtual loopback address for communication. You must enable the policy 'Virtual IP loopback support' for this setting to take effect.
```

Setting Name: `VirtualLoopbackPrograms`

Setting Value:
```
jsonencode([
    {Program Path 1},
    {Program Path 2},
    ...
])
```

### Visual quality
Description:
```
The desired visual quality of the images. Higher visual quality will result in increased bandwidth usage. For improved responsiveness in constrained bandwidth scenarios use lower visual quality. By default medium quality is selected.

In cases where preserving image data is vital (for example, when displaying X-ray images where no loss of quality is acceptable), Always lossless should be selected, to assure no lossy data is ever sent. Build to lossless may send lossy images when there are cases of high activity, but will improve quality to lossless once activity ceases.

If 'Always lossless' or 'Build to lossless' is selected, the effect of this setting is changed if the 'Allow visually lossless compression' setting is configured and set to true. In this case, visually lossless compression is used instead of true lossless compression. 'Always lossless' is then always visually lossless and 'Build to lossless' is then build to visually lossless. For more information on true lossless vs. visually lossless, see the help for 'Allow visually lossless compression' setting.
Related setting
```

Setting Name: `VisualQuality`

Setting Value: `Low`, `Medium`, `High`, `BuildToLossless`, `AlwaysLossless`

### Wait for printers to be created (server desktop)
Description:
```
Allows or prevents a delay in connecting to a session so that desktop printers can be auto-created. By default, a connection delay does not occur. This setting applies only to published server desktops.
```

Setting Name: `WaitForPrintersToBeCreated`

Setting Value:
```
{
    enabled = true | false
}
```

### Watermark custom text
Description:
```
Custom text that is displayed in watermark text labels. It should not exceed 25 characters, otherwise it will be truncated.

Additional customization can be achieved using custom text policy from CVAD version 2206 onwards. These customizations can be achieved by using custom tags in the text. Max number of characters is increased to 1024.

These are available tags for watermark settings:

<font=value> - change font of watermark text. value is name of a font available on the VDA. example: <font=Courier New>
<fontzoom=value> - font zoom factor. value will be 200 for 200% zoom on watermark text. example: <fontzoom=200>
<position=value> - change position of watermark text. value can be center, topleft, topright, bottomleft, bottomright. only applicable with single style. example: <position=topright>
<rotation=value> - rotate watermark text. value is in degree and can be between -360 and 360. example: <rotation=45>
<style=value> - change display style. value can be single, xstyle, tile. This tag will override WatermarkStyle policy. example: <style=single>

With single style configuration, a single watermark text label will be displayed at the center of the session, position tag can be used to change the location. With xstyle configuration, five watermark labels will be displayed in the session, with one in the center and four in the corners. With tile configuration, multiple labels will be displayed in session, watermark will be equally spaced on the entire screen.

These are available tags for changing watermark text:

<username> - user name
<domain> - domain name of logged in user account
<hostname> - machine name of VDA
<clientip> - ip address of end point
<serverip> - ip address of VDA
<date> - date when the session was established
<time> - time when the session was established
<newline> - create next line

If tags for watermark text are used then all other watermark policy related to text will be ignored.

Examples:

1. change font of default text
<font=Courier New>
2. show only username with 200% zoom
<username><fontzoom=200>
3. show username in first line and date in second line using "Consolas" font
<username><newline><date><font=Consolas>

Tags for watermark settings can be used with other watermark policies. Unsupported tags will be shown as regular text.

Watermark text should not exceed 25 characters.
```

Setting Name: `WatermarkCustomText`

Setting Value: `{Watermark Text}`

### Watermark transparency
Description:
```
You can specify watermark opacity from 0 - 100. The larger the value specified, the more opaque the watermark.
This setting is effective only when session watermark is enabled.
```

Setting Name: `WatermarkTransparency`

Setting Value: `{Watermark Opacity}`

### WebSockets connections
Description:
```
Allow or prevent WebSockets connections.

The WebSocket Protocol enables two-way communication between browser-based applications and servers without opening multiple HTTP connections. Having fewer connections enhances security and reduces overhead on the XenApp server.
```

Setting Name: `AcceptWebSocketsConnections`

Setting Value:
```
{
    enabled = true | false
}
```

### WebSockets port number
Description:
```
TCP port number for incoming WebSockets connections.
```

Setting Name: `WebSocketsPort`

Setting Value: `{TCP Port Number}`

### WebSockets trusted origin server list
Description:
```
Comma-separated list of trusted origin servers expressed as URLs with the option of using wildcards.

It is usually the address of the Receiver for Web site. Only Websockets connections originating from one of these addresses will be accepted by XenApp. The generic syntax for this address is:<protocol>://<FQDN of host>:<port>/ The protocols should be HTTP or HTTPS. Port is an optional variable and if it is not specified, ports 80 and 443 are used for HTTP and HTTPS respectively. Wildcard characters can also be used to extend this syntax. If this field contains just a '*', it indicates that connections from all origin servers will be accepted.

Examples
https://abc.domain.com:8080/
http://abc.def.domain.com:8080/
https://hostname.domain.com:8081/
https://*.trusteddomain.com/
'*' is not treated as wild card character if used with IP.
'http://10.105.*.*' is an invalid trusted origin as per current design.
```

Setting Name: `WSTrustedOriginServerList`

Setting Value: `https://abc.domain.com:8080/`

Related Settings:
```
WebSockets connections

WebSockets port number
```

### WIA redirection
Description:
```
Enables or disables WIA scanner redirection. When enabled WIA scanner will be redirected to the virtual session and users will be able to scan using Windows Image Acquisition (WIA) from the Virtual machine.
```

Setting Name: `AllowWIARedirection`

Setting Value:
```
{
    enabled = true | false
}
```

### Windows Media client-side content fetching
Description:
```
This setting enables a user device to stream multimedia files directly from the source provider on the Internet or Intranet, rather than through the host server.

By default, the streaming of multimedia files to the user device direct from the source provider is allowed.

Allowing this setting improves network utilization and server scalability by moving any processing on the media from the host server to the user device. It also removes the requirement that an advanced multimedia framework such as Microsoft DirectShow or Media Foundation be installed on the user device; the user device requires only the ability to play a file from a URL.

When adding this setting to a policy, make sure the Windows Media Redirection setting is present and set to Allowed. If this setting is disabled, the streaming of multimedia files to the user device direct from the source provider is also disabled.
```

Setting Name: `MultimediaAccelerationEnableCSF`

Setting Value:
```
{
    enabled = true | false
}
```

### Windows Media fallback prevention
Description:
```
This setting is available only on VDA versions XenApp and XenDesktop 7.6 Feature Pack 3 and later.

Specify the video load management as the following:

`Unconfigured`: 'Not Configured' - equivalent to 'Play all content': if you do not configure this policy setting, the 'Play all content' methods are used.

`SFSR`: 'Play all content' - attempt client-side content fetching, then Windows Media Redirection. If unsuccessful, play on server.

`SFCR`: 'Play all content only on client' - attempt client-side fetching, then Windows Media Redirection. If unsuccessful, content will not play.

`CFCR`: 'Play only client-accessible content on client' - attempt only client-side fetching. If unsuccessful, content will not play.
```

Setting Name: `VideoLoadManagement`

Setting Value: `CFCR`, `SFCR`, `SFSR`, `Unconfigured`

### Windows Media redirection
Description:
```
Controls and optimizes the way Virtual Apps servers deliver streaming audio and video to users. By default, this setting is allowed.

Allowing this setting increases the quality of audio and video rendered from the server to a level that compares with audio and video played locally on a client device. Virtual Apps streams multimedia to the client in the original, compressed form and allows the client device to decompress and render the media.

Windows Media Redirection optimizes multimedia files that are encoded with codecs that adhere to Microsoft's DirectShow, DirectX Media Objects, and Media Foundation standards. To play back a given multimedia file, a codec compatible with the encoding format of the multimedia file must be present on the client device.

By default, audio is disabled on Citrix Receiver. To allow users to run multimedia applications in ICA sessions, turn on audio or give the users permission to turn on audio themselves in their plug-in interface.

Choose 'Prohibited' only if playing media using Windows Media Redirection sounds worse than when rendered using basic ICA compression and regular audio. This is rare but can happen under low bandwidth conditions; for example, with media in which there is a very low frequency of key frames.
```

Setting Name: `MultimediaAcceleration`

Setting Value:
```
{
    enabled = true | false
}
```
