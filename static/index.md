---
title: killdeer.digital
description: Tiny SSH-first VM hosting page for plover.digital.
---

# killdeer.digital :: VM hosting by plover.digital

text first. ssh first. web page second.

Killdeer.digital is the VM control panel. If you already have an account, go in over SSH.
If you are looking around first, start with the help command below. The site is intentionally
tiny and the commands are the product.

## Connect

Interactive menu:

`ssh [username]@killdeer.digital`

Omit the command to enter the interactive menu.

Direct command line:

`ssh killdeer.digital help`

Start with `help` to see the available controls.

Replace `[username]` with your Killdeer account username.

## Manage Your VMs

Killdeer gives you control over your VMs and account through the same SSH connection you use to access your VMs.

- **Interactive:** Browse a menu for VM and account actions.
- **Command line:** Run individual actions directly or script your workflow.
- **Billing controls:** Set VM limits, runtime budgets, and what happens when a budget is reached.

## Access A VM

VM login command:

`ssh plover@[ip]`

Replace `[ip]` with the VM IP address from `list` or `status`.

Default VM user: `plover`.
If your account was set up with SSH, your key is already installed for `plover`.
If your account was not set up with SSH, the VM password will be emailed to you.
If you connect through the Killdeer console, use the password from that email.

## Storage That Stays Available

VM disks are stored on replicated, fault-tolerant Ceph storage. If hardware fails, healthy replicas give us a way to recover your data.

Replication protects against hardware failure. It is not a substitute for backups against accidental deletion,
compromised credentials, or application-level mistakes.

## Help Command Output

See:

- https://killdeer.digital/ssh-help.txt

The live helper still prints some examples with `killdeer.plover.digital`.
This page normalizes them to the public hostname `killdeer.digital`.

The help output covers command-line VM management and billing controls. This site also includes sizes,
pricing, OS images, and signup credits.

## Sizes, Pricing And Billing

```text
SIZE      CPU     RAM   DISK    BASE    RUNTIME     EST. 24/7
Micro     1 vCPU  1 GB  10 GB   $2/mo   $0.0055/hr  ~$6/mo
Basic     1 vCPU  2 GB  25 GB   $4/mo   $0.0110/hr  ~$12/mo
Standard  2 vCPU  2 GB  50 GB   $6/mo   $0.0164/hr  ~$18/mo
Premium   2 vCPU  4 GB  50 GB   $6/mo   $0.0247/hr  ~$24/mo
Ultra     4 vCPU  4 GB  100 GB  $11/mo  $0.0397/hr  ~$40/mo
Mega      4 vCPU  8 GB  100 GB  $11/mo  $0.0507/hr  ~$48/mo
```

Manage VM limits and monthly runtime budgets from the same SSH experience. Choose whether reaching a budget
sends a notification, stops VMs, or locks further runtime.

Base fee includes $1/mo for IPv4 + $1/mo per 10GB storage.
IPv6-only VMs save $1/mo (no IPv4 charge).
Runtime is charged hourly only when the VM is powered on.
Base fee is prorated when VMs are created or deleted mid-cycle.

## OS Images

Public command:

`ssh killdeer.digital os`

```text
Available OS images:

Main distros:

DISTRO    VERSIONS
Alpine    3.23
Debian    13
Fedora    44
Rocky     10
Ubuntu    26.04, 24.04

Experimental distros:

DISTRO      VERSIONS
Arch Linux  current
NixOS       25.11

Ubuntu 26.04 is the default Ubuntu image.
The SSH CLI also accepts full active image names.
```

## Promotions

- New users get $20 credit.
- Students get double credit with a valid `.edu` email.
- Referrals: $5 for you, $5 for them.

## Notes For Humans And Agents

- Public SSH endpoint: `killdeer.digital`
- Preferred login syntax: `ssh [username]@killdeer.digital`
- Machine-readable guide: https://killdeer.digital/llms.txt
- Full agent bundle: https://killdeer.digital/llms-full.txt
- Normalized CLI help text: https://killdeer.digital/ssh-help.txt
- Sizes and pricing: https://killdeer.digital/sizes.txt
- OS images: https://killdeer.digital/os.txt
- CLI metadata: https://killdeer.digital/api/v1/cli.json
- OS image metadata: https://killdeer.digital/api/v1/images.json
- Agent skills index: https://killdeer.digital/.well-known/agent-skills/index.json
- Agent skill for CLI usage: https://killdeer.digital/.well-known/agent-skills/killdeer-cli/SKILL.md

## Get In Touch

- Email: machines@plover.digital
- Discord: https://discord.gg/AhD77Raqru

## Mailing List

Use the `Join the mailing list` action on the homepage to open the Mailjet-hosted signup modal.
If the modal is blocked, contact Killdeer over email or Discord.

## Footer Note

killdeer.digital is a project by plover.digital, an independent service provider building tools and infrastructure for developers.
