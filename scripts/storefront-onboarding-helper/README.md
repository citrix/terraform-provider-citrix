# Storefront Provider Deployment Onboarding Automation

This automation script is designed to onboard an existing storefront deployment to Terraform. It collects the list of resources from Storefront server, imports them into Terraform, and generates the Terraform skeletons for the site resources. Please note that this storefront onboarding script is still in **Tech Preview**.

## Environment Requirements

- PowerShell version `5.0` or higher
- Citrix Provider version `0.6.3` or higher
- Connection between local machine to the remote Storefront Server

## Getting Started

1. Create a new folder for your Terraform project.
2. Initialize Terraform in the newly created folder by running the following command:
  ```shell
  terraform init
  ```
3. Set up the Storefront Terraform provider locally. For instructions, refer to the storefront config section of [StoreFront Terraform Documentation](https://github.com/citrix/terraform-provider-citrix/blob/main/StoreFront.md).

## Onboarding Process

1. Create a new folder for your Terraform project.
2. Copy the `storefront-terraform-onboarding.ps1` script and `terraform.tf` to the terraform project directory created in step 1.
3. Open a PowerShell session with **Administrator privileges**.
4. Navigate to the directory where the `storefront-terraform-onboarding.ps1` script is located.
5. Set the execution policy by running the following command in the PowerShell session:
    ```powershell
    Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass -Force
    ```
6. Run the script with the following command:
    ```powershell
    .\storefront-terraform-onboarding.ps1 -StorefrontHostname "{SFServerHostName}" -ADAdminUsername "{ADAdminUsername}"-ADAdminPassword "{ADAdminPassword}" 
    ```
    Replace the placeholders `{...}` with your actual values. Here's what each parameter means:
    - `SFServerHostName`: The Storefront Server Hostname. This can be an IP address or a FQDN
    - `ADAdminUsername`: The Active Directory Admin Username for authentication with the Storefront Server.
    - `ADAdminPassword`: The Active Directory Admin Password for authentication with the Storefront Server.
    - `DisableSSLValidation` The switch parameter to disable SSL verification when connecting to the target StoreFront server. 

7. Wait for the script to complete. The execution time will depend on the complexity of the onboarding process and the resources being imported.

8. Once the script has finished running, check the `resource.tf` file for the output. The Terraform state file should also be updated with the site terraform resources.

