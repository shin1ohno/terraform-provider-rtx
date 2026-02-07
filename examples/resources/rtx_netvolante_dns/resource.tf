# NetVolante DNS (Yamaha's free DDNS service)
resource "rtx_netvolante_dns" "primary" {
  interface    = "pp 1"
  hostname     = "myrouter.aa0.netvolante.jp"
  server       = 1
  timeout      = 60
  ipv6_enabled = false
}

# NetVolante DNS on server 2 for redundancy
resource "rtx_netvolante_dns" "backup" {
  interface    = "pp 1"
  hostname     = "backup.aa0.netvolante.jp"
  server       = 2
  timeout      = 60
  ipv6_enabled = false
}
