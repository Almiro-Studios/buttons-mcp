# Security policy

## Reporting a vulnerability

Please do **not** open a public issue for security problems. Use GitHub's private reporting instead: **Security → Report a vulnerability** on this repository.

We aim to reply within a week.

## How credentials are handled

- `buttons-mcp login` stores the Buttons username and password in the operating system keychain, keyed by server URL. They are never written to disk in plain text by this tool.
- `BUTTONS_USERNAME` / `BUTTONS_PASSWORD` environment variables override the keychain. Treat them as secrets.
- By default the server accepts Buttons' self-signed TLS certificate (`BUTTONS_INSECURE_TLS=true`), which is appropriate for `localhost`. Use a trusted certificate and set it to `false` when Buttons runs on another machine.
- `call_api` can run any mutation the logged-in user is allowed to run. Use an account with only the permissions you need.
