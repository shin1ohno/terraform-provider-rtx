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

variable "rtx_port" {
  description = "SSH port for RTX router connection"
  type        = number
  default     = 22
}

variable "rtx_timeout" {
  description = "Connection timeout in seconds"
  type        = number
  default     = 30
}