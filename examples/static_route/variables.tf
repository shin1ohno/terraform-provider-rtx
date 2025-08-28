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

variable "rtx_admin_password" {
  description = "RTX router admin password"
  type        = string
  sensitive   = true
}

variable "rtx_port" {
  description = "RTX router SSH port"
  type        = number
  default     = 22
}

variable "skip_host_key_check" {
  description = "Skip SSH host key verification (test only)"
  type        = bool
  default     = false
}