variable "rtx_host" {
  description = "RTX router IP address or hostname"
  type        = string
}

variable "rtx_username" {
  description = "RTX router username"
  type        = string
}

variable "rtx_password" {
  description = "RTX router password"
  type        = string
  sensitive   = true
}

variable "admin_password" {
  description = "RTX router administrator password"
  type        = string
  sensitive   = true
}

variable "skip_host_key_check" {
  description = "Skip SSH host key checking (use only in test environments)"
  type        = bool
  default     = false
}