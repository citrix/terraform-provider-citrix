resource "citrix_daas_delivery_group" "example-delivery-group" {
    name = "example-delivery-group"
    associated_machine_catalogs = [
        {
        machine_catalog = citrix_daas_machine_catalog.example-machine-catalog.id
        count = 1
        }
    ]
    autoscale_enabled = true 
    users = [
        "user@example.com",
    ]
    autoscale_settings = {
        disconnect_peak_idle_session_after_seconds = 3600
        log_off_peak_disconnected_session_after_seconds = 3600
        peak_log_off_action = "Nothing"
        power_time_schemes = [
            {
                days_of_week = [
                    "Monday",
                    "Tuesday",
                    "Wednesday",
                    "Thursday",
                    "Friday"
                ]
                display_name = "weekdays schedule"
                peak_time_ranges = [
                    "09:00-17:00"
                ]
                pool_size_schedules = [
                    {
                        "time_range": "00:00-00:00",
                        "pool_size": 1
                    }
                ],
                pool_using_percentage = false
            },
        ]
    }
}