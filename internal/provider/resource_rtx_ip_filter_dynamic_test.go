package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func TestBuildIPFilterDynamicFromResourceData_Form1(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.IPFilterDynamic
		wantErr  bool
	}{
		{
			name: "basic Form 1 with protocol www",
			input: map[string]interface{}{
				"filter_id":       100,
				"source":          "*",
				"destination":     "*",
				"protocol":        "www",
				"syslog":          false,
				"filter_list":     []interface{}{},
				"in_filter_list":  []interface{}{},
				"out_filter_list": []interface{}{},
			},
			expected: client.IPFilterDynamic{
				Number:   100,
				Source:   "*",
				Dest:     "*",
				Protocol: "www",
				SyslogOn: false,
			},
			wantErr: false,
		},
		{
			name: "Form 1 with syslog enabled",
			input: map[string]interface{}{
				"filter_id":       200,
				"source":          "192.168.1.0/24",
				"destination":     "*",
				"protocol":        "ftp",
				"syslog":          true,
				"filter_list":     []interface{}{},
				"in_filter_list":  []interface{}{},
				"out_filter_list": []interface{}{},
			},
			expected: client.IPFilterDynamic{
				Number:   200,
				Source:   "192.168.1.0/24",
				Dest:     "*",
				Protocol: "ftp",
				SyslogOn: true,
			},
			wantErr: false,
		},
		{
			name: "Form 1 with smtp protocol",
			input: map[string]interface{}{
				"filter_id":       300,
				"source":          "*",
				"destination":     "10.0.0.1",
				"protocol":        "smtp",
				"syslog":          false,
				"filter_list":     []interface{}{},
				"in_filter_list":  []interface{}{},
				"out_filter_list": []interface{}{},
			},
			expected: client.IPFilterDynamic{
				Number:   300,
				Source:   "*",
				Dest:     "10.0.0.1",
				Protocol: "smtp",
				SyslogOn: false,
			},
			wantErr: false,
		},
		{
			name: "Form 1 with dns protocol",
			input: map[string]interface{}{
				"filter_id":       400,
				"source":          "*",
				"destination":     "*",
				"protocol":        "dns",
				"syslog":          false,
				"filter_list":     []interface{}{},
				"in_filter_list":  []interface{}{},
				"out_filter_list": []interface{}{},
			},
			expected: client.IPFilterDynamic{
				Number:   400,
				Source:   "*",
				Dest:     "*",
				Protocol: "dns",
				SyslogOn: false,
			},
			wantErr: false,
		},
		{
			name: "Form 1 with https protocol",
			input: map[string]interface{}{
				"filter_id":       500,
				"source":          "*",
				"destination":     "*",
				"protocol":        "https",
				"syslog":          true,
				"filter_list":     []interface{}{},
				"in_filter_list":  []interface{}{},
				"out_filter_list": []interface{}{},
			},
			expected: client.IPFilterDynamic{
				Number:   500,
				Source:   "*",
				Dest:     "*",
				Protocol: "https",
				SyslogOn: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceRTXIPFilterDynamic().Schema, tt.input)
			result, err := buildIPFilterDynamicFromResourceData(d)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Number, result.Number)
				assert.Equal(t, tt.expected.Source, result.Source)
				assert.Equal(t, tt.expected.Dest, result.Dest)
				assert.Equal(t, tt.expected.Protocol, result.Protocol)
				assert.Equal(t, tt.expected.SyslogOn, result.SyslogOn)
			}
		})
	}
}

