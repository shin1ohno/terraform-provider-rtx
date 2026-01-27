# Design Document: Plugin Framework Migration

## Overview

This document describes the technical design for migrating terraform-provider-rtx from Terraform Plugin SDK v2 to Terraform Plugin Framework. The migration involves rewriting ~45,000 lines of provider code across 97 resource files while maintaining functional equivalence.

## Current Architecture (SDK v2)

```
internal/
├── provider/
│   ├── provider.go              # SDK v2 provider definition
│   ├── resource_rtx_*.go        # 97 resource implementations
│   ├── data_source_rtx_*.go     # Data sources
│   ├── schema_helpers.go        # WriteOnlyStringSchema, etc.
│   ├── state_funcs.go           # State manipulation helpers
│   └── acctest/                 # Acceptance test helpers
├── client/
│   ├── client.go                # SSH client (keep as-is)
│   └── *_service.go             # CRUD services (keep as-is)
├── rtx/
│   └── parsers/                 # RTX output parsers (keep as-is)
└── logging/                     # Zerolog logging (keep as-is)
```

## Target Architecture (Plugin Framework)

```
internal/
├── provider/
│   ├── provider.go              # Framework provider definition
│   ├── resources/
│   │   ├── ipsec_tunnel/
│   │   │   ├── resource.go      # Resource implementation
│   │   │   ├── model.go         # Typed Go struct for state
│   │   │   └── schema.go        # Schema definition
│   │   ├── l2tp/
│   │   │   ├── resource.go
│   │   │   ├── model.go
│   │   │   └── schema.go
│   │   └── ...                  # Other resources
│   ├── datasources/
│   │   └── ...                  # Data sources
│   ├── planmodifiers/           # Custom plan modifiers
│   ├── validators/              # Custom validators
│   └── fwhelpers/               # Framework helpers
├── client/                      # Keep as-is
├── rtx/                         # Keep as-is
└── logging/                     # Keep as-is
```

## Code Reuse Analysis

