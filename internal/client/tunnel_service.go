package client

import (
	"context"
	"fmt"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// TunnelService handles unified tunnel configuration operations
type TunnelService struct {
	executor Executor
	client   *rtxClient
}

// NewTunnelService creates a new unified tunnel service
func NewTunnelService(executor Executor, client *rtxClient) *TunnelService {
	return &TunnelService{
		executor: executor,
		client:   client,
	}
}

// Get retrieves a specific unified tunnel configuration
func (s *TunnelService) Get(ctx context.Context, tunnelID int) (*Tunnel, error) {
	tunnels, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, tunnel := range tunnels {
		if tunnel.ID == tunnelID {
			return &tunnel, nil
		}
	}

	return nil, fmt.Errorf("tunnel %d not found", tunnelID)
}

// List retrieves all unified tunnel configurations
func (s *TunnelService) List(ctx context.Context) ([]Tunnel, error) {
	output, err := s.executor.Run(ctx, "show config")
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	parser := parsers.NewTunnelParser()
	parsed, err := parser.ParseTunnelConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse tunnel config: %w", err)
	}

	// Convert from parser types to client types
	tunnels := make([]Tunnel, len(parsed))
	for i, p := range parsed {
		tunnels[i] = convertFromParserTunnel(p)
	}

	return tunnels, nil
}

// Create creates a new unified tunnel
func (s *TunnelService) Create(ctx context.Context, tunnel Tunnel) error {
	// Validate configuration
	parserTunnel := convertToParserTunnel(tunnel)
	if err := parsers.ValidateTunnel(parserTunnel); err != nil {
		return fmt.Errorf("invalid tunnel config: %w", err)
	}

	// Build commands
	commands := parsers.BuildTunnelCommands(parserTunnel)

	// Execute all commands in batch
	output, err := s.executor.RunBatch(ctx, commands)
	if err != nil {
		return fmt.Errorf("failed to execute tunnel batch commands: %w", err)
	}
	if containsError(string(output)) {
		return fmt.Errorf("tunnel batch commands failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("failed to save tunnel config: %w", err)
		}
	}

	return nil
}

// Update modifies an existing unified tunnel
func (s *TunnelService) Update(ctx context.Context, tunnel Tunnel) error {
	// Validate configuration
	parserTunnel := convertToParserTunnel(tunnel)
	if err := parsers.ValidateTunnel(parserTunnel); err != nil {
		return fmt.Errorf("invalid tunnel config: %w", err)
	}

	// For update, we rebuild the entire tunnel configuration
	// This ensures all settings are correctly applied
	commands := parsers.BuildTunnelCommands(parserTunnel)

	// Execute all commands in batch
	output, err := s.executor.RunBatch(ctx, commands)
	if err != nil {
		return fmt.Errorf("failed to execute tunnel batch commands: %w", err)
	}
	if containsError(string(output)) {
		return fmt.Errorf("tunnel batch commands failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("failed to save tunnel config: %w", err)
		}
	}

	return nil
}

// Delete removes a unified tunnel
func (s *TunnelService) Delete(ctx context.Context, tunnelID int) error {
	commands := parsers.BuildDeleteTunnelCommands(tunnelID)

	// Execute all commands in batch
	output, err := s.executor.RunBatch(ctx, commands)
	if err != nil {
		return fmt.Errorf("failed to execute tunnel delete batch commands: %w", err)
	}
	if containsError(string(output)) {
		return fmt.Errorf("tunnel delete batch commands failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("failed to save config after tunnel delete: %w", err)
		}
	}

	return nil
}

