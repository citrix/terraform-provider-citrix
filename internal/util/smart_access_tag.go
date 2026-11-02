// Copyright © 2026. Citrix Systems, Inc.

package util

import (
	"context"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ RefreshableListItemWithAttributes[citrixorchestration.SmartAccessTagResponseModel] = SmartAccessTagModel{}

// Terraform model for a SmartAccess filter matched against the tags of a user's connection
type SmartAccessTagModel struct {
	FilterName  types.String `tfsdk:"filter_name"`
	FilterValue types.String `tfsdk:"filter_value"`
}

func (r SmartAccessTagModel) GetKey() string {
	return r.FilterName.ValueString() + r.FilterValue.ValueString()
}

// RefreshListItem implements RefreshableListItemWithAttributes.
func (r SmartAccessTagModel) RefreshListItem(ctx context.Context, diagnostics *diag.Diagnostics, remote citrixorchestration.SmartAccessTagResponseModel) ResourceModelWithAttributes {
	if !strings.EqualFold(r.FilterName.ValueString(), remote.GetFarm()) {
		r.FilterName = types.StringValue(remote.GetFarm())
	}
	if !strings.EqualFold(r.FilterValue.ValueString(), remote.GetFilter()) {
		r.FilterValue = types.StringValue(remote.GetFilter())
	}

	return r
}

func (SmartAccessTagModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"filter_name": schema.StringAttribute{
				Description: "The name of the filter.",
				Required:    true,
			},
			"filter_value": schema.StringAttribute{
				Description: "The value of the filter.",
				Required:    true,
			},
		},
	}
}

func (SmartAccessTagModel) GetAttributes() map[string]schema.Attribute {
	return SmartAccessTagModel{}.GetSchema().Attributes
}

func (SmartAccessTagModel) GetDataSourceSchema() datasourceSchema.NestedAttributeObject {
	return datasourceSchema.NestedAttributeObject{
		Attributes: map[string]datasourceSchema.Attribute{
			"filter_name": datasourceSchema.StringAttribute{
				Description: "The name of the filter.",
				Computed:    true,
			},
			"filter_value": datasourceSchema.StringAttribute{
				Description: "The value of the filter.",
				Computed:    true,
			},
		},
	}
}

func (SmartAccessTagModel) GetDataSourceAttributes() map[string]datasourceSchema.Attribute {
	return SmartAccessTagModel{}.GetDataSourceSchema().Attributes
}

func GetSmartAccessTagsRequestModel(smartAccessTags []SmartAccessTagModel) []citrixorchestration.SmartAccessTagRequestModel {
	smartAccessTagRequests := []citrixorchestration.SmartAccessTagRequestModel{}
	for _, smartAccessTag := range smartAccessTags {
		var smartAccessTagRequestModel citrixorchestration.SmartAccessTagRequestModel
		smartAccessTagRequestModel.SetFarm(smartAccessTag.FilterName.ValueString())
		smartAccessTagRequestModel.SetFilter(smartAccessTag.FilterValue.ValueString())
		smartAccessTagRequests = append(smartAccessTagRequests, smartAccessTagRequestModel)
	}
	return smartAccessTagRequests
}

func ParseSmartAccessTagsToPluginModel(smartAccessTags []citrixorchestration.SmartAccessTagResponseModel) []SmartAccessTagModel {
	res := []SmartAccessTagModel{}
	for _, smartAccessTag := range smartAccessTags {
		res = append(res, SmartAccessTagModel{
			FilterName:  types.StringValue(smartAccessTag.GetFarm()),
			FilterValue: types.StringValue(smartAccessTag.GetFilter()),
		})
	}
	return res
}
