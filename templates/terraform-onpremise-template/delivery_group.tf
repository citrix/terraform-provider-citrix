resource "citrix_daas_delivery_group" "example-delivery-group" {
    name = "example-delivery-group"
    associated_machine_catalogs = [
        {
            machine_catalog = citrix_daas_machine_catalog.example-catalog.id
            count = 1
        }
    ]
    autoscale_enabled = true
    users = [
        "<domain user email>",
    ]
    autoscale_settings = {
        power_time_schemes = [
            {
                days_of_week = [
                    "Monday",
                    "Tuesday",
                    "Wednesday",
                    "Thursday",
                    "Friday"
                ]
                name = "weekdays test"
                display_name = "weekdays schedule"
                peak_time_ranges = [
                    "09:00-17:00"
                ]
                pool_size_schedules = [
                    {
                        time_range = "00:00-00:00",
                        pool_size = 1
                    }
                ]
                pool_using_percentage = false
            },
        ]
    }
}