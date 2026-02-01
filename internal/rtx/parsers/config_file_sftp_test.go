package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sftpTestConfig contains actual RTX configuration from SFTP read (sensitive data masked)
// This tests holistic parsing of a real-world configuration
const sftpTestConfig = `#
# Admin
#

login password TEST_LOGIN_PASS
administrator password TEST_ADMIN_PASS
login user testuser testpass
user attribute administrator=off connection=off gui-page=dashboard,lan-map,config login-timer=300
user attribute testuser connection=serial,telnet,remote,ssh,sftp,http gui-page=dashboard,lan-map,config login-timer=3600
timezone +09:00
console character ja.utf8
console lines infinity
console prompt "[RTX1210] "
system packet-buffer small max-buffer=5000 max-free=1300
system packet-buffer middle max-buffer=10000 max-free=4950
system packet-buffer large max-buffer=20000 max-free=5600
switch control use bridge1 on terminal=on
lan-map snapshot use bridge1 on terminal=wired-only

httpd host any
httpd proxy-access l2ms permit on

operation http revision-up permit on
description yno RTX1210-Test
sshd service on
sshd host lan2 bridge1
sshd host key generate 2869 TEST_SSH_KEY
sftpd host bridge1
external-memory statistics filename prefix usb1:rtx1210
statistics traffic on
statistics nat on
switch control watch interval 2 5
lan-map terminal watch interval 1800 10
lan-map sysname RTX1210_Test

syslog host 192.168.1.20
syslog local address 192.168.1.253
syslog facility local0
syslog notice on
syslog info on
syslog debug on

#
# WAN connection
#

description lan2 wan
ip lan2 address dhcp
ip lan2 nat descriptor 1000
ip lan2 secure filter in  200020 200021 200022 200023 200024 200025 200103 200100 200102 200104 200101 200105 200099
ip lan2 secure filter out 200020 200021 200022 200023 200024 200025 200026 200027 200099 dynamic 200080 200081 200082 200083 200084 200085

ipv6 lan2 mtu 1500
ipv6 lan2 dhcp service client ir=on
ipv6 lan2 secure filter in 101000 101002 101099
ipv6 lan2 secure filter out 101099 dynamic 101080 101081 101082 101083 101084 101085 101098 101099

ngn type lan2 off
sip use off

#
# IP configuration
#

ip routing process fast
ip route change log on
ip filter source-route on
ip filter directed-broadcast on
ip route default gateway dhcp lan2
ip route 10.33.128.0/21 gateway 192.168.1.20 gateway 192.168.1.21
ip route 100.64.0.0/10 gateway 192.168.1.20 gateway 192.168.1.21

ipv6 routing process fast
ipv6 prefix 1 ra-prefix@lan2::/64
ipv6 bridge1 address ra-prefix@lan2::1/64
ipv6 lan1 address ra-prefix@lan2::2/64
ipv6 lan1 rtadv send 1 o_flag=on
ipv6 lan1 dhcp service server

bridge member bridge1 lan1 tunnel1
ip bridge1 address 192.168.1.253/16
ip lan1 proxyarp on

#
# Services
#

dhcp service server
dhcp server rfc2131 compliant except remain-silent
dhcp scope 1 192.168.1.20-192.168.1.99/16 gateway 192.168.1.253 expire 12:00 maxexpire 24:00
dhcp scope bind 1 192.168.1.20 01 00 30 93 11 0e 33
dhcp scope bind 1 192.168.1.21 01 00 3e e1 c3 54 b4
dhcp scope bind 1 192.168.1.22 01 00 3e e1 c3 54 b5
dhcp scope bind 1 192.168.1.28 24:59:e5:54:5e:5a
dhcp scope bind 1 192.168.1.29 b8:0b:da:ef:77:33
dhcp scope option 1 dns=1.1.1.1,1.0.0.1
dhcp scope option 1 router=192.168.1.253

dhcp client release linkdown on
dns host bridge1
dns service recursive
dns cache use off
dns server select 1 100.100.100.100 edns=on any home.local
dns server select 500000 1.1.1.1 edns=on 1.0.0.1 edns=on a .
dns server select 500100 2606:4700:4700::1111 edns=on 2606:4700:4700::1001 edns=on aaaa .
dns private address spoof on

#
# Tunnels
#

pp disable all
pp select anonymous
 pp bind tunnel2
 pp auth request mschap-v2
 pp auth username testuser testpass
 ppp ipcp ipaddress on
 ppp ipcp msext on
 ppp ccp type none
 ip pp remote address pool dhcp
 ip pp mtu 1258
 pp enable anonymous
no tunnel enable all
tunnel select 1
 tunnel encapsulation l2tpv3
 tunnel endpoint name test.example.com fqdn
 ipsec tunnel 101
  ipsec sa policy 101 1 esp aes-cbc sha-hmac
  ipsec ike keepalive log 1 off
  ipsec ike keepalive use 1 on heartbeat 10 6
  ipsec ike local address 1 192.168.1.253
  ipsec ike log 1 key-info message-info payload-info
  ipsec ike nat-traversal 1 on
  ipsec ike pre-shared-key 1 text TEST_PSK
  ipsec ike remote address 1 test.example.com
  ipsec ike remote name 1 test-id key-id
 l2tp always-on on
 l2tp hostname test-RTX1210
 l2tp tunnel auth on TEST_AUTH
 l2tp tunnel disconnect time off
 l2tp keepalive use on 60 3
 l2tp keepalive log off
 l2tp syslog on
 l2tp local router-id 192.168.1.253
 l2tp remote router-id 192.168.1.254
 l2tp remote end-id testuser
 ip tunnel secure filter in 200028 200099
 ip tunnel tcp mss limit auto
 tunnel enable 1

 tunnel select 2
 tunnel encapsulation l2tp
 ipsec tunnel 1
  ipsec sa policy 1 2 esp aes-cbc sha-hmac
  ipsec ike keepalive use 2 off
  ipsec ike nat-traversal 2 on
  ipsec ike pre-shared-key 2 text TEST_PSK
  ipsec ike remote address 2 any
 l2tp tunnel disconnect time off
 ip tunnel tcp mss limit auto
 tunnel enable 2

ip filter 200000 reject 10.0.0.0/8 * * * *
ip filter 200001 reject 172.16.0.0/12 * * * *
ip filter 200002 reject 192.168.0.0/16 * * * *
ip filter 200010 reject * 10.0.0.0/8 * * *
ip filter 200011 reject * 172.16.0.0/12 * * *
ip filter 200012 reject * 192.168.0.0/16 * * *
ip filter 200020 reject * * udp,tcp 135 *
ip filter 200021 reject * * udp,tcp * 135
ip filter 200022 reject * * udp,tcp netbios_ns-netbios_ssn *
ip filter 200023 reject * * udp,tcp * netbios_ns-netbios_ssn
ip filter 200024 reject * * udp,tcp 445 *
ip filter 200025 reject * * udp,tcp * 445
ip filter 200026 restrict * * tcpfin * www,21,nntp
ip filter 200027 restrict * * tcprst * www,21,nntp
ip filter 200028 reject * * udp dhcps,dhcpc dhcps,dhcpc
ip filter 200099 pass * * * * *
ip filter 200100 pass * * udp * 500
ip filter 200101 pass * * udp * 1701
ip filter 200102 pass * * udp * 4500
ip filter 200103 reject * * udp dhcps,dhcpc dhcps,dhcpc
ip filter 200104 pass * * tcp * www
ip filter 200105 pass * * esp
ip filter 500000 restrict * * * * *
ip filter dynamic 200080 * * ftp syslog=off
ip filter dynamic 200081 * * domain syslog=off
ip filter dynamic 200082 * * www syslog=off
ip filter dynamic 200083 * * smtp syslog=off
ip filter dynamic 200084 * * pop3 syslog=off
ip filter dynamic 200085 * * submission syslog=off
ip filter dynamic 200098 * * tcp syslog=off
ip filter dynamic 200099 * * udp syslog=off

#
# Ethernet filters
#

# default filter
ethernet filter 100 pass *:*:*:*:*:* *:*:*:*:*:*

# no internet access from living TV
ethernet filter 1 reject-nolog bc:5c:17:05:59:3a *:*:*:*:*:*
ethernet filter 2 reject-nolog *:*:*:*:*:* bc:5c:17:05:59:3a

# apply the filters
ethernet lan1 filter in 1 100
ethernet lan1 filter out 2 100

nat descriptor type 1000 masquerade
nat descriptor address outer 1000 primary
nat descriptor masquerade incoming 1000 reject
nat descriptor masquerade static 1000 1 192.168.1.253 esp
nat descriptor masquerade static 1000 2 192.168.1.253 udp 500
nat descriptor masquerade static 1000 3 192.168.1.253 udp 4500
nat descriptor masquerade static 1000 4 192.168.1.253 udp 1701
nat descriptor masquerade static 1000 900 192.168.1.20 tcp 55000

ipsec auto refresh on
ipsec transport 1 101 udp 1701
ipsec transport 3 3 udp 1701
l2tp service on l2tpv3 l2tp

ipv6 filter 101000 pass * * icmp6 * *
ipv6 filter 101002 pass * * udp * 546
ipv6 filter 101098 reject * * * * *
ipv6 filter 101099 pass * * * * *
ipv6 filter 402100 pass * * * * *
ipv6 filter dynamic 101080 * * ftp syslog=off
ipv6 filter dynamic 101081 * * domain syslog=off
ipv6 filter dynamic 101082 * * www syslog=off
ipv6 filter dynamic 101083 * * smtp syslog=off
ipv6 filter dynamic 101084 * * pop3 syslog=off
ipv6 filter dynamic 101085 * * submission syslog=off
ipv6 filter dynamic 101098 * * tcp syslog=off
ipv6 filter dynamic 101099 * * udp syslog=off
`

