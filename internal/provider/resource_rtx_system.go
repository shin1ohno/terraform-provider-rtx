package provider

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXSystem() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages system-level settings on RTX routers. This is a singleton resource - there is only one system configuration per router.",
		CreateContext: resourceRTXSystemCreate,
		ReadContext:   resourceRTXSystemRead,
		UpdateContext: resourceRTXSystemUpdate,
		DeleteContext: resourceRTXSystemDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXSystemImport,
		},

		Schema: map[string]*schema.Schema{
			"timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Timezone as UTC offset (e.g., '+09:00' for JST, '-05:00' for EST)",
				ValidateFunc: validateTimezone,
			},
			"console": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Console settings",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"character": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Character encoding (ja.utf8, ja.sjis, ascii, euc-jp)",
							ValidateFunc: validation.StringInSlice([]string{"ja.utf8", "ja.sjis", "ascii", "euc-jp"}, false),
						},
						"lines": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Lines per page (positive integer or 'infinity')",
							ValidateFunc: validateConsoleLines,
						},
						"prompt": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Custom prompt string",
						},
					},
				},
			},
			"packet_buffer": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    3,
				Description: "Packet buffer tuning settings (small, middle, large)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Buffer size category (small, middle, large)",
							ValidateFunc: validation.StringInSlice([]string{"small", "middle", "large"}, false),
						},
						"max_buffer": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Maximum buffer count",
							ValidateFunc: validation.IntAtLeast(1),
						},
						"max_free": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Maximum free buffer count",
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},
			"statistics": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Statistics collection settings",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"traffic": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Enable traffic statistics collection",
						},
						"nat": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Enable NAT statistics collection",
						},
					},
				},
			},
		},
	}
}

func resourceRTXSystemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildSystemConfigFromResourceData(d)

	log.Printf("[DEBUG] Creating system configuration: %+v", config)

	err := apiClient.client.ConfigureSystem(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to configure system: %v", err)
	}

	// Use fixed ID "system" for singleton resource
	d.SetId("system")

	// Read back to ensure consistency
	return resourceRTXSystemRead(ctx, d, meta)
}

func resourceRTXSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	log.Printf("[DEBUG] Reading system configuration")

	config, err := apiClient.client.GetSystemConfig(ctx)
	if err != nil {
		return diag.Errorf("Failed to read system configuration: %v", err)
	}

	// Update the state
	if err := d.Set("timezone", config.Timezone); err != nil {
		return diag.FromErr(err)
	}

	// Convert Console to nested block
	if config.Console != nil {
		console := []map[string]interface{}{}
		if config.Console.Character != "" || config.Console.Lines != "" || config.Console.Prompt != "" {
			consoleMap := map[string]interface{}{
				"character": config.Console.Character,
				"lines":     config.Console.Lines,
				"prompt":    config.Console.Prompt,
			}
			console = append(console, consoleMap)
		}
		if err := d.Set("console", console); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err := d.Set("console", []map[string]interface{}{}); err != nil {
			return diag.FromErr(err)
		}
	}

	// Convert PacketBuffers to list
	packetBuffers := make([]map[string]interface{}, len(config.PacketBuffers))
	for i, pb := range config.PacketBuffers {
		packetBuffers[i] = map[string]interface{}{
			"size":       pb.Size,
			"max_buffer": pb.MaxBuffer,
			"max_free":   pb.MaxFree,
		}
	}
	if err := d.Set("packet_buffer", packetBuffers); err != nil {
		return diag.FromErr(err)
	}

	// Convert Statistics to nested block
	if config.Statistics != nil {
		statistics := []map[string]interface{}{
			{
				"traffic": config.Statistics.Traffic,
				"nat":     config.Statistics.NAT,
			},
		}
		if err := d.Set("statistics", statistics); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err := d.Set("statistics", []map[string]interface{}{}); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceRTXSystemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	config := buildSystemConfigFromResourceData(d)

	log.Printf("[DEBUG] Updating system configuration: %+v", config)

	err := apiClient.client.UpdateSystemConfig(ctx, config)
	if err != nil {
		return diag.Errorf("Failed to update system configuration: %v", err)
	}

	return resourceRTXSystemRead(ctx, d, meta)
}

func resourceRTXSystemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	log.Printf("[DEBUG] Deleting (resetting) system configuration")

	err := apiClient.client.ResetSystem(ctx)
	if err != nil {
		return diag.Errorf("Failed to reset system configuration: %v", err)
	}

	return nil
}

func resourceRTXSystemImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Only accept "system" as valid import ID (singleton resource)
	if importID != "system" {
		return nil, fmt.Errorf("invalid import ID format, expected 'system' for singleton resource, got: %s", importID)
	}

	log.Printf("[DEBUG] Importing system configuration")

	// Verify configuration exists and retrieve it
	config, err := apiClient.client.GetSystemConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to import system configuration: %v", err)
	}

	// Set the ID
	d.SetId("system")

	// Set all attributes
	d.Set("timezone", config.Timezone)

	// Set console
	if config.Console != nil {
		console := []map[string]interface{}{}
		if config.Console.Character != "" || config.Console.Lines != "" || config.Console.Prompt != "" {
			consoleMap := map[string]interface{}{
				"character": config.Console.Character,
				"lines":     config.Console.Lines,
				"prompt":    config.Console.Prompt,
			}
			console = append(console, consoleMap)
		}
		d.Set("console", console)
	}

	// Set packet_buffer
	packetBuffers := make([]map[string]interface{}, len(config.PacketBuffers))
	for i, pb := range config.PacketBuffers {
		packetBuffers[i] = map[string]interface{}{
			"size":       pb.Size,
			"max_buffer": pb.MaxBuffer,
			"max_free":   pb.MaxFree,
		}
	}
	d.Set("packet_buffer", packetBuffers)

	// Set statistics
	if config.Statistics != nil {
		statistics := []map[string]interface{}{
			{
				"traffic": config.Statistics.Traffic,
				"nat":     config.Statistics.NAT,
			},
		}
		d.Set("statistics", statistics)
	}

	return []*schema.ResourceData{d}, nil
}

// buildSystemConfigFromResourceData creates a SystemConfig from Terraform resource data
func buildSystemConfigFromResourceData(d *schema.ResourceData) client.SystemConfig {
	config := client.SystemConfig{
		Timezone:      d.Get("timezone").(string),
		PacketBuffers: []client.PacketBufferConfig{},
	}

	// Handle console block
	if v, ok := d.GetOk("console"); ok {
		consoleList := v.([]interface{})
		if len(consoleList) > 0 {
			consoleMap := consoleList[0].(map[string]interface{})
			config.Console = &client.ConsoleConfig{
				Character: consoleMap["character"].(string),
				Lines:     consoleMap["lines"].(string),
				Prompt:    consoleMap["prompt"].(string),
			}
		}
	}

	// Handle packet_buffer list
	if v, ok := d.GetOk("packet_buffer"); ok {
		pbList := v.([]interface{})
		for _, pbRaw := range pbList {
			pbMap := pbRaw.(map[string]interface{})
			config.PacketBuffers = append(config.PacketBuffers, client.PacketBufferConfig{
				Size:      pbMap["size"].(string),
				MaxBuffer: pbMap["max_buffer"].(int),
				MaxFree:   pbMap["max_free"].(int),
			})
		}
	}

	// Handle statistics block
	if v, ok := d.GetOk("statistics"); ok {
		statsList := v.([]interface{})
		if len(statsList) > 0 {
			statsMap := statsList[0].(map[string]interface{})
			config.Statistics = &client.StatisticsConfig{
				Traffic: statsMap["traffic"].(bool),
				NAT:     statsMap["nat"].(bool),
			}
		}
	}

	return config
}

// validateTimezone validates that a string is a valid timezone format (±HH:MM)
func validateTimezone(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, nil
	}

	pattern := regexp.MustCompile(`^[\+\-]\d{2}:\d{2}$`)
	if !pattern.MatchString(value) {
		return nil, []error{fmt.Errorf("%q must be a valid UTC offset (e.g., '+09:00', '-05:00')", k)}
	}

	return nil, nil
}

// validateConsoleLines validates the console lines setting
func validateConsoleLines(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, nil
	}

	if value == "infinity" {
		return nil, nil
	}

	// Try to parse as positive integer
	lines := strings.TrimSpace(value)
	n, err := strconv.Atoi(lines)
	if err != nil || n <= 0 {
		return nil, []error{fmt.Errorf("%q must be a positive integer or 'infinity'", k)}
	}

	return nil, nil
}
