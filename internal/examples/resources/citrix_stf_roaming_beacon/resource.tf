resource "citrix_stf_roaming_beacon" "testSTFRoamingBeacon" {
  internal_ip = "https://example.internal.url/"
  external_ips = ["https://abc1.com/" , "https://abc2.com/"]
  site_id = 1
}