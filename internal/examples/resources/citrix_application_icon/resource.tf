resource "citrix_application_icon" "example-application-icon" {
  raw_data                    = "example-raw-data"
}

# You can use the following PowerShell commands to convert an .ico file to base64:
# $pic = Get-Content 'fileName.ico' -Encoding Byte
# $picBase64 = [System.Convert]::ToBase64String($pic)
