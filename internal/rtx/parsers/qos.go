package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// QoSConfig represents QoS configuration on an RTX router
type QoSConfig struct {
	Interface    string     `json:"interface"`               // Interface name (lan1, wan1, etc.)
	QueueType    string     `json:"queue_type,omitempty"`    // Queue type (priority, cbq, etc.)
	Classes      []QoSClass `json:"classes,omitempty"`       // QoS classes
	ShapeAverage int        `json:"shape_average,omitempty"` // Shape average rate in bps
	ShapeBurst   int        `json:"shape_burst,omitempty"`   // Shape burst size in bytes
	Speed        int        `json:"speed,omitempty"`         // Interface speed in bps
}

// QoSClass represents a QoS class configuration
type QoSClass struct {
	Name             string `json:"name"`                        // Class name
	Filter           int    `json:"filter,omitempty"`            // Filter number
	Priority         string `json:"priority,omitempty"`          // Priority level (high, normal, low)
	BandwidthPercent int    `json:"bandwidth_percent,omitempty"` // Bandwidth percentage (1-100)
	PoliceCIR        int    `json:"police_cir,omitempty"`        // Committed Information Rate in bps
	QueueLimit       int    `json:"queue_limit,omitempty"`       // Queue limit (depth)
}

// ClassMap represents a class-map configuration for traffic classification
type ClassMap struct {
	Name                 string `json:"name"`                             // Class map name
	MatchProtocol        string `json:"match_protocol,omitempty"`         // Protocol to match (http, sip, etc.)
	MatchDestinationPort []int  `json:"match_destination_port,omitempty"` // Destination ports to match
	MatchSourcePort      []int  `json:"match_source_port,omitempty"`      // Source ports to match
	MatchDSCP            string `json:"match_dscp,omitempty"`             // DSCP value to match
	MatchFilter          int    `json:"match_filter,omitempty"`           // IP filter number to match
}

// PolicyMap represents a policy-map configuration
type PolicyMap struct {
	Name    string           `json:"name"`              // Policy map name
	Classes []PolicyMapClass `json:"classes,omitempty"` // Policy map classes
}

// PolicyMapClass represents a class within a policy map
type PolicyMapClass struct {
	Name             string `json:"name"`                        // Class name (references class-map)
	Priority         string `json:"priority,omitempty"`          // Priority level
	BandwidthPercent int    `json:"bandwidth_percent,omitempty"` // Bandwidth percentage
	PoliceCIR        int    `json:"police_cir,omitempty"`        // Committed Information Rate
	QueueLimit       int    `json:"queue_limit,omitempty"`       // Queue limit
}

// ServicePolicy represents a service-policy attachment to an interface
type ServicePolicy struct {
	Interface string `json:"interface"`  // Interface name
	Direction string `json:"direction"`  // input or output
	PolicyMap string `json:"policy_map"` // Policy map name
}

// ShapeConfig represents traffic shaping configuration
type ShapeConfig struct {
	Interface    string `json:"interface"`             // Interface name
	Direction    string `json:"direction"`             // input or output
	ShapeAverage int    `json:"shape_average"`         // Average rate in bps
	ShapeBurst   int    `json:"shape_burst,omitempty"` // Burst size in bytes
}

// QoSParser parses QoS configuration output
type QoSParser struct{}

// NewQoSParser creates a new QoS parser
func NewQoSParser() *QoSParser {
	return &QoSParser{}
}

