resource "citrix_gac_discovery" "example-gac-discovery" {
    domain = "example-domain.com"
    service_urls = ["https://example.com:443", "https://example2.com:80"]
    allowed_web_store_urls = ["https://example.com", "https://example2.com"]
}