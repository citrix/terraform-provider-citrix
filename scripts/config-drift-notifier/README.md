# Citrix Provider Config Drift Notification

This script is designed to notify the customer the configuration drift, which occurs when the real-world state of your infrastructure differs from the state defined in your configuration.
Please note that this onboarding script is a template for slack notification only.

## Environment Requirements

- PowerShell version `5.0` or higher
- Citrix Provider version `0.3.6` or higher

## Getting Started

1. Copy the `config-drift.ps1` script under `/config-drift-notifier` to the directory where the Citrix Terraform provider is located.
2. Test the script with the following command:
    ```powershell
    .\config-drift.ps1 -SlackWebhookUrl {SlackWebhookUrl} 
3. An optional parameter FilterList is provided to notify the user about only the resources in the list. An example can be  `-FilterList "citrix_machine_catalog.machine_catalog_0"` or just  `-FilterList "citrix_machine_catalog"`
