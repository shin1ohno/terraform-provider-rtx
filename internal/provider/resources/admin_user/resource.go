package admin_user

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
	_ resource.Resource                = &AdminUserResource{}
	_ resource.ResourceWithImportState = &AdminUserResource{}
)

// NewAdminUserResource creates a new admin user resource.
func NewAdminUserResource() resource.Resource {
	return &AdminUserResource{}
}

// AdminUserResource defines the resource implementation.
type AdminUserResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *AdminUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_user"
}

// Schema defines the schema for the resource.
func (r *AdminUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages admin user accounts on RTX routers. Each user can have different permissions and access methods.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: "Username for the admin user (cannot be changed after creation). Must start with a letter and contain only alphanumeric characters and underscores.",
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
			"password": schema.StringAttribute{
				Description: "Password for the admin user. This value is write-only and will not be stored in state.",
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
			},
			"encrypted": schema.BoolAttribute{
				Description: "Whether the password is already encrypted. If true, the password value will be used as-is.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"administrator": schema.BoolAttribute{
				Description: "Whether the user has administrator privileges.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"connection_methods": schema.SetAttribute{
				Description: "Allowed connection methods for the user.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf("serial", "telnet", "remote", "ssh", "sftp", "http"),
					),
				},
			},
			"gui_pages": schema.SetAttribute{
				Description: "Allowed GUI pages for the user.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf("dashboard", "lan-map", "config"),
					),
				},
			},
			"login_timer": schema.Int64Attribute{
				Description: "Login timeout in seconds. 0 means infinite (no timeout).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *AdminUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *AdminUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AdminUserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_admin_user", data.Username.ValueString())
	logger := logging.FromContext(ctx)

	user := data.ToClient()
	logger.Debug().Str("resource", "rtx_admin_user").Msgf("Creating admin user: %s", user.Username)

	if err := r.client.CreateAdminUser(ctx, user); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create admin user",
			fmt.Sprintf("Could not create admin user: %v", err),
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
func (r *AdminUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AdminUserModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the user from the router.
func (r *AdminUserResource) read(ctx context.Context, data *AdminUserModel, diagnostics *diag.Diagnostics) {
	username := data.Username.ValueString()

	ctx = logging.WithResource(ctx, "rtx_admin_user", username)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_admin_user").Msgf("Reading admin user: %s", username)

	user, err := r.client.GetAdminUser(ctx, username)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_admin_user").Msgf("Admin user %s not found", username)
			data.Username = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read admin user", fmt.Sprintf("Could not read admin user %s: %v", username, err))
		return
	}

	data.FromClient(user)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AdminUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AdminUserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_admin_user", data.Username.ValueString())
	logger := logging.FromContext(ctx)

	user := data.ToClient()
	logger.Debug().Str("resource", "rtx_admin_user").Msgf("Updating admin user: %s", user.Username)

	if err := r.client.UpdateAdminUser(ctx, user); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update admin user",
			fmt.Sprintf("Could not update admin user: %v", err),
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
func (r *AdminUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AdminUserModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := data.Username.ValueString()

	ctx = logging.WithResource(ctx, "rtx_admin_user", username)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_admin_user").Msgf("Deleting admin user: %s", username)

	if err := r.client.DeleteAdminUser(ctx, username); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete admin user",
			fmt.Sprintf("Could not delete admin user %s: %v", username, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *AdminUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("username"), req, resp)
}
