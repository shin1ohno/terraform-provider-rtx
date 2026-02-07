# DynDNS-style DDNS provider
resource "rtx_ddns" "dyndns" {
  server_id = 1
  url       = "https://members.dyndns.org/nic/update"
  hostname  = "myhost.dyndns.org"
  username  = "ddns-user"
  password  = "example!PASS123"
}

# No-IP DDNS provider
resource "rtx_ddns" "noip" {
  server_id = 2
  url       = "https://dynupdate.no-ip.com/nic/update"
  hostname  = "myhost.no-ip.org"
  username  = "noip-user"
  password  = "example!PASS123"
}
