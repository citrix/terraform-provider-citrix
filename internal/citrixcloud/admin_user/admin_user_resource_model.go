// Copyright © 2024. Citrix Systems, Inc.

package cc_admin_user

import (
	"context"

	"github.com/citrix/citrix-daas-rest-go/ccadmins"

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
	UserId       types.String `tfsdk:"user_id"`
	UcOid        types.String `tfsdk:"ucoid"`
	AccessType   types.String `tfsdk:"access_type"`
	DisplayName  types.String `tfsdk:"display_name"`
	Email        types.String `tfsdk:"email"`
	FirstName    types.String `tfsdk:"first_name"`
	LastName     types.String `tfsdk:"last_name"`
	ProviderType types.String `tfsdk:"provider_type"`
	Type         types.String `tfsdk:"type"`
}

func (CCAdminUserResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "Citrix Cloud --- Manages an administrator user for cloud environment.",

		Attributes: map[string]schema.Attribute{
			"user_id": schema.StringAttribute{
				Description: "Id of the administrator.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ucoid": schema.StringAttribute{
				Description: "Universal claim organization identifier of the administrator.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_type": schema.StringAttribute{
				Description: "Access Type of the user. Currently, this attribute can only be set to `Full`",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(ccadmins.CITRIXCLOUDSERVICESADMINISTRATORSAPIMODELSADMINISTRATORACCESSTYPE_FULL),
					),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "Display name for the user.",
				Optional:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email of the user where the invitation link will be sent.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"first_name": schema.StringAttribute{
				Description: "First name of the user.",
				Optional:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "Last name of the user.",
				Optional:    true,
			},
			"provider_type": schema.StringAttribute{
				Description: "Identity provider for the administrator or group you want to add. Currently, this attribute can only be set to `CitrixSts`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(ccadmins.CITRIXCLOUDSERVICESADMINISTRATORSAPIMODELSADMINISTRATORPROVIDERTYPE_CITRIX_STS),
					),
				},
			},
			"type": schema.StringAttribute{
				Description: "Type of administrator being added. Currently, this attribute can only be set to `AdministratorUser`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(ccadmins.CITRIXCLOUDSERVICESADMINISTRATORSAPIMODELSADMINISTRATORTYPE_ADMINISTRATOR_USER),
					),
				},
			},
		},
	}
}

func (CCAdminUserResourceModel) GetAttributes() map[string]schema.Attribute {
	return CCAdminUserResourceModel{}.GetSchema().Attributes
}

func (r CCAdminUserResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, adminUser *ccadmins.CitrixCloudServicesAdministratorsApiModelsAdministratorResult) CCAdminUserResourceModel {

	r.UserId = types.StringValue(adminUser.GetUserId())
	r.UcOid = types.StringValue(adminUser.GetUcOid())
	r.AccessType = types.StringValue(string(adminUser.GetAccessType()))
	r.Email = types.StringValue(adminUser.GetEmail())
	r.ProviderType = types.StringValue(string(adminUser.GetProviderType()))
	r.Type = types.StringValue(string(adminUser.GetType()))

	if !r.DisplayName.IsNull() {
		r.DisplayName = types.StringValue(adminUser.GetDisplayName())
	}

	if !r.FirstName.IsNull() {
		r.FirstName = types.StringValue(adminUser.GetFirstName())
	}
	if !r.LastName.IsNull() {
		r.LastName = types.StringValue(adminUser.GetLastName())
	}

	return r
}
