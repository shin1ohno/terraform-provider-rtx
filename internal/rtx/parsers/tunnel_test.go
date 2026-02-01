package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTunnelParser_ParseIPsecTunnel(t *testing.T) {
	config := `tunnel select 1
 ipsec tunnel 1
 ipsec sa policy 1 1 esp aes-cbc sha-hmac
 ipsec ike local address 1 192.168.1.1
 ipsec ike remote address 1 192.168.2.1
 ipsec ike pre-shared-key 1 text secret123
 ipsec ike keepalive use 1 on dpd 30 3
 tunnel enable 1
`

	parser := NewTunnelParser()
	tunnels, err := parser.ParseTunnelConfig(config)

	require.NoError(t, err)
	require.Len(t, tunnels, 1)

	tunnel := tunnels[0]
	assert.Equal(t, 1, tunnel.ID)
	assert.Equal(t, "ipsec", tunnel.Encapsulation)
	assert.True(t, tunnel.Enabled)

	require.NotNil(t, tunnel.IPsec)
	assert.Equal(t, 1, tunnel.IPsec.IPsecTunnelID)
	assert.Equal(t, "192.168.1.1", tunnel.IPsec.LocalAddress)
	assert.Equal(t, "192.168.2.1", tunnel.IPsec.RemoteAddress)
	assert.Equal(t, "secret123", tunnel.IPsec.PreSharedKey)

	require.NotNil(t, tunnel.IPsec.Keepalive)
	assert.True(t, tunnel.IPsec.Keepalive.Enabled)
	assert.Equal(t, "dpd", tunnel.IPsec.Keepalive.Mode)
	assert.Equal(t, 30, tunnel.IPsec.Keepalive.Interval)
	assert.Equal(t, 3, tunnel.IPsec.Keepalive.Retry)
}

func TestTunnelParser_ParseL2TPv3Tunnel(t *testing.T) {
	config := `tunnel select 1
 tunnel encapsulation l2tpv3
 ipsec tunnel 101
 ipsec sa policy 101 1 esp aes-cbc sha-hmac
 ipsec ike local address 101 192.168.1.253
 ipsec ike remote address 101 itm.ohno.be
 ipsec ike pre-shared-key 101 text secret123
 ipsec ike keepalive use 101 on heartbeat 10 6
 l2tp always-on on
 l2tp hostname ebisu-RTX1210
 l2tp tunnel auth on password123
 l2tp keepalive use on 60 3
 l2tp local router-id 192.168.1.253
 l2tp remote router-id 192.168.1.254
 l2tp remote end-id shin1
 ip tunnel secure filter in 200028 200099
 ip tunnel tcp mss limit auto
 tunnel enable 1
`

	parser := NewTunnelParser()
	tunnels, err := parser.ParseTunnelConfig(config)

	require.NoError(t, err)
	require.Len(t, tunnels, 1)

	tunnel := tunnels[0]
	assert.Equal(t, 1, tunnel.ID)
	assert.Equal(t, "l2tpv3", tunnel.Encapsulation)
	assert.True(t, tunnel.Enabled)

	// IPsec block
	require.NotNil(t, tunnel.IPsec)
	assert.Equal(t, 101, tunnel.IPsec.IPsecTunnelID)
	assert.Equal(t, "192.168.1.253", tunnel.IPsec.LocalAddress)
	assert.Equal(t, "itm.ohno.be", tunnel.IPsec.RemoteAddress)
	assert.Equal(t, "secret123", tunnel.IPsec.PreSharedKey)
	assert.Equal(t, []int{200028, 200099}, tunnel.IPsec.SecureFilterIn)
	assert.Equal(t, "auto", tunnel.IPsec.TCPMSSLimit)

	require.NotNil(t, tunnel.IPsec.Keepalive)
	assert.True(t, tunnel.IPsec.Keepalive.Enabled)
	assert.Equal(t, "heartbeat", tunnel.IPsec.Keepalive.Mode)
	assert.Equal(t, 10, tunnel.IPsec.Keepalive.Interval)
	assert.Equal(t, 6, tunnel.IPsec.Keepalive.Retry)

	// L2TP block
	require.NotNil(t, tunnel.L2TP)
	assert.Equal(t, "ebisu-RTX1210", tunnel.L2TP.Hostname)
	assert.True(t, tunnel.L2TP.AlwaysOn)
	assert.Equal(t, "192.168.1.253", tunnel.L2TP.LocalRouterID)
	assert.Equal(t, "192.168.1.254", tunnel.L2TP.RemoteRouterID)
	assert.Equal(t, "shin1", tunnel.L2TP.RemoteEndID)

	require.NotNil(t, tunnel.L2TP.TunnelAuth)
	assert.True(t, tunnel.L2TP.TunnelAuth.Enabled)
	assert.Equal(t, "password123", tunnel.L2TP.TunnelAuth.Password)

	require.NotNil(t, tunnel.L2TP.Keepalive)
	assert.True(t, tunnel.L2TP.Keepalive.Enabled)
	assert.Equal(t, 60, tunnel.L2TP.Keepalive.Interval)
	assert.Equal(t, 3, tunnel.L2TP.Keepalive.Retry)
}

