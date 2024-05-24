// Currently, only settings objects with value type of integer, boolean, strings and list of strings is supported.
resource "citrix_gac_settings" "test_settings_configuration" {
    service_url = "https://<your_service_url>:443"
    name = "test-settings"
    description = "Test settings configuration"
    app_settings = {
        windows = [
            {
                user_override = false,
                category = "ICA Client",
                settings = [
                    {
                        name = "Allow Client Clipboard Redirection",
                        value_string = "true"
                    }
                ]
            },
            {
                user_override = false,
                category = "Browser",
                settings = [
                    {
                        name = "delete browsing data on exit",
                        value_list = [
                            "browsing_history",
                            "download_history"
                        ]
                    },
                    {
                        name = "relaunch notification period",
                        value_string = "3600000"
                    }
                ]
            }
        ],
        html5 = [
            {
                category = "Virtual Channel",
                user_override = false,
                settings = [
                    {
                        name = "Clipboard Operations Between VDA And Local Device",
                        value_string = "true"
                    }  
                ]
            }
        ],
        ios = [
            {
                category = "Audio",
                user_override = false,
                settings = [
                    {
                        name = "audio",
                        value_string = "true"
                    }  
                ]
            }
        ],
        macos = [
            {
                category = "ica client",
                user_override = false,
                settings = [
                    {
                        name = "Reconnect Apps and Desktops",
                        value_list = [
                            "startWorkspace",
                            "refreshApps"
                        ]
                    }
                ]
            }
        ]
    }
}