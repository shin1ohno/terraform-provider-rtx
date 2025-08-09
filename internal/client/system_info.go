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
	cmd := Command{
		Key:     "show system information",
		Payload: "show system information",
	}
	
	result, err := c.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}
	
	// Parse the raw output
	info := parseSystemInfo(string(result.Raw))
	return info, nil
}

// parseSystemInfo parses the output of "show system information" command
func parseSystemInfo(output string) *SystemInfo {
	info := &SystemInfo{}
	
	// Model extraction (e.g., "RTX1210" or "Model: RTX1210")
	if match := regexp.MustCompile(`(?:Model:\s*)?RTX\d+`).FindString(output); match != "" {
		info.Model = strings.TrimPrefix(match, "Model: ")
		info.Model = strings.TrimSpace(info.Model)
	}
	
	// Firmware version extraction
	if match := regexp.MustCompile(`(?:Firmware Version:|Rev\.)\s*([\d.]+)`).FindStringSubmatch(output); len(match) > 1 {
		info.FirmwareVersion = match[1]
	}
	
	// Serial number extraction
	if match := regexp.MustCompile(`(?:Serial Number:|serial=)([A-Z0-9]+)`).FindStringSubmatch(output); len(match) > 1 {
		info.SerialNumber = match[1]
	}
	
	// MAC address extraction
	if match := regexp.MustCompile(`(?:MAC Address:|MAC-Address=)([0-9A-Fa-f:]+)`).FindStringSubmatch(output); len(match) > 1 {
		info.MACAddress = strings.ToUpper(match[1])
	}
	
	// Uptime extraction
	if match := regexp.MustCompile(`Uptime:\s*(.+)`).FindStringSubmatch(output); len(match) > 1 {
		info.Uptime = strings.TrimSpace(match[1])
	}
	
	return info
}