# Enable SFTP service on LAN interface
# Note: Requires rtx_sshd to be enabled
resource "rtx_sftpd" "file_transfer" {
  hosts = ["lan1"]
}
