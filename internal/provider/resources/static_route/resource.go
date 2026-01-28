package static_route

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	_ resource.Resource                = &StaticRouteResource{}
	_ resource.ResourceWithImportState = &StaticRouteResource{}
)

// NewStaticRouteResource creates a new static route resource.
func NewStaticRouteResource() resource.Resource {
	return &StaticRouteResource{}
}

// StaticRouteResource defines the resource implementation.
type StaticRouteResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *StaticRouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_route"
}

// Schema defines the schema for the resource.
func (r *StaticRouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages static routes on RTX routers. A static route defines a fixed path for network traffic to reach a specific destination network.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier in the format 'prefix/mask'.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"prefix": schema.StringAttribute{
				Description: "The destination network prefix (e.g., '10.0.0.0' for a network, '0.0.0.0' for default route).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					ipv4AddressValidator{},
				},
			},
			"mask": schema.StringAttribute{
				Description: "The subnet mask in dotted decimal notation (e.g., '255.255.255.0').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					subnetMaskValidator{},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"next_hop": schema.ListNestedBlock{
				Description: "Next hop configuration for this route. Multiple next hops enable load balancing or failover.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"gateway": schema.StringAttribute{
							Description: "Next hop gateway IP address (e.g., '192.168.1.1'). Either gateway or interface must be specified.",
							Optional:    true,
							Validators: []validator.String{
								optionalIPv4AddressValidator{},
							},
						},
						"interface": schema.StringAttribute{
							Description: "Outgoing interface (e.g., 'pp 1', 'tunnel 1'). Either gateway or interface must be specified.",
							Optional:    true,
						},
						"distance": schema.Int64Attribute{
							Description: "Administrative distance (weight). Lower values are preferred. Range: 1-100. Defaults to router default if not specified.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 100),
							},
						},
						"permanent": schema.BoolAttribute{
							Description: "Keep route even when next hop is unreachable (keepalive). Defaults to router default if not specified.",
							Optional:    true,
							Computed:    true,
						},
						"filter": schema.Int64Attribute{
							Description: "IP filter number to apply to this route. 0 means no filter. Defaults to router default if not specified.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(0),
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *StaticRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*fwhelpers.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *fwhelpers.ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = providerData.Client
}

// Create creates the resource and sets the initial Terraform state.
func (r *StaticRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StaticRouteModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	route := data.ToClient()

	// Add resource context for logging
	id := fmt.Sprintf("%s/%s", route.Prefix, route.Mask)
	ctx = logging.WithResource(ctx, "rtx_static_route", id)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_static_route").Msgf("Creating static route: %+v", route)

	if err := r.client.CreateStaticRoute(ctx, route); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create static route",
			fmt.Sprintf("Could not create static route: %v", err),
		)
		return
	}

	// Set the ID
	data.ID = types.StringValue(id)

	// Invalidate cache to ensure we read the latest config
	r.client.InvalidateCache()

	// Read back the created resource
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *StaticRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StaticRouteModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if resource was deleted outside of Terraform
	if data.Prefix.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the static route from the router.
func (r *StaticRouteResource) read(ctx context.Context, data *StaticRouteModel, diagnostics *diag.Diagnostics) {
	prefix := fwhelpers.GetStringValue(data.Prefix)
	mask := fwhelpers.GetStringValue(data.Mask)

	ctx = logging.WithResource(ctx, "rtx_static_route", data.ID.ValueString())
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_static_route").Msgf("Reading static route: %s/%s", prefix, mask)

	// Always use SSH for static routes to ensure we get the latest config
	// SFTP cache may not be updated immediately after Create/Update
	route, err := r.client.GetStaticRoute(ctx, prefix, mask)
	if err != nil {
		// Check if route doesn't exist
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_static_route").Msgf("Static route %s/%s not found", prefix, mask)
			// Resource has been deleted outside of Terraform
			data.Prefix = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read static route", fmt.Sprintf("Could not read static route %s/%s: %v", prefix, mask, err))
		return
	}

	logger.Debug().Str("resource", "rtx_static_route").Msgf("Read route with %d next hops", len(route.NextHops))

	// Update data from the route
	data.FromClient(route)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *StaticRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StaticRouteModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	route := data.ToClient()

	ctx = logging.WithResource(ctx, "rtx_static_route", data.ID.ValueString())
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_static_route").Msgf("Updating static route: %+v", route)

	if err := r.client.UpdateStaticRoute(ctx, route); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update static route",
			fmt.Sprintf("Could not update static route: %v", err),
		)
		return
	}

	// Invalidate cache to ensure we read the latest config
	r.client.InvalidateCache()

	// Read back the updated resource
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *StaticRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StaticRouteModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	prefix := fwhelpers.GetStringValue(data.Prefix)
	mask := fwhelpers.GetStringValue(data.Mask)

	ctx = logging.WithResource(ctx, "rtx_static_route", data.ID.ValueString())
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_static_route").Msgf("Deleting static route: %s/%s", prefix, mask)

	if err := r.client.DeleteStaticRoute(ctx, prefix, mask); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete static route",
			fmt.Sprintf("Could not delete static route %s/%s: %v", prefix, mask, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *StaticRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse import ID as "prefix/mask" (e.g., "10.0.0.0/255.0.0.0" or "0.0.0.0/0.0.0.0")
	prefix, mask, err := parseStaticRouteID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format 'prefix/mask': %v", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("prefix"), prefix)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("mask"), mask)...)
}

