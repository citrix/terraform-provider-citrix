// Copyright © 2024. Citrix Systems, Inc.

package machine_catalog

import (
	"context"
	"regexp"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MachinePropertiesModel struct {
	Name             types.String `tfsdk:"name"`
	MachineCatalogId types.String `tfsdk:"machine_catalog_id"`
	Tags             types.Set    `tfsdk:"tags"`
}

func (MachinePropertiesModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages the properties of a machine.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The Name of the machine. For domain joined machines, the name must be in format <domain>\\<machine> format. Must be all in lowercase.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.LowerCaseRegex), "must be all in lowercase"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"machine_catalog_id": schema.StringAttribute{
				Description: "The ID of the machine catalog to which the machine belongs.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"tags": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tags to associate with the machine.",
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

func (MachinePropertiesModel) GetAttributes() map[string]schema.Attribute {
	return MachinePropertiesModel{}.GetSchema().Attributes
}

func (MachinePropertiesModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

func (r MachinePropertiesModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, machine *citrixorchestration.MachineDetailResponseModel, tagIds []string) MachinePropertiesModel {
	machineName := strings.ToLower(machine.GetName())
	r.Name = types.StringValue(machineName)

	machineCatalog := machine.GetMachineCatalog()
	r.MachineCatalogId = types.StringValue(machineCatalog.GetId())

	if len(tagIds) > 0 {
		r.Tags = util.StringArrayToStringSet(ctx, diagnostics, tagIds)
	} else {
		r.Tags = types.SetNull(types.StringType)
	}

	return r
}
