resource "citrix_delivery_group" "example-delivery-group" {
    name = "example-delivery-group"
    minimum_functional_level    = "L7_20"
    associated_machine_catalogs = [
        {
            machine_catalog = citrix_machine_catalog.example-aws-catalog.id
            machine_count = 1
        }
    ]
    desktops = [
        {
            published_name = "Example Desktop"
            description = "Desription for example desktop"
            restricted_access_users = {
                allow_list = [
                    "example\\user1"
                ]
                block_list = [
                    "example\\user2",
                ]
            }
            enabled = true
            enable_session_roaming = false
        }
        
    ] 
    autoscale_settings = {
        autoscale_enabled = true
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
    restricted_access_users = {
        allow_list = [
            "example\\user1"
        ]
        block_list = [
            "example\\user2",
        ]
    }
}
