// Copyright © 2024. Citrix Systems, Inc.

package policies

import (
	"encoding/json"
	"regexp"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/citrix/terraform-provider-citrix/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ PolicyFilterInterface = AccessControlFilterModel{}
	_ PolicyFilterInterface = BranchRepeaterFilterModel{}
	_ PolicyFilterInterface = ClientIPFilterModel{}
	_ PolicyFilterInterface = ClientNameFilterModel{}
	_ PolicyFilterInterface = DeliveryGroupFilterModel{}
	_ PolicyFilterInterface = DeliveryGroupTypeFilterModel{}
	_ PolicyFilterInterface = OuFilterModel{}
	_ PolicyFilterInterface = UserFilterModel{}
	_ PolicyFilterInterface = TagFilterModel{}
)

type PolicyFilterInterface interface {
	GetFilterRequest(diagnostics *diag.Diagnostics, serverValue string) (citrixorchestration.FilterRequest, error)
}

type AccessControlFilterModel struct {
	Allowed    types.Bool   `tfsdk:"allowed"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	Connection types.String `tfsdk:"connection"`
	Condition  types.String `tfsdk:"condition"`
	Gateway    types.String `tfsdk:"gateway"`
}

func (AccessControlFilterModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
			"connection": schema.StringAttribute{
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

type BranchRepeaterFilterModel struct {
	Allowed types.Bool `tfsdk:"allowed"`
}

func (BranchRepeaterFilterModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Definition of branch repeater policy filter.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
		},
	}
}

func (BranchRepeaterFilterModel) GetAttributes() map[string]schema.Attribute {
	return BranchRepeaterFilterModel{}.GetSchema().Attributes
}

func (filter BranchRepeaterFilterModel) GetFilterRequest(diagnostics *diag.Diagnostics, serverValue string) (citrixorchestration.FilterRequest, error) {
	branchRepeaterFilterRequest := citrixorchestration.FilterRequest{}
	branchRepeaterFilterRequest.SetFilterType("BranchRepeater")
	branchRepeaterFilterRequest.SetIsAllowed(filter.Allowed.ValueBool())
	branchRepeaterFilterRequest.SetIsEnabled(true)
	branchRepeaterFilterRequest.SetFilterData("")
	return branchRepeaterFilterRequest, nil
}

type ClientIPFilterModel struct {
	Allowed   types.Bool   `tfsdk:"allowed"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	IpAddress types.String `tfsdk:"ip_address"`
}

func (ClientIPFilterModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
			"ip_address": schema.StringAttribute{
				Description: "IP Address of the client to be filtered.",
				Required:    true,
				Validators: []validator.String{
					validators.ValidateIPFilter(),
				},
			},
		},
	}
}

func (ClientIPFilterModel) GetAttributes() map[string]schema.Attribute {
	return ClientIPFilterModel{}.GetSchema().Attributes
}

func (filter ClientIPFilterModel) GetFilterRequest(diagnostics *diag.Diagnostics, serverValue string) (citrixorchestration.FilterRequest, error) {
	filterRequest := citrixorchestration.FilterRequest{}
	filterRequest.SetFilterType("ClientIP")
	filterRequest.SetFilterData(filter.IpAddress.ValueString())
	filterRequest.SetIsAllowed(filter.Allowed.ValueBool())
	filterRequest.SetIsEnabled(filter.Enabled.ValueBool())

	return filterRequest, nil
}

type ClientNameFilterModel struct {
	Allowed    types.Bool   `tfsdk:"allowed"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	ClientName types.String `tfsdk:"client_name"`
}

func (ClientNameFilterModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
			"client_name": schema.StringAttribute{
				Description: "Name of the client to be filtered.",
				Required:    true,
			},
		},
	}
}

func (ClientNameFilterModel) GetAttributes() map[string]schema.Attribute {
	return ClientNameFilterModel{}.GetSchema().Attributes
}

func (filter ClientNameFilterModel) GetFilterRequest(diagnostics *diag.Diagnostics, serverValue string) (citrixorchestration.FilterRequest, error) {
	filterRequest := citrixorchestration.FilterRequest{}
	filterRequest.SetFilterType("ClientName")
	filterRequest.SetFilterData(filter.ClientName.ValueString())
	filterRequest.SetIsAllowed(filter.Allowed.ValueBool())
	filterRequest.SetIsEnabled(filter.Enabled.ValueBool())

	return filterRequest, nil
}

type DeliveryGroupFilterModel struct {
	Allowed         types.Bool   `tfsdk:"allowed"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	DeliveryGroupId types.String `tfsdk:"delivery_group_id"`
}

