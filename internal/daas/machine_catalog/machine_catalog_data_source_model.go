// Copyright © 2024. Citrix Systems, Inc.

package machine_catalog

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

// MachineCatalogDataSourceModel defines the Machine Catalog data source implementation.
type MachineCatalogDataSourceModel struct {
	Id                       types.String   `tfsdk:"id"`
	Name                     types.String   `tfsdk:"name"`
	MachineCatalogFolderPath types.String   `tfsdk:"machine_catalog_folder_path"`
	Vdas                     []vda.VdaModel `tfsdk:"vdas"`    // List[VdaModel]
	Tenants                  types.Set      `tfsdk:"tenants"` // Set[String]
	Tags                     types.Set      `tfsdk:"tags"`    // Set[string]
}

func (MachineCatalogDataSourceModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Read data of an existing machine catalog.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the machine catalog.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("name")), // Ensures that only one of either Id or Name is provided. It will also cause a validation error if none are specified.
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "Id must be a valid GUID"),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the machine catalog.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"machine_catalog_folder_path": schema.StringAttribute{
				Description: "The path to the folder in which the machine catalog is located.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.AdminFolderPathWithBackslashRegex), "Admin Folder Path must not start or end with a backslash"),
					stringvalidator.RegexMatches(regexp.MustCompile(util.AdminFolderPathSpecialCharactersRegex), "Admin Folder Path must not contain any of the following special characters: / ; : # . * ? = < > | [ ] ( ) { } \" ' ` ~ "),
					stringvalidator.AlsoRequires(path.MatchRoot("name")),
				},
			},
			"vdas": schema.ListNestedAttribute{
				Description:  "The VDAs associated with the machine catalog.",
				Computed:     true,
				NestedObject: vda.VdaModel{}.GetSchema(),
			},
			"tenants": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tenants to associate with the machine catalog.",
				Computed:    true,
			},
			"tags": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "A set of identifiers of tags to associate with the machine catalog.",
				Computed:    true,
			},
		},
	}
}

func (r MachineCatalogDataSourceModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, catalog *citrixorchestration.MachineCatalogDetailResponseModel, vdas *citrixorchestration.MachineResponseModelCollection, tags []string) MachineCatalogDataSourceModel {
	r.Id = types.StringValue(catalog.GetId())
	r.Name = types.StringValue(catalog.GetName())

	adminFolder := catalog.GetAdminFolder()
	adminFolderPath := strings.TrimSuffix(adminFolder.GetName(), "\\")
	if adminFolderPath != "" {
		r.MachineCatalogFolderPath = types.StringValue(adminFolderPath)
	} else {
		r.MachineCatalogFolderPath = types.StringNull()
	}

	res := []vda.VdaModel{}
	for _, model := range vdas.GetItems() {
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

	r.Tenants = util.RefreshTenantSet(ctx, diagnostics, catalog.GetTenants())
	r.Tags = util.StringArrayToStringSet(ctx, diagnostics, tags)

	return r
}
