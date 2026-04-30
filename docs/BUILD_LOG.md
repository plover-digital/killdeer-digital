# killdeer.digital build log

```
    __   _ ____    __                      ___       _ __        __
   / /__(_) / /___/ /__  ___  _____   ____/ (_)___ _(_) /_____ _/ /
  / //_/ / / / __  / _ \/ _ \/ ___/  / __  / / __ `/ / __/ __ `/ /
 / ,< / / / / /_/ /  __/  __/ /     / /_/ / / /_/ / / /_/ /_/ / /
/_/|_/_/_/_/\__,_/\___/\___/_/      \__,_/_/\__, /_/\__/\__,_/_/
                                           /____/
```

This file is the running design + architecture diary for the site.

Rule of the road:
- Update this document before moving on to the next implementation or verification step.

## Snapshot

Date:
- 2026-04-20

Project goal:
- Build a tiny website for Plover.digital's Killdeer VM hosting service.
- Keep the actual website very small and text-forward.
- Push users toward the SSH control plane instead of building a web dashboard.
- Help both humans and AI agents understand how to operate the service.
- Collect a mailing list through a Mailjet-hosted form.

## Inputs We Used

Reference vibe:
- `https://plover.digital/`
- `https://www.terminal.shop/`

Live control-plane reference:
- `ssh killdeer.digital help`

Important live detail:
- The current helper output still prints some examples using `killdeer.plover.digital`.
- The public hostname we want to teach people is `killdeer.digital`.

Decision:
- Normalize the website copy and text resources to `killdeer.digital`.
- Keep a note in the docs and UI explaining why the examples differ from the live helper text.

## Design Direction

Mood board in plain ASCII:

```text
small page
big command
paper + terminal
friendly but dry
minimal, not sterile
```

Design choices made:
- Single page HTML with inline CSS and JS so the site stays tiny.
- Text-first layout inspired by the sparse command-led flow of terminal.shop.
- Copy-to-clipboard buttons because the main job of the page is to get the user to the right SSH command quickly.
- Soft terminal/paper styling instead of a heavy "fake hacker terminal" gimmick.
- Monospace-first presentation to match the SSH-first product.
- Explicit "humans and agents" guidance baked into the page copy.

Why:
- The service itself is command-line driven, so the site should feel like a short note passed between operators.
- Plover.digital already sets the tone with a stripped-down, trust-building presentation.
- Terminal.shop shows that a single command can carry the whole conversion path if the page stays focused.

## Architecture Direction

Current shape:

```text
browser
  |
  +--> Caddy
         |
         +--> Go site
                |
                +--> GET /
                +--> GET /index.md
                +--> GET /llms.txt
                +--> GET /llms-full.txt
                +--> GET /ssh-help.txt
                +--> GET /sizes.txt
                +--> GET /robots.txt
                +--> GET /sitemap.xml
                |
                +--> Mailjet iframe + script
                          |
                          +--> Mailjet hosted signup form
```

Files added so far:
- `Caddyfile`
- `static/index.html`
- `static/ssh-help.txt`
- `static/llms.txt`
- `main.go`
- `go.mod`
- `Containerfile`
- `compose.yaml`
- `.env.example`
- `.gitignore`
- `README.md`

## Why These Architecture Choices

### Tiny Go server

Choice:
- Use a very small Go HTTP server instead of a bigger framework.

Why:
- The site only needs to serve a handful of static text files.
- Go compiles to a single binary, which makes Podman deployment simple.
- It keeps the operational story close to the product story: small, direct, boring in a good way.

### Tiny Caddy front end

Choice:
- Put Caddy in front of the Go app in the compose stack.

Why:
- Caddy is a clean reverse proxy for the public-facing container port.
- It gives the project an obvious place for future host routing or TLS changes.
- The Go app can stay private to the compose network.

### Embedded static files

Choice:
- Embed `index.html`, `llms.txt`, and `ssh-help.txt` into the Go binary.

Why:
- Avoids serving the rest of the repository by accident.
- Keeps the container image straightforward.
- Makes the site portable and easy to deploy.

### Mailjet-hosted mailing list embed

Choice:
- The homepage embeds Mailjet's hosted signup form.

Why:
- The site stays smaller and simpler.
- Mailjet owns the signup UI and confirmation flow directly.
- The Go server does not need newsletter-specific credentials or API glue code.

### Agent-facing text files

Choice:
- Add `llms.txt` and `ssh-help.txt`.

Why:
- Agents often do better with plain text than with stylized HTML alone.
- `ssh-help.txt` gives a normalized command reference that is easy to scrape.
- `llms.txt` gives policy and behavior notes in a clean machine-readable format.

## Known Caveats

- The mailing list section now depends on Mailjet's hosted iframe and embed script being available.
- The rest of the site remains readable and useful even if the embed is blocked.

## Current Implementation Status

Done:
- Drafted the single-page site.
- Added normalized SSH help text.
- Added AI-agent helper text.
- Added a small Go server.
- Added Podman container scaffolding.

In progress:
- Compile and smoke-test the Go server.
- Build and verify the Podman image.

Next step queued:
- Fix any compile/runtime issues found during `go test`, then perform a quick local HTTP smoke test.

## Verification Notes

### 2026-04-20 :: compile pass

Command:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`

Result:
- Pass

Issue found just before the pass:
- `main.go` was missing the `context` import used by the Mailjet subscription helper.

Fix:
- Added the missing `context` import.

Current confidence:
- The Go server compiles.
- The next useful check is a live local HTTP smoke test against `/`, `/llms.txt`, `/ssh-help.txt`, and `/healthz`.

### 2026-04-20 :: local HTTP smoke test

Server boot command:
- `GOCACHE=/tmp/killdeer-go-cache PORT=8080 go run .`

Route checks:
- `GET /` -> `200 OK`
- `GET /llms.txt` -> `200 OK`
- `GET /ssh-help.txt` -> `200 OK`
- `GET /healthz` -> `200 OK`

What this confirms:
- The embedded static files are being served correctly.
- The content headers are in place.
- The `llms.txt` and normalized SSH help files are reachable as plain text.
- The HTML shell renders as one self-contained page, which keeps the site small and deployable.

Small environment note:
- Because the dev server had to bind outside the sandbox, the follow-up curl checks also had to run with the same elevated scope.

Next step queued:
- Check `/subscribe` behavior in the "Mailjet not configured yet" state, then build the Podman image.

### 2026-04-20 :: signup endpoint behavior without Mailjet config

Request:
- `POST /subscribe` with a valid test email and consent set to `true`

Observed response:
- `503 Service Unavailable`
- JSON message: `The mailing list is not configured yet. Email machines@plover.digital in the meantime.`

Why this is good:
- The page has a graceful failure mode before secrets are wired in.
- Users still get a concrete path forward instead of a dead button.
- We can deploy the site shell before the Mailjet environment variables are present and still behave honestly.

Next step queued:
- Build the Podman image and confirm the containerized app compiles cleanly.

### 2026-04-20 :: first Podman build attempt

Command:
- `podman build -t killdeer-digital .`

Observed result:
- Failed before the build graph started.
- Podman could not connect to its machine/socket.

Error summary:
- `unable to connect to Podman socket`
- Suggested follow-up from Podman: `podman machine start`

Interpretation:
- The container recipe still needs verification.
- The current blocker is environment state, not necessarily a problem in `Containerfile`.

Next step queued:
- Start the Podman machine, then rerun `podman build -t killdeer-digital .`.

### 2026-04-20 :: Podman machine startup

Command:
- `podman machine start`

Observed result:
- Success
- Podman reported the default machine started in rootless mode.

What matters for this project:
- Rootless mode is fine here because the site listens on port `8080`, not a privileged port.
- The container environment is ready for a real image build attempt.

Next step queued:
- Rerun `podman build -t killdeer-digital .`.

### 2026-04-20 :: second Podman build attempt

Command:
- `podman build -t killdeer-digital .`

Observed result:
- Failed again with the same socket connection error.

Interesting mismatch:
- `podman machine start` reported success.
- `podman build` still tried to connect to `127.0.0.1:56096` and got `connection refused`.

Interpretation:
- This now looks like a Podman connection/configuration issue rather than a containerfile problem.
- The next thing to inspect is Podman's active connection state and socket configuration.

Next step queued:
- Inspect Podman connection settings, then retry the build against the live machine/socket.

### 2026-04-20 :: Podman connection inspection

Commands:
- `podman system connection list`
- `podman info`

Observed result:
- The default Podman connection is still configured as:
  `ssh://core@127.0.0.1:56096/run/user/501/podman/podman.sock`
- `podman info` still fails because that SSH endpoint refuses the connection.

What this means:
- The machine startup message alone is not enough.
- Either the VM did not stay reachable on the expected forwarded port, or the active Podman connection metadata is stale.

Next step queued:
- Inspect machine state directly and determine whether the VM is actually up.

### 2026-04-20 :: Podman machine state check

Commands:
- `podman machine list`
- `ps -ax`

Observed result:
- `podman machine list` shows:
  - `podman-machine-default*`
  - `LAST UP: Never`
- That does not match the earlier "started successfully" message.

Interpretation:
- Podman's startup messaging and its machine state are disagreeing.
- The VM likely never became usable, which explains the dead SSH socket and repeated build failure.

What this means for the project:
- The site code and Go server have working verification.
- The remaining blocker is local Podman environment health, not the app itself.

Next step queued:
- Probe the machine/socket directly once more and decide whether to keep debugging Podman or mark container verification as blocked by the local environment.

### 2026-04-20 :: direct machine/socket probe

Commands:
- `ls -l /var/folders/ch/jhfxf4q57gqg_hjy5hc6x9b80000gn/T/podman/podman-machine-default-api.sock`
- `podman machine inspect podman-machine-default`

Observed result:
- The advertised API socket path exists on disk.
- `podman machine inspect` reports:
  - `State: stopped`
  - `LastUp: 0001-01-01T00:00:00Z`

Interpretation:
- The socket file alone is not proof of a live machine.
- Podman is leaving stale-looking connection artifacts behind even though the VM is not up.

