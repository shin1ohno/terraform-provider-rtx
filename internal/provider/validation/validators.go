package validation

import (
	"context"
	"net"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// IPv4AddressValidator returns a validator that checks if the string is a valid IPv4 address.
func IPv4AddressValidator() validator.String {
	return &ipv4AddressValidator{}
}

type ipv4AddressValidator struct{}

func (v ipv4AddressValidator) Description(ctx context.Context) string {
	return "value must be a valid IPv4 address"
}

func (v ipv4AddressValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be a valid IPv4 address"
}

func (v ipv4AddressValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	ip := net.ParseIP(value)
	if ip == nil || ip.To4() == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv4 Address",
			"The value must be a valid IPv4 address.",
		)
	}
}

// IPv6AddressValidator returns a validator that checks if the string is a valid IPv6 address.
func IPv6AddressValidator() validator.String {
	return &ipv6AddressValidator{}
}

type ipv6AddressValidator struct{}

func (v ipv6AddressValidator) Description(ctx context.Context) string {
	return "value must be a valid IPv6 address"
}

func (v ipv6AddressValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be a valid IPv6 address"
}

func (v ipv6AddressValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	ip := net.ParseIP(value)
	if ip == nil || ip.To4() != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv6 Address",
			"The value must be a valid IPv6 address.",
		)
	}
}

// CIDRValidator returns a validator that checks if the string is a valid CIDR notation.
func CIDRValidator() validator.String {
	return &cidrValidator{}
}

type cidrValidator struct{}

func (v cidrValidator) Description(ctx context.Context) string {
	return "value must be a valid CIDR notation"
}

func (v cidrValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be a valid CIDR notation"
}

func (v cidrValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	_, _, err := net.ParseCIDR(value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid CIDR Notation",
			"The value must be a valid CIDR notation (e.g., '192.168.1.0/24').",
		)
	}
}

// MACAddressValidator returns a validator that checks if the string is a valid MAC address.
func MACAddressValidator() validator.String {
	return &macAddressValidator{}
}

type macAddressValidator struct{}

func (v macAddressValidator) Description(ctx context.Context) string {
	return "value must be a valid MAC address"
}

func (v macAddressValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be a valid MAC address"
}

func (v macAddressValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	_, err := net.ParseMAC(value)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid MAC Address",
			"The value must be a valid MAC address (e.g., '00:11:22:33:44:55').",
		)
	}
}

// InterfaceNameValidator returns a validator that checks if the string is a valid RTX interface name.
func InterfaceNameValidator() validator.String {
	return &interfaceNameValidator{}
}

type interfaceNameValidator struct{}

func (v interfaceNameValidator) Description(ctx context.Context) string {
	return "value must be a valid RTX interface name"
}

func (v interfaceNameValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be a valid RTX interface name"
}

func (v interfaceNameValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	pattern := regexp.MustCompile(`^(lan[0-9]+|pp[0-9]+|tunnel[0-9]+|bridge[0-9]+|loopback[0-9]+|vlan[0-9]+)$`)
	if !pattern.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Interface Name",
			"The value must be a valid RTX interface name (e.g., 'lan1', 'pp1', 'tunnel1', 'bridge1').",
		)
	}
}

// prefixRefPattern matches the RTX's interface-relative prefix forms, the same two
// the ipv6 prefix parser accepts:
//
//	ra-prefix@<interface>::/<length>     the interface's RA-derived prefix
//	dhcp-prefix@<interface>::/<length>   the DHCPv6-PD delegated prefix
//
// See internal/rtx/parsers/ipv6_prefix.go, which builds and parses the same forms.
var prefixRefPattern = regexp.MustCompile(`^(?:ra|dhcp)-prefix@[A-Za-z0-9._-]+::/\d{1,3}$`)

// IPv6FilterAddressValidator returns a validator for the source/destination of an
// IPv6 filter: "*", a literal IPv6 address, an IPv6 CIDR, or one of the RTX's
// interface-relative prefix references.
//
// Without it these attributes are unconstrained strings and a typo passes straight
// through to the router. `dhcp-prefix@lan1::/64` where `lan2` was meant is accepted
// by the RTX, applies cleanly, shows no drift, and produces a filter that matches
// nothing — a stateful WAN filter that quietly stops admitting return traffic is
// exactly the failure this is meant to catch at plan time.
func IPv6FilterAddressValidator() validator.String {
	return &ipv6FilterAddressValidator{}
}

type ipv6FilterAddressValidator struct{}

func (v ipv6FilterAddressValidator) Description(ctx context.Context) string {
	return `value must be "*", an IPv6 address, an IPv6 CIDR, or ra-prefix@<interface>::/<length> / dhcp-prefix@<interface>::/<length>`
}

func (v ipv6FilterAddressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ipv6FilterAddressValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "*" || prefixRefPattern.MatchString(value) {
		return
	}

	if ip := net.ParseIP(value); ip != nil && ip.To4() == nil {
		return
	}

	if ip, _, err := net.ParseCIDR(value); err == nil && ip.To4() == nil {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid IPv6 Filter Address",
		`The value must be "*", an IPv6 address, an IPv6 CIDR, or an interface-relative prefix `+
			`reference such as "dhcp-prefix@lan2::/64" or "ra-prefix@lan2::/64". `+
			`Got: `+value,
	)
}
