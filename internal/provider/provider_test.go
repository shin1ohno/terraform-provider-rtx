package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProvider_Schema(t *testing.T) {
	p := New("test")

	// Test that provider schema is valid
	if err := p.InternalValidate(); err != nil {
		t.Fatalf("provider internal validation failed: %v", err)
	}
}

func TestProvider_HasAllExpectedSchemaFields(t *testing.T) {
	p := New("test")

	expectedFields := []string{
		"host",
		"username",
		"password",
		"admin_password",
		"port",
		"timeout",
		"ssh_host_key",
		"known_hosts_file",
		"skip_host_key_check",
		"max_parallelism",
		"use_sftp",
		"sftp_config_path",
	}

	for _, field := range expectedFields {
		if _, ok := p.Schema[field]; !ok {
			t.Errorf("expected schema field %q to exist", field)
		}
	}
}

func TestProvider_UseSFTPSchema(t *testing.T) {
	p := New("test")

	useSFTPSchema, ok := p.Schema["use_sftp"]
	if !ok {
		t.Fatal("expected use_sftp schema to exist")
	}

	if useSFTPSchema.Type != schema.TypeBool {
		t.Errorf("expected use_sftp type to be TypeBool, got %v", useSFTPSchema.Type)
	}

	if useSFTPSchema.Required {
		t.Error("expected use_sftp to be optional")
	}

	if useSFTPSchema.DefaultFunc == nil {
		t.Error("expected use_sftp to have DefaultFunc for environment variable")
	}
}

func TestProvider_SFTPConfigPathSchema(t *testing.T) {
	p := New("test")

	sftpConfigPathSchema, ok := p.Schema["sftp_config_path"]
	if !ok {
		t.Fatal("expected sftp_config_path schema to exist")
	}

	if sftpConfigPathSchema.Type != schema.TypeString {
		t.Errorf("expected sftp_config_path type to be TypeString, got %v", sftpConfigPathSchema.Type)
	}

	if sftpConfigPathSchema.Required {
		t.Error("expected sftp_config_path to be optional")
	}

	if sftpConfigPathSchema.DefaultFunc == nil {
		t.Error("expected sftp_config_path to have DefaultFunc for environment variable")
	}
}

func TestProvider_UseSFTPEnvDefault(t *testing.T) {
	p := New("test")

	useSFTPSchema := p.Schema["use_sftp"]

	// Test default value (no env var set)
	if err := os.Unsetenv("RTX_USE_SFTP"); err != nil {
		t.Fatalf("failed to unset env var: %v", err)
	}
	val, err := useSFTPSchema.DefaultFunc()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != false {
		t.Errorf("expected default value to be false, got %v", val)
	}

	// Test with env var set to true
	if err := os.Setenv("RTX_USE_SFTP", "true"); err != nil {
		t.Fatalf("failed to set env var: %v", err)
	}
	defer func() { _ = os.Unsetenv("RTX_USE_SFTP") }()
	val, err = useSFTPSchema.DefaultFunc()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "true" {
		t.Errorf("expected value from env var to be 'true', got %v", val)
	}
}

func TestProvider_SFTPConfigPathEnvDefault(t *testing.T) {
	p := New("test")

	sftpConfigPathSchema := p.Schema["sftp_config_path"]

	// Test default value (no env var set)
	if err := os.Unsetenv("RTX_SFTP_CONFIG_PATH"); err != nil {
		t.Fatalf("failed to unset env var: %v", err)
	}
	val, err := sftpConfigPathSchema.DefaultFunc()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "" {
		t.Errorf("expected default value to be empty string, got %v", val)
	}

	// Test with env var set
	if err := os.Setenv("RTX_SFTP_CONFIG_PATH", "/system/config0"); err != nil {
		t.Fatalf("failed to set env var: %v", err)
	}
	defer func() { _ = os.Unsetenv("RTX_SFTP_CONFIG_PATH") }()
	val, err = sftpConfigPathSchema.DefaultFunc()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "/system/config0" {
		t.Errorf("expected value from env var to be '/system/config0', got %v", val)
	}
}
