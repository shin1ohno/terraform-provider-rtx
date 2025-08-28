package parsers

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestDhcpScopeParsers(t *testing.T) {
	tests := []struct {
		name       string
		model      string
		configLine string
		want       *DhcpScope
		wantErr    bool
	}{
		// 正常系: 基本設定（ID、レンジ、プレフィックスのみ）
		{
			name:       "RTX830 basic scope",
			model:      "RTX830",
			configLine: "dhcp scope 1 192.168.100.2-192.168.100.191/24",
			want: &DhcpScope{
				ID:         1,
				RangeStart: "192.168.100.2",
				RangeEnd:   "192.168.100.191",
				Prefix:     24,
			},
			wantErr: false,
		},
		{
			name:       "RTX1210 basic scope",
			model:      "RTX1210",
			configLine: "dhcp scope 2 10.0.0.100-10.0.0.200/16",
			want: &DhcpScope{
				ID:         2,
				RangeStart: "10.0.0.100",
				RangeEnd:   "10.0.0.200",
				Prefix:     16,
			},
			wantErr: false,
		},
		// 正常系: 全オプション指定
		{
			name:       "RTX830 full options",
			model:      "RTX830",
			configLine: "dhcp scope 1 192.168.100.2-192.168.100.191/24 gateway 192.168.100.1 dns 8.8.8.8 8.8.4.4 lease 7 domain example.com",
			want: &DhcpScope{
				ID:         1,
				RangeStart: "192.168.100.2",
				RangeEnd:   "192.168.100.191",
				Prefix:     24,
				Gateway:    "192.168.100.1",
				DNSServers: []string{"8.8.8.8", "8.8.4.4"},
				Lease:      7,
				DomainName: "example.com",
			},
			wantErr: false,
		},
		{
			name:       "RTX1210 full options",
			model:      "RTX1210",
			configLine: "dhcp scope 3 172.16.0.10-172.16.0.50/20 gateway 172.16.0.1 dns 1.1.1.1 lease 24 domain test.local",
			want: &DhcpScope{
				ID:         3,
				RangeStart: "172.16.0.10",
				RangeEnd:   "172.16.0.50",
				Prefix:     20,
				Gateway:    "172.16.0.1",
				DNSServers: []string{"1.1.1.1"},
				Lease:      24,
				DomainName: "test.local",
			},
			wantErr: false,
		},
		// 正常系: DNSサーバー複数指定
		{
			name:       "RTX830 multiple DNS",
			model:      "RTX830",
			configLine: "dhcp scope 5 192.168.1.50-192.168.1.100/24 dns 8.8.8.8 8.8.4.4 1.1.1.1",
			want: &DhcpScope{
				ID:         5,
				RangeStart: "192.168.1.50",
				RangeEnd:   "192.168.1.100",
				Prefix:     24,
				DNSServers: []string{"8.8.8.8", "8.8.4.4", "1.1.1.1"},
			},
			wantErr: false,
		},
		// 正常系: オプションの順序が異なる場合
		{
			name:       "RTX1210 options different order",
			model:      "RTX1210",
			configLine: "dhcp scope 4 10.1.1.10-10.1.1.20/28 domain corp.local lease 12 gateway 10.1.1.1 dns 192.168.1.1",
			want: &DhcpScope{
				ID:         4,
				RangeStart: "10.1.1.10",
				RangeEnd:   "10.1.1.20",
				Prefix:     28,
				Gateway:    "10.1.1.1",
				DNSServers: []string{"192.168.1.1"},
				Lease:      12,
				DomainName: "corp.local",
			},
			wantErr: false,
		},
		// 正常系: expire with hh:mm format (実機出力形式)
		{
			name:       "RTX830 expire with time format",
			model:      "RTX830",
			configLine: "dhcp scope 1 192.168.1.20-192.168.1.99/24 gateway 192.168.1.253 expire 12:00",
			want: &DhcpScope{
				ID:         1,
				RangeStart: "192.168.1.20",
				RangeEnd:   "192.168.1.99",
				Prefix:     24,
				Gateway:    "192.168.1.253",
				Lease:      43200, // 12:00 = 12*60*60 seconds
			},
			wantErr: false,
		},
		// 正常系: expire with ma option (実機出力形式)
		{
			name:       "RTX830 expire with ma option",
			model:      "RTX830",
			configLine: "dhcp scope 1 192.168.1.20-192.168.1.99/16 gateway 192.168.1.253 expire 12:00 ma",
			want: &DhcpScope{
				ID:         1,
				RangeStart: "192.168.1.20",
				RangeEnd:   "192.168.1.99",
				Prefix:     16,
				Gateway:    "192.168.1.253",
				Lease:      43200, // 12:00 = 12*60*60 seconds
			},
			wantErr: false,
		},
		// 正常系: maxexpire option
		{
			name:       "RTX830 with maxexpire",
			model:      "RTX830",
			configLine: "dhcp scope 2 10.0.0.10-10.0.0.50/24 expire 60 maxexpire 1440",
			want: &DhcpScope{
				ID:         2,
				RangeStart: "10.0.0.10",
				RangeEnd:   "10.0.0.50",
				Prefix:     24,
				Lease:      3600, // 60 minutes = 3600 seconds
			},
			wantErr: false,
		},
		// 異常系: 不正な設定行フォーマット
		{
			name:       "invalid format - missing dhcp keyword",
			model:      "RTX830",
			configLine: "scope 1 192.168.1.1-192.168.1.10/24",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "invalid format - missing scope keyword",
			model:      "RTX830",
			configLine: "dhcp 1 192.168.1.1-192.168.1.10/24",
			want:       nil,
			wantErr:    true,
		},
		// 異常系: 存在しないスコープID
		{
			name:       "invalid scope ID - non-numeric",
			model:      "RTX830",
			configLine: "dhcp scope abc 192.168.1.1-192.168.1.10/24",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "invalid scope ID - zero",
			model:      "RTX830",
			configLine: "dhcp scope 0 192.168.1.1-192.168.1.10/24",
			want:       nil,
			wantErr:    true,
		},
		// 異常系: IPアドレス範囲の形式エラー
		{
			name:       "invalid IP range - missing dash",
			model:      "RTX830",
			configLine: "dhcp scope 1 192.168.1.1 192.168.1.10/24",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "invalid IP range - invalid start IP",
			model:      "RTX830",
			configLine: "dhcp scope 1 999.999.999.999-192.168.1.10/24",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "invalid IP range - invalid end IP",
			model:      "RTX830",
			configLine: "dhcp scope 1 192.168.1.1-999.999.999.999/24",
			want:       nil,
			wantErr:    true,
		},
		// 異常系: 不正なプレフィックス長
		{
			name:       "invalid prefix - missing slash",
			model:      "RTX830",
			configLine: "dhcp scope 1 192.168.1.1-192.168.1.10 24",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "invalid prefix - out of range",
			model:      "RTX830",
			configLine: "dhcp scope 1 192.168.1.1-192.168.1.10/33",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "invalid prefix - negative",
			model:      "RTX830",
			configLine: "dhcp scope 1 192.168.1.1-192.168.1.10/-1",
			want:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDhcpScope(tt.configLine)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDhcpScope() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				gotJSON, _ := json.MarshalIndent(got, "", "  ")
				wantJSON, _ := json.MarshalIndent(tt.want, "", "  ")
				t.Errorf("ParseDhcpScope() got:\n%s\nwant:\n%s", gotJSON, wantJSON)
			}
		})
	}
}

