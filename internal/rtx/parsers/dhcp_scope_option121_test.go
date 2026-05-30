package parsers

import (
	"reflect"
	"testing"
)

// TestEncodeClasslessRoutes verifies the RFC 3442 (DHCP option 121) hex-byte
// CSV encoding for the spec-provided example routes.
func TestEncodeClasslessRoutes(t *testing.T) {
	testCases := []struct {
		name        string
		routes      []ClasslessRoute
		expectedHex string
	}{
		{
			name:        "default_route",
			routes:      []ClasslessRoute{{Destination: "0.0.0.0/0", Gateway: "192.168.1.253"}},
			expectedHex: "00,c0,a8,01,fd",
		},
		{
			name:        "slash18_non_byte_boundary",
			routes:      []ClasslessRoute{{Destination: "10.33.128.0/18", Gateway: "192.168.1.60"}},
			expectedHex: "12,0a,21,80,c0,a8,01,3c",
		},
		{
			name:        "slash10",
			routes:      []ClasslessRoute{{Destination: "100.64.0.0/10", Gateway: "192.168.1.60"}},
			expectedHex: "0a,64,40,c0,a8,01,3c",
		},
		{
			name: "full_multi_route",
			routes: []ClasslessRoute{
				{Destination: "0.0.0.0/0", Gateway: "192.168.1.253"},
				{Destination: "10.33.128.0/18", Gateway: "192.168.1.60"},
				{Destination: "100.64.0.0/10", Gateway: "192.168.1.60"},
			},
			expectedHex: "00,c0,a8,01,fd,12,0a,21,80,c0,a8,01,3c,0a,64,40,c0,a8,01,3c",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := encodeClasslessRoutes(tc.routes)
			if err != nil {
				t.Fatalf("encodeClasslessRoutes(%+v) error = %v", tc.routes, err)
			}
			if got != tc.expectedHex {
				t.Errorf("encodeClasslessRoutes() = %q, want %q", got, tc.expectedHex)
			}
		})
	}
}

