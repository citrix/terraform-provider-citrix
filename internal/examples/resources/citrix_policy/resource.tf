resource "citrix_policy" "example_policy" {
    policy_set_id   = citrix_policy_set_v2.example_policy_set_v2.id
    name            = "example_policy"
    description     = "example policy description"
    enabled         = true
}