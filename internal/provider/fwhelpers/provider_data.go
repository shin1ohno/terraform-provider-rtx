package fwhelpers

import (
	"github.com/sh1/terraform-provider-rtx/internal/client"
)

// ProviderData holds the provider-configured data for use by resources and data sources.
type ProviderData struct {
	Client client.Client
}
