# TODO

- **ipv6_interface: destroy leaves `ipv6 <iface> address` lines on the device**
  Observed 2026-08-03 on RTX1210 Rev.14.01.42 (home-monitor #117/#119): destroying an
  `rtx_ipv6_interface` removed the rtadv / dhcp-service lines but left
  `ipv6 lan1 address ra-prefix@lan2::2/64` and `ipv6 bridge1 address ra-prefix@lan2::1/64`
  in the running config. Terraform state dropped the resource, so the leftovers became
  invisible drift that needed a manual admin-session cleanup.
  First step: acceptance test that reads back `show config` after destroy, then make the
  delete path issue `no ipv6 <iface> address <value>` for every configured address block.

- **ipv6_interface: schema accepts `rtadv` / `dhcpv6_service` on bridge interfaces the firmware rejects**
  `ipv6 bridge1 rtadv send 1 o_flag=on m_flag=off` fails on-device with コマンド名エラー —
  per the official IPv6 command reference, `rtadv send` / `dhcp service` take LAN
  interfaces only, on ALL current models (rt-common applies to RTX3510/1300/1220/1210/840/830/vRX).
  The shared interface regex `^(lan|bridge|pp|tunnel)\d+$` lets bridgeN through validation,
  so the failure only surfaces mid-apply (caused the home-monitor #118 revert).
  First step: restrict rtadv / dhcpv6_service to lanN (plus pp where documented) via schema
  validators, with a test asserting bridgeN is rejected at plan time.
