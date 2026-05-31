package parsers

import (
	"testing"
)

func TestSNMPParser_ParseSNMPConfig(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *SNMPConfig
		wantErr bool
	}{
		{
			name: "full configuration",
			input: `snmp sysname RTX830-Main
snmp syslocation Tokyo Data Center
snmp syscontact admin@example.com
snmp community read-only public
snmp community read-write private
snmp trap community public
snmp host 192.168.1.100
snmp host 192.168.1.101
snmp trap enable snmp coldstart warmstart linkdown linkup`,
			want: &SNMPConfig{
				SysName:     "RTX830-Main",
				SysLocation: "Tokyo Data Center",
				SysContact:  "admin@example.com",
				Communities: []SNMPCommunity{
					{Name: "public", Permission: "ro"},
					{Name: "private", Permission: "rw"},
				},
				Hosts: []SNMPHost{
					{Address: "192.168.1.100", Community: "public"},
					{Address: "192.168.1.101", Community: "public"},
				},
				TrapEnable: []string{"coldstart", "warmstart", "linkdown", "linkup"},
			},
		},
		{
			name: "minimal configuration",
			input: `snmp community read-only public
snmp host 192.168.1.100`,
			want: &SNMPConfig{
				Communities: []SNMPCommunity{
					{Name: "public", Permission: "ro"},
				},
				Hosts: []SNMPHost{
					{Address: "192.168.1.100"},
				},
				TrapEnable: []string{},
			},
		},
		{
			name: "community with ACL",
			input: `snmp community read-only public 10
snmp community read-write admin 20`,
			want: &SNMPConfig{
				Communities: []SNMPCommunity{
					{Name: "public", Permission: "ro", ACL: "10"},
					{Name: "admin", Permission: "rw", ACL: "20"},
				},
				Hosts:      []SNMPHost{},
				TrapEnable: []string{},
			},
		},
		{
			name:  "empty configuration",
			input: "",
			want: &SNMPConfig{
				Communities: []SNMPCommunity{},
				Hosts:       []SNMPHost{},
				TrapEnable:  []string{},
			},
		},
		{
			name: "sysname with spaces (quoted)",
			input: `snmp sysname My Router Name
snmp syslocation Building A, Floor 3`,
			want: &SNMPConfig{
				SysName:     "My Router Name",
				SysLocation: "Building A, Floor 3",
				Communities: []SNMPCommunity{},
				Hosts:       []SNMPHost{},
				TrapEnable:  []string{},
			},
		},
		{
			name:  "all trap types",
			input: `snmp trap enable snmp all`,
			want: &SNMPConfig{
				Communities: []SNMPCommunity{},
				Hosts:       []SNMPHost{},
				TrapEnable:  []string{"all"},
			},
		},
		{
			name: "multiple communities",
			input: `snmp community read-only monitoring
snmp community read-only public
snmp community read-write admin
snmp community read-write secure`,
			want: &SNMPConfig{
				Communities: []SNMPCommunity{
					{Name: "monitoring", Permission: "ro"},
					{Name: "public", Permission: "ro"},
					{Name: "admin", Permission: "rw"},
					{Name: "secure", Permission: "rw"},
				},
				Hosts:      []SNMPHost{},
				TrapEnable: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewSNMPParser()
			got, err := p.ParseSNMPConfig(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSNMPConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Check SysName
			if got.SysName != tt.want.SysName {
				t.Errorf("SysName = %q, want %q", got.SysName, tt.want.SysName)
			}

			// Check SysLocation
			if got.SysLocation != tt.want.SysLocation {
				t.Errorf("SysLocation = %q, want %q", got.SysLocation, tt.want.SysLocation)
			}

			// Check SysContact
			if got.SysContact != tt.want.SysContact {
				t.Errorf("SysContact = %q, want %q", got.SysContact, tt.want.SysContact)
			}

			// Check Communities
			if len(got.Communities) != len(tt.want.Communities) {
				t.Errorf("Communities count = %d, want %d", len(got.Communities), len(tt.want.Communities))
			} else {
				for i, c := range got.Communities {
					if c.Name != tt.want.Communities[i].Name {
						t.Errorf("Community[%d].Name = %q, want %q", i, c.Name, tt.want.Communities[i].Name)
					}
					if c.Permission != tt.want.Communities[i].Permission {
						t.Errorf("Community[%d].Permission = %q, want %q", i, c.Permission, tt.want.Communities[i].Permission)
					}
					if c.ACL != tt.want.Communities[i].ACL {
						t.Errorf("Community[%d].ACL = %q, want %q", i, c.ACL, tt.want.Communities[i].ACL)
					}
				}
			}

			// Check Hosts
			if len(got.Hosts) != len(tt.want.Hosts) {
				t.Errorf("Hosts count = %d, want %d", len(got.Hosts), len(tt.want.Hosts))
			} else {
				for i, h := range got.Hosts {
					if h.Address != tt.want.Hosts[i].Address {
						t.Errorf("Host[%d].Address = %q, want %q", i, h.Address, tt.want.Hosts[i].Address)
					}
					if h.Community != tt.want.Hosts[i].Community {
						t.Errorf("Host[%d].Community = %q, want %q", i, h.Community, tt.want.Hosts[i].Community)
					}
					if h.Version != tt.want.Hosts[i].Version {
						t.Errorf("Host[%d].Version = %q, want %q", i, h.Version, tt.want.Hosts[i].Version)
					}
				}
			}

			// Check TrapEnable
			if len(got.TrapEnable) != len(tt.want.TrapEnable) {
				t.Errorf("TrapEnable count = %d, want %d", len(got.TrapEnable), len(tt.want.TrapEnable))
			} else {
				for i, trap := range got.TrapEnable {
					if trap != tt.want.TrapEnable[i] {
						t.Errorf("TrapEnable[%d] = %q, want %q", i, trap, tt.want.TrapEnable[i])
					}
				}
			}
		})
	}
}