func TestBuildIPFilterDynamicFromResourceData_Form2(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.IPFilterDynamic
		wantErr  bool
	}{
		{
			name: "basic Form 2 with filter_list only",
			input: map[string]interface{}{
				"filter_id":       100,
				"source":          "*",
				"destination":     "*",
				"protocol":        "",
				"syslog":          false,
				"filter_list":     []interface{}{1000, 1001},
				"in_filter_list":  []interface{}{},
				"out_filter_list": []interface{}{},
			},
			expected: client.IPFilterDynamic{
				Number:     100,
				Source:     "*",
				Dest:       "*",
				Protocol:   "",
				SyslogOn:   false,
				FilterList: []int{1000, 1001},
			},
			wantErr: false,
		},
		{
			name: "Form 2 with filter_list and in_filter_list",
			input: map[string]interface{}{
				"filter_id":       200,
				"source":          "192.168.1.0/24",
				"destination":     "*",
				"protocol":        "",
				"syslog":          false,
				"filter_list":     []interface{}{1000},
				"in_filter_list":  []interface{}{2000, 2001},
				"out_filter_list": []interface{}{},
			},
			expected: client.IPFilterDynamic{
				Number:       200,
				Source:       "192.168.1.0/24",
				Dest:         "*",
				Protocol:     "",
				SyslogOn:     false,
				FilterList:   []int{1000},
				InFilterList: []int{2000, 2001},
			},
			wantErr: false,
		},
		{
			name: "Form 2 with all filter lists",
			input: map[string]interface{}{
				"filter_id":       300,
				"source":          "*",
				"destination":     "10.0.0.0/8",
				"protocol":        "",
				"syslog":          true,
				"filter_list":     []interface{}{1000, 1001, 1002},
				"in_filter_list":  []interface{}{2000},
				"out_filter_list": []interface{}{3000, 3001},
			},
			expected: client.IPFilterDynamic{
				Number:        300,
				Source:        "*",
				Dest:          "10.0.0.0/8",
				Protocol:      "",
				SyslogOn:      true,
				FilterList:    []int{1000, 1001, 1002},
				InFilterList:  []int{2000},
				OutFilterList: []int{3000, 3001},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceRTXIPFilterDynamic().Schema, tt.input)
			result, err := buildIPFilterDynamicFromResourceData(d)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Number, result.Number)
				assert.Equal(t, tt.expected.Source, result.Source)
				assert.Equal(t, tt.expected.Dest, result.Dest)
				assert.Equal(t, tt.expected.SyslogOn, result.SyslogOn)
				assert.Equal(t, tt.expected.FilterList, result.FilterList)
				assert.Equal(t, tt.expected.InFilterList, result.InFilterList)
				assert.Equal(t, tt.expected.OutFilterList, result.OutFilterList)
			}
		})
	}
}

func TestBuildIPFilterDynamicFromResourceData_WithTimeout(t *testing.T) {
	input := map[string]interface{}{
		"filter_id":       100,
		"source":          "*",
		"destination":     "*",
		"protocol":        "www",
		"syslog":          false,
		"timeout":         60,
		"filter_list":     []interface{}{},
		"in_filter_list":  []interface{}{},
		"out_filter_list": []interface{}{},
	}

	d := schema.TestResourceDataRaw(t, resourceRTXIPFilterDynamic().Schema, input)
	result, err := buildIPFilterDynamicFromResourceData(d)
	assert.NoError(t, err)
	assert.Equal(t, 100, result.Number)
	assert.Equal(t, "www", result.Protocol)
	assert.NotNil(t, result.Timeout)
	assert.Equal(t, 60, *result.Timeout)
}

func TestBuildIPFilterDynamicFromResourceData_MissingProtocolAndFilterList(t *testing.T) {
	input := map[string]interface{}{
		"filter_id":       100,
		"source":          "*",
		"destination":     "*",
		"protocol":        "",
		"syslog":          false,
		"filter_list":     []interface{}{},
		"in_filter_list":  []interface{}{},
		"out_filter_list": []interface{}{},
	}

	d := schema.TestResourceDataRaw(t, resourceRTXIPFilterDynamic().Schema, input)
	_, err := buildIPFilterDynamicFromResourceData(d)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "either 'protocol' or 'filter_list' must be specified")
}

func TestFlattenIPFilterDynamicToResourceData_Form1(t *testing.T) {
	filter := &client.IPFilterDynamic{
		Number:   100,
		Source:   "*",
		Dest:     "*",
		Protocol: "www",
		SyslogOn: true,
	}

	input := map[string]interface{}{
		"filter_id":       0,
		"source":          "",
		"destination":     "",
		"protocol":        "",
		"syslog":          false,
		"filter_list":     []interface{}{},
		"in_filter_list":  []interface{}{},
		"out_filter_list": []interface{}{},
	}

	d := schema.TestResourceDataRaw(t, resourceRTXIPFilterDynamic().Schema, input)

	err := flattenIPFilterDynamicToResourceData(filter, d)
	assert.NoError(t, err)
	assert.Equal(t, 100, d.Get("filter_id"))
	assert.Equal(t, "*", d.Get("source"))
	assert.Equal(t, "*", d.Get("destination"))
	assert.Equal(t, "www", d.Get("protocol"))
	assert.Equal(t, true, d.Get("syslog"))
}

