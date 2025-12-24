resource "citrix_policy_setting" "multi_stream_virtual_channel_stream_assignment" {
    name = "MultiStreamAssignment"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = "CTXCAM,0;CTXCTL,1;CTXCLIP,2"
}