func TestDhcpScopeParserCanHandle(t *testing.T) {
	tests := []struct {
		parser    DhcpScopeParser
		model     string
		canHandle bool
	}{
		{&rtx830DhcpScopeParser{}, "RTX830", true},
		{&rtx830DhcpScopeParser{}, "RTX1210", false},
		{&rtx12xxDhcpScopeParser{}, "RTX1210", true},
		{&rtx12xxDhcpScopeParser{}, "RTX1220", true},
		{&rtx12xxDhcpScopeParser{}, "RTX830", false},
	}

	for _, tt := range tests {
		name := reflect.TypeOf(tt.parser).Elem().Name() + "/" + tt.model
		t.Run(name, func(t *testing.T) {
			got := tt.parser.CanHandle(tt.model)
			if got != tt.canHandle {
				t.Errorf("CanHandle(%s) = %v, want %v", tt.model, got, tt.canHandle)
			}
		})
	}
}

func TestDhcpScopeParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		parser  DhcpScopeParser
		raw     string
		want    []*DhcpScope
		wantErr bool
	}{
		{
			name:   "RTX830 multiple scopes",
			parser: &rtx830DhcpScopeParser{},
			raw: `dhcp scope 1 192.168.100.2-192.168.100.191/24 gateway 192.168.100.1 dns 8.8.8.8
dhcp scope 2 10.0.0.10-10.0.0.20/16 lease 12`,
			want: []*DhcpScope{
				{
					ID:         1,
					RangeStart: "192.168.100.2",
					RangeEnd:   "192.168.100.191",
					Prefix:     24,
					Gateway:    "192.168.100.1",
					DNSServers: []string{"8.8.8.8"},
				},
				{
					ID:         2,
					RangeStart: "10.0.0.10",
					RangeEnd:   "10.0.0.20",
					Prefix:     16,
					Lease:      12,
				},
			},
			wantErr: false,
		},
		{
			name:   "RTX1210 single scope",
			parser: &rtx12xxDhcpScopeParser{},
			raw:    "dhcp scope 1 172.16.0.100-172.16.0.200/24 gateway 172.16.0.1 dns 1.1.1.1 1.0.0.1 lease 24 domain example.org",
			want: []*DhcpScope{
				{
					ID:         1,
					RangeStart: "172.16.0.100",
					RangeEnd:   "172.16.0.200",
					Prefix:     24,
					Gateway:    "172.16.0.1",
					DNSServers: []string{"1.1.1.1", "1.0.0.1"},
					Lease:      24,
					DomainName: "example.org",
				},
			},
			wantErr: false,
		},
		{
			name:    "empty input",
			parser:  &rtx830DhcpScopeParser{},
			raw:     "",
			want:    []*DhcpScope{},
			wantErr: false,
		},
		{
			name:    "no dhcp scope lines",
			parser:  &rtx830DhcpScopeParser{},
			raw:     "some other configuration line\nanother line without dhcp scope",
			want:    []*DhcpScope{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.parser.Parse(tt.raw)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			scopes, ok := result.([]*DhcpScope)
			if !ok {
				t.Fatalf("expected []*DhcpScope, got %T", result)
			}

			if !reflect.DeepEqual(scopes, tt.want) {
				gotJSON, _ := json.MarshalIndent(scopes, "", "  ")
				wantJSON, _ := json.MarshalIndent(tt.want, "", "  ")
				t.Errorf("Parse() got:\n%s\nwant:\n%s", gotJSON, wantJSON)
			}
		})
	}
}
