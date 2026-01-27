package shape

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// ShapeModel describes the resource data model.
type ShapeModel struct {
	ID           types.String `tfsdk:"id"`
	Interface    types.String `tfsdk:"interface"`
	Direction    types.String `tfsdk:"direction"`
	ShapeAverage types.Int64  `tfsdk:"shape_average"`
	ShapeBurst   types.Int64  `tfsdk:"shape_burst"`
}

// ToClient converts the Terraform model to a client.ShapeConfig.
func (m *ShapeModel) ToClient() client.ShapeConfig {
	return client.ShapeConfig{
		Interface:    fwhelpers.GetStringValue(m.Interface),
		Direction:    fwhelpers.GetStringValue(m.Direction),
		ShapeAverage: fwhelpers.GetInt64Value(m.ShapeAverage),
		ShapeBurst:   fwhelpers.GetInt64Value(m.ShapeBurst),
	}
}

// FromClient updates the Terraform model from a client.ShapeConfig.
func (m *ShapeModel) FromClient(sc *client.ShapeConfig) {
	m.Interface = types.StringValue(sc.Interface)
	m.Direction = types.StringValue(sc.Direction)
	m.ShapeAverage = types.Int64Value(int64(sc.ShapeAverage))
	if sc.ShapeBurst > 0 {
		m.ShapeBurst = types.Int64Value(int64(sc.ShapeBurst))
	} else {
		m.ShapeBurst = types.Int64Null()
	}
}
