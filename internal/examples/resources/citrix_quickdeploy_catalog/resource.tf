
# Quick Deploy Non-Domain-Joined Managed Catalog with Default Power Schedule
resource citrix_quickdeploy_catalog default-power-schedule-catalog {
    name = "example-quickdeploy-catalog"
    catalog_type = "MultiSession"
    region = "East US"
    subscription_name = "Citrix Managed"
    template_image_id = "<Template Image ID>"
    machine_size = "d2asv5"
    storage_type = "StandardSSD_LRS"
    number_of_users = 2
    max_users_per_vm = 4
    power_schedule = {}
}

# Quick Deploy Non-Domain-Joined Managed Catalog with custom Power Schedule and custom Machine Naming Scheme
resource citrix_quickdeploy_catalog custom-power-schedule-catalog {
    name = "example-quickdeploy-catalog-custom-schedule"
    catalog_type = "MultiSession"
    region = "East US"
    subscription_name = "Citrix Managed"
    template_image_id = "<Template Image ID>"
    machine_size = "d2asv5"
    storage_type = "StandardSSD_LRS"
    number_of_users = 4
    max_users_per_vm = 4
    machine_naming_scheme = {
        naming_scheme = "example-vda-#"
        naming_scheme_type = "Numeric"
    }
    power_schedule = {
        peak_min_instances = 2
        off_peak_min_instances = 1
        weekdays = ["monday", "tuesday", "wednesday", "thursday", "friday"]
        peak_start_time = 9
        peak_end_time = 18
        peak_time_zone_id = "Pacific Standard Time"
        peak_off_delay = 15
    }
}

# Quick Deploy Domain-Joined Managed Catalog with Default Power Schedule
data "citrix_quickdeploy_onprem_network_connection" "example_network_connection" {
    name = "example-network-connection-name"
}

resource citrix_quickdeploy_catalog domain-joined-catalog {
    name = "example-quickdeploy-catalog"
    catalog_type = "MultiSession"
    region = "East US"
    subscription_name = "Citrix Managed"
    template_image_id = "<Template Image ID>"
    machine_size = "d2asv5"
    storage_type = "StandardSSD_LRS"
    number_of_users = 2
    max_users_per_vm = 4
    power_schedule = {}
    on_prem_connectivity = {
        onprem_network_connection_id = data.citrix_quickdeploy_onprem_network_connection.example_network_connection.id
        domain_identity = {
            domain                   = "acme.net"
            service_account          = "admin"
            service_account_password = "<service account password>"
        }
    }
}