# Plugin for Terraform Provider for Citrix® Developer Guide

This documentation will guide you through the process of setting up your dev environment for running Plugin for Terraform Provider for Citrix® server locally on your dev machine.

## Table of Contents
- [Plugin for Terraform Provider for Citrix® Developer Guide](#plugin-for-terraform-provider-for-citrix-developer-guide)
  - [Table of Contents](#table-of-contents)
  - [Install Dependencies](#install-dependencies)
  - [Load project in VSCode for Go Development](#load-project-in-vscode-for-go-development)
  - [Debugging Provider code in VSCode](#debugging-provider-code-in-vscode)
    - [Add VSCode Launch Configuration](#add-vscode-launch-configuration)
    - [Start Debugger](#start-debugger)
    - [Attach Local Provider to PowerShell](#attach-local-provider-to-powershell)
  - [Debugging with citrix-daas-rest-go client code in Visual Studio Code](#debugging-with-citrix-daas-rest-go-client-code-in-visual-studio-code)
  - [Handling Terraform lists/sets and nested objects](#handling-terraform-listssets-and-nested-objects)
    - [Converting to Go native types](#converting-to-go-native-types)
    - [Preserving order in lists](#preserving-order-in-lists)
  - [Running the tests](#running-the-tests)
  - [Commonly faced errors](#commonly-faced-errors)
  - [Plugin for Terraform Provider for StoreFront Developer Guide](#plugin-for-terraform-provider-for-storefront-developer-guide)

## Install Dependencies
* Install Go on your local system: https://go.dev/doc/install
  * `choco install golang`
* Install latest version of Terraform (installing via Chocolatey recommended)
  * `choco install terraform`

## Load project in VSCode for Go Development
Visual Studio Code requires the `Go` extension to be able to load go projects, resolve internal references and even cross package references. Once the `Go` extension is installed, you should be able to load `terraform-provider-citrix` in VSCode. `Go` plugin requires the `go.mod` file to be in the root work directory when you load the project.

## Debugging Provider code in VSCode

### Add VSCode Launch Configuration

In order to debug the terraform provider module locally, you will need to setup a VS debugger config for the go project. Create `terraform-provider-citrix/.vscode/launch.json` and add the following configuration block:

```json
{
    "configurations": [
        {
            "name": "Debug Terraform Provider",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            // this assumes your workspace is the terraform-provider-citrix directory
            "program": "${workspaceFolder}",
            "env": {},
            "args": [
                "-debug",
            ]
        }
    ]
}
```

### Start Debugger

Once you have the debug config setup, you can hit `Run` on the VSCode debugger. You should see the following output on the debugger console:

```powershell
Starting: C:\Users\{local user}\go\bin\dlv.exe dap --listen=127.0.0.1:54615 from {Root of repo}\terraform-provider-citrix
DAP server listening at: 127.0.0.1:54615
Type 'dlv help' for list of commands.
{"@level":"debug","@message":"plugin address","@timestamp":"2023-05-22T15:21:11.753013-04:00","address":"127.0.0.1:54834","network":"tcp"}
Provider started. To attach Terraform CLI, set the TF_REATTACH_PROVIDERS environment variable with the following:

    Command Prompt:	set "TF_REATTACH_PROVIDERS={"registry.terraform.io/citrix/citrix":{"Protocol":"grpc","ProtocolVersion":6,"Pid":38724,"Test":true,"Addr":{"Network":"tcp","String":"127.0.0.1:54834"}}}"
    PowerShell:	$env:TF_REATTACH_PROVIDERS='{"registry.terraform.io/citrix/citrix":{"Protocol":"grpc","ProtocolVersion":6,"Pid":38724,"Test":true,"Addr":{"Network":"tcp","String":"127.0.0.1:54834"}}}'
```

### Attach Local Provider to PowerShell

Start a PowerShell session for running your terraform cli for debugging, and copy paste the following command:

    $env:TF_REATTACH_PROVIDERS='{"registry.terraform.io/citrix/citrix":{"Protocol":"grpc","ProtocolVersion":6,"Pid":38724,"Test":true,"Addr":{"Network":"tcp","String":"127.0.0.1:54834"}}}'

Now you are good to run terraform jobs to debug the provider code. Make sure to re-attach the provider server everytime you restart the debugger as the port can change per debugging session.

## Debugging with citrix-daas-rest-go client code in Visual Studio Code

Optionally, you can also debug [citrix-daas-rest-go](https://github.com/citrix/citrix-daas-rest-go) client, which is the Citrix DaaS Rest client for Go. By debugging with the go client, you can inspect the raw response and error message from Citrix DaaS APIs by setting up breakpoints in provider. You can also check the function implementation and models the provider server uses.

Clone the Go client from <https://github.com/citrix/citrix-daas-rest-go>. Go to `terraform-provider-citrix/go.mod` and uncomment the following line to intercept `citrix-daas-rest-go` client with local package:

    replace github.com/citrix/citrix-daas-rest-go => {your local repo directory}/citrix-daas-rest-go

Run [Debugging Provider code in Visual Studio Code](#debugging-provider-code-in-visual-studio-code) again and you will be able to step into the client functions.

Set a breakpoint in `terraform-provider-citrix/internal/provider/provider.go::Configure`

## Handling Terraform lists/sets and nested objects
### Converting to Go native types
When the Terraform configuration, state, or plan is being converted into a Go model we must use `types.List` and `types.Object` for lists and nested objects rather than go native slices and structs. This is in order to support Null/Unknown values. Unknown is especially important because any variables in the .tf configuration files can be unknown in `ValidateConfig` and `ModifyPlan`. However, handling these Terraform List and Object types is cumbersome as they are dynamically typed at runtime. See [this doc](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values) for more information. 

In order to reduce errors this project has introduced a system to convert between Terraform List/Object and Go native slices/structs. When data needs to be operated on it should be first converted to the Go native representation, then converted back to the Terraform representation. The following helper methods can handle this for you.

| From | To | Function | Notes |
|------|----|----------|-------|
| `types.Object` | `T` | `ObjectValueToTypedObject` | `T` must implement `ModelWithAttributes` |
| `T` | `types.Object` | `TypedObjectToObjectValue` | `T` must implement `ModelWithAttributes` |
| `types.List` | `T[]` | `ObjectListToTypedArray[T]` | `T` must implement `ModelWithAttributes`. For a list of nested objects |
| `T[]` | `types.List` | `TypedArrayToObjectList[T]` | `T` must implement `ModelWithAttributes`. For a list of nested objects |
| `types.List` | `string[]` | `StringListToStringArray` | For a list of strings |
| `string[]` | `types.List` | `StringArrayToStringList` | For a list of strings |
| `types.Set` | `string[]` | `StringSetToStringArray` | For a set of strings |
| `string[]` | `types.Set` | `StringArrayToStringSet` | For a set of strings |

In order to use the first 4 of these methods, the struct `T` needs to implement the [ModelWithAttributes](internal/util/types.go) interface which is ultimately populated from the attribute's Schema. This gives the Terraform type system the necessary information to populate a `types.Object` or `types.List` with a nested object.

### Preserving order in lists
Often time the order of elements in a list does not matter to the service. In this case one of the following helper functions should be used. These functions will get state list in sync with the remote list while preserving the order in the state when possible. 

| Function | Input | Notes |
|----------|-------|-------|
| `RefreshList` | `[]string` | |
| `RefreshUsersList` | `types.Set` | Will ensure users are not duplicated by UPN or SAMname |
| `RefreshListValues` | `types.List` of `string` | |
| `RefreshListValueProperties` | `types.List` of `types.Object` | Each element will have its `RefreshListItem` method called. The element's type must implement the `RefreshableListItemWithAttributes` interface |

## Running the tests

Before running the tests, you need to provide values for environment variables required by the test files. 
The environment parameters that need to be specified can be found in the following template files:
1. To Run Tests for the Cloud Environment: `settings.cloud.example.json`
2. To Run Tests for the On-Premise environment: `setings.onprem.example.json`

Copy the environment parameters from the appropriate template file and paste them in the GO `settings.json` file.
To navigate to `settings.json` file, follow the steps below:
1. Click on the `Extensions` icon on the left panel of VS Code.
2. Search for the `Go` extension and click on the `gear` icon next to it. 

    ![Go Extension in VS Code](./images/go-extension.png "Go Extension in VS Code")

3. In the search bar, type in `go.testEnvVars`. From the search result, click on `Edit in settings.json` under `Go: Test Env Vars`.
4. Paste the contents of the template file that you copied earlier.
5. Update the missing values in the file and run the commands mentioned below

```powershell
➥ cd {Root of repo}/terraform-provider-citrix
➥ $env:TF_ACC = 1
➥ go test -count=1 -run='{name of resource to test}' -v ./internal/test
```

## Commonly faced errors
```powershell
    error obtaining VCS status: exit status 128
        Use -buildvcs=false to disable VCS stamping.
    exit status 1
```

To solve this issue, run the following command at the root of the repository:
```powershell
    git config --global --add safe.directory [path to dir/repo]
```

## Plugin for Terraform Provider for StoreFront Developer Guide

The test running process is the same as  [Running the tests](#running-the-tests) with additional parameter in the settings.cloud.example.json or settings.onprem.example.json  `StoreFront env variable` section