// Copyright © 2026. Citrix Systems, Inc.

package testresource

import (
	"context"

	"github.com/citrix/terraform-provider-citrix/custom-linters/test/testutil"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &TestResource{}
	_ resource.ResourceWithConfigure      = &TestResource{}
	_ resource.ResourceWithImportState    = &TestResource{}
	_ resource.ResourceWithModifyPlan     = &TestResource{}
	_ resource.ResourceWithValidateConfig = &TestResource{}
)

type TestResource struct{}

// Valid: Basic panic handler as first statement
func (r *TestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// Valid: Single comment before defer
func (r *TestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Comment
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// Valid: Multiple comments before defer
func (r *TestResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// First comment
	// Second comment
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// Valid: Block comment before defer
func (r *TestResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	/*
		Block comment
		spanning multiple lines
	*/
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// Valid: Multi-line function signature
func (r *TestResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// Valid: ValidateConfig with panic handler
func (r *TestResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// Valid: Schema is excluded
func (r *TestResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
}

// Valid: Metadata is excluded
func (r *TestResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
}

// Valid: Configure is excluded
func (r *TestResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
}

// Valid: ImportState is excluded
func (r *TestResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}

var (
	_ resource.Resource                   = &EdgeCaseResource{}
	_ resource.ResourceWithModifyPlan     = &EdgeCaseResource{}
	_ resource.ResourceWithValidateConfig = &EdgeCaseResource{}
)

type EdgeCaseResource struct{}

// violations:1
// Invalid: Missing panic handler completely
func (r *EdgeCaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

// violations:1
// Invalid: Wrong defer function
func (r *EdgeCaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer wrongFunction()
}

// violations:1
// Invalid: Other defer before panic handler
func (r *EdgeCaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer cleanup()
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// Valid: Linter doesn't check arguments, only that PanicHandler is called
func (r *EdgeCaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer testutil.PanicHandler(nil)
}

// violations:1
// Invalid: Variable declaration before defer (defer must be first statement)
func (r *EdgeCaseResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	wrongVar := &resp.Diagnostics
	defer testutil.PanicHandler(wrongVar)
}

// violations:1
// Invalid: Variable declaration before defer
func (r *EdgeCaseResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	x := 1
	defer testutil.PanicHandler(&resp.Diagnostics)
	_ = x
}

func (r *EdgeCaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
}

func (r *EdgeCaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
}

var _ datasource.DataSource = &TestDataSource{}

type TestDataSource struct{}

// Valid: DataSource with panic handler
func (d *TestDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// Valid: DataSource Schema is excluded
func (d *TestDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
}

// Valid: DataSource Metadata is excluded
func (d *TestDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
}

var _ datasource.DataSource = &BadDataSource{}

type BadDataSource struct{}

// violations:1
// Invalid: DataSource missing panic handler
func (d *BadDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
}

func (d *BadDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
}

func (d *BadDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
}

var _ resource.Resource = &CodeBeforeDeferResource{}

type CodeBeforeDeferResource struct{}

// violations:1
// Invalid: Assignment before defer
func (r *CodeBeforeDeferResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var x = 1
	defer testutil.PanicHandler(&resp.Diagnostics)
	_ = x
}

// violations:1
// Invalid: Function call before defer
func (r *CodeBeforeDeferResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	helper()
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// violations:1
// Invalid: If statement before defer
func (r *CodeBeforeDeferResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if true { //nolint:staticcheck // Empty branch intentionally used to test code before defer
		// something
	}
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// violations:1
// Invalid: For loop before defer
func (r *CodeBeforeDeferResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	for i := range 10 {
		_ = i
	}
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *CodeBeforeDeferResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
}

func (r *CodeBeforeDeferResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
}

func wrongFunction() {}
func cleanup()       {}
func helper()        {}

// Helper methods and model types that should NOT be flagged by the linter

type DeliveryGroupAccessPolicyModel struct{}

// Valid: Helper method with non-standard signature should be ignored
// This is NOT a Terraform SDK method because it doesn't have the Request/Response signature
func (accessPolicy DeliveryGroupAccessPolicyModel) ValidateConfig(ctx context.Context, diagnostics interface{}, index int) bool {
	// No panic handler needed - this is a helper method, not a Terraform SDK interface method
	return true
}

type DeliveryGroupAppProtection struct{}

// Valid: Helper method with custom signature should be ignored
func (appProtection DeliveryGroupAppProtection) ValidateConfig(ctx context.Context, diagnostics interface{}) bool {
	// No panic handler needed - this is a helper method
	return true
}

type ImageUpdateRebootOptionsModel struct{}

// Valid: Helper method with single parameter should be ignored
func (rebootOptions ImageUpdateRebootOptionsModel) ValidateConfig(diagnostics interface{}) {
	// No panic handler needed - this is a helper method
}

type NameValueStringPairModel struct{}

// Valid: Helper method with custom signature should be ignored
func (r NameValueStringPairModel) ValidateConfig(ctx context.Context, diagnostics interface{}, index int) bool {
	// No panic handler needed - this is a helper method
	return true
}

// Valid: Method with standard name but wrong signature (2 params instead of 3)
type TwoParamResource struct{}

func (r *TwoParamResource) Create(ctx context.Context, req resource.CreateRequest) {
	// No panic handler needed - not a Terraform SDK method signature
}

// Valid: Method with standard name but wrong signature (4 params)
type FourParamResource struct{}

func (r *FourParamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse, extra int) {
	// No panic handler needed - not a Terraform SDK method signature
}

// Valid: Method with standard name but non-Request second parameter
type WrongSecondParamResource struct{}

func (r *WrongSecondParamResource) Update(ctx context.Context, config string, resp *resource.UpdateResponse) {
	// No panic handler needed - second param is not a Request type
}

// Valid: Method with standard name but non-Response third parameter
type WrongThirdParamResource struct{}

func (r *WrongThirdParamResource) Delete(ctx context.Context, req resource.DeleteRequest, msg string) {
	// No panic handler needed - third param is not a Response type
}

// Valid: Standalone function (not a method) with Terraform method name
func Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// No panic handler needed - this is not a method on a receiver
}

// Valid: Standalone function Read
func Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// No panic handler needed - standalone functions are ignored
}