Decision:
- One final debug-style startup attempt is reasonable.
- If that still leaves the machine in `stopped`, container verification should be marked as blocked by the local Podman environment.

Next step queued:
- Run one final Podman startup check with more detail, then decide whether to stop debugging the environment.

### 2026-04-20 :: debug Podman startup

Command:
- `podman --log-level=debug machine start`

Observed result:
- Much better signal than the earlier plain startup.
- Podman launched `vfkit`.
- The debug log reported:
  - `virtual machine is running`
  - `waiting for VM to stop`

Interpretation:
- The VM can start in this environment.
- The earlier failures may have been a timing or connection bootstrap issue.
- While this debug-start process is holding the VM open, it is worth retrying the image build immediately.

Next step queued:
- Retry `podman build -t killdeer-digital .` while the debug-start session is active.

### 2026-04-20 :: successful Podman image build

Command:
- `podman build -t killdeer-digital .`

Observed result:
- Success
- Image tagged as `localhost/killdeer-digital:latest`

What this confirms:
- The `Containerfile` is valid.
- The multi-stage Go build works in the container.
- The final runtime image is constructed correctly with the compiled binary.

Podman note:
- The build succeeded only after a debug-level machine start kept the VM clearly alive.
- This suggests the earlier failures were due to the local Podman machine/bootstrap state, not the app code.

Next step queued:
- Run the built container and smoke-test the HTTP routes through Podman.

### 2026-04-20 :: first container smoke test

Command:
- `podman run --rm -d -p 18080:8080 --name killdeer-digital-smoke killdeer-digital`

Observed result:
- Container started and returned an ID.
- Follow-up `curl` checks on `http://127.0.0.1:18080` failed:
  - first with `Empty reply from server`
  - then with `Couldn't connect to server`

Interpretation:
- The container likely exited or crashed very quickly after startup.
- The next useful signal is container status + logs.

Next step queued:
- Inspect the temporary container with `podman ps -a` and `podman logs`.

### 2026-04-20 :: container status inspection after first smoke run

Commands:
- `podman ps -a`
- `podman logs killdeer-digital-smoke`

Observed result:
- The smoke-test container was already gone.
- `podman logs` reported no container with that name.

Why:
- The container was launched with `--rm`, so it auto-removed itself after exiting.

Decision:
- Rerun the container without `--rm` so the next failure leaves inspectable logs behind.

Next step queued:
- Start a non-ephemeral smoke-test container and inspect its logs if it exits again.

### 2026-04-20 :: second container smoke test with logs

Commands:
- `podman run -d -p 18080:8080 --name killdeer-digital-smoke killdeer-digital`
- `podman ps -a --filter name=killdeer-digital-smoke`
- `podman logs killdeer-digital-smoke`

Observed result:
- The container exited almost immediately with status `2`.
- Logs show a Go runtime `SIGSEGV` during startup while binding the HTTP listener.

Most likely cause:
- The `Containerfile` hard-coded `GOARCH=amd64`.
- The Podman machine on this Mac is arm64-oriented.
- That mismatch can produce a binary that technically builds but misbehaves at runtime in this environment.

Decision:
- Remove the hard-coded amd64 target and let the Go build use the container's native Linux architecture.

Next step queued:
- Update the `Containerfile`, rebuild the image, and rerun the container smoke test.

### 2026-04-20 :: container architecture fix

Change made:
- Updated `Containerfile`
- Removed the hard-coded `GOARCH=amd64` from the Go build step

Why:
- The image should build for the native Linux architecture inside the Podman machine.
- Hard-coding amd64 made the container less portable and broke runtime verification on this arm64-oriented setup.

New expectation:
- Rebuilt images should launch cleanly in the local Podman VM.

Next step queued:
- Rebuild the image and rerun the container smoke test.

### 2026-04-20 :: rebuilt image after architecture fix

Command:
- `podman build -t killdeer-digital .`

Observed result:
- Success
- New image tagged as `localhost/killdeer-digital:latest`

What this tells us:
- The architecture-friendly build path is valid.
- The next real check is runtime, not compile time.

Next step queued:
- Remove the old exited smoke container, run the rebuilt image, and test the routes again.

### 2026-04-20 :: successful container runtime verification

Commands:
- `podman run -d -p 18080:8080 --name killdeer-digital-smoke-fixed killdeer-digital`
- `podman ps --filter name=killdeer-digital-smoke-fixed`
- `curl -i http://127.0.0.1:18080/healthz`
- `curl -i http://127.0.0.1:18080/llms.txt`

Observed result:
- Container status: `Up`
- `GET /healthz` -> `200 OK`
- `GET /llms.txt` -> `200 OK`

Final conclusion:
- The site runs correctly in Podman after removing the hard-coded amd64 build target.
- The Go server, static assets, agent files, and container recipe are all verified locally.

Next step queued:
- Clean up temporary smoke-test processes and containers.

### 2026-04-20 :: cleanup

Command:
- `podman rm -f killdeer-digital-smoke-fixed killdeer-digital-smoke`

Observed result:
- Both temporary smoke-test containers were removed.

Remaining local cleanup:
- Stop the `go run` development server used for the earlier HTTP smoke tests.

Final cleanup note:
- The temporary `go run` development server was stopped after verification.
- I did not force the broader Podman machine back down, since that is part of your local container environment and not specific to this repo.

### 2026-04-20 :: visual refresh request

New user direction:
- Get the Killdeer bird ASCII art onto the page somehow.
- Push the site closer to the stark black-and-white feel of `plover.digital`.
- Keep it ASCII-forward and fun.

Observed design gap in the current version:
- The first pass leaned warmer and more "paper terminal" than the user now wants.
- It had too much color and too many boxed surfaces compared with the plainness of `plover.digital`.

Design response:
- Strip the palette down to black, white, and gray.
- Reduce the amount of chrome around each section.
- Make the hero feel more like a plaintext announcement than a styled product card.
- Integrate the bird as part of the page identity, not as a side asset.

Next step queued:
- Update `static/index.html` to use a more monochrome, plover-like presentation with the bird in the hero.

### 2026-04-20 :: monochrome bird refresh

Changes made in `static/index.html`:
- Replaced the warm paper/terminal palette with a black, white, and gray palette.
- Removed most of the boxed-card treatment.
- Changed section headings to plaintext-style `-- Section --` labels.
- Added the Killdeer bird art directly into the hero/masthead.
- Kept the command-first flow and copy buttons intact.

Why:
- This gets the page closer to the spare, text-heavy tone of `plover.digital`.
- The bird now feels like part of the site's identity instead of an afterthought.
- The overall look is simpler, sharper, and a little more playful.

Next step queued:
- Verify that the embedded site still compiles and that the served HTML includes the new bird/monochrome markup.

### 2026-04-20 :: verification after monochrome refresh

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `rg -n "Killdeer|-- Access --|Notes For Humans And Agents|Occasional updates" static/index.html`
- `PORT=8081 GOCACHE=/tmp/killdeer-go-cache go run .`
- `curl -s http://127.0.0.1:8081/ | sed -n '300,370p'`

Observed result:
- The Go app still compiles cleanly with the embedded assets.
- The served HTML includes:
  - the Killdeer bird art in the hero
  - the new `-- Access --` section heading
  - the plainer monochrome structure

Conclusion:
- The visual refresh is live in the served page, not just on disk.

Remaining cleanup:
- Stop the temporary `go run` process on port `8081`.

Final cleanup note:
- The temporary verification server on port `8081` was stopped after the served-HTML check.

### 2026-04-20 :: public help redaction request

New user direction:
- Remove admin commands from public view.

Interpretation:
- This applies to the website and other public resources served by it.
- It does not imply changing the live SSH helper on the server in this repo.

Plan:
- Remove the admin command block from `static/index.html`.
- Remove the admin command block from `static/ssh-help.txt`.
- Remove the admin command block from `static/llms.txt`.
- Leave a note in docs so the redaction is intentional rather than accidental drift.

Next step queued:
- Redact admin commands from all public site assets and verify the text no longer appears.

### 2026-04-20 :: admin block redacted from public assets

Changes made:
- Removed the admin command block from `static/index.html`
- Removed the admin command block from `static/ssh-help.txt`
- Removed the admin command block from `static/llms.txt`
- Reworded the agent note in `static/llms.txt` so it explicitly says the public docs omit administrator-only commands

Result:
- The public website and its public text resources no longer enumerate admin-only actions.

Next step queued:
- Verify that no public asset still contains the admin block and that the app still compiles with the updated embedded files.

### 2026-04-20 :: verification after admin redaction

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `rg -n "Admin Commands|admin hosts|admin menu|ordinary customers" static/index.html static/ssh-help.txt static/llms.txt`

Observed result:
- The Go app still compiles cleanly.
- The search returned no matches in the public assets.

Conclusion:
- Admin-only commands are no longer exposed by the website or its public text resources.

### 2026-04-20 :: compose freshness investigation

New user question:
- Why the latest changes are not showing up when running `podman-compose up --build`

Investigation plan:
- Inspect `compose.yaml`
- Check what container/image `podman-compose` is likely running
- Compare that with the known-good direct `podman build` / `podman run` path already verified

Next step queued:
- Inspect the compose setup and current container/image state.

### 2026-04-20 :: compose investigation findings

What I found:
- `podman-compose` is creating and running `killdeer-digital_killdeer-site_1`
- The compose container logs show the site *is* starting and serving requests
- The app currently sends `Cache-Control: public, max-age=300` for the embedded files it serves

Interpretation:
- This is not primarily a "compose failed to rebuild" problem.
- A browser can legitimately keep showing a stale `index.html` for up to five minutes after a rebuild.
- That makes a tiny static-looking site feel like `up --build` did nothing even when the new container is serving traffic.

Decision:
- Change the public embedded assets to no-cache / no-store behavior so rebuilds are visible immediately.

Next step queued:
- Update the cache headers in `main.go`, then verify the new response headers.

### 2026-04-20 :: cache header fix for fresh rebuilds