func TestSNMPParser_ParseHostAccessControl(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantV1    []string
		wantV2c   []string
		wantHosts []string // trap-host addresses that must remain routed to Hosts
	}{
		{
			name:    "snmp host any -> v1 access control",
			input:   "snmp host any",
			wantV1:  []string{"any"},
			wantV2c: []string{},
		},
		{
			name:    "snmp host none -> v1 access control",
			input:   "snmp host none",
			wantV1:  []string{"none"},
			wantV2c: []string{},
		},
		{
			name:    "snmp host IPv4 range -> v1 access control",
			input:   "snmp host 192.168.1.1-192.168.1.10",
			wantV1:  []string{"192.168.1.1-192.168.1.10"},
			wantV2c: []string{},
		},
		{
			name:    "snmp host lan1 interface token -> v1 access control",
			input:   "snmp host lan1",
			wantV1:  []string{"lan1"},
			wantV2c: []string{},
		},
		{
			name:    "snmp host bridge2 interface token -> v1 access control",
			input:   "snmp host bridge2",
			wantV1:  []string{"bridge2"},
			wantV2c: []string{},
		},
		{
			name:    "snmpv2c host any -> v2c access control",
			input:   "snmpv2c host any",
			wantV1:  []string{},
			wantV2c: []string{"any"},
		},
		{
			name:    "snmpv2c host with communities captures host only",
			input:   "snmpv2c host 192.168.1.100 public private",
			wantV1:  []string{},
			wantV2c: []string{"192.168.1.100"},
		},
		{
			name:    "snmpv2c host IPv4 range -> v2c access control",
			input:   "snmpv2c host 192.168.1.1-192.168.1.10",
			wantV1:  []string{},
			wantV2c: []string{"192.168.1.1-192.168.1.10"},
		},
		{
			name:      "pure IP snmp host stays a trap host (Hosts), not access control",
			input:     "snmp host 192.168.1.100",
			wantV1:    []string{},
			wantV2c:   []string{},
			wantHosts: []string{"192.168.1.100"},
		},
		{
			name: "mixed: trap host IP + v1 any + v2c any",
			input: `snmp host 192.168.1.100
snmp host any
snmpv2c host any`,
			wantV1:    []string{"any"},
			wantV2c:   []string{"any"},
			wantHosts: []string{"192.168.1.100"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewSNMPParser()
			got, err := p.ParseSNMPConfig(tt.input)
			if err != nil {
				t.Fatalf("ParseSNMPConfig() error = %v", err)
			}

			if !stringSlicesEqualTest(got.HostAccessV1, tt.wantV1) {
				t.Errorf("HostAccessV1 = %v, want %v", got.HostAccessV1, tt.wantV1)
			}
			if !stringSlicesEqualTest(got.HostAccessV2c, tt.wantV2c) {
				t.Errorf("HostAccessV2c = %v, want %v", got.HostAccessV2c, tt.wantV2c)
			}

			gotHosts := make([]string, len(got.Hosts))
			for i, h := range got.Hosts {
				gotHosts[i] = h.Address
			}
			if tt.wantHosts == nil {
				tt.wantHosts = []string{}
			}
			if !stringSlicesEqualTest(gotHosts, tt.wantHosts) {
				t.Errorf("Hosts addresses = %v, want %v", gotHosts, tt.wantHosts)
			}
		})
	}
}

