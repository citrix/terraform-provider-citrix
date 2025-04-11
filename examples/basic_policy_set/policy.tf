// Policy is depending on the policy set
resource "citrix_policy" "first_basic_policy" {
    policy_set_id   = citrix_policy_set_v2.basic_policy_set.id
    name            = "first basic policy"
    description     = "basic policy description"
    enabled         = true
}

resource "citrix_policy" "second_basic_policy" {
    policy_set_id   = citrix_policy_set_v2.basic_policy_set.id
    name            = "second basic policy"
    description     = "second basic policy description"
    enabled         = true
}

resource "citrix_policy" "third_basic_policy" {
    policy_set_id   = citrix_policy_set_v2.basic_policy_set.id
    name            = "third basic policy"
    description     = "third basic policy description"
    enabled         = true
}
