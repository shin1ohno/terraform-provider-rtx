variable "rtx_host" {
  description = "Hostname or IP address of the RTX router"
  type        = string
}

variable "rtx_username" {
  description = "Username for RTX router authentication"
  type        = string
}

variable "rtx_password" {
  description = "Password for RTX router authentication"
  type        = string
  sensitive   = true
}

variable "rtx_admin_password" {
  description = "Administrator password for configuration changes (optional, defaults to rtx_password)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "skip_host_key_check" {
  description = "Skip SSH host key verification (not recommended for production)"
  type        = bool
  default     = false
}
