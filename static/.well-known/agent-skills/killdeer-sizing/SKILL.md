---
name: killdeer-sizing
description: Help choose a Killdeer VM size and explain billing notes.
---

# Killdeer Sizing

Use this skill when a user asks what Killdeer VM size to choose or how Killdeer pricing works.

## Size Table

| Size | CPU | RAM | Disk | Base | Runtime | Estimated 24/7 |
| --- | --- | --- | --- | --- | --- | --- |
| Micro | 1 vCPU | 1 GB | 10 GB | $2/mo | $0.0055/hr | ~$6/mo |
| Basic | 1 vCPU | 2 GB | 25 GB | $4/mo | $0.0110/hr | ~$12/mo |
| Standard | 2 vCPU | 2 GB | 50 GB | $6/mo | $0.0164/hr | ~$18/mo |
| Premium | 2 vCPU | 4 GB | 50 GB | $6/mo | $0.0247/hr | ~$24/mo |
| Ultra | 4 vCPU | 4 GB | 100 GB | $11/mo | $0.0397/hr | ~$40/mo |
| Mega | 4 vCPU | 8 GB | 100 GB | $11/mo | $0.0507/hr | ~$48/mo |

## Billing Notes

Base fee includes `$1/mo` for IPv4 plus `$1/mo` per 10GB storage.
IPv6-only VMs save `$1/mo` because there is no IPv4 charge.
Runtime is charged hourly only when the VM is powered on.
Base fee is prorated when VMs are created or deleted mid-cycle.

## CLI Usage

Show live sizes:

```sh
ssh [username]@killdeer.digital sizes
```

Create a VM with a chosen size:

```sh
ssh [username]@killdeer.digital create <name> <size> <os> [ip-type]
```

Ask for the user's Killdeer username before constructing account-specific commands.

## Canonical References

- Size metadata: https://killdeer.digital/api/v1/sizes.json
- Plain-text size table: https://killdeer.digital/sizes.txt
- CLI skill: https://killdeer.digital/.well-known/agent-skills/killdeer-cli/SKILL.md