### Components to Keep Unchanged
- **internal/client/**: SSH client and all `*_service.go` files - business logic unchanged
- **internal/rtx/parsers/**: RTX output parsers - no changes needed
- **internal/logging/**: Zerolog infrastructure - works with Framework

### Components to Migrate
- **provider.go**: Complete rewrite to Framework pattern
- **resource_rtx_*.go**: Rewrite each resource to Framework pattern
- **schema_helpers.go**: Replace with Framework equivalents
- **state_funcs.go**: Replace with typed model operations

## Architecture

### Provider Structure

```go
// internal/provider/provider.go
package provider

import (
    "github.com/hashicorp/terraform-plugin-framework/provider"
    "github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

type rtxProvider struct {
    version string
}

func New(version string) func() provider.Provider {
    return func() provider.Provider {
        return &rtxProvider{version: version}
    }
}

func (p *rtxProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "host": schema.StringAttribute{
                Required:    true,
                Description: "RTX router hostname or IP",
            },
            "password": schema.StringAttribute{
                Optional:    true,
                Sensitive:   true,
                WriteOnly:   true,  // NEW: Not stored in state
                Description: "SSH password",
            },
            // ...
        },
    }
}
```

### Resource Pattern

```go
// internal/provider/resources/ipsec_tunnel/resource.go
package ipsec_tunnel

import (
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

type Resource struct {
    client *client.Client
}

func NewResource() resource.Resource {
    return &Resource{}
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_ipsec_tunnel"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed: true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "tunnel_id": schema.Int64Attribute{
                Required: true,
            },
            "pre_shared_key": schema.StringAttribute{
                Optional:    true,
                Sensitive:   true,
                WriteOnly:   true,  // KEY: Not stored in state
                Description: "IKE pre-shared key (write-only)",
            },
            // ...
        },
    }
}
```

### Typed Model Pattern

```go
// internal/provider/resources/ipsec_tunnel/model.go
package ipsec_tunnel

import "github.com/hashicorp/terraform-plugin-framework/types"

// Model represents the Terraform state for rtx_ipsec_tunnel
type Model struct {
    ID            types.String `tfsdk:"id"`
    TunnelID      types.Int64  `tfsdk:"tunnel_id"`
    Enabled       types.Bool   `tfsdk:"enabled"`
    PreSharedKey  types.String `tfsdk:"pre_shared_key"`  // WriteOnly, not in state
    LocalAddress  types.String `tfsdk:"local_address"`
    RemoteAddress types.String `tfsdk:"remote_address"`
    DPDEnabled    types.Bool   `tfsdk:"dpd_enabled"`
    DPDInterval   types.Int64  `tfsdk:"dpd_interval"`
    DPDRetry      types.Int64  `tfsdk:"dpd_retry"`
    // Nested blocks
    IKEv2Proposal  *IKEv2ProposalModel  `tfsdk:"ikev2_proposal"`
    IPsecTransform *IPsecTransformModel `tfsdk:"ipsec_transform"`
}

type IKEv2ProposalModel struct {
    EncryptionAES256 types.Bool  `tfsdk:"encryption_aes256"`
    GroupFourteen    types.Bool  `tfsdk:"group_fourteen"`
    IntegritySHA256  types.Bool  `tfsdk:"integrity_sha256"`
    LifetimeSeconds  types.Int64 `tfsdk:"lifetime_seconds"`
}
```

### CRUD Implementation

```go
// internal/provider/resources/ipsec_tunnel/resource.go (continued)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan Model
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Convert to domain model and call service
    tunnel := r.modelToIPsecTunnel(plan)
    if err := r.client.IPsecTunnel().Create(ctx, tunnel); err != nil {
        resp.Diagnostics.AddError("Create failed", err.Error())
        return
    }

    // Set state (pre_shared_key is WriteOnly, not saved)
    plan.ID = types.StringValue(fmt.Sprintf("%d", plan.TunnelID.ValueInt64()))
    diags = resp.State.Set(ctx, &plan)
    resp.Diagnostics.Append(diags...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state Model
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Read from router
    tunnel, err := r.client.IPsecTunnel().Read(ctx, int(state.TunnelID.ValueInt64()))
    if err != nil {
        resp.State.RemoveResource(ctx)
        return
    }

    // Update state from router (pre_shared_key not read back)
    r.ipsecTunnelToModel(tunnel, &state)
    diags = resp.State.Set(ctx, &state)
    resp.Diagnostics.Append(diags...)
}
```

## Components and Interfaces

### Provider Component
- **Purpose**: Initialize provider, configure client, register resources
- **Interfaces**: `provider.Provider`, `provider.ProviderWithMetaSchema`
- **Dependencies**: internal/client

### Resource Components (per resource)
- **Purpose**: Implement CRUD for single resource type
- **Interfaces**: `resource.Resource`, `resource.ResourceWithImportState`, `resource.ResourceWithConfigure`
- **Dependencies**: internal/client/*_service.go

### Plan Modifiers
- **Purpose**: Custom plan modification logic
- **Location**: internal/provider/planmodifiers/
- **Examples**: `UseStateForUnknown()`, `RequiresReplace()`

### Validators
- **Purpose**: Custom validation logic
- **Location**: internal/provider/validators/
- **Examples**: IP address validation, CIDR validation

## Data Models

### Write-Only Attributes

Resources with write-only sensitive attributes:

| Resource | Attribute | Type |
|----------|-----------|------|
| rtx_ipsec_tunnel | pre_shared_key | string |
| rtx_l2tp | tunnel_auth_password | string |
| rtx_l2tp | ipsec_profile.pre_shared_key | string |
| rtx_admin | admin_password | string |
| rtx_admin | login_password | string |
| rtx_admin_user | password | string |
| rtx_ddns | password | string |
| provider | password | string |
| provider | private_key | string |
| provider | private_key_passphrase | string |
| provider | admin_password | string |

### Migration Mapping

SDK v2 → Framework type mapping:

| SDK v2 | Framework |
|--------|-----------|
| schema.TypeString | types.String |
| schema.TypeInt | types.Int64 |
| schema.TypeBool | types.Bool |
| schema.TypeFloat | types.Float64 |
| schema.TypeList | types.List |
| schema.TypeSet | types.Set |
| schema.TypeMap | types.Map |
| Nested *schema.Resource | Nested attribute/block |

## Error Handling

### Error Scenarios

1. **Terraform Version < 1.11**
   - **Handling**: Check version in provider Configure, return error diagnostic
   - **User Impact**: Clear message: "This provider requires Terraform 1.11 or later for write-only attribute support"

2. **SSH Connection Failure**
   - **Handling**: Return error diagnostic with connection details
   - **User Impact**: Same as current behavior

3. **Invalid Configuration**
   - **Handling**: Use Framework validators
   - **User Impact**: Validation errors shown during plan

## Testing Strategy

### Unit Testing
- Test typed model conversions
- Test validators in isolation
- Test plan modifiers

### Integration Testing
- Migrate existing acceptance tests to Framework testing patterns
- Use `resource.Test()` from `terraform-plugin-testing`

```go
func TestAccIPsecTunnel_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccIPsecTunnelConfig_basic(),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("rtx_ipsec_tunnel.test", "tunnel_id", "1"),
                ),
            },
        },
    })
}
```

### Write-Only Testing
- Verify sensitive values not in state after apply
- Use `terraform show -json` to confirm absence

## Migration Phases

### Phase 1: Infrastructure Setup
1. Add Framework dependencies to go.mod
2. Create new directory structure
3. Implement provider skeleton
4. Setup acceptance test infrastructure

### Phase 2: High-Priority Resources (Sensitive)
1. rtx_ipsec_tunnel
2. rtx_l2tp
3. rtx_admin
4. rtx_admin_user
5. rtx_ddns

### Phase 3: Normal-Priority Resources
1. Migrate remaining resources alphabetically
2. Each resource: schema → model → CRUD → tests

### Phase 4: Cleanup
1. Remove SDK v2 dependencies
2. Remove old resource files
3. Update documentation
4. Release as v1.0.0

## File Changes Summary

| Action | Files | Estimated LOC |
|--------|-------|---------------|
| Delete | internal/provider/resource_rtx_*.go (old) | -40,000 |
| Delete | internal/provider/schema_helpers.go | -200 |
| Delete | internal/provider/state_funcs.go | -200 |
| Create | internal/provider/provider.go (new) | +300 |
| Create | internal/provider/resources/*/resource.go | +25,000 |
| Create | internal/provider/resources/*/model.go | +5,000 |
| Create | internal/provider/resources/*/schema.go | +10,000 |
| Create | internal/provider/validators/*.go | +500 |
| Create | internal/provider/planmodifiers/*.go | +300 |
| Modify | go.mod | Framework deps |
| Keep | internal/client/* | 0 |
| Keep | internal/rtx/* | 0 |
| Keep | internal/logging/* | 0 |

**Net change**: ~-5,000 LOC (cleaner typed code)