// TestConfigFileParser_SFTPConfig_L2TPTunnel1 tests L2TPv3 tunnel parsing (REQ-1)
func TestConfigFileParser_SFTPConfig_L2TPTunnel1(t *testing.T) {
	parser := NewConfigFileParser()
	result, err := parser.Parse(sftpTestConfig)
	require.NoError(t, err)

	tunnels := result.ExtractL2TPTunnels()
	require.NotEmpty(t, tunnels, "should extract L2TP tunnels")

	// Find tunnel 1
	var tunnel1 *L2TPConfig
	for i := range tunnels {
		if tunnels[i].ID == 1 {
			tunnel1 = &tunnels[i]
			break
		}
	}
	require.NotNil(t, tunnel1, "tunnel 1 should exist")

	t.Run("AlwaysOn", func(t *testing.T) {
		assert.True(t, tunnel1.AlwaysOn, "l2tp always-on on should set AlwaysOn=true")
	})

	t.Run("KeepaliveEnabled", func(t *testing.T) {
		assert.True(t, tunnel1.KeepaliveEnabled, "l2tp keepalive use on should set KeepaliveEnabled=true")
	})

	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "test-RTX1210", tunnel1.Name, "l2tp hostname should be extracted")
	})

	t.Run("Version", func(t *testing.T) {
		assert.Equal(t, "l2tpv3", tunnel1.Version, "tunnel encapsulation l2tpv3 should set Version=l2tpv3")
	})

	t.Run("Mode", func(t *testing.T) {
		assert.Equal(t, "l2vpn", tunnel1.Mode, "L2TPv3 should have Mode=l2vpn")
	})
}

