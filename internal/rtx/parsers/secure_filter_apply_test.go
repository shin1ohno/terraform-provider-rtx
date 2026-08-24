package parsers

import (
	"reflect"
	"testing"
)

// Verbatim `show config | grep "secure filter"` output from the HND RTX1210
// (Rev.14.01.42) on 2026-08-23, wrapping included. The two long lines are the
// reason this fixture is kept literal: the router breaks them at ~80 columns and
// the `dynamic ...` suffix lands on the continuation line.
const hndSecureFilterOutput = `ip lan2 secure filter in 45 50 55 60 65 70 75 80 dynamic 1 6 11 16 21 26 100 10
5
ip lan2 secure filter out 1 6 11 16 21 26 31 36 41
ipv6 lan2 secure filter in 1 6 11 dynamic 6 31 36 41 46 51
ipv6 lan2 secure filter out 100 105 110 115 120 125 130 135 140 145 150 155 160
 165 dynamic 6 31 36 41 46 51
 ip tunnel secure filter in 300 305
 ipv6 tunnel secure filter in 300
 ipv6 tunnel secure filter out 300`

func TestParseInterfaceIPv6SecureFilterWithDynamic_UnwrappedLine(t *testing.T) {
	got, err := ParseInterfaceIPv6SecureFilterWithDynamic(hndSecureFilterOutput)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	in, ok := got["lan2"]["in"]
	if !ok {
		t.Fatal("lan2 in missing")
	}
	if want := []int{1, 6, 11}; !reflect.DeepEqual(in.StaticIDs, want) {
		t.Errorf("lan2 in static = %v, want %v", in.StaticIDs, want)
	}
	if want := []int{6, 31, 36, 41, 46, 51}; !reflect.DeepEqual(in.DynamicIDs, want) {
		t.Errorf("lan2 in dynamic = %v, want %v", in.DynamicIDs, want)
	}
}

// The out line wraps mid-list. Before the rejoin the continuation carrying
// `dynamic 6 31 36 41 46 51` was dropped, so DynamicIDs came back empty and a
// device-side change to it could never be seen.
func TestParseInterfaceIPv6SecureFilterWithDynamic_WrappedLineKeepsDynamicSuffix(t *testing.T) {
	got, err := ParseInterfaceIPv6SecureFilterWithDynamic(hndSecureFilterOutput)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	out, ok := got["lan2"]["out"]
	if !ok {
		t.Fatal("lan2 out missing")
	}
	want := []int{100, 105, 110, 115, 120, 125, 130, 135, 140, 145, 150, 155, 160, 165}
	if !reflect.DeepEqual(out.StaticIDs, want) {
		t.Errorf("lan2 out static = %v, want %v", out.StaticIDs, want)
	}
	if want := []int{6, 31, 36, 41, 46, 51}; !reflect.DeepEqual(out.DynamicIDs, want) {
		t.Errorf("lan2 out dynamic = %v, want %v", out.DynamicIDs, want)
	}
}

// The v4 in line wraps mid-NUMBER: it ends "... 100 10" and continues "5", which
// has to rejoin as 105 rather than 10 and 5.
func TestParseInterfaceSecureFilterWithDynamic_MidNumberWrap(t *testing.T) {
	got, err := ParseInterfaceSecureFilterWithDynamic(hndSecureFilterOutput)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	in, ok := got["lan2"]["in"]
	if !ok {
		t.Fatal("lan2 in missing")
	}
	if want := []int{45, 50, 55, 60, 65, 70, 75, 80}; !reflect.DeepEqual(in.StaticIDs, want) {
		t.Errorf("lan2 in static = %v, want %v", in.StaticIDs, want)
	}
	if want := []int{1, 6, 11, 16, 21, 26, 100, 105}; !reflect.DeepEqual(in.DynamicIDs, want) {
		t.Errorf("lan2 in dynamic = %v, want %v", in.DynamicIDs, want)
	}
}

// The IPv4 parser must not adopt the ipv6 lines and vice versa — both families are
// in one grep now.
func TestParseInterfaceSecureFilter_FamiliesDoNotBleed(t *testing.T) {
	v4, err := ParseInterfaceSecureFilterWithDynamic(hndSecureFilterOutput)
	if err != nil {
		t.Fatalf("v4 parse: %v", err)
	}
	if got := v4["lan2"]["out"].StaticIDs; !reflect.DeepEqual(got, []int{1, 6, 11, 16, 21, 26, 31, 36, 41}) {
		t.Errorf("v4 lan2 out picked up the wrong line: %v", got)
	}

	v6, err := ParseInterfaceIPv6SecureFilterWithDynamic(hndSecureFilterOutput)
	if err != nil {
		t.Fatalf("v6 parse: %v", err)
	}
	if got := v6["lan2"]["in"].StaticIDs; !reflect.DeepEqual(got, []int{1, 6, 11}) {
		t.Errorf("v6 lan2 in picked up the wrong line: %v", got)
	}
}

func TestBuildShowSecureFilterCommand_MatchesApplyLinesNotDefinitions(t *testing.T) {
	// The bug this replaced: grepping "ipv6 filter" cannot match
	// "ipv6 lan2 secure filter ...".
	if got, want := BuildShowSecureFilterCommand(), `show config | grep "secure filter"`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