Change made:
- Updated `main.go`
- Changed embedded asset responses from `Cache-Control: public, max-age=300` to `Cache-Control: no-store, max-age=0`

Why:
- The homepage and public text assets are tiny and change during active development.
- Freshness matters more than caching here.
- This removes the five-minute window where a browser can make a rebuild look stale.

Next step queued:
- Verify the app still compiles and that the served response headers now disable caching.

### 2026-04-20 :: verification after cache fix

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `PORT=8080 GOCACHE=/tmp/killdeer-go-cache go run .`
- `curl -i http://127.0.0.1:8080/`

Observed result:
- The Go app still compiles cleanly.
- The homepage now returns:
  - `Cache-Control: no-store, max-age=0`

Conclusion:
- A browser should no longer sit on a stale homepage for five minutes after a rebuild.

Remaining cleanup:
- Stop the temporary verification server on port `8080`.

Final cleanup note:
- The temporary verification server on port `8080` was stopped after the header check.

### 2026-04-20 :: live compose container mismatch check

New user report:
- `killdeer-digital_killdeer-site_1` is still up
- The user is still seeing the wrong site from that running container

Investigation plan:
- Inspect the running compose container directly
- Fetch the HTML currently being served from port `8080`
- Compare the served output with the current repository version

Next step queued:
- Inspect `killdeer-digital_killdeer-site_1` and fetch its live response.

### 2026-04-20 :: live compose container findings

What the running container is serving:
- The live response on `http://127.0.0.1:8080/` is the *older* warm paper-terminal version
- It still includes:
  - `Cache-Control: public, max-age=300`
  - the earlier warm palette
  - the pre-redaction admin block

Important image detail:
- The compose container is running:
  - `localhost/killdeer-digital_killdeer-site:latest`
- The direct manual verification earlier used:
  - `localhost/killdeer-digital:latest`

Interpretation:
- There are currently two different local images in play.
- `podman-compose` is using its own auto-generated image tag, which has drifted away from the manually rebuilt/tested image.

Decision:
- Make the compose service use an explicit image tag so compose and direct Podman runs point at the same image name.

Next step queued:
- Update `compose.yaml` to use `image: killdeer-digital:latest`, then document the operational note.

### 2026-04-20 :: compose image alignment fix

Changes made:
- Updated `compose.yaml`
- Added `image: killdeer-digital:latest` to the `killdeer-site` service
- Updated `README.md` with a note about `podman-compose up --build --force-recreate`

Why:
- This removes the split between:
  - compose's auto-generated `killdeer-digital_killdeer-site:latest`
  - manual `killdeer-digital:latest`
- It makes compose and direct Podman runs refer to the same named image.

Expected outcome:
- Future compose rebuilds should stop drifting onto a separate stale image tag.

Next step queued:
- Verify the config change on disk and summarize the exact command the user should run next.

### 2026-04-20 :: sizes output addition request

New user direction:
- Add the output of `ssh killdeer.digital sizes`

Interpretation:
- This should be part of the public documentation surface, not just a private note.
- It belongs on the website and in the public text resources that help humans and agents.

Plan:
- Add a visible sizes/pricing section to `static/index.html`
- Add a plain-text `static/sizes.txt`
- Mention that resource from `static/llms.txt`

Next step queued:
- Update the public site and text assets with the `sizes` command output.

### 2026-04-20 :: sizes output added to public docs

Changes made:
- Added `static/sizes.txt`
- Updated `main.go` to embed and serve `/sizes.txt`
- Updated `static/index.html` with:
  - a link to `/sizes.txt`
  - a visible `-- Sizes --` section
- Updated `static/llms.txt` with:
  - a link to the new plain-text sizes resource
  - the sizes output and pricing notes
- Updated `README.md` to mention the new resource

Result:
- Humans and agents now have a dedicated public plain-text pricing resource.
- The homepage also shows the sizes output in a readable preformatted block.

Next step queued:
- Verify the app still compiles and that the new `/sizes.txt` route is present in the served content set.

### 2026-04-20 :: verification after sizes addition

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `GOCACHE=/tmp/killdeer-go-cache PORT=8083 go run .`
- `curl -i http://127.0.0.1:8083/sizes.txt`
- `curl -s http://127.0.0.1:8083/ | rg -n "Sizes and pricing|-- Sizes --|Micro     1 vCPU"`

Observed result:
- The Go app still compiles cleanly.
- `/sizes.txt` returns `200 OK`
- The homepage now includes:
  - the `/sizes.txt` link
  - the `-- Sizes --` section
  - the expected size/pricing rows

Conclusion:
- The sizes output is now part of both the public HTML page and the public plain-text docs set.

Remaining cleanup:
- Stop the temporary verification server on port `8083`.

Final cleanup note:
- The temporary verification server on port `8083` was stopped after the sizes checks.

### 2026-04-20 :: design target shift to isitagentready.com

New user direction:
- Align the site as much as possible with `https://isitagentready.com/`

Interpretation:
- Keep the Killdeer content and SSH-first product story
- Shift the presentation toward the pacing, structure, and feel of that reference
- Preserve the minimal / text-forward constraint while making the layout feel more intentional

Next step queued:
- Inspect the reference site and then adapt `static/index.html` to match its design language more closely.

### 2026-04-20 :: reference-aligned homepage redesign

Changes made in `static/index.html`:
- Reworked the page into a cleaner app-like structure inspired by `isitagentready.com`
- Added a header with a compact brand area and a single outbound action
- Changed the hero into:
  - a strong headline
  - a short explanatory subtitle
  - the Killdeer bird art as a companion panel
- Replaced the older stacked sections with:
  - one primary action panel
  - pill links to public docs
  - accordion-style information cards
  - a cleaner mailing-list panel

Content mapping to the reference:
- Reference hero -> Killdeer SSH-first hero
- Reference scanner section -> primary SSH command / help / sizes action panel
- Reference FAQ cards -> Killdeer usage, sizes, onboarding, and docs accordions

Small implementation detail:
- The copy button behavior was updated so each button restores its original label after showing `Copied`.

Next step queued:
- Verify the embedded app still compiles and inspect the served HTML to confirm the new reference-aligned structure is present.

### 2026-04-20 :: verification after reference-aligned redesign

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `GOCACHE=/tmp/killdeer-go-cache PORT=8084 go run .`
- `curl -s http://127.0.0.1:8084/ | rg -n "Access your VMs over SSH|Primary action|What can you do from SSH|What sizes are available|Get occasional updates"`
- `curl -i http://127.0.0.1:8084/`

Observed result:
- The Go app still compiles cleanly.
- The served HTML now includes the new:
  - hero title
  - primary action panel
  - accordion-style info cards
  - signup section
- The homepage still returns `Cache-Control: no-store, max-age=0`

Conclusion:
- The embedded app is serving the new `isitagentready.com`-influenced structure, not just storing it on disk.

Remaining cleanup:
- Stop the temporary verification server on port `8084`.

Final cleanup note:
- The temporary verification server on port `8084` was stopped after the redesign checks.

### 2026-04-20 :: homepage pricing visibility request

New user direction:
- Put the sizes and pricing on the homepage

Interpretation:
- The pricing should be directly visible on page load
- It should not require expanding an accordion to discover the current VM lineup

Plan:
- Add a visible homepage section for sizes and pricing
- Keep `/sizes.txt` as the plain-text source of truth
- Optionally leave a pricing FAQ item in place, but the core pricing table should be visible by default

Next step queued:
- Update `static/index.html` so the sizes/pricing table is visible on the homepage.

### 2026-04-20 :: pricing promoted into main homepage flow

Changes made:
- Added a visible `Sizes and pricing` card directly under the primary action section in `static/index.html`
- Removed the old accordion item that hid the pricing table behind `What sizes are available?`

Result:
- The VM lineup and pricing are now visible on page load
- The homepage no longer requires an accordion click to compare sizes

Next step queued:
- Verify the homepage still compiles and that the visible pricing section is present in the served HTML.

### 2026-04-20 :: design simplification request

New user direction:
- Roll back the homepage complexity
- Make it feel closer to `plover.digital` again
- Keep it ASCII, simple, and fun

Interpretation:
- The recent `isitagentready.com`-style structure went too far
- The page should feel more like a direct plaintext landing page than a little web app
- Keep the useful public content:
  - SSH access command
  - help/docs links
  - sizes/pricing visible on the homepage

Plan:
- Remove the app-like header / pills / accordion treatment
- Return to a more linear, monochrome, text-first layout
- Keep the bird, the public docs links, and the visible pricing table

Next step queued:
- Rewrite `static/index.html` into a plainer ASCII-first layout with visible pricing.

### 2026-04-20 :: simplified homepage rollback

Changes made in `static/index.html`:
- Removed the app-like header, pills, and accordion treatment
- Returned to a plain single-column document with simple horizontal section breaks
- Kept:
  - the bird art
  - the SSH login command
  - public docs links
  - visible sizes/pricing
  - the mailing-list form
- Reframed the whole page as a small public note rather than a product UI

Tone/design effect:
- Much closer to `plover.digital`
- More monochrome
- More ASCII-forward
- Less "tool" and more "homepage note"

Next step queued:
- Verify the simplified homepage compiles and that the served HTML reflects the rollback.

### 2026-04-20 :: screenshot-guided design correction

New user feedback:
- The earlier simple version shown in the screenshot was much closer
- The site should return toward that exact feel

Interpretation:
- Use the screenshot as the local target
- Keep the monochrome, text-first, single-column structure from that earlier version
- Keep the useful additions like visible sizes/pricing, but fit them into that older layout

Plan:
- Rebuild `static/index.html` around the screenshot's structure:
  - topline
  - short secondary line
  - bird + title block
  - linear sections with horizontal rules
  - simple copy buttons
  - visible pricing in the same style

Next step queued:
- Rewrite `static/index.html` to match the screenshot-guided simpler layout.

### 2026-04-20 :: screenshot-aligned homepage rebuild

Changes made in `static/index.html`:
- Restored the earlier topline + subline structure
- Kept the bird as the central hero element
- Returned to simple linear sections separated by horizontal rules
- Restored the older style of small copy buttons beside commands
- Kept the newer `Sizes And Pricing` section visible on the homepage
- Removed the more recent app-like sectioning and generic footer tag line

