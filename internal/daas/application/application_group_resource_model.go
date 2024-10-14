// Copyright © 2024. Citrix Systems, Inc.

package application

import (
	"context"
	"regexp"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// ApplicationGroupResource maps the resource schema data.
type ApplicationGroupResourceModel struct {
	Id                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Description                types.String `tfsdk:"description"`
	RestrictToTag              types.String `tfsdk:"restrict_to_tag"`
	IncludedUsers              types.Set    `tfsdk:"included_users"`   // Set[string]
	DeliveryGroups             types.Set    `tfsdk:"delivery_groups"`  // Set[string]
	Scopes                     types.Set    `tfsdk:"scopes"`           // Set[string]
	BuiltInScopes              types.Set    `tfsdk:"built_in_scopes"`  //Set[String]
	InheritedScopes            types.Set    `tfsdk:"inherited_scopes"` //Set[String]
	ApplicationGroupFolderPath types.String `tfsdk:"application_group_folder_path"`
	Tenants                    types.Set    `tfsdk:"tenants"`  // Set[string]
	Metadata                   types.List   `tfsdk:"metadata"` // List[NameValueStringPairModel]
	Tags                       types.Set    `tfsdk:"tags"`     // Set[string]
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
			"built_in_scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the built-in scopes of the application group.",
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"inherited_scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the inherited scopes of the application group.",
				Computed:    true,
			},
			"application_group_folder_path": schema.StringAttribute{
				Description: "The path of the folder in which the application group is located.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.AdminFolderPathWithBackslashRegex), "Admin Folder Path must not start or end with a backslash"),
					stringvalidator.RegexMatches(regexp.MustCompile(util.AdminFolderPathSpecialCharactersRegex), "Admin Folder Path must not contain any of the following special characters: / ; : # . * ? = < > | [ ] ( ) { } \" ' ` ~ "),
				},
			},
			"tenants": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tenants to associate with the application group.",
				Computed:    true,
			},
			"metadata": util.GetMetadataListSchema("Application Group"),
			"tags": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tags to associate with the application group.",
				Optional:    true,
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

func (ApplicationGroupResourceModel) GetAttributes() map[string]schema.Attribute {
	return ApplicationGroupResourceModel{}.GetSchema().Attributes
}

func (r ApplicationGroupResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, applicationGroup *citrixorchestration.ApplicationGroupDetailResponseModel, dgs *citrixorchestration.ApplicationGroupDeliveryGroupResponseModelCollection, tags []string) ApplicationGroupResourceModel {
	// Overwrite application with refreshed state
	r.Id = types.StringValue(applicationGroup.GetId())
	r.Name = types.StringValue(applicationGroup.GetName())

	// Set optional values
	r.Description = types.StringValue(applicationGroup.GetDescription())

	restrictToTag := applicationGroup.GetRestrictToTag()
	retrictToTagName := restrictToTag.GetName()
	if retrictToTagName != "" {
		r.RestrictToTag = types.StringValue(retrictToTagName)
	} else {
		r.RestrictToTag = types.StringNull()
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

	scopeIdsInPlan := util.StringSetToStringArray(ctx, diagnostics, r.Scopes)
	scopeIds, builtInScopes, inheritedScopeIds, err := util.CategorizeScopes(ctx, client, diagnostics, applicationGroup.GetScopes(), citrixorchestration.SCOPEDOBJECTTYPE_DELIVERY_GROUP, resultDeliveryGroupIds, scopeIdsInPlan)
	if err != nil {
		return r
	}
	r.Scopes = util.StringArrayToStringSet(ctx, diagnostics, scopeIds)
	r.BuiltInScopes = util.StringArrayToStringSet(ctx, diagnostics, builtInScopes)
	r.InheritedScopes = util.StringArrayToStringSet(ctx, diagnostics, inheritedScopeIds)

	adminFolder := applicationGroup.GetAdminFolder()
	adminFolderPath := strings.TrimSuffix(adminFolder.GetName(), "\\")
	if adminFolderPath != "" {
		r.ApplicationGroupFolderPath = types.StringValue(adminFolderPath)
	} else {
		r.ApplicationGroupFolderPath = types.StringNull()
	}

	if len(applicationGroup.GetTenants()) > 0 {
		var remoteTenants []string
		for _, tenant := range applicationGroup.GetTenants() {
			remoteTenants = append(remoteTenants, tenant.GetId())
		}
		r.Tenants = util.StringArrayToStringSet(ctx, diagnostics, remoteTenants)
	} else {
		r.Tenants = types.SetNull(types.StringType)
	}

	effectiveMetadata := util.GetEffectiveMetadata(util.ObjectListToTypedArray[util.NameValueStringPairModel](ctx, diagnostics, r.Metadata), applicationGroup.GetMetadata())

	if len(effectiveMetadata) > 0 {
		r.Metadata = util.RefreshListValueProperties[util.NameValueStringPairModel, citrixorchestration.NameValueStringPairModel](ctx, diagnostics, r.Metadata, effectiveMetadata, util.GetOrchestrationNameValueStringPairKey)
	} else {
		r.Metadata = util.TypedArrayToObjectList[util.NameValueStringPairModel](ctx, diagnostics, nil)
	}

	r.Tags = util.RefreshTagSet(ctx, diagnostics, tags)

	return r
}
