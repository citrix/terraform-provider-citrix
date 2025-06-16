# Citrix Provider Site Onboarding Automation

This automation script is designed to onboard an existing site to Terraform. It collects the list of resources from DDC, imports them into Terraform, and generates the Terraform skeletons for the site resources. Please note that this onboarding script is still in **tech preview**.

## Environment Requirements

- PowerShell version `5.0` or higher
- Citrix Provider version `1.0.21`
- For On-Premises Customers: CVAD DDC `version 2311` or newer.

## Workflow:

![](./images/Onboarding%20Automation%20Workflow.png)

## Getting Started

1. Create a new folder for your Terraform project.
2. Initialize Terraform in the newly created folder by running the following command:
  ```shell
  terraform init
  ```
3. Set up the Citrix Terraform provider locally. For instructions, refer to the [Citrix Provider Documentation](https://registry.terraform.io/providers/citrix/citrix/latest/docs).
4. (Cloud only) create a [Citrix Cloud service principal](https://developer-docs.citrix.com/en-us/citrix-cloud/citrix-cloud-api-overview/get-started-with-citrix-cloud-apis#citrix-cloud-api-access-with-service-principals) with at least the `Read Only Administrator` role to the DDC. This will be used for the `ClientId` and `ClientSecret` in the next step.

## Recommended Approach for New Users

For administrators who are new to Terraform and Citrix, we recommend starting small by onboarding a limited set of resources rather than your entire Citrix site at once. This approach makes the learning process more manageable and less overwhelming.

We recommend using the following parameters when running the onboarding script:

- Use the `-ResourceTypes` parameter to specify just a few resource types (for example `citrix_zone`, `citrix_delivery_group`)
- Use the `-NamesOrIds` parameter to filter for specific resources by name or ID

Example:
```powershell
.\terraform-onboarding.ps1 -CustomerId "{CustomerId}" -ClientId "{ClientId}" -ClientSecret "{ClientSecret}" -ResourceTypes "citrix_zone","citrix_delivery_group" -NamesOrIds "Primary Zone","Sales Delivery Group"
```

This incremental approach allows you to become familiar with Terraform concepts and the Citrix provider while working with a smaller, more focused set of resources.

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
    .\terraform-onboarding.ps1 -CustomerId "{CustomerId}" -ClientId "{ClientId}" -ClientSecret "{ClientSecret}" -Environment "{Environment}"
    ```
  - For Citrix on-premises customers
    ```powershell
    .\terraform-onboarding.ps1 -ClientId "{Username}" -ClientSecret "{Password}" -DomainFqdn "{Domain FQDN}" -HostName "{HostName}"
    ```
    Replace the placeholders `{...}` with your actual values. Here's what each parameter means, in addition to the optional parameters:
    - `CustomerId`: 
      - For Citrix Cloud customers **only** (Required): Your Citrix Cloud customer ID. This is only applicable for Citrix Cloud customers.
    - `ClientId`: Your client ID for Citrix DaaS service authentication.
      - For Citrix On-Premises customers: Use this to specify a DDC administrator username.
      - For Citrix Cloud customers: Use this to specify a Citrix Cloud service principal ID.
    - `ClientSecret`: Your client secret for Citrix DaaS service authentication.
      - For Citrix on-premises customers: Use this to specify a DDC administrator password.
      - For Citrix Cloud customers: Use this to specify Citrix Cloud service principal secret.
    - `DomainFqdn`: The domain FQDN of the Active Directory. Only required for on-premises customers.
      - For Citrix on-premises customers **only** (Required): Use this to specify Domain FQDN.
    - `HostName`: The host name or base URL of your Citrix DaaS service.
      - For Citrix on-premises customers (Required): Use this to specify Delivery Controller hostname.
      - For Citrix Cloud customers (Optional): Use this to force override the Citrix DaaS service hostname.
    - `Environment`: 
      - For Citrix Cloud customers **only** (Optional): Your Citrix Cloud environment. The available options are `Production`, `Japan`, and `Gov`. The default value is `Production`.
    - `ResourceTypes` (Array):
      - Optional list of resource types to onboard. When specified, only those resources will be onboarded, the rest skipped. This helps make the onboarding process more manageable by limiting the scope.
      - By default (if `-NoDependencyRelationship` is not specified), will resolve all dependency relationships between resources as long as the dependent resource is included.
      - Available resource types include: `citrix_admin_folder`, `citrix_admin_role`, `citrix_admin_scope`, `citrix_admin_user`, `citrix_application`, `citrix_application_group`, `citrix_application_icon`, `citrix_aws_hypervisor`, `citrix_azure_hypervisor`, `citrix_delivery_group`, `citrix_gcp_hypervisor`, `citrix_image_definition`, `citrix_machine_catalog`, `citrix_nutanix_hypervisor`, `citrix_openshift_hypervisor`, `citrix_policy_set`, `citrix_scvmm_hypervisor`, `citrix_service_account`, `citrix_storefront_server`, `citrix_tag`, `citrix_vsphere_hypervisor`, `citrix_wem_configuration_set`, `citrix_wem_directory_object`, `citrix_xenserver_hypervisor`, `citrix_zone`.
      - `citrix_<hypervisorType>_resource_pools` are included with the `citrix_<hypervisorType>_hypervisor` resource.
      - `citrix_image_version` is included with the `citrix_image_definition` resource.
    - `NamesOrIds` (Array):
      - Optional string array parameter to filter resources by name or ID. Only resources with a Name or ID matching any of these values will be onboarded.
      - This allows you to onboard multiple specific resources by name or ID in a single operation.
      - By default if (`-NoDependencyRelationship` is not specified), will resolve all dependency relationships between resources as long as the dependent resource is included.
    - `NoDependencyRelationship` (Switch): Add this switch to remove dependency relationships between resources by replacing resource references with resource IDs.
    - `DisableSSLValidation` (Switch):
      - For Citrix on-premises customers **only**: Add this switch to disable SSL validation on both the PowerShell session and the provider client. SSL validation has to be disabled for this script to work if your on-premises DDC does not have a valid SSL certificate.
    - `ShowClientSecret` (Switch):
      - Use this switch to include the client secret value in the generated Terraform configuration file. Defaults to `$false` to ensure security.

7. Wait for the script to complete. The execution time will depend on the complexity of the onboarding process and the resources being imported.
8. Once the script has finished running, check the `.tf` files for the output. The Terraform state file should also be updated with the site terraform resources.
9. Please note that the onboarding script masks out values for all sensitive attributes present in the generated terraform files. Please update these placeholders with the appropriate values.
10. At this point if you run `terraform plan`, you should **only** see the sensitive attributes from step 9 being updated.
11. Run `terraform apply`. This will synchronize the state file with the values of the sensitive attributes updated in step 9.
12. If you run `terraform plan` again, you should see the following message: `No changes. Your infrastructure matches the configuration.`. This indicates that all the Citrix resources have been successfully onboarded.

## Examples
```powershell
# Cloud customer importing all resources
.\terraform-onboarding.ps1 -CustomerId "{CustomerId}" -ClientId "{ClientId}" -ClientSecret "{ClientSecret}"

# Cloud customer import all resources without dependency relationships
.\terraform-onboarding.ps1 -CustomerId "{CustomerId}" -ClientId "{ClientId}" -ClientSecret "{ClientSecret}" -NoDependencyRelationship

# Cloud customer importing a subset of resources
.\terraform-onboarding.ps1 -CustomerId "{CustomerId}" -ClientId "{ClientId}" -ClientSecret "{ClientSecret}" -ResourceTypes "citrix_zone","citrix_delivery_group" -NamesOrIds "Primary Zone","Sales Delivery Group"

# Cloud customer importing a subset of resources without dependency relationships
.\terraform-onboarding.ps1 -CustomerId "{CustomerId}" -ClientId "{ClientId}" -ClientSecret "{ClientSecret}" -ResourceTypes "citrix_zone","citrix_delivery_group" -NamesOrIds "Primary Zone","Sales Delivery Group" -NoDependencyRelationship

# On-premises customer importing all resources
.\terraform-onboarding.ps1 -ClientId "{Username}" -ClientSecret "{Password}" -DomainFqdn "{Domain FQDN}" -HostName "{HostName}"

# On-premises customer import all resources without dependency relationships
.\terraform-onboarding.ps1 -ClientId "{Username}" -ClientSecret "{Password}" -DomainFqdn "{Domain FQDN}" -HostName "{HostName}" -NoDependencyRelationship

# On-premises customer importing a subset of resources
.\terraform-onboarding.ps1 -ClientId "{Username}" -ClientSecret "{Password}" -DomainFqdn "{Domain FQDN}" -HostName "{HostName}" -ResourceTypes "citrix_zone","citrix_delivery_group" -NamesOrIds "Primary Zone","Sales Delivery Group"

# On-premises customer importing a subset of resources without dependency relationships
.\terraform-onboarding.ps1 -ClientId "{Username}" -ClientSecret "{Password}" -DomainFqdn "{Domain FQDN}" -HostName "{HostName}" -ResourceTypes "citrix_zone","citrix_delivery_group" -NamesOrIds "Primary Zone","Sales Delivery Group" -NoDependencyRelationship
```

## Known Issues/Debugging:
1. While running the script for On-Premises customers if it throws an exception as stated below:

```
Invoke-WebRequest : The underlying connection was closed: Could not establish trust relationship for the SSL/TLS secure channel.
```
Solution : 
Disable SSL validation by adding `-DisableSSLValidation` to the command.
```powershell
.\terraform-onboarding.ps1 -ClientId "{Username}" -ClientSecret "{Password}" -DomainFqdn "{Domain FQDN}" -HostName "{HostName}" -DisableSSLValidation
```