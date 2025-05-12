# Helper Script to fetch CVAD Object IDs from Web Studio

This helper script is designed to fetch the CVAD object IDs from the Webstudio. It collects the list of resources from DDC, and generates a JSON consisting of the object ids for all the object types. Please note that this script is still in **tech preview**.

## Environment Requirements

- PowerShell version `5.0` or higher
- For On-Premises Customers: CVAD DDC `version 2311` or newer.

## Getting Started

1. (Cloud only) create a [Citrix Cloud service principal](https://developer-docs.citrix.com/en-us/citrix-cloud/citrix-cloud-api-overview/get-started-with-citrix-cloud-apis#citrix-cloud-api-access-with-service-principals) with at least the `Read Only Administrator` role to the DDC. This will be used for the `ClientId` and `ClientSecret` in the next step.

## Process to obtain the Object IDs

1. Create a new folder for your Object Ids project.
2. Copy the `cvad-object-ids.ps1` script to the Object Ids project directory created in step 1.
3. Open a PowerShell session.
4. Navigate to the directory where the `cvad-object-ids.ps1` script is located.
5. Set the execution policy by running the following command in the PowerShell session:
    ```powershell
    Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass -Force
    ```
6. Run the script with the following command:
  - For Citrix Cloud customers
    ```powershell
    .\cvad-object-ids.ps1 -CustomerId "{CustomerId}" -ClientId "{ClientId}" -ClientSecret "{ClientSecret}" -Environment "{Environment}"
    ```
  - For Citrix on-premises customers
    ```powershell
    .\cvad-object-ids.ps1 -ClientId "{ClientId}" -ClientSecret "{ClientSecret}" -DomainFqdn "{Domain FQDN}" -HostName "{HostName}"
    ```
    Replace the placeholders `{...}` with your actual values. Here's what each parameter means:
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
      - For Citrix Cloud customers **only** (Optional): Your Citrix Cloud environment. The available options are `Production` and `Staging`. The default value is `Production`.
    - `DisableSSLValidation` (Switch):
      - For Citrix on-premises customers **only**: Add this switch to disable SSL validation on both the PowerShell session and the provider client. SSL validation has to be disabled for this script to work if your on-premises DDC does not have a valid SSL certificate.

7. Once the script has finished running, check the `object_ids.json` file for the object ids. The json file will consist of a list of object types with each object type having a nested list of objects consisting of object name and the corresponding object id.


## Known Issues/Debugging:
1. While running the script for On-Premises customers if it throws an exception as stated below:

```
Invoke-WebRequest : The underlying connection was closed: Could not establish trust relationship for the SSL/TLS secure channel.
```
Solution : 
Disable SSL validation by adding `-DisableSSLValidation` to the command.
```powershell
.\cvad-object-ids.ps1 -ClientId "{ClientId}" -ClientSecret "{ClientSecret}" -DomainFqdn "{Domain FQDN}" -HostName "{HostName}" -DisableSSLValidation
```