Result:
- The page should now feel much closer to the screenshot the user shared
- The content additions remain, but the presentation is back in the simple plover-style lane

Next step queued:
- Verify the rebuilt homepage compiles and that the served HTML reflects the screenshot-guided version.

### 2026-04-20 :: verification after screenshot-guided rebuild

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `GOCACHE=/tmp/killdeer-go-cache PORT=8086 go run .`
- `curl -s http://127.0.0.1:8086/ | rg -n "text first\\. ssh first\\. web page second\\.|-- Access --|-- Sizes And Pricing --|-- Notes For Humans And Agents --|-- Mailing List --"`
- `curl -i http://127.0.0.1:8086/`

Observed result:
- The Go app still compiles cleanly.
- The served homepage includes:
  - the topline + subline
  - the simpler `-- Access --` section
  - a visible `-- Sizes And Pricing --` section
  - the notes, help, and mailing-list sections in the older linear layout
- The homepage still returns `Cache-Control: no-store, max-age=0`

Conclusion:
- The served page is back in the lane shown by the user's screenshot, while keeping visible pricing.

Remaining cleanup:
- Stop the temporary verification server on port `8086`.

### 2026-04-21 :: switch mailing list to Mailjet embed

New user direction:
- Replace the custom signup form with the Mailjet-hosted embed snippet

Interpretation:
- The homepage should stop pretending to be a tiny app for newsletter signup
- Mailjet can own the signup UI and submission flow directly
- The site stays smaller and the architecture gets simpler

Plan:
- Remove the custom `/subscribe` flow from the public page
- Update the homepage mailing list section to embed the Mailjet form
- Rewrite the markdown and agent-facing docs so they describe the hosted embed honestly
- Simplify the Go server and README to match the new setup

Why:
- The embed is the source of truth the user wants to use
- It removes custom signup JavaScript and server-side Mailjet plumbing
- The page becomes more honest, smaller, and easier to maintain

Next step queued:
- Patch the homepage and project docs to reflect the new Mailjet-hosted signup flow.

### 2026-04-21 :: Mailjet embed implementation

Changes made:
- Replaced the custom mailing list form in `static/index.html` with the Mailjet-hosted iframe and script snippet
- Removed the homepage signup submission JavaScript and kept only the small copy-button behavior
- Updated `static/index.md` and `static/llms-full.txt` so they describe the hosted embed honestly
- Updated `README.md`, `.env.example`, and `compose.yaml` to remove stale Mailjet API configuration guidance
- Updated the architecture notes at the top of this log so the current system shape is accurate again

Why:
- The page should match the real signup path the user wants to run
- Removing the custom newsletter flow keeps the project smaller and easier to reason about
- Cleaning up the container and environment docs prevents future false leads during setup

Next step queued:
- Run `gofmt`, `go test`, and a local HTTP smoke test to confirm the simpler mailing-list setup serves correctly.

### 2026-04-21 :: verification after Mailjet embed switch

Commands:
- `gofmt -w main.go`
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `GOCACHE=/tmp/killdeer-go-cache PORT=8090 go run .`
- `curl -i http://127.0.0.1:8090/`
- `curl -i http://127.0.0.1:8090/index.md`
- `curl -i http://127.0.0.1:8090/llms-full.txt`

Observed result:
- The Go app still compiles cleanly after removing the custom Mailjet proxy flow.
- The homepage now serves the Mailjet iframe snippet and external Mailjet embed script.
- The HTML no longer exposes the old custom signup form or `/subscribe` submission flow.
- The CSP now explicitly allows the Mailjet embed origins needed by the hosted form.
- The markdown homepage and full plain-text bundle now describe the mailing list as a Mailjet-hosted embed.
- `.env.example` and `compose.yaml` no longer suggest unused Mailjet API secrets.

Conclusion:
- The mailing list architecture is now simpler, smaller, and documented consistently across the site and project docs.

Remaining cleanup:
- Stop the temporary verification server on port `8090`.

Final cleanup note:
- The temporary verification server on port `8090` was stopped after the homepage, markdown page, and full plain-text bundle checks completed.

### 2026-04-21 :: add Discord contact link

New user direction:
- Add the Discord invite link alongside the existing support email

Plan:
- Update the homepage contact copy to show both email and Discord
- Mirror the same support path in the markdown and agent-facing text

Why:
- The site should offer a fast community contact path without replacing the direct email path
- Keeping the support links consistent across human and agent docs reduces ambiguity

Next step queued:
- Patch the homepage and text resources, then do a quick content check.

### 2026-04-21 :: verification after Discord contact update

Checks:
- Confirmed the homepage now lists both `machines@plover.digital` and `discord.gg/AhD77Raqru`
- Confirmed the Mailjet support fallback copy now offers Discord as well as email
- Confirmed the markdown homepage and full plain-text bundle include the Discord support path

Conclusion:
- Humans and agents now see the same contact options, with both direct email and Discord available.

### 2026-04-21 :: move Mailjet signup into modal

New user direction:
- Move the Mailjet signup into a modal like the older `plover.digital` site

Reference studied:
- `/Users/wokuno/Downloads/index.html`

Observed pattern in the older site:
- The page stays text-first and uncluttered
- A single inline link opens the signup flow
- The Mailjet form lives inside a centered overlay
- The open/close JavaScript is tiny and direct

Plan:
- Replace the always-visible Mailjet embed block with a text trigger
- Add a simple monochrome modal wrapper around the Mailjet iframe
- Keep the implementation inline and tiny, matching the rest of the site
- Preserve keyboard and click-away closing so the modal feels calm instead of sticky

Why:
- The homepage gets quieter again
- The mailing list stays available without taking up a whole section visually
- The interaction will feel closer to the older Plover site the user liked

Next step queued:
- Patch the homepage modal markup, styles, and script, then run a quick local smoke check.

### 2026-04-21 :: Mailjet modal implementation

Changes made:
- Replaced the always-visible Mailjet signup block in `static/index.html` with a small text trigger
- Added a centered monochrome modal overlay that contains the Mailjet iframe
- Added tiny inline JavaScript for open, close, click-away, and `Escape` behavior
- Updated `static/index.md` and `static/llms-full.txt` so they describe the modal-based signup flow accurately

Why:
- The page gets closer to the older `plover.digital` pattern the user referenced
- The mailing list remains available without permanently taking over the section
- The interaction stays tiny, text-first, and easy to maintain

Next step queued:
- Run a quick local HTTP check to confirm the modal trigger, modal shell, and Mailjet iframe are all present in the served homepage.

### 2026-04-21 :: plover CSS modal reference

Additional reference studied:
- `/Users/wokuno/Downloads/index.css`

Useful cues from that stylesheet:
- The modal wrapper is very plain and full-screen
- The close affordance sits in the top-right and stays visually small
- The modal treatment is flatter than a modern app dialog

Adjustment to make:
- Reduce the current Killdeer modal card feel
- Keep the overlay sparse and monochrome
- Let the interaction feel more like a simple site interrupt than a product UI component

Next step queued:
- Simplify the modal styling in `static/index.html`, then re-run the local smoke check.

### 2026-04-21 :: modal styling adjustment from plover CSS

Changes made:
- Flattened the modal overlay styling so it feels more like a simple page interrupt
- Moved the close control into the top-right edge of the modal
- Removed the heavier card feel from the modal shell
- Added a very light monochrome texture to the overlay instead of a modern app-style backdrop

Why:
- The older `plover.digital` modal feels plain and direct
- The flatter treatment fits the Killdeer page better than a more polished dialog card

Next step queued:
- Restart the temporary local server so the embedded static homepage picks up the new modal styling, then verify the served HTML.

### 2026-04-21 :: verification after modal conversion

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `GOCACHE=/tmp/killdeer-go-cache PORT=8090 go run .`
- `curl -i http://127.0.0.1:8090/`

Observed result:
- The Go app still compiles cleanly.
- The served homepage now includes a text trigger for the mailing list instead of an always-visible Mailjet block.
- The served HTML includes the modal wrapper, modal shell, top-right close control, and Mailjet iframe.
- The flatter overlay styling is now present in the served CSS, which moves the interaction closer to the older `plover.digital` feel.

Conclusion:
- The Mailjet signup is now modal-driven and visually calmer, while staying monochrome and text-first.

Remaining cleanup:
- Stop the temporary verification server on port `8090`.

### 2026-04-21 :: broader plover site reference available

New user direction:
- The rest of the `Plover-Digital-LLC.github.io` site is available for reference if useful

Plan:
- Inspect that project for modal-related assets and patterns
- Reuse ideas only where they fit the simpler Killdeer page

Why:
- The old site may have small visual details worth borrowing
- Looking at the real source is better than guessing from memory

Next step queued:
- Review the Plover site files related to the modal and support visuals.

### 2026-04-21 :: findings from broader plover reference

Files confirmed:
- `/Users/wokuno/Desktop/Plover-Digital-LLC.github.io/index.html`
- `/Users/wokuno/Desktop/Plover-Digital-LLC.github.io/index.css`
- `/Users/wokuno/Desktop/Plover-Digital-LLC.github.io/dither.png`
- `/Users/wokuno/Desktop/Plover-Digital-LLC.github.io/close_black_24dp.svg`

Useful references found:
- The site uses a small monospace contact box for email and Discord
- The old modal assets exist if a more literal port is ever wanted
- The broader site still keeps everything text-first and sparse

Decision:
- Keep the current Killdeer implementation lightweight for now
- Borrow the feel, not a full asset copy, unless the user asks for a more literal match

Final cleanup note:
- The temporary verification server on port `8090` was stopped after the modal smoke check.

### 2026-04-21 :: dedicated get in touch section

New user direction:
- Give the email and Discord links their own contact section

Plan:
- Remove the direct contact lines from the mixed notes section
- Add a small standalone support section for email and Discord
- Mirror the structure in the markdown homepage and agent bundle where useful

Why:
- Contact info reads more clearly when it has its own home
- The notes section can stay focused on machine-readable docs and control-plane references

