// Copyright © 2024. Citrix Systems, Inc.

package validators

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.String = ConflictsWithOnValuesValidator{}
	_ validator.Bool   = ConflictsWithOnValuesValidator{}
)

// ConflictsWithOnValuesValidator is the underlying struct implementing AlsoRequiresOnValue.
type ConflictsWithOnValuesValidator struct {
	OnStringValues  []string
	OnBoolValues    []bool
	PathExpressions path.Expressions
}

// Description implements validator.String.
func (v ConflictsWithOnValuesValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

// MarkdownDescription implements validator.String.
func (v ConflictsWithOnValuesValidator) MarkdownDescription(context.Context) string {
	if len(v.OnStringValues) > 0 {
		return fmt.Sprintf("If the current attribute is set to one of [%s], none of the following can be set: %q", strings.Join(v.OnStringValues, ","), v.PathExpressions)
	} else if len(v.OnBoolValues) > 0 {
		boolValueArray := []string{}
		for _, boolValue := range v.OnBoolValues {
			boolValueArray = append(boolValueArray, strconv.FormatBool(boolValue))
		}
		return fmt.Sprintf("If the current attribute is set to one of [%v], none of the following can be set: %q", strings.Join(boolValueArray, ","), v.PathExpressions)
	}
	return ""
}

// ValidateString implements validator.String.
func (v ConflictsWithOnValuesValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If attribute configuration is null, there is nothing else to validate
	if req.ConfigValue.IsNull() {
		return
	}

	for _, value := range v.OnStringValues {
		if value == req.ConfigValue.ValueString() {
			stringvalidator.ConflictsWith(v.PathExpressions...).ValidateString(ctx, req, resp)
			return
		}
	}
}

// ValidateBool implements validator.Bool.
func (v ConflictsWithOnValuesValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// If attribute configuration is null, there is nothing else to validate
	if req.ConfigValue.IsNull() {
		return
	}

	for _, value := range v.OnBoolValues {
		if value == req.ConfigValue.ValueBool() {
			boolvalidator.ConflictsWith(v.PathExpressions...).ValidateBool(ctx, req, resp)
			return
		}
	}
}

// ConflictsWithOnValues checks that a set of path.Expression has a non-null value,
// if the current attribute or block is set to one of the values defined in onValues array.
//
// Relative path.Expression will be resolved using the attribute or block
// being validated.
func ConflictsWithOnStringValues(onValues []string, expressions ...path.Expression) validator.String {
	return ConflictsWithOnValuesValidator{
		OnStringValues:  onValues,
		PathExpressions: expressions,
	}
}

func ConflictsWithOnBoolValues(onValues []bool, expressions ...path.Expression) validator.Bool {
	return ConflictsWithOnValuesValidator{
		OnBoolValues:    onValues,
		PathExpressions: expressions,
	}
}
