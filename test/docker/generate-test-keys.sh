#!/bin/bash
# Generate test SSH host keys for CI/CD

set -e

KEYS_DIR="test_host_keys"
mkdir -p "$KEYS_DIR"

# Generate RSA key for testing
ssh-keygen -t rsa -b 2048 -f "$KEYS_DIR/ssh_host_rsa_key" -N "" -C "test-rtx-router"

# Generate ED25519 key for testing
ssh-keygen -t ed25519 -f "$KEYS_DIR/ssh_host_ed25519_key" -N "" -C "test-rtx-router"

# Create a known_hosts file for testing
echo "[localhost]:2222 $(cat $KEYS_DIR/ssh_host_rsa_key.pub)" > "$KEYS_DIR/test_known_hosts"
echo "[127.0.0.1]:2222 $(cat $KEYS_DIR/ssh_host_rsa_key.pub)" >> "$KEYS_DIR/test_known_hosts"

# Extract the base64 encoded host key for fixed key testing
echo "Test host key (base64):"
ssh-keygen -f "$KEYS_DIR/ssh_host_rsa_key.pub" -e -m RFC4716 | grep -v "^----" | tr -d '\n'
echo

echo "Test keys generated in $KEYS_DIR/"