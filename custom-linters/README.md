# Custom Linters

Custom golangci-lint linters for the Citrix Terraform Provider. These linters enforce provider-specific patterns and best practices to improve code quality and prevent common issues.

## Overview

This directory contains custom static analysis tools that extend golangci-lint to catch provider-specific issues at build time. Each linter is implemented as a separate Go analysis package and is integrated into our CI/CD pipeline. The linters are largely AI generated backed by comprehensive tests to ensure the validity of their implementation.

## Available Linters

### ExecuteWithRetry

**Purpose:** Enforces that API GET operations (read operations) use the `ExecuteWithRetry` wrapper instead of calling `Execute()` directly. This ensures proper resilience for operations that retrieve data from the API.

**Why?** The `ExecuteWithRetry` wrapper provides automatic retry logic with exponential backoff for transient failures (429 rate limits, 5xx server errors). This is especially important for GET operations since they're idempotent and safe to retry. This improves provider resilience and user experience by handling temporary API issues gracefully without manual intervention.

**How it Works:**
- Recursively analyzes the entire expression tree leading to `.Execute()` calls
- Identifies GET operations by checking for request type patterns containing `ApiGet*`, `ApiFetch*`, `ApiList*`, etc.
- Works with any chaining pattern: direct calls, `AddRequestData()`, `Async()`, or future wrapper methods
- StoreFront APIs are excluded (they use direct `Execute()` - ExecuteWithRetry is not compatible with StoreFront's PowerShell remoting architecture)
- Only GET operations are required to use `ExecuteWithRetry`
- Non-GET operations (Create, Update, Delete, Patch, Set, etc.) can use direct `Execute()` calls

**Common Fixes Required:**
When converting `.Execute()` to `ExecuteWithRetry`, you may need to:
1. **Add missing import**: `citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"`
2. **Fix return type**: Use the correct response model type (check API documentation or test files)
   - Example: `ZoneResponseModel` → `ZoneDetailResponseModel`
   - Example: `ServiceAccountsResponseModelCollection` → `ServiceAccountResponseModelCollection`
   - Example: `PvsSiteResponseModelCollection` → `PvsStreamingSiteResponseModelCollection`
3. **Use pointer type**: Type parameter must be a pointer: `ExecuteWithRetry[*Type]` not `ExecuteWithRetry[Type]`

**Detected Patterns:**
```go
// ✗ All of these will be flagged for GET operations:
request.Execute()                                    // Direct call
AddRequestData(request, client).Execute()           // Wrapped call
AddRequestData(request, client).Async(true).Execute() // Chained call
futureWrapper(request).someMethod().Execute()       // Any chaining pattern
```

**Required Pattern for GET Operations:**
```go
result, httpResp, err := citrixdaasclient.ExecuteWithRetry[ResponseType](request, client)
```

**Code Examples:**

```go
// ✓ Correct - Using ExecuteWithRetry for GET
func getHypervisor(ctx context.Context, client *citrixdaasclient.CitrixDaasClient) {
    getRequest := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsGetHypervisor(ctx, hypervisorId)
    hypervisor, _, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.HypervisorDetailResponseModel](getRequest, client)
}

// ✓ Correct - Direct Execute() for POST/PUT/DELETE (any chaining pattern allowed)
func createHypervisor(ctx context.Context, client *citrixdaasclient.CitrixDaasClient) {
    createRequest := client.ApiClient.HypervisorsAPIsDAAS.HypervisorsCreateHypervisor(ctx)
    _, err := citrixdaasclient.AddRequestData(createRequest, client).Execute()
}
```

### PanicHandler

**Purpose:** Enforces that all Terraform provider SDK interface functions start with `defer util.PanicHandler(&resp.Diagnostics)`.

**Target Functions:** Create, Read, Update, Delete, ModifyPlan, ValidateConfig  
**Excluded Functions:** Schema, ImportState, Metadata, Configure

**Why?** Prevents provider crashes by catching unexpected panics and converting them into proper Terraform diagnostics. Users see actionable error messages instead of stack traces.

**Code Example:**

```go
// ✓ Correct
func (r *myResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    defer util.PanicHandler(&resp.Diagnostics)
    // ...
}
```

### UnknownCheck

**Purpose:** Enforces that `IsUnknown()` is checked before `IsNull()` in `ValidateConfig` and `ModifyPlan` functions.

**Target Functions:** ValidateConfig, ModifyPlan

**Why?** When users reference Terraform variables, those values are initially unknown during validation and plan phases. Checking `IsNull()` without first checking `IsUnknown()` can lead to incorrect evaluation and should be treated with caution.

**Example Scenario:**
```hcl
resource "citrix_delivery_group" "example" {
  name = var.delivery_group_name  # Unknown during validation
}
```

**Code Example:**

```go
// ✓ Correct - Check IsUnknown first
func (r *myResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
    var data MyResourceModel
    req.Config.Get(ctx, &data)
    
    if !data.Name.IsUnknown() && data.Name.IsNull() {
        resp.Diagnostics.AddError("Name is null", "Name cannot be null")
    }
}
```


## Scripts
To supplement the golangci-lint custom linters, it is also useful to have scripts to perform analysis.

### validate-docs

**Purpose:** Validates that generated Terraform documentation has proper categorization metadata (`subcategory` field) by checking that schema descriptions contain the `" --- "` separator pattern.

**Why?** The `terraform-plugin-docs` generator parses schema descriptions in the format `"CATEGORY --- Description"` to extract `subcategory` frontmatter. Without this, resources aren't properly categorized in the documentation.

**Usage:** Runs automatically during `make generate`. Fix violations by adding the separator to schema descriptions:
```go
Description: "Site Configuration --- Manages a backup schedule"
```
## Development

### Architecture

```
custom-linters/
├── <linter-name>/
│   ├── <linter-name>.go              # Analyzer implementation
│   └── <linter-name>_test.go         # Unit tests
├── test/
│   ├── testresource/                 # Test resources
│   │   ├── <linter-name>_test_cases.go
│   └── testutil/                     # Test utilities
│       ├── testutil.go               # Shared test runner
├── util/
│   └── terraform.go                  # Shared Terraform utilities
├── plugin.go                         # Plugin registration
└── README.md
```

### Adding a New Linter

1. Create package directory: `mkdir custom-linters/mylinter`

2. Implement analyzer in `mylinter/mylinter.go` using `golang.org/x/tools/go/analysis`

3. Add test in `mylinter/mylinter_test.go`:
   ```go
   func TestAllCases(t *testing.T) {
       testCases := testutil.ExtractTestCasesFromComments(t, "package/path", "mylinter_test_cases.go")
       testutil.RunAnalyzerTestCases(t, mylinter.Analyzer, "package/path", "mylinter_test_cases.go", testCases)
   }
   ```

4. Create test resources in `test/testresource/mylinter_test_cases.go` with annotations:
   ```go
   // Valid: correct pattern (defaults to 0 violations)
   func (r *GoodResource) Create(...) {
       defer testutil.PanicHandler(&resp.Diagnostics)
   }
   
   // violations:1
   // Invalid: missing panic handler
   func (r *BadResource) Create(...) {
       // no defer statement
   }
   ```
   Use `violations:N` where N > 0. Functions without annotations default to 0 violations.

5. Register plugin in `plugin.go` (see existing linters for pattern)

6. Enable in `.golangci.yml` under `linters.enable` and `linters-settings.custom`

7. Rebuild: `make custom-gcl && go test ./custom-linters/...`

8. Update this README with linter docs

### Shared Utilities

- **`testutil.ExtractTestCasesFromComments()`** - Parses `violations:N` annotations from test files
- **`testutil.RunAnalyzerTestCases()`** - Test runner validating violations per function
- **`testutil.PanicHandler()`** - Stub implementation for test resources
- **`util.IsTerraformMethod()`** - Check if a function is a Terraform SDK method

## Testing

```bash
# Test all linters
cd custom-linters && go test ./...

# Test specific linter
cd custom-linters/panichandler && go test -v
cd custom-linters/unknowncheck && go test -v

# Run linters on codebase
make custom-gcl
./custom-linters/bin/custom-gcl run ./...
```

## Usage

The linters run automatically as part of `make lint`. To rebuild after making changes:

```bash
make custom-gcl
```

To run only specific linters:

```bash
./custom-linters/bin/custom-gcl run --enable-only panichandler,unknowncheck ./...
```