// ParseQoSConfig parses the output of "show config" command for QoS configuration
func (p *QoSParser) ParseQoSConfig(raw string, iface string) (*QoSConfig, error) {
	config := &QoSConfig{
		Interface: iface,
		Classes:   []QoSClass{},
	}

	lines := strings.Split(raw, "\n")

	// Patterns for QoS configuration lines
	// queue <interface> type <type>
	queueTypePattern := regexp.MustCompile(`^\s*queue\s+` + regexp.QuoteMeta(iface) + `\s+type\s+(\S+)\s*$`)
	// queue <interface> class filter <n> <filter>
	classFilterPattern := regexp.MustCompile(`^\s*queue\s+` + regexp.QuoteMeta(iface) + `\s+class\s+filter\s+(\d+)\s+(\d+)\s*$`)
	// queue <interface> class priority <class> <priority>
	classPriorityPattern := regexp.MustCompile(`^\s*queue\s+` + regexp.QuoteMeta(iface) + `\s+class\s+priority\s+(\d+)\s+(\S+)\s*$`)
	// queue <interface> length <class> <length>
	queueLengthPattern := regexp.MustCompile(`^\s*queue\s+` + regexp.QuoteMeta(iface) + `\s+length\s+(\d+)\s+(\d+)\s*$`)
	// speed <interface> <speed>
	speedPattern := regexp.MustCompile(`^\s*speed\s+` + regexp.QuoteMeta(iface) + `\s+(\d+)\s*$`)

	classMap := make(map[int]*QoSClass) // key: class number

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try queue type pattern
		if matches := queueTypePattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.QueueType = matches[1]
			continue
		}

		// Try class filter pattern
		if matches := classFilterPattern.FindStringSubmatch(line); len(matches) >= 3 {
			classNum, _ := strconv.Atoi(matches[1])
			filterNum, _ := strconv.Atoi(matches[2])

			class, exists := classMap[classNum]
			if !exists {
				class = &QoSClass{Name: fmt.Sprintf("class%d", classNum)}
				classMap[classNum] = class
			}
			class.Filter = filterNum
			continue
		}

		// Try class priority pattern
		if matches := classPriorityPattern.FindStringSubmatch(line); len(matches) >= 3 {
			classNum, _ := strconv.Atoi(matches[1])
			priority := matches[2]

			class, exists := classMap[classNum]
			if !exists {
				class = &QoSClass{Name: fmt.Sprintf("class%d", classNum)}
				classMap[classNum] = class
			}
			class.Priority = priority
			continue
		}

		// Try queue length pattern
		if matches := queueLengthPattern.FindStringSubmatch(line); len(matches) >= 3 {
			classNum, _ := strconv.Atoi(matches[1])
			length, _ := strconv.Atoi(matches[2])

			class, exists := classMap[classNum]
			if !exists {
				class = &QoSClass{Name: fmt.Sprintf("class%d", classNum)}
				classMap[classNum] = class
			}
			class.QueueLimit = length
			continue
		}

		// Try speed pattern
		if matches := speedPattern.FindStringSubmatch(line); len(matches) >= 2 {
			speed, _ := strconv.Atoi(matches[1])
			config.Speed = speed
			continue
		}
	}

	// Convert class map to slice
	for _, class := range classMap {
		config.Classes = append(config.Classes, *class)
	}

	return config, nil
}

// ParseClassMap parses class-map configuration
func (p *QoSParser) ParseClassMap(raw string, name string) (*ClassMap, error) {
	cm := &ClassMap{
		Name:                 name,
		MatchDestinationPort: []int{},
		MatchSourcePort:      []int{},
	}

	lines := strings.Split(raw, "\n")

	// Patterns for class-map
	// queue class filter <n> protocol=<protocol>
	protocolPattern := regexp.MustCompile(`queue\s+\S+\s+class\s+filter\s+\d+\s+protocol=(\S+)`)
	// queue class filter <n> dstport=<port>
	dstPortPattern := regexp.MustCompile(`queue\s+\S+\s+class\s+filter\s+\d+\s+.*dstport=(\d+)`)
	// ip filter <n> ... (for filter-based matching)
	filterPattern := regexp.MustCompile(`ip\s+filter\s+(\d+)\s+`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try protocol pattern
		if matches := protocolPattern.FindStringSubmatch(line); len(matches) >= 2 {
			cm.MatchProtocol = matches[1]
			continue
		}

		// Try destination port pattern
		if matches := dstPortPattern.FindStringSubmatch(line); len(matches) >= 2 {
			port, _ := strconv.Atoi(matches[1])
			cm.MatchDestinationPort = append(cm.MatchDestinationPort, port)
			continue
		}

		// Try filter pattern
		if matches := filterPattern.FindStringSubmatch(line); len(matches) >= 2 {
			filter, _ := strconv.Atoi(matches[1])
			cm.MatchFilter = filter
			continue
		}
	}

	return cm, nil
}

