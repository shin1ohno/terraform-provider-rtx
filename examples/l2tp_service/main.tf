# L2TP service configuration for RTX router
#
# This is a singleton resource - only one can exist per router.
# It controls the global L2TP service state.

# Enable L2TP service with both L2TPv3 and L2TPv2 protocols
resource "rtx_l2tp_service" "main" {
  enabled   = true
  protocols = ["l2tpv3", "l2tp"]
}

# Example: Enable only L2TPv3 for site-to-site VPN
# resource "rtx_l2tp_service" "l2tpv3_only" {
#   enabled   = true
#   protocols = ["l2tpv3"]
# }

# Example: Enable only L2TPv2 for remote access VPN
# resource "rtx_l2tp_service" "l2tp_only" {
#   enabled   = true
#   protocols = ["l2tp"]
# }
