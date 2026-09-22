package kron_schedule

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// validate runs ValidateConfig against a config where every attribute not in
// vals is null, the way Terraform hands over an attribute the practitioner
// left out.
func validate(t *testing.T, vals map[string]tftypes.Value) diag.Diagnostics {
	t.Helper()
	ctx := context.Background()
	r := &KronScheduleResource{}

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("schema type is not an object")
	}

	full := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, typ := range objType.AttributeTypes {
		if v, ok := vals[name]; ok {
			full[name] = v
		} else {
			full[name] = tftypes.NewValue(typ, nil)
		}
	}

	req := resource.ValidateConfigRequest{
		Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: tftypes.NewValue(objType, full)},
	}
	var resp resource.ValidateConfigResponse
	r.ValidateConfig(ctx, req, &resp)
	return resp.Diagnostics
}

var (
	unknownString = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	unknownBool   = tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue)
	listOfString  = tftypes.List{ElementType: tftypes.String}
	unknownList   = tftypes.NewValue(listOfString, tftypes.UnknownValue)
)

func commands(cmds ...string) tftypes.Value {
	elems := make([]tftypes.Value, len(cmds))
	for i, c := range cmds {
		elems[i] = tftypes.NewValue(tftypes.String, c)
	}
	return tftypes.NewValue(listOfString, elems)
}

func errorSummaries(diags diag.Diagnostics) string {
	var out []string
	for _, d := range diags.Errors() {
		out = append(out, d.Detail())
	}
	return strings.Join(out, "; ")
}

// Terraform validates before variables and for_each values are known, so a
// computed schedule reaches ValidateConfig with unknown attributes. Those
// used to be read as "not set" and rejected.
func TestValidateConfigAcceptsUnknownValues(t *testing.T) {
	cases := map[string]map[string]tftypes.Value{
		"unknown at_time (for_each)": {
			"schedule_id":   tftypes.NewValue(tftypes.Number, 100),
			"at_time":       unknownString,
			"command_lines": commands("ethernet lan1 filter in 100"),
		},
		"unknown command_lines (built from a local)": {
			"schedule_id":   tftypes.NewValue(tftypes.Number, 100),
			"at_time":       tftypes.NewValue(tftypes.String, "21:00"),
			"command_lines": unknownList,
		},
		"known list with an unknown element": {
			"schedule_id":   tftypes.NewValue(tftypes.Number, 100),
			"at_time":       unknownString,
			"command_lines": tftypes.NewValue(listOfString, []tftypes.Value{unknownString}),
		},
		"unknown policy_list": {
			"schedule_id": tftypes.NewValue(tftypes.Number, 100),
			"at_time":     tftypes.NewValue(tftypes.String, "21:00"),
			"policy_list": unknownString,
		},
		"explicit recurring=false with unknown on_startup": {
			"schedule_id":   tftypes.NewValue(tftypes.Number, 100),
			"on_startup":    unknownBool,
			"recurring":     tftypes.NewValue(tftypes.Bool, false),
			"command_lines": commands("ntpdate ntp.nict.jp"),
		},
	}

	for name, vals := range cases {
		t.Run(name, func(t *testing.T) {
			if diags := validate(t, vals); diags.HasError() {
				t.Errorf("want no error, got: %s", errorSummaries(diags))
			}
		})
	}
}

// The guards must not swallow the checks once the values are known.
func TestValidateConfigStillRejectsKnownGaps(t *testing.T) {
	cases := map[string]struct {
		vals map[string]tftypes.Value
		want string
	}{
		"no timing at all": {
			vals: map[string]tftypes.Value{
				"schedule_id":   tftypes.NewValue(tftypes.Number, 100),
				"command_lines": commands("ntpdate ntp.nict.jp"),
			},
			want: "One of 'at_time', 'on_startup', or 'date' must be specified.",
		},
		"no command at all": {
			vals: map[string]tftypes.Value{
				"schedule_id": tftypes.NewValue(tftypes.Number, 100),
				"at_time":     tftypes.NewValue(tftypes.String, "21:00"),
			},
			want: "Either 'policy_list' or 'command_lines' must be specified.",
		},
		"empty command list": {
			vals: map[string]tftypes.Value{
				"schedule_id":   tftypes.NewValue(tftypes.Number, 100),
				"at_time":       tftypes.NewValue(tftypes.String, "21:00"),
				"command_lines": commands(),
			},
			want: "Either 'policy_list' or 'command_lines' must be specified.",
		},
		"recurring=false on a plain daily schedule": {
			vals: map[string]tftypes.Value{
				"schedule_id":   tftypes.NewValue(tftypes.Number, 100),
				"at_time":       tftypes.NewValue(tftypes.String, "21:00"),
				"recurring":     tftypes.NewValue(tftypes.Bool, false),
				"command_lines": commands("ntpdate ntp.nict.jp"),
			},
			want: "recurring cannot be false",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			diags := validate(t, tc.vals)
			if !diags.HasError() {
				t.Fatalf("want error containing %q, got none", tc.want)
			}
			if got := errorSummaries(diags); !strings.Contains(got, tc.want) {
				t.Errorf("want error containing %q, got: %s", tc.want, got)
			}
		})
	}
}
