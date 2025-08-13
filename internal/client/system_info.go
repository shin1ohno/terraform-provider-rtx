package client

import (
	"context"
	"regexp"
	"strings"
)

// SystemInfo represents RTX router system information
type SystemInfo struct {
	Model           string
	FirmwareVersion string
	SerialNumber    string
	MACAddress      string
	Uptime          string
}

// GetSystemInfo retrieves system information from the router
func (c *rtxClient) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	// RTX routers use "show environment" instead of "show system information"
	cmd := Command{
		Key:     "show environment",
		Payload: "show environment",
	}
	
	result, err := c.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}
	
	// Parse the raw output
	info := parseSystemInfo(string(result.Raw))
	return info, nil
}

// parseSystemInfo parses the output of "show environment" command
func parseSystemInfo(output string) *SystemInfo {
	info := &SystemInfo{}
	
	// Model extraction (e.g., "RTX1210 Rev.14.01.42")
	if match := regexp.MustCompile(`(RTX\d+)\s+Rev\.`).FindStringSubmatch(output); len(match) > 1 {
		info.Model = match[1]
	}
	
	// Firmware version extraction (e.g., "RTX1210 Rev.14.01.42")
	if match := regexp.MustCompile(`RTX\d+\s+Rev\.([\d.]+)`).FindStringSubmatch(output); len(match) > 1 {
		info.FirmwareVersion = match[1]
	}
	
	// Serial number extraction (e.g., "serial=S4H104289")
	if match := regexp.MustCompile(`serial=([A-Z0-9]+)`).FindStringSubmatch(output); len(match) > 1 {
		info.SerialNumber = match[1]
	}
	
	// MAC address extraction - take the first one (e.g., "MAC-Address=ac:44:f2:3a:2a:fd")
	if match := regexp.MustCompile(`MAC-Address=([0-9a-fA-F:]+)`).FindStringSubmatch(output); len(match) > 1 {
		info.MACAddress = strings.ToLower(match[1])
	}
	
	// Uptime extraction (e.g., "Elapsed time from boot: 14days 12:18:49")
	if match := regexp.MustCompile(`Elapsed time from boot:\s*(.+)`).FindStringSubmatch(output); len(match) > 1 {
		info.Uptime = strings.TrimSpace(match[1])
	}
	
	return info
}