func TestFlattenIPFilterDynamicToResourceData_Form2(t *testing.T) {
	filter := &client.IPFilterDynamic{
		Number:        200,
		Source:        "192.168.1.0/24",
		Dest:          "*",
		Protocol:      "",
		SyslogOn:      false,
		FilterList:    []int{1000, 1001},
		InFilterList:  []int{2000},
		OutFilterList: []int{3000},
	}

	input := map[string]interface{}{
		"filter_id":       0,
		"source":          "",
		"destination":     "",
		"protocol":        "",
		"syslog":          false,
		"filter_list":     []interface{}{},
		"in_filter_list":  []interface{}{},
		"out_filter_list": []interface{}{},
	}

	d := schema.TestResourceDataRaw(t, resourceRTXIPFilterDynamic().Schema, input)

	err := flattenIPFilterDynamicToResourceData(filter, d)
	assert.NoError(t, err)
	assert.Equal(t, 200, d.Get("filter_id"))
	assert.Equal(t, "192.168.1.0/24", d.Get("source"))
	assert.Equal(t, "*", d.Get("destination"))
	assert.Equal(t, "", d.Get("protocol"))
	assert.Equal(t, false, d.Get("syslog"))

	filterList := d.Get("filter_list").([]interface{})
	assert.Len(t, filterList, 2)
	assert.Equal(t, 1000, filterList[0])
	assert.Equal(t, 1001, filterList[1])

	inFilterList := d.Get("in_filter_list").([]interface{})
	assert.Len(t, inFilterList, 1)
	assert.Equal(t, 2000, inFilterList[0])

	outFilterList := d.Get("out_filter_list").([]interface{})
	assert.Len(t, outFilterList, 1)
	assert.Equal(t, 3000, outFilterList[0])
}

func TestFlattenIPFilterDynamicToResourceData_WithTimeout(t *testing.T) {
	timeout := 120
	filter := &client.IPFilterDynamic{
		Number:   100,
		Source:   "*",
		Dest:     "*",
		Protocol: "ftp",
		SyslogOn: false,
		Timeout:  &timeout,
	}

	input := map[string]interface{}{
		"filter_id":       0,
		"source":          "",
		"destination":     "",
		"protocol":        "",
		"syslog":          false,
		"filter_list":     []interface{}{},
		"in_filter_list":  []interface{}{},
		"out_filter_list": []interface{}{},
	}

	d := schema.TestResourceDataRaw(t, resourceRTXIPFilterDynamic().Schema, input)

	err := flattenIPFilterDynamicToResourceData(filter, d)
	assert.NoError(t, err)
	assert.Equal(t, 120, d.Get("timeout"))
}

func TestExpandIntList(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []int
	}{
		{
			name:     "empty list",
			input:    []interface{}{},
			expected: nil, // expandIntList returns nil for empty input
		},
		{
			name:     "single element",
			input:    []interface{}{100},
			expected: []int{100},
		},
		{
			name:     "multiple elements",
			input:    []interface{}{100, 200, 300},
			expected: []int{100, 200, 300},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandIntList(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResourceRTXIPFilterDynamicSchema(t *testing.T) {
	resource := resourceRTXIPFilterDynamic()

	// Verify required fields
	assert.NotNil(t, resource.Schema["filter_id"])
	assert.True(t, resource.Schema["filter_id"].Required)
	assert.True(t, resource.Schema["filter_id"].ForceNew)

	assert.NotNil(t, resource.Schema["source"])
	assert.True(t, resource.Schema["source"].Required)

	assert.NotNil(t, resource.Schema["destination"])
	assert.True(t, resource.Schema["destination"].Required)

	// Verify optional fields
	assert.NotNil(t, resource.Schema["protocol"])
	assert.True(t, resource.Schema["protocol"].Optional)

	assert.NotNil(t, resource.Schema["filter_list"])
	assert.True(t, resource.Schema["filter_list"].Optional)

	assert.NotNil(t, resource.Schema["in_filter_list"])
	assert.True(t, resource.Schema["in_filter_list"].Optional)

	assert.NotNil(t, resource.Schema["out_filter_list"])
	assert.True(t, resource.Schema["out_filter_list"].Optional)

	assert.NotNil(t, resource.Schema["syslog"])
	assert.True(t, resource.Schema["syslog"].Optional)
	assert.Equal(t, false, resource.Schema["syslog"].Default)

	assert.NotNil(t, resource.Schema["timeout"])
	assert.True(t, resource.Schema["timeout"].Optional)

	// Verify ConflictsWith settings
	assert.Contains(t, resource.Schema["protocol"].ConflictsWith, "filter_list")
	assert.Contains(t, resource.Schema["protocol"].ConflictsWith, "in_filter_list")
	assert.Contains(t, resource.Schema["protocol"].ConflictsWith, "out_filter_list")

	assert.Contains(t, resource.Schema["filter_list"].ConflictsWith, "protocol")
	assert.Contains(t, resource.Schema["in_filter_list"].ConflictsWith, "protocol")
	assert.Contains(t, resource.Schema["out_filter_list"].ConflictsWith, "protocol")
}