func TestTunnelParser_ParseL2TPv2Tunnel(t *testing.T) {
	config := `tunnel select 1
 tunnel encapsulation l2tp
 ipsec tunnel 1
 ipsec sa policy 1 1 esp aes-cbc sha-hmac
 ipsec ike pre-shared-key 1 text secret123
 tunnel enable 1
pp select anonymous
 pp bind tunnel1
 pp auth accept chap
 pp auth request chap
 pp auth myname vpnuser password123
 ip pp remote address pool 192.168.100.100-192.168.100.200
`

	parser := NewTunnelParser()
	tunnels, err := parser.ParseTunnelConfig(config)

	require.NoError(t, err)
	require.Len(t, tunnels, 1)

	tunnel := tunnels[0]
	assert.Equal(t, 1, tunnel.ID)
	assert.Equal(t, "l2tp", tunnel.Encapsulation)
	assert.True(t, tunnel.Enabled)

	// IPsec block
	require.NotNil(t, tunnel.IPsec)
	assert.Equal(t, 1, tunnel.IPsec.IPsecTunnelID)
	assert.Equal(t, "secret123", tunnel.IPsec.PreSharedKey)

	// L2TP block (from anonymous PP)
	require.NotNil(t, tunnel.L2TP)
	require.NotNil(t, tunnel.L2TP.Authentication)
	assert.Equal(t, "chap", tunnel.L2TP.Authentication.Method)
	assert.Equal(t, "chap", tunnel.L2TP.Authentication.RequestMethod)
	assert.Equal(t, "vpnuser", tunnel.L2TP.Authentication.Username)
	assert.Equal(t, "password123", tunnel.L2TP.Authentication.Password)

	require.NotNil(t, tunnel.L2TP.IPPool)
	assert.Equal(t, "192.168.100.100", tunnel.L2TP.IPPool.Start)
	assert.Equal(t, "192.168.100.200", tunnel.L2TP.IPPool.End)
}

func TestBuildTunnelCommands_IPsec(t *testing.T) {
	tunnel := Tunnel{
		ID:            1,
		Encapsulation: "ipsec",
		Enabled:       true,
		IPsec: &TunnelIPsec{
			IPsecTunnelID: 1,
			LocalAddress:  "192.168.1.1",
			RemoteAddress: "192.168.2.1",
			PreSharedKey:  "secret123",
			IKEv2Proposal: IKEv2Proposal{
				EncryptionAES128: true,
				IntegritySHA1:    true,
				GroupFourteen:    true,
			},
			Transform: IPsecTransform{
				Protocol:         "esp",
				EncryptionAES128: true,
				IntegritySHA1:    true,
			},
			Keepalive: &TunnelIPsecKeepalive{
				Enabled:  true,
				Mode:     "dpd",
				Interval: 30,
				Retry:    3,
			},
		},
	}

	commands := BuildTunnelCommands(tunnel)

	assert.Contains(t, commands, "tunnel select 1")
	assert.Contains(t, commands, "ipsec tunnel 1")
	assert.Contains(t, commands, "ipsec ike local address 1 192.168.1.1")
	assert.Contains(t, commands, "ipsec ike remote address 1 192.168.2.1")
	assert.Contains(t, commands, "ipsec ike pre-shared-key 1 text secret123")
	assert.Contains(t, commands, "ipsec ike keepalive use 1 on dpd 30 3")
	assert.Contains(t, commands, "tunnel enable 1")
}

