# Citrix Provider Site Onboarding Automation

This automation script is designed to onboard an existing site to Terraform. It collects the list of resources from DDC, imports them into Terraform, and generates the Terraform skeletons for the site resources. Please note that this onboarding script is still in **tech preview**.

## Environment Requirements

- PowerShell version `5.0` or higher
- Citrix Provider version `0.5.2` or higher
- For On-Premises Customers: CVAD DDC `version 2308` or newer.

## Workflow:

![](./images/Onboarding%20Automation%20Workflow.png)

## Getting Started

1. Create a new folder for your Terraform project.
2. Initialize Terraform in the newly created folder by running the following command:
  ```shell
  terraform init
  ```
3. Set up the Citrix Terraform provider locally. For instructions, refer to the [Citrix Provider Documentation](https://registry.terraform.io/providers/citrix/citrix/latest/docs).

## Onboarding Process

1. Create a new folder for your Terraform project.
2. Copy the `terraform-onboarding.ps1` script and `terraform.tf` to the terraform project directory created in step 1.
3. Open a PowerShell session with **Administrator privileges**.
4. Navigate to the directory where the `terraform-onboarding.ps1` script is located.
5. Set the execution policy by running the following command in the PowerShell session:
    ```powershell
    Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass -Force
    ```
6. Run the script with the following command:
  - For Citrix Cloud customers
    ```powershell
    .\terraform-onboarding.ps1 -CustomerId "{CustomerId}" -ClientId "{ClientId}" -ClientSecret "{ClientSecret}" -HostName "{HostName}" -Environment "{Environment}"
    ```
  - For Citrix on-premises customers
    ```powershell
    .\terraform-onboarding.ps1 -ClientId "{ClientId}" -ClientSecret "{ClientSecret}" -DomainFqdn "{Domain FQDN}" -HostName "{HostName}"
    ```
    Replace the placeholders `{...}` with your actual values. Here's what each parameter means:
    - `CustomerId`: 
      - For Citrix Cloud customers **only** (Required): Your Citrix Cloud customer ID. This is only applicable for Citrix Cloud customers.
    - `ClientId`: Your client ID for Citrix DaaS service authentication.
      - For Citrix On-Premises customers: Use this to specify a DDC administrator username.
      - For Citrix Cloud customers: Use this to specify Cloud API Key Client Id.
    - `ClientSecret`: Your client secret for Citrix DaaS service authentication.
      - For Citrix on-premises customers: Use this to specify a DDC administrator password.
      - For Citrix Cloud customers: Use this to specify Cloud API Key Client Secret.
    - `DomainFqdn`: Your client secret for Citrix DaaS service authentication.
      - For Citrix on-premises customers **only** (Required): Use this to specify Domain FQDN.
    - `HostName`: The host name or base URL of your Citrix DaaS service.
      - For Citrix on-premises customers (Required): Use this to specify Delivery Controller hostname.
      - For Citrix Cloud customers (Optional): Use this to force override the Citrix DaaS service hostname.
    - `Environment`: 
      - For Citrix Cloud customers **only** (Optional): Your Citrix Cloud environment. The available options are `Production` and `Staging`. The default value is `Production`.
    - `SetDependencyRelationship` (Switch): Add this switch to create dependency relationships between resources by replacing resouce ID with resource references.
    - `DisableSSLValidation` (Switch): 
      - For Citrix on-premises customers **only**: Add this switch to disable SSL validation on both the PowerShell session and the provider client. SSL validation has to be disabled for this script to work if your on-premises DDC does not have a valid SSL certificate.

7. Wait for the script to complete. The execution time will depend on the complexity of the onboarding process and the resources being imported.

8. Once the script has finished running, check the `resource.tf` file for the output. The Terraform state file should also be updated with the site terraform resources.


## Known Issues/Debugging:
1. While running the script for On-Premises customers if it throws an exception as stated below:

```
Invoke-WebRequest : The underlying connection was closed: Could not establish trust relationship for the SSL/TLS secure channel.
```
Solution : 
Disable SSL validation by adding `-DisableSSLValidation` to the command.
```powershell
.\terraform-onboarding.ps1 -ClientId "{ClientId}" -ClientSecret "{ClientSecret}" -DomainFqdn "{Domain FQDN}" -HostName "{HostName}" -DisableSSLValidation
```