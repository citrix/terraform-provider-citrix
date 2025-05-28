# Quick Deploy Template Image with VHD URI
resource citrix_quickdeploy_template_image test_image {
    name = "example-template-image"
    notes = "Example Windows Gen 1 template image imported to US East region in the Citrix Managed Azure Subscription via VHD URI"
    subscription_name = "Citrix Managed"
    region = "East US"
    vhd_uri = "<Image VHD URI>"
    machine_generation = "V1"
    os_platform = "Windows"
}

# Quick Deploy Gen 2 Template Image with vTPM and Secure Boot enabled
resource citrix_quickdeploy_template_image test_image {
    name = "example-template-image"
    notes = "Example Windows Gen 2 template image with vTPM and Secure Boot enabled"
    subscription_name = "Citrix Managed"
    region = "East US"
    vhd_uri = "<Encrypted Image VHD URI>"
    machine_generation = "V2"
    os_platform = "Windows"
    vtpm_enabled = true
    secure_boot_enabled = true
    guest_disk_uri = "<Guest Disk VHD URI>"
}