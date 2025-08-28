variable "rtx_host" {
  description = "RTX router IP address or hostname"
  type        = string
  default     = "192.168.1.253"
}

variable "rtx_username" {
  description = "RTX router username"
  type        = string
  default     = "testuser"
}

variable "rtx_password" {
  description = "RTX router password"
  type        = string
  sensitive   = true
  default     = "testpass"
}

variable "rtx_admin_password" {
  description = "RTX router admin password"
  type        = string
  sensitive   = true
  default     = "password"
}

variable "rtx_port" {
  description = "RTX router SSH port"
  type        = number
  default     = 22
}