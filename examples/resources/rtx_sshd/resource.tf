# Enable SSH service on all interfaces
resource "rtx_sshd" "ssh_access" {
  enabled = true
}

# Enable SSH restricted to specific interfaces
# resource "rtx_sshd" "secure_ssh" {
#   enabled = true
#   hosts   = ["lan1", "lan2"]
# }