// TestConfigFileParser_SFTPConfig_L2TPTunnel2 tests L2TPv2 tunnel parsing (REQ-2)
func TestConfigFileParser_SFTPConfig_L2TPTunnel2(t *testing.T) {
	parser := NewConfigFileParser()
	result, err := parser.Parse(sftpTestConfig)
	require.NoError(t, err)

	tunnels := result.ExtractL2TPTunnels()

	// Find tunnel 2
	var tunnel2 *L2TPConfig
	for i := range tunnels {
		if tunnels[i].ID == 2 {
			tunnel2 = &tunnels[i]
			break
		}
	}
	require.NotNil(t, tunnel2, "tunnel 2 should exist")

	t.Run("Version", func(t *testing.T) {
		assert.Equal(t, "l2tp", tunnel2.Version, "tunnel encapsulation l2tp should set Version=l2tp")
	})

	t.Run("Mode", func(t *testing.T) {
		assert.Equal(t, "lns", tunnel2.Mode, "L2TPv2 should have Mode=lns (not default l2vpn)")
	})
}

// TestConfigFileParser_SFTPConfig_L2TPService tests L2TP service parsing (REQ-3)
func TestConfigFileParser_SFTPConfig_L2TPService(t *testing.T) {
	parser := NewConfigFileParser()
	result, err := parser.Parse(sftpTestConfig)
	require.NoError(t, err)

	service := result.ExtractL2TPService()
	require.NotNil(t, service, "L2TP service should be extracted")

	t.Run("Enabled", func(t *testing.T) {
		assert.True(t, service.Enabled, "l2tp service on should set Enabled=true")
	})

	t.Run("Protocols", func(t *testing.T) {
		assert.Equal(t, []string{"l2tpv3", "l2tp"}, service.Protocols,
			"l2tp service on l2tpv3 l2tp should extract both protocols")
	})
}

