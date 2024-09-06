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
	Id                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Description                types.String `tfsdk:"description"`
	RestrictToTag              types.String `tfsdk:"restrict_to_tag"`
	IncludedUsers              types.Set    `tfsdk:"included_users"`  // Set[string]
	DeliveryGroups             types.Set    `tfsdk:"delivery_groups"` // Set[string]
	Scopes                     types.Set    `tfsdk:"scopes"`          // Set[string]
	ApplicationGroupFolderPath types.String `tfsdk:"application_group_folder_path"`
	Tenants                    types.Set    `tfsdk:"tenants"` // Set[string]
}

func (ApplicationGroupResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Resource for creating and managing application group.",
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
				Description: "Users who can use this application group. " +
					"\n\n-> **Note** User must be in `Domain\\UserOrGroupName` or `user@domain.com` format",
				Optional: true,
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
					setvalidator.ValueStringsAre(
						validator.String(
							stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
						),
					),
				},
			},
			"application_group_folder_path": schema.StringAttribute{
				Description: "The path of the folder in which the application group is located.",
				Optional:    true,
			},
			"tenants": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tenants to associate with the application group.",
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (ApplicationGroupResourceModel) GetAttributes() map[string]schema.Attribute {
	return ApplicationGroupResourceModel{}.GetSchema().Attributes
}

func (r ApplicationGroupResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, applicationGroup *citrixorchestration.ApplicationGroupDetailResponseModel, dgs *citrixorchestration.ApplicationGroupDeliveryGroupResponseModelCollection) ApplicationGroupResourceModel {
	// Overwrite application with refreshed state
	r.Id = types.StringValue(applicationGroup.GetId())
	r.Name = types.StringValue(applicationGroup.GetName())

	// Set optional values
	if applicationGroup.GetDescription() != "" {
		r.Description = types.StringValue(applicationGroup.GetDescription())
	} else {
		r.Description = types.StringNull()
	}

	restrictToTag := applicationGroup.GetRestrictToTag()
	retrictToTagName := restrictToTag.GetName()
	if retrictToTagName != "" {
		r.RestrictToTag = types.StringValue(retrictToTagName)
	} else {
		r.RestrictToTag = types.StringNull()
	}

	if applicationGroup.GetScopes() != nil {
		scopeIdsInState := util.StringSetToStringArray(ctx, diagnostics, r.Scopes)
		scopeIds := util.GetIdsForFilteredScopeObjects(scopeIdsInState, applicationGroup.GetScopes())
		r.Scopes = util.StringArrayToStringSet(ctx, diagnostics, scopeIds)
	}

	// Set included users
	includedUsers := applicationGroup.GetIncludedUsers() //only set IncludedUsers to null when it is null in the configuration
	if len(includedUsers) == 0 && r.IncludedUsers.IsNull() {
		r.IncludedUsers = types.SetNull(types.StringType)
	} else { // If included users is not null or empty list, we need to update the list
		r.IncludedUsers = util.RefreshUsersList(ctx, diagnostics, r.IncludedUsers, includedUsers)
	}

	resultDeliveryGroupIds := []string{}
	for _, deliveryGroup := range dgs.Items {
		resultDeliveryGroupIds = append(resultDeliveryGroupIds, deliveryGroup.GetId())
	}
	r.DeliveryGroups = util.StringArrayToStringSet(ctx, diagnostics, resultDeliveryGroupIds)

	adminFolder := applicationGroup.GetAdminFolder()
	adminFolderPath := adminFolder.GetName()
	if adminFolderPath != "" {
		r.ApplicationGroupFolderPath = types.StringValue(adminFolderPath)
	} else {
		r.ApplicationGroupFolderPath = types.StringNull()
	}

	if len(applicationGroup.GetTenants()) > 0 || !r.Tenants.IsNull() {
		var remoteTenants []string
		for _, tenant := range applicationGroup.GetTenants() {
			remoteTenants = append(remoteTenants, tenant.GetId())
		}
		r.Tenants = util.StringArrayToStringSet(ctx, diagnostics, remoteTenants)
	} else {
		r.Tenants = types.SetNull(types.StringType)
	}

	return r
}
