package client

import (
	"context"
	"fmt"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// IPsecTunnelService handles IPsec tunnel configuration operations
type IPsecTunnelService struct {
	executor Executor
	client   *rtxClient
}

// NewIPsecTunnelService creates a new IPsec tunnel service
func NewIPsecTunnelService(executor Executor, client *rtxClient) *IPsecTunnelService {
	return &IPsecTunnelService{
		executor: executor,
		client:   client,
	}
}

// Get retrieves a specific IPsec tunnel configuration
func (s *IPsecTunnelService) Get(ctx context.Context, tunnelID int) (*IPsecTunnel, error) {
	tunnels, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, tunnel := range tunnels {
		if tunnel.ID == tunnelID {
			return &tunnel, nil
		}
	}

	return nil, fmt.Errorf("IPsec tunnel %d not found", tunnelID)
}

// List retrieves all IPsec tunnel configurations
func (s *IPsecTunnelService) List(ctx context.Context) ([]IPsecTunnel, error) {
	output, err := s.executor.Run(ctx, parsers.BuildShowIPsecConfigCommand())
	if err != nil {
		return nil, fmt.Errorf("failed to get IPsec config: %w", err)
	}

	parser := parsers.NewIPsecTunnelParser()
	parsed, err := parser.ParseIPsecTunnelConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IPsec config: %w", err)
	}

	// Convert from parser types to client types
	tunnels := make([]IPsecTunnel, len(parsed))
	for i, p := range parsed {
		tunnels[i] = convertFromParserIPsecTunnel(p)
	}

	return tunnels, nil
}

// Create creates a new IPsec tunnel
func (s *IPsecTunnelService) Create(ctx context.Context, tunnel IPsecTunnel) error {
	// Validate configuration
	parserTunnel := convertToParserIPsecTunnel(tunnel)
	if err := parsers.ValidateIPsecTunnel(parserTunnel); err != nil {
		return fmt.Errorf("invalid IPsec tunnel config: %w", err)
	}

	commands := []string{}

	// 1. Select tunnel
	commands = append(commands, parsers.BuildTunnelSelectCommand(tunnel.ID))

	// 2. Create IPsec tunnel
	commands = append(commands, parsers.BuildIPsecTunnelCommand(tunnel.ID))

	// 3. Set local address
	commands = append(commands, parsers.BuildIPsecIKELocalAddressCommand(tunnel.ID, tunnel.LocalAddress))

	// 4. Set remote address
	commands = append(commands, parsers.BuildIPsecIKERemoteAddressCommand(tunnel.ID, tunnel.RemoteAddress))

	// 5. Set pre-shared key
	commands = append(commands, parsers.BuildIPsecIKEPreSharedKeyCommand(tunnel.ID, tunnel.PreSharedKey))

	// 6. Set IKE encryption
	commands = append(commands, parsers.BuildIPsecIKEEncryptionCommand(tunnel.ID, parserTunnel.IKEv2Proposal))

	// 7. Set IKE hash
	commands = append(commands, parsers.BuildIPsecIKEHashCommand(tunnel.ID, parserTunnel.IKEv2Proposal))

	// 8. Set IKE group
	commands = append(commands, parsers.BuildIPsecIKEGroupCommand(tunnel.ID, parserTunnel.IKEv2Proposal))

	// 9. Set SA policy
	policyNum := 100 + tunnel.ID
	commands = append(commands, parsers.BuildIPsecSAPolicyCommand(policyNum, tunnel.ID, parserTunnel.IPsecTransform))

	// 10. Configure DPD if enabled
	if tunnel.DPDEnabled {
		commands = append(commands, parsers.BuildIPsecIKEKeepaliveCommand(tunnel.ID, tunnel.DPDInterval, tunnel.DPDRetry))
	} else {
		commands = append(commands, parsers.BuildIPsecIKEKeepaliveOffCommand(tunnel.ID))
	}

	// Execute all commands
	for _, cmd := range commands {
		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to execute IPsec command '%s': %w", cmd, err)
		}
		if containsError(string(output)) {
			return fmt.Errorf("IPsec command '%s' failed: %s", cmd, string(output))
		}
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save IPsec config: %w", err)
	}

	return nil
}