// TestDecodeClasslessRoutes verifies decoding the spec hex back into routes,
// and that decode is the exact inverse of encode (decode∘encode == identity).
func TestDecodeClasslessRoutes(t *testing.T) {
	testCases := []struct {
		name           string
		hex            string
		expectedRoutes []ClasslessRoute
	}{
		{
			name:           "default_route",
			hex:            "00,c0,a8,01,fd",
			expectedRoutes: []ClasslessRoute{{Destination: "0.0.0.0/0", Gateway: "192.168.1.253"}},
		},
		{
			name:           "slash18_non_byte_boundary",
			hex:            "12,0a,21,80,c0,a8,01,3c",
			expectedRoutes: []ClasslessRoute{{Destination: "10.33.128.0/18", Gateway: "192.168.1.60"}},
		},
		{
			name:           "slash10",
			hex:            "0a,64,40,c0,a8,01,3c",
			expectedRoutes: []ClasslessRoute{{Destination: "100.64.0.0/10", Gateway: "192.168.1.60"}},
		},
		{
			name: "full_multi_route",
			hex:  "00,c0,a8,01,fd,12,0a,21,80,c0,a8,01,3c,0a,64,40,c0,a8,01,3c",
			expectedRoutes: []ClasslessRoute{
				{Destination: "0.0.0.0/0", Gateway: "192.168.1.253"},
				{Destination: "10.33.128.0/18", Gateway: "192.168.1.60"},
				{Destination: "100.64.0.0/10", Gateway: "192.168.1.60"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := decodeClasslessRoutes(tc.hex)
			if err != nil {
				t.Fatalf("decodeClasslessRoutes(%q) error = %v", tc.hex, err)
			}
			if !reflect.DeepEqual(got, tc.expectedRoutes) {
				t.Errorf("decodeClasslessRoutes(%q) = %+v, want %+v", tc.hex, got, tc.expectedRoutes)
			}

			// decode∘encode == identity (re-encode the decoded routes, must match hex)
			reEncoded, err := encodeClasslessRoutes(got)
			if err != nil {
				t.Fatalf("re-encode error = %v", err)
			}
			if reEncoded != tc.hex {
				t.Errorf("re-encode mismatch: got %q, want %q", reEncoded, tc.hex)
			}
		})
	}
}

// TestClasslessRoutesRoundTrip verifies encode(routes) -> decode -> equal routes
// for each spec route family.
func TestClasslessRoutesRoundTrip(t *testing.T) {
	families := [][]ClasslessRoute{
		{{Destination: "0.0.0.0/0", Gateway: "192.168.1.253"}},
		{{Destination: "10.33.128.0/18", Gateway: "192.168.1.60"}},
		{{Destination: "100.64.0.0/10", Gateway: "192.168.1.60"}},
		{
			{Destination: "0.0.0.0/0", Gateway: "192.168.1.253"},
			{Destination: "10.33.128.0/18", Gateway: "192.168.1.60"},
			{Destination: "100.64.0.0/10", Gateway: "192.168.1.60"},
		},
	}

	for i, routes := range families {
		hex, err := encodeClasslessRoutes(routes)
		if err != nil {
			t.Fatalf("family %d: encode error = %v", i, err)
		}
		decoded, err := decodeClasslessRoutes(hex)
		if err != nil {
			t.Fatalf("family %d: decode error = %v", i, err)
		}
		if !reflect.DeepEqual(decoded, routes) {
			t.Errorf("family %d round-trip mismatch:\n  original: %+v\n  decoded:  %+v", i, routes, decoded)
		}
	}
}

// TestEncodeClasslessRoutes_Errors verifies input validation.
func TestEncodeClasslessRoutes_Errors(t *testing.T) {
	testCases := []struct {
		name   string
		routes []ClasslessRoute
	}{
		{"bad_cidr", []ClasslessRoute{{Destination: "not-a-cidr", Gateway: "192.168.1.1"}}},
		{"mask_over_32", []ClasslessRoute{{Destination: "10.0.0.0/33", Gateway: "192.168.1.1"}}},
		{"bad_gateway", []ClasslessRoute{{Destination: "10.0.0.0/8", Gateway: "999.1.1.1"}}},
		{"empty_gateway", []ClasslessRoute{{Destination: "10.0.0.0/8", Gateway: ""}}},
		{"ipv6_destination", []ClasslessRoute{{Destination: "2001:db8::/32", Gateway: "192.168.1.1"}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := encodeClasslessRoutes(tc.routes); err == nil {
				t.Errorf("encodeClasslessRoutes(%+v) expected error, got nil", tc.routes)
			}
		})
	}
}

// TestDecodeClasslessRoutes_Errors verifies decode rejects malformed input.
func TestDecodeClasslessRoutes_Errors(t *testing.T) {
	testCases := []struct {
		name string
		hex  string
	}{
		{"malformed_hex", "zz,01,02"},
		{"mask_over_32", "21,0a,0a,0a,0a,c0,a8,01,01"}, // 0x21 = 33
		{"truncated_destination", "18,0a"},             // /24 needs 3 dest octets, only 1 present + no gw
		{"truncated_gateway", "08,0a,c0,a8"},           // /8: 1 dest octet, only 3 gw octets
		{"empty_token", "00,,c0,a8,01,fd"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := decodeClasslessRoutes(tc.hex); err == nil {
				t.Errorf("decodeClasslessRoutes(%q) expected error, got nil", tc.hex)
			}
		})
	}
}

// TestBuildDHCPScopeOptionsCommand_Option121 verifies the build command emits
// 121=<hex> and that it coexists with dns=/router= on the same line.
func TestBuildDHCPScopeOptionsCommand_Option121(t *testing.T) {
	testCases := []struct {
		name     string
		opts     DHCPScopeOptions
		expected string
	}{
		{
			name: "only_classless_routes",
			opts: DHCPScopeOptions{
				ClasslessStaticRoutes: []ClasslessRoute{
					{Destination: "0.0.0.0/0", Gateway: "192.168.1.253"},
					{Destination: "10.33.128.0/18", Gateway: "192.168.1.60"},
					{Destination: "100.64.0.0/10", Gateway: "192.168.1.60"},
				},
			},
			expected: "dhcp scope option 1 121=00,c0,a8,01,fd,12,0a,21,80,c0,a8,01,3c,0a,64,40,c0,a8,01,3c",
		},
		{
			name: "coexists_with_dns_and_router",
			opts: DHCPScopeOptions{
				DNSServers: []string{"192.168.1.253"},
				Routers:    []string{"192.168.1.253"},
				ClasslessStaticRoutes: []ClasslessRoute{
					{Destination: "0.0.0.0/0", Gateway: "192.168.1.253"},
					{Destination: "10.33.128.0/18", Gateway: "192.168.1.60"},
					{Destination: "100.64.0.0/10", Gateway: "192.168.1.60"},
				},
			},
			expected: "dhcp scope option 1 dns=192.168.1.253 router=192.168.1.253 121=00,c0,a8,01,fd,12,0a,21,80,c0,a8,01,3c,0a,64,40,c0,a8,01,3c",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildDHCPScopeOptionsCommand(1, tc.opts)
			if got != tc.expected {
				t.Errorf("BuildDHCPScopeOptionsCommand() =\n  %q\nwant\n  %q", got, tc.expected)
			}
		})
	}
}

