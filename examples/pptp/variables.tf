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

variable "pptp_username" {
  description = "PPTP authentication username"
  type        = string
}

variable "pptp_password" {
  description = "PPTP authentication password"
  type        = string
  sensitive   = true
}
