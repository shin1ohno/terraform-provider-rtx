package client

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DefaultConfigPath is the fallback path when config number cannot be detected
// RTX routers store config files in /system/configN format
// Note: Absolute path with leading "/" is required for RTX SFTP access
const DefaultConfigPath = "/system/config0"

// ConfigPathResolver resolves the SFTP path for the router's startup configuration file.
// It executes "show environment" and parses the output to determine the active config number.
type ConfigPathResolver struct {
	executor Executor
}

// NewConfigPathResolver creates a new ConfigPathResolver with the given executor.
func NewConfigPathResolver(executor Executor) *ConfigPathResolver {
	return &ConfigPathResolver{
		executor: executor,
	}
}

// Resolve determines the SFTP path for the router's configuration file.
// It executes "show environment" and parses the output to find the default config number.
// Returns the path in format "/system/configN" where N is the config number.
// Falls back to "/system/config0" if detection fails for any reason.
func (r *ConfigPathResolver) Resolve(ctx context.Context) (string, error) {
	output, err := r.executor.Run(ctx, "show environment")
	if err != nil {
		// Fallback on execution error
		return DefaultConfigPath, nil
	}

	configNum, found := parseConfigNumber(string(output))
	if !found {
		// Fallback if config number not found in output
		return DefaultConfigPath, nil
	}

	return fmt.Sprintf("/system/config%d", configNum), nil
}

// parseConfigNumber extracts the config number from show environment output.
// It handles both Japanese ("デフォルト設定ファイル") and English ("Default config file") formats.
// Returns the config number and true if found, otherwise 0 and false.
func parseConfigNumber(output string) (int, bool) {
	if output == "" {
		return 0, false
	}

	// Patterns to match both Japanese and English output formats
	// Japanese: "デフォルト設定ファイル: configN"
	// English: "Default config file: configN"
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)デフォルト設定ファイル:\s*config(\d+)`),
		regexp.MustCompile(`(?i)default\s+config\s+file:\s*config(\d+)`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(output)
		if len(matches) >= 2 {
			numStr := strings.TrimSpace(matches[1])
			num, err := strconv.Atoi(numStr)
			if err == nil {
				return num, true
			}
		}
	}

	return 0, false
}
