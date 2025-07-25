
# Quick Deploy Managed Catalog with Default Power Schedule
resource citrix_quickdeploy_catalog test-1 {
    name = "example-quickdeploy-catalog"
    catalog_type = "MultiSession"
    region = "East US"
    subscription_name = "Citrix Managed"
    template_image_id = "<Template Image ID>"
    machine_size = "d2asv5"
    storage_type = "StandardSSD_LRS"
    number_of_machines = 2
    max_users_per_vm = 4
    power_schedule = {}
}

# Quick Deploy Managed Catalog with custom Power Schedule and custom Machine Naming Scheme
resource citrix_quickdeploy_catalog test-1 {
    name = "example-quickdeploy-catalog-custom-schedule"
    catalog_type = "MultiSession"
    region = "East US"
    subscription_name = "Citrix Managed"
    template_image_id = "<Template Image ID>"
    machine_size = "d2asv5"
    storage_type = "StandardSSD_LRS"
    number_of_machines = 4
    max_users_per_vm = 4
    machine_naming_scheme = {
        naming_scheme = "example-vda-#"
        naming_scheme_type = "Numeric"
    }
    power_schedule = {
        peak_buffer_capacity = 30
        off_peak_buffer_capacity = 15
        peak_min_instances = 2
        off_peak_min_instances = 1
        max_users_per_vm = 4
        weekdays = ["monday", "tuesday", "wednesday", "thursday", "friday"]
        peak_start_time = 9
        peak_end_time = 18
        peak_time_zone_id = "Pacific Standard Time"
        peak_off_delay = 15
    }
}