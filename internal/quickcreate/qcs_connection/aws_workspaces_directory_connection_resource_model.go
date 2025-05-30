package qcs_connection

import (
	"context"
	"regexp"

	quickcreateservice "github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AwsWorkspacesDirectoryConnectionModel struct {
	DirectoryConnectionId           types.String `tfsdk:"id"`
	AccountId                       types.String `tfsdk:"account"`
	Name                            types.String `tfsdk:"name"`
	ZoneId                          types.String `tfsdk:"zone"`
	ResourceLocationId              types.String `tfsdk:"resource_location"`
	Directory                       types.String `tfsdk:"directory"`
	Subnets                         types.Set    `tfsdk:"subnets"`
	Tenancy                         types.String `tfsdk:"tenancy"`
	UserEnabledAsLocalAdministrator types.Bool   `tfsdk:"user_enabled_as_local_administrator"`
	SecurityGroup                   types.String `tfsdk:"security_group"`
	DefaultOu                       types.String `tfsdk:"default_ou"`
}

func (AwsWorkspacesDirectoryConnectionModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "DaaS Quick Deploy - AWS WorkSpaces Core --- Manages an AWS WorkSpaces directory connection.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the directory connection.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the directory connection.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"account": schema.StringAttribute{
				Description: "ID of the account the directory connection is associated with.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zone": schema.StringAttribute{
				Description: "ID of the zone the directory connection is associated with. Only one of `zone` and `resource_location` attributes can be specified.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("resource_location"),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_location": schema.StringAttribute{
				Description: "ID of the resource location the directory connection is associated with. Only one of `resource_location` and `zone` attributes can be specified.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"directory": schema.StringAttribute{
				Description: "ID of the AWS directory the directory connection is associated with.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.AwsDirectoryId), "The Directory Id provided contains invalid characters or is not in the correct format."),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnets": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "IDs of the subnets the directory connection is associated with.",
				Required:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(regexp.MustCompile(util.AwsSubnetIdFormat), "The Subnet ID provided contains invalid characters or is not in the correct format."),
					),
					setvalidator.SizeBetween(2, 2),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"tenancy": schema.StringAttribute{
				Description: "Tenancy of the directory connection. Possible values are `SHARED` and `DEDICATED`. Defaults to `DEDICATED`.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(quickcreateservice.AWSEDCDIRECTORYTENANCY_SHARED),
						string(quickcreateservice.AWSEDCDIRECTORYTENANCY_DEDICATED),
					),
				},
				Default: stringdefault.StaticString(string(quickcreateservice.AWSEDCDIRECTORYTENANCY_DEDICATED)),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_enabled_as_local_administrator": schema.BoolAttribute{
				Description: "Enable users to be local administrators. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"security_group": schema.StringAttribute{
				Description: "ID of the security group the directory connection is associated with.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.AwsSecurityGroupId), "The Security Group ID provided contains invalid characters or is not in the correct format."),
				},
			},
			"default_ou": schema.StringAttribute{
				Description: "Default OU for VDAs in the directory connection.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.OuPathFormat), "The organizational unit path provided contains invalid characters or is not in the correct format."),
				},
			},
		},
	}
}

func (AwsWorkspacesDirectoryConnectionModel) GetAttributes() map[string]schema.Attribute {
	return AwsWorkspacesDirectoryConnectionModel{}.GetSchema().Attributes
}

func (AwsWorkspacesDirectoryConnectionModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

func (r AwsWorkspacesDirectoryConnectionModel) RefreshPropertyValues(ctx context.Context, diagnostics *diag.Diagnostics, isResource bool, directory *quickcreateservice.AwsEdcDirectoryConnection) AwsWorkspacesDirectoryConnectionModel {
	r.DirectoryConnectionId = types.StringValue(directory.GetConnectionId())
	r.Name = types.StringValue(directory.GetName())
	r.AccountId = types.StringValue(directory.GetAccountId())
	if !r.ZoneId.IsNull() || !isResource {
		r.ZoneId = types.StringValue(directory.GetZoneId())
	}
	if !r.ResourceLocationId.IsNull() || !isResource {
		r.ResourceLocationId = types.StringValue(directory.GetResourceLocationId())
	}
	r.Directory = types.StringValue(directory.GetDirectoryId())
	r.Subnets = util.StringArrayToStringSet(ctx, diagnostics, []string{directory.GetSubnet1Id(), directory.GetSubnet2Id()})

	r.Tenancy = types.StringValue(string(directory.GetTenancy()))
	r.UserEnabledAsLocalAdministrator = types.BoolValue(directory.GetUserEnabledAsLocalAdministrator())
	r.SecurityGroup = types.StringValue(directory.GetSecurityGroupId())
	r.DefaultOu = types.StringValue(directory.GetDefaultOU())

	return r
}
