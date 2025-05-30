// Copyright © 2024. Citrix Systems, Inc.
package cma_image

import (
	"context"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixquickdeploy"
	catalogservice "github.com/citrix/citrix-daas-rest-go/citrixquickdeploy"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CitrixManagedAzureImageResourceModel struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Notes             types.String `tfsdk:"notes"`
	SubscriptionName  types.String `tfsdk:"subscription_name"`
	Region            types.String `tfsdk:"region"`
	VhdUri            types.String `tfsdk:"vhd_uri"`
	MachineGeneration types.String `tfsdk:"machine_generation"`
	OsPlatform        types.String `tfsdk:"os_platform"`
	VtpmEnabled       types.Bool   `tfsdk:"vtpm_enabled"`
	SecureBootEnabled types.Bool   `tfsdk:"secure_boot_enabled"`
	GuestDiskUri      types.String `tfsdk:"guest_disk_uri"`
}

func (CitrixManagedAzureImageResourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "DaaS Quick Deploy - Citrix Managed Azure --- Manages an Citrix Managed Azure image.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the image.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the image.",
				Required:    true,
			},
			"notes": schema.StringAttribute{
				Description: "Notes of the image. Note length cannot exceed 1024 characters.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1024),
				},
			},
			"subscription_name": schema.StringAttribute{
				Description: "The name of the Citrix Managed Azure subscription to import the image. Defaults to `Citrix Managed` if omitted.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("Citrix Managed"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Description: "The Azure region to import the image.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vhd_uri": schema.StringAttribute{
				Description: "The Azure-generated URL for the Image's Virtual Hard Disk.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"machine_generation": schema.StringAttribute{
				Description: "The generation of the virtual machine this image will be used to create. Choose between `V1` and `V2`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(util.HypervGen1),
						string(util.HypervGen2),
					),
				},
			},
			"os_platform": schema.StringAttribute{
				Description: "The OS platform of the image. Choose between `Windows` and `Linux`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(util.OSPlatform_Windows),
						string(util.OSPlatform_Linux),
					),
				},
			},
			"vtpm_enabled": schema.BoolAttribute{
				Description: "Defines whether the image supports vTPM TrustedLaunch. Only applicable for V2 generation images. Default is `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"secure_boot_enabled": schema.BoolAttribute{
				Description: "Defines whether the image supports Secure Boot. Only applicable for V2 generation images. Default is `false`." +
					"\n\n~> **Please Note** When using Secure Boot, the guest disk URI must be specified in the `guest_disk_uri` attribute.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"guest_disk_uri": schema.StringAttribute{
				Description: "The Azure-generated URL for the Guest Disk for the encrypted Template Image VHD. Only applicable for V2 generation images. Required only when `secure_boot_enabled` is set to `true`.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (CitrixManagedAzureImageResourceModel) GetAttributes() map[string]schema.Attribute {
	return CitrixManagedAzureImageResourceModel{}.GetSchema().Attributes
}

func (CitrixManagedAzureImageResourceModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

func (r CitrixManagedAzureImageResourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, isResource bool, image *catalogservice.TemplateImageDetails, region *catalogservice.DeploymentRegionModel) CitrixManagedAzureImageResourceModel {
	r.Id = types.StringValue(image.GetId())
	r.Name = types.StringValue(image.GetName())
	r.Notes = types.StringValue(image.GetNotes())
	r.SubscriptionName = types.StringValue(image.GetSubscriptionName())
	if region == nil || r.shouldSetRegion(*region) {
		// Set region only if region is not set in state, or inconsistent with plan
		// If region is nil, it will not be consistent with plan
		r.Region = types.StringValue(image.GetRegion())
	}
	r.MachineGeneration = types.StringValue(image.GetHyperVGen())
	r.OsPlatform = types.StringValue(string(image.GetOsPlatform()))
	r.VtpmEnabled = types.BoolValue(image.GetVtpmEnabled())
	r.SecureBootEnabled = types.BoolValue(image.GetSecureBootEnabled())

	return r
}

func (r CitrixManagedAzureImageResourceModel) shouldSetRegion(region citrixquickdeploy.DeploymentRegionModel) bool {
	// Always store name in state for the first time, but allow either if already specified in state or plan
	return r.Region.ValueString() == "" ||
		(!strings.EqualFold(r.Region.ValueString(), region.GetName()) && !strings.EqualFold(r.Region.ValueString(), region.GetId()))
}