// Update modifies an existing IPsec tunnel
func (s *IPsecTunnelService) Update(ctx context.Context, tunnel IPsecTunnel) error {
	// Validate configuration
	parserTunnel := convertToParserIPsecTunnel(tunnel)
	if err := parsers.ValidateIPsecTunnel(parserTunnel); err != nil {
		return fmt.Errorf("invalid IPsec tunnel config: %w", err)
	}

	commands := []string{}

	// Select tunnel
	commands = append(commands, parsers.BuildTunnelSelectCommand(tunnel.ID))

	// Update local address
	commands = append(commands, parsers.BuildIPsecIKELocalAddressCommand(tunnel.ID, tunnel.LocalAddress))

	// Update remote address
	commands = append(commands, parsers.BuildIPsecIKERemoteAddressCommand(tunnel.ID, tunnel.RemoteAddress))

	// Update pre-shared key if provided
	if tunnel.PreSharedKey != "" {
		commands = append(commands, parsers.BuildIPsecIKEPreSharedKeyCommand(tunnel.ID, tunnel.PreSharedKey))
	}

	// Update IKE settings
	commands = append(commands, parsers.BuildIPsecIKEEncryptionCommand(tunnel.ID, parserTunnel.IKEv2Proposal))
	commands = append(commands, parsers.BuildIPsecIKEHashCommand(tunnel.ID, parserTunnel.IKEv2Proposal))
	commands = append(commands, parsers.BuildIPsecIKEGroupCommand(tunnel.ID, parserTunnel.IKEv2Proposal))

	// Update SA policy
	policyNum := 100 + tunnel.ID
	commands = append(commands, parsers.BuildIPsecSAPolicyCommand(policyNum, tunnel.ID, parserTunnel.IPsecTransform))

	// Update DPD settings
	if tunnel.DPDEnabled {
		commands = append(commands, parsers.BuildIPsecIKEKeepaliveCommand(tunnel.ID, tunnel.DPDInterval, tunnel.DPDRetry))
	} else {
		commands = append(commands, parsers.BuildIPsecIKEKeepaliveOffCommand(tunnel.ID))
	}

	// Execute all commands
	for _, cmd := range commands {
		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to execute IPsec command '%s': %w", cmd, err)
		}
		if containsError(string(output)) {
			return fmt.Errorf("IPsec command '%s' failed: %s", cmd, string(output))
		}
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save IPsec config: %w", err)
	}

	return nil
}

// Delete removes an IPsec tunnel
func (s *IPsecTunnelService) Delete(ctx context.Context, tunnelID int) error {
	commands := []string{
		parsers.BuildDeleteIPsecTunnelCommand(tunnelID),
		parsers.BuildDeleteTunnelSelectCommand(tunnelID),
	}

	for _, cmd := range commands {
		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to execute IPsec delete command '%s': %w", cmd, err)
		}
		if containsError(string(output)) {
			return fmt.Errorf("IPsec delete command '%s' failed: %s", cmd, string(output))
		}
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save config after IPsec delete: %w", err)
	}

	return nil
}

