package validation

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func validateFilterAddress(t *testing.T, value types.String) bool {
	t.Helper()
	req := validator.StringRequest{Path: path.Root("source"), ConfigValue: value}
	resp := &validator.StringResponse{}
	IPv6FilterAddressValidator().ValidateString(context.Background(), req, resp)
	return !resp.Diagnostics.HasError()
}

func TestIPv6FilterAddressValidator_Accepts(t *testing.T) {
	for _, v := range []string{
		"*",
		"2001:db8::1",
		"fe80::1",
		"2001:db8::/32",
		"dhcp-prefix@lan2::/64",
		"ra-prefix@lan2::/64",
		"dhcp-prefix@lan1::/56",
	} {
		if !validateFilterAddress(t, types.StringValue(v)) {
			t.Errorf("expected %q to be accepted", v)
		}
	}
}

func TestIPv6FilterAddressValidator_Rejects(t *testing.T) {
	for _, v := range []string{
		"",
		"192.168.1.1",         // IPv4 literal — this is the IPv6 filter
		"192.168.1.0/24",      // IPv4 CIDR
		"dhcp-prefix@lan2",    // missing ::/<length>
		"dhcp-prefix::/64",    // missing @<interface>
		"prefix@lan2::/64",    // neither ra- nor dhcp-
		"dhcp-prefix@lan2::/", // missing length
		"not-an-address",
	} {
		if validateFilterAddress(t, types.StringValue(v)) {
			t.Errorf("expected %q to be rejected", v)
		}
	}
}

func TestIPv6FilterAddressValidator_NullAndUnknownAreSkipped(t *testing.T) {
	// Required attributes are still unknown during plan; validating then would
	// fail every plan that computes the value.
	if !validateFilterAddress(t, types.StringNull()) {
		t.Error("null must be skipped")
	}
	if !validateFilterAddress(t, types.StringUnknown()) {
		t.Error("unknown must be skipped")
	}
}

// A wrong interface is syntactically valid and the router accepts it, so the
// validator cannot catch it. Pinned here so nobody expects it to.
func TestIPv6FilterAddressValidator_WrongInterfaceStillPasses(t *testing.T) {
	if !validateFilterAddress(t, types.StringValue("dhcp-prefix@lan1::/64")) {
		t.Error("a syntactically valid prefix ref must pass regardless of which interface it names")
	}
}
