# Changelog

All notable changes are listed here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses [Semantic Versioning](https://semver.org/).

## [0.1.0] - 2026-11-13

### Added

- MCP server for Bitfocus Buttons over its internal tRPC API (tested with Buttons 1.9.1 beta).
- Tools: `list_positions`, `list_buttons`, `get_layout`, `set_button`, `delete_button`, `set_color_rules`, `set_load_gradient`, `list_variables`, `call_api`.
- `buttons-mcp login` / `logout` with OS keychain storage; terminal prompt or macOS dialog.
- Safe saves using the layout revision, and automatic re-login when the session expires.

[0.1.0]: https://github.com/Almiro-Studios/buttons-mcp/releases/tag/v0.1.0
