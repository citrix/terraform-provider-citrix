// Copyright © 2025. Citrix Systems, Inc.

package service_account

import (
	"context"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ServiceAccountModel struct {
	Id                                   types.String `tfsdk:"id"`
	DisplayName                          types.String `tfsdk:"display_name"`
	Description                          types.String `tfsdk:"description"`
	IdentityProviderType                 types.String `tfsdk:"identity_provider_type"`
	IdentityProviderIdentifier           types.String `tfsdk:"identity_provider_identifier"`
	AccountId                            types.String `tfsdk:"account_id"`
	AccountSecret                        types.String `tfsdk:"account_secret"`
	AccountSecretFormat                  types.String `tfsdk:"account_secret_format"`
	SecretExpiryTime                     types.String `tfsdk:"secret_expiry_time"`
	EnableIntuneEnrolledDeviceManagement types.Bool   `tfsdk:"enable_intune_enrolled_device_management"`
	Scopes                               types.Set    `tfsdk:"scopes"` // Set[string]
}

func (ServiceAccountModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Resource for creating and managing service accounts.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "A friendly name for the service account.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description for the service account.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"identity_provider_type": schema.StringAttribute{
				Description: "The identity provider type for the service account. Possible values are `ActiveDirectory` and `AzureAD`." +
					"\n\n -> **Note** 'Device.ReadWrite.All' permission is required for the service principal for Azure AD joined device management.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixorchestration.IDENTITYTYPE_ACTIVE_DIRECTORY),
						string(citrixorchestration.IDENTITYTYPE_AZURE_AD),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"identity_provider_identifier": schema.StringAttribute{
				Description: "The identity provider identifier for the service account." +
					"\n\n -> **Note** For Active Directory, this is the domain name in the FQDN format. For example, `domain.com`. For AzureAD, this is the tenant ID.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"account_id": schema.StringAttribute{
				Description: "The account ID of the service account." +
					"\n\n -> **Note** For Active Directory, this is the username. Username should be in `Domain\\UserName` format. For AzureAD, this is the application ID.",
				Required: true,
			},
			"account_secret": schema.StringAttribute{
				Description: "The password for the service account." +
					"\n\n -> **Note** For Active Directory, this is the password. For AzureAD, this is the client secret.",
				Required:  true,
				Sensitive: true,
			},
			"account_secret_format": schema.StringAttribute{
				Description: "The format of the account secret. Possible values are `PlainText` and `Base64`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixorchestration.IDENTITYPASSWORDFORMAT_BASE64),
						string(citrixorchestration.IDENTITYPASSWORDFORMAT_PLAIN_TEXT),
					),
				},
			},
			"secret_expiry_time": schema.StringAttribute{
				Description: "The UTC expiration date of the account secret." +
					"\n\n -> **Note** The expected format is `YYYY-MM-DD`.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^((?:19|20|21)\d\d)[-](0[1-9]|1[012])[-](0[1-9]|[12][0-9]|3[01])$`), "ensure date is valid and is in the format YYYY-MM-DD"),
				},
			},
			"enable_intune_enrolled_device_management": schema.BoolAttribute{
				Description: "Indicates whether the service account can perform Microsoft Intune enrolled device management. This is applicable only for AzureAD identity provider type." +
					"\n\n -> **Note** 'DeviceManagementManagedDevices.ReadWrite.All' permission is required for the service principal before enabling this capability.",
				Optional: true,
			},
			"scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the scopes for the service account to be a part of.",
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
		},
	}
}

func (ServiceAccountModel) GetAttributes() map[string]schema.Attribute {
	return ServiceAccountModel{}.GetSchema().Attributes
}

func (r ServiceAccountModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, serviceAccountModel *citrixorchestration.ServiceAccountResponseModel) ServiceAccountModel {
	r.Id = types.StringValue(serviceAccountModel.GetServiceAccountUid())
	if !strings.EqualFold(r.DisplayName.ValueString(), serviceAccountModel.GetDisplayName()) {
		r.DisplayName = types.StringValue(serviceAccountModel.GetDisplayName())
	}
	r.Description = types.StringValue(serviceAccountModel.GetDescription())
	r.IdentityProviderType = types.StringValue(serviceAccountModel.GetIdentityProviderType())
	if !strings.EqualFold(r.IdentityProviderIdentifier.ValueString(), serviceAccountModel.GetIdentityProviderIdentifier()) {
		r.IdentityProviderIdentifier = types.StringValue(serviceAccountModel.GetIdentityProviderIdentifier())
	}
	if !strings.EqualFold(r.AccountId.ValueString(), serviceAccountModel.GetAccountId()) {
		r.AccountId = types.StringValue(serviceAccountModel.GetAccountId())
	}

	expiryDateTime, err := time.Parse(time.RFC3339, serviceAccountModel.GetSecretExpiryTime())
	if err == nil {
		year := expiryDateTime.Year()
		month := expiryDateTime.Month()
		day := expiryDateTime.Day()

		if year == 9999 && month == 12 && day == 31 {
			r.SecretExpiryTime = types.StringNull()
		} else {
			expiryDateStr := expiryDateTime.Format(time.DateOnly) // "YYYY-MM-DD"
			r.SecretExpiryTime = types.StringValue(expiryDateStr)
		}
	}

	capabilities := serviceAccountModel.GetCapabilities()
	if slices.ContainsFunc(capabilities, func(capability citrixorchestration.ServiceAccountCapabilityReference) bool {
		return strings.EqualFold(capability.GetName(), util.ServiceAccountIntuneEnrolledDeviceManagementCapability)
	}) {
		r.EnableIntuneEnrolledDeviceManagement = types.BoolValue(true)
	} else if !r.EnableIntuneEnrolledDeviceManagement.IsNull() {
		r.EnableIntuneEnrolledDeviceManagement = types.BoolValue(false)
	}

	scopeIdsInState := util.StringSetToStringArray(ctx, diagnostics, r.Scopes)
	scopeIds := util.GetIdsForFilteredScopeObjects(scopeIdsInState, serviceAccountModel.GetScopes())
	r.Scopes = util.StringArrayToStringSet(ctx, diagnostics, scopeIds)

	return r
}
