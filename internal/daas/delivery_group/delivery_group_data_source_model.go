// Copyright © 2024. Citrix Systems, Inc.

package delivery_group

import (
	"context"
	"regexp"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/terraform-provider-citrix/internal/daas/vda"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DeliveryGroupDataSourceModel defines the Delivery Group data source implementation.
type DeliveryGroupDataSourceModel struct {
	Id                      types.String   `tfsdk:"id"`
	Name                    types.String   `tfsdk:"name"`
	DeliveryType            types.String   `tfsdk:"delivery_type"`
	DeliveryGroupFolderPath types.String   `tfsdk:"delivery_group_folder_path"`
	InMaintenanceMode       types.Bool     `tfsdk:"in_maintenance_mode"`
	Vdas                    []vda.VdaModel `tfsdk:"vdas"`    // List[VdaModel]
	Tenants                 types.Set      `tfsdk:"tenants"` // Set[string]
	Tags                    types.Set      `tfsdk:"tags"`    // Set[string]
	SecureIcaRequired       types.Bool     `tfsdk:"secure_ica_required"`
}

func (DeliveryGroupDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Read data of an existing delivery group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the delivery group.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")), // Ensures that only one of either Id or Name is provided. It will also cause a validation error if none are specified.
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "Id must be a valid GUID"),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the delivery group.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"delivery_type": schema.StringAttribute{
				Description: "The delivery type of the delivery group.",
				Computed:    true,
			},
			"delivery_group_folder_path": schema.StringAttribute{
				Description: "The path to the folder in which the delivery group is located.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.AdminFolderPathWithBackslashRegex), "Admin Folder Path must not start or end with a backslash"),
					stringvalidator.RegexMatches(regexp.MustCompile(util.AdminFolderPathSpecialCharactersRegex), "Admin Folder Path must not contain any of the following special characters: / ; : # . * ? = < > | [ ] ( ) { } \" ' ` ~ "),
					stringvalidator.AlsoRequires(path.MatchRoot("name")),
				},
			},
			"in_maintenance_mode": schema.BoolAttribute{
				Description: "Indicates whether the delivery group is in maintenance mode.",
				Computed:    true,
			},
			"vdas": schema.ListNestedAttribute{
				Description:  "The VDAs associated with the delivery group.",
				Computed:     true,
				NestedObject: vda.VdaModel{}.GetSchema(),
			},
			"tenants": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tenants to associate with the delivery group.",
				Computed:    true,
			},
			"tags": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tags to associate with the delivery group.",
				Computed:    true,
			},
			"secure_ica_required": schema.BoolAttribute{
				Description: "Indicates whether secure ICA is required for the delivery group.",
				Computed:    true,
			},
		},
	}
}

func (r DeliveryGroupDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, deliveryGroup *citrixorchestration.DeliveryGroupDetailResponseModel, vdas []citrixorchestration.MachineResponseModel, tags []string) DeliveryGroupDataSourceModel {
	r.Id = types.StringValue(deliveryGroup.GetId())
	r.Name = types.StringValue(deliveryGroup.GetName())

	adminFolder := deliveryGroup.GetAdminFolder()
	adminFolderPath := strings.TrimSuffix(adminFolder.GetName(), "\\")
	if adminFolderPath != "" {
		r.DeliveryGroupFolderPath = types.StringValue(adminFolderPath)
	} else {
		r.DeliveryGroupFolderPath = types.StringNull()
	}

	deliveryType := string(deliveryGroup.GetDeliveryType())
	r.DeliveryType = types.StringValue(deliveryType)
	r.InMaintenanceMode = types.BoolValue(deliveryGroup.GetInMaintenanceMode())

	res := []vda.VdaModel{}
	for _, model := range vdas {
		machineName := model.GetName()
		hosting := model.GetHosting()
		hostedMachineId := hosting.GetHostedMachineId()
		machineCatalog := model.GetMachineCatalog()
		machineCatalogId := machineCatalog.GetId()
		deliveryGroup := model.GetDeliveryGroup()
		deliveryGroupId := deliveryGroup.GetId()

		res = append(res, vda.VdaModel{
			Id:                       types.StringValue(model.GetId()),
			MachineName:              types.StringValue(machineName),
			HostedMachineId:          types.StringValue(hostedMachineId),
			AssociatedMachineCatalog: types.StringValue(machineCatalogId),
			AssociatedDeliveryGroup:  types.StringValue(deliveryGroupId),
		})
	}

	r.Vdas = res

	r.Tenants = util.RefreshTenantSet(ctx, diagnostics, deliveryGroup.GetTenants())
	r.Tags = util.RefreshTagSet(ctx, diagnostics, tags)

	r.SecureIcaRequired = types.BoolValue(deliveryGroup.GetSecureIcaRequired())

	return r
}
