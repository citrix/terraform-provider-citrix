resource citrix_autoscale_plugin_template example-template {
    name       = "<template-name>"
    type       = "Holiday"
    dates = [
        "2025-12-25",
        "2026-01-01"
    ]
}