Next step queued:
- Patch the homepage and text resources, then do a quick content verification.

### 2026-04-21 :: verification after contact section split

Checks:
- Confirmed the homepage now has a dedicated `-- Get In Touch --` section
- Confirmed email and Discord were removed from the mixed notes section
- Confirmed the markdown homepage and full agent bundle now expose contact info in a standalone contact block

Conclusion:
- Contact info is easier to scan now, and the notes section stays focused on product and agent guidance.

### 2026-04-21 :: add Caddy in front of the app

New user direction:
- Put Caddy in front of the Go site in the compose setup

Plan:
- Add a tiny `Caddyfile` that reverse proxies to the Go app container
- Update `compose.yaml` so Caddy is the public-facing service
- Keep the Go app on the internal compose network only
- Update the README and architecture notes to reflect the two-container layout

Why:
- Caddy makes the front-door story cleaner
- It gives the project a straightforward place for future routing, headers, or TLS work
- The Go app can stay focused on serving the site itself

Next step queued:
- Patch the compose setup, add the `Caddyfile`, and refresh the project docs.

### 2026-04-21 :: Caddy front end implementation

Changes made:
- Added a minimal `Caddyfile` that reverse proxies to `killdeer-site:8080`
- Updated `compose.yaml` so `caddy` is the public-facing service
- Changed `killdeer-site` from a published host port to internal `expose`
- Updated `README.md` and this build log to describe the two-container stack

Why:
- The compose stack now has a clearer front door
- The app container no longer needs to be published directly to the host
- Future routing or header changes have a natural home in Caddy

Next step queued:
- Sanity-check the config and confirm the compose file still renders correctly.

### 2026-04-21 :: verification after Caddy compose change

Checks:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...` passed
- `podman-compose config` rendered the updated stack successfully
- Confirmed `caddy` publishes host port `8080` to container port `80`
- Confirmed `killdeer-site` is only exposed internally on port `8080`

Conclusion:
- The compose stack is now set up with Caddy in front of the Go app, and the config validates cleanly.

### 2026-04-21 :: billing subdomain redirect

New user direction:
- Redirect `billing.killdeer.digital` to the Stripe billing login URL at the Caddy layer

Plan:
- Add a host-specific Caddy site block for `billing.killdeer.digital`
- Keep the main catch-all site block proxying to the Go app
- Do a config sanity check after the change

Why:
- Billing belongs at the edge, not in the app
- Caddy is the right place for a simple hostname-based redirect like this

Next step queued:
- Patch the `Caddyfile` and verify the config shape.

### 2026-04-21 :: verification after billing redirect

Checks:
- Confirmed `Caddyfile` now contains a dedicated site block for `billing.killdeer.digital`
- Confirmed that block redirects to `https://billing.stripe.com/p/login/8wM4kaa9S7vb7u0bII`
- Re-ran `podman-compose config` to make sure the compose stack still renders cleanly with the updated mounted `Caddyfile`

Conclusion:
- The billing subdomain redirect is now handled at the Caddy layer and the container config still validates cleanly.

### 2026-04-21 :: make Caddy bind IP configurable through .env

New user direction:
- Put the host bind IP for Caddy in `.env`

Plan:
- Add `BIND_IP` to `.env.example`
- Update `compose.yaml` so the Caddy published port uses `${BIND_IP}`
- Refresh the README so deployment instructions mention the bind IP setting explicitly

Why:
- The compose file stays reusable across machines
- The bind address can change without editing the checked-in compose config
- This is safer and cleaner than hardcoding one host IP into the repo

Next step queued:
- Patch `compose.yaml`, `.env.example`, and the README, then verify the rendered compose config.

### 2026-04-21 :: verification after env-driven bind IP

Checks:
- Confirmed `.env.example` now includes `BIND_IP=0.0.0.0`
- Confirmed `compose.yaml` now publishes Caddy with `${BIND_IP:-0.0.0.0}:${PORT:-8080}:80`
- Re-ran `podman-compose config` and confirmed the rendered default port mapping is `0.0.0.0:8080:80`
- Confirmed the README now documents using `BIND_IP=216.66.77.166` in `.env`

Conclusion:
- The host bind IP is now configurable through `.env`, while keeping a safe default for local development.

### 2026-04-21 :: sedna Caddyfile mount failure

Observed deployment issue:
- On `sedna`, `podman-compose up` starts the Go app but Caddy exits with:
- `Error: reading config from file: open /etc/caddy/Caddyfile: permission denied`

Interpretation:
- The problem is the mounted `Caddyfile`, not the app binary or reverse proxy config itself
- On Podman hosts with SELinux, bind mounts often need relabeling so the container can read them

Plan:
- Update the `Caddyfile` bind mount in `compose.yaml` to use Podman-friendly relabeling
- Refresh the README so the fix is documented for remote hosts like `sedna`

Why:
- This is the most likely cause of a read-only config file being unreadable inside the container
- It keeps the setup simple and avoids asking the user to manually relabel files on every deploy

Next step queued:
- Patch the compose bind mount and document the Podman/SELinux note.

### 2026-04-21 :: verification after Podman mount relabel fix

Checks:
- Confirmed the Caddy volume mount is now `./Caddyfile:/etc/caddy/Caddyfile:ro,Z`
- Re-ran `podman-compose config` and confirmed the updated mount renders correctly
- Added a README note explaining why the `:Z` relabel is present

Conclusion:
- The compose stack now includes the Podman/SELinux relabel fix that should resolve Caddy's `permission denied` error on `sedna`.

### 2026-04-21 :: add HTTPS port and document low-port setup

New user direction:
- Pass the HTTPS port through to Caddy
- Document the system setup needed so Podman can hand ports `80` and `443` to the Caddy container

Assumption:
- The main public hostname served by Caddy is `killdeer.digital`

Plan:
- Publish both HTTP and HTTPS ports for the Caddy container through compose
- Switch the main Caddy site block to the real hostname so Caddy can use automatic HTTPS
- Add host setup notes for rootless Podman and low ports
- Update `.env.example` and `README.md` so the new shape is easy to deploy on `sedna`

Why:
- Exposing `443` only matters if Caddy is actually serving a hostname-based HTTPS site
- Low-port publishing is a host concern, so the repo should document it clearly instead of pretending compose can solve it alone

Next step queued:
- Patch the Caddy config, compose/env files, and deployment docs, then verify the rendered stack.

### 2026-04-21 :: verification after HTTPS port wiring

Checks:
- Confirmed `Caddyfile` now serves the main site from `killdeer.digital, http://killdeer.digital`
- Confirmed `compose.yaml` now publishes both `${HTTP_PORT:-80}:80` and `${HTTPS_PORT:-443}:443`
- Confirmed `.env.example` now defaults to `HTTP_PORT=80` and `HTTPS_PORT=443`
- Re-ran `podman-compose config` and confirmed the rendered default mapping is `0.0.0.0:80:80` and `0.0.0.0:443:443`

Conclusion:
- The repo is now wired for both HTTP and HTTPS on Caddy, with the remaining requirement being host support for low ports on the deployment machine.

### 2026-04-21 :: sync markdown and agent text to updated homepage

New user direction:
- Make sure recent `static/index.html` edits are reflected in the markdown and LLM text files

Observed homepage changes:
- `-- Help Command Output --` now appears immediately after `-- Access --`
- `Join the mailing list` now lives inside `-- Get In Touch --`
- The contact section now includes the modal fallback copy directly
- The closing tiny footer copy is now a short `plover.digital` project blurb

Plan:
- Update `static/index.md` to match the new section order and support copy
- Update `static/llms-full.txt` so the richer plain-text bundle reflects the current homepage structure
- Refresh `static/llms.txt` with any now-relevant contact and mailing-list notes

Why:
- The markdown and agent text should describe the live site, not an older draft of it
- Keeping those companion files aligned helps both crawlers and humans trust the alternate formats

Next step queued:
- Patch the text companions, then run a quick content verification pass.

### 2026-04-21 :: verification after text companion sync

Checks:
- Confirmed `static/index.md` now reflects the homepage's current section order, including early help output and the updated contact-plus-mailing-list flow
- Confirmed `static/llms-full.txt` now describes the mailing list modal from the contact section and includes the current footer note
- Confirmed `static/llms.txt` now includes the current contact paths and modal signup note

Conclusion:
- The markdown and machine-readable text now describe the live homepage instead of an older homepage layout.

### 2026-04-21 :: add plover analytics script

New user direction:
- Add the provided Plover analytics script so the site can track users

Plan:
- Add the script tag to `static/index.html`
- Update the Go server CSP so the analytics script can load
- Record the external dependency in the build log

Why:
- The analytics script is an external browser dependency, so the HTML and CSP need to move together
- Documenting it makes the privacy and architecture story easier to understand later

Next step queued:
- Patch the homepage and CSP, then run a quick source verification pass.

### 2026-04-21 :: verification after analytics integration

Checks:
- Confirmed `static/index.html` now includes the provided analytics script with site id `bf210ad5f698`
- Confirmed `main.go` now allows `https://analytics.plover.digital` in both `script-src` and `connect-src`

Conclusion:
- The analytics script is now present in the homepage and should no longer be blocked by the site's CSP.

### 2026-04-21 :: harden copy button fallback

New user direction:
- The copy button is frequently landing in `copy blocked`

Interpretation:
- The current implementation only uses `navigator.clipboard.writeText()`
- That API often fails on plain HTTP pages or in stricter browser contexts

Plan:
- Add a hidden textarea + `document.execCommand("copy")` fallback
- Only show `copy blocked` if both copy paths fail

Why:
- The homepage is often being tested over plain HTTP during setup
- A small fallback keeps the button useful in more environments without changing the page design

Next step queued:
- Patch the inline homepage script and do a quick source verification pass.

### 2026-04-21 :: verification after copy fallback fix

Checks:
- Confirmed `static/index.html` now includes a `legacyCopy()` fallback using `document.execCommand("copy")`
- Confirmed the copy button now routes through a shared `copyText()` helper instead of relying only on `navigator.clipboard.writeText()`
- Confirmed `copy blocked` is now only shown after both copy paths fail

