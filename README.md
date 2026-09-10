# sovereign-recon

**Node:** AEG-NODE-865854 | **Role:** RECON | **Priority:** P1 | **Swarm:** AE Hub 8-Node Orchestrator

Reconnaissance module of the Sovereign stack. First node in the AE Hub
orchestrator swarm — maps the terrain before any other node acts.

## Position in the Swarm

| Order | Node | Role |
|-------|------|------|
| P1 | AEG-NODE-865854 | RECON (this repo) |
| P2 | AEG-NODE-534000 | INFIL |
| P3 | AEG-NODE-570081 | EXFIL |
| P4 | AEG-NODE-328975 | DEFEND |
| P5 | AEG-NODE-644943 | OFFEND |
| P6 | AEG-NODE-011260 | INTEL |
| P7 | AEG-NODE-681081 | MEDIC |
| P8 | AEG-NODE-338476 | GHOST |

## Layout

- `armada/` — OSINT Armada tool integrations and container configs
- `configs/` — node configuration
- `docs/` — operating procedures

## Related

- `sovereign-osint` — populated OSINT tooling (BUNDLEMIT sibling)
- `sovereign-agent` — agentic AI layer that tasks this node
