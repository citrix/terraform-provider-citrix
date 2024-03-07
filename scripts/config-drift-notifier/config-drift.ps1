
# Copyright © 2024. Citrix Systems, Inc. All Rights Reserved.
<#
This script is a template for creating custom scripts to notify the user about configuration drift.
.SYNOPSIS
    <#
    Script to notify the user about configuration drift, which occurs when the real-world state of your infrastructure differs from the state defined in your configuration.
    
.DESCRIPTION 
    The script should be run by a wrapping automation tool to notify the user about configuration drift. The script uses the Terraform plan command to identify configuration drift and sends a notification to the user using a Slack webhook URL.
    A list can be provided to notify the user about the resources only in the list. The list should be in the format of a comma-separated string. For example, "resource1,resource2,resource3".

.Parameter SlackWebhookUrl
    The Slack webhook URL to send the notification. The script can be modified to use other messaging services.

.Parameter FilterList
    The filter to notify the user about only the resources in the list. The list should be in the format of a comma-separated string. For example, "resource1,resource2,resource3". 

#>  

[CmdletBinding()]
Param (
    [Parameter(Mandatory = $false)]
    [string] $FilterList,

    [Parameter(Mandatory = $true)]
    [string] $SlackWebhookUrl
)

# Define the command and parameters
$command = "terraform"
$parameters = "plan", "-refresh-only", "-no-color"
# Run the command and capture the output
$output = & $command $parameters

# Initialize a flag variable
$flag = $false

# Initialize an empty array to store the relevant lines
$relevantLines = @()
$skipFlag = $false

foreach ($line in $output) {
    if ($line.Trim().StartsWith("This is a refresh-only plan")) {
        $flag = $false
        break
    }

    if ($flag) {
        if($line.Trim().StartsWith("#")) {
            if($FilterList -ne ""){             #if filter list is provided
                $skipFlag = $true
                foreach ($filter in $FilterList.Split(',')) {
                    if ($line -like "*$filter*") {
                        $skipFlag = $false
                        $relevantLines += "*"+$line+ "*"
                    }
                }
            }else{                               #if filter list is not provided   
                $skipFlag = $line.Trim().EndsWith("deleted")
                $relevantLines += "*"+$line+ "*"
            }
        }
        elseif (-not $skipFlag) {
            if($line.Trim().StartsWith("+")){
                $relevantLines += "*"+$line+ "*"
            }else{
                $relevantLines += $line
            }
            
       }
    }

    if ($line.Trim().EndsWith("which may have affected this plan:")) {
        $relevantLines += "Terraform Config Drift Detected: `n ---------------------------------------------------------------------------------------------------"
        $flag = $true
    }
}



$relevantString = $relevantLines -join "`n"

#slack notification, feel free to modify the payload as per your requirement or use other messaging services
if("" -ne $relevantString){
    $payload = ConvertTo-Json -InputObject @{ text = $relevantString}
    Invoke-RestMethod -Uri $SlackWebhookUrl -Method Post -Body $payload -ContentType 'application/json'
}
