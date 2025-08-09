#!/bin/sh
# Simple RTX router command simulator for testing

# Simulate RTX prompt
PROMPT="RTX1210# "

# Function to simulate command responses
execute_command() {
    case "$1" in
        "show environment")
            cat << EOF
RTX1210 Rev.14.01.38 (Fri Jul 12 10:45:32 2024)
main:  RTX1210 ver=00 serial=XXXXXXXXXX MAC-Address=XX:XX:XX:XX:XX:XX
Temperature: 42.5C
CPU: 15%
Memory: 55%
EOF
            ;;
        "show system information")
            cat << EOF
RTX1210 System Information
Model: RTX1210
Firmware Version: 14.01.38
Serial Number: XXXXXXXXXX
MAC Address: XX:XX:XX:XX:XX:XX
Boot Time: 2024-01-01 00:00:00
Uptime: 10 days 5:30:45
EOF
            ;;
        "show status boot")
            cat << EOF
RTX1210 Rev.14.01.38 (Fri Jul 12 10:45:32 2024)
Reboot by power on
Uptime: 10 days 5:30:45
EOF
            ;;
        "show config")
            cat << EOF
# RTX1210 Configuration
ip route default gateway 192.168.1.1
ip lan1 address 192.168.1.254/24
dns server 8.8.8.8 8.8.4.4
EOF
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