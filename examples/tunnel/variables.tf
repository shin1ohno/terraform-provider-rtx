# Provider configuration variables
variable "rtx_host" {
  description = "RTX router hostname or IP address"
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

# Tunnel configuration variables
variable "ipsec_psk" {
  description = "IPsec pre-shared key"
  type        = string
  sensitive   = true
  default     = "example!PSK123"
}

variable "l2tp_password" {
  description = "L2TP tunnel authentication password"
  type        = string
  sensitive   = true
  default     = "example!L2TP456"
}
