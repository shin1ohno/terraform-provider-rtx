package client

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

func TestConvertFromParserL2TPConfig_TunnelAuth(t *testing.T) {
	tests := []struct {
		name             string
		parserConfig     parsers.L2TPConfig
		expectedEnabled  bool
		expectedPassword string
		expectNilAuth    bool
	}{
		{
			name: "tunnel auth enabled with password",
			parserConfig: parsers.L2TPConfig{
				ID:      1,
				Version: "l2tpv3",
				Mode:    "l2vpn",
				Enabled: true,
				L2TPv3Config: &parsers.L2TPv3Config{
					LocalRouterID:  "1.1.1.1",
					RemoteRouterID: "2.2.2.2",
					TunnelAuth: &parsers.L2TPTunnelAuth{
						Enabled:  true,
						Password: "secret123",
					},
				},
			},
			expectedEnabled:  true,
			expectedPassword: "secret123",
			expectNilAuth:    false,
		},
		{
			name: "tunnel auth enabled without password",
			parserConfig: parsers.L2TPConfig{
				ID:      1,
				Version: "l2tpv3",
				Mode:    "l2vpn",
				Enabled: true,
				L2TPv3Config: &parsers.L2TPv3Config{
					LocalRouterID:  "1.1.1.1",
					RemoteRouterID: "2.2.2.2",
					TunnelAuth: &parsers.L2TPTunnelAuth{
						Enabled:  true,
						Password: "",
					},
				},
			},
			expectedEnabled:  true,
			expectedPassword: "",
			expectNilAuth:    false,
		},
		{
			name: "tunnel auth disabled",
			parserConfig: parsers.L2TPConfig{
				ID:      1,
				Version: "l2tpv3",
				Mode:    "l2vpn",
				Enabled: true,
				L2TPv3Config: &parsers.L2TPv3Config{
					LocalRouterID:  "1.1.1.1",
					RemoteRouterID: "2.2.2.2",
					TunnelAuth: &parsers.L2TPTunnelAuth{
						Enabled:  false,
						Password: "",
					},
				},
			},
			expectedEnabled:  false,
			expectedPassword: "",
			expectNilAuth:    false,
		},
		{
			name: "no tunnel auth config",
			parserConfig: parsers.L2TPConfig{
				ID:      1,
				Version: "l2tpv3",
				Mode:    "l2vpn",
				Enabled: true,
				L2TPv3Config: &parsers.L2TPv3Config{
					LocalRouterID:  "1.1.1.1",
					RemoteRouterID: "2.2.2.2",
					TunnelAuth:     nil,
				},
			},
			expectNilAuth: true,
		},
		{
			name: "no L2TPv3 config",
			parserConfig: parsers.L2TPConfig{
				ID:           1,
				Version:      "l2tpv3",
				Mode:         "l2vpn",
				Enabled:      true,
				L2TPv3Config: nil,
			},
			expectNilAuth: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertFromParserL2TPConfig(tt.parserConfig)

			if tt.expectNilAuth {
				if tt.parserConfig.L2TPv3Config == nil {
					assert.Nil(t, result.L2TPv3Config)
				} else {
					assert.NotNil(t, result.L2TPv3Config)
					assert.Nil(t, result.L2TPv3Config.TunnelAuth)
				}
			} else {
				assert.NotNil(t, result.L2TPv3Config)
				assert.NotNil(t, result.L2TPv3Config.TunnelAuth)
				assert.Equal(t, tt.expectedEnabled, result.L2TPv3Config.TunnelAuth.Enabled)
				assert.Equal(t, tt.expectedPassword, result.L2TPv3Config.TunnelAuth.Password)
			}
		})
	}
}

func TestConvertToParserL2TPConfig_TunnelAuth(t *testing.T) {
	tests := []struct {
		name             string
		clientConfig     L2TPConfig
		expectedEnabled  bool
		expectedPassword string
		expectNilAuth    bool
	}{
		{
			name: "tunnel auth enabled with password",
			clientConfig: L2TPConfig{
				ID:      1,
				Version: "l2tpv3",
				Mode:    "l2vpn",
				Enabled: true,
				L2TPv3Config: &L2TPv3Config{
					LocalRouterID:  "1.1.1.1",
					RemoteRouterID: "2.2.2.2",
					TunnelAuth: &L2TPTunnelAuth{
						Enabled:  true,
						Password: "secret123",
					},
				},
			},
			expectedEnabled:  true,
			expectedPassword: "secret123",
			expectNilAuth:    false,
		},
		{
			name: "no tunnel auth config",
			clientConfig: L2TPConfig{
				ID:      1,
				Version: "l2tpv3",
				Mode:    "l2vpn",
				Enabled: true,
				L2TPv3Config: &L2TPv3Config{
					LocalRouterID:  "1.1.1.1",
					RemoteRouterID: "2.2.2.2",
					TunnelAuth:     nil,
				},
			},
			expectNilAuth: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToParserL2TPConfig(tt.clientConfig)

			if tt.expectNilAuth {
				assert.NotNil(t, result.L2TPv3Config)
				assert.Nil(t, result.L2TPv3Config.TunnelAuth)
			} else {
				assert.NotNil(t, result.L2TPv3Config)
				assert.NotNil(t, result.L2TPv3Config.TunnelAuth)
				assert.Equal(t, tt.expectedEnabled, result.L2TPv3Config.TunnelAuth.Enabled)
				assert.Equal(t, tt.expectedPassword, result.L2TPv3Config.TunnelAuth.Password)
			}
		})
	}
}