// TestConfigFileParser_SFTPConfig_System tests system config parsing (REQ-4)
func TestConfigFileParser_SFTPConfig_System(t *testing.T) {
	parser := NewConfigFileParser()
	result, err := parser.Parse(sftpTestConfig)
	require.NoError(t, err)

	system := result.ExtractSystem()
	require.NotNil(t, system, "System config should be extracted")

	t.Run("Timezone", func(t *testing.T) {
		assert.Equal(t, "+09:00", system.Timezone, "timezone should be extracted")
	})

	t.Run("Console_Character", func(t *testing.T) {
		require.NotNil(t, system.Console, "Console config should exist")
		assert.Equal(t, "ja.utf8", system.Console.Character, "console character should be extracted")
	})

	t.Run("Console_Lines", func(t *testing.T) {
		require.NotNil(t, system.Console, "Console config should exist")
		assert.Equal(t, "infinity", system.Console.Lines, "console lines should be extracted")
	})

	t.Run("Console_Prompt", func(t *testing.T) {
		require.NotNil(t, system.Console, "Console config should exist")
		assert.Equal(t, "[RTX1210] ", system.Console.Prompt, "console prompt should be extracted")
	})

	t.Run("PacketBuffers", func(t *testing.T) {
		require.NotNil(t, system.PacketBuffers, "PacketBuffers should exist")
		assert.Len(t, system.PacketBuffers, 3, "should have 3 packet buffer configs")

		// Verify each buffer type exists
		var hasSmall, hasMiddle, hasLarge bool
		for _, pb := range system.PacketBuffers {
			switch pb.Size {
			case "small":
				hasSmall = true
				assert.Equal(t, 5000, pb.MaxBuffer)
				assert.Equal(t, 1300, pb.MaxFree)
			case "middle":
				hasMiddle = true
				assert.Equal(t, 10000, pb.MaxBuffer)
				assert.Equal(t, 4950, pb.MaxFree)
			case "large":
				hasLarge = true
				assert.Equal(t, 20000, pb.MaxBuffer)
				assert.Equal(t, 5600, pb.MaxFree)
			}
		}
		assert.True(t, hasSmall, "should have small packet buffer")
		assert.True(t, hasMiddle, "should have middle packet buffer")
		assert.True(t, hasLarge, "should have large packet buffer")
	})

	t.Run("Statistics_Traffic", func(t *testing.T) {
		require.NotNil(t, system.Statistics, "Statistics config should exist")
		assert.True(t, system.Statistics.Traffic, "statistics traffic on should be parsed")
	})

	t.Run("Statistics_NAT", func(t *testing.T) {
		require.NotNil(t, system.Statistics, "Statistics config should exist")
		assert.True(t, system.Statistics.NAT, "statistics nat on should be parsed")
	})
}

// TestConfigFileParser_SFTPConfig_DHCPBindings tests DHCP binding parsing (REQ-5)
func TestConfigFileParser_SFTPConfig_DHCPBindings(t *testing.T) {
	parser := NewConfigFileParser()
	result, err := parser.Parse(sftpTestConfig)
	require.NoError(t, err)

	bindings := result.ExtractDHCPBindings()
	require.Len(t, bindings, 5, "should have 5 DHCP bindings")

	t.Run("AllBindings_ScopeID", func(t *testing.T) {
		for i, b := range bindings {
			assert.Equal(t, 1, b.ScopeID, "binding %d should have ScopeID=1", i)
		}
	})

	t.Run("IPAddresses", func(t *testing.T) {
		ips := make([]string, len(bindings))
		for i, b := range bindings {
			ips[i] = b.IPAddress
		}

		assert.Contains(t, ips, "192.168.1.20", "should contain IP 192.168.1.20")
		assert.Contains(t, ips, "192.168.1.21", "should contain IP 192.168.1.21")
		assert.Contains(t, ips, "192.168.1.22", "should contain IP 192.168.1.22")
		assert.Contains(t, ips, "192.168.1.28", "should contain IP 192.168.1.28")
		assert.Contains(t, ips, "192.168.1.29", "should contain IP 192.168.1.29")
	})
}
