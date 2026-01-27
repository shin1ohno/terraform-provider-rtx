package sshd_host_key

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &SSHDHostKeyResource{}
	_ resource.ResourceWithImportState = &SSHDHostKeyResource{}
)

// NewSSHDHostKeyResource creates a new SSHD host key resource.
func NewSSHDHostKeyResource() resource.Resource {
	return &SSHDHostKeyResource{}
}

// SSHDHostKeyResource defines the resource implementation.
type SSHDHostKeyResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *SSHDHostKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sshd_host_key"
}

// Schema defines the schema for the resource.
func (r *SSHDHostKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages SSH host key on RTX routers. " +
			"This is a singleton resource - only one instance should exist per router. " +
			"The host key is used for SSH server authentication. " +
			"If no host key exists, creating this resource will generate one.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this singleton resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"fingerprint": schema.StringAttribute{
				Description: "SSH host key fingerprint.",
				Computed:    true,
			},
			"algorithm": schema.StringAttribute{
				Description: "Host key algorithm (e.g., ssh-rsa).",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SSHDHostKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*fwhelpers.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *fwhelpers.ProviderData, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = providerData.Client
}

// Create creates the resource and sets the initial Terraform state.
func (r *SSHDHostKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SSHDHostKeyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_sshd_host_key", "sshd_host_key")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_sshd_host_key").Msg("Creating SSHD host key resource")

	// Get current host key info
	keyInfo, err := r.client.GetSSHDHostKey(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get SSHD host key info",
			fmt.Sprintf("Could not get SSHD host key info: %v", err),
		)
		return
	}

	// If no key exists (empty fingerprint), generate one
	if keyInfo.Fingerprint == "" {
		logger.Info().Str("resource", "rtx_sshd_host_key").Msg("No host key exists, generating new SSHD host key")

		err = r.client.GenerateSSHDHostKey(ctx)
		if err != nil {
			// Check if error indicates key already exists (safety net triggered)
			// This can happen if the parser failed to detect the existing key
			if strings.Contains(err.Error(), "already exists") {
				logger.Info().Str("resource", "rtx_sshd_host_key").Msg("Host key already exists on router (detected during generation), reading existing key info")
				// Re-read to get the existing key info
				keyInfo, err = r.client.GetSSHDHostKey(ctx)
				if err != nil {
					resp.Diagnostics.AddError(
						"Failed to read existing SSHD host key info",
						fmt.Sprintf("Could not read existing SSHD host key info: %v", err),
					)
					return
				}
			} else {
				resp.Diagnostics.AddError(
					"Failed to generate SSHD host key",
					fmt.Sprintf("Could not generate SSHD host key: %v", err),
				)
				return
			}
		} else {
			// Get the new key info
			keyInfo, err = r.client.GetSSHDHostKey(ctx)
			if err != nil {
				resp.Diagnostics.AddError(
					"Failed to get newly generated SSHD host key info",
					fmt.Sprintf("Could not get newly generated SSHD host key info: %v", err),
				)
				return
			}
		}
	}

	data.FromClient(keyInfo)

	logger.Debug().
		Str("resource", "rtx_sshd_host_key").
		Str("fingerprint", keyInfo.Fingerprint).
		Str("algorithm", keyInfo.Algorithm).
		Msg("SSHD host key resource created")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *SSHDHostKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SSHDHostKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_sshd_host_key", data.ID.ValueString())
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_sshd_host_key").Msg("Reading SSHD host key configuration")

	// Get current host key info
	keyInfo, err := r.client.GetSSHDHostKey(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get SSHD host key info",
			fmt.Sprintf("Could not get SSHD host key info: %v", err),
		)
		return
	}

	// If fingerprint is empty, the key doesn't exist - remove from state
	if keyInfo.Fingerprint == "" {
		logger.Debug().Str("resource", "rtx_sshd_host_key").Msg("SSHD host key not found, removing from state")
		resp.State.RemoveResource(ctx)
		return
	}

	data.FromClient(keyInfo)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SSHDHostKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This resource has no configurable attributes, so Update is a no-op
	var data SSHDHostKeyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *SSHDHostKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SSHDHostKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_sshd_host_key", data.ID.ValueString())
	logger := logging.FromContext(ctx)

	// No-op - host keys should persist on the router
	// Deleting from Terraform state doesn't delete the actual key
	logger.Debug().Str("resource", "rtx_sshd_host_key").Msg("Removing SSHD host key from Terraform state (key persists on router)")
}

// ImportState imports an existing resource into Terraform.
func (r *SSHDHostKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ctx = logging.WithResource(ctx, "rtx_sshd_host_key", "sshd_host_key")
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_sshd_host_key").Msg("Importing SSHD host key configuration")

	// Set the ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue("sshd_host_key"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Verify host key exists and get info
	keyInfo, err := r.client.GetSSHDHostKey(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to import SSHD host key",
			fmt.Sprintf("Could not get SSHD host key info: %v", err),
		)
		return
	}

	// If no host key exists, return error
	if keyInfo.Fingerprint == "" {
		resp.Diagnostics.AddError(
			"No SSHD host key found",
			"No SSHD host key found on router - generate one first or use 'terraform apply' to create",
		)
		return
	}

	// Set attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("fingerprint"), types.StringValue(keyInfo.Fingerprint))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("algorithm"), types.StringValue(keyInfo.Algorithm))...)
}
