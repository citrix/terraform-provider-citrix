# Only allow access to applications tagged with the tag from connections through the corporate gateway
resource "citrix_resource_access_policy" "example_tag_and_type_policy" {
    name          = "Example Resource Access Policy"
    description   = "Restrict tagged applications to the corporate gateway"
    tag_id        = citrix_tag.example_tag.id
    resource_type = "Application"

    enable_criteria_for_include_connections = true
    include_connections_criteria_type       = "MatchAny"
    include_criteria_filters = [
        {
            filter_name  = "CorpGateway"
            filter_value = "CorpVPN"
        }
    ]
}

# Deny access to all desktops from connections matching the untrusted network filter
resource "citrix_resource_access_policy" "example_type_policy" {
    name          = "Example Desktop Resource Access Policy"
    resource_type = "Desktop"

    enable_criteria_for_exclude_connections = true
    exclude_criteria_filters = [
        {
            filter_name  = "CorpGateway"
            filter_value = "UntrustedNetwork"
        }
    ]
}