// convertToParserTunnel converts client.Tunnel to parsers.Tunnel
func convertToParserTunnel(tunnel Tunnel) parsers.Tunnel {
	result := parsers.Tunnel{
		ID:               tunnel.ID,
		Encapsulation:    tunnel.Encapsulation,
		Enabled:          tunnel.Enabled,
		Name:             tunnel.Name,
		EndpointName:     tunnel.EndpointName,
		EndpointNameType: tunnel.EndpointNameType,
	}

	// Convert IPsec block
	if tunnel.IPsec != nil {
		result.IPsec = &parsers.TunnelIPsec{
			IPsecTunnelID:     tunnel.IPsec.IPsecTunnelID,
			LocalAddress:      tunnel.IPsec.LocalAddress,
			RemoteAddress:     tunnel.IPsec.RemoteAddress,
			PreSharedKey:      tunnel.IPsec.PreSharedKey,
			NATTraversal:      tunnel.IPsec.NATTraversal,
			IKERemoteName:     tunnel.IPsec.IKERemoteName,
			IKERemoteNameType: tunnel.IPsec.IKERemoteNameType,
			IKEKeepaliveLog:   tunnel.IPsec.IKEKeepaliveLog,
			IKELog:            tunnel.IPsec.IKELog,
			SecureFilterIn:    tunnel.IPsec.SecureFilterIn,
			SecureFilterOut:   tunnel.IPsec.SecureFilterOut,
			TCPMSSLimit:       tunnel.IPsec.TCPMSSLimit,
			IKEv2Proposal: parsers.IKEv2Proposal{
				EncryptionAES256: tunnel.IPsec.IKEv2Proposal.EncryptionAES256,
				EncryptionAES128: tunnel.IPsec.IKEv2Proposal.EncryptionAES128,
				Encryption3DES:   tunnel.IPsec.IKEv2Proposal.Encryption3DES,
				IntegritySHA256:  tunnel.IPsec.IKEv2Proposal.IntegritySHA256,
				IntegritySHA1:    tunnel.IPsec.IKEv2Proposal.IntegritySHA1,
				IntegrityMD5:     tunnel.IPsec.IKEv2Proposal.IntegrityMD5,
				GroupFourteen:    tunnel.IPsec.IKEv2Proposal.GroupFourteen,
				GroupFive:        tunnel.IPsec.IKEv2Proposal.GroupFive,
				GroupTwo:         tunnel.IPsec.IKEv2Proposal.GroupTwo,
				LifetimeSeconds:  tunnel.IPsec.IKEv2Proposal.LifetimeSeconds,
			},
			Transform: parsers.IPsecTransform{
				Protocol:         tunnel.IPsec.Transform.Protocol,
				EncryptionAES256: tunnel.IPsec.Transform.EncryptionAES256,
				EncryptionAES128: tunnel.IPsec.Transform.EncryptionAES128,
				Encryption3DES:   tunnel.IPsec.Transform.Encryption3DES,
				IntegritySHA256:  tunnel.IPsec.Transform.IntegritySHA256,
				IntegritySHA1:    tunnel.IPsec.Transform.IntegritySHA1,
				IntegrityMD5:     tunnel.IPsec.Transform.IntegrityMD5,
				PFSGroupFourteen: tunnel.IPsec.Transform.PFSGroupFourteen,
				PFSGroupFive:     tunnel.IPsec.Transform.PFSGroupFive,
				PFSGroupTwo:      tunnel.IPsec.Transform.PFSGroupTwo,
				LifetimeSeconds:  tunnel.IPsec.Transform.LifetimeSeconds,
			},
		}

		// Convert IPsec keepalive
		if tunnel.IPsec.Keepalive != nil {
			result.IPsec.Keepalive = &parsers.TunnelIPsecKeepalive{
				Enabled:  tunnel.IPsec.Keepalive.Enabled,
				Mode:     tunnel.IPsec.Keepalive.Mode,
				Interval: tunnel.IPsec.Keepalive.Interval,
				Retry:    tunnel.IPsec.Keepalive.Retry,
			}
		}
	}

	// Convert L2TP block
	if tunnel.L2TP != nil {
		result.L2TP = &parsers.TunnelL2TP{
			Hostname:       tunnel.L2TP.Hostname,
			AlwaysOn:       tunnel.L2TP.AlwaysOn,
			DisconnectTime: tunnel.L2TP.DisconnectTime,
			KeepaliveLog:   tunnel.L2TP.KeepaliveLog,
			SyslogEnabled:  tunnel.L2TP.SyslogEnabled,
			LocalRouterID:  tunnel.L2TP.LocalRouterID,
			RemoteRouterID: tunnel.L2TP.RemoteRouterID,
			RemoteEndID:    tunnel.L2TP.RemoteEndID,
		}

		// Convert L2TP keepalive
		if tunnel.L2TP.Keepalive != nil {
			result.L2TP.Keepalive = &parsers.TunnelL2TPKeepalive{
				Enabled:  tunnel.L2TP.Keepalive.Enabled,
				Interval: tunnel.L2TP.Keepalive.Interval,
				Retry:    tunnel.L2TP.Keepalive.Retry,
			}
		}

		// Convert L2TP tunnel auth
		if tunnel.L2TP.TunnelAuth != nil {
			result.L2TP.TunnelAuth = &parsers.TunnelL2TPAuth{
				Enabled:  tunnel.L2TP.TunnelAuth.Enabled,
				Password: tunnel.L2TP.TunnelAuth.Password,
			}
		}

		// Convert L2TP authentication (L2TPv2)
		if tunnel.L2TP.Authentication != nil {
			result.L2TP.Authentication = &parsers.L2TPAuth{
				Method:        tunnel.L2TP.Authentication.Method,
				RequestMethod: tunnel.L2TP.Authentication.RequestMethod,
				Username:      tunnel.L2TP.Authentication.Username,
				Password:      tunnel.L2TP.Authentication.Password,
			}
		}

		// Convert L2TP IP pool (L2TPv2)
		if tunnel.L2TP.IPPool != nil {
			result.L2TP.IPPool = &parsers.L2TPIPPool{
				Start: tunnel.L2TP.IPPool.Start,
				End:   tunnel.L2TP.IPPool.End,
			}
		}
	}

	return result
}

