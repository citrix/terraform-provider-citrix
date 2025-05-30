// Copyright © 2024. Citrix Systems, Inc.

package image_definition

import (
	"context"
	"regexp"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AzureImageDefinitionModel struct {
	ResourceGroup    types.String `tfsdk:"resource_group"` // Optional, If not specified, create new resource group, cannot use existing image gallery
	UseImageGallery  types.Bool   `tfsdk:"use_image_gallery"`
	ImageGalleryName types.String `tfsdk:"image_gallery_name"`
}

func (AzureImageDefinitionModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Details of the Azure Image Definition.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"resource_group": schema.StringAttribute{
				Description: "Existing resource group to store the image definition. If not specified, a new resource group will be created.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"use_image_gallery": schema.BoolAttribute{
				Description: "Whether image gallery is used to store the image definition. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"image_gallery_name": schema.StringAttribute{
				Description: "Name of the existing image gallery. If not specified and `use_image_gallery` is `true`, a new image gallery will be created.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
	}
}

func (AzureImageDefinitionModel) GetAttributes() map[string]schema.Attribute {
	return AzureImageDefinitionModel{}.GetSchema().Attributes
}

type ImageDefinitionModel struct {
	Id                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Description               types.String `tfsdk:"description"`
	OsType                    types.String `tfsdk:"os_type"`
	SessionSupport            types.String `tfsdk:"session_support"`
	Hypervisor                types.String `tfsdk:"hypervisor"`
	HypervisorResourcePool    types.String `tfsdk:"hypervisor_resource_pool"`
	AzureImageDefinitionModel types.Object `tfsdk:"azure_image_definition"`
	LatestVersion             types.Int64  `tfsdk:"latest_version"`
	Timeout                   types.Object `tfsdk:"timeout"`
}

func (ImageDefinitionModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages an image definition. **Note that this feature is in Tech Preview.**",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The GUID identifier of the image definition.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the image definition.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the image definition.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"os_type": schema.StringAttribute{
				Description: "Operating system type of the image definition. Valid values are `Windows` and `Linux`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixorchestration.OSTYPE_WINDOWS),
						string(citrixorchestration.OSTYPE_LINUX),
					),
				},
			},
			"session_support": schema.StringAttribute{
				Description: "Session support of the image definition. Valid values are `MultiSession` and `SingleSession`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(citrixorchestration.SESSIONSUPPORT_MULTI_SESSION),
						string(citrixorchestration.SESSIONSUPPORT_SINGLE_SESSION),
					),
				},
			},
			"hypervisor": schema.StringAttribute{
				Description: "ID of the hypervisor connection to be used for image definition.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			// The resource pool isn't actually used by the image definition, but the definition internally will use any of the pools.
			// This attribute ensures there is at least 1 pool as required by the API.
			"hypervisor_resource_pool": schema.StringAttribute{
				Description: "ID of the hypervisor resource pool to be used for image definition.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"azure_image_definition": AzureImageDefinitionModel{}.GetSchema(),
			"latest_version": schema.Int64Attribute{
				Description: "Latest version of the image definition.",
				Computed:    true,
			},
			"timeout": ImageDefinitionTimeout{}.GetSchema(),
		},
	}
}

func (ImageDefinitionModel) GetAttributes() map[string]schema.Attribute {
	return ImageDefinitionModel{}.GetSchema().Attributes
}

func (ImageDefinitionModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

type ImageDefinitionTimeout struct {
	Create types.Int32 `tfsdk:"create"`
	Delete types.Int32 `tfsdk:"delete"`
}

func getImageDefinitionTimeoutConfigs() util.TimeoutConfigs {
	return util.TimeoutConfigs{
		Create:        true,
		CreateDefault: 10,
		CreateMin:     5,

		Delete:        true,
		DeleteDefault: 10,
		DeleteMin:     5,
	}
}

func (ImageDefinitionTimeout) GetSchema() schema.SingleNestedAttribute {
	return util.GetTimeoutSchema("image definition", getImageDefinitionTimeoutConfigs())
}

func (ImageDefinitionTimeout) GetAttributes() map[string]schema.Attribute {
	return ImageDefinitionTimeout{}.GetSchema().Attributes
}

func (r ImageDefinitionModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, isResource bool, imageDefinition *citrixorchestration.ImageDefinitionResponseModel) ImageDefinitionModel {
	r.Id = types.StringValue(imageDefinition.GetId())
	r.Name = types.StringValue(imageDefinition.GetName())
	r.Description = types.StringValue(imageDefinition.GetDescription())
	r.OsType = types.StringValue(util.OrchestrationOSTypeEnumToString(imageDefinition.GetOsType()))
	r.SessionSupport = types.StringValue(util.SessionSupportEnumToString(imageDefinition.GetVDASessionSupport()))

	connections := imageDefinition.GetHypervisorConnections()
	var azureImgDefAttrMap map[string]attr.Type
	var err error

	// Set AzureImageDefinitionModel to default null value
	if isResource {
		azureImgDefAttrMap, err = util.ResourceAttributeMapFromObject(AzureImageDefinitionModel{})
	} else {
		azureImgDefAttrMap, err = util.DataSourceAttributeMapFromObject(AzureImageDefinitionModel{})
	}
	if err != nil {
		diagnostics.AddWarning("Error converting schema to attribute map. Error: ", err.Error())
		return r
	}
	r.AzureImageDefinitionModel = types.ObjectNull(azureImgDefAttrMap)

	// Set Hypervisor to default null value
	r.Hypervisor = types.StringNull()
	// r.HypervisorResourcePool is not updated, it is only used to create a dependency relationship.

	if len(connections) > 0 {
		connection := connections[0]
		customProperties := connection.GetCustomProperties()
		switch connection.GetPluginFactoryName() {
		case util.AZURERM_FACTORY_NAME:
			model := AzureImageDefinitionModel{
				ResourceGroup:    types.StringNull(),
				UseImageGallery:  types.BoolNull(),
				ImageGalleryName: types.StringNull(),
			}
			for _, property := range customProperties {
				if property.GetName() == "ResourceGroups" {
					model.ResourceGroup = types.StringValue(property.GetValue())
				} else if property.GetName() == "UseSharedImageGallery" {
					model.UseImageGallery = util.StringToTypeBool(property.GetValue())
				} else if property.GetName() == "ImageGallery" {
					model.ImageGalleryName = types.StringValue(property.GetValue())
				}
			}
			// Override default null value for AzureImageDefinitionModel
			r.AzureImageDefinitionModel = util.TypedObjectToObjectValue(ctx, diagnostics, model)
		}
		// Override default null value for Hypervisor
		r.Hypervisor = types.StringValue(connection.GetId())
	}

	r.LatestVersion = types.Int64Value(int64(imageDefinition.GetLatestVersion()))

	return r
}
