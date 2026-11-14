# buttons-mcp

[![CI](https://github.com/Almiro-Studios/buttons-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/Almiro-Studios/buttons-mcp/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Almiro-Studios/buttons-mcp.svg)](https://pkg.go.dev/github.com/Almiro-Studios/buttons-mcp)
[![Go Report Card](https://goreportcard.com/badge/github.com/Almiro-Studios/buttons-mcp)](https://goreportcard.com/report/github.com/Almiro-Studios/buttons-mcp)
[![Release](https://img.shields.io/github/v/release/Almiro-Studios/buttons-mcp)](https://github.com/Almiro-Studios/buttons-mcp/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![MCP](https://img.shields.io/badge/MCP-server-8A2BE2)](https://modelcontextprotocol.io)

An [MCP](https://modelcontextprotocol.io) server for the **[Elgato Stream Deck Studio](https://www.elgato.com/stream-deck-studio)** running **[Bitfocus Buttons](https://bitfocus.io/buttons)**. It lets an AI assistant (Claude, Cursor, …) build and edit your layouts: put live values on keys, colour them by load, and make keys trigger actions (ATEM cuts, routing, streaming).

> **Unofficial.** Buttons has no public API. This server talks to the same internal tRPC API the Buttons web UI uses, so it can break when Buttons updates. Tested with Buttons 1.9.1 (beta).

## Before you start

buttons-mcp does not talk to the Stream Deck directly. It controls **Bitfocus Buttons**, the software that drives the Stream Deck Studio, so you need both:

1. **A Stream Deck Studio.** Elgato's 32-key, 2-dial rack unit. It connects over Ethernet (PoE+) or USB-C. Other surfaces that Buttons supports may work too, but the Studio is what this project is built and tested for.
2. **Bitfocus Buttons, installed and set up:**
   1. Download it from [user.bitfocus.io/download](https://user.bitfocus.io/download) (free account; the free tier is enough) and install it on a computer on the same network as the Studio.
   2. Open the Buttons web UI at `https://localhost:4443` (accept the self-signed certificate warning) and create the first admin user. That username and password are what `buttons-mcp login` asks for.
   3. Add the Studio under **Connect → Surface**. On the network it is discovered automatically (it advertises itself as `Stream Deck Studio …`); you can also enter its IP address.
   4. Create a **Position** (a layout), assign the Studio surface to it, and click **Apply** once. The MCP tools edit positions, so you need at least one.
   5. Add the **connections** (modules) you want to control or read from: ATEM, OBS, vMix, Generic File Reader, … under **Connect → Connection**. You can also ask your assistant to do this later with `call_api`.
3. **The Elgato Stream Deck app is not needed** and should not be paired with the Studio at the same time: Buttons takes control of the device. If the Studio's firmware is too old for Buttons, update it with Bitfocus or Elgato's instructions first.

## Use with your AI app

Each button opens a short setup page: install the binary, run `buttons-mcp login`, add the server (a one-click `cursor://` install for Cursor, a copy-paste command or config for the others), then open the app with a starter prompt prefilled.

[![Add to Cursor](https://img.shields.io/badge/Add%20to-Cursor-000000?style=for-the-badge&logo=cursor&logoColor=white)](https://almiro-studios.github.io/buttons-mcp/open/cursor.html)
[![Use with Claude Code](https://img.shields.io/badge/Use%20with-Claude%20Code-D97757?style=for-the-badge&logo=claude&logoColor=white)](https://almiro-studios.github.io/buttons-mcp/open/claude-code.html)
[![Use with Claude Desktop](https://img.shields.io/badge/Use%20with-Claude%20Desktop-D97757?style=for-the-badge&logo=claude&logoColor=white)](https://almiro-studios.github.io/buttons-mcp/open/claude.html)
[![Use with Codex](https://img.shields.io/badge/Use%20with-Codex-412991?style=for-the-badge&logo=openai&logoColor=white)](https://almiro-studios.github.io/buttons-mcp/open/codex.html)

Already have the binary and login? [Install straight into Cursor](https://cursor.com/install-mcp?name=buttons&config=eyJjb21tYW5kIjogImJ1dHRvbnMtbWNwIn0%3D).

## What it can do

| Tool | What it does |
|---|---|
| `list_positions` | List positions (layouts mapped to surfaces). |
| `list_buttons` | List keys on a position: grid location, text, background, number of colour rules. |
| `get_layout` | Raw layout document and revision, for advanced edits. |
| `set_button` | Create or update the key at `x`,`y`: text, icon ([Tabler](https://tabler.io/icons) name, e.g. `brand-github`), background colour, and an image: a website's own icon (`site_icon`), an `image_url`, or a local `image_file` (PNG, JPEG, GIF or SVG). `style` sets the layout (`bigIcon` fills the key; image-only keys use it by default); `text_color` sets the text colour. |
| `delete_button` | Remove the key at `x`,`y`. |
| `set_color_rules` | Replace a key's colour rules: each rule sets the background when its expression is true. |
| `set_load_gradient` | Colour a key from green to red by a numeric variable, in 2–8 steps. |
| `list_actions` | List the actions each connection offers (switch an ATEM input, route a Videohub, start a stream, …) with option ids, types, defaults and choices. |
| `set_button_actions` | Set what a key does on press and release: connection actions with option values and delays. |
| `list_connections` / `get_connection` | List connections, or show one connection's settings and status. |
| `add_connection` | Add a connection for any installed module (e.g. `bmd-atem`, `obs-studio`, `generic-http`) and set its config. |
| `configure_connection` / `delete_connection` | Change a connection's settings (merged) or label, or remove it. |
| `list_variables` | List connections, to find variables for `$(connection.variable)`. |
| `call_api` | Call any Buttons API procedure (query or mutation). Covers everything the UI can do that the tools above don't: connections, positions, surfaces, presets, … |

The dedicated tools are safe shortcuts for common edits. `call_api` gives full access: procedure names follow the UI, for example `connection.list`, `connection.get`, `connection.add`, `position.all`, `position.create`, `position.update`, `position.delete`, `positionDocument.saved`, `positionDocument.save`. Inputs are plain JSON; most queries take an object (use `{}` when there is nothing to pass).

Every edit loads the saved layout, changes it, and saves it back with the revision it loaded. If someone else saved in between, the save is rejected instead of overwriting their change. Saved changes go live on the surface right away.

Key text uses Buttons variables, written `$(connectionKey.variable)`, for example `CI $(FileReader.contents)`. Colours are Buttons palette entries written `[name, shade]`, for example `["green","500"]` or `["red","700"]`.

## Install

Download a binary for macOS, Linux or Windows from [Releases](https://github.com/Almiro-Studios/buttons-mcp/releases) and put `buttons-mcp` on your `PATH`.

Or build it with Go 1.26 or newer:

```bash
go install github.com/Almiro-Studios/buttons-mcp/cmd/buttons-mcp@latest
```

## Log in once

```bash
buttons-mcp login
```

This asks for your Buttons username and password, checks them against Buttons, and stores them in your system keychain (macOS Keychain, Windows Credential Manager, or Secret Service on Linux). On macOS without a terminal it shows a dialog instead. Your password never goes into an MCP config file.

`buttons-mcp logout` removes the saved login.

For CI or headless machines, set `BUTTONS_USERNAME` and `BUTTONS_PASSWORD` instead; they take precedence over the keychain.

## Add it to your MCP client

Claude Code:

```bash
claude mcp add buttons -- buttons-mcp
```

Claude Desktop, Cursor and others (`mcpServers` JSON):

```json
{
  "mcpServers": {
    "buttons": { "command": "buttons-mcp" }
  }
}
```

Use the full path to the binary if `$(go env GOPATH)/bin` is not on the client's `PATH`.

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `BUTTONS_URL` | `https://localhost:4443` | Buttons web address (a `wss://…/api/wstrpc` URL also works). |
| `BUTTONS_INSECURE_TLS` | `true` | Accept the self-signed certificate Buttons ships with. Set `false` if you use a trusted certificate. |
| `BUTTONS_USERNAME` / `BUTTONS_PASSWORD` | — | Override the keychain login. |

## Walkthrough: a CI dashboard on a Stream Deck Studio

1. Write your numbers to files, e.g. `~/.deck-stats/ci_queued.txt`, from a script on a timer.
2. In Buttons, add a **Generic File Reader** connection per file. Its variable key (e.g. `FileReader`) exposes `$(FileReader.contents)`.
3. Ask your assistant:
   *"On the Desk position, put `CI $(FileReader.contents)` on key 0,0 and colour it green at 0 to red at 20 queued jobs."*

   It calls `set_button`, then `set_load_gradient`.

## Use cases

Things you can ask your assistant once the server is connected. Each needs the matching Buttons connection (module) set up; the assistant can list connections with `list_variables` and add new ones with `call_api` (`connection.add`).

1. **CI and build dashboard.** *"Put the queued GitHub Actions jobs from `$(FileReader.contents)` on key 0,0 and colour it green at 0, red at 20."* Feed numbers from a script into Generic File Reader, Generic HTTP or MQTT connections.
2. **Live stream health (OBS / vMix).** *"Make a row showing OBS streaming and recording state, dropped frames and CPU; turn the dropped-frames key red above 1%."* Uses the `obs-studio` or `studiocoast-vmix` module variables.
3. **ATEM program/preview tally.** *"Build a 2×8 grid for ATEM inputs 1–8: program bus on top, preview below, red when on program and green when on preview."* Uses `bmd-atem` variables in colour rules.
4. **Videohub routing panel.** *"Lay out one key per Videohub output showing which input it is routed from."* Uses `bmd-videohub` routing variables as key text.
5. **YouTube / Twitch live stats.** *"Show concurrent viewers and stream uptime for my YouTube live event and the Twitch channel, with a gradient from grey to green as viewers grow."* Uses `youtube-live` and `twitch-api`.
6. **Audio levels and mutes.** *"Show mute state for X32 channels 1–8 and colour muted channels red."* Uses `behringer-x32` or `allenheath-sq`.
7. **Recording / playback status.** *"Show HyperDeck 1 and 2 transport state and remaining record time; turn the time key orange under 10 minutes and red under 2."* Uses `bmd-hyperdeck`.
8. **Service and network monitoring.** *"Add keys for our production servers that go red when Generic Ping fails, and show the last speed test."* Uses `generic-ping` and `generic-speedtest`.
9. **Show-calling with ProPresenter.** *"Show the current and next slide titles and the countdown timer, red under 30 seconds."* Uses `renewedvision-propresenter-api`.
10. **Website launcher.** *"Put Sketchpad, GitHub and Portainer on the right side using each site's icon, and open the site in Chrome when pressed."* Uses `site_icon` for the faces and a `generic-http` action calling a small local opener service (see [Opening websites from keys](#opening-websites-from-keys)).
11. **Studio environment.** *"Show Key Light brightness, Hue scene and room temperature from Home Assistant, and colour the temperature key blue to red."* Uses `elgato-keylight`, `philips-hue` and `homeassistant-server`.

Keys can both **show** and **do**: `set_button` and the colour tools handle the display, and `set_button_actions` makes a key cut to an ATEM input, route a Videohub output, start a stream or run any other connection action on press or release. The assistant finds the right action and its options with `list_actions`.

## Opening websites from keys

Buttons runs actions on the Buttons server and has no "open a URL in my browser" action. To make a key open a website on your computer, run a tiny local service that opens a fixed list of sites, and call it with a `generic-http` GET action. A minimal example (Python, localhost only, allowlisted names):

```python
import http.server, subprocess
SITES = {"github": "https://github.com", "docs": "https://example.com/docs"}
class H(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        url = SITES.get(self.path.removeprefix("/open/"))
        if url: subprocess.Popen(["open", url])  # macOS; use xdg-open on Linux
        self.send_response(204 if url else 404); self.end_headers()
http.server.HTTPServer(("127.0.0.1", 8765), H).serve_forever()
```

Then ask your assistant: *"Add a generic-http connection called opener, and make key 15,0 do GET http://127.0.0.1:8765/open/github with GitHub's icon."*

`site_icon` reads the icon a site declares (apple-touch-icon, PNG or SVG `<link rel=icon>`) and falls back to Google's public favicon service, which receives the site's domain.

## How it works

- Login: `POST /api/trpc/auth.localLogin` returns a session cookie.
- Everything else goes over the WebSocket at `/api/wstrpc` (tRPC, superjson envelope), carrying that cookie. If the session expires, the server logs in again and retries once.
- Layouts: `positionDocument.saved` (load) and `positionDocument.save` (save with `expectedRevision`).
- Colour rules are Internal → Expression feedbacks whose expression is stored as `_EXP_(…)`.

Debugging aid: `buttons-mcp dump -` prints all positions; `buttons-mcp dump <positionId>` prints one saved layout.

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md), and please follow the [code of conduct](CODE_OF_CONDUCT.md). Security issues: see [SECURITY.md](SECURITY.md). Release notes: [CHANGELOG.md](CHANGELOG.md).

Not affiliated with or endorsed by Bitfocus. "Bitfocus Buttons" and "Stream Deck" are trademarks of their owners.

## License

[MIT](LICENSE) © 2026 Almiro Studios
