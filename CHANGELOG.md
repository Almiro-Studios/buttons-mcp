# Changelog

All notable changes are listed here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- `list_actions` and `set_button_actions` tools: keys can now run connection actions (ATEM cuts, Videohub routes, stream start/stop, …) on press and release.
- Setup pages per AI client (Cursor, Claude Code, Claude Desktop, Codex) on GitHub Pages, with a `cursor://` MCP install link.
- Use cases section in the README.
- Connection tools: `list_connections`, `get_connection`, `add_connection`, `configure_connection`, `delete_connection`.
- Key images in `set_button`: `site_icon` (a website's own icon), `image_url` and `image_file`; PNG, JPEG, GIF and SVG are converted to the PNG Buttons stores.
- `set_button` options `style` (e.g. `bigIcon` for a full-size icon or image) and `text_color`; image-only keys default to `bigIcon`.

### Fixed

- `call_api` rejected JSON objects as input.
- Cursor web install link now uses `cursor.com/install-mcp`.

## [0.1.0] - 2026-11-13

### Added

- MCP server for Bitfocus Buttons over its internal tRPC API (tested with Buttons 1.9.1 beta).
- Tools: `list_positions`, `list_buttons`, `get_layout`, `set_button`, `delete_button`, `set_color_rules`, `set_load_gradient`, `list_variables`, `call_api`.
- `buttons-mcp login` / `logout` with OS keychain storage; terminal prompt or macOS dialog.
- Safe saves using the layout revision, and automatic re-login when the session expires.

[Unreleased]: https://github.com/Almiro-Studios/buttons-mcp/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Almiro-Studios/buttons-mcp/releases/tag/v0.1.0