Conclusion:
- The copy button should now work in more setup-time browser contexts, especially when the site is being served over plain HTTP.

### 2026-04-21 :: persist Caddy certificate state

Observed deployment signal:
- Caddy hit Let's Encrypt rate limits for `killdeer.digital`
- The logs show repeated certificate issuance attempts across container runs

Interpretation:
- Caddy should be persisting its ACME account, certificates, and renewal state between runs
- Without persistent storage, recreating the container can look like a brand-new instance and trigger unnecessary certificate orders

Plan:
- Add persistent volumes for Caddy `/data` and `/config` in `compose.yaml`
- Update the README so the TLS persistence behavior is documented
- Re-render the compose config to confirm the storage wiring

Why:
- This is the standard Caddy deployment pattern
- It avoids repeated certificate issuance attempts and helps stay clear of Let's Encrypt rate limits

Next step queued:
- Patch the compose stack and docs, then verify the rendered config.

### 2026-04-21 :: verification after Caddy volume persistence

Checks:
- Confirmed `compose.yaml` now mounts `caddy-data:/data` and `caddy-config:/config`
- Re-ran `podman-compose config` and confirmed both named volumes render in the stack
- Confirmed the README now explains why those volumes should survive redeploys

Conclusion:
- Caddy's ACME account, certificate cache, and renewal state are now configured to persist between container recreations.

### 2026-04-21 :: promotions section request

New user direction:
- Add a promotions section covering signup and referral credits

Changes made:
- Added a `-- Promotions --` section to `static/index.html` directly after pricing
- Added matching `## Promotions` sections to:
  - `static/index.md`
  - `static/llms-full.txt`

Promotion copy used:
- New users get `$20` credit
- Students get double credit with a valid `.edu` email
- Referrals: `$5` for you, `$5` for them

Why:
- The pricing section is the natural place for incentive details
- Mirroring the same offer into the markdown and full plain-text bundle keeps the public site and agent-facing docs aligned

Next step queued:
- Verify the new promotions content compiles and serves correctly.

### 2026-04-21 :: verification after promotions section

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `GOCACHE=/tmp/killdeer-go-cache PORT=8090 go run .`
- `curl -i http://127.0.0.1:8090/`
- `curl -i http://127.0.0.1:8090/index.md`
- `curl -i http://127.0.0.1:8090/llms-full.txt`

Observed result:
- The Go app still compiles cleanly.
- The served homepage includes the new `-- Promotions --` section after pricing.
- The markdown homepage and the full plain-text agent bundle both include matching `Promotions` sections.

Conclusion:
- The promotions copy is now consistent across the human-facing page and the main machine-readable resources.

Remaining cleanup:
- Stop the temporary verification server on port `8090`.

Final cleanup note:
- The temporary verification server on port `8090` was stopped after the promotions checks completed.

### 2026-04-21 :: switch mailing list to Mailjet embed

New user direction:
- Use the provided Mailjet embedded form for the mailing list

Interpretation:
- Replace the current custom signup form on the homepage with the Mailjet-hosted embed
- Update the docs so they no longer describe the mailing list as a local server proxy flow
- Simplify the server if the custom subscription endpoint is no longer needed

Next step queued:
- Inspect the current homepage and server signup flow, then patch the site to use the Mailjet embed end-to-end.

Final cleanup note:
- No temporary server process was left running after verification.

### 2026-04-20 :: agent-readiness review against Cloudflare guidance

References reviewed:
- `https://isitagentready.com/`
- `https://blog.cloudflare.com/agent-readiness/`
- Cloudflare docs for `robots.txt`, sitemaps, and markdown negotiation

Current repo assessment before changes:
- Good:
  - `llms.txt` already exists
  - public plain-text docs already exist for SSH help and sizes
  - the homepage already contains explicit notes for humans and agents
- Missing easy wins:
  - no `robots.txt`
  - no `sitemap.xml`
  - no `llms-full.txt`
  - no markdown representation of the homepage
  - no `Accept: text/markdown` handling on `/`
  - no discovery `Link` headers advertising markdown or agent-facing resources
- Probably unnecessary for this project right now:
  - MCP server card
  - API catalog
  - OAuth discovery
  - WebMCP
  - commerce protocols

Interpretation:
- `killdeer.digital` is a tiny content site plus a mailing-list form, not a public API platform
- We should implement the discoverability and content-accessibility features that fit a small SSH-first homepage
- We should not pretend the site exposes machine-callable capabilities that it does not actually offer

Planned changes:
- Add `robots.txt` with sitemap discovery and explicit content-signal preferences
- Add `sitemap.xml` for the public resources this site actually serves
- Add `llms-full.txt` as a fuller bundled plain-text resource for agents
- Add `index.md` as a markdown version of the homepage
- Teach `/` to negotiate `text/markdown`
- Add response `Link` headers on the homepage to advertise the markdown and agent-facing resources

Preference decision for content signals:
- Allow `search=yes`
- Allow `ai-input=yes`
- Set `ai-train=no`

Why:
- The site explicitly wants to help agents and humans access public operational docs
- Training rights are a different question from inference and retrieval, so we keep those separate

Next step queued:
- Patch the server and static resources to add the agent-readiness basics without making the site more complex.

### 2026-04-20 :: agent-readiness implementation

Changes made:
- Added `static/robots.txt`
- Added `static/sitemap.xml`
- Added `static/index.md`
- Added `static/llms-full.txt`
- Reshaped `static/llms.txt` into a shorter index-style guide
- Added hidden discovery links in `static/index.html`
- Added visible links for `/index.md` and `/llms-full.txt` in the notes section
- Updated `main.go` to serve:
  - `/index.md`
  - `/llms-full.txt`
  - `/robots.txt`
  - `/sitemap.xml`
- Added markdown content negotiation so `/` serves markdown when the request includes `Accept: text/markdown`
- Added response `Link` headers on the homepage to advertise markdown and agent-facing resources
- Added a `Content-Signal` response header aligned with the robots policy
- Updated `README.md` to document the new agent-facing routes

Policy choices implemented:
- `search=yes`
- `ai-input=yes`
- `ai-train=no`

Why these choices fit the project:
- The site is intentionally public and meant to help agents retrieve operational docs
- The site is not offering training rights by default
- The site remains an SSH-first landing page instead of pretending to be an API product

Scope intentionally left out:
- MCP server card
- API catalog
- OAuth discovery
- WebMCP
- commerce protocols

Why those were skipped:
- `killdeer.digital` does not currently expose a public API or tool endpoint that those discovery standards would honestly describe

Next step queued:
- Verify the new routes, headers, and markdown negotiation behavior.

### 2026-04-20 :: verification after agent-readiness patch

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `GOCACHE=/tmp/killdeer-go-cache PORT=8090 go run .`
- `curl -i http://127.0.0.1:8090/`
- `curl -i -H 'Accept: text/markdown' http://127.0.0.1:8090/`
- `curl -i http://127.0.0.1:8090/robots.txt`
- `curl -i http://127.0.0.1:8090/sitemap.xml`

Observed result:
- The Go app still compiles cleanly.
- `GET /` now returns:
  - `Link` headers advertising `/index.md`, `/llms.txt`, `/llms-full.txt`, `/ssh-help.txt`, `/sizes.txt`, and `/sitemap.xml`
  - `Vary: Accept`
  - `Content-Signal: search=yes, ai-input=yes, ai-train=no`
- `GET /` with `Accept: text/markdown` now returns `Content-Type: text/markdown; charset=utf-8`
- `GET /robots.txt` returns the new crawler policy and sitemap directive
- `GET /sitemap.xml` returns the expected XML document for the public resources

Conclusion:
- The site now covers the most relevant `isitagentready.com` basics for a tiny public content site:
  - discoverability
  - content accessibility for agents
  - explicit content-signal policy
- Richer protocol discovery is still intentionally out of scope because the site does not expose a public API or MCP endpoint

Remaining cleanup:
- Stop the temporary verification server on port `8090`.

Final cleanup note:
- The temporary verification server on port `8090` was stopped after the top-rule and realignment checks completed.

### 2026-04-20 :: exact slant title request

New user input:
- Provided the exact desired output of `figlet -f slant killdeer digital`

Interpretation:
- Use the single combined `killdeer digital` slant banner instead of the separate `KILLDEER` and `DIGITAL` blocks
- Keep the bird merged with the title, but swap the title text itself to the exact user-supplied output

Next step queued:
- Replace the current split slant title with the exact single combined slant banner and verify the hero layout.

### 2026-04-20 :: broader font exploration request

New user direction:
- Look for some other fonts too

Interpretation:
- Do a broader `figlet` comparison before locking in the next hero title treatment
- Favor a short curated set of strong candidates rather than trial-and-error edits in the page

Next step queued:
- Sample additional `figlet` fonts and narrow them down to a few serious candidates for the merged hero.

### 2026-04-20 :: broader figlet shortlist

Additional fonts sampled:
- `standard`
- `block`
- `lean`
- `smslant`
- `small`
- `big`
- `colossal`
- `mini`

Shortlist:
- `standard`
- `big`
- `smslant`
- `block`

Recommendation:
- `standard` is the strongest all-around candidate

Why:
- `standard` stays bold and high-contrast without becoming too wide or gimmicky
- `big` is strong but starts to feel heavier and more billboard-like next to the bird
- `smslant` is elegant and compact, but not as bold as the user has been asking for
- `block` has personality, but it introduces more visual noise and competes with the bird

Current decision state:
- No further hero font swap yet in this step
- Wait for user preference before replacing the current `slant` title with another font

### 2026-04-20 :: showfigfonts exploration request

New user direction:
- Look at `showfigfonts`
- Compare what `killdeer digital` looks like in uppercase and lowercase

Interpretation:
- Use the local figlet toolchain to compare actual font behavior before choosing the next title treatment
- Pay special attention to whether uppercase or lowercase produces a better merged hero feel

Next step queued:
- Inspect `showfigfonts` output and generate uppercase/lowercase samples for the strongest candidate fonts.

