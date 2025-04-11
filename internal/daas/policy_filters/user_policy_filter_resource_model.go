// Copyright © 2024. Citrix Systems, Inc.

package policy_filters

import (
	"context"
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

type UserFilterModel struct {
	Id       types.String `tfsdk:"id"`
	PolicyId types.String `tfsdk:"policy_id"`
	Allowed  types.Bool   `tfsdk:"allowed"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	UserSid  types.String `tfsdk:"sid"`
}

func (UserFilterModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: getPolicyFilterResourceDescription("User"),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the user policy filter.",
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
			"sid": schema.StringAttribute{
				Description: "SID of the user or user group to be filtered.",
				Required:    true,
			},
		},
	}
}

func (UserFilterModel) GetAttributes() map[string]schema.Attribute {
	return UserFilterModel{}.GetSchema().Attributes
}

func (filter UserFilterModel) GetId() string {
	return filter.Id.ValueString()
}

func (filter UserFilterModel) GetPolicyId() string {
	return filter.PolicyId.ValueString()
}

func (filter UserFilterModel) GetFilterRequest(diagnostics *diag.Diagnostics, serverValue string) (citrixorchestration.FilterRequest, error) {
	filterRequest := citrixorchestration.FilterRequest{}
	filterRequest.SetFilterType("User")

	filterRequest.SetFilterData(filter.UserSid.ValueString())
	filterRequest.SetIsAllowed(filter.Allowed.ValueBool())
	filterRequest.SetIsEnabled(filter.Enabled.ValueBool())
	return filterRequest, nil
}

func (r UserFilterModel) RefreshPropertyValues(ctx context.Context, diags *diag.Diagnostics, filter citrixorchestration.FilterResponse) UserFilterModel {
	r.Id = types.StringValue(filter.GetFilterGuid())
	r.PolicyId = types.StringValue(filter.GetPolicyGuid())
	r.Allowed = types.BoolValue(filter.GetIsAllowed())
	r.Enabled = types.BoolValue(filter.GetIsEnabled())
	r.UserSid = types.StringValue(filter.GetFilterData())

	return r
}
