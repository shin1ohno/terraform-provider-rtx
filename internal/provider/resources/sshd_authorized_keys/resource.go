package sshd_authorized_keys

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &SSHDAuthorizedKeysResource{}
	_ resource.ResourceWithImportState = &SSHDAuthorizedKeysResource{}
)

// NewSSHDAuthorizedKeysResource creates a new SSHD authorized keys resource.
func NewSSHDAuthorizedKeysResource() resource.Resource {
	return &SSHDAuthorizedKeysResource{}
}

// SSHDAuthorizedKeysResource defines the resource implementation.
type SSHDAuthorizedKeysResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *SSHDAuthorizedKeysResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sshd_authorized_keys"
}

// Schema defines the schema for the resource.
func (r *SSHDAuthorizedKeysResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages SSH authorized keys for a user on RTX routers. " +
			"This resource allows you to configure SSH public key authentication for admin users. " +
			"The router returns full public keys, allowing import and drift detection.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: "Username to manage authorized keys for. Must be an existing admin user. Changing this value forces a new resource to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`),
						"must start with a letter and contain only alphanumeric characters and underscores",
					),
				},
			},
			"keys": schema.ListAttribute{
				Description: "List of SSH public keys in OpenSSH format (e.g., 'ssh-ed25519 AAAA... user@host'). Each key must be a valid SSH public key.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"key_count": schema.Int64Attribute{
				Description: "Number of authorized keys registered for this user.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SSHDAuthorizedKeysResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *SSHDAuthorizedKeysResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SSHDAuthorizedKeysModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := data.Username.ValueString()
	ctx = logging.WithResource(ctx, "rtx_sshd_authorized_keys", username)
	logger := logging.FromContext(ctx)

	keys := data.ToKeyStrings()
	logger.Debug().
		Str("resource", "rtx_sshd_authorized_keys").
		Str("username", username).
		Int("keyCount", len(keys)).
		Msg("Creating SSHD authorized keys")

	if err := r.client.SetSSHDAuthorizedKeys(ctx, username, keys); err != nil {
		resp.Diagnostics.AddError(
			"Failed to set SSHD authorized keys",
			fmt.Sprintf("Could not set SSHD authorized keys for user %s: %v", username, err),
		)
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *SSHDAuthorizedKeysResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SSHDAuthorizedKeysModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If username became null, resource was not found
	if data.Username.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the authorized keys from the router.
func (r *SSHDAuthorizedKeysResource) read(ctx context.Context, data *SSHDAuthorizedKeysModel, diagnostics *diag.Diagnostics) {
	username := data.Username.ValueString()

	ctx = logging.WithResource(ctx, "rtx_sshd_authorized_keys", username)
	logger := logging.FromContext(ctx)

	logger.Debug().
		Str("resource", "rtx_sshd_authorized_keys").
		Str("username", username).
		Msg("Reading SSHD authorized keys")

	keys, err := r.client.GetSSHDAuthorizedKeys(ctx, username)
	if err != nil {
		// Check if user or keys don't exist
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no authorized keys") {
			logger.Debug().
				Str("resource", "rtx_sshd_authorized_keys").
				Str("username", username).
				Msg("No authorized keys found, removing from state")
			data.Username = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read SSHD authorized keys",
			fmt.Sprintf("Could not read SSHD authorized keys for user %s: %v", username, err))
		return
	}

	// If no keys returned, remove from state
	if len(keys) == 0 {
		logger.Debug().
			Str("resource", "rtx_sshd_authorized_keys").
			Str("username", username).
			Msg("No authorized keys returned, removing from state")
		data.Username = types.StringNull()
		return
	}

	data.FromClient(keys)

	logger.Debug().
		Str("resource", "rtx_sshd_authorized_keys").
		Str("username", username).
		Int("keyCount", len(keys)).
		Msg("Read SSHD authorized keys successfully")
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SSHDAuthorizedKeysResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SSHDAuthorizedKeysModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := data.Username.ValueString()
	ctx = logging.WithResource(ctx, "rtx_sshd_authorized_keys", username)
	logger := logging.FromContext(ctx)

	keys := data.ToKeyStrings()
	logger.Debug().
		Str("resource", "rtx_sshd_authorized_keys").
		Str("username", username).
		Int("keyCount", len(keys)).
		Msg("Updating SSHD authorized keys")

	// SetSSHDAuthorizedKeys will delete all existing keys and re-register new ones
	if err := r.client.SetSSHDAuthorizedKeys(ctx, username, keys); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update SSHD authorized keys",
			fmt.Sprintf("Could not update SSHD authorized keys for user %s: %v", username, err),
		)
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *SSHDAuthorizedKeysResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SSHDAuthorizedKeysModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := data.Username.ValueString()
	ctx = logging.WithResource(ctx, "rtx_sshd_authorized_keys", username)
	logger := logging.FromContext(ctx)

	logger.Debug().
		Str("resource", "rtx_sshd_authorized_keys").
		Str("username", username).
		Msg("Deleting SSHD authorized keys")

	if err := r.client.DeleteSSHDAuthorizedKeys(ctx, username); err != nil {
		// Check if already gone
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete SSHD authorized keys",
			fmt.Sprintf("Could not delete SSHD authorized keys for user %s: %v", username, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *SSHDAuthorizedKeysResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	username := req.ID

	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("resource", "rtx_sshd_authorized_keys").
		Str("username", username).
		Msg("Importing SSHD authorized keys")

	// Set the username from the import ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("username"), username)...)
}
