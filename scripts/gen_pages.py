#!/usr/bin/env python3
# Copyright 2026 Almiro Studios
# SPDX-License-Identifier: MIT
"""Generate the GitHub Pages install pages in docs/open/.

Each page installs buttons-mcp into one AI client: a one-click deeplink where
the client supports one, the exact command or config to copy otherwise, and a
"try it" link that opens the client with a starter prompt prefilled.

Run from the repo root: python3 scripts/gen_pages.py
"""
import base64
import html
import json
import pathlib
import urllib.parse

REPO = "https://github.com/Almiro-Studios/buttons-mcp"
NAME = "buttons"
SERVER = {"command": "buttons-mcp"}
TRY_PROMPT = ("Use the buttons MCP server: call list_positions, then list_buttons for my first position, "
              "and describe my Stream Deck layout.")

def q(s):
    return urllib.parse.quote(s, safe="")

cfg_b64 = base64.b64encode(json.dumps(SERVER).encode()).decode()
mcp_json = json.dumps({"mcpServers": {NAME: SERVER}}, indent=2)

CLIENTS = {
    "cursor": {
        "name": "Cursor",
        "install": {
            "button": f"cursor://anysphere.cursor-deeplink/mcp/install?name={NAME}&config={q(cfg_b64)}",
            "fallback": f"https://cursor.com/install-mcp?name={NAME}&config={q(cfg_b64)}",
        },
        "manual": ("Or add this to <code>~/.cursor/mcp.json</code>:", mcp_json),
        "try": f"cursor://anysphere.cursor-deeplink/prompt?text={q(TRY_PROMPT)}",
    },
    "claude-code": {
        "name": "Claude Code",
        "manual": ("Run in your terminal:", f"claude mcp add {NAME} --scope user -- buttons-mcp"),
        "try": f"claude-cli://open?q={q(TRY_PROMPT)}",
    },
    "claude": {
        "name": "Claude Desktop",
        "manual": ("Open Settings → Developer → Edit Config and add this to <code>claude_desktop_config.json</code>, then restart Claude:", mcp_json),
        "try": f"claude://claude.ai/new?q={q(TRY_PROMPT)}",
    },
    "codex": {
        "name": "Codex",
        "manual": ("Run in your terminal:", f"codex mcp add {NAME} -- buttons-mcp"),
        "manual2": ("Or add this to <code>~/.codex/config.toml</code>:", f'[mcp_servers.{NAME}]\ncommand = "buttons-mcp"'),
        "try": f"codex://new?prompt={q(TRY_PROMPT)}",
    },
}

CSS = """
:root{color-scheme:light dark;--bg:#fff;--fg:#1b1b1f;--muted:#5b5b66;--card:#f4f4f6;--line:#e2e2e8;--accent:#6d28d9;--on-accent:#fff}
@media (prefers-color-scheme:dark){:root{--bg:#111114;--fg:#ececf1;--muted:#a0a0ab;--card:#1d1d22;--line:#2c2c33;--accent:#a78bfa;--on-accent:#111114}}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--fg);font:16px/1.55 system-ui,-apple-system,sans-serif;padding:40px 16px}
main{max-width:680px;margin:auto}
h1{font-size:1.6rem;margin:0 0 4px}
ol{padding-left:0;list-style:none;counter-reset:s}
ol>li{counter-increment:s;border:1px solid var(--line);border-radius:10px;padding:16px 16px 16px 52px;margin:12px 0;position:relative}
ol>li::before{content:counter(s);position:absolute;left:16px;top:16px;width:24px;height:24px;border-radius:50%;background:var(--accent);color:var(--on-accent);font-weight:700;font-size:.85rem;display:grid;place-items:center}
h2{font-size:1.05rem;margin:0 0 6px}
p{margin:6px 0}
.muted{color:var(--muted)}
a{color:var(--accent)}
a.button{display:inline-block;background:var(--accent);color:var(--on-accent);padding:9px 16px;border-radius:8px;text-decoration:none;font-weight:600;margin:4px 0}
.code{position:relative}
pre{background:var(--card);padding:12px 14px;border-radius:8px;overflow-x:auto;font-size:13.5px;margin:6px 0}
button.copy{position:absolute;top:6px;right:6px;font:inherit;font-size:12px;padding:3px 8px;border-radius:6px;border:1px solid var(--line);background:var(--bg);color:var(--fg);cursor:pointer}
code{font-size:.9em}
"""

def code_block(text):
    return f'<div class="code"><pre>{html.escape(text)}</pre><button class="copy" type="button">Copy</button></div>'

def page(key, c):
    n = html.escape(c["name"])
    steps = []
    steps.append(f"""<li><h2>Install the binary</h2>
<p>Download <code>buttons-mcp</code> for your OS from <a href="{REPO}/releases/latest">Releases</a> and put it on your <code>PATH</code>, or build it with Go:</p>
{code_block("go install github.com/Almiro-Studios/buttons-mcp/cmd/buttons-mcp@latest")}</li>""")
    steps.append(f"""<li><h2>Log in to Buttons</h2>
<p>Saves your Buttons username and password in the system keychain. Run once:</p>
{code_block("buttons-mcp login")}</li>""")
    add = [f"<li><h2>Add it to {n}</h2>"]
    if "install" in c:
        add.append(f'<p><a class="button" href="{html.escape(c["install"]["button"])}">Add to {n}</a></p>')
        add.append(f'<p class="muted">Not opening? <a href="{html.escape(c["install"]["fallback"])}" target="_blank" rel="noopener">Install via cursor.com</a>.</p>')
    for k in ("manual", "manual2"):
        if k in c:
            label, text = c[k]
            add.append(f"<p>{label}</p>{code_block(text)}")
    add.append("</li>")
    steps.append("".join(add))
    steps.append(f"""<li><h2>Try it</h2>
<p>Opens {n} with a starter prompt prefilled. You review it before it runs.</p>
<p><a class="button" href="{html.escape(c["try"])}">Open in {n}</a></p></li>""")
    return f"""<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>buttons-mcp for {n}</title>
<meta name="description" content="Set up the buttons-mcp MCP server for Bitfocus Buttons in {n}.">
<style>{CSS}</style>
</head>
<body>
<main>
<h1>buttons-mcp for {n}</h1>
<p class="muted">Control Bitfocus Buttons (Stream Deck Studio) layouts from {n}. <a href="{REPO}#readme">About the project</a></p>
<ol>
{''.join(steps)}
</ol>
</main>
<script>
document.querySelectorAll('button.copy').forEach(function (b) {{
  b.addEventListener('click', function () {{
    navigator.clipboard.writeText(b.previousElementSibling.textContent).then(function () {{
      b.textContent = 'Copied'; setTimeout(function () {{ b.textContent = 'Copy'; }}, 1500);
    }});
  }});
}});
</script>
</body>
</html>
"""

def main():
    out = pathlib.Path("docs/open")
    out.mkdir(parents=True, exist_ok=True)
    for old in out.glob("*.html"):
        old.unlink()
    for key, c in CLIENTS.items():
        (out / f"{key}.html").write_text(page(key, c), encoding="utf-8")
    print("cursor web install link:", CLIENTS["cursor"]["install"]["fallback"])

if __name__ == "__main__":
    main()
