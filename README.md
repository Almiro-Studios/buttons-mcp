# buttons-mcp

[![CI](https://github.com/Almiro-Studios/buttons-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/Almiro-Studios/buttons-mcp/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Almiro-Studios/buttons-mcp.svg)](https://pkg.go.dev/github.com/Almiro-Studios/buttons-mcp)
[![Go Report Card](https://goreportcard.com/badge/github.com/Almiro-Studios/buttons-mcp)](https://goreportcard.com/report/github.com/Almiro-Studios/buttons-mcp)
[![Release](https://img.shields.io/github/v/release/Almiro-Studios/buttons-mcp)](https://github.com/Almiro-Studios/buttons-mcp/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![MCP](https://img.shields.io/badge/MCP-server-8A2BE2)](https://modelcontextprotocol.io)

An [MCP](https://modelcontextprotocol.io) server for [Bitfocus Buttons](https://bitfocus.io/buttons), the control software for Stream Deck Studio and other surfaces. It lets an AI assistant (Claude, Cursor, …) build and edit your layouts: put live values on keys, and colour them by load.

> **Unofficial.** Buttons has no public API. This server talks to the same internal tRPC API the Buttons web UI uses, so it can break when Buttons updates. Tested with Buttons 1.9.1 (beta).

## Use with your AI app

One click opens the app with a setup prompt prefilled (you review it before it runs). It installs the server, asks you to run `buttons-mcp login`, adds the MCP config, and shows your layout.

[![Use with Cursor](https://img.shields.io/badge/Use%20with-Cursor-000000?logo=cursor&logoColor=white)](https://almiro-studios.github.io/buttons-mcp/open/cursor.html)
[![Add to Cursor](https://cursor.com/deeplink/mcp-install-dark.svg)](https://almiro-studios.github.io/buttons-mcp/open/cursor-mcp.html)
[![Use with Claude Code](https://img.shields.io/badge/Use%20with-Claude%20Code-D97757?logo=claude&logoColor=white)](https://almiro-studios.github.io/buttons-mcp/open/claude-code.html)
[![Use with Claude Desktop](https://img.shields.io/badge/Use%20with-Claude%20Desktop-D97757?logo=claude&logoColor=white)](https://almiro-studios.github.io/buttons-mcp/open/claude.html)
[![Use with Codex](https://img.shields.io/badge/Use%20with-Codex-412991?logo=openai&logoColor=white)](https://almiro-studios.github.io/buttons-mcp/open/codex.html)

"Add to Cursor" only registers the server; install the binary and run `buttons-mcp login` first.

## What it can do

| Tool | What it does |
|---|---|
| `list_positions` | List positions (layouts mapped to surfaces). |
| `list_buttons` | List keys on a position: grid location, text, background, number of colour rules. |
| `get_layout` | Raw layout document and revision, for advanced edits. |
| `set_button` | Create or update the key at `x`,`y`: text, icon, background colour. |
| `delete_button` | Remove the key at `x`,`y`. |
| `set_color_rules` | Replace a key's colour rules: each rule sets the background when its expression is true. |
| `set_load_gradient` | Colour a key from green to red by a numeric variable, in 2–8 steps. |
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

## Example: a CI dashboard on a Stream Deck Studio

1. Write your numbers to files, e.g. `~/.deck-stats/ci_queued.txt`, from a script on a timer.
2. In Buttons, add a **Generic File Reader** connection per file. Its variable key (e.g. `FileReader`) exposes `$(FileReader.contents)`.
3. Ask your assistant:
   *"On the Desk position, put `CI $(FileReader.contents)` on key 0,0 and colour it green at 0 to red at 20 queued jobs."*

   It calls `set_button`, then `set_load_gradient`.

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
