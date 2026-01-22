# Reconciliation

## Product principles
- Service resources use Cisco-like naming (`httpd`, `sshd`, `sftpd`) and keep secrets (keys) sensitive; state does not track runtime status.

## Implementation alignment
- HTTPD: host (any/interface) and proxy_access flags supported. SSHD: enable flag and hosts list present; SFTPD: hosts list. CRUD/import implemented as singletons.
- Basic validation of interface names applied.
- Gaps: no host key generation/storage handling, no multi-interface binding controls for HTTPD, no SFTP/SSH listen address/port tuning, no warnings about exposure, and imports rely on router defaults for missing fields.
