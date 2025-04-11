// Copyright © 2024. Citrix Systems, Inc.

package policy_filters

import (
	"context"
	"encoding/json"
	"regexp"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AccessControlFilterModel struct {
	Id         types.String `tfsdk:"id"`
	PolicyId   types.String `tfsdk:"policy_id"`
	Allowed    types.Bool   `tfsdk:"allowed"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	Connection types.String `tfsdk:"connection_type"`
	Condition  types.String `tfsdk:"condition"`
	Gateway    types.String `tfsdk:"gateway"`
}

func (AccessControlFilterModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: getPolicyFilterResourceDescription("Access Control"),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the policy filter.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_id": schema.StringAttribute{
				Description: "Id of the policy to which the filter belongs.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
			"connection_type": schema.StringAttribute{
				Description: "Gateway connection for the policy filter.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{
						"WithAccessGateway",
						"WithoutAccessGateway"}...),
				},
			},
			"condition": schema.StringAttribute{
				Description: "Gateway condition for the policy filter.",
				Required:    true,
			},
			"gateway": schema.StringAttribute{
				Description: "Gateway for the policy filter.",
				Required:    true,
			},
		},
	}
}

func (AccessControlFilterModel) GetAttributes() map[string]schema.Attribute {
	return AccessControlFilterModel{}.GetSchema().Attributes
}

func (filter AccessControlFilterModel) GetId() string {
	return filter.Id.ValueString()
}

func (filter AccessControlFilterModel) GetPolicyId() string {
	return filter.PolicyId.ValueString()
}

func (filter AccessControlFilterModel) GetFilterRequest(diagnostics *diag.Diagnostics, serverValue string) (citrixorchestration.FilterRequest, error) {
	filterRequest := citrixorchestration.FilterRequest{}
	filterRequest.SetFilterType("AccessControl")

	policyFilterDataClientModel := PolicyFilterGatewayDataClientModel{
		Connection: filter.Connection.ValueString(),
		Condition:  filter.Condition.ValueString(),
		Gateway:    filter.Gateway.ValueString(),
	}

	policyFilterDataJson, err := json.Marshal(policyFilterDataClientModel)
	if err != nil {
		diagnostics.AddError(
			"Error constructing Access Control Policy Filter request.",
			"An unexpected error occurred: "+err.Error(),
		)
		return filterRequest, err
	}
	filterRequest.SetFilterData(string(policyFilterDataJson))
	filterRequest.SetIsAllowed(filter.Allowed.ValueBool())
	filterRequest.SetIsEnabled(filter.Enabled.ValueBool())
	return filterRequest, nil
}

func (r AccessControlFilterModel) RefreshPropertyValues(ctx context.Context, diags *diag.Diagnostics, filter citrixorchestration.FilterResponse) AccessControlFilterModel {
	var gatewayFilterData PolicyFilterGatewayDataClientModel
	_ = json.Unmarshal([]byte(filter.GetFilterData()), &gatewayFilterData)
	r.Id = types.StringValue(filter.GetFilterGuid())
	r.PolicyId = types.StringValue(filter.GetPolicyGuid())
	r.Allowed = types.BoolValue(filter.GetIsAllowed())
	r.Enabled = types.BoolValue(filter.GetIsEnabled())
	r.Connection = types.StringValue(gatewayFilterData.Connection)
	r.Condition = types.StringValue(gatewayFilterData.Condition)
	r.Gateway = types.StringValue(gatewayFilterData.Gateway)

	return r
}