// ParseServicePolicy parses service-policy configuration
func (p *QoSParser) ParseServicePolicy(raw string, iface string) (*ServicePolicy, error) {
	sp := &ServicePolicy{
		Interface: iface,
	}

	lines := strings.Split(raw, "\n")

	// Pattern for service policy (RTX uses queue command for this)
	// queue <interface> type priority
	queuePattern := regexp.MustCompile(`^\s*queue\s+` + regexp.QuoteMeta(iface) + `\s+type\s+(\S+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if matches := queuePattern.FindStringSubmatch(line); len(matches) >= 2 {
			sp.PolicyMap = matches[1]
			sp.Direction = "output" // RTX queuing is typically output-based
			break
		}
	}

	if sp.PolicyMap == "" {
		return nil, fmt.Errorf("service policy not found for interface %s", iface)
	}

	return sp, nil
}

// ParseShapeConfig parses traffic shaping configuration
func (p *QoSParser) ParseShapeConfig(raw string, iface string) (*ShapeConfig, error) {
	sc := &ShapeConfig{
		Interface: iface,
		Direction: "output", // Default direction
	}

	lines := strings.Split(raw, "\n")

	// Pattern for speed command
	speedPattern := regexp.MustCompile(`^\s*speed\s+` + regexp.QuoteMeta(iface) + `\s+(\d+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if matches := speedPattern.FindStringSubmatch(line); len(matches) >= 2 {
			speed, _ := strconv.Atoi(matches[1])
			sc.ShapeAverage = speed
			break
		}
	}

	if sc.ShapeAverage == 0 {
		return nil, fmt.Errorf("shape configuration not found for interface %s", iface)
	}

	return sc, nil
}

// BuildQueueTypeCommand builds the command to set queue type
// Command format: queue <interface> type <type>
func BuildQueueTypeCommand(iface string, queueType string) string {
	return fmt.Sprintf("queue %s type %s", iface, queueType)
}

// BuildQueueClassFilterCommand builds the command to set class filter
// Command format: queue <interface> class filter <n> <filter>
func BuildQueueClassFilterCommand(iface string, classNum int, filterNum int) string {
	return fmt.Sprintf("queue %s class filter %d %d", iface, classNum, filterNum)
}

// BuildQueueClassPriorityCommand builds the command to set class priority
// Command format: queue <interface> class priority <class> <priority>
func BuildQueueClassPriorityCommand(iface string, classNum int, priority string) string {
	return fmt.Sprintf("queue %s class priority %d %s", iface, classNum, priority)
}

// BuildSpeedCommand builds the command to set interface speed (traffic shaping)
// Command format: speed <interface> <bandwidth>
func BuildSpeedCommand(iface string, bandwidth int) string {
	return fmt.Sprintf("speed %s %d", iface, bandwidth)
}

// BuildQueueLengthCommand builds the command to set queue length
// Command format: queue <interface> length <class> <length>
func BuildQueueLengthCommand(iface string, classNum int, length int) string {
	return fmt.Sprintf("queue %s length %d %d", iface, classNum, length)
}

// BuildDeleteQueueTypeCommand builds the command to delete queue type
// Command format: no queue <interface> type
func BuildDeleteQueueTypeCommand(iface string) string {
	return fmt.Sprintf("no queue %s type", iface)
}

// BuildDeleteQueueClassFilterCommand builds the command to delete class filter
// Command format: no queue <interface> class filter <n>
func BuildDeleteQueueClassFilterCommand(iface string, classNum int) string {
	return fmt.Sprintf("no queue %s class filter %d", iface, classNum)
}

// BuildDeleteQueueClassPriorityCommand builds the command to delete class priority
// Command format: no queue <interface> class priority <n>
func BuildDeleteQueueClassPriorityCommand(iface string, classNum int) string {
	return fmt.Sprintf("no queue %s class priority %d", iface, classNum)
}

// BuildDeleteSpeedCommand builds the command to delete speed setting
// Command format: no speed <interface>
func BuildDeleteSpeedCommand(iface string) string {
	return fmt.Sprintf("no speed %s", iface)
}

// BuildDeleteQueueLengthCommand builds the command to delete queue length
// Command format: no queue <interface> length <class>
func BuildDeleteQueueLengthCommand(iface string, classNum int) string {
	return fmt.Sprintf("no queue %s length %d", iface, classNum)
}

// BuildDeleteQoSCommand builds all commands to remove QoS configuration
// This removes queue type and speed settings
func BuildDeleteQoSCommand(iface string) []string {
	return []string{
		fmt.Sprintf("no queue %s type", iface),
		fmt.Sprintf("no speed %s", iface),
	}
}

// BuildShowQoSCommand builds the command to show QoS configuration
func BuildShowQoSCommand(iface string) string {
	return fmt.Sprintf("show config | grep \"queue %s\\|speed %s\"", iface, iface)
}

// BuildShowAllQoSCommand builds the command to show all QoS configuration
func BuildShowAllQoSCommand() string {
	return "show config | grep \"queue\\|speed\""
}

// ValidateQoSConfig validates QoS configuration
func ValidateQoSConfig(config QoSConfig) error {
	// Validate interface
	if config.Interface == "" {
		return fmt.Errorf("interface is required")
	}

	// Validate queue type if specified
	validQueueTypes := map[string]bool{
		"priority": true,
		"cbq":      true,
		"fifo":     true,
		"shaping":  true,
	}
	if config.QueueType != "" && !validQueueTypes[config.QueueType] {
		return fmt.Errorf("invalid queue type %q, must be one of: priority, cbq, fifo, shaping", config.QueueType)
	}

	// Validate classes
	totalBandwidth := 0
	for i, class := range config.Classes {
		// Validate priority
		validPriorities := map[string]bool{
			"high":   true,
			"normal": true,
			"low":    true,
		}
		if class.Priority != "" && !validPriorities[class.Priority] {
			return fmt.Errorf("class %d: invalid priority %q, must be one of: high, normal, low", i, class.Priority)
		}

		// Validate bandwidth percent
		if class.BandwidthPercent < 0 || class.BandwidthPercent > 100 {
			return fmt.Errorf("class %d: bandwidth_percent must be between 0 and 100, got %d", i, class.BandwidthPercent)
		}
		totalBandwidth += class.BandwidthPercent

		// Validate queue limit
		if class.QueueLimit < 0 {
			return fmt.Errorf("class %d: queue_limit must be non-negative, got %d", i, class.QueueLimit)
		}
	}

	// Validate total bandwidth doesn't exceed 100%
	if totalBandwidth > 100 {
		return fmt.Errorf("total bandwidth_percent across all classes (%d%%) exceeds 100%%", totalBandwidth)
	}

	// Validate speed
	if config.Speed < 0 {
		return fmt.Errorf("speed must be non-negative, got %d", config.Speed)
	}

	// Validate shape average
	if config.ShapeAverage < 0 {
		return fmt.Errorf("shape_average must be non-negative, got %d", config.ShapeAverage)
	}

	return nil
}

// ValidateClassMap validates class-map configuration
func ValidateClassMap(cm ClassMap) error {
	if cm.Name == "" {
		return fmt.Errorf("class map name is required")
	}

	// Validate name format (alphanumeric and underscore only)
	validNamePattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !validNamePattern.MatchString(cm.Name) {
		return fmt.Errorf("class map name must start with a letter and contain only letters, numbers, underscores, and hyphens")
	}

	// Validate destination ports
	for _, port := range cm.MatchDestinationPort {
		if port < 1 || port > 65535 {
			return fmt.Errorf("invalid destination port %d, must be between 1 and 65535", port)
		}
	}

	// Validate source ports
	for _, port := range cm.MatchSourcePort {
		if port < 1 || port > 65535 {
			return fmt.Errorf("invalid source port %d, must be between 1 and 65535", port)
		}
	}

	return nil
}

// ValidatePolicyMap validates policy-map configuration
func ValidatePolicyMap(pm PolicyMap) error {
	if pm.Name == "" {
		return fmt.Errorf("policy map name is required")
	}

	// Validate name format
	validNamePattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !validNamePattern.MatchString(pm.Name) {
		return fmt.Errorf("policy map name must start with a letter and contain only letters, numbers, underscores, and hyphens")
	}

	// Validate classes
	totalBandwidth := 0
	for i, class := range pm.Classes {
		if class.Name == "" {
			return fmt.Errorf("class %d: name is required", i)
		}

		// Validate priority
		validPriorities := map[string]bool{
			"high":   true,
			"normal": true,
			"low":    true,
			"":       true, // Optional
		}
		if !validPriorities[class.Priority] {
			return fmt.Errorf("class %d: invalid priority %q", i, class.Priority)
		}

		// Validate bandwidth percent
		if class.BandwidthPercent < 0 || class.BandwidthPercent > 100 {
			return fmt.Errorf("class %d: bandwidth_percent must be between 0 and 100", i)
		}
		totalBandwidth += class.BandwidthPercent

		// Validate police CIR
		if class.PoliceCIR < 0 {
			return fmt.Errorf("class %d: police_cir must be non-negative", i)
		}

		// Validate queue limit
		if class.QueueLimit < 0 {
			return fmt.Errorf("class %d: queue_limit must be non-negative", i)
		}
	}

	// Validate total bandwidth
	if totalBandwidth > 100 {
		return fmt.Errorf("total bandwidth_percent across all classes (%d%%) exceeds 100%%", totalBandwidth)
	}

	return nil
}

// ValidateServicePolicy validates service-policy configuration
func ValidateServicePolicy(sp ServicePolicy) error {
	if sp.Interface == "" {
		return fmt.Errorf("interface is required")
	}

	// Validate direction
	validDirections := map[string]bool{
		"input":  true,
		"output": true,
	}
	if !validDirections[sp.Direction] {
		return fmt.Errorf("direction must be 'input' or 'output', got %q", sp.Direction)
	}

	if sp.PolicyMap == "" {
		return fmt.Errorf("policy_map is required")
	}

	return nil
}

// ValidateShapeConfig validates shape configuration
func ValidateShapeConfig(sc ShapeConfig) error {
	if sc.Interface == "" {
		return fmt.Errorf("interface is required")
	}

	// Validate direction
	validDirections := map[string]bool{
		"input":  true,
		"output": true,
	}
	if !validDirections[sc.Direction] {
		return fmt.Errorf("direction must be 'input' or 'output', got %q", sc.Direction)
	}

	if sc.ShapeAverage <= 0 {
		return fmt.Errorf("shape_average must be positive, got %d", sc.ShapeAverage)
	}

	if sc.ShapeBurst < 0 {
		return fmt.Errorf("shape_burst must be non-negative, got %d", sc.ShapeBurst)
	}

	return nil
}
