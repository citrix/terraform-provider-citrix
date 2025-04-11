// Policy Filters are depending on the `citrix_policy` resource.
// Since the `citrix_policy` resource depends on `citrix_policy_set_v2` resource, the policy filter resources have an implicit dependency on the `citrix_policy_set_v2` resource.
resource "citrix_access_control_policy_filter" "access_control_policy_filter" {
    policy_id       = citrix_policy.first_basic_policy.id
    enabled         = true
    allowed         = true
    connection_type = "WithAccessGateway"
    condition       = "*"
    gateway         = "*"
}

resource "citrix_branch_repeater_policy_filter" "branch_repeater_filter" {
    policy_id   = citrix_policy.first_basic_policy.id
    allowed     = true
}

resource "citrix_client_ip_policy_filter" "client_ip_filter" {
    policy_id   = citrix_policy.first_basic_policy.id
    enabled     = true
    allowed     = true
    ip_address  = var.policy_filter_client_ip
}

resource "citrix_client_name_policy_filter" "client_name_filter" {
    policy_id    = citrix_policy.first_basic_policy.id
    enabled      = true
    allowed      = true
    client_name  = var.policy_filter_client_name
}

resource "citrix_client_platform_policy_filter" "client_platform_filter" {
    policy_id   = citrix_policy.first_basic_policy.id
    enabled     = true
    allowed     = true
    platform    = var.policy_filter_client_platform
}

resource "citrix_delivery_group_policy_filter" "delivery_group_filter" {
    policy_id          = citrix_policy.first_basic_policy.id
    enabled            = true
    allowed            = true
    delivery_group_id  = var.policy_filter_delivery_group_id
}

resource "citrix_delivery_group_type_policy_filter" "delivery_group_type_filter" {
    policy_id    	  	= citrix_policy.first_basic_policy.id
    enabled      	  	= true
    allowed      	  	= true
    delivery_group_type = var.policy_filter_delivery_group_type
}

resource "citrix_ou_policy_filter" "ou_filter" {
    policy_id   = citrix_policy.first_basic_policy.id
    enabled     = true
    allowed     = true
    ou	 		= var.policy_filter_ou
}

resource "citrix_tag_policy_filter" "tag_filter" {
    policy_id   = citrix_policy.first_basic_policy.id
    enabled     = true
    allowed     = true
    tag 		= var.policy_filter_tag_id
}

resource "citrix_user_policy_filter" "user_filter" {
    policy_id   = citrix_policy.first_basic_policy.id
    enabled     = true
    allowed     = true
    sid 		= var.policy_filter_user_sid
}