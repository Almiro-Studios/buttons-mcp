# Contributing

Thanks for helping improve buttons-mcp! Issues and pull requests are welcome.

## Before you start

- For bugs, open an issue with your Buttons version (Settings → General), what you asked the assistant or which tool you called, and what happened.
- For larger changes, open an issue first so we can agree on the approach.

## Development

Requires Go 1.26+ and a running Bitfocus Buttons install for manual testing.

```bash
make build   # builds bin/buttons-mcp
make test    # go vet + go test
```

Log in once with `bin/buttons-mcp login`, then try changes without an MCP client:

```bash
bin/buttons-mcp dump -              # all positions
bin/buttons-mcp dump <positionId>   # one saved layout document
```

## Layout of the code

| Path | Contents |
|---|---|
| `cmd/buttons-mcp` | Entry point, CLI subcommands, MCP tool definitions. |
| `internal/trpc` | tRPC client: HTTP login plus the WebSocket session. |
| `internal/layout` | Helpers that edit layout documents (keys, colour rules). |
| `internal/creds` | Login prompt and OS keychain storage. |

## Guidelines

- Keep the dedicated tools small and safe; anything exotic can go through `call_api`.
- Layout documents are edited as generic JSON so unknown fields survive. Don't add structs that drop fields.
- Every Go file starts with the copyright/SPDX header. Exported identifiers get doc comments.
- Run `gofmt`, `go vet` and the tests before opening a PR.
- Buttons' API is internal and changes between releases. If you adapt to a new version, note it in `CHANGELOG.md`.

By contributing, you agree that your contributions are licensed under the MIT License.
