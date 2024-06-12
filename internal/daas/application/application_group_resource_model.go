// Copyright © 2024. Citrix Systems, Inc.

package application

import (
	"context"
	"regexp"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// ApplicationGroupResource maps the resource schema data.
type ApplicationGroupResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	RestrictToTag  types.String `tfsdk:"restrict_to_tag"`
	IncludedUsers  types.Set    `tfsdk:"included_users"`  // Set[string]
	DeliveryGroups types.Set    `tfsdk:"delivery_groups"` // Set[string]
	Scopes         types.Set    `tfsdk:"scopes"`          // Set[string]
}

func GetApplicationGroupSchema() schema.Schema {
	return schema.Schema{
		Description: "Resource for creating and managing application group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the application group.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the application group to create.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the application group.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"restrict_to_tag": schema.StringAttribute{
				Description: "The tag to restrict the application group to.",
				Optional:    true,
			},
			"included_users": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "Users who can use this application group. Must be in `Domain\\UserOrGroupName` or `user@domain.com` format",
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.SamAndUpnRegex), "must be in `Domain\\UserOrGroupName` or `user@domain.com` format"),
						),
					),
				},
			},
			"delivery_groups": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "Delivery groups to associate with the application group.",
				Required:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
				},
			},
			"scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the scopes for the application group to be a part of.",
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
				},
			},
		},
	}
}

func (appGroup ApplicationGroupResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, application *citrixorchestration.ApplicationGroupDetailResponseModel, dgs *citrixorchestration.ApplicationGroupDeliveryGroupResponseModelCollection) ApplicationGroupResourceModel {
	// Overwrite application with refreshed state
	appGroup.Id = types.StringValue(application.GetId())
	appGroup.Name = types.StringValue(application.GetName())

	// Set optional values
	if application.GetDescription() != "" {
		appGroup.Description = types.StringValue(application.GetDescription())
	} else {
		appGroup.Description = types.StringNull()
	}

	restrictToTag := application.GetRestrictToTag()
	retrictToTagName := restrictToTag.GetName()
	if retrictToTagName != "" {
		appGroup.RestrictToTag = types.StringValue(retrictToTagName)
	} else {
		appGroup.RestrictToTag = types.StringNull()
	}

	if application.GetScopes() != nil {
		scopeIds := util.GetIdsForScopeObjects(application.GetScopes())
		appGroup.Scopes = util.StringArrayToStringSet(ctx, diagnostics, scopeIds)
	}

	// Set included users
	includedUsers := application.GetIncludedUsers() //only set IncludedUsers to null when it is null in the configuration
	if len(includedUsers) == 0 && appGroup.IncludedUsers.IsNull() {
		appGroup.IncludedUsers = types.SetNull(types.StringType)
	} else { // If included users is not null or empty list, we need to update the list
		appGroup.IncludedUsers = util.RefreshUsersList(ctx, diagnostics, appGroup.IncludedUsers, includedUsers)
	}

	resultDeliveryGroupIds := []string{}
	for _, deliveryGroup := range dgs.Items {
		resultDeliveryGroupIds = append(resultDeliveryGroupIds, deliveryGroup.GetId())
	}
	appGroup.DeliveryGroups = util.StringArrayToStringSet(ctx, diagnostics, resultDeliveryGroupIds)

	return appGroup
}