func TestBuildSNMPHostAccessCommand(t *testing.T) {
	// Note: a bare IPv4 is intentionally absent — it is not a valid snmp_host
	// (SNMPv1 access-control) value (the schema validator + ValidateSNMPConfig
	// reject it). The IPv4 range form is valid because it carries a hyphen.
	tests := []struct {
		v    string
		want string
	}{
		{"any", "snmp host any"},
		{"none", "snmp host none"},
		{"192.168.1.1-192.168.1.10", "snmp host 192.168.1.1-192.168.1.10"},
		{"lan1", "snmp host lan1"},
	}
	for _, tt := range tests {
		t.Run(tt.v, func(t *testing.T) {
			if got := BuildSNMPHostAccessCommand(tt.v); got != tt.want {
				t.Errorf("BuildSNMPHostAccessCommand(%q) = %q, want %q", tt.v, got, tt.want)
			}
		})
	}
}

func TestBuildSNMPv2cHostCommand(t *testing.T) {
	tests := []struct {
		v    string
		want string
	}{
		{"any", "snmpv2c host any"},
		{"none", "snmpv2c host none"},
		{"192.168.1.76", "snmpv2c host 192.168.1.76"},
		{"bridge1", "snmpv2c host bridge1"},
	}
	for _, tt := range tests {
		t.Run(tt.v, func(t *testing.T) {
			if got := BuildSNMPv2cHostCommand(tt.v); got != tt.want {
				t.Errorf("BuildSNMPv2cHostCommand(%q) = %q, want %q", tt.v, got, tt.want)
			}
		})
	}
}

func TestBuildDeleteSNMPHostAccessCommands(t *testing.T) {
	t.Run("delete v1 host access", func(t *testing.T) {
		if got := BuildDeleteSNMPHostAccessCommand("any"); got != "no snmp host any" {
			t.Errorf("BuildDeleteSNMPHostAccessCommand() = %q, want %q", got, "no snmp host any")
		}
	})
	t.Run("delete v2c host access", func(t *testing.T) {
		if got := BuildDeleteSNMPv2cHostCommand("any"); got != "no snmpv2c host any" {
			t.Errorf("BuildDeleteSNMPv2cHostCommand() = %q, want %q", got, "no snmpv2c host any")
		}
	})
}

