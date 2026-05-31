package parsers

import (
	"reflect"
	"testing"
)

// TestParseScopeConfig_WrappedOption121 reproduces the v0.14.2 read-back failure
// for rtx_dhcp_scope.ebisu_main. RTX wraps the long `dhcp scope option 1 ...
// classless_static_route=<hex>` line at the console width (80), splitting the
// word "classless_static_route" across two physical lines. Before the de-wrap,
// the parser saw "...classless_sta" (no '=', dropped) and an orphan
// "tic_route=..." line (no pattern match), yielding ZERO classless routes and
// the "element N has vanished" inconsistent-result error. After de-wrapping, all
// three routes parse, with dns/router preserved.
//
// The input below is the verbatim `show config | grep "dhcp scope"` output from
// the live RTX1210 (hnd), wrap point included.
func TestParseScopeConfig_WrappedOption121(t *testing.T) {
	raw := "dhcp scope 1 192.168.1.20-192.168.1.99/16 expire 12:00\n" +
		"dhcp scope bind 1 192.168.1.20 bc:24:11:00:00:64\n" +
		"dhcp scope option 1 dns=192.168.1.61,1.1.1.1 router=192.168.1.253 classless_sta\n" +
		"tic_route=00,c0,a8,01,fd,12,0a,21,80,c0,a8,01,3c,0a,64,40,c0,a8,01,3c\n"

	parser := NewDHCPScopeParser()
	scope, err := parser.ParseSingleScope(raw, 1)
	if err != nil {
		t.Fatalf("ParseSingleScope error = %v", err)
	}

	wantRoutes := []ClasslessRoute{
		{Destination: "0.0.0.0/0", Gateway: "192.168.1.253"},
		{Destination: "10.33.128.0/18", Gateway: "192.168.1.60"},
		{Destination: "100.64.0.0/10", Gateway: "192.168.1.60"},
	}
	if !reflect.DeepEqual(scope.Options.ClasslessStaticRoutes, wantRoutes) {
		t.Errorf("ClasslessStaticRoutes after de-wrap =\n  %+v\nwant\n  %+v",
			scope.Options.ClasslessStaticRoutes, wantRoutes)
	}

	// dns/router (before the wrap point) must still parse.
	if !reflect.DeepEqual(scope.Options.DNSServers, []string{"192.168.1.61", "1.1.1.1"}) {
		t.Errorf("DNSServers = %v, want [192.168.1.61 1.1.1.1]", scope.Options.DNSServers)
	}
	if !reflect.DeepEqual(scope.Options.Routers, []string{"192.168.1.253"}) {
		t.Errorf("Routers = %v, want [192.168.1.253]", scope.Options.Routers)
	}
}

// TestParseScopeConfig_WrappedOption121_MultiLineWrap covers a value long enough
// to wrap across three physical lines (defensive: the de-wrap must join every
// continuation, not just one).
func TestParseScopeConfig_WrappedOption121_MultiLineWrap(t *testing.T) {
	// Same logical line as above, split at two arbitrary points.
	raw := "dhcp scope 1 192.168.1.20-192.168.1.99/16 expire 12:00\n" +
		"dhcp scope option 1 dns=192.168.1.61,1.1.1.1 router=192.168.1.253 classl\n" +
		"ess_static_route=00,c0,a8,01,fd,12,0a,21,80,c0,a8,01,3c,0a,64,40,c0,a8,0\n" +
		"1,3c\n"

	parser := NewDHCPScopeParser()
	scope, err := parser.ParseSingleScope(raw, 1)
	if err != nil {
		t.Fatalf("ParseSingleScope error = %v", err)
	}
	if len(scope.Options.ClasslessStaticRoutes) != 3 {
		t.Fatalf("expected 3 classless routes across a 3-line wrap, got %d: %+v",
			len(scope.Options.ClasslessStaticRoutes), scope.Options.ClasslessStaticRoutes)
	}
}

// TestParseScopeConfig_UnwrappedUnchanged guards that de-wrapping is inert for
// the common single-line (unwrapped) case — every line already contains
// "dhcp scope", so nothing is joined.
func TestParseScopeConfig_UnwrappedUnchanged(t *testing.T) {
	raw := "dhcp scope 1 192.168.1.0/24 expire 72:00\n" +
		"dhcp scope option 1 dns=192.168.1.253 router=192.168.1.253 classless_static_route=00,c0,a8,01,fd\n"

	parser := NewDHCPScopeParser()
	scope, err := parser.ParseSingleScope(raw, 1)
	if err != nil {
		t.Fatalf("ParseSingleScope error = %v", err)
	}
	wantRoutes := []ClasslessRoute{{Destination: "0.0.0.0/0", Gateway: "192.168.1.253"}}
	if !reflect.DeepEqual(scope.Options.ClasslessStaticRoutes, wantRoutes) {
		t.Errorf("ClasslessStaticRoutes = %+v, want %+v", scope.Options.ClasslessStaticRoutes, wantRoutes)
	}
}
