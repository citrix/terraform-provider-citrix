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
  - [Running the tests](#running-the-tests)
  - [Commonly faced errors](#commonly-faced-errors)

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

Optionally, you can also debug `citrix-daas-rest-go` client, which is the Citrix DaaS Rest client for Go. By debugging with the go client, you can inspect the raw response and error message from Citrix DaaS APIs by setting up breakpoints in provider. You can also check the function implementation and models the provider server uses.

Clone Go client from <https://github.com/citrix/citrix-daas-rest-go>. Go to `terraform-provider-citrix/go.mod` and uncomment the following line to intercept `citrix-daas-rest-go` client with local package:

    replace github.com/citrix/citrix-daas-rest-go => {your local repo directory}/citrix-daas-rest-go

Run [Debugging Provider code in Visual Studio Code](#debugging-provider-code-in-visual-studio-code) again and you will be able to step into the client functions.

Set a breakpoint in `terraform-provider-citrix/internal/provider/provider.go::Configure`

## Running the tests
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