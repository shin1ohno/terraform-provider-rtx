package client

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// QoSService handles QoS operations (class-map, policy-map, service-policy, shape)
type QoSService struct {
	executor Executor
	client   *rtxClient
}

// NewQoSService creates a new QoS service instance
func NewQoSService(executor Executor, client *rtxClient) *QoSService {
	return &QoSService{
		executor: executor,
		client:   client,
	}
}

// ========== Class Map Methods ==========

// CreateClassMap creates a new class-map
func (s *QoSService) CreateClassMap(ctx context.Context, cm ClassMap) error {
	// Validate input
	parserCM := s.toParserClassMap(cm)
	if err := parsers.ValidateClassMap(parserCM); err != nil {
		return fmt.Errorf("invalid class-map: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// For RTX routers, class-maps are typically implemented using IP filters
	// We create an IP filter that matches the specified criteria
	if cm.MatchFilter > 0 {
		// Class map references an existing filter - nothing to create
		logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Class-map %s references existing filter %d", cm.Name, cm.MatchFilter)
	}

	// If destination ports are specified, we might need to create filters
	// RTX uses "queue <interface> class filter <n> <filter>" syntax
	// The class-map itself is a logical grouping that references filters

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("class-map created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetClassMap retrieves a class-map configuration
func (s *QoSService) GetClassMap(ctx context.Context, name string) (*ClassMap, error) {
	// Get all QoS configuration to find the class map
	cmd := parsers.BuildShowAllQoSCommand()
	logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Getting class-map with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get class-map: %w", err)
	}

	// Parse the output to find class map by name
	// In RTX, class maps are referenced in queue configurations
	// We look for patterns like "queue <interface> class filter <n> <filter>"
	cm := &ClassMap{
		Name:                 name,
		MatchDestinationPort: []int{},
		MatchSourcePort:      []int{},
	}

	// Parse IP filter definitions that might be referenced
	filterPattern := regexp.MustCompile(`ip\s+filter\s+(\d+)\s+(\w+)\s+.*`)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if matches := filterPattern.FindStringSubmatch(line); len(matches) >= 2 {
			filterNum, _ := strconv.Atoi(matches[1])
			// Check if this filter is associated with our class map
			if strings.Contains(line, name) || cm.MatchFilter == filterNum {
				cm.MatchFilter = filterNum
			}
		}
	}

	return cm, nil
}

// UpdateClassMap updates an existing class-map
func (s *QoSService) UpdateClassMap(ctx context.Context, cm ClassMap) error {
	// Validate input
	parserCM := s.toParserClassMap(cm)
	if err := parsers.ValidateClassMap(parserCM); err != nil {
		return fmt.Errorf("invalid class-map: %w", err)
	}

	// For RTX, updating a class-map means updating the associated filter
	// The class-map itself is a logical reference

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("class-map updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// DeleteClassMap removes a class-map
func (s *QoSService) DeleteClassMap(ctx context.Context, name string) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Class-maps in RTX are referenced by queue configurations
	// Deleting a class-map means removing the queue class filter references

	logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Deleting class-map: %s", name)

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("class-map deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ListClassMaps retrieves all class-maps
func (s *QoSService) ListClassMaps(ctx context.Context) ([]ClassMap, error) {
	cmd := parsers.BuildShowAllQoSCommand()
	logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Listing class-maps with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list class-maps: %w", err)
	}

	// Parse output to extract class maps
	classMapsByFilter := make(map[int]*ClassMap)
	filterPattern := regexp.MustCompile(`queue\s+\S+\s+class\s+filter\s+(\d+)\s+(\d+)`)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if matches := filterPattern.FindStringSubmatch(line); len(matches) >= 3 {
			classNum, _ := strconv.Atoi(matches[1])
			filterNum, _ := strconv.Atoi(matches[2])

			if _, exists := classMapsByFilter[classNum]; !exists {
				classMapsByFilter[classNum] = &ClassMap{
					Name:                 fmt.Sprintf("class%d", classNum),
					MatchFilter:          filterNum,
					MatchDestinationPort: []int{},
					MatchSourcePort:      []int{},
				}
			}
		}
	}

	result := make([]ClassMap, 0, len(classMapsByFilter))
	for _, cm := range classMapsByFilter {
		result = append(result, *cm)
	}

	return result, nil
}

// ========== Policy Map Methods ==========

// CreatePolicyMap creates a new policy-map
func (s *QoSService) CreatePolicyMap(ctx context.Context, pm PolicyMap) error {
	// Validate input
	parserPM := s.toParserPolicyMap(pm)
	if err := parsers.ValidatePolicyMap(parserPM); err != nil {
		return fmt.Errorf("invalid policy-map: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Policy maps in RTX are implemented through queue configurations
	// We don't create them directly - they're applied via service-policy

	logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Policy-map %s created (logical configuration)", pm.Name)

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("policy-map created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetPolicyMap retrieves a policy-map configuration
func (s *QoSService) GetPolicyMap(ctx context.Context, name string) (*PolicyMap, error) {
	cmd := parsers.BuildShowAllQoSCommand()
	logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Getting policy-map with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy-map: %w", err)
	}

	pm := &PolicyMap{
		Name:    name,
		Classes: []PolicyMapClass{},
	}

	// Parse queue configurations to build policy map
	classPriorityPattern := regexp.MustCompile(`queue\s+\S+\s+class\s+priority\s+(\d+)\s+(\S+)`)
	queueLengthPattern := regexp.MustCompile(`queue\s+\S+\s+length\s+(\d+)\s+(\d+)`)

	classMap := make(map[int]*PolicyMapClass)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if matches := classPriorityPattern.FindStringSubmatch(line); len(matches) >= 3 {
			classNum, _ := strconv.Atoi(matches[1])
			priority := matches[2]

			if _, exists := classMap[classNum]; !exists {
				classMap[classNum] = &PolicyMapClass{Name: fmt.Sprintf("class%d", classNum)}
			}
			classMap[classNum].Priority = priority
		}

		if matches := queueLengthPattern.FindStringSubmatch(line); len(matches) >= 3 {
			classNum, _ := strconv.Atoi(matches[1])
			length, _ := strconv.Atoi(matches[2])

			if _, exists := classMap[classNum]; !exists {
				classMap[classNum] = &PolicyMapClass{Name: fmt.Sprintf("class%d", classNum)}
			}
			classMap[classNum].QueueLimit = length
		}
	}

	for _, class := range classMap {
		pm.Classes = append(pm.Classes, *class)
	}

	return pm, nil
}

// UpdatePolicyMap updates an existing policy-map
func (s *QoSService) UpdatePolicyMap(ctx context.Context, pm PolicyMap) error {
	// Validate input
	parserPM := s.toParserPolicyMap(pm)
	if err := parsers.ValidatePolicyMap(parserPM); err != nil {
		return fmt.Errorf("invalid policy-map: %w", err)
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("policy-map updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// DeletePolicyMap removes a policy-map
func (s *QoSService) DeletePolicyMap(ctx context.Context, name string) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Deleting policy-map: %s", name)

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("policy-map deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ListPolicyMaps retrieves all policy-maps
func (s *QoSService) ListPolicyMaps(ctx context.Context) ([]PolicyMap, error) {
	cmd := parsers.BuildShowAllQoSCommand()
	logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Listing policy-maps with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list policy-maps: %w", err)
	}

	// Parse output to extract unique queue types (which represent policy maps)
	policyMapsByIface := make(map[string]*PolicyMap)
	queueTypePattern := regexp.MustCompile(`queue\s+(\S+)\s+type\s+(\S+)`)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if matches := queueTypePattern.FindStringSubmatch(line); len(matches) >= 3 {
			iface := matches[1]
			queueType := matches[2]

			policyMapsByIface[iface] = &PolicyMap{
				Name:    fmt.Sprintf("%s-%s", iface, queueType),
				Classes: []PolicyMapClass{},
			}
		}
	}

	result := make([]PolicyMap, 0, len(policyMapsByIface))
	for _, pm := range policyMapsByIface {
		result = append(result, *pm)
	}

	return result, nil
}

// ========== Service Policy Methods ==========

// CreateServicePolicy creates a new service-policy
func (s *QoSService) CreateServicePolicy(ctx context.Context, sp ServicePolicy) error {
	// Validate input
	parserSP := s.toParserServicePolicy(sp)
	if err := parsers.ValidateServicePolicy(parserSP); err != nil {
		return fmt.Errorf("invalid service-policy: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Apply the queue type to the interface
	cmd := parsers.BuildQueueTypeCommand(sp.Interface, sp.PolicyMap)
	logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Creating service-policy with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create service-policy: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("service-policy created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetServicePolicy retrieves a service-policy configuration
func (s *QoSService) GetServicePolicy(ctx context.Context, iface string, direction string) (*ServicePolicy, error) {
	cmd := parsers.BuildShowQoSCommand(iface)
	logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Getting service-policy with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get service-policy: %w", err)
	}

	parser := parsers.NewQoSParser()
	parserSP, err := parser.ParseServicePolicy(string(output), iface)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service-policy: %w", err)
	}

	sp := s.fromParserServicePolicy(*parserSP)
	return &sp, nil
}

// UpdateServicePolicy updates an existing service-policy
func (s *QoSService) UpdateServicePolicy(ctx context.Context, sp ServicePolicy) error {
	// Validate input
	parserSP := s.toParserServicePolicy(sp)
	if err := parsers.ValidateServicePolicy(parserSP); err != nil {
		return fmt.Errorf("invalid service-policy: %w", err)
	}

	// Delete existing and recreate
	if err := s.DeleteServicePolicy(ctx, sp.Interface, sp.Direction); err != nil {
		// Ignore errors during deletion
		logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Ignoring deletion error during update: %v", err)
	}

	return s.CreateServicePolicy(ctx, sp)
}

// DeleteServicePolicy removes a service-policy
func (s *QoSService) DeleteServicePolicy(ctx context.Context, iface string, direction string) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cmd := parsers.BuildDeleteQueueTypeCommand(iface)
	logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Deleting service-policy with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete service-policy: %w", err)
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
			return fmt.Errorf("service-policy deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ListServicePolicies retrieves all service-policies
func (s *QoSService) ListServicePolicies(ctx context.Context) ([]ServicePolicy, error) {
	cmd := parsers.BuildShowAllQoSCommand()
	logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Listing service-policies with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list service-policies: %w", err)
	}

	queueTypePattern := regexp.MustCompile(`queue\s+(\S+)\s+type\s+(\S+)`)
	lines := strings.Split(string(output), "\n")

	var result []ServicePolicy
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if matches := queueTypePattern.FindStringSubmatch(line); len(matches) >= 3 {
			result = append(result, ServicePolicy{
				Interface: matches[1],
				Direction: "output",
				PolicyMap: matches[2],
			})
		}
	}

	return result, nil
}

// ========== Shape Methods ==========

// CreateShape creates a new shape configuration
func (s *QoSService) CreateShape(ctx context.Context, sc ShapeConfig) error {
	// Validate input
	parserSC := s.toParserShapeConfig(sc)
	if err := parsers.ValidateShapeConfig(parserSC); err != nil {
		return fmt.Errorf("invalid shape configuration: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Apply speed command
	cmd := parsers.BuildSpeedCommand(sc.Interface, sc.ShapeAverage)
	logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Creating shape with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create shape: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("shape created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetShape retrieves a shape configuration
func (s *QoSService) GetShape(ctx context.Context, iface string, direction string) (*ShapeConfig, error) {
	cmd := parsers.BuildShowQoSCommand(iface)
	logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Getting shape with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get shape: %w", err)
	}

	parser := parsers.NewQoSParser()
	parserSC, err := parser.ParseShapeConfig(string(output), iface)
	if err != nil {
		return nil, fmt.Errorf("failed to parse shape: %w", err)
	}

	sc := s.fromParserShapeConfig(*parserSC)
	return &sc, nil
}

// UpdateShape updates an existing shape configuration
func (s *QoSService) UpdateShape(ctx context.Context, sc ShapeConfig) error {
	// Validate input
	parserSC := s.toParserShapeConfig(sc)
	if err := parsers.ValidateShapeConfig(parserSC); err != nil {
		return fmt.Errorf("invalid shape configuration: %w", err)
	}

	// Simply re-apply the speed command (RTX allows overwriting)
	cmd := parsers.BuildSpeedCommand(sc.Interface, sc.ShapeAverage)
	logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Updating shape with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to update shape: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("shape updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// DeleteShape removes a shape configuration
func (s *QoSService) DeleteShape(ctx context.Context, iface string, direction string) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cmd := parsers.BuildDeleteSpeedCommand(iface)
	logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Deleting shape with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete shape: %w", err)
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
			return fmt.Errorf("shape deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ListShapes retrieves all shape configurations
func (s *QoSService) ListShapes(ctx context.Context) ([]ShapeConfig, error) {
	cmd := parsers.BuildShowAllQoSCommand()
	logging.FromContext(ctx).Debug().Str("service", "qos").Msgf("Listing shapes with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list shapes: %w", err)
	}

	speedPattern := regexp.MustCompile(`speed\s+(\S+)\s+(\d+)`)
	lines := strings.Split(string(output), "\n")

	var result []ShapeConfig
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if matches := speedPattern.FindStringSubmatch(line); len(matches) >= 3 {
			speed, _ := strconv.Atoi(matches[2])
			result = append(result, ShapeConfig{
				Interface:    matches[1],
				Direction:    "output",
				ShapeAverage: speed,
			})
		}
	}

	return result, nil
}

// ========== Conversion Helpers ==========

func (s *QoSService) toParserClassMap(cm ClassMap) parsers.ClassMap {
	return parsers.ClassMap{
		Name:                 cm.Name,
		MatchProtocol:        cm.MatchProtocol,
		MatchDestinationPort: cm.MatchDestinationPort,
		MatchSourcePort:      cm.MatchSourcePort,
		MatchDSCP:            cm.MatchDSCP,
		MatchFilter:          cm.MatchFilter,
	}
}

func (s *QoSService) toParserPolicyMap(pm PolicyMap) parsers.PolicyMap {
	classes := make([]parsers.PolicyMapClass, len(pm.Classes))
	for i, c := range pm.Classes {
		classes[i] = parsers.PolicyMapClass{
			Name:             c.Name,
			Priority:         c.Priority,
			BandwidthPercent: c.BandwidthPercent,
			PoliceCIR:        c.PoliceCIR,
			QueueLimit:       c.QueueLimit,
		}
	}
	return parsers.PolicyMap{
		Name:    pm.Name,
		Classes: classes,
	}
}

func (s *QoSService) toParserServicePolicy(sp ServicePolicy) parsers.ServicePolicy {
	return parsers.ServicePolicy{
		Interface: sp.Interface,
		Direction: sp.Direction,
		PolicyMap: sp.PolicyMap,
	}
}

func (s *QoSService) fromParserServicePolicy(psp parsers.ServicePolicy) ServicePolicy {
	return ServicePolicy{
		Interface: psp.Interface,
		Direction: psp.Direction,
		PolicyMap: psp.PolicyMap,
	}
}

func (s *QoSService) toParserShapeConfig(sc ShapeConfig) parsers.ShapeConfig {
	return parsers.ShapeConfig{
		Interface:    sc.Interface,
		Direction:    sc.Direction,
		ShapeAverage: sc.ShapeAverage,
		ShapeBurst:   sc.ShapeBurst,
	}
}

func (s *QoSService) fromParserShapeConfig(psc parsers.ShapeConfig) ShapeConfig {
	return ShapeConfig{
		Interface:    psc.Interface,
		Direction:    psc.Direction,
		ShapeAverage: psc.ShapeAverage,
		ShapeBurst:   psc.ShapeBurst,
	}
}
