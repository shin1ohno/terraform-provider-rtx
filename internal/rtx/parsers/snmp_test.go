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