// convertFromParserTunnel converts parsers.Tunnel to client.Tunnel
func convertFromParserTunnel(p parsers.Tunnel) Tunnel {
	result := Tunnel{
		ID:               p.ID,
		Encapsulation:    p.Encapsulation,
		Enabled:          p.Enabled,
		Name:             p.Name,
		EndpointName:     p.EndpointName,
		EndpointNameType: p.EndpointNameType,
	}

	// Convert IPsec block
	if p.IPsec != nil {
		result.IPsec = &TunnelIPsec{
			IPsecTunnelID:     p.IPsec.IPsecTunnelID,
			LocalAddress:      p.IPsec.LocalAddress,
			RemoteAddress:     p.IPsec.RemoteAddress,
			PreSharedKey:      p.IPsec.PreSharedKey,
			NATTraversal:      p.IPsec.NATTraversal,
			IKERemoteName:     p.IPsec.IKERemoteName,
			IKERemoteNameType: p.IPsec.IKERemoteNameType,
			IKEKeepaliveLog:   p.IPsec.IKEKeepaliveLog,
			IKELog:            p.IPsec.IKELog,
			SecureFilterIn:    p.IPsec.SecureFilterIn,
			SecureFilterOut:   p.IPsec.SecureFilterOut,
			TCPMSSLimit:       p.IPsec.TCPMSSLimit,
			IKEv2Proposal: IKEv2Proposal{
				EncryptionAES256: p.IPsec.IKEv2Proposal.EncryptionAES256,
				EncryptionAES128: p.IPsec.IKEv2Proposal.EncryptionAES128,
				Encryption3DES:   p.IPsec.IKEv2Proposal.Encryption3DES,
				IntegritySHA256:  p.IPsec.IKEv2Proposal.IntegritySHA256,
				IntegritySHA1:    p.IPsec.IKEv2Proposal.IntegritySHA1,
				IntegrityMD5:     p.IPsec.IKEv2Proposal.IntegrityMD5,
				GroupFourteen:    p.IPsec.IKEv2Proposal.GroupFourteen,
				GroupFive:        p.IPsec.IKEv2Proposal.GroupFive,
				GroupTwo:         p.IPsec.IKEv2Proposal.GroupTwo,
				LifetimeSeconds:  p.IPsec.IKEv2Proposal.LifetimeSeconds,
			},
			Transform: IPsecTransform{
				Protocol:         p.IPsec.Transform.Protocol,
				EncryptionAES256: p.IPsec.Transform.EncryptionAES256,
				EncryptionAES128: p.IPsec.Transform.EncryptionAES128,
				Encryption3DES:   p.IPsec.Transform.Encryption3DES,
				IntegritySHA256:  p.IPsec.Transform.IntegritySHA256,
				IntegritySHA1:    p.IPsec.Transform.IntegritySHA1,
				IntegrityMD5:     p.IPsec.Transform.IntegrityMD5,
				PFSGroupFourteen: p.IPsec.Transform.PFSGroupFourteen,
				PFSGroupFive:     p.IPsec.Transform.PFSGroupFive,
				PFSGroupTwo:      p.IPsec.Transform.PFSGroupTwo,
				LifetimeSeconds:  p.IPsec.Transform.LifetimeSeconds,
			},
		}

		// Convert IPsec keepalive
		if p.IPsec.Keepalive != nil {
			result.IPsec.Keepalive = &TunnelIPsecKeepalive{
				Enabled:  p.IPsec.Keepalive.Enabled,
				Mode:     p.IPsec.Keepalive.Mode,
				Interval: p.IPsec.Keepalive.Interval,
				Retry:    p.IPsec.Keepalive.Retry,
			}
		}
	}

	// Convert L2TP block
	if p.L2TP != nil {
		result.L2TP = &TunnelL2TP{
			Hostname:       p.L2TP.Hostname,
			AlwaysOn:       p.L2TP.AlwaysOn,
			DisconnectTime: p.L2TP.DisconnectTime,
			KeepaliveLog:   p.L2TP.KeepaliveLog,
			SyslogEnabled:  p.L2TP.SyslogEnabled,
			LocalRouterID:  p.L2TP.LocalRouterID,
			RemoteRouterID: p.L2TP.RemoteRouterID,
			RemoteEndID:    p.L2TP.RemoteEndID,
		}

		// Convert L2TP keepalive
		if p.L2TP.Keepalive != nil {
			result.L2TP.Keepalive = &TunnelL2TPKeepalive{
				Enabled:  p.L2TP.Keepalive.Enabled,
				Interval: p.L2TP.Keepalive.Interval,
				Retry:    p.L2TP.Keepalive.Retry,
			}
		}

		// Convert L2TP tunnel auth
		if p.L2TP.TunnelAuth != nil {
			result.L2TP.TunnelAuth = &TunnelL2TPAuth{
				Enabled:  p.L2TP.TunnelAuth.Enabled,
				Password: p.L2TP.TunnelAuth.Password,
			}
		}

		// Convert L2TP authentication (L2TPv2)
		if p.L2TP.Authentication != nil {
			result.L2TP.Authentication = &L2TPAuth{
				Method:        p.L2TP.Authentication.Method,
				RequestMethod: p.L2TP.Authentication.RequestMethod,
				Username:      p.L2TP.Authentication.Username,
				Password:      p.L2TP.Authentication.Password,
			}
		}

		// Convert L2TP IP pool (L2TPv2)
		if p.L2TP.IPPool != nil {
			result.L2TP.IPPool = &L2TPIPPool{
				Start: p.L2TP.IPPool.Start,
				End:   p.L2TP.IPPool.End,
			}
		}
	}

	return result
}
