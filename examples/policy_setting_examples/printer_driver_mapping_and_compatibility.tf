resource "citrix_policy_setting" "printer_driver_mapping_and_compatibility" {
    name = "PrinterDriverMappings"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "Microsoft XPS Document Writer,Deny,Deny",
        "Send to Microsoft OneNote *,UPD_Only",
        "Printer Driver,Replace=Printer Driver Replaced",
        "Printer Driver Allowed,Allow"
    ])
}
