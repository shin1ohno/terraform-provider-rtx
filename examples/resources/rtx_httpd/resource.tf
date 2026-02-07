# Allow HTTP management from any interface
resource "rtx_httpd" "web_management" {
  host = "any"
}

# Restrict HTTP management to LAN only
# resource "rtx_httpd" "secure_web" {
#   host         = "lan1"
#   proxy_access = true
# }
