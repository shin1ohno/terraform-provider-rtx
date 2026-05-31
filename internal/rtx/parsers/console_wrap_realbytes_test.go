package parsers

import (
	"reflect"
	"testing"
)

// These tests use the VERBATIM bytes captured from the live RTX1210 (hnd) over
// the provider's SSH session (TF_LOG=debug "raw output"), including the command
// echo, "Searching ...", the CRLF wrap, and the trailing "[RTX1210] >" prompt.
// They reproduce the exact production read-backs that the v0.14.4 de-wrap failed
// (it joined the prompt onto the wrapped value, corrupting it). \r\n is included
// deliberately.

// rtx_dhcp_scope: option 121 (classless_static_route) wraps mid-token
// ("classless_sta" + "tic_route="); the prompt and pre-output noise must not be
// joined onto the value.
func TestDewrap_RealDHCPScopeReadBack(t *testing.T) {
	raw := " show config | grep \"dhcp scope\"\r\n" +
		"Searching ...\r\n" +
		"dhcp scope 1 192.168.1.20-192.168.1.99/16 expire 12:00\r\n" +
		"dhcp scope bind 1 192.168.1.20 bc:24:11:00:00:64\r\n" +
		"dhcp scope bind 1 192.168.1.21 01 00 3e e1 c3 54 b4\r\n" +
		"dhcp scope option 1 dns=192.168.1.61,1.1.1.1 router=192.168.1.253 classless_sta\r\n" +
		"tic_route=00,c0,a8,01,fd,12,0a,21,80,c0,a8,01,3c,0a,64,40,c0,a8,01,3c\r\n" +
		"[RTX1210] >"

	scope, err := NewDHCPScopeParser().ParseSingleScope(raw, 1)
	if err != nil {
		t.Fatalf("ParseSingleScope error = %v", err)
	}
	want := []ClasslessRoute{
		{Destination: "0.0.0.0/0", Gateway: "192.168.1.253"},
		{Destination: "10.33.128.0/18", Gateway: "192.168.1.60"},
		{Destination: "100.64.0.0/10", Gateway: "192.168.1.60"},
	}
	if !reflect.DeepEqual(scope.Options.ClasslessStaticRoutes, want) {
		t.Errorf("classless_static_routes =\n  %+v\nwant\n  %+v", scope.Options.ClasslessStaticRoutes, want)
	}
	if !reflect.DeepEqual(scope.Options.DNSServers, []string{"192.168.1.61", "1.1.1.1"}) {
		t.Errorf("dns_servers = %v", scope.Options.DNSServers)
	}
	if !reflect.DeepEqual(scope.Options.Routers, []string{"192.168.1.253"}) {
		t.Errorf("routers = %v", scope.Options.Routers)
	}
}

// rtx_dns_server: `show config | grep dns` also returns the wrapped dhcp scope
// option line (it contains "dns="), plus an IPv6 `dns server select` that wraps
// at a space ("...edns=on " + "aaaa ."). The aaaa selector must survive; the
// dhcp line and prompt must not corrupt any dns entry.
func TestDewrap_RealDNSServerReadBack(t *testing.T) {
	raw := " show config | grep dns\r\n" +
		"Searching ...\r\n" +
		"dhcp scope option 1 dns=192.168.1.61,1.1.1.1 router=192.168.1.253 classless_sta\r\n" +
		"tic_route=00,c0,a8,01,fd,12,0a,21,80,c0,a8,01,3c,0a,64,40,c0,a8,01,3c\r\n" +
		"dns host bridge1\r\n" +
		"dns service recursive\r\n" +
		"dns server select 1 10.33.128.2 edns=on any home.local\r\n" +
		"dns server select 6 1.1.1.1 edns=on 1.0.0.1 edns=on a .\r\n" +
		"dns server select 11 2606:4700:4700::1111 edns=on 2606:4700:4700::1001 edns=on \r\n" +
		"aaaa .\r\n" +
		"dns private address spoof on\r\n" +
		"[RTX1210] >"

	cfg, err := NewDNSParser().ParseDNSConfig(raw)
	if err != nil {
		t.Fatalf("ParseDNSConfig error = %v", err)
	}

	var sel11 *DNSServerSelect
	for i := range cfg.ServerSelect {
		if cfg.ServerSelect[i].ID == 11 {
			sel11 = &cfg.ServerSelect[i]
		}
	}
	if sel11 == nil {
		t.Fatalf("server select 11 (aaaa) missing — wrap continuation dropped; got %d selects: %+v",
			len(cfg.ServerSelect), cfg.ServerSelect)
	}
	if sel11.RecordType != "aaaa" {
		t.Errorf("select 11 RecordType = %q, want aaaa", sel11.RecordType)
	}
	if sel11.QueryPattern != "." {
		t.Errorf("select 11 QueryPattern = %q, want '.'", sel11.QueryPattern)
	}
	if len(sel11.Servers) != 2 {
		t.Errorf("select 11 expected 2 servers, got %d: %+v", len(sel11.Servers), sel11.Servers)
	} else {
		if sel11.Servers[0].Address != "2606:4700:4700::1111" || sel11.Servers[1].Address != "2606:4700:4700::1001" {
			t.Errorf("select 11 servers = %+v", sel11.Servers)
		}
		if !sel11.Servers[0].EDNS || !sel11.Servers[1].EDNS {
			t.Errorf("select 11 servers should both have edns=on: %+v", sel11.Servers)
		}
	}

	// All three selectors present, none corrupted by the dhcp line/prompt.
	if len(cfg.ServerSelect) != 3 {
		t.Errorf("expected 3 server_select entries, got %d: %+v", len(cfg.ServerSelect), cfg.ServerSelect)
	}
	if !cfg.PrivateSpoof {
		t.Errorf("private address spoof should be parsed as on")
	}
}
