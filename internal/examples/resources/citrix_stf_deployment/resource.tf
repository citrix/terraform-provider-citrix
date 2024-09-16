resource "citrix_stf_deployment" "example-stf-deployment" {
	site_id      = "1"	
	host_base_url = "https://<storefront machine hostname>"
	roaming_gateway = [
		{
			name = "Example Roaming Gateway Name"
			logon_type = "None"
			gateway_url = "https://example.gateway.url/"
			subnet_ip_address = "10.0.0.1"
		}
	]
	roaming_beacon = {
		internal_ip = "https://example.internalip.url"
		external_ips = ["https://example.externalip.url"]
	}
}