Final cleanup note:
- The temporary verification server on port `8090` was stopped after the `slant` title checks completed.

### 2026-04-20 :: top rule removal and hero realignment request

New user direction:
- Remove the section line on top
- Realign the merged bird/title ASCII block

Interpretation:
- Remove the border that currently sits above the page content
- Keep the `slant` title treatment, but adjust its spacing relative to the bird so the composition feels cleaner

Next step queued:
- Patch the top layout rule and tighten the hero ASCII alignment, then verify the result.

### 2026-04-20 :: top rule removal and hero realignment implementation

Changes made in `static/index.html`:
- Removed the top border from `.stack`
- Shifted the `slant` title block left so it sits closer to the bird
- Pulled the lower `DIGITAL` block left as well so the two title rows feel more connected to the bird and each other

Why:
- The extra top rule made the page feel boxed in before the content even started
- The hero title was drifting too far away from the bird, which weakened the merged-poster effect

Next step queued:
- Verify the top-rule removal and the tighter hero alignment in the served page.

### 2026-04-20 :: verification after top-rule removal and realignment

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `GOCACHE=/tmp/killdeer-go-cache PORT=8090 go run .`
- `curl -i http://127.0.0.1:8090/`

Observed result:
- The Go app still compiles cleanly.
- The served homepage no longer has the top border line before the content.
- The `slant` title block is now visibly closer to the bird on both the upper and lower title rows.

Conclusion:
- The page opens more cleanly, and the hero reads as a tighter merged mark.

Remaining cleanup:
- Stop the temporary verification server on port `8090`.

Final cleanup note:
- The temporary verification server on port `8090` was stopped after the big bold title checks completed.

### 2026-04-20 :: figlet exploration available

New user update:
- `figlet` is now installed locally for font exploration

Interpretation:
- We can stop approximating title treatments by hand
- The next useful step is to sample a handful of stronger fonts directly and use those results to improve the merged hero title

Next step queued:
- Inspect the local `figlet` installation, list useful fonts, and generate a short set of big bold candidates for `KILLDEER DIGITAL`.

### 2026-04-20 :: figlet font exploration results

Fonts sampled:
- `slant`
- `big`
- `doom`
- `banner3-D`

Decision:
- Use `slant`

Why:
- It matches the strong, forward-leaning energy the user pointed to from `plover.digital`
- It feels bold without becoming a giant rectangular wall
- It merges more naturally with the bird than the heavier novelty fonts

Next step queued:
- Replace the current hand-built bold title with the actual `figlet -f slant` output for `KILLDEER` and `DIGITAL`.

### 2026-04-20 :: slant title implementation

Changes made in `static/index.html`:
- Replaced the hand-built bold title with the actual `figlet -f slant` output for `KILLDEER` and `DIGITAL`
- Kept the bird merged with the title inside the same composite ASCII block
- Increased the composite font clamp slightly to let the `slant` title read with more confidence

Why:
- This is the closest match to the user's `plover.digital` reference
- Using the real figlet output is stronger and cleaner than approximating the style by hand

Next step queued:
- Verify the `slant` title treatment compiles and renders cleanly inside the merged hero.

### 2026-04-20 :: verification after slant title switch

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `GOCACHE=/tmp/killdeer-go-cache PORT=8090 go run .`
- `curl -i http://127.0.0.1:8090/`

Observed result:
- The Go app still compiles cleanly.
- The served homepage now shows the actual `figlet -f slant` output for `KILLDEER` and `DIGITAL` inside the merged bird/title hero.
- The slanted title is visibly stronger and closer to the `plover.digital` reference than the previous hand-built or lighter treatments.

Conclusion:
- The hero title now has the right bold directional feel, while keeping the same overall poster composition.

Remaining cleanup:
- Stop the temporary verification server on port `8090`.

Final cleanup note:
- No temporary server process was left running after the playful title verification.

### 2026-04-20 :: big bold title request

New user direction:
- Pick a better ASCII art font
- Make it big and bold

Interpretation:
- The current playful title is still too light
- The hero should use a heavier, more assertive title treatment while keeping the merged bird/title composition

Next step queued:
- Replace the current title art with a larger, bolder ASCII wordmark and verify the hero still fits cleanly.

### 2026-04-20 :: big bold title implementation

Changes made in `static/index.html`:
- Replaced the lighter playful title with a larger, bolder slanted ASCII wordmark
- Kept the bird merged with the title inside the same composite ASCII block
- Nudged the composite font sizing slightly larger to support the heavier title without letting it spill too aggressively

Why:
- The user asked for something big and bold
- The slanted blockier title is closer to the strong first-impression energy of `plover.digital`

Next step queued:
- Verify the bolder title treatment compiles and renders cleanly.

### 2026-04-20 :: verification after big bold title pass

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `GOCACHE=/tmp/killdeer-go-cache PORT=8090 go run .`
- `curl -i http://127.0.0.1:8090/`

Observed result:
- The Go app still compiles cleanly.
- The served homepage now shows the larger slanted bold `KILLDEER DIGITAL` title inside the merged bird/title composite.
- The updated sizing keeps the heavier title legible without changing the rest of the page structure.

Conclusion:
- The hero title is now much closer to the "big and bold" direction the user asked for.

Remaining cleanup:
- Stop the temporary verification server on port `8090`.

Final cleanup note:
- The temporary verification server on port `8090` was stopped after the playful title checks completed.

Final cleanup note:
- The temporary verification server on port `8090` was stopped after the hero composition checks completed.

### 2026-04-20 :: title wordmark refinement request

New user direction:
- The current `KILLDEER DIGITAL` title treatment feels off
- The title should have a more fun ASCII design
- The TAAG reference should guide the style exploration

Interpretation:
- Keep the merged bird-plus-title composition
- Change the title art itself, not the overall hero structure
- Favor a title that feels playful and handmade rather than too rigid or blocky

Next step queued:
- Explore a few ASCII wordmark treatments and replace the current title art with a stronger option.

### 2026-04-20 :: playful title wordmark implementation

Changes made in `static/index.html`:
- Replaced the stiffer `KILLDEER DIGITAL` title art inside the composite hero with a more playful line-art wordmark
- Kept the bird merged with the title inside one composite ASCII block
- Reduced the composite hero font clamp and line height slightly so the wider title treatment still fits cleanly

Why:
- The previous title felt too rigid next to the bird
- The new line-art title has more character and feels closer to the handmade ASCII energy the user is asking for

Next step queued:
- Verify the playful title treatment compiles and renders cleanly.

### 2026-04-20 :: verification after playful title refinement

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `GOCACHE=/tmp/killdeer-go-cache PORT=8090 go run .`
- `curl -i http://127.0.0.1:8090/`

Observed result:
- The Go app still compiles cleanly.
- The served homepage now shows the more playful line-art `KILLDEER DIGITAL` treatment inside the merged bird/title composite.
- The smaller font clamp keeps the wider title art inside the composition without changing the rest of the page layout.

Conclusion:
- The hero keeps the merged poster composition the user wanted, but the title has a more lively ASCII voice.

Remaining cleanup:
- Stop the temporary verification server on port `8090`.

Final cleanup note:
- The temporary verification server on port `8090` was stopped after the lighter hero checks completed.

### 2026-04-20 :: ASCII merge cleanup request

New user direction:
- Clean up the merging of the ASCII in the hero

Interpretation:
- The current wordmark and bird combination likely feels visually crowded or poorly separated
- The right move is to improve spacing and alignment inside the hero without changing the broader page structure

Plan:
- Inspect the current hero markup and CSS
- Adjust the relationship between the hero wordmark, bird, and copy so they read as one composition instead of colliding
- Keep the overall monochrome, text-first style

Next step queued:
- Review the current hero structure and patch the ASCII composition.

### 2026-04-20 :: ASCII merge cleanup implementation

Changes made in `static/index.html`:
- Restored the opening topline block as its own section
- Rebuilt the hero so the lighter ASCII wordmark sits in its own `<pre>`
- Rebuilt the lower hero row so the bird and copy sit in separate elements again
- Removed the broken merged markup that had the wordmark, bird, and intro copy colliding inside one malformed block

Why:
- The problem was not just spacing; the HTML structure itself had become tangled
- Separating the three hero pieces again gives the ASCII room to breathe and makes the page predictable to style

Next step queued:
- Verify that the repaired hero compiles and serves clean HTML again.

### 2026-04-20 :: correction on ASCII cleanup direction

New user clarification:
- The user had been intentionally moving the ASCII around because they liked the tighter composition

Correction:
- The right goal is not to fully separate the wordmark and bird
- The right goal is to keep the intertwined composition, but make it structurally clean

Updated plan:
- Rebuild the hero ASCII as one intentional composite block
- Preserve the tighter bird + wordmark relationship
- Avoid malformed HTML by composing the ASCII inside a single `<pre>`

Next step queued:
- Replace the split hero art with a single clean composite ASCII block.

### 2026-04-20 :: hero composition refinement

New user direction:
- Move the opening topline/subline/tiny text below the bird
- Merge the bird and the title wordmark together
- Remove the extra hero copy paragraph block

Interpretation:
- The hero should be a single ASCII composition first
- The smaller descriptive site text should sit beneath that composition, not above or beside it
- The page should open more like a single poster and less like art plus explanatory blurb

Next step queued:
- Restructure the hero so the merged ASCII sits first and the smaller site text follows beneath it.

### 2026-04-20 :: hero composition restructure implementation

Changes made in `static/index.html`:
- Removed the separate opening section that sat above the hero
- Kept the bird and title wordmark merged inside one composite ASCII block
- Moved the smaller topline/subline/tiny text beneath the composite ASCII inside the hero
- Removed the extra hero prose block entirely

Why:
- This makes the top of the page read as one unified composition
- The supporting site text now feels like a caption to the ASCII instead of a competing content block
- Removing the extra copy keeps the opening more direct and more in line with the user's preferred minimal tone

Next step queued:
- Verify the restructured hero compiles and renders cleanly.

### 2026-04-20 :: verification after hero composition restructure

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `GOCACHE=/tmp/killdeer-go-cache PORT=8090 go run .`
- `curl -i http://127.0.0.1:8090/`

Observed result:
- The Go app still compiles cleanly.
- The served homepage now opens with:
  - the merged bird + title ASCII block
  - the topline/subline/tiny site text directly underneath it
  - no extra hero prose block

Conclusion:
- The hero now reads more like a single poster composition, which aligns with the user's direction.

Remaining cleanup:
- Stop the temporary verification server on port `8090`.

Final cleanup note:
- The temporary verification server on port `8090` was stopped after the divider checks completed.

### 2026-04-20 :: alternate hero wordmark request

New user input:
- Provided a new ASCII `KILLDEER DIGITAL` wordmark block

Interpretation:
- Replace the current hero banner text with the new supplied version
- Keep the bird and the simpler page structure from the latest pass

Plan:
- Update `static/index.html` only
- Swap the hero wordmark to the new ASCII block
- Keep the existing bird + short copy layout unless the new mark forces a spacing tweak

Next step queued:
- Patch the hero banner markup with the new ASCII lockup.

### 2026-04-20 :: follow-up ASCII inspiration

New user input:
- Shared several more `KILLDEER DIGITAL` ASCII treatments

Decision:
- Prefer the lighter line-art version over the heavy filled block or fullwidth Unicode variant

Why:
- The site has been moving back toward a simpler `plover.digital` feel
- The lighter mark leaves more breathing room for the bird and the surrounding copy
- Staying with plain ASCII is a better fit than switching to fullwidth Unicode characters

Next step queued:
- Replace the current hero wordmark with the lighter ASCII version and adjust sizing if needed.

### 2026-04-20 :: lighter hero wordmark implementation

Changes made in `static/index.html`:
- Replaced the filled block hero banner with the lighter line-art `KILLDEER DIGITAL` wordmark
- Increased the hero wordmark font clamp slightly so the thinner strokes still read as the main hero element
- Kept the bird, hero copy, and overall hero layout unchanged

Why:
- The lighter banner is more in tune with the small, text-first feeling of the rest of the page
- It gives the hero more character without making it feel loud or billboard-like

Next step queued:
- Verify the lighter hero mark still compiles and renders cleanly.

### 2026-04-20 :: verification after lighter hero wordmark

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `GOCACHE=/tmp/killdeer-go-cache PORT=8090 go run .`
- `curl -i http://127.0.0.1:8090/`

Observed result:
- The Go app still compiles cleanly.
- The served homepage now includes the lighter line-art `KILLDEER DIGITAL` banner.
- The larger font clamp keeps the thinner ASCII strokes readable without overpowering the bird.

Conclusion:
- The hero is now more in line with the site's simpler monochrome tone.

Remaining cleanup:
- Stop the temporary verification server on port `8090`.

Final cleanup note:
- The temporary verification server on port `8090` was stopped after the hero checks completed.

### 2026-04-20 :: section divider cleanup request

New user direction:
- There are too many section lines on the homepage right now

Interpretation:
- The current layout is over-segmented because nearly every section carries its own rule
- The right fix is to reduce the divider count, not to change the overall information architecture

Plan:
- Update `static/index.html` only
- Simplify the divider system so fewer sections draw horizontal rules
- Keep enough structure that the page still scans cleanly

Next step queued:
- Inspect the current section markup and patch the divider treatment.

### 2026-04-20 :: section divider cleanup implementation

Changes made in `static/index.html`:
- Removed the default bottom rule from every section
- Introduced a dedicated `.section-rule` class for the few places that still need a divider
- Kept horizontal rules only on:
  - the hero break
  - the mailing-list break

Why:
- The section titles already provide enough structure for the middle of the page
- Reducing the divider count makes the layout feel calmer and more like a plaintext note
- Keeping only a couple of rules preserves rhythm without making the page feel boxed in

Next step queued:
- Verify the lighter divider treatment still compiles and renders cleanly.

### 2026-04-20 :: verification after divider cleanup

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `GOCACHE=/tmp/killdeer-go-cache PORT=8090 go run .`
- `curl -i http://127.0.0.1:8090/`

Observed result:
- The Go app still compiles cleanly.
- The served homepage now applies horizontal rules only to:
  - the hero section
  - the mailing-list section
- The access, pricing, notes, help, and footer sections now rely on spacing and headings instead of extra divider lines.

Conclusion:
- The page keeps its structure, but the visual rhythm is quieter and closer to a simple plaintext note.

Remaining cleanup:
- Stop the temporary verification server on port `8090`.

Final cleanup note:
- The temporary verification server on port `8090` was stopped after the HTTP checks completed.

### 2026-04-20 :: hero refresh request

New user direction:
- Keep the bird
- Make the hero more interesting
- Use the large ASCII `KILLDEER DIGITAL` wordmark the user provided
- Stay simple, monochrome, and fun

Interpretation:
- The best move is to make the hero feel more like a little terminal poster
- Use the bird as the character element
- Use the big text mark as the main visual anchor
- Keep the rest of the page quiet so the hero does the work

Plan:
- Update `static/index.html` only
- Add a larger ASCII lockup for the hero
- Preserve the bird
- Tighten the hero layout so it still works on smaller screens

Next step queued:
- Inspect the current hero markup and patch the homepage hero section.

### 2026-04-20 :: hero refresh implementation

Changes made in `static/index.html`:
- Added the user's large ASCII `KILLDEER DIGITAL` wordmark as the main hero lockup
- Kept the bird as a second visual anchor under the wordmark
- Reworked the hero into a two-part layout:
  - large ASCII banner
  - bird + short explanatory copy
- Shortened the hero copy so the top of the page reads faster
- Added responsive hero layout rules so the bird and copy stack cleanly on smaller screens

Why:
- The big wordmark gives the homepage a stronger first impression without breaking the text-first style
- The bird keeps the page playful and recognizable
- Shorter copy lets the ASCII art carry more of the mood

Next step queued:
- Verify the refreshed hero still compiles and renders cleanly.

### 2026-04-20 :: verification after hero refresh

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `GOCACHE=/tmp/killdeer-go-cache PORT=8090 go run .`
- `curl -i http://127.0.0.1:8090/`

Observed result:
- The Go app still compiles cleanly.
- The served homepage includes:
  - the large ASCII `KILLDEER DIGITAL` wordmark
  - the preserved bird art
  - the shorter hero copy and metadata line
  - the responsive stacked hero layout rules for smaller screens

Conclusion:
- The hero now has a stronger visual center while staying monochrome, text-first, and playful.

Remaining cleanup:
- Stop the temporary verification server on port `8090`.

Final cleanup note:
- The temporary verification server on port `8086` was stopped after the screenshot-aligned checks.

### 2026-04-20 :: small homepage polish request

New user direction:
- Make the block cursor blink again on the login command
- Remove the scrollbar beside the bird art
- Remove the `=====` line under the bird

Plan:
- Update `static/index.html` only
- Restore the cursor animation
- Tighten the bird container overflow behavior
- Remove the decorative rule made of equals signs from the bird block

Next step queued:
- Patch the homepage markup and CSS, then verify the served HTML reflects the polish changes.

### 2026-04-20 :: homepage polish implementation

Changes made in `static/index.html`:
- Restored the blinking block cursor on the main SSH login command
- Changed the bird art container to hide overflow instead of exposing a scrollbar
- Removed the `=====` rule from the bird block so the hero stays lighter and closer to the earlier version

Why:
- The blinking block makes the command feel live again without making the page noisy
- The bird should read like a single piece of ASCII art, not like a scrollable widget
- Dropping the equals-sign divider keeps the page closer to the tiny plaintext look the user preferred

Next step queued:
- Verify the homepage still compiles and that the served HTML reflects the cursor and bird cleanup.

### 2026-04-20 :: verification after homepage polish

Commands:
- `GOCACHE=/tmp/killdeer-go-cache go test ./...`
- `GOCACHE=/tmp/killdeer-go-cache PORT=8086 go run .`
- `curl -s http://127.0.0.1:8086/ | rg -n "@keyframes blink|ssh \[username\]@killdeer\.digital|cursor|================================================================================"`

Observed result:
- The Go app still compiles cleanly.
- The served homepage includes the blinking cursor markup and animation again.
- The served bird block no longer includes the equals-sign divider.
- The bird section now serves with `overflow: hidden`, which removes the internal scrollbar treatment from that block.

Conclusion:
- The homepage is back to the simpler screenshot-guided look, with the cursor blink restored and the bird hero cleaned up.

Remaining cleanup:
- Stop the temporary verification server on port `8086`.

### 2026-04-29 :: pricing and OS listing refresh

New user direction:
- Update public listings from the latest `ssh killdeer.digital sizes` and `ssh killdeer.digital os` output.

Changes made:
- Updated `static/sizes.txt`, homepage pricing, Markdown docs, JSON size metadata, and the sizing agent skill with the lower runtime rates and new 24/7 estimates.
- Updated `static/os.txt`, homepage OS table, Markdown docs, JSON image metadata, and the CLI agent skill for Debian 13, Fedora 44, Ubuntu 26.04 as the Ubuntu shorthand default, and the new active-image wording.
- Expanded the homepage OS table to show shorthand, image, and OS columns so default shorthands are visible.

Verification:
- `GOCACHE=/private/tmp/killdeer-go-cache go test ./...`
- `jq . static/api/v1/sizes.json`
- `jq . static/api/v1/images.json`
- `curl -s http://127.0.0.1:8090/sizes.txt`
- `curl -s http://127.0.0.1:8090/os.txt`
- `curl -s http://127.0.0.1:8090/api/v1/sizes.json`
- `curl -s http://127.0.0.1:8090/api/v1/images.json`
- `curl -s http://127.0.0.1:8090/`

Observed result:
- The Go app still compiles cleanly.
- The served text routes, JSON metadata routes, and homepage all include the refreshed prices and OS image list.
- A stale-value scan found no remaining old rates, old Debian/Fedora image versions, or old full-image wording.

Cleanup:
- The temporary verification server on port `8090` was stopped.
