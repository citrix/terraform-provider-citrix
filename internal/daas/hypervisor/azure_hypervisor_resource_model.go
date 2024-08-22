// Copyright © 2024. Citrix Systems, Inc.

package hypervisor

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HypervisorResourceModel maps the resource schema data.
type AzureHypervisorResourceModel struct {
	/**** Connection Details ****/
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Zone   types.String `tfsdk:"zone"`
	Scopes types.Set    `tfsdk:"scopes"` // Set[string]
	/** Azure Connection **/
	ApplicationId                   types.String `tfsdk:"application_id"`
	ApplicationSecret               types.String `tfsdk:"application_secret"`
	ApplicationSecretExpirationDate types.String `tfsdk:"application_secret_expiration_date"`
	SubscriptionId                  types.String `tfsdk:"subscription_id"`
	ActiveDirectoryId               types.String `tfsdk:"active_directory_id"`
	EnableAzureADDeviceManagement   types.Bool   `tfsdk:"enable_azure_ad_device_management"`
}

func (AzureHypervisorResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages an Azure hypervisor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the hypervisor.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the hypervisor.",
				Required:    true,
			},
			"zone": schema.StringAttribute{
				Description: "Id of the zone the hypervisor is associated with.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "Application ID of the service principal used to access the Azure APIs.",
				Required:    true,
			},
			"application_secret": schema.StringAttribute{
				Description: "The Application Secret of the service principal used to access the Azure APIs.",
				Required:    true,
				Sensitive:   true,
			},
			"application_secret_expiration_date": schema.StringAttribute{
				Description: "The expiration date of the application secret of the service principal used to access the Azure APIs. " +
					"\n\n-> **Note** Expiration date format is `YYYY-MM-DD`.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^((?:19|20|21)\d\d)[-](0[1-9]|1[012])[-](0[1-9]|[12][0-9]|3[01])$`), "ensure date is valid and is in the format YYYY-MM-DD"),
				},
			},
			"subscription_id": schema.StringAttribute{
				Description: "Azure Subscription ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"active_directory_id": schema.StringAttribute{
				Description: "Azure Active Directory ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"enable_azure_ad_device_management": schema.BoolAttribute{
				Description: "Enable Azure AD device management. Default is false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The IDs of the scopes for the hypervisor to be a part of.",
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

func (AzureHypervisorResourceModel) GetAttributes() map[string]schema.Attribute {
	return AzureHypervisorResourceModel{}.GetSchema().Attributes
}

func (r AzureHypervisorResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, hypervisor *citrixorchestration.HypervisorDetailResponseModel) AzureHypervisorResourceModel {
	r.Id = types.StringValue(hypervisor.GetId())
	r.Name = types.StringValue(hypervisor.GetName())
	hypZone := hypervisor.GetZone()
	r.Zone = types.StringValue(hypZone.GetId())
	r.ApplicationId = types.StringValue(hypervisor.GetApplicationId())
	r.SubscriptionId = types.StringValue(hypervisor.GetSubscriptionId())
	r.ActiveDirectoryId = types.StringValue(hypervisor.GetActiveDirectoryId())
	scopeIds := util.GetIdsForScopeObjects(hypervisor.GetScopes())
	r.Scopes = util.StringArrayToStringSet(ctx, diagnostics, scopeIds)

	customPropertiesString := hypervisor.GetCustomProperties()
	var customProperties []citrixorchestration.NameValueStringPairModel
	err := json.Unmarshal([]byte(customPropertiesString), &customProperties)
	if err != nil {
		return r
	}

	for _, customProperty := range customProperties {
		if customProperty.GetName() == "AzureAdDeviceManagement" {
			enabled, _ := strconv.ParseBool(customProperty.GetValue())
			r.EnableAzureADDeviceManagement = types.BoolValue(enabled)
		}
	}

	return r
}