// convertToParserIPsecTunnel converts client IPsecTunnel to parser IPsecTunnel
func convertToParserIPsecTunnel(tunnel IPsecTunnel) parsers.IPsecTunnel {
	return parsers.IPsecTunnel{
		ID:            tunnel.ID,
		Name:          tunnel.Name,
		LocalAddress:  tunnel.LocalAddress,
		RemoteAddress: tunnel.RemoteAddress,
		PreSharedKey:  tunnel.PreSharedKey,
		LocalNetwork:  tunnel.LocalNetwork,
		RemoteNetwork: tunnel.RemoteNetwork,
		DPDEnabled:    tunnel.DPDEnabled,
		DPDInterval:   tunnel.DPDInterval,
		DPDRetry:      tunnel.DPDRetry,
		Enabled:       tunnel.Enabled,
		IKEv2Proposal: parsers.IKEv2Proposal{
			EncryptionAES256: tunnel.IKEv2Proposal.EncryptionAES256,
			EncryptionAES128: tunnel.IKEv2Proposal.EncryptionAES128,
			Encryption3DES:   tunnel.IKEv2Proposal.Encryption3DES,
			IntegritySHA256:  tunnel.IKEv2Proposal.IntegritySHA256,
			IntegritySHA1:    tunnel.IKEv2Proposal.IntegritySHA1,
			IntegrityMD5:     tunnel.IKEv2Proposal.IntegrityMD5,
			GroupFourteen:    tunnel.IKEv2Proposal.GroupFourteen,
			GroupFive:        tunnel.IKEv2Proposal.GroupFive,
			GroupTwo:         tunnel.IKEv2Proposal.GroupTwo,
			LifetimeSeconds:  tunnel.IKEv2Proposal.LifetimeSeconds,
		},
		IPsecTransform: parsers.IPsecTransform{
			Protocol:         tunnel.IPsecTransform.Protocol,
			EncryptionAES256: tunnel.IPsecTransform.EncryptionAES256,
			EncryptionAES128: tunnel.IPsecTransform.EncryptionAES128,
			Encryption3DES:   tunnel.IPsecTransform.Encryption3DES,
			IntegritySHA256:  tunnel.IPsecTransform.IntegritySHA256,
			IntegritySHA1:    tunnel.IPsecTransform.IntegritySHA1,
			IntegrityMD5:     tunnel.IPsecTransform.IntegrityMD5,
			PFSGroupFourteen: tunnel.IPsecTransform.PFSGroupFourteen,
			PFSGroupFive:     tunnel.IPsecTransform.PFSGroupFive,
			PFSGroupTwo:      tunnel.IPsecTransform.PFSGroupTwo,
			LifetimeSeconds:  tunnel.IPsecTransform.LifetimeSeconds,
		},
	}
}

// convertFromParserIPsecTunnel converts parser IPsecTunnel to client IPsecTunnel
func convertFromParserIPsecTunnel(p parsers.IPsecTunnel) IPsecTunnel {
	return IPsecTunnel{
		ID:            p.ID,
		Name:          p.Name,
		LocalAddress:  p.LocalAddress,
		RemoteAddress: p.RemoteAddress,
		PreSharedKey:  p.PreSharedKey,
		LocalNetwork:  p.LocalNetwork,
		RemoteNetwork: p.RemoteNetwork,
		DPDEnabled:    p.DPDEnabled,
		DPDInterval:   p.DPDInterval,
		DPDRetry:      p.DPDRetry,
		Enabled:       p.Enabled,
		IKEv2Proposal: IKEv2Proposal{
			EncryptionAES256: p.IKEv2Proposal.EncryptionAES256,
			EncryptionAES128: p.IKEv2Proposal.EncryptionAES128,
			Encryption3DES:   p.IKEv2Proposal.Encryption3DES,
			IntegritySHA256:  p.IKEv2Proposal.IntegritySHA256,
			IntegritySHA1:    p.IKEv2Proposal.IntegritySHA1,
			IntegrityMD5:     p.IKEv2Proposal.IntegrityMD5,
			GroupFourteen:    p.IKEv2Proposal.GroupFourteen,
			GroupFive:        p.IKEv2Proposal.GroupFive,
			GroupTwo:         p.IKEv2Proposal.GroupTwo,
			LifetimeSeconds:  p.IKEv2Proposal.LifetimeSeconds,
		},
		IPsecTransform: IPsecTransform{
			Protocol:         p.IPsecTransform.Protocol,
			EncryptionAES256: p.IPsecTransform.EncryptionAES256,
			EncryptionAES128: p.IPsecTransform.EncryptionAES128,
			Encryption3DES:   p.IPsecTransform.Encryption3DES,
			IntegritySHA256:  p.IPsecTransform.IntegritySHA256,
			IntegritySHA1:    p.IPsecTransform.IntegritySHA1,
			IntegrityMD5:     p.IPsecTransform.IntegrityMD5,
			PFSGroupFourteen: p.IPsecTransform.PFSGroupFourteen,
			PFSGroupFive:     p.IPsecTransform.PFSGroupFive,
			PFSGroupTwo:      p.IPsecTransform.PFSGroupTwo,
			LifetimeSeconds:  p.IPsecTransform.LifetimeSeconds,
		},
	}
}
