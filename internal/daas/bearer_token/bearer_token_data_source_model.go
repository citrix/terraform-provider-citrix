// Copyright © 2024. Citrix Systems, Inc.

package bearer_token

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BearerTokenDataSourceModel struct {
	BearerToken types.String `tfsdk:"bearer_token"`
}

func (BearerTokenDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "CVAD --- Data source to get bearer token.",

		Attributes: map[string]schema.Attribute{
			"bearer_token": schema.StringAttribute{
				Description: "Value of the bearer token.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func (r BearerTokenDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, bearerToken string) BearerTokenDataSourceModel {
	r.BearerToken = types.StringValue(bearerToken)

	return r
}
