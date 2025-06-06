---
name: Bug report
about: Report any issues using the Citrix Terraform Provider
title: "[Bug]"
labels: bug
assignees: ''

---

<!-- Thanks for taking the time to fill out this bug report! Before submitting this issue please check the [open bugs](https://github.com/citrix/terraform-provider-citrix/issues?q=is%3Aissue+is%3Aopen+label%3Abug) to ensure the bug has not already been reported. If it has been reported give it a 👍 -->


## Describe the bug

<!-- Summary of the issue -->

**Terraform command (import, apply, etc):**
**Resource impacted:**
<!-- If this bug is present when using the Citrix service UI or REST APIs then it is not a bug in the provider but rather a bug in the underlying service or the environment. In some cases there can be an enhancement in the provider to handle the error better. Please open a feature request instead of a bug in this case. For more information see [CONTRIBUTING.md#provider-issue-vs-product-issue-vs-configuration-issue](https://github.com/citrix/terraform-provider-citrix/blob/main/CONTRIBUTING.md#provider-issue-vs-product-issue-vs-configuration-issue). -->
**Issue reproducible outside of Terraform:** <!-- Yes/No/Not verified -->


## Versions

<!-- Use the `terraform -v` command to find the Terraform and Citrix Provider versions. -->
**Terraform:** 
**citrix/citrix provider:** 
**Operation system:** 

**Environment type:** <!-- Cloud or On-premises -->
**Hypervisor type (if applicable):** <!-- Azure, AWS, GCP, vSphere, XenServer, Nutanix, etc. -->

<!-- For on-premises customers fill out any that apply with the CU or LTSR version (eg 2402). -->
**CVAD (DDC, VDA, etc):** 
**Storefront:** 

## Terraform configuration files
<!-- Paste or attach any relevant `.tf` files with secrets and identifying information removed. -->


## Terraform console output
<!-- Paste the output from Terraform CLI including any errors and the transactionIds if present. Errors with TransactionIDs are critical for troubleshooting.

If the output references a file in the temp directory include it as well. -->


## Terraform log file
<!-- If the issue is reproducible enable Terraform debug logging using one of the commands below. Then reproduce the issue and include the resulting log file. More information about Terraform logging is available [here](https://developer.hashicorp.com/terraform/plugin/log/managing#enable-logging). -->

<!-- cmd: -->
```bat
set TF_LOG="DEBUG"
set TF_LOG_PATH="./citrix-provider-issue.txt"
terraform <command>
```

<!-- Powershell: -->
```powershell
$env:TF_LOG="DEBUG"
$env:TF_LOG_PATH="./citrix-provider-issue.txt"
terraform <command>
```

<!-- bash: -->
```bash
export TF_LOG="DEBUG"
export TF_LOG_PATH="./citrix-provider-issue.txt"
terraform <command>
```