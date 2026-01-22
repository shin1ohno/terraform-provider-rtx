# SNMP Community Strings - SENSITIVE VALUES
# Best practice: Use environment variables or a secrets manager
# Example: export TF_VAR_snmp_community_ro="your_community_string"

variable "snmp_community_ro" {
  description = "SNMP read-only community string"
  type        = string
  sensitive   = true
  default     = "public" # Change this in production!

  validation {
    condition     = length(var.snmp_community_ro) >= 8
    error_message = "Community string should be at least 8 characters for security."
  }
}

variable "snmp_community_rw" {
  description = "SNMP read-write community string"
  type        = string
  sensitive   = true
  default     = "private" # Change this in production!

  validation {
    condition     = length(var.snmp_community_rw) >= 12
    error_message = "Read-write community string should be at least 12 characters for security."
  }
}