// parseStaticRouteID parses the resource ID into prefix and mask
func parseStaticRouteID(id string) (prefix, mask string, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected format 'prefix/mask', got %q", id)
	}

	prefix = parts[0]
	mask = parts[1]

	// Validate both parts are valid IPs
	if net.ParseIP(prefix) == nil {
		return "", "", fmt.Errorf("invalid prefix IP address: %s", prefix)
	}
	if net.ParseIP(mask) == nil {
		return "", "", fmt.Errorf("invalid mask: %s", mask)
	}

	return prefix, mask, nil
}

// Custom validators

// ipv4AddressValidator validates an IPv4 address
type ipv4AddressValidator struct{}

func (v ipv4AddressValidator) Description(ctx context.Context) string {
	return "must be a valid IPv4 address"
}

func (v ipv4AddressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ipv4AddressValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv4 Address",
			"Value cannot be empty",
		)
		return
	}

	ip := net.ParseIP(value)
	if ip == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv4 Address",
			fmt.Sprintf("Value %q is not a valid IP address", value),
		)
		return
	}

	if ip.To4() == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv4 Address",
			fmt.Sprintf("Value %q is not an IPv4 address", value),
		)
	}
}

// optionalIPv4AddressValidator validates an optional IPv4 address
type optionalIPv4AddressValidator struct{}

func (v optionalIPv4AddressValidator) Description(ctx context.Context) string {
	return "must be a valid IPv4 address if provided"
}

func (v optionalIPv4AddressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v optionalIPv4AddressValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		return // Empty is OK for optional field
	}

	ip := net.ParseIP(value)
	if ip == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv4 Address",
			fmt.Sprintf("Value %q is not a valid IP address", value),
		)
		return
	}

	if ip.To4() == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IPv4 Address",
			fmt.Sprintf("Value %q is not an IPv4 address", value),
		)
	}
}

// subnetMaskValidator validates a subnet mask in dotted decimal notation
type subnetMaskValidator struct{}

func (v subnetMaskValidator) Description(ctx context.Context) string {
	return "must be a valid subnet mask in dotted decimal notation (e.g., '255.255.255.0')"
}

func (v subnetMaskValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v subnetMaskValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Subnet Mask",
			"Value cannot be empty",
		)
		return
	}

	ip := net.ParseIP(value)
	if ip == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Subnet Mask",
			fmt.Sprintf("Value %q is not a valid subnet mask in dotted decimal notation (e.g., '255.255.255.0')", value),
		)
		return
	}

	// Ensure it's IPv4
	if ip.To4() == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Subnet Mask",
			fmt.Sprintf("Value %q is not an IPv4 subnet mask", value),
		)
		return
	}

	// Validate that it's a valid mask (contiguous 1s followed by 0s)
	maskBytes := ip.To4()
	mask := net.IPv4Mask(maskBytes[0], maskBytes[1], maskBytes[2], maskBytes[3])
	ones, bits := mask.Size()
	if bits == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Subnet Mask",
			fmt.Sprintf("Value %q is not a valid subnet mask", value),
		)
		return
	}

	// Verify the mask is contiguous
	maskInt := uint32(maskBytes[0])<<24 | uint32(maskBytes[1])<<16 | uint32(maskBytes[2])<<8 | uint32(maskBytes[3])
	expectedMask := uint32(0xFFFFFFFF) << (32 - ones)
	if maskInt != expectedMask {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Subnet Mask",
			fmt.Sprintf("Value %q is not a valid contiguous subnet mask", value),
		)
	}
}
