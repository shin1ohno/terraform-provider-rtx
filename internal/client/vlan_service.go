package client

import (
	"context"
	"fmt"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// VLANService handles VLAN operations
type VLANService struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewVLANService creates a new VLAN service instance
func NewVLANService(executor Executor, client *rtxClient) *VLANService {
	return &VLANService{
		executor: executor,
		client:   client,
	}
}

// CreateVLAN creates a new VLAN
func (s *VLANService) CreateVLAN(ctx context.Context, vlan VLAN) error {
	// Convert client.VLAN to parsers.VLAN for validation
	parserVLAN := s.toParserVLAN(vlan)

	// Validate input
	if err := parsers.ValidateVLAN(parserVLAN); err != nil {
		return fmt.Errorf("invalid VLAN: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get existing VLANs to find available slot
	existingVLANs, err := s.ListVLANs(ctx)
	if err != nil {
		// If we can't list, assume no existing VLANs
		logging.FromContext(ctx).Debug().Str("service", "vlan").Msgf("Could not list existing VLANs: %v, assuming empty", err)
		existingVLANs = []VLAN{}
	}

	// Convert to parser VLANs for slot finding
	parserExisting := make([]parsers.VLAN, len(existingVLANs))
	for i, v := range existingVLANs {
		parserExisting[i] = s.toParserVLAN(v)
	}

	// Check if VLAN with same ID already exists on this interface
	for _, existing := range existingVLANs {
		if existing.Interface == vlan.Interface && existing.VlanID == vlan.VlanID {
			return fmt.Errorf("VLAN %d already exists on interface %s", vlan.VlanID, vlan.Interface)
		}
	}

	// Find next available slot
	slot := parsers.FindNextAvailableSlot(parserExisting, vlan.Interface)
	if slot == -1 {
		return fmt.Errorf("no available VLAN slots on interface %s", vlan.Interface)
	}

	// Build and execute VLAN creation command
	cmd := parsers.BuildVLANCommand(vlan.Interface, slot, vlan.VlanID)
	logging.FromContext(ctx).Debug().Str("service", "vlan").Msgf("Creating VLAN with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create VLAN: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Compute the VLAN interface name
	vlanInterface := fmt.Sprintf("%s/%d", vlan.Interface, slot)

	// Configure IP address if specified
	if vlan.IPAddress != "" && vlan.IPMask != "" {
		ipCmd := parsers.BuildVLANIPCommand(vlanInterface, vlan.IPAddress, vlan.IPMask)
		logging.FromContext(ctx).Debug().Str("service", "vlan").Msgf("Setting VLAN IP with command: %s", ipCmd)

		output, err = s.executor.Run(ctx, ipCmd)
		if err != nil {
			return fmt.Errorf("failed to set VLAN IP: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("IP command failed: %s", string(output))
		}
	}

	// Configure description if specified
	if vlan.Name != "" {
		descCmd := parsers.BuildVLANDescriptionCommand(vlanInterface, vlan.Name)
		logging.FromContext(ctx).Debug().Str("service", "vlan").Msgf("Setting VLAN description with command: %s", descCmd)

		output, err = s.executor.Run(ctx, descCmd)
		if err != nil {
			return fmt.Errorf("failed to set VLAN description: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("description command failed: %s", string(output))
		}
	}

	// Configure shutdown state if specified
	if vlan.Shutdown {
		shutdownCmd := parsers.BuildVLANDisableCommand(vlanInterface)
		logging.FromContext(ctx).Debug().Str("service", "vlan").Msgf("Disabling VLAN with command: %s", shutdownCmd)

		output, err = s.executor.Run(ctx, shutdownCmd)
		if err != nil {
			return fmt.Errorf("failed to disable VLAN: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("disable command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("VLAN created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetVLAN retrieves a VLAN configuration
func (s *VLANService) GetVLAN(ctx context.Context, iface string, vlanID int) (*VLAN, error) {
	cmd := parsers.BuildShowVLANCommand(iface, vlanID)
	logging.FromContext(ctx).Debug().Str("service", "vlan").Msgf("Getting VLAN with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get VLAN: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "vlan").Msgf("VLAN raw output: %q", string(output))

	parser := parsers.NewVLANParser()
	parserVLAN, err := parser.ParseSingleVLAN(string(output), iface, vlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VLAN: %w", err)
	}

	// Convert parsers.VLAN to client.VLAN
	vlan := s.fromParserVLAN(*parserVLAN)
	return &vlan, nil
}

// UpdateVLAN updates an existing VLAN
// Note: interface and vlan_id changes require recreation
func (s *VLANService) UpdateVLAN(ctx context.Context, vlan VLAN) error {
	parserVLAN := s.toParserVLAN(vlan)

	// Validate input
	if err := parsers.ValidateVLAN(parserVLAN); err != nil {
		return fmt.Errorf("invalid VLAN: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get current VLAN configuration
	currentVLAN, err := s.GetVLAN(ctx, vlan.Interface, vlan.VlanID)
	if err != nil {
		return fmt.Errorf("failed to get current VLAN: %w", err)
	}

	vlanInterface := currentVLAN.VlanInterface

	// Update IP address if changed
	if vlan.IPAddress != currentVLAN.IPAddress || vlan.IPMask != currentVLAN.IPMask {
		// Remove old IP if exists
		if currentVLAN.IPAddress != "" {
			deleteIPCmd := parsers.BuildDeleteVLANIPCommand(vlanInterface)
			logging.FromContext(ctx).Debug().Str("service", "vlan").Msgf("Removing old VLAN IP with command: %s", deleteIPCmd)
			_, _ = s.executor.Run(ctx, deleteIPCmd) // Ignore errors for cleanup
		}

		// Set new IP if specified
		if vlan.IPAddress != "" && vlan.IPMask != "" {
			ipCmd := parsers.BuildVLANIPCommand(vlanInterface, vlan.IPAddress, vlan.IPMask)
			logging.FromContext(ctx).Debug().Str("service", "vlan").Msgf("Setting VLAN IP with command: %s", ipCmd)

			output, err := s.executor.Run(ctx, ipCmd)
			if err != nil {
				return fmt.Errorf("failed to set VLAN IP: %w", err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("IP command failed: %s", string(output))
			}
		}
	}

	// Update description if changed
	if vlan.Name != currentVLAN.Name {
		// Remove old description if exists
		if currentVLAN.Name != "" {
			deleteDescCmd := parsers.BuildDeleteVLANDescriptionCommand(vlanInterface)
			logging.FromContext(ctx).Debug().Str("service", "vlan").Msgf("Removing old VLAN description with command: %s", deleteDescCmd)
			_, _ = s.executor.Run(ctx, deleteDescCmd) // Ignore errors for cleanup
		}

		// Set new description if specified
		if vlan.Name != "" {
			descCmd := parsers.BuildVLANDescriptionCommand(vlanInterface, vlan.Name)
			logging.FromContext(ctx).Debug().Str("service", "vlan").Msgf("Setting VLAN description with command: %s", descCmd)

			output, err := s.executor.Run(ctx, descCmd)
			if err != nil {
				return fmt.Errorf("failed to set VLAN description: %w", err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("description command failed: %s", string(output))
			}
		}
	}

	// Update shutdown state if changed
	if vlan.Shutdown != currentVLAN.Shutdown {
		var stateCmd string
		if vlan.Shutdown {
			stateCmd = parsers.BuildVLANDisableCommand(vlanInterface)
		} else {
			stateCmd = parsers.BuildVLANEnableCommand(vlanInterface)
		}
		logging.FromContext(ctx).Debug().Str("service", "vlan").Msgf("Changing VLAN state with command: %s", stateCmd)

		output, err := s.executor.Run(ctx, stateCmd)
		if err != nil {
			return fmt.Errorf("failed to change VLAN state: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("state command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("VLAN updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// DeleteVLAN removes a VLAN
func (s *VLANService) DeleteVLAN(ctx context.Context, iface string, vlanID int) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get current VLAN to find the interface name
	currentVLAN, err := s.GetVLAN(ctx, iface, vlanID)
	if err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return fmt.Errorf("failed to get current VLAN: %w", err)
	}

	cmd := parsers.BuildDeleteVLANCommand(currentVLAN.VlanInterface)
	logging.FromContext(ctx).Debug().Str("service", "vlan").Msgf("Deleting VLAN with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete VLAN: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		// Check if it's already gone
		if strings.Contains(strings.ToLower(string(output)), "not found") {
			return nil
		}
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("VLAN deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ListVLANs retrieves all VLANs
func (s *VLANService) ListVLANs(ctx context.Context) ([]VLAN, error) {
	cmd := parsers.BuildShowAllVLANsCommand()
	logging.FromContext(ctx).Debug().Str("service", "vlan").Msgf("Listing VLANs with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list VLANs: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "vlan").Msgf("VLANs raw output: %q", string(output))

	parser := parsers.NewVLANParser()
	parserVLANs, err := parser.ParseVLANConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse VLANs: %w", err)
	}

	// Convert parsers.VLAN to client.VLAN
	vlans := make([]VLAN, len(parserVLANs))
	for i, pv := range parserVLANs {
		vlans[i] = s.fromParserVLAN(pv)
	}

	return vlans, nil
}

// toParserVLAN converts client.VLAN to parsers.VLAN
func (s *VLANService) toParserVLAN(vlan VLAN) parsers.VLAN {
	return parsers.VLAN{
		VlanID:        vlan.VlanID,
		Name:          vlan.Name,
		Interface:     vlan.Interface,
		VlanInterface: vlan.VlanInterface,
		IPAddress:     vlan.IPAddress,
		IPMask:        vlan.IPMask,
		Shutdown:      vlan.Shutdown,
	}
}

// fromParserVLAN converts parsers.VLAN to client.VLAN
func (s *VLANService) fromParserVLAN(pv parsers.VLAN) VLAN {
	return VLAN{
		VlanID:        pv.VlanID,
		Name:          pv.Name,
		Interface:     pv.Interface,
		VlanInterface: pv.VlanInterface,
		IPAddress:     pv.IPAddress,
		IPMask:        pv.IPMask,
		Shutdown:      pv.Shutdown,
	}
}
