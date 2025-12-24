// Copyright © 2025. Citrix Systems, Inc.

package util

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MachineDomainIdentityModel struct {
	Domain                 types.String `tfsdk:"domain"`
	Ou                     types.String `tfsdk:"domain_ou"`
	ServiceAccountDomain   types.String `tfsdk:"service_account_domain"`
	ServiceAccount         types.String `tfsdk:"service_account"`
	ServiceAccountPassword types.String `tfsdk:"service_account_password"`
	ServiceAccountId       types.String `tfsdk:"service_account_id"`
}

func (MachineDomainIdentityModel) GetSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "The domain identity for machines in the machine catalog." + "<br />" +
			"Required when identity_type is set to `ActiveDirectory`",
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The AD domain where machine accounts will be created. Specify this in FQDN format; for example, MyDomain.com.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(DomainFqdnRegex), "must be in FQDN format"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_ou": schema.StringAttribute{
				Description: "The organization unit that computer accounts will be created into.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"service_account_domain": schema.StringAttribute{
				Description: "The domain name of the service account. Specify this in FQDN format; for example, MyServiceDomain.com." +
					"\n\n~> **Please Note** Use this property if domain of the service account which is used to create the machine accounts resides in a domain different from what's specified in property `domain` where the machine accounts are created.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(DomainFqdnRegex), "must be in FQDN format"),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("service_account"),
					}...),
				},
			},
			"service_account": schema.StringAttribute{
				Description: "Service account for the domain. Only the username is required; do not include the domain name.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(NoPathRegex), "must not include domain name, only specify the username"),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("service_account_password"),
					}...),
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRelative().AtParent().AtName("service_account_id"),
					}...),
				},
			},
			"service_account_password": schema.StringAttribute{
				Description: "Service account password for the domain.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("service_account"),
					}...),
				},
			},
			"service_account_id": schema.StringAttribute{
				Description: "The service account Id to be used for managing the machine accounts.",
				Optional:    true,
			},
		},
	}
}

func (MachineDomainIdentityModel) GetAttributes() map[string]schema.Attribute {
	return MachineDomainIdentityModel{}.GetSchema().Attributes
}

func (MachineDomainIdentityModel) GetCmaSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "The domain identity for creating machines in the domain-joined Citrix Managed Azure catalog. Only required when the machines in catalog are domain-joined",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The AD domain where machine accounts will be created. Specify this in FQDN format; for example, MyDomain.com.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(DomainFqdnRegex), "must be in FQDN format"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_ou": schema.StringAttribute{
				Description: "The organization unit that computer accounts will be created into.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"service_account_domain": schema.StringAttribute{
				Description: "The domain name of the service account if it is in a different domain from where the machines resides. **This is not yet supported in Citrix Managed Azure Catalogs.**",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(DomainFqdnRegex), "must be in FQDN format"),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("service_account"),
					}...),
				},
			},
			"service_account": schema.StringAttribute{
				Description: "Service account for the domain. Only the username is required; do not include the domain name.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(NoPathRegex), "must not include domain name, only specify the username"),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("service_account_password"),
					}...),
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRelative().AtParent().AtName("service_account_id"),
					}...),
				},
			},
			"service_account_password": schema.StringAttribute{
				Description: "Service account password for the domain.",
				Required:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRelative().AtParent().AtName("service_account"),
					}...),
				},
			},
			"service_account_id": schema.StringAttribute{
				Description: "The service account Id to be used for managing the machine accounts. **This is not yet supported in Citrix Managed Azure Catalogs.**",
				Optional:    true,
			},
		},
	}
}

func (MachineDomainIdentityModel) GetCmaAttributes() map[string]schema.Attribute {
	return MachineDomainIdentityModel{}.GetCmaSchema().Attributes
}
