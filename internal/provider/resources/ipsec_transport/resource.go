package ipsec_transport

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &IPsecTransportResource{}
	_ resource.ResourceWithImportState = &IPsecTransportResource{}
)

// NewIPsecTransportResource creates a new IPsec transport resource.
func NewIPsecTransportResource() resource.Resource {
	return &IPsecTransportResource{}
}

// IPsecTransportResource defines the resource implementation.
type IPsecTransportResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *IPsecTransportResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipsec_transport"
}

// Schema defines the schema for the resource.
func (r *IPsecTransportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages IPsec transport mode configuration on RTX routers. Used for L2TP over IPsec and other transport mode VPN configurations.",
		Attributes: map[string]schema.Attribute{
			"transport_id": schema.Int64Attribute{
				Description: "Transport ID (1-6000).",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 6000),
				},
			},
			"tunnel_id": schema.Int64Attribute{
				Description: "Associated IPsec tunnel ID (1-6000).",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 6000),
				},
			},
			"protocol": schema.StringAttribute{
				Description: "Transport protocol: 'udp' or 'tcp'.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive("udp", "tcp"),
				},
			},
			"port": schema.Int64Attribute{
				Description: "Port number (1-65535). Common value is 1701 for L2TP.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *IPsecTransportResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *IPsecTransportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IPsecTransportModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add resource context for logging
	ctx = logging.WithResource(ctx, "rtx_ipsec_transport", strconv.FormatInt(data.TransportID.ValueInt64(), 10))
	logger := logging.FromContext(ctx)

	transport := data.ToClient()
	logger.Debug().Str("resource", "rtx_ipsec_transport").Msgf("Creating IPsec transport: %+v", transport)

	if err := r.client.CreateIPsecTransport(ctx, transport); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create IPsec transport",
			fmt.Sprintf("Could not create IPsec transport: %v", err),
		)
		return
	}

	// Read back the created resource
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *IPsecTransportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IPsecTransportModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if resource was not found
	if data.TransportID.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the transport from the router.
func (r *IPsecTransportResource) read(ctx context.Context, data *IPsecTransportModel, diagnostics *diag.Diagnostics) {
	transportID := int(data.TransportID.ValueInt64())

	ctx = logging.WithResource(ctx, "rtx_ipsec_transport", strconv.Itoa(transportID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_ipsec_transport").Msgf("Reading IPsec transport: %d", transportID)

	transport, err := r.client.GetIPsecTransport(ctx, transportID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_ipsec_transport").Msgf("IPsec transport %d not found", transportID)
			// Resource has been deleted outside of Terraform
			data.TransportID = types.Int64Null()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read IPsec transport", fmt.Sprintf("Could not read IPsec transport %d: %v", transportID, err))
		return
	}

	// Update data from the transport
	data.FromClient(transport)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *IPsecTransportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IPsecTransportModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = logging.WithResource(ctx, "rtx_ipsec_transport", strconv.FormatInt(data.TransportID.ValueInt64(), 10))
	logger := logging.FromContext(ctx)

	transport := data.ToClient()
	logger.Debug().Str("resource", "rtx_ipsec_transport").Msgf("Updating IPsec transport: %+v", transport)

	if err := r.client.UpdateIPsecTransport(ctx, transport); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update IPsec transport",
			fmt.Sprintf("Could not update IPsec transport: %v", err),
		)
		return
	}

	// Read back the updated resource
	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *IPsecTransportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IPsecTransportModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	transportID := int(data.TransportID.ValueInt64())

	ctx = logging.WithResource(ctx, "rtx_ipsec_transport", strconv.Itoa(transportID))
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_ipsec_transport").Msgf("Deleting IPsec transport: %d", transportID)

	if err := r.client.DeleteIPsecTransport(ctx, transportID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete IPsec transport",
			fmt.Sprintf("Could not delete IPsec transport %d: %v", transportID, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *IPsecTransportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	transportID, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected transport ID as integer, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("transport_id"), int64(transportID))...)
}
