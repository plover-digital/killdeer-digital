# killdeer.digital

Small landing page for Plover.digital's Killdeer VM hosting service.

Ongoing design + architecture notes live in [docs/BUILD_LOG.md](docs/BUILD_LOG.md).

## What is here

- A single text-forward landing page at `static/index.html`
- A normalized copy of `ssh killdeer.digital help` at `static/ssh-help.txt`
- A plain-text copy of `ssh killdeer.digital sizes` at `static/sizes.txt`
- A plain-text copy of `ssh killdeer.digital os` at `static/os.txt`
- An agent-facing guide at `static/llms.txt`
- A fuller agent bundle at `static/llms-full.txt`
- Read-only public metadata at `static/api/v1/cli.json`, `static/api/v1/sizes.json`, and `static/api/v1/images.json`
- Agent skills under `static/.well-known/agent-skills/`
- API discovery through `static/.well-known/api-catalog` and `static/openapi.json`
- A markdown version of the homepage at `static/index.md`
- Agent discovery files at `static/robots.txt` and `static/sitemap.xml`
- A tiny Go server that serves the page and discovery files
- A tiny Caddy front end for the containerized setup
- A `Containerfile` and `compose.yaml` for Podman

## Local development

```sh
go run .
```

The site listens on `http://localhost:8080` by default.

Useful local routes:

- `/` for the homepage
- `/index.md` for the markdown version of the homepage
- `/llms.txt` for the short agent index
- `/llms-full.txt` for the fuller agent bundle
- `/ssh-help.txt` for normalized command help
- `/sizes.txt` for sizes and pricing
- `/os.txt` for available OS images
- `/api/v1/cli.json` for structured SSH CLI guidance
- `/api/v1/sizes.json` for structured size and pricing data
- `/api/v1/images.json` for structured OS image data
- `/.well-known/agent-skills/index.json` for agent skill discovery
- `/.well-known/api-catalog` for public metadata API discovery
- `/openapi.json` for the read-only metadata API description
- `/robots.txt` and `/sitemap.xml` for crawler discovery

If a client requests `/` with `Accept: text/markdown`, the server now serves the markdown homepage instead of HTML.

## Podman

```sh
cp .env.example .env
podman compose up --build
```

Then visit `http://localhost:8080`.

The main bind settings live in `.env`:

- `BIND_IP` controls which host IP Caddy publishes on
- `HTTP_PORT` controls which host port maps to Caddy's port `80`
- `HTTPS_PORT` controls which host port maps to Caddy's port `443`

Example:

```sh
BIND_IP=216.66.77.166
HTTP_PORT=80
HTTPS_PORT=443
```

If you are running this on a Podman host with SELinux, the Caddy config bind mount is relabeled through the compose file with `:Z` so Caddy can read `/etc/caddy/Caddyfile` correctly.

For rootless Podman on a Linux host, publishing low ports like `80` and `443` is a host-level setting.
If the host still blocks unprivileged low ports, a common setup is:

```sh
echo 'net.ipv4.ip_unprivileged_port_start=80' | sudo tee /etc/sysctl.d/99-podman-low-ports.conf
sudo sysctl --system
```

That change allows rootless containers to bind ports `80` and higher, which covers both `80` and `443`.
If you do not want to change the host setting, use higher host ports in `.env` instead.

Caddy certificate state is stored in persistent named volumes:

- `caddy-data` keeps ACME accounts, issued certificates, and renewal state
- `caddy-config` keeps Caddy's persisted runtime config state

Those volumes should be kept between deploys so Caddy can reuse existing certificates instead of repeatedly requesting new ones.

If you want to be certain the running container is replaced after a site change, use:

```sh
podman-compose up --build --force-recreate
```

The compose stack now looks like this:

- `caddy` listens on the host HTTP/HTTPS ports and reverse proxies requests
- `killdeer-site` serves the embedded Go site on the internal compose network

The app image is still pinned to `killdeer-digital:latest` so direct `podman build` and compose runs refer to the same local image tag.

## Mailing list embed

The homepage uses a Mailjet-hosted embed for the mailing list.

This means:

- the Go server does not need Mailjet API credentials
- local development only needs `PORT`
- if the signup form needs to change, update the Mailjet iframe and script snippet in `static/index.html`

## Note about the SSH help text

The live helper currently emits examples that still mention `killdeer.plover.digital`. The site intentionally normalizes those examples to the public hostname `killdeer.digital`.
