---
name: killdeer-cli
description: Use the Killdeer SSH command line to help a user manage developer VMs.
---

# Killdeer CLI

Use this skill when a user wants to understand or operate the Killdeer VM hosting control plane.

Killdeer is SSH-first. The public control surface is the command line over `killdeer.digital`.
Do not invent HTTP VM-management API calls.

## Hostname

Use `killdeer.digital` as the public SSH hostname.

## First Step

If the user has not provided their Killdeer username, ask for it before constructing account-specific commands.

Use this login pattern:

```sh
ssh [username]@killdeer.digital
```

Use this public discovery command:

```sh
ssh killdeer.digital help
```

If a user omits the command after login, Killdeer opens the interactive menu.

## Access A VM

Use this VM login pattern:

```sh
ssh plover@[ip]
```

Replace `[ip]` with the VM IP address from `list` or `status`.

The default VM user is `plover`. If the user's Killdeer account was set up with SSH,
their key is already installed for `plover`. Otherwise, the VM password is emailed to them.
If the user connects through the Killdeer console, they must use the password from that email.

## Common Commands

List VMs:

```sh
ssh [username]@killdeer.digital list
```

Inspect a VM:

```sh
ssh [username]@killdeer.digital status <name>
```

Start, stop, or restart a VM:

```sh
ssh [username]@killdeer.digital start <name>
ssh [username]@killdeer.digital stop <name>
ssh [username]@killdeer.digital restart <name>
```

Resize a stopped VM:

```sh
ssh [username]@killdeer.digital resize <name> <size>
```

Create a VM:

```sh
ssh [username]@killdeer.digital create <name> <size> <os> [ip-type]
```

Allowed `ip-type` values are `--ipv4`, `--ipv6`, and `--dualstack`. The default is `--ipv4`.

Delete a VM:

```sh
ssh [username]@killdeer.digital delete <name> --confirm
```

Deletion is destructive and requires `--confirm`. Present delete commands carefully.

Attach to serial console:

```sh
ssh [username]@killdeer.digital console <name>
```

Exit the console with `Ctrl+]`.

Manage shutdown timers:

```sh
ssh [username]@killdeer.digital timer <name>
ssh [username]@killdeer.digital timer <name> 30m
ssh [username]@killdeer.digital timer <name> +2h
ssh [username]@killdeer.digital timer <name> -30m
ssh [username]@killdeer.digital timer <name> cancel
ssh [username]@killdeer.digital timer <name> force
```

Timer durations include `30m`, `2h`, and `1d`. The maximum is 7 days.

List sizes and OS images:

```sh
ssh [username]@killdeer.digital sizes
ssh [username]@killdeer.digital images
ssh [username]@killdeer.digital os
```

The `images` command and `os` alias list the same OS images. The CLI accepts either the image shorthand or full active image name when creating a VM.

Current public image list from `ssh killdeer.digital os`:

```text
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
```

Ubuntu 26.04 is the default Ubuntu image. Run `ssh [username]@killdeer.digital os` when exact image names are needed.

## Aliases

`ls` maps to `list`.
`st` and `info` map to `status`.
`rm` maps to `delete`.
`os` maps to `images`.

## Canonical References

- CLI metadata: https://killdeer.digital/api/v1/cli.json
- Normalized help text: https://killdeer.digital/ssh-help.txt
- Size metadata: https://killdeer.digital/api/v1/sizes.json
- OS image list: https://killdeer.digital/os.txt
- OS image metadata: https://killdeer.digital/api/v1/images.json
- Full agent guide: https://killdeer.digital/llms-full.txt
