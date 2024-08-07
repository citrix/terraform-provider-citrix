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
	_ validator.String = AlsoRequiresOnValuesValidator{}
	_ validator.Bool   = AlsoRequiresOnValuesValidator{}
)

// AlsoRequiresOnValuesValidator is the underlying struct implementing AlsoRequiresOnValue.
type AlsoRequiresOnValuesValidator struct {
	OnStringValues  []string
	OnBoolValues    []bool
	PathExpressions path.Expressions
}

// Description implements validator.String.
func (v AlsoRequiresOnValuesValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

// MarkdownDescription implements validator.String.
func (v AlsoRequiresOnValuesValidator) MarkdownDescription(context.Context) string {
	if len(v.OnStringValues) > 0 {
		return fmt.Sprintf("If the current attribute is set to one of [%s], the following also need to be set: %q", strings.Join(v.OnStringValues, ","), v.PathExpressions)
	} else if len(v.OnBoolValues) > 0 {
		boolValueArray := []string{}
		for _, boolValue := range v.OnBoolValues {
			boolValueArray = append(boolValueArray, strconv.FormatBool(boolValue))
		}
		return fmt.Sprintf("If the current attribute is set to one of [%v], the following also need to be set: %q", strings.Join(boolValueArray, ","), v.PathExpressions)
	}
	return ""
}

// ValidateString implements validator.String.
func (v AlsoRequiresOnValuesValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If attribute configuration is null, there is nothing else to validate
	if req.ConfigValue.IsNull() {
		return
	}

	for _, value := range v.OnStringValues {
		if value == req.ConfigValue.ValueString() {
			stringvalidator.AlsoRequires(v.PathExpressions...).ValidateString(ctx, req, resp)
			return
		}
	}
}

// ValidateBool implements validator.Bool.
func (v AlsoRequiresOnValuesValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// If attribute configuration is null, there is nothing else to validate
	if req.ConfigValue.IsNull() {
		return
	}

	for _, value := range v.OnBoolValues {
		if value == req.ConfigValue.ValueBool() {
			boolvalidator.AlsoRequires(v.PathExpressions...).ValidateBool(ctx, req, resp)
			return
		}
	}
}

// AlsoRequiresOnValues checks that a set of path.Expression has a non-null value,
// if the current attribute or block is set to one of the values defined in onValues array.
//
// Relative path.Expression will be resolved using the attribute or block
// being validated.
func AlsoRequiresOnStringValues(onValues []string, expressions ...path.Expression) validator.String {
	return AlsoRequiresOnValuesValidator{
		OnStringValues:  onValues,
		PathExpressions: expressions,
	}
}

func AlsoRequiresOnBoolValues(onValues []bool, expressions ...path.Expression) validator.Bool {
	return AlsoRequiresOnValuesValidator{
		OnBoolValues:    onValues,
		PathExpressions: expressions,
	}
}