// TestParseOptions_Option121_Coexistence verifies parsing a full
// "dhcp scope option" line populates DNSServers, Routers AND
// ClasslessStaticRoutes simultaneously (no field clobbers another).
func TestParseOptions_Option121_Coexistence(t *testing.T) {
	parser := NewDHCPScopeParser()
	line := "dhcp scope option 1 dns=192.168.1.253 router=192.168.1.253 domain=home.local 121=00,c0,a8,01,fd,12,0a,21,80,c0,a8,01,3c,0a,64,40,c0,a8,01,3c"

	scopes, err := parser.ParseScopeConfig(line)
	if err != nil {
		t.Fatalf("ParseScopeConfig error = %v", err)
	}
	if len(scopes) != 1 {
		t.Fatalf("Scopes count = %d, want 1", len(scopes))
	}
	s := scopes[0]

	if len(s.Options.DNSServers) != 1 || s.Options.DNSServers[0] != "192.168.1.253" {
		t.Errorf("DNSServers = %v, want [192.168.1.253]", s.Options.DNSServers)
	}
	if len(s.Options.Routers) != 1 || s.Options.Routers[0] != "192.168.1.253" {
		t.Errorf("Routers = %v, want [192.168.1.253]", s.Options.Routers)
	}
	if s.Options.DomainName != "home.local" {
		t.Errorf("DomainName = %v, want home.local", s.Options.DomainName)
	}

	wantRoutes := []ClasslessRoute{
		{Destination: "0.0.0.0/0", Gateway: "192.168.1.253"},
		{Destination: "10.33.128.0/18", Gateway: "192.168.1.60"},
		{Destination: "100.64.0.0/10", Gateway: "192.168.1.60"},
	}
	if !reflect.DeepEqual(s.Options.ClasslessStaticRoutes, wantRoutes) {
		t.Errorf("ClasslessStaticRoutes =\n  %+v\nwant\n  %+v", s.Options.ClasslessStaticRoutes, wantRoutes)
	}
}

// TestDHCPScopeOption121_BuildParseRoundTrip verifies build->parse->equal
// for the full multi-route options line (no perpetual drift).
func TestDHCPScopeOption121_BuildParseRoundTrip(t *testing.T) {
	parser := NewDHCPScopeParser()

	opts := DHCPScopeOptions{
		DNSServers: []string{"192.168.1.253"},
		Routers:    []string{"192.168.1.253"},
		ClasslessStaticRoutes: []ClasslessRoute{
			{Destination: "0.0.0.0/0", Gateway: "192.168.1.253"},
			{Destination: "10.33.128.0/18", Gateway: "192.168.1.60"},
			{Destination: "100.64.0.0/10", Gateway: "192.168.1.60"},
		},
	}

	cmd := BuildDHCPScopeOptionsCommand(1, opts)

	scopes, err := parser.ParseScopeConfig(cmd)
	if err != nil {
		t.Fatalf("ParseScopeConfig(%q) error = %v", cmd, err)
	}
	if len(scopes) != 1 {
		t.Fatalf("Scopes count = %d, want 1", len(scopes))
	}
	parsed := scopes[0]

	if !reflect.DeepEqual(parsed.Options.ClasslessStaticRoutes, opts.ClasslessStaticRoutes) {
		t.Errorf("Round-trip drift in ClasslessStaticRoutes:\n  built from: %+v\n  parsed:     %+v",
			opts.ClasslessStaticRoutes, parsed.Options.ClasslessStaticRoutes)
	}
}
