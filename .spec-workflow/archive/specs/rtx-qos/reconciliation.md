# Reconciliation

## Product principles
- QoS resources adopt Cisco MQC naming (class_map/policy_map/service_policy/shape) consistent with Cisco compatibility principle; state contains config only.

## Implementation alignment
- Resources exist for class_map (protocol/ports), policy_map (class actions with priority/bandwidth/police), service_policy (interface/direction binding), and shape (rate/queue length); basic CRUD/import implemented.
- Matches cover simple protocol/port filters and queue length; shaping supports rate/committed/burst fields.
- Gaps: missing DSCP/TOS/interface matching, queue types/WRED/ECN, weighted fair queuing, bandwidth guarantees per interface, drop policies, priority levels beyond boolean, and validation for bandwidth units; no coverage for multi-queue scheduling or class-default behaviors.
