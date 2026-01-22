package client

import (
	"bytes"
	"regexp"
)

// defaultPromptDetector is the default implementation for detecting RTX router prompts
type defaultPromptDetector struct {
	pattern *regexp.Regexp
}

// NewDefaultPromptDetector creates a new prompt detector with default RTX patterns
func NewDefaultPromptDetector() PromptDetector {
	// RTX routers typically end with hostname# or hostname>
	// Based on Ansible RTX terminal plugin: [>#]\s*$
	pattern := regexp.MustCompile(`[>#]\s*$`)
	return &defaultPromptDetector{pattern: pattern}
}

// DetectPrompt checks if the output contains a router prompt
func (d *defaultPromptDetector) DetectPrompt(output []byte) (matched bool, prompt string) {
	matches := d.pattern.FindAll(output, -1)
	if len(matches) == 0 {
		return false, ""
	}

	// Return the last match as the prompt
	lastMatch := matches[len(matches)-1]
	return true, string(bytes.TrimSpace(lastMatch))
}

// customPromptDetector allows for custom prompt patterns
type customPromptDetector struct {
	pattern *regexp.Regexp
}

// NewCustomPromptDetector creates a prompt detector with a custom pattern
func NewCustomPromptDetector(pattern string) (PromptDetector, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &customPromptDetector{pattern: re}, nil
}

// DetectPrompt checks if the output contains the custom prompt pattern
func (d *customPromptDetector) DetectPrompt(output []byte) (matched bool, prompt string) {
	match := d.pattern.Find(output)
	if match == nil {
		return false, ""
	}
	return true, string(bytes.TrimSpace(match))
}