func (DeliveryGroupFilterModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
			"delivery_group_id": schema.StringAttribute{
				Description: "Id of the delivery group to be filtered.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
		},
	}
}

func (DeliveryGroupFilterModel) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupFilterModel{}.GetSchema().Attributes
}

func (filter DeliveryGroupFilterModel) GetFilterRequest(diagnostics *diag.Diagnostics, serverValue string) (citrixorchestration.FilterRequest, error) {
	filterRequest := citrixorchestration.FilterRequest{}
	filterRequest.SetFilterType("DesktopGroup")

	policyFilterDataClientModel := PolicyFilterUuidDataClientModel{
		Uuid:   filter.DeliveryGroupId.ValueString(),
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

type DeliveryGroupTypeFilterModel struct {
	Allowed           types.Bool   `tfsdk:"allowed"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	DeliveryGroupType types.String `tfsdk:"delivery_group_type"`
}

func (DeliveryGroupTypeFilterModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
			"delivery_group_type": schema.StringAttribute{
				Description: "Type of the delivery groups to be filtered.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{
						"Private",
						"PrivateApp",
						"Shared",
						"SharedApp"}...),
				},
			},
		},
	}
}

func (DeliveryGroupTypeFilterModel) GetAttributes() map[string]schema.Attribute {
	return DeliveryGroupTypeFilterModel{}.GetSchema().Attributes
}

func (filter DeliveryGroupTypeFilterModel) GetFilterRequest(diagnostics *diag.Diagnostics, serverValue string) (citrixorchestration.FilterRequest, error) {
	filterRequest := citrixorchestration.FilterRequest{}
	filterRequest.SetFilterType("DesktopKind")

	filterRequest.SetFilterData(filter.DeliveryGroupType.ValueString())
	filterRequest.SetIsAllowed(filter.Allowed.ValueBool())
	filterRequest.SetIsEnabled(filter.Enabled.ValueBool())
	return filterRequest, nil
}

type OuFilterModel struct {
	Allowed types.Bool   `tfsdk:"allowed"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Ou      types.String `tfsdk:"ou"`
}

func (OuFilterModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
			"ou": schema.StringAttribute{
				Description: "Organizational Unit to be filtered.",
				Required:    true,
			},
		},
	}
}

func (OuFilterModel) GetAttributes() map[string]schema.Attribute {
	return OuFilterModel{}.GetSchema().Attributes
}

func (filter OuFilterModel) GetFilterRequest(diagnostics *diag.Diagnostics, serverValue string) (citrixorchestration.FilterRequest, error) {
	filterRequest := citrixorchestration.FilterRequest{}
	filterRequest.SetFilterType("OU")

	filterRequest.SetFilterData(filter.Ou.ValueString())
	filterRequest.SetIsAllowed(filter.Allowed.ValueBool())
	filterRequest.SetIsEnabled(filter.Enabled.ValueBool())
	return filterRequest, nil
}

type UserFilterModel struct {
	Allowed types.Bool   `tfsdk:"allowed"`
	Enabled types.Bool   `tfsdk:"enabled"`
	UserSid types.String `tfsdk:"sid"`
}

func (UserFilterModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
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

func (filter UserFilterModel) GetFilterRequest(diagnostics *diag.Diagnostics, serverValue string) (citrixorchestration.FilterRequest, error) {
	filterRequest := citrixorchestration.FilterRequest{}
	filterRequest.SetFilterType("User")

	filterRequest.SetFilterData(filter.UserSid.ValueString())
	filterRequest.SetIsAllowed(filter.Allowed.ValueBool())
	filterRequest.SetIsEnabled(filter.Enabled.ValueBool())
	return filterRequest, nil
}

type TagFilterModel struct {
	Allowed types.Bool   `tfsdk:"allowed"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Tag     types.String `tfsdk:"tag"`
}

func (TagFilterModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Required:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Required:    true,
			},
			"tag": schema.StringAttribute{
				Description: "Tag to be filtered.",
				Required:    true,
			},
		},
	}
}

func (TagFilterModel) GetAttributes() map[string]schema.Attribute {
	return TagFilterModel{}.GetSchema().Attributes
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
