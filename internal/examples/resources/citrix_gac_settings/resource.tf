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
            },
            {
                user_override = false,
                category = "dazzle",
                settings = [
                    {
                        name = "Local App Whitelist",
                        local_app_allow_list = [
                            {
                                arguments = "www.citrix.com",
                                name = "Google Chrome",
                                path = "C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe"
                            },
                            {
                                arguments = "www.citrix2.com",
                                name = "Google Chrome2",
                                path = "C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe"
                            },
                        ]
                    }
                ]
            },
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
            },
             {
                user_override = false,
                category = "Browser",
                settings = [
                    {
                        name = "managed bookmarks",
                        managed_bookmarks = [
                            {
                                name = "Citrix",
                                url = "https://www.citrix.com/"
                            },
                            {
                                name = "Citrix Workspace app",
                                url = "https://www.citrix.com/products/receiver.html"
                            }
                        ]
                    }
                ]
            }
        ],
        linux = [
            {
                category = "root",
                user_override = false,
                settings = [
                    {
                        name = "enable fido2",
                        value_string = "true"
                    }
                ]
            }
        ]
    }
}