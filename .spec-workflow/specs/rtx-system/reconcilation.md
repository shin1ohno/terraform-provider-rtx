# Reconciliation

## Product principles
- System resource uses Cisco-like naming for global settings; state stores only config values (no runtime stats).

## Implementation alignment
- Supports timezone offsets, console character/lines/prompt, packet_buffer tuning (small/middle/large), and statistics toggles; singleton CRUD/import implemented.
- Validation for timezone/console values present.
- Gaps: no hostname/system name, no console speed/auth controls, limited timezone handling (offset only, no named zones/DST), statistics options limited, and packet-buffer max_free/max_buffer not validated against platform defaults.
