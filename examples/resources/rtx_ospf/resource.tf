# OSPF configuration with multiple areas
resource "rtx_ospf" "main" {
  process_id = 1
  router_id  = "10.0.0.1"
  distance   = 110

  # Backbone area
  area {
    area_id = "0.0.0.0"
    type    = "normal"
  }

  # Stub area
  area {
    area_id    = "0.0.0.1"
    type       = "stub"
    no_summary = false
  }

  # Networks in backbone area
  network {
    ip       = "10.0.0.0"
    wildcard = "0.0.0.255"
    area     = "0.0.0.0"
  }

  network {
    ip       = "192.168.1.0"
    wildcard = "0.0.0.255"
    area     = "0.0.0.1"
  }

  redistribute_static           = true
  redistribute_connected        = true
  default_information_originate = true
}
