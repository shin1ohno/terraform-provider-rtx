#!/bin/bash

# Terraform import script for RTX static routes
# Imports actual existing routes from router configuration

echo "Starting Terraform import for RTX static routes..."

# First, check current routes
echo "Checking current static routes..."
terraform refresh

# Show current routes from router
echo "Current routes from data source:"
terraform output current_static_routes

echo ""
echo "Note: Import requires exact format: 'destination||gateway||interface'"
echo ""
echo "Based on actual router config:"
echo "  ip route default gateway dhcp lan2"
echo "  ip route 10.33.128.0/21 gateway 192.168.1.20"
echo "  ip route 100.64.0.0/10 gateway 192.168.1.20"
echo "  ip route 192.168.10.0/24 gateway 192.168.1.10 metric 10"
echo ""

echo "Attempting to import existing static routes..."

# Import default route
echo "Importing default route..."
terraform import rtx_static_route.default_route "0.0.0.0/0||dhcp||lan2"

if [ $? -eq 0 ]; then
    echo "✓ Successfully imported default route"
else
    echo "✗ Failed to import default route"
fi

echo ""

# Import route to 10.33.128.0/21 network
echo "Importing route to 10.33.128.0/21..."
terraform import rtx_static_route.private_network_route "10.33.128.0/21||192.168.1.20||"

if [ $? -eq 0 ]; then
    echo "✓ Successfully imported 10.33.128.0/21 route"
else
    echo "✗ Failed to import 10.33.128.0/21 route"
fi

echo ""

# Import route to 100.64.0.0/10 network  
echo "Importing route to 100.64.0.0/10..."
terraform import rtx_static_route.cgn_network_route "100.64.0.0/10||192.168.1.20||"

if [ $? -eq 0 ]; then
    echo "✓ Successfully imported 100.64.0.0/10 route"
else
    echo "✗ Failed to import 100.64.0.0/10 route"
fi

echo ""

# Import route to 192.168.10.0/24 network
echo "Importing route to 192.168.10.0/24..."
terraform import rtx_static_route.lan_network_route "192.168.10.0/24||192.168.1.10||"

if [ $? -eq 0 ]; then
    echo "✓ Successfully imported 192.168.10.0/24 route"
else
    echo "✗ Failed to import 192.168.10.0/24 route"
fi

echo ""
echo "Import process completed."
echo "Running terraform plan to verify imported resources..."
terraform plan