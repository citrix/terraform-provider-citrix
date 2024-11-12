// Copyright © 2024. Citrix Systems, Inc.

package policies

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (AccessControlFilterModel) GetDataSourceSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the policy filter.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Computed:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Computed:    true,
			},
			"connection": schema.StringAttribute{
				Description: "Gateway connection for the policy filter.",
				Computed:    true,
			},
			"condition": schema.StringAttribute{
				Description: "Gateway condition for the policy filter.",
				Computed:    true,
			},
			"gateway": schema.StringAttribute{
				Description: "Gateway for the policy filter.",
				Computed:    true,
			},
		},
	}
}

func (AccessControlFilterModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return AccessControlFilterModel{}.GetDataSourceSchema().Attributes
}
func (BranchRepeaterFilterModel) GetDataSourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Definition of branch repeater policy filter.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the branch repeater policy filter.",
				Computed:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Computed:    true,
			},
		},
	}
}

func (BranchRepeaterFilterModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return BranchRepeaterFilterModel{}.GetDataSourceSchema().Attributes
}

func (ClientIPFilterModel) GetDataSourceSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the client ip policy filter.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Computed:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Computed:    true,
			},
			"ip_address": schema.StringAttribute{
				Description: "IP Address of the client to be filtered.",
				Computed:    true,
			},
		},
	}
}

func (ClientIPFilterModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return ClientIPFilterModel{}.GetDataSourceSchema().Attributes
}

func (ClientNameFilterModel) GetDataSourceSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the client name policy filter.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Computed:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Computed:    true,
			},
			"client_name": schema.StringAttribute{
				Description: "Name of the client to be filtered.",
				Computed:    true,
			},
		},
	}
}

func (ClientNameFilterModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return ClientNameFilterModel{}.GetDataSourceSchema().Attributes
}

func (DeliveryGroupFilterModel) GetDataSourceSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the delivery group policy filter.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Computed:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Computed:    true,
			},
			"delivery_group_id": schema.StringAttribute{
				Description: "Id of the delivery group to be filtered.",
				Computed:    true,
			},
		},
	}
}

func (DeliveryGroupFilterModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return DeliveryGroupFilterModel{}.GetDataSourceSchema().Attributes
}

func (DeliveryGroupTypeFilterModel) GetDataSourceSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the delivery group type policy filter.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Computed:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Computed:    true,
			},
			"delivery_group_type": schema.StringAttribute{
				Description: "Type of the delivery groups to be filtered.",
				Computed:    true,
			},
		},
	}
}

func (DeliveryGroupTypeFilterModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return DeliveryGroupTypeFilterModel{}.GetDataSourceSchema().Attributes
}

func (OuFilterModel) GetDataSourceSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the organizational unit policy filter.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Computed:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Computed:    true,
			},
			"ou": schema.StringAttribute{
				Description: "Organizational Unit to be filtered.",
				Computed:    true,
			},
		},
	}
}

func (OuFilterModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return OuFilterModel{}.GetDataSourceSchema().Attributes
}

func (UserFilterModel) GetDataSourceSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the user policy filter.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Computed:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Computed:    true,
			},
			"sid": schema.StringAttribute{
				Description: "SID of the user or user group to be filtered.",
				Computed:    true,
			},
		},
	}
}

func (UserFilterModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return UserFilterModel{}.GetDataSourceSchema().Attributes
}

func (TagFilterModel) GetDataSourceSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the tag policy filter.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicate whether the filter is being enabled.",
				Computed:    true,
			},
			"allowed": schema.BoolAttribute{
				Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
				Computed:    true,
			},
			"tag": schema.StringAttribute{
				Description: "Tag to be filtered.",
				Computed:    true,
			},
		},
	}
}

func (TagFilterModel) GetDataSourceAttributes() map[string]schema.Attribute {
	return TagFilterModel{}.GetDataSourceSchema().Attributes
}
