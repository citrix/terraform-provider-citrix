// Copyright © 2024. Citrix Systems, Inc.

package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.String = AlsoRequiresOnValuesValidator{}
)

// AlsoRequiresOnValuesValidator is the underlying struct implementing AlsoRequiresOnValue.
type AlsoRequiresOnValuesValidator struct {
	OnValues        []string
	PathExpressions path.Expressions
}

// Description implements validator.String.
func (v AlsoRequiresOnValuesValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

// MarkdownDescription implements validator.String.
func (v AlsoRequiresOnValuesValidator) MarkdownDescription(context.Context) string {
	return fmt.Sprintf("If the current attribute is set to one of [%s], the following also need to be set: %q", strings.Join(v.OnValues, ","), v.PathExpressions)
}

// ValidateString implements validator.String.
func (v AlsoRequiresOnValuesValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If attribute configuration is null, there is nothing else to validate
	if req.ConfigValue.IsNull() {
		return
	}

	for _, value := range v.OnValues {
		if value == req.ConfigValue.ValueString() {
			stringvalidator.AlsoRequires(v.PathExpressions...).ValidateString(ctx, req, resp)
			return
		}
	}
}

// AlsoRequiresOnValues checks that a set of path.Expression has a non-null value,
// if the current attribute or block is set to one of the values defined in onValues array.
//
// Relative path.Expression will be resolved using the attribute or block
// being validated.
func AlsoRequiresOnValues(onValues []string, expressions ...path.Expression) validator.String {
	return AlsoRequiresOnValuesValidator{
		OnValues:        onValues,
		PathExpressions: expressions,
	}
}
