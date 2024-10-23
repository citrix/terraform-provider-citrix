// Copyright © 2024. Citrix Systems, Inc.

package validators

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var (
	_ validator.String = AlsoRequiresOneOfOnValuesValidator{}
	_ validator.Bool   = AlsoRequiresOneOfOnValuesValidator{}
)

// AlsoRequiresOneOfOnValuesValidator is the underlying struct implementing AlsoRequiresOnValue.
type AlsoRequiresOneOfOnValuesValidator struct {
	OnStringValues  []string
	OnBoolValues    []bool
	PathExpressions path.Expressions
}

type AlsoRequiresOneOfOnValuesValidatorRequest struct {
	Config         tfsdk.Config
	ConfigValue    attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type AlsoRequiresOneOfOnValuesValidatorResponse struct {
	Diagnostics diag.Diagnostics
}

// Description implements validator.String.
func (v AlsoRequiresOneOfOnValuesValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

// MarkdownDescription implements validator.String.
func (v AlsoRequiresOneOfOnValuesValidator) MarkdownDescription(context.Context) string {
	if len(v.OnStringValues) > 0 {
		return fmt.Sprintf("If the current attribute is set to one of [%s], exactly one of the following also need to be set: %q", strings.Join(v.OnStringValues, ","), v.PathExpressions)
	} else if len(v.OnBoolValues) > 0 {
		boolValueArray := []string{}
		for _, boolValue := range v.OnBoolValues {
			boolValueArray = append(boolValueArray, strconv.FormatBool(boolValue))
		}
		return fmt.Sprintf("If the current attribute is set to one of [%v], exactly one of the following also need to be set: %q", strings.Join(boolValueArray, ","), v.PathExpressions)
	}
	return ""
}

func (v AlsoRequiresOneOfOnValuesValidator) Validate(ctx context.Context, req AlsoRequiresOneOfOnValuesValidatorRequest, res *AlsoRequiresOneOfOnValuesValidatorResponse) {
	// If attribute configuration is null, there is nothing else to validate
	if req.ConfigValue.IsNull() {
		return
	}

	expressions := req.PathExpression.MergeExpressions(v.PathExpressions...)

	matchedCount := 0
	for _, expression := range expressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)

		res.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
			// If the user specifies the same attribute this validator is applied to,
			// also as part of the input, skip it
			if mp.Equal(req.Path) {
				continue
			}

			var mpVal attr.Value
			diags := req.Config.GetAttribute(ctx, mp, &mpVal)
			res.Diagnostics.Append(diags...)

			// Collect all errors
			if diags.HasError() {
				continue
			}

			// Delay validation until all involved attribute have a known value
			if mpVal.IsUnknown() {
				return
			}

			if mpVal.IsNull() {
				continue
			}

			matchedCount++
		}
	}
	if matchedCount == 0 {
		res.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			req.Path,
			fmt.Sprintf("One of attributes %q must be specified", expressions),
		))
	} else if matchedCount > 1 {
		res.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			req.Path,
			fmt.Sprintf("Only one of attributes %q can be specified", expressions),
		))
	}
}

// ValidateString implements validator.String.
func (v AlsoRequiresOneOfOnValuesValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If attribute configuration is null, there is nothing else to validate
	validateReq := AlsoRequiresOneOfOnValuesValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AlsoRequiresOneOfOnValuesValidatorResponse{}

	for _, value := range v.OnStringValues {
		if value == req.ConfigValue.ValueString() {
			v.Validate(ctx, validateReq, validateResp)
			return
		}
	}

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

// ValidateBool implements validator.Bool.
func (v AlsoRequiresOneOfOnValuesValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// If attribute configuration is null, there is nothing else to validate
	validateReq := AlsoRequiresOneOfOnValuesValidatorRequest{
		Config:         req.Config,
		ConfigValue:    req.ConfigValue,
		Path:           req.Path,
		PathExpression: req.PathExpression,
	}
	validateResp := &AlsoRequiresOneOfOnValuesValidatorResponse{}

	for _, value := range v.OnBoolValues {
		if value == req.ConfigValue.ValueBool() {
			v.Validate(ctx, validateReq, validateResp)
			return
		}
	}

	resp.Diagnostics.Append(validateResp.Diagnostics...)
}

// AlsoRequiresOneOfOnValues checks that a set of path.Expression has a non-null value,
// if the current attribute or block is set to one of the values defined in onValues array.
//
// Relative path.Expression will be resolved using the attribute or block
// being validated.
func AlsoRequiresOneOfOnStringValues(onValues []string, expressions ...path.Expression) validator.String {
	return AlsoRequiresOneOfOnValuesValidator{
		OnStringValues:  onValues,
		PathExpressions: expressions,
	}
}

func AlsoRequiresOneOfOnBoolValues(onValues []bool, expressions ...path.Expression) validator.Bool {
	return AlsoRequiresOneOfOnValuesValidator{
		OnBoolValues:    onValues,
		PathExpressions: expressions,
	}
}
