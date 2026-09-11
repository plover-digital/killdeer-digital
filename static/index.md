# killdeer.digital

Linux VMs. Managed over SSH.

We're a small, independent team in Minneapolis building our own hosting stack. Create and manage your virtual machines from your terminal.

## Connect

Sign up or open the terminal UI (TUI) over SSH to manage your VMs:

```sh
ssh [username]@killdeer.digital
```

Replace `[username]` with your username. Signing in for the first time with an available username creates your account. New accounts receive **$20 credit**.

You can also create a VM directly from the command line. For example:

```sh
ssh [username]@killdeer.digital create test-vm micro alpine-3.23 --ipv6
```

This example creates an IPv6-only Micro VM named `test-vm` running Alpine Linux. [All commands](https://killdeer.digital/ssh-help.txt).

## Sizes & pricing

Monthly estimates with the VM running 24/7, including IPv4.

| Size | vCPU | RAM | Disk | Est. / month |
| --- | ---: | ---: | ---: | ---: |
| Micro | 1 | 1 GB | 10 GB | ~$6 |
| Basic | 1 | 2 GB | 25 GB | ~$12 |
| Standard | 2 | 2 GB | 50 GB | ~$18 |
| Premium | 2 | 4 GB | 50 GB | ~$24 |
| Ultra | 4 | 4 GB | 100 GB | ~$40 |
| Mega | 4 | 8 GB | 100 GB | ~$48 |

Pay a base fee for storage and IPv4, plus hourly runtime while your VM is powered on. IPv6-only VMs save $1/month.

Base fees are prorated when you create or delete a VM. [Full pricing & hourly rates](https://killdeer.digital/sizes.txt).

## Documentation

- [Commands](https://killdeer.digital/ssh-help.txt)
- [Operating systems](https://killdeer.digital/os.txt)
- [Agent guide](https://killdeer.digital/llms.txt)

Operated by [Plover Digital](https://plover.digital/). [Service status](https://status.plover.digital/).
