# Basic syslog configuration with single host
resource "rtx_syslog" "basic" {
  host {
    address = "192.168.1.100"
  }

  facility = "user"
  notice   = true
  info     = false
  debug    = false
}

# Full syslog configuration with multiple hosts and all options
resource "rtx_syslog" "full" {
  host {
    address = "192.168.1.100"
  }

  host {
    address = "192.168.1.101"
    port    = 1514
  }

  host {
    address = "syslog.example.com"
    port    = 5514
  }

  local_address = "192.168.1.1"
  facility      = "local0"
  notice        = true
  info          = true
  debug         = true
}
