resource "citrix_policy_setting" "universal_printing_optimization_defaults" {
    name = "UPDCompressionDefaults"
    policy_id = citrix_policy.example_policy.id
    use_default = false
    value = "ImageCompression=StandardQuality, HeavyweightCompression=False, ImageCaching=True, FontCaching=True, AllowNonAdminsToModify=False"
}
