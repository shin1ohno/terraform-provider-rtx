#!/bin/sh
# Simple RTX router command simulator for testing

# Get model from environment variable or default to RTX1210
RTX_MODEL=${RTX_MODEL:-RTX1210}

# Simulate RTX prompt based on model
PROMPT="$RTX_MODEL# "

# Function to simulate command responses
execute_command() {
    case "$1" in
        "show environment")
            cat << EOF
$RTX_MODEL Rev.14.01.38 (Fri Jul 12 10:45:32 2024)
main:  $RTX_MODEL ver=00 serial=XXXXXXXXXX MAC-Address=XX:XX:XX:XX:XX:XX
Temperature: 42.5C
CPU: 15%
Memory: 55%
EOF
            ;;
        "show system information")
            cat << EOF
$RTX_MODEL System Information
Model: $RTX_MODEL
Firmware Version: 14.01.38
Serial Number: XXXXXXXXXX
MAC Address: XX:XX:XX:XX:XX:XX
Boot Time: 2024-01-01 00:00:00
Uptime: 10 days 5:30:45
EOF
            ;;
        "show status boot")
            cat << EOF
$RTX_MODEL Rev.14.01.38 (Fri Jul 12 10:45:32 2024)
Reboot by power on
Uptime: 10 days 5:30:45
EOF
            ;;
        "show config")
            cat << EOF
# $RTX_MODEL Configuration
ip route default gateway 192.168.1.1
ip lan1 address 192.168.1.254/24
dns server 8.8.8.8 8.8.4.4
EOF
            ;;
        "show ip route")
            # Simulate routing table output
            cat << EOF
Codes: C - connected, S - static, R - RIP, O - OSPF,
       B - BGP, D - DHCP, I - Implicit-gateway

Destination         Gateway          Interface        Type
default             203.0.113.1      WAN1             S
10.10.10.0/24       *                VLAN10           C
172.16.0.0/24       10.10.10.5       VLAN10           S
192.168.1.0/24      *                LAN1             C
203.0.113.0/30      *                WAN1             C
EOF
            ;;
        "show interface")
            # Different output based on model
            case "$RTX_MODEL" in
                "RTX830")
                    cat << EOF
LAN1: UP
  IP Address: 192.168.1.254/24
  MAC Address: 00:A0:DE:12:34:56
LAN2: DOWN
  IP Address: Not configured
  MAC Address: 00:A0:DE:12:34:57
WAN1: UP
  IP Address: 203.0.113.1/30
  MAC Address: 00:A0:DE:12:34:58
PP1: UP
  IP Address: PPPoE
VLAN10: UP
  IP Address: 10.10.10.1/24
  MAC Address: 00:A0:DE:12:34:59
EOF
                    ;;
                "RTX1210"|"RTX1220")
                    cat << EOF
Interface LAN1
  Status: up
  IPv4: 192.168.1.254/24
  IPv6: Not configured
  Ethernet address: 00:A0:DE:AB:CD:01
  MTU: 1500

Interface LAN2
  Status: down
  IPv4: Not configured
  IPv6: Not configured
  Ethernet address: 00:A0:DE:AB:CD:02
  MTU: 1500

Interface WAN1
  Status: up
  IPv4: 203.0.113.1/30
  IPv6: 2001:db8::1/64
  Ethernet address: 00:A0:DE:AB:CD:03
  MTU: 1500

Interface PP1
  Status: up
  Type: PPPoE
  MTU: 1454

Interface VLAN10
  Status: up
  IPv4: 10.10.10.1/24
  Ethernet address: 00:A0:DE:AB:CD:04
  MTU: 1500
EOF
                    ;;
                *)
                    echo "Error: Unknown model $RTX_MODEL"
                    ;;
            esac
            ;;
        "exit"|"quit")
            exit 0
            ;;
        *)
            echo "Error: Unknown command '$1'"
            ;;
    esac
}

# Main loop - simulate interactive shell
while true; do
    printf "%s" "$PROMPT"
    read -r cmd
    [ -z "$cmd" ] && continue
    execute_command "$cmd"
done