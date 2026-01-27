package shape

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	_ resource.Resource                = &ShapeResource{}
	_ resource.ResourceWithImportState = &ShapeResource{}
)

// NewShapeResource creates a new shape resource.
func NewShapeResource() resource.Resource {
	return &ShapeResource{}
}

// ShapeResource defines the resource implementation.
type ShapeResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *ShapeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_shape"
}

// Schema defines the schema for the resource.
func (r *ShapeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages QoS traffic shaping configurations on RTX routers. Shaping limits the rate of outgoing traffic on an interface.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier in the format 'interface:direction'.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"interface": schema.StringAttribute{
				Description: "The interface to apply traffic shaping to (e.g., 'lan1', 'wan1').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"direction": schema.StringAttribute{
				Description: "Direction for shaping: 'input' or 'output'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("input", "output"),
				},
			},
			"shape_average": schema.Int64Attribute{
				Description: "Average rate limit in bits per second (bps).",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"shape_burst": schema.Int64Attribute{
				Description: "Burst size in bytes (optional).",
				Optional:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ShapeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ShapeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ShapeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sc := data.ToClient()
	resourceID := fmt.Sprintf("%s:%s", sc.Interface, sc.Direction)

	ctx = logging.WithResource(ctx, "rtx_shape", resourceID)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_shape").Msgf("Creating shape: %+v", sc)

	if err := r.client.CreateShape(ctx, sc); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create shape",
			fmt.Sprintf("Could not create shape: %v", err),
		)
		return
	}

	data.ID = types.StringValue(resourceID)

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ShapeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ShapeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If resource was not found, remove from state
	if data.ID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the shape from the router.
func (r *ShapeResource) read(ctx context.Context, data *ShapeModel, diagnostics *diag.Diagnostics) {
	iface, direction, err := parseShapeID(data.ID.ValueString())
	if err != nil {
		fwhelpers.AppendDiagError(diagnostics, "Invalid resource ID", fmt.Sprintf("Could not parse resource ID: %v", err))
		return
	}

	ctx = logging.WithResource(ctx, "rtx_shape", data.ID.ValueString())
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_shape").Msgf("Reading shape: %s:%s", iface, direction)

	sc, err := r.client.GetShape(ctx, iface, direction)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_shape").Msgf("Shape %s:%s not found", iface, direction)
			data.ID = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read shape", fmt.Sprintf("Could not read shape %s:%s: %v", iface, direction, err))
		return
	}

	data.FromClient(sc)
	data.ID = types.StringValue(fmt.Sprintf("%s:%s", sc.Interface, sc.Direction))
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ShapeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ShapeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state for ID
	var stateData ShapeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = stateData.ID

	ctx = logging.WithResource(ctx, "rtx_shape", data.ID.ValueString())
	logger := logging.FromContext(ctx)

	sc := data.ToClient()
	logger.Debug().Str("resource", "rtx_shape").Msgf("Updating shape: %+v", sc)

	if err := r.client.UpdateShape(ctx, sc); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update shape",
			fmt.Sprintf("Could not update shape: %v", err),
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
func (r *ShapeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ShapeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	iface, direction, err := parseShapeID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid resource ID",
			fmt.Sprintf("Could not parse resource ID: %v", err),
		)
		return
	}

	ctx = logging.WithResource(ctx, "rtx_shape", data.ID.ValueString())
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_shape").Msgf("Deleting shape: %s:%s", iface, direction)

	if err := r.client.DeleteShape(ctx, iface, direction); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete shape",
			fmt.Sprintf("Could not delete shape %s:%s: %v", iface, direction, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *ShapeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importID := req.ID

	// Validate import ID format
	iface, direction, err := parseShapeID(importID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Invalid import ID format, expected 'interface:direction' (e.g., 'lan1:output'): %v", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), importID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("interface"), iface)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("direction"), direction)...)
}

// parseShapeID parses the resource ID into interface and direction.
func parseShapeID(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected format 'interface:direction', got %q", id)
	}

	iface := parts[0]
	direction := parts[1]

	if iface == "" {
		return "", "", fmt.Errorf("interface cannot be empty")
	}
	if direction != "input" && direction != "output" {
		return "", "", fmt.Errorf("direction must be 'input' or 'output', got %q", direction)
	}

	return iface, direction, nil
}
