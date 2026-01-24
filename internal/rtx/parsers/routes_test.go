package parsers

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRTX830RoutesParser_ParseRoutes(t *testing.T) {
	parser := &rtx830RoutesParser{
		BaseRoutesParser: BaseRoutesParser{
			modelPatterns: map[string]*regexp.Regexp{
				"route": regexp.MustCompile(`^([SCROBPD])\s+(\S+)\s+(?:via\s+(\S+))?\s*(?:dev\s+(\S+))?\s*(?:metric\s+(\d+))?`),
			},
		},
	}

	tests := []struct {
		name     string
		raw      string
		expected []Route
	}{
		{
			name: "Static route with metric",
			raw:  "S   0.0.0.0/0         via 192.168.1.1    dev LAN1 metric 1",
			expected: []Route{
				{
					Protocol:    "S",
					Destination: "0.0.0.0/0",
					Gateway:     "192.168.1.1",
					Interface:   "LAN1",
					Metric:      intPtr(1),
				},
			},
		},
		{
			name: "Connected route without gateway",
			raw:  "C   192.168.1.0/24    dev LAN1",
			expected: []Route{
				{
					Protocol:    "C",
					Destination: "192.168.1.0/24",
					Gateway:     "*",
					Interface:   "LAN1",
					Metric:      nil,
				},
			},
		},
		{
			name: "RIP route",
			raw:  "R   10.0.0.0/24       via 192.168.1.254  dev LAN1 metric 2",
			expected: []Route{
				{
					Protocol:    "R",
					Destination: "10.0.0.0/24",
					Gateway:     "192.168.1.254",
					Interface:   "LAN1",
					Metric:      intPtr(2),
				},
			},
		},
		{
			name: "Multiple routes",
			raw: `S   0.0.0.0/0         via 192.168.1.1    dev LAN1 metric 1
C   192.168.1.0/24    dev LAN1
R   10.0.0.0/24       via 192.168.1.254  dev LAN1 metric 2`,
			expected: []Route{
				{
					Protocol:    "S",
					Destination: "0.0.0.0/0",
					Gateway:     "192.168.1.1",
					Interface:   "LAN1",
					Metric:      intPtr(1),
				},
				{
					Protocol:    "C",
					Destination: "192.168.1.0/24",
					Gateway:     "*",
					Interface:   "LAN1",
					Metric:      nil,
				},
				{
					Protocol:    "R",
					Destination: "10.0.0.0/24",
					Gateway:     "192.168.1.254",
					Interface:   "LAN1",
					Metric:      intPtr(2),
				},
			},
		},
		{
			name:     "Empty output",
			raw:      "",
			expected: []Route{},
		},
		{
			name:     "Invalid format",
			raw:      "Invalid line\nAnother invalid line",
			expected: []Route{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseRoutes(tt.raw)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRTX12xxRoutesParser_ParseRoutes(t *testing.T) {
	parser := &rtx12xxRoutesParser{
		BaseRoutesParser: BaseRoutesParser{
			modelPatterns: map[string]*regexp.Regexp{
				"header": regexp.MustCompile(`^Destination\s+Gateway\s+Interface\s+Protocol\s+Metric`),
				"route":  regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+|-)$`),
			},
		},
	}

	tests := []struct {
		name     string
		raw      string
		expected []Route
	}{
		{
			name: "Standard RTX12xx format",
			raw: `Destination     Gateway         Interface   Protocol Metric
0.0.0.0/0       192.168.1.1     LAN1        S        1
192.168.1.0/24  *               LAN1        C        -
10.0.0.0/24     192.168.1.254   LAN1        R        2`,
			expected: []Route{
				{
					Destination: "0.0.0.0/0",
					Gateway:     "192.168.1.1",
					Interface:   "LAN1",
					Protocol:    "S",
					Metric:      intPtr(1),
				},
				{
					Destination: "192.168.1.0/24",
					Gateway:     "*",
					Interface:   "LAN1",
					Protocol:    "C",
					Metric:      nil,
				},
				{
					Destination: "10.0.0.0/24",
					Gateway:     "192.168.1.254",
					Interface:   "LAN1",
					Protocol:    "R",
					Metric:      intPtr(2),
				},
			},
		},
		{
			name: "OSPF route",
			raw: `Destination     Gateway         Interface   Protocol Metric
172.16.1.0/24   192.168.1.100   LAN1        O        110`,
			expected: []Route{
				{
					Destination: "172.16.1.0/24",
					Gateway:     "192.168.1.100",
					Interface:   "LAN1",
					Protocol:    "O",
					Metric:      intPtr(110),
				},
			},
		},
		{
			name: "BGP route",
			raw: `Destination     Gateway         Interface   Protocol Metric
172.16.2.0/24   192.168.1.101   LAN1        B        20`,
			expected: []Route{
				{
					Destination: "172.16.2.0/24",
					Gateway:     "192.168.1.101",
					Interface:   "LAN1",
					Protocol:    "B",
					Metric:      intPtr(20),
				},
			},
		},
		{
			name:     "Without header",
			raw:      `0.0.0.0/0       192.168.1.1     LAN1        S        1`,
			expected: []Route{}, // Should not parse without header
		},
		{
			name:     "Empty output",
			raw:      "",
			expected: []Route{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseRoutes(tt.raw)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoutesParser_Parse(t *testing.T) {
	rtx830Parser := &rtx830RoutesParser{
		BaseRoutesParser: BaseRoutesParser{
			modelPatterns: map[string]*regexp.Regexp{
				"route": regexp.MustCompile(`^([SCROBPD])\s+(\S+)\s+(?:via\s+(\S+))?\s*(?:dev\s+(\S+))?\s*(?:metric\s+(\d+))?`),
			},
		},
	}

	rtx12xxParser := &rtx12xxRoutesParser{
		BaseRoutesParser: BaseRoutesParser{
			modelPatterns: map[string]*regexp.Regexp{
				"header": regexp.MustCompile(`^Destination\s+Gateway\s+Interface\s+Protocol\s+Metric`),
				"route":  regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+|-)$`),
			},
		},
	}

	// Test RTX830 parser
	t.Run("RTX830_Parse", func(t *testing.T) {
		raw := "S   0.0.0.0/0         via 192.168.1.1    dev LAN1 metric 1"
		result, err := rtx830Parser.Parse(raw)
		assert.NoError(t, err)

		routes, ok := result.([]Route)
		assert.True(t, ok)
		assert.Len(t, routes, 1)
		assert.Equal(t, "S", routes[0].Protocol)
		assert.Equal(t, "0.0.0.0/0", routes[0].Destination)
		assert.Equal(t, "192.168.1.1", routes[0].Gateway)
	})

	// Test RTX12xx parser
	t.Run("RTX12xx_Parse", func(t *testing.T) {
		raw := `Destination     Gateway         Interface   Protocol Metric
0.0.0.0/0       192.168.1.1     LAN1        S        1`
		result, err := rtx12xxParser.Parse(raw)
		assert.NoError(t, err)

		routes, ok := result.([]Route)
		assert.True(t, ok)
		assert.Len(t, routes, 1)
		assert.Equal(t, "0.0.0.0/0", routes[0].Destination)
		assert.Equal(t, "192.168.1.1", routes[0].Gateway)
	})
}

func TestRoutesParser_CanHandle(t *testing.T) {
	rtx830Parser := &rtx830RoutesParser{}
	rtx12xxParser := &rtx12xxRoutesParser{}

	// Test RTX830
	assert.True(t, rtx830Parser.CanHandle("RTX830"))
	assert.False(t, rtx830Parser.CanHandle("RTX1210"))
	assert.False(t, rtx830Parser.CanHandle("RTX1220"))

	// Test RTX12xx
	assert.True(t, rtx12xxParser.CanHandle("RTX1210"))
	assert.True(t, rtx12xxParser.CanHandle("RTX1220"))
	assert.True(t, rtx12xxParser.CanHandle("RTX1200"))
	assert.False(t, rtx12xxParser.CanHandle("RTX830"))
}


