resource "citrix_delivery_group" "example-delivery-group" {
    name = "example-delivery-group"
    associated_machine_catalogs = [
        {
            machine_catalog = citrix_machine_catalog.example-azure-mtsession.id
            machine_count = 1
        }
    ]
    desktops = [
        {
            published_name = "Example Desktop"
            description = "Desription for example desktop"
            restricted_access_users = {
                allow_list = [
                    "user1@example.com"
                ]
                block_list = [
                    "user2@example.com",
                ]
            }
            enabled = true
            enable_session_roaming = false
        }
        
    ] 
    autoscale_settings = {
        autoscale_enabled = true
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
            "user1@example.com"
        ]
        block_list = [
            "user2@example.com",
        ]
    }
    reboot_schedules = [
		{
			name = "example_reboot_schedule_weekly"
			reboot_schedule_enabled = true
			frequency = "Weekly"
			frequency_factor = 1
			days_in_week = [
				"Monday",
				"Tuesday",
				"Wednesday"
				]
			start_time = "12:12"
			start_date = "2024-05-25"
			reboot_duration_minutes = 0
			ignore_maintenance_mode = true
			natural_reboot_schedule = false
		},
		{
			name = "example_reboot_schedule_monthly"
			description = "example reboot schedule"
			reboot_schedule_enabled = true
			frequency = "Monthly"
			frequency_factor = 2
			week_in_month = "First"
			day_in_month = "Monday"
			start_time = "12:12"
			start_date = "2024-04-21"
			ignore_maintenance_mode = true
			reboot_duration_minutes = 120
			natural_reboot_schedule = false
			reboot_notification_to_users = {
				notification_duration_minutes = 15
				notification_message = "test message"
				notification_title = "test title"
				notification_repeat_every_5_minutes = true
			}
		}
	]
	
    policy_set_id = citrix_policy_set.example-policy-set.id
}