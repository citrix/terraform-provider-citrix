// Copyright © 2026. Citrix Systems, Inc.

package testresource

import (
	"context"

	"github.com/citrix/terraform-provider-citrix/custom-linters/test/testutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Test resources for unknowncheck linter
var (
	_ resource.Resource                   = &GoodValidateConfig{}
	_ resource.ResourceWithValidateConfig = &GoodValidateConfig{}
	_ resource.Resource                   = &GoodModifyPlan{}
	_ resource.ResourceWithModifyPlan     = &GoodModifyPlan{}
	_ resource.Resource                   = &BadValidateConfig{}
	_ resource.ResourceWithValidateConfig = &BadValidateConfig{}
	_ resource.Resource                   = &BadModifyPlan{}
	_ resource.ResourceWithModifyPlan     = &BadModifyPlan{}
	_ resource.Resource                   = &ComplexCases{}
	_ resource.ResourceWithValidateConfig = &ComplexCases{}
	_ resource.Resource                   = &RawFieldCases{}
	_ resource.ResourceWithModifyPlan     = &RawFieldCases{}
	_ resource.Resource                   = &NegatedIsNullCases{}
	_ resource.ResourceWithValidateConfig = &NegatedIsNullCases{}
	_ resource.Resource                   = &AndConditionCases{}
	_ resource.ResourceWithValidateConfig = &AndConditionCases{}
)

type GoodValidateConfig struct{}

// CRUD methods below are stubs to satisfy resource.Resource interface requirements

func (r *GoodValidateConfig) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *GoodValidateConfig) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *GoodValidateConfig) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *GoodValidateConfig) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// Valid: IsUnknown checked before IsNull
func (r *GoodValidateConfig) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
	var data struct {
		Name types.String `tfsdk:"name"`
	}
	req.Config.Get(ctx, &data)

	if !data.Name.IsUnknown() && data.Name.IsNull() {
		resp.Diagnostics.AddError("Name is null", "Name cannot be null")
	}
}

// Valid: Only IsUnknown without IsNull is fine
func (r *GoodValidateConfig) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
}

func (r *GoodValidateConfig) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
}

type GoodModifyPlan struct{}

// CRUD methods below are stubs to satisfy resource.Resource interface requirements

func (r *GoodModifyPlan) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *GoodModifyPlan) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *GoodModifyPlan) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *GoodModifyPlan) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// Valid: IsUnknown checked in left side of AND
func (r *GoodModifyPlan) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
	var data struct {
		Id types.String `tfsdk:"id"`
	}
	req.Plan.Get(ctx, &data)

	if !data.Id.IsUnknown() && data.Id.IsNull() {
		resp.Diagnostics.AddError("Id is null", "Id cannot be null")
	}
}

// Valid: Multiple fields with proper checks
func (r *GoodModifyPlan) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
}

func (r *GoodModifyPlan) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
}

type BadValidateConfig struct{}

// CRUD methods below are stubs to satisfy resource.Resource interface requirements

func (r *BadValidateConfig) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *BadValidateConfig) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *BadValidateConfig) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *BadValidateConfig) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// violations:2
// Invalid: IsNull without IsUnknown check (positive checks only)
func (r *BadValidateConfig) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
	var data struct {
		Name        types.String `tfsdk:"name"`
		Description types.String `tfsdk:"description"`
		Value       types.Int64  `tfsdk:"value"`
	}
	req.Config.Get(ctx, &data)

	// Invalid: checking if field.IsNull() is true without checking IsUnknown first
	if data.Name.IsNull() {
		resp.Diagnostics.AddError("Name is null", "Name cannot be null")
	}

	// Valid: !field.IsNull() is checking if field has a value, no need for IsUnknown check
	if !data.Description.IsNull() { //nolint:staticcheck // Test case for checking IsNull without IsUnknown
		// Do something
	}

	// Invalid: checking if field.IsNull() is true without checking IsUnknown first
	if data.Value.IsNull() {
		resp.Diagnostics.AddError("Value is null", "Value cannot be null")
	}
}

func (r *BadValidateConfig) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
}

func (r *BadValidateConfig) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
}

type BadModifyPlan struct{}

// CRUD methods below are stubs to satisfy resource.Resource interface requirements

func (r *BadModifyPlan) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *BadModifyPlan) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *BadModifyPlan) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *BadModifyPlan) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// violations:1
// Invalid: IsNull without IsUnknown check in ModifyPlan (positive check only)
func (r *BadModifyPlan) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
	var plan struct {
		Id   types.String `tfsdk:"id"`
		Tags types.List   `tfsdk:"tags"`
	}
	req.Plan.Get(ctx, &plan)

	// Invalid: checking if field.IsNull() is true without checking IsUnknown first
	if plan.Id.IsNull() {
		resp.Diagnostics.AddError("Id is null", "Id cannot be null")
	}

	// Valid: !field.IsNull() is checking if field has a value, no need for IsUnknown check
	if !plan.Tags.IsNull() { //nolint:staticcheck // Test case for checking IsNull without IsUnknown
		// Do something
	}
}

func (r *BadModifyPlan) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
}

func (r *BadModifyPlan) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
}

type ComplexCases struct{}

// CRUD methods below are stubs to satisfy resource.Resource interface requirements

func (r *ComplexCases) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *ComplexCases) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *ComplexCases) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *ComplexCases) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// violations:1
// Mixed cases
func (r *ComplexCases) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
	var data struct {
		Name   types.String `tfsdk:"name"`
		Config types.Object `tfsdk:"config"`
	}
	req.Config.Get(ctx, &data)

	// Valid: !field.IsNull() checks if field has a value, no need for IsUnknown check
	if !data.Name.IsUnknown() && !data.Name.IsNull() { //nolint:staticcheck // Test case for valid IsUnknown before IsNull pattern
		// Do something
	}

	// Invalid: field.IsNull() without checking IsUnknown first
	if data.Config.IsNull() || someOtherCondition() {
		resp.Diagnostics.AddError("Config issue", "Config has an issue")
	}
}

func (r *ComplexCases) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
}

func (r *ComplexCases) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
}

func someOtherCondition() bool {
	return false
}

type RawFieldCases struct{}

// CRUD methods below are stubs to satisfy resource.Resource interface requirements

func (r *RawFieldCases) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *RawFieldCases) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *RawFieldCases) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *RawFieldCases) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// violations:0
// Valid: req.Plan.Raw and req.State.Raw do not need IsUnknown checks
func (r *RawFieldCases) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)

	// Valid: Raw fields are special and don't need IsUnknown checks
	if req.Plan.Raw.IsNull() {
		return
	}

	if !req.State.Raw.IsNull() { //nolint:staticcheck // Test case - empty branch is intentional
		// Do something with state
	}

	if req.Config.Raw.IsNull() {
		return
	}

	var plan struct {
		Id types.String `tfsdk:"id"`
	}
	req.Plan.Get(ctx, &plan)
}

func (r *RawFieldCases) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
}

func (r *RawFieldCases) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
}

type NegatedIsNullCases struct{}

// CRUD methods below are stubs to satisfy resource.Resource interface requirements

func (r *NegatedIsNullCases) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *NegatedIsNullCases) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *NegatedIsNullCases) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *NegatedIsNullCases) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// violations:0
// Valid: !field.IsNull() doesn't need IsUnknown check
func (r *NegatedIsNullCases) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
	var data struct {
		Name   types.String `tfsdk:"name"`
		Value  types.String `tfsdk:"value"`
		Config types.Object `tfsdk:"config"`
	}
	req.Config.Get(ctx, &data)

	// Valid: !field.IsNull() checks if field has a value
	if !data.Name.IsNull() { //nolint:staticcheck // Test case
		// Do something when name has a value
	}

	// Valid: !field.IsNull() in complex condition
	if someOtherCondition() && !data.Value.IsNull() { //nolint:staticcheck // Test case
		// Do something
	}

	// Valid: Multiple !field.IsNull() checks
	if !data.Config.IsNull() && !data.Name.IsNull() { //nolint:staticcheck // Test case
		// Do something
	}
}

func (r *NegatedIsNullCases) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
}

func (r *NegatedIsNullCases) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
}

type AndConditionCases struct{}

// CRUD methods below are stubs to satisfy resource.Resource interface requirements

func (r *AndConditionCases) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *AndConditionCases) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *AndConditionCases) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

func (r *AndConditionCases) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
}

// violations:0
// Valid: IsNull and IsUnknown checked in AND condition (order doesn't matter in AND)
func (r *AndConditionCases) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	defer testutil.PanicHandler(&resp.Diagnostics)
	var data struct {
		Name   types.String `tfsdk:"name"`
		Value  types.String `tfsdk:"value"`
		Config types.Object `tfsdk:"config"`
	}
	req.Config.Get(ctx, &data)

	// Valid: Both checked in AND (IsNull before IsUnknown is OK in AND)
	if !data.Name.IsNull() && !data.Name.IsUnknown() { //nolint:staticcheck // Test case - empty branch is intentional
		// Do something
	}

	// Valid: Both checked in AND (IsUnknown before IsNull is also OK)
	if !data.Value.IsUnknown() && !data.Value.IsNull() { //nolint:staticcheck // Test case
		// Do something
	}

	// Valid: Even with negations reversed
	if data.Config.IsNull() && data.Config.IsUnknown() { //nolint:staticcheck // Test case
		// Do something
	}
}

func (r *AndConditionCases) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
}

func (r *AndConditionCases) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
}
