package service_policy

import (
	"context"
	"fmt"
	"strings"

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
	_ resource.Resource                = &ServicePolicyResource{}
	_ resource.ResourceWithImportState = &ServicePolicyResource{}
)

// NewServicePolicyResource creates a new service policy resource.
func NewServicePolicyResource() resource.Resource {
	return &ServicePolicyResource{}
}

// ServicePolicyResource defines the resource implementation.
type ServicePolicyResource struct {
	client client.Client
}

// Metadata returns the resource type name.
func (r *ServicePolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_policy"
}

// Schema defines the schema for the resource.
func (r *ServicePolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Attaches a QoS policy-map to an interface on RTX routers. Service policies control traffic entering or leaving an interface.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The resource ID in format 'interface:direction'.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"interface": schema.StringAttribute{
				Description: "The interface to attach the policy to (e.g., 'lan1', 'wan1').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"direction": schema.StringAttribute{
				Description: "Direction for the policy: 'input' or 'output'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("input", "output"),
				},
			},
			"policy_map": schema.StringAttribute{
				Description: "The policy-map name or queue type to apply (e.g., 'priority', 'cbq', or a policy-map name).",
				Required:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ServicePolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ServicePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServicePolicyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sp := data.ToClient()
	resourceID := fmt.Sprintf("%s:%s", sp.Interface, sp.Direction)

	ctx = logging.WithResource(ctx, "rtx_service_policy", resourceID)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_service_policy").Msgf("Creating service-policy: %+v", sp)

	if err := r.client.CreateServicePolicy(ctx, sp); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create service-policy",
			fmt.Sprintf("Could not create service-policy: %v", err),
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
func (r *ServicePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServicePolicyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.read(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if resource was not found
	if data.Interface.IsNull() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// read is a helper function that reads the service policy from the router.
func (r *ServicePolicyResource) read(ctx context.Context, data *ServicePolicyModel, diagnostics *diag.Diagnostics) {
	iface, direction, err := parseServicePolicyID(data.ID.ValueString())
	if err != nil {
		fwhelpers.AppendDiagError(diagnostics, "Invalid resource ID", fmt.Sprintf("Could not parse resource ID: %v", err))
		return
	}

	ctx = logging.WithResource(ctx, "rtx_service_policy", data.ID.ValueString())
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_service_policy").Msgf("Reading service-policy: %s:%s", iface, direction)

	sp, err := r.client.GetServicePolicy(ctx, iface, direction)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debug().Str("resource", "rtx_service_policy").Msgf("Service-policy %s:%s not found", iface, direction)
			data.Interface = types.StringNull()
			return
		}
		fwhelpers.AppendDiagError(diagnostics, "Failed to read service-policy", fmt.Sprintf("Could not read service-policy %s:%s: %v", iface, direction, err))
		return
	}

	data.FromClient(sp)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ServicePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ServicePolicyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve ID from state since it's computed
	var state ServicePolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = state.ID

	sp := data.ToClient()

	ctx = logging.WithResource(ctx, "rtx_service_policy", data.ID.ValueString())
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_service_policy").Msgf("Updating service-policy: %+v", sp)

	if err := r.client.UpdateServicePolicy(ctx, sp); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update service-policy",
			fmt.Sprintf("Could not update service-policy: %v", err),
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
func (r *ServicePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServicePolicyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	iface, direction, err := parseServicePolicyID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid resource ID",
			fmt.Sprintf("Could not parse resource ID: %v", err),
		)
		return
	}

	ctx = logging.WithResource(ctx, "rtx_service_policy", data.ID.ValueString())
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_service_policy").Msgf("Deleting service-policy: %s:%s", iface, direction)

	if err := r.client.DeleteServicePolicy(ctx, iface, direction); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete service-policy",
			fmt.Sprintf("Could not delete service-policy %s:%s: %v", iface, direction, err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *ServicePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importID := req.ID

	// Validate import ID format
	iface, direction, err := parseServicePolicyID(importID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Import ID must be in format 'interface:direction' (e.g., 'lan1:output'): %v", err),
		)
		return
	}

	ctx = logging.WithResource(ctx, "rtx_service_policy", importID)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_service_policy").Msgf("Importing service-policy: %s:%s", iface, direction)

	// Set the ID and let Read populate the rest
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), importID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("interface"), iface)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("direction"), direction)...)
}

// parseServicePolicyID parses the resource ID into interface and direction.
func parseServicePolicyID(id string) (string, string, error) {
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
