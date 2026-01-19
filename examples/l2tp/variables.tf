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

variable "l2tp_username" {
  description = "L2TP authentication username"
  type        = string
}

variable "l2tp_password" {
  description = "L2TP authentication password"
  type        = string
  sensitive   = true
}

variable "tunnel_password" {
  description = "L2TPv3 tunnel password"
  type        = string
  sensitive   = true
}
