package fwhelpers

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestIntSliceToList_Nil(t *testing.T) {
	result := IntSliceToList(nil)
	assert.True(t, result.IsNull(), "nil input should produce ListNull")
}

func TestIntSliceToList_Empty(t *testing.T) {
	result := IntSliceToList([]int{})
	assert.False(t, result.IsNull(), "empty slice should NOT produce ListNull")
	assert.Equal(t, 0, len(result.Elements()), "empty slice should produce empty list")
}

func TestIntSliceToList_Populated(t *testing.T) {
	result := IntSliceToList([]int{1, 2, 3})
	assert.False(t, result.IsNull())
	elements := result.Elements()
	assert.Equal(t, 3, len(elements))
	assert.Equal(t, int64(1), elements[0].(types.Int64).ValueInt64())
	assert.Equal(t, int64(2), elements[1].(types.Int64).ValueInt64())
	assert.Equal(t, int64(3), elements[2].(types.Int64).ValueInt64())
}

func TestListToIntSlice_Null(t *testing.T) {
	result := ListToIntSlice(types.ListNull(types.Int64Type))
	assert.Nil(t, result, "null list should produce nil")
}

func TestListToIntSlice_Empty(t *testing.T) {
	emptyList, _ := types.ListValue(types.Int64Type, []attr.Value{})
	result := ListToIntSlice(emptyList)
	assert.NotNil(t, result, "empty list should NOT produce nil")
	assert.Equal(t, 0, len(result), "empty list should produce empty slice")
}

func TestListToIntSlice_Populated(t *testing.T) {
	list := IntSliceToList([]int{10, 20, 30})
	result := ListToIntSlice(list)
	assert.Equal(t, []int{10, 20, 30}, result)
}

func TestStringSliceToList_Nil(t *testing.T) {
	result := StringSliceToList(nil)
	assert.True(t, result.IsNull(), "nil input should produce ListNull")
}

func TestStringSliceToList_Empty(t *testing.T) {
	result := StringSliceToList([]string{})
	assert.False(t, result.IsNull(), "empty slice should NOT produce ListNull")
	assert.Equal(t, 0, len(result.Elements()), "empty slice should produce empty list")
}

func TestStringSliceToList_Populated(t *testing.T) {
	result := StringSliceToList([]string{"a", "b", "c"})
	assert.False(t, result.IsNull())
	elements := result.Elements()
	assert.Equal(t, 3, len(elements))
	assert.Equal(t, "a", elements[0].(types.String).ValueString())
	assert.Equal(t, "b", elements[1].(types.String).ValueString())
	assert.Equal(t, "c", elements[2].(types.String).ValueString())
}

func TestListToStringSlice_Null(t *testing.T) {
	result := ListToStringSlice(types.ListNull(types.StringType))
	assert.Nil(t, result, "null list should produce nil")
}

func TestListToStringSlice_Empty(t *testing.T) {
	emptyList, _ := types.ListValue(types.StringType, []attr.Value{})
	result := ListToStringSlice(emptyList)
	assert.NotNil(t, result, "empty list should NOT produce nil")
	assert.Equal(t, 0, len(result), "empty list should produce empty slice")
}

func TestListToStringSlice_Populated(t *testing.T) {
	list := StringSliceToList([]string{"hello", "world"})
	result := ListToStringSlice(list)
	assert.Equal(t, []string{"hello", "world"}, result)
}

// Round-trip tests to verify nil vs empty preservation
func TestIntSlice_RoundTrip_Nil(t *testing.T) {
	list := IntSliceToList(nil)
	result := ListToIntSlice(list)
	assert.Nil(t, result, "nil → ListNull → nil round-trip should preserve nil")
}

func TestIntSlice_RoundTrip_Empty(t *testing.T) {
	list := IntSliceToList([]int{})
	result := ListToIntSlice(list)
	assert.NotNil(t, result, "[]int{} → empty ListValue → []int{} round-trip should preserve non-nil")
	assert.Equal(t, 0, len(result))
}

func TestStringSlice_RoundTrip_Nil(t *testing.T) {
	list := StringSliceToList(nil)
	result := ListToStringSlice(list)
	assert.Nil(t, result, "nil → ListNull → nil round-trip should preserve nil")
}

func TestStringSlice_RoundTrip_Empty(t *testing.T) {
	list := StringSliceToList([]string{})
	result := ListToStringSlice(list)
	assert.NotNil(t, result, "[]string{} → empty ListValue → []string{} round-trip should preserve non-nil")
	assert.Equal(t, 0, len(result))
}