func TestValidateSNMPConfig_HostAccess(t *testing.T) {
	tests := []struct {
		name    string
		config  SNMPConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid v1 any",
			config:  SNMPConfig{HostAccessV1: []string{"any"}},
			wantErr: false,
		},
		{
			name:    "valid v1 none + range + lan + bridge (no bare IP)",
			config:  SNMPConfig{HostAccessV1: []string{"none", "192.168.1.1-192.168.1.10", "lan1", "bridge2"}},
			wantErr: false,
		},
		{
			name:    "bare IPv4 rejected for snmp_host (must use host block / snmpv2c_host)",
			config:  SNMPConfig{HostAccessV1: []string{"192.168.1.76"}},
			wantErr: true,
			errMsg:  "invalid snmp_host value",
		},
		{
			name:    "valid v2c any",
			config:  SNMPConfig{HostAccessV2c: []string{"any"}},
			wantErr: false,
		},
		{
			name:    "valid v2c bare IPv4 accepted",
			config:  SNMPConfig{HostAccessV2c: []string{"192.168.1.76"}},
			wantErr: false,
		},
		{
			name:    "valid v2c IPv4 range accepted",
			config:  SNMPConfig{HostAccessV2c: []string{"192.168.1.1-192.168.1.10"}},
			wantErr: false,
		},
		{
			name:    "invalid v1 token",
			config:  SNMPConfig{HostAccessV1: []string{"everyone"}},
			wantErr: true,
			errMsg:  "invalid snmp_host value",
		},
		{
			name:    "invalid v2c token",
			config:  SNMPConfig{HostAccessV2c: []string{"wan0"}},
			wantErr: true,
			errMsg:  "invalid snmpv2c_host value",
		},
		{
			name:    "malformed IPv4 octet rejected by tightened regex (v2c)",
			config:  SNMPConfig{HostAccessV2c: []string{"192.168.1.999"}},
			wantErr: true,
			errMsg:  "invalid snmpv2c_host value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSNMPConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSNMPConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && !snmpContains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateSNMPConfig() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

// TestSNMPParser_DewrapConsoleLine verifies that a long `snmp ...` line that the
// RTX console wrapped at 80 columns (bare CRLF, no continuation marker) is
// reassembled before pattern matching, so the field does not silently vanish on
// read-back (would otherwise produce "Provider produced inconsistent result
// after apply"). Mirrors the de-wrap coverage in dns/dhcp_scope.
func TestSNMPParser_DewrapConsoleLine(t *testing.T) {
	p := NewSNMPParser()

	t.Run("wrapped trap enable line parses as one logical line", func(t *testing.T) {
		// A `snmp trap enable snmp ...` line, wrapped by the RTX console with a
		// bare CRLF mid-token ("authenticat" + "ion") and no continuation marker.
		// dewrapConsoleLines must rejoin the continuation onto the logical line.
		const head = "snmp trap enable snmp coldstart warmstart linkdown linkup authenticat"
		const tail = "ion enterprise"
		raw := head + "\r\n" + tail

		config, err := p.ParseSNMPConfig(raw)
		if err != nil {
			t.Fatalf("ParseSNMPConfig error: %v", err)
		}

		want := []string{"coldstart", "warmstart", "linkdown", "linkup", "authentication", "enterprise"}
		if !stringSlicesEqualTest(config.TrapEnable, want) {
			t.Errorf("TrapEnable = %v, want %v (wrapped continuation was dropped?)", config.TrapEnable, want)
		}
	})

	t.Run("wrapped community read-only line parses as one logical line", func(t *testing.T) {
		// `snmp community read-only <name> <acl>` wrapped mid the community name.
		const head = "snmp community read-only verylongcommunitystringnamethatwraps"
		const tail = "atconsolewidth 100"
		raw := head + "\r\n" + tail

		config, err := p.ParseSNMPConfig(raw)
		if err != nil {
			t.Fatalf("ParseSNMPConfig error: %v", err)
		}
		if len(config.Communities) != 1 {
			t.Fatalf("Communities count = %d, want 1 (wrapped continuation dropped?)", len(config.Communities))
		}
		if config.Communities[0].Name != "verylongcommunitystringnamethatwrapsatconsolewidth" {
			t.Errorf("Community Name = %q, want reassembled name", config.Communities[0].Name)
		}
		if config.Communities[0].ACL != "100" {
			t.Errorf("Community ACL = %q, want 100", config.Communities[0].ACL)
		}
	})
}

// TestSNMPParser_HostAccessBuildParseRoundTrip verifies that a valid SNMPv1
// snmp_host value ("any") and a valid SNMPv2c snmpv2c_host value (bare IPv4)
// survive Build -> Parse unchanged, each into its own access-control list and
// NOT into the trap Hosts block.
func TestSNMPParser_HostAccessBuildParseRoundTrip(t *testing.T) {
	p := NewSNMPParser()

	raw := BuildSNMPHostAccessCommand("any") + "\n" + BuildSNMPv2cHostCommand("192.168.1.76")
	config, err := p.ParseSNMPConfig(raw)
	if err != nil {
		t.Fatalf("ParseSNMPConfig error: %v", err)
	}

	if !stringSlicesEqualTest(config.HostAccessV1, []string{"any"}) {
		t.Errorf("HostAccessV1 = %v, want [any]", config.HostAccessV1)
	}
	if !stringSlicesEqualTest(config.HostAccessV2c, []string{"192.168.1.76"}) {
		t.Errorf("HostAccessV2c = %v, want [192.168.1.76]", config.HostAccessV2c)
	}
	if len(config.Hosts) != 0 {
		t.Errorf("Hosts = %v, want empty (access-control values must not land in trap Hosts)", config.Hosts)
	}
}

// stringSlicesEqualTest is a local test helper comparing two string slices.
func stringSlicesEqualTest(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestBuildSNMPSysNameCommand(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "RTX830-Main",
			want: "snmp sysname RTX830-Main",
		},
		{
			name: "My Router",
			want: "snmp sysname My Router",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildSNMPSysNameCommand(tt.name)
			if got != tt.want {
				t.Errorf("BuildSNMPSysNameCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildSNMPSysLocationCommand(t *testing.T) {
	tests := []struct {
		location string
		want     string
	}{
		{
			location: "Tokyo",
			want:     "snmp syslocation Tokyo",
		},
		{
			location: "Building A, Floor 3",
			want:     "snmp syslocation Building A, Floor 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.location, func(t *testing.T) {
			got := BuildSNMPSysLocationCommand(tt.location)
			if got != tt.want {
				t.Errorf("BuildSNMPSysLocationCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildSNMPSysContactCommand(t *testing.T) {
	tests := []struct {
		contact string
		want    string
	}{
		{
			contact: "admin@example.com",
			want:    "snmp syscontact admin@example.com",
		},
		{
			contact: "John Doe",
			want:    "snmp syscontact John Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.contact, func(t *testing.T) {
			got := BuildSNMPSysContactCommand(tt.contact)
			if got != tt.want {
				t.Errorf("BuildSNMPSysContactCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildSNMPCommunityCommand(t *testing.T) {
	tests := []struct {
		name      string
		community SNMPCommunity
		want      string
	}{
		{
			name: "read-only without ACL",
			community: SNMPCommunity{
				Name:       "public",
				Permission: "ro",
			},
			want: "snmp community read-only public",
		},
		{
			name: "read-write without ACL",
			community: SNMPCommunity{
				Name:       "private",
				Permission: "rw",
			},
			want: "snmp community read-write private",
		},
		{
			name: "read-only with ACL",
			community: SNMPCommunity{
				Name:       "monitoring",
				Permission: "ro",
				ACL:        "10",
			},
			want: "snmp community read-only monitoring 10",
		},
		{
			name: "read-write with ACL",
			community: SNMPCommunity{
				Name:       "admin",
				Permission: "rw",
				ACL:        "20",
			},
			want: "snmp community read-write admin 20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildSNMPCommunityCommand(tt.community)
			if got != tt.want {
				t.Errorf("BuildSNMPCommunityCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildSNMPHostCommand(t *testing.T) {
	tests := []struct {
		name string
		host SNMPHost
		want string
	}{
		{
			name: "simple host",
			host: SNMPHost{
				Address: "192.168.1.100",
			},
			want: "snmp host 192.168.1.100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildSNMPHostCommand(tt.host)
			if got != tt.want {
				t.Errorf("BuildSNMPHostCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildSNMPTrapCommunityCommand(t *testing.T) {
	got := BuildSNMPTrapCommunityCommand("public")
	want := "snmp trap community public"
	if got != want {
		t.Errorf("BuildSNMPTrapCommunityCommand() = %q, want %q", got, want)
	}
}

func TestBuildSNMPTrapEnableCommand(t *testing.T) {
	tests := []struct {
		name      string
		trapTypes []string
		want      string
	}{
		{
			name:      "single trap type",
			trapTypes: []string{"coldstart"},
			want:      "snmp trap enable snmp coldstart",
		},
		{
			name:      "multiple trap types",
			trapTypes: []string{"coldstart", "warmstart", "linkdown", "linkup"},
			want:      "snmp trap enable snmp coldstart warmstart linkdown linkup",
		},
		{
			name:      "all traps",
			trapTypes: []string{"all"},
			want:      "snmp trap enable snmp all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildSNMPTrapEnableCommand(tt.trapTypes)
			if got != tt.want {
				t.Errorf("BuildSNMPTrapEnableCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildDeleteSNMPCommands(t *testing.T) {
	t.Run("DeleteSysName", func(t *testing.T) {
		got := BuildDeleteSNMPSysNameCommand()
		want := "no snmp sysname"
		if got != want {
			t.Errorf("BuildDeleteSNMPSysNameCommand() = %q, want %q", got, want)
		}
	})

	t.Run("DeleteSysLocation", func(t *testing.T) {
		got := BuildDeleteSNMPSysLocationCommand()
		want := "no snmp syslocation"
		if got != want {
			t.Errorf("BuildDeleteSNMPSysLocationCommand() = %q, want %q", got, want)
		}
	})

	t.Run("DeleteSysContact", func(t *testing.T) {
		got := BuildDeleteSNMPSysContactCommand()
		want := "no snmp syscontact"
		if got != want {
			t.Errorf("BuildDeleteSNMPSysContactCommand() = %q, want %q", got, want)
		}
	})

	t.Run("DeleteCommunity", func(t *testing.T) {
		community := SNMPCommunity{Name: "public", Permission: "ro"}
		got := BuildDeleteSNMPCommunityCommand(community)
		want := "no snmp community read-only public"
		if got != want {
			t.Errorf("BuildDeleteSNMPCommunityCommand() = %q, want %q", got, want)
		}
	})

	t.Run("DeleteHost", func(t *testing.T) {
		got := BuildDeleteSNMPHostCommand("192.168.1.100")
		want := "no snmp host 192.168.1.100"
		if got != want {
			t.Errorf("BuildDeleteSNMPHostCommand() = %q, want %q", got, want)
		}
	})

	t.Run("DeleteTrapCommunity", func(t *testing.T) {
		got := BuildDeleteSNMPTrapCommunityCommand()
		want := "no snmp trap community"
		if got != want {
			t.Errorf("BuildDeleteSNMPTrapCommunityCommand() = %q, want %q", got, want)
		}
	})

	t.Run("DeleteTrapEnable", func(t *testing.T) {
		got := BuildDeleteSNMPTrapEnableCommand()
		want := "no snmp trap enable snmp"
		if got != want {
			t.Errorf("BuildDeleteSNMPTrapEnableCommand() = %q, want %q", got, want)
		}
	})
}

func TestBuildShowSNMPConfigCommand(t *testing.T) {
	got := BuildShowSNMPConfigCommand()
	want := "show config | grep snmp"
	if got != want {
		t.Errorf("BuildShowSNMPConfigCommand() = %q, want %q", got, want)
	}
}

func TestValidateSNMPConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  SNMPConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			config: SNMPConfig{
				SysName:     "Router",
				SysLocation: "Tokyo",
				SysContact:  "admin@example.com",
				Communities: []SNMPCommunity{
					{Name: "public", Permission: "ro"},
					{Name: "private", Permission: "rw"},
				},
				Hosts: []SNMPHost{
					{Address: "192.168.1.100"},
				},
				TrapEnable: []string{"coldstart", "warmstart"},
			},
			wantErr: false,
		},
		{
			name: "empty community name",
			config: SNMPConfig{
				Communities: []SNMPCommunity{
					{Name: "", Permission: "ro"},
				},
			},
			wantErr: true,
			errMsg:  "community name cannot be empty",
		},
		{
			name: "invalid community permission",
			config: SNMPConfig{
				Communities: []SNMPCommunity{
					{Name: "public", Permission: "invalid"},
				},
			},
			wantErr: true,
			errMsg:  "community permission must be 'ro' or 'rw'",
		},
		{
			name: "community name too long",
			config: SNMPConfig{
				Communities: []SNMPCommunity{
					{Name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Permission: "ro"},
				},
			},
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
		{
			name: "empty host address",
			config: SNMPConfig{
				Hosts: []SNMPHost{
					{Address: ""},
				},
			},
			wantErr: true,
			errMsg:  "host address cannot be empty",
		},
		{
			name: "invalid host IP",
			config: SNMPConfig{
				Hosts: []SNMPHost{
					{Address: "not-an-ip"},
				},
			},
			wantErr: true,
			errMsg:  "invalid host IP address",
		},
		{
			name: "invalid SNMP version",
			config: SNMPConfig{
				Hosts: []SNMPHost{
					{Address: "192.168.1.100", Version: "3"},
				},
			},
			wantErr: true,
			errMsg:  "invalid SNMP version",
		},
		{
			name: "invalid trap type",
			config: SNMPConfig{
				TrapEnable: []string{"invalid-trap"},
			},
			wantErr: true,
			errMsg:  "invalid trap type",
		},
		{
			name: "valid SNMP version 1",
			config: SNMPConfig{
				Hosts: []SNMPHost{
					{Address: "192.168.1.100", Version: "1"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid SNMP version 2c",
			config: SNMPConfig{
				Hosts: []SNMPHost{
					{Address: "192.168.1.100", Version: "2c"},
				},
			},
			wantErr: false,
		},
		{
			name: "all valid trap types",
			config: SNMPConfig{
				TrapEnable: []string{"all", "authentication", "coldstart", "warmstart", "linkdown", "linkup", "enterprise"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSNMPConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSNMPConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !snmpContains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateSNMPConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateSNMPCommunity(t *testing.T) {
	tests := []struct {
		name      string
		community SNMPCommunity
		wantErr   bool
	}{
		{
			name:      "valid read-only",
			community: SNMPCommunity{Name: "public", Permission: "ro"},
			wantErr:   false,
		},
		{
			name:      "valid read-write",
			community: SNMPCommunity{Name: "private", Permission: "rw"},
			wantErr:   false,
		},
		{
			name:      "valid with ACL",
			community: SNMPCommunity{Name: "admin", Permission: "rw", ACL: "10"},
			wantErr:   false,
		},
		{
			name:      "empty name",
			community: SNMPCommunity{Name: "", Permission: "ro"},
			wantErr:   true,
		},
		{
			name:      "invalid permission",
			community: SNMPCommunity{Name: "test", Permission: "invalid"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSNMPCommunity(tt.community)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSNMPCommunity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSNMPHost(t *testing.T) {
	tests := []struct {
		name    string
		host    SNMPHost
		wantErr bool
	}{
		{
			name:    "valid host",
			host:    SNMPHost{Address: "192.168.1.100"},
			wantErr: false,
		},
		{
			name:    "valid with community",
			host:    SNMPHost{Address: "10.0.0.1", Community: "public"},
			wantErr: false,
		},
		{
			name:    "valid with version",
			host:    SNMPHost{Address: "10.0.0.1", Version: "2c"},
			wantErr: false,
		},
		{
			name:    "empty address",
			host:    SNMPHost{Address: ""},
			wantErr: true,
		},
		{
			name:    "invalid IP",
			host:    SNMPHost{Address: "invalid"},
			wantErr: true,
		},
		{
			name:    "invalid version",
			host:    SNMPHost{Address: "192.168.1.100", Version: "3"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSNMPHost(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSNMPHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function for error message checking
func snmpContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
