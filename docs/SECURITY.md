# Security Considerations

This document outlines the security features and best practices for the Terraform Provider for RTX routers.

## Password Security

### Current Implementation

The provider implements the following security measures for password handling:

1. **Sensitive Attribute**: The `password` field is marked with `Sensitive: true` in the Terraform schema. This provides:
   - Redaction in `terraform plan` and `terraform apply` output
   - Redaction in CLI error messages and provider logs
   - Protection from display in the Terraform UI

2. **Environment Variable Support**: Passwords can be provided via the `RTX_PASSWORD` environment variable to avoid hardcoding in configuration files.

### Best Practices

1. **Use Environment Variables**: Always prefer environment variables over hardcoding passwords in `.tf` files:
   ```bash
   export RTX_PASSWORD="your-secure-password"
   ```

2. **Secure State Storage**: Use encrypted remote backends for Terraform state:
   - Terraform Cloud with encryption at rest
   - AWS S3 with SSE (Server-Side Encryption)
   - Google Cloud Storage with CMEK (Customer-Managed Encryption Keys)
   - Azure Blob Storage with encryption

3. **Consider SSH Key Authentication**: For enhanced security, consider implementing SSH key-based authentication in future versions.

## SSH Host Key Verification

### Implementation

The provider supports multiple methods for SSH host key verification:

1. **Fixed Host Key**: Specify a known host key directly:
   ```hcl
   provider "rtx" {
     ssh_host_key = "AAAAB3NzaC1yc2EAAAA..."
   }
   ```

2. **Known Hosts File**: Use a standard SSH known_hosts file:
   ```hcl
   provider "rtx" {
     known_hosts_file = "~/.ssh/known_hosts"
   }
   ```

3. **Skip Verification** (NOT RECOMMENDED): For testing only:
   ```hcl
   provider "rtx" {
     skip_host_key_check = true
   }
   ```

### Security Recommendations

1. Always use host key verification in production environments
2. Store host keys securely and verify them out-of-band
3. Regularly audit and update host keys as needed
4. Never use `skip_host_key_check` in production

## Connection Security

- All connections use SSH protocol with encryption
- Connection timeouts are enforced to prevent hanging connections
- Context-aware operations ensure proper resource cleanup

## Logging and Debugging

- The provider never logs passwords or sensitive data
- Debug output redacts sensitive information
- Error messages are sanitized to prevent information leakage

## Future Enhancements

1. **SSH Key Authentication**: Add support for public key authentication
2. **Vault Integration**: Support for HashiCorp Vault for credential management
3. **MFA Support**: Multi-factor authentication for enhanced security
4. **Audit Logging**: Detailed audit trails for all configuration changes