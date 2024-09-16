// Copyright © 2024. Citrix Systems, Inc.
package qcs_account

import (
	"context"
	"encoding/base64"

	quickcreateservice "github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AwsWorkspacesCloudFormationDataSourceModel struct {
	Content types.String `tfsdk:"content"`
}

func (AwsWorkspacesCloudFormationDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "DaaS Quick Deploy - AWS WorkSpaces Core --- Data source to fetch AWS WorkSpaces CloudFormation template.",
		Attributes: map[string]schema.Attribute{
			"content": schema.StringAttribute{
				Description: "Content of the CloudFormation template.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func (AwsWorkspacesCloudFormationDataSourceModel) GetAttributes() map[string]schema.Attribute {
	return AwsWorkspacesCloudFormationDataSourceModel{}.GetSchema().Attributes
}

func (r AwsWorkspacesCloudFormationDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, accountResourceFile *quickcreateservice.AwsEdcAccountResourceFile) AwsWorkspacesCloudFormationDataSourceModel {
	content, err := base64.StdEncoding.DecodeString(accountResourceFile.GetFileContent())
	if err != nil {
		diagnostics.AddError(
			"Failed to decode CloudFormation template file",
			err.Error(),
		)
		return r
	}
	r.Content = types.StringValue(string(content))

	return r
}