func TestBuildTunnelCommands_L2TPv3(t *testing.T) {
	tunnel := Tunnel{
		ID:            1,
		Encapsulation: "l2tpv3",
		Enabled:       true,
		IPsec: &TunnelIPsec{
			IPsecTunnelID: 101,
			LocalAddress:  "192.168.1.253",
			RemoteAddress: "itm.ohno.be",
			PreSharedKey:  "secret123",
			IKEv2Proposal: IKEv2Proposal{
				EncryptionAES128: true,
				IntegritySHA1:    true,
				GroupFourteen:    true,
			},
			Transform: IPsecTransform{
				Protocol:         "esp",
				EncryptionAES128: true,
				IntegritySHA1:    true,
			},
			Keepalive: &TunnelIPsecKeepalive{
				Enabled:  true,
				Mode:     "heartbeat",
				Interval: 10,
				Retry:    6,
			},
			SecureFilterIn: []int{200028, 200099},
			TCPMSSLimit:    "auto",
		},
		L2TP: &TunnelL2TP{
			Hostname:       "ebisu-RTX1210",
			LocalRouterID:  "192.168.1.253",
			RemoteRouterID: "192.168.1.254",
			RemoteEndID:    "shin1",
			AlwaysOn:       true,
			TunnelAuth: &TunnelL2TPAuth{
				Enabled:  true,
				Password: "password123",
			},
			Keepalive: &TunnelL2TPKeepalive{
				Enabled:  true,
				Interval: 60,
				Retry:    3,
			},
		},
	}

	commands := BuildTunnelCommands(tunnel)

	assert.Contains(t, commands, "tunnel select 1")
	assert.Contains(t, commands, "tunnel encapsulation l2tpv3")
	assert.Contains(t, commands, "ipsec tunnel 101")
	// IKE commands use tunnel_id (1), not ipsec_tunnel_id (101)
	// RTX uses separate ID spaces for IPsec tunnels and IKE gateways
	assert.Contains(t, commands, "ipsec ike local address 1 192.168.1.253")
	assert.Contains(t, commands, "ipsec ike remote address 1 itm.ohno.be")
	assert.Contains(t, commands, "ipsec ike pre-shared-key 1 text secret123")
	assert.Contains(t, commands, "ipsec ike keepalive use 1 on heartbeat 10 6")
	assert.Contains(t, commands, "ip tunnel secure filter in 200028 200099")
	assert.Contains(t, commands, "ip tunnel tcp mss limit auto")
	assert.Contains(t, commands, "l2tp hostname ebisu-RTX1210")
	assert.Contains(t, commands, "l2tp local router-id 192.168.1.253")
	assert.Contains(t, commands, "l2tp remote router-id 192.168.1.254")
	assert.Contains(t, commands, "l2tp remote end-id shin1")
	assert.Contains(t, commands, "l2tp always-on on")
	assert.Contains(t, commands, "l2tp tunnel auth on password123")
	assert.Contains(t, commands, "l2tp keepalive use on 60 3")
	assert.Contains(t, commands, "tunnel enable 1")
}

func TestValidateTunnel(t *testing.T) {
	tests := []struct {
		name        string
		tunnel      Tunnel
		expectError bool
		errContains string
	}{
		{
			name: "valid IPsec tunnel",
			tunnel: Tunnel{
				ID:            1,
				Encapsulation: "ipsec",
				IPsec: &TunnelIPsec{
					PreSharedKey: "secret",
				},
			},
			expectError: false,
		},
		{
			name: "valid L2TPv3 tunnel",
			tunnel: Tunnel{
				ID:            1,
				Encapsulation: "l2tpv3",
				L2TP:          &TunnelL2TP{},
			},
			expectError: false,
		},
		{
			name: "valid L2TP tunnel",
			tunnel: Tunnel{
				ID:            1,
				Encapsulation: "l2tp",
				IPsec: &TunnelIPsec{
					PreSharedKey: "secret",
				},
				L2TP: &TunnelL2TP{},
			},
			expectError: false,
		},
		{
			name: "invalid tunnel_id",
			tunnel: Tunnel{
				ID:            0,
				Encapsulation: "ipsec",
			},
			expectError: true,
			errContains: "tunnel_id must be positive",
		},
		{
			name: "invalid encapsulation",
			tunnel: Tunnel{
				ID:            1,
				Encapsulation: "invalid",
			},
			expectError: true,
			errContains: "encapsulation must be",
		},
		{
			name: "IPsec encapsulation missing ipsec block",
			tunnel: Tunnel{
				ID:            1,
				Encapsulation: "ipsec",
			},
			expectError: true,
			errContains: "ipsec block is required",
		},
		{
			name: "IPsec encapsulation with L2TP block",
			tunnel: Tunnel{
				ID:            1,
				Encapsulation: "ipsec",
				IPsec: &TunnelIPsec{
					PreSharedKey: "secret",
				},
				L2TP: &TunnelL2TP{},
			},
			expectError: true,
			errContains: "l2tp block is not allowed",
		},
		{
			name: "L2TPv3 missing L2TP block",
			tunnel: Tunnel{
				ID:            1,
				Encapsulation: "l2tpv3",
			},
			expectError: true,
			errContains: "l2tp block is required",
		},
		{
			name: "L2TP missing IPsec block",
			tunnel: Tunnel{
				ID:            1,
				Encapsulation: "l2tp",
				L2TP:          &TunnelL2TP{},
			},
			expectError: true,
			errContains: "ipsec block is required",
		},
		// Note: pre_shared_key validation is handled by Terraform schema, not here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTunnel(tt.tunnel)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
