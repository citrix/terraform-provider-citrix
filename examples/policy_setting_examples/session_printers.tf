resource "citrix_policy_setting" "session_printers" {
    name = "SessionPrinters"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = jsonencode([
        "\\\\printserver\\printer1,model=HP LaserJet,location=Building A",
        "\\\\printserver\\printer2,model=Canon Pixma,location=Building B"
    ])
}
