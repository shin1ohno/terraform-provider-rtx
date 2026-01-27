package class_map

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	_ resource.Resource                = &ClassMapResource{}
	_ resource.ResourceWithImportState = &ClassMapResource{}
)

// NewClassMapResource creates a new class map resource.
func NewClassMapResource() resource.Resource {
	return &ClassMapResource{}
}

// ClassMapResource defines the resource implementation.
type ClassMapResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *ClassMapResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_class_map"
}

// Schema defines the schema for the resource.
func (r *ClassMapResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages QoS class-map configurations on RTX routers. Class-maps classify traffic based on various match criteria.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The class-map name. Must start with a letter and contain only alphanumeric characters, underscores, and hyphens.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`),
						"must start with a letter and contain only alphanumeric characters, underscores, and hyphens",
					),
				},
			},
			"match_protocol": schema.StringAttribute{
				Description: "Protocol to match (e.g., 'sip', 'http', 'ftp').",
				Optional:    true,
			},
			"match_destination_port": schema.ListAttribute{
				Description: "List of destination ports to match.",
				Optional:    true,
				ElementType: types.Int64Type,
				Validators: []validator.List{
					listvalidator.ValueInt64sAre(
						int64validator.Between(1, 65535),
					),
				},
			},
			"match_source_port": schema.ListAttribute{
				Description: "List of source ports to match.",
				Optional:    true,
				ElementType: types.Int64Type,
				Validators: []validator.List{
					listvalidator.ValueInt64sAre(
						int64validator.Between(1, 65535),
					),
				},
			},
			"match_dscp": schema.StringAttribute{
				Description: "DSCP value to match (e.g., 'ef', 'af11', '46').",
				Optional:    true,
			},
			"match_filter": schema.Int64Attribute{
				Description: "IP filter number to reference for matching (1-65535).",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ClassMapResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ClassMapResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ClassMapModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_class_map", data.Name.ValueString())
	logger := logging.FromContext(ctx)

	cm := data.ToClient()
	logger.Debug().Str("resource", "rtx_class_map").Msgf("Creating class-map: %s", cm.Name)

	if err := r.client.CreateClassMap(ctx, cm); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create class-map",
			fmt.Sprintf("Could not create class-map: %v", err),
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
func (r *ClassMapResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ClassMapModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if resource was removed
	if data.Name.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the class-map from the router.
func (r *ClassMapResource) read(ctx context.Context, data *ClassMapModel, diagnostics *diag.Diagnostics) {
	name := data.Name.ValueString()

	ctx = logging.WithResource(ctx, "rtx_class_map", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_class_map").Msgf("Reading class-map: %s", name)

	cm, err := r.client.GetClassMap(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_class_map").Msgf("Class-map %s not found", name)
			data.Name = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read class-map", fmt.Sprintf("Could not read class-map %s: %v", name, err))
		return
	}

	data.FromClient(cm)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ClassMapResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ClassMapModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_class_map", data.Name.ValueString())
	logger := logging.FromContext(ctx)

	cm := data.ToClient()
	logger.Debug().Str("resource", "rtx_class_map").Msgf("Updating class-map: %s", cm.Name)

	if err := r.client.UpdateClassMap(ctx, cm); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update class-map",
			fmt.Sprintf("Could not update class-map: %v", err),
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
func (r *ClassMapResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ClassMapModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	ctx = logging.WithResource(ctx, "rtx_class_map", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_class_map").Msgf("Deleting class-map: %s", name)

	if err := r.client.DeleteClassMap(ctx, name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete class-map",
			fmt.Sprintf("Could not delete class-map %s: %v", name, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *ClassMapResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
