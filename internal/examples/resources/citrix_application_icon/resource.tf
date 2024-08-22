resource "citrix_application_icon" "example-application-icon" {
  raw_data                    = filebase64("path/to/icon.ico")
}
# Use filebase64 to encode a file's content in base64 format.