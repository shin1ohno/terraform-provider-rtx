# SFTP Configuration Guide

This guide explains how to configure SFTP (SSH File Transfer Protocol) on Yamaha RTX routers and use the Terraform provider with SFTP capabilities.

## Overview

SFTP provides secure file transfer functionality on RTX routers. It allows you to:
- Download configuration files from the router
- Transfer firmware files
- Access log files and other router data

The RTX provider can use SFTP to read configuration files directly from the router, enabling more efficient configuration state management.

## Enabling SFTP on RTX Routers

### Prerequisites

SFTP requires SSH to be enabled on the router. The SFTP daemon runs on top of the SSH service.

### RTX Router Commands

Enable SSH and SFTP on your RTX router using the following commands:

```
# Enable SSH daemon
sshd service on
sshd host lan1

# Enable SFTP daemon
sftpd host lan1
```

**Command options for `sftpd host`:**

| Command | Description |
|---------|-------------|
| `sftpd host <ip_range>` | Allow SFTP access from specific IP addresses |
| `sftpd host any` | Allow SFTP access from any host (not recommended for production) |
| `sftpd host none` | Disable SFTP access |
| `sftpd host lan` | Allow SFTP access from all LAN interfaces |
| `sftpd host lan1` | Allow SFTP access from lan1 interface |
| `no sftpd host` | Remove SFTP host configuration |

### Terraform Configuration

Use the `rtx_sshd` and `rtx_sftpd` resources to manage these settings:

```hcl
# Enable SSH daemon
resource "rtx_sshd" "ssh" {
  enabled = true
  hosts   = ["lan1"]
}

# Enable SFTP daemon
resource "rtx_sftpd" "sftp" {
  hosts = ["lan1"]

  depends_on = [rtx_sshd.ssh]
}
```

## Provider Configuration Options

The RTX provider uses the following settings for SFTP connections:

| Option | Description | Default |
|--------|-------------|---------|
| `host` | RTX router hostname or IP address | Required |
| `username` | SSH username | Required |
| `password` | SSH password | Required |
| `port` | SSH port | 22 |
| `timeout` | Connection timeout in seconds | 30 |
| `ssh_host_key` | Base64-encoded SSH host public key | - |
| `known_hosts_file` | Path to known_hosts file | ~/.ssh/known_hosts |
| `skip_host_key_check` | Skip host key verification (insecure) | false |

### Example Provider Configuration

```hcl
provider "rtx" {
  host     = "192.168.1.1"
  username = "admin"
  password = var.rtx_password
  port     = 22
  timeout  = 30

  # For production: use known_hosts or ssh_host_key
  known_hosts_file = "~/.ssh/known_hosts"

  # For testing only (insecure)
  # skip_host_key_check = true
}
```

## User Permissions

To use SFTP, the user account must have appropriate permissions configured:

```
# Configure user with SFTP access
login user <username> <password> connection=sftp
```

The `connection` attribute controls which connection methods are allowed:
- `all` - Allow all connection types
- `ssh` - Allow SSH connections
- `sftp` - Allow SFTP connections
- Multiple values can be combined with commas

Users with `administrator` attribute enabled can access SFTP using the administrator password.

## Troubleshooting

### Connection Refused

**Symptoms:**
- "connection refused" error when attempting SFTP connection

**Solutions:**
1. Verify SSH is enabled: `show status sshd`
2. Verify SFTP is enabled: `show config | grep sftpd`
3. Check host restrictions: `show config | grep "sftpd host"`
4. Ensure the connecting IP is in the allowed range

### Authentication Failed

**Symptoms:**
- "authentication failed" or "permission denied" errors

**Solutions:**
1. Verify username and password are correct
2. Check user has SFTP connection permission: `show config | grep "login user"`
3. Ensure user's `connection` attribute includes `sftp` or `all`

### Host Key Verification Failed

**Symptoms:**
- "host key verification failed" error

**Solutions:**
1. Add the router's host key to known_hosts:
   ```bash
   ssh-keyscan -p 22 192.168.1.1 >> ~/.ssh/known_hosts
   ```
2. Or provide the host key directly in provider configuration:
   ```hcl
   provider "rtx" {
     ssh_host_key = "AAAAB3NzaC1yc2E..."
   }
   ```
3. For testing only, use `skip_host_key_check = true` (not recommended for production)

### File Not Found

**Symptoms:**
- "no such file or directory" when downloading files

**Solutions:**
1. Verify the file path exists on the router
2. Check file permissions
3. Use `show file list` command on the router to list available files

### Timeout Errors

**Symptoms:**
- Connection timeouts during SFTP operations

**Solutions:**
1. Increase the timeout value in provider configuration
2. Check network connectivity to the router
3. Verify the router is not under heavy load

## Security Considerations

### Host Key Verification

Always verify the SSH host key in production environments:

- **Best:** Use `ssh_host_key` with the router's public key
- **Good:** Use `known_hosts_file` with a managed known_hosts file
- **Avoid:** Using `skip_host_key_check = true` in production

### Interface Restrictions

Restrict SFTP access to trusted interfaces only:

```hcl
resource "rtx_sftpd" "sftp" {
  # Only allow SFTP from internal LAN
  hosts = ["lan1"]
}
```

Avoid using `any` for the host setting in production environments.

### User Access Control

- Create dedicated users for Terraform with minimal required permissions
- Use strong passwords
- Consider using SSH key authentication when supported
- Regularly audit user access and permissions

### Network Segmentation

- Place management interfaces (SSH/SFTP) on a dedicated management VLAN
- Use firewall rules to restrict access to management services
- Consider using a bastion host for administrative access

## Supported Models

SFTP is supported on the following RTX router models:
- vRX Series (Amazon EC2, VMware ESXi)
- RTX5000
- RTX3510
- RTX3500
- RTX1300
- RTX1220
- RTX1210
- RTX840
- RTX830

## Related Resources

- [rtx_sshd Resource](resources/sshd.md) - SSH daemon configuration
- [rtx_sftpd Resource](resources/sftpd.md) - SFTP daemon configuration
- [Provider Configuration](index.md) - Main provider documentation
