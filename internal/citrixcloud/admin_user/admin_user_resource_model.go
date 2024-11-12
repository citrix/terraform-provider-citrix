// Copyright © 2024. Citrix Systems, Inc.

package cc_admin_user

import (
	"context"
	"regexp"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/ccadmins"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CCAdminUserResourceModel maps the resource schema data.
type CCAdminUserResourceModel struct {
	AdminId            types.String `tfsdk:"admin_id"`
	AccessType         types.String `tfsdk:"access_type"`
	DisplayName        types.String `tfsdk:"display_name"`
	Email              types.String `tfsdk:"email"`
	FirstName          types.String `tfsdk:"first_name"`
	LastName           types.String `tfsdk:"last_name"`
	ProviderType       types.String `tfsdk:"provider_type"`
	Type               types.String `tfsdk:"type"`
	Policies           types.List   `tfsdk:"policies"` // List[CCAdminPolicyResourceModel]
	ExternalProviderId types.String `tfsdk:"external_provider_id"`
	ExternalUserId     types.String `tfsdk:"external_user_id"`
}

var _ util.RefreshableListItemWithAttributes[ccadmins.AdministratorAccessPolicyModel] = CCAdminPolicyResourceModel{}

type CCAdminPolicyResourceModel struct {
	Name        types.String `tfsdk:"name"`
	ServiceName types.String `tfsdk:"service_name"`
	Scopes      types.Set    `tfsdk:"scopes"` // Set[string]
}

func (r CCAdminPolicyResourceModel) GetKey() string {
	return r.Name.ValueString()
}

func (CCAdminPolicyResourceModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the policy to be associated with the admin user.",
				Required:    true,
			},
			"service_name": schema.StringAttribute{
				Description: "Name of the service to be associated with the admin user. Currently, this attribute can be set to `XenDesktop`, `Platform`, `CAS`, or `WEM`.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						ADMINISTRATORSERVICENAME_XENDESKTOP,
						ADMINISTRATORSERVICENAME_PLATFORM,
						ADMINISTRATORACCESSTYPE_CAS,
						ADMINISTRATORACCESSTYPE_WEM,
					),
				},
			},
			"scopes": schema.SetAttribute{
				Description: "Scope names to be associated with the admin user.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (CCAdminPolicyResourceModel) GetAttributes() map[string]schema.Attribute {
	return CCAdminPolicyResourceModel{}.GetSchema().Attributes
}

func (CCAdminUserResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "Citrix Cloud --- Manages an administrator user for cloud environment.",

		Attributes: map[string]schema.Attribute{
			"admin_id": schema.StringAttribute{
				Description: "Id of the administrator.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_type": schema.StringAttribute{
				Description: "Access Type of the user. Currently, this attribute can be set to `Full` or `Custom`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(ccadmins.ADMINISTRATORACCESSTYPE_FULL),
						string(ccadmins.ADMINISTRATORACCESSTYPE_CUSTOM),
					),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "Display name for the user.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				Description: "Email of the user where the invitation link will be sent.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"first_name": schema.StringAttribute{
				Description: "First name of the user.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"last_name": schema.StringAttribute{
				Description: "Last name of the user.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provider_type": schema.StringAttribute{
				Description: "Identity provider for the administrator or group you want to add. Currently, this attribute can be set to `CitrixSTS`,`AzureAd` or `Ad`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(ccadmins.ADMINISTRATORPROVIDERTYPE_CITRIX_STS),
						string(ccadmins.ADMINISTRATORPROVIDERTYPE_AZURE_AD),
						string(ccadmins.ADMINISTRATOREXTERNALPROVIDERTYPE_AD),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Type of administrator being added. Currently, this attribute can be set to `AdministratorUser` or `AdministratorGroup`. Note: `AdministratorGroup` is only supported for `AzureAd` and `Ad` provider type.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(ccadmins.ADMINISTRATORTYPE_ADMINISTRATOR_USER),
						string(ccadmins.ADMINISTRATORTYPE_ADMINISTRATOR_GROUP),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policies": schema.ListNestedAttribute{
				Description:  "Policies to be associated with the admin user. Only applicable when access_type is Custom.",
				Optional:     true,
				NestedObject: CCAdminPolicyResourceModel{}.GetSchema(),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"external_provider_id": schema.StringAttribute{
				Description: " External provider Id for directory. For `AzureAd`, specify the external tenant ID. For `Ad`, specify the AD domain name in FQDN format (e.g., MyDomain.com)",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"external_user_id": schema.StringAttribute{
				Description: "External objectId for user or group from the directory",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validator.String(
						stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					),
				},
			},
		},
	}
}

func (CCAdminUserResourceModel) GetAttributes() map[string]schema.Attribute {
	return CCAdminUserResourceModel{}.GetSchema().Attributes
}

func (r CCAdminUserResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, adminUser ccadmins.AdministratorResult) CCAdminUserResourceModel {
	r.AccessType = types.StringValue(string(adminUser.GetAccessType()))
	r.Type = types.StringValue(string(adminUser.GetType()))

	// Set the admin id based on the type of the admin user
	if r.Type.ValueString() == string(ccadmins.ADMINISTRATORTYPE_ADMINISTRATOR_GROUP) {
		r.AdminId = types.StringValue(adminUser.GetUcOid())
	} else {
		r.AdminId = types.StringValue(adminUser.GetUserId())
	}

	if !providerTypeExists(adminUser.GetLegacyProviders(), r.ProviderType.ValueString()) {
		r.ProviderType = types.StringValue(string(adminUser.GetProviderType()))
	}

	if !strings.EqualFold(r.Email.ValueString(), adminUser.GetEmail()) {
		r.Email = types.StringValue(adminUser.GetEmail())
	}

	if !r.DisplayName.IsNull() {
		r.DisplayName = types.StringValue(adminUser.GetDisplayName())
	}
	if !r.FirstName.IsNull() {
		r.FirstName = types.StringValue(adminUser.GetFirstName())
	}
	if !r.LastName.IsNull() {
		r.LastName = types.StringValue(adminUser.GetLastName())
	}

	if !r.ExternalProviderId.IsNull() {
		r.ExternalProviderId = types.StringValue(adminUser.GetProviderId())
	}

	if !r.ExternalUserId.IsNull() {
		r.ExternalUserId = types.StringValue(getExternalUserId(adminUser.GetExternalOid()))
	}
	return r
}

func (r CCAdminPolicyResourceModel) RefreshListItem(ctx context.Context, diags *diag.Diagnostics, adminAccessPolicy ccadmins.AdministratorAccessPolicyModel) util.ResourceModelWithAttributes {
	// Update the name and service name from the admin access policy
	r.Name = types.StringValue(adminAccessPolicy.GetDisplayName())
	r.ServiceName = types.StringValue(adminAccessPolicy.GetServiceName())

	// Initialize scope names and get remote scope choices
	var scopeNames []string
	remoteScopeChoices := adminAccessPolicy.GetScopeChoices()
	scopes := util.StringSetToStringArray(ctx, diags, r.Scopes)

	// Create a map for quick lookup of scope names from the scopes
	scopeNameMap := make(map[string]string)
	for _, scope := range scopes {
		scopeNameMap[strings.ToLower(scope)] = scope
	}

	if remoteScopeChoices.GetChoices() != nil {
		for _, remoteScope := range remoteScopeChoices.GetChoices() {
			// Checks if the scope choice is selected or not.
			checkable := remoteScope.GetCheckable()
			if checkable.GetValue() {
				// Check if the remote scope name exists in the scopes
				if configScopeName, exists := scopeNameMap[strings.ToLower(remoteScope.GetDisplayName())]; exists {
					scopeNames = append(scopeNames, configScopeName)
				} else {
					scopeNames = append(scopeNames, remoteScope.GetDisplayName())
				}
			}
		}
		r.Scopes = util.StringArrayToStringSet(ctx, diags, scopeNames)
	} else {
		r.Scopes = types.SetNull(types.StringType)
	}
	return r
}

// Filters policies based on the names
func filterPolicies(remotePolicies []ccadmins.AdministratorAccessPolicyModel, policies []CCAdminPolicyResourceModel) []ccadmins.AdministratorAccessPolicyModel {
	var filteredPolicies []ccadmins.AdministratorAccessPolicyModel

	// Create a map for quick lookup of policy names from the policies
	policyNameMap := make(map[string]string)
	for _, policy := range policies {
		policyNameMap[strings.ToLower(policy.Name.ValueString())] = policy.Name.ValueString()
	}

	for _, remotePolicy := range remotePolicies {
		// Check if the scope choice is selected
		checkable := remotePolicy.GetCheckable()
		if checkable.GetValue() {
			// Check if the remote policy name exists in the policies
			trimmedRemotePolicyDisplayName := strings.TrimSuffix(remotePolicy.GetDisplayName(), util.AdminUserMonitorAccessPolicySuffix)
			if configPolicyName, exists := policyNameMap[strings.ToLower(remotePolicy.GetDisplayName())]; exists {
				remotePolicy.SetDisplayName(configPolicyName)
			} else if configPolicyName, exists := policyNameMap[strings.ToLower(trimmedRemotePolicyDisplayName)]; exists {
				remotePolicy.SetDisplayName(configPolicyName)
			}
			filteredPolicies = append(filteredPolicies, remotePolicy)
		}
	}

	return filteredPolicies
}

func (r CCAdminUserResourceModel) RefreshPropertyValuesForPolicies(ctx context.Context, diagnostics *diag.Diagnostics, adminAccessPolicy *ccadmins.AdministratorAccessModel) CCAdminUserResourceModel {
	if adminAccessPolicy.GetPolicies() != nil {
		policies := util.ObjectListToTypedArray[CCAdminPolicyResourceModel](ctx, diagnostics, r.Policies)
		filteredPolicies := filterPolicies(adminAccessPolicy.GetPolicies(), policies)
		r.Policies = util.RefreshListValueProperties[CCAdminPolicyResourceModel, ccadmins.AdministratorAccessPolicyModel](ctx, diagnostics, r.Policies, filteredPolicies, util.GetCCAdminAccessPolicyNameKey)
	}
	return r
}
