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

type TagFilterModel struct {
	Id       types.String `tfsdk:"id"`
	PolicyId types.String `tfsdk:"policy_id"`
	Allowed  types.Bool   `tfsdk:"allowed"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Tag      types.String `tfsdk:"tag"`
}

func (TagFilterModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: getPolicyFilterResourceDescription("Tag"),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the tag policy filter.",
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
			"tag": schema.StringAttribute{
				Description: "ID of the tag to be filtered.",
				Required:    true,
			},
		},
	}
}

func (TagFilterModel) GetAttributes() map[string]schema.Attribute {
	return TagFilterModel{}.GetSchema().Attributes
}

func (filter TagFilterModel) GetId() string {
	return filter.Id.ValueString()
}

func (filter TagFilterModel) GetPolicyId() string {
	return filter.PolicyId.ValueString()
}

func (filter TagFilterModel) GetFilterRequest(diagnostics *diag.Diagnostics, serverValue string) (citrixorchestration.FilterRequest, error) {
	filterRequest := citrixorchestration.FilterRequest{}
	filterRequest.SetFilterType("DesktopTag")

	policyFilterDataClientModel := PolicyFilterUuidDataClientModel{
		Uuid:   filter.Tag.ValueString(),
		Server: serverValue,
	}

	policyFilterDataJson, err := json.Marshal(policyFilterDataClientModel)
	if err != nil {
		diagnostics.AddError(
			"Error adding Access Control Policy Filter to Policy Set. ",
			"An unexpected error occurred: "+err.Error(),
		)
		return filterRequest, err
	}

	filterRequest.SetFilterData(string(policyFilterDataJson))
	filterRequest.SetIsAllowed(filter.Allowed.ValueBool())
	filterRequest.SetIsEnabled(filter.Enabled.ValueBool())
	return filterRequest, nil
}

func (r TagFilterModel) RefreshPropertyValues(ctx context.Context, diags *diag.Diagnostics, filter citrixorchestration.FilterResponse) TagFilterModel {
	var uuidFilterData PolicyFilterUuidDataClientModel
	_ = json.Unmarshal([]byte(filter.GetFilterData()), &uuidFilterData)

	r.Id = types.StringValue(filter.GetFilterGuid())
	r.PolicyId = types.StringValue(filter.GetPolicyGuid())
	r.Allowed = types.BoolValue(filter.GetIsAllowed())
	r.Enabled = types.BoolValue(filter.GetIsEnabled())
	r.Tag = types.StringValue(uuidFilterData.Uuid)

	return r
}
