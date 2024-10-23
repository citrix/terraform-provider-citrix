resource "citrix_desktop_icon" "example-desktop-icon" {
  raw_data                    = filebase64("path/to/desktopicon.ico")
}
# Use filebase64 to encode a file's content in base64 format.