package main

import (
	"fmt"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

func main() {
	// Test the static route parser with actual router config
	routeConfig := `ip route 10.33.128.0/21 gateway 192.168.1.20 gateway 192.168.1.21
ip route 100.64.0.0/10 gateway 192.168.1.20 gateway 192.168.1.21`

	routes, err := parsers.ParseStaticRoutes([]byte(routeConfig))
	if err != nil {
		fmt.Printf("Error parsing routes: %v\n", err)
		return
	}

	fmt.Printf("Parsed %d routes:\n", len(routes))
	for i, route := range routes {
		fmt.Printf("Route %d:\n", i+1)
		fmt.Printf("  Destination: %s\n", route.Destination)
		fmt.Printf("  GatewayIP: %s\n", route.GatewayIP)
		fmt.Printf("  GatewayInterface: %s\n", route.GatewayInterface)
		fmt.Printf("  Interface: %s\n", route.Interface)
		fmt.Printf("  Metric: %d\n", route.Metric)
		fmt.Printf("  Weight: %d\n", route.Weight)
		fmt.Printf("  Description: %s\n", route.Description)
		fmt.Printf("  Hide: %t\n\n", route.Hide)
	}
}