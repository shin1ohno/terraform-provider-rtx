# Requirements: Terraform Plan Differences Fix

## Overview

This specification addresses the 4 remaining differences identified during `terraform plan` execution in the `examples/import` directory. These differences must be resolved to achieve a clean `terraform plan` with no changes.

## Detected Differences

### Difference 1: rtx_dhcp_scope.scope1 - Network Forces Replacement

**Status:** Must be replaced (forces replacement)

**Terraform Plan Output:**
```hcl
-/+ resource "rtx_dhcp_scope" "scope1" {
      ~ id          = "1" -> (known after apply)
      + network     = "192.168.0.0/16" # forces replacement
      + range_end   = (known after apply)
      + range_start = (known after apply)
        # (1 unchanged attribute hidden)

      ~ options {
          - dns_servers = [] -> null
            # (2 unchanged attributes hidden)
        }
    }
```

**Current State:**
```hcl
resource "rtx_dhcp_scope" "scope1" {
    id          = "1"
    lease_time  = null
    network     = null
    range_end   = null
    range_start = null
    scope_id    = 1

    options {
        dns_servers = []
        domain_name = null
        routers     = ["192.168.1.253"]
    }
}
```

**1-a. RTX Command (Source):**
```
dhcp scope 1 192.168.1.20-192.168.1.99/16 gateway 192.168.1.253 expire 12:00 maxexpire 24:00
dhcp scope option 1 router=192.168.1.253
```

**1-b. Expected main.tf:**
```hcl
resource "rtx_dhcp_scope" "scope1" {
  scope_id = 1
  network  = "192.168.0.0/16"
  # lease_time not imported - provider read limitation

  options {
    routers = ["192.168.1.253"]
  }
}
```

---

### Difference 2: rtx_ipv6_filter_dynamic.main - Will Be Created

**Status:** Will be created (resource not in state)

**Terraform Plan Output:**
```hcl
+ resource "rtx_ipv6_filter_dynamic" "main" {
      + id = (known after apply)

      + entry {
          + destination = "*"
          + number      = 101080
          + protocol    = "ftp"
          + source      = "*"
          + syslog      = false
        }
      + entry {
          + destination = "*"
          + number      = 101081
          + protocol    = "domain"
          + source      = "*"
          + syslog      = false
        }
      # ... 6 more entries (101082-101099)
    }
```

**Current State:** No state (resource never imported)

**1-a. RTX Command (Source):**
```
ipv6 filter dynamic 101080 * * ftp syslog=off
ipv6 filter dynamic 101081 * * domain syslog=off
ipv6 filter dynamic 101082 * * www syslog=off
ipv6 filter dynamic 101083 * * smtp syslog=off
ipv6 filter dynamic 101084 * * pop3 syslog=off
ipv6 filter dynamic 101085 * * submission syslog=off
ipv6 filter dynamic 101098 * * tcp syslog=off
ipv6 filter dynamic 101099 * * udp syslog=off
```

**1-b. Expected main.tf:**
```hcl
resource "rtx_ipv6_filter_dynamic" "main" {
  entry {
    number      = 101080
    source      = "*"
    destination = "*"
    protocol    = "ftp"
    syslog      = false
  }

  entry {
    number      = 101081
    source      = "*"
    destination = "*"
    protocol    = "domain"
    syslog      = false
  }

  entry {
    number      = 101082
    source      = "*"
    destination = "*"
    protocol    = "www"
    syslog      = false
  }

  entry {
    number      = 101083
    source      = "*"
    destination = "*"
    protocol    = "smtp"
    syslog      = false
  }

  entry {
    number      = 101084
    source      = "*"
    destination = "*"
    protocol    = "pop3"
    syslog      = false
  }

  entry {
    number      = 101085
    source      = "*"
    destination = "*"
    protocol    = "submission"
    syslog      = false
  }

  entry {
    number      = 101098
    source      = "*"
    destination = "*"
    protocol    = "tcp"
    syslog      = false
  }

  entry {
    number      = 101099
    source      = "*"
    destination = "*"
    protocol    = "udp"
    syslog      = false
  }
}
```

---

### Difference 3: rtx_l2tp.tunnel1 - tunnel_auth_enabled Mismatch

**Status:** Will be updated in-place

**Terraform Plan Output:**
```hcl
~ resource "rtx_l2tp" "tunnel1" {
        id                 = "1"
        name               = "ebisu-RTX1210"
        # (13 unchanged attributes hidden)

      ~ l2tpv3_config {
          ~ tunnel_auth_enabled  = true -> false
          - tunnel_auth_password = (sensitive value) -> null
            # (6 unchanged attributes hidden)
        }

        # (1 unchanged block hidden)
    }
```

**Current State:**
```hcl
resource "rtx_l2tp" "tunnel1" {
    # ... other attributes ...
    l2tpv3_config {
        tunnel_auth_enabled  = true
        tunnel_auth_password = (sensitive value)
        local_router_id      = "192.168.1.253"
        remote_router_id     = "192.168.1.254"
        remote_end_id        = "shin1"
        # ...
    }
}
```

**1-a. RTX Command (Source):**
Based on the state showing `tunnel_auth_enabled = true`, the RTX configuration likely includes:
```
l2tp tunnel 1 auth on
l2tp tunnel 1 auth password <password>
```

**1-b. Expected main.tf:**
```hcl
resource "rtx_l2tp" "tunnel1" {
  tunnel_id          = 1
  name               = "ebisu-RTX1210"
  version            = "l2tpv3"
  mode               = "l2vpn"
  tunnel_destination = "itm.ohno.be"
  tunnel_dest_type   = "fqdn"
  always_on          = true
  keepalive_enabled  = true
  keepalive_interval = 60
  keepalive_retry    = 3
  disconnect_time    = 0

  l2tpv3_config {
    local_router_id     = "192.168.1.253"
    remote_router_id    = "192.168.1.254"
    remote_end_id       = "shin1"
    tunnel_auth_enabled = false  # <-- This is the source of the difference
  }

  ipsec_profile {
    enabled   = true
    tunnel_id = 101
  }
}
```

---

### Difference 4: rtx_nat_masquerade.nat1000 - Will Be Created

**Status:** Will be created (resource not in state)

**Terraform Plan Output:**
```hcl
+ resource "rtx_nat_masquerade" "nat1000" {
      + descriptor_id = 1000
      + id            = (known after apply)
      + outer_address = "primary"

      + static_entry {
          + entry_number        = 2
          + inside_local        = "192.168.1.253"
          + inside_local_port   = 500
          + outside_global      = "primary"
          + outside_global_port = 500
          + protocol            = "udp"
        }
      # ... more static_entry blocks (3, 4, 900)
    }
```

**Current State:** No state (resource never imported)

**1-a. RTX Command (Source):**
Based on interface configuration referencing `nat descriptor 1000`:
```
nat descriptor type 1000 masquerade
nat descriptor address outer 1000 primary
nat descriptor masquerade static 1000 1 192.168.1.253 esp
nat descriptor masquerade static 1000 2 192.168.1.253 udp 500
nat descriptor masquerade static 1000 3 192.168.1.253 udp 4500
nat descriptor masquerade static 1000 4 192.168.1.253 udp 1701
nat descriptor masquerade static 1000 900 192.168.1.20 tcp 55000
```

**1-b. Expected main.tf:**
```hcl
resource "rtx_nat_masquerade" "nat1000" {
  descriptor_id = 1000
  outer_address = "primary"

  # Entry 1: ESP protocol (protocol-only, no ports)
  static_entry {
    entry_number   = 1
    inside_local   = "192.168.1.253"
    outside_global = "primary"
    protocol       = "esp"
  }

  static_entry {
    entry_number        = 2
    inside_local        = "192.168.1.253"
    inside_local_port   = 500
    outside_global      = "primary"
    outside_global_port = 500
    protocol            = "udp"
  }

  static_entry {
    entry_number        = 3
    inside_local        = "192.168.1.253"
    inside_local_port   = 4500
    outside_global      = "primary"
    outside_global_port = 4500
    protocol            = "udp"
  }

  static_entry {
    entry_number        = 4
    inside_local        = "192.168.1.253"
    inside_local_port   = 1701
    outside_global      = "primary"
    outside_global_port = 1701
    protocol            = "udp"
  }

  static_entry {
    entry_number        = 900
    inside_local        = "192.168.1.20"
    inside_local_port   = 55000
    outside_global      = "primary"
    outside_global_port = 55000
    protocol            = "tcp"
  }
}
```

---

## Root Cause Analysis Summary

| Resource | Issue Type | Root Cause |
|----------|------------|------------|
| rtx_dhcp_scope.scope1 | Provider Import Bug | Import function doesn't capture `network` field |
| rtx_ipv6_filter_dynamic.main | Missing Import | Resource never imported to state |
| rtx_l2tp.tunnel1 | Configuration Mismatch | main.tf doesn't match actual RTX config |
| rtx_nat_masquerade.nat1000 | Missing Import | Resource never imported to state |

---

## Acceptance Criteria

### AC-1: DHCP Scope Network Field
- [ ] After import, `network` field should be populated with correct CIDR notation
- [ ] `terraform plan` shows no changes for rtx_dhcp_scope.scope1

### AC-2: IPv6 Dynamic Filter Import
- [ ] `terraform import rtx_ipv6_filter_dynamic.main main` succeeds
- [ ] Imported state matches the RTX configuration
- [ ] `terraform plan` shows no changes for rtx_ipv6_filter_dynamic.main

### AC-3: L2TP Tunnel Auth Configuration
- [ ] main.tf is updated to match actual RTX configuration OR
- [ ] RTX configuration is intentionally changed via terraform apply
- [ ] `terraform plan` shows no changes for rtx_l2tp.tunnel1

### AC-4: NAT Masquerade Import
- [ ] `terraform import rtx_nat_masquerade.nat1000 1000` succeeds
- [ ] Imported state matches the RTX configuration
- [ ] `terraform plan` shows no changes for rtx_nat_masquerade.nat1000

### AC-5: Overall
- [ ] Running `terraform plan -parallelism=2` shows "No changes. Your infrastructure matches the configuration."
