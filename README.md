# Go MCP Server Skeleton

This repository contains a minimal [Model Context Protocol](https://github.com/modelcontextprotocol/specification) server written in Go and designed to run with a very small resource footprint. It speaks JSON-RPC 2.0 over STDIN/STDOUT, exposes a single sample tool (`echo`), and ships with a compact web UI for manual experimentation.

## Features
- Pure Go implementation with no external runtime dependencies.
- Lightweight message transport that understands `Content-Length` framed JSON over STDIN/STDOUT.
- Basic JSON-RPC dispatch with initialization handshake, tool listing, and tool invocation flow.
- Tiny HTTP server that serves the bundled web console and exposes the demo tool over REST.

## Getting Started

### Prerequisites
- Go 1.22 or newer.

### Build

```bash
go build ./cmd/mcpserver
go build ./cmd/mcpwebui
```

### Run (JSON-RPC transport)

The `mcpserver` binary communicates over STDIN/STDOUT using MCP's JSON-RPC framing, so it is typically launched by an MCP-compatible client. For manual experimentation you can pipe JSON messages:

```bash
printf 'Content-Length: 85\r\n\r\n{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"clientInfo":{"name":"demo"}}}\n' \
  | ./mcpserver
```

You should receive a matching JSON-RPC response containing the advertised capabilities.

### Run (web UI)

Launch the HTTP server to interact with the sample `echo` tool from a browser:

```bash
./mcpwebui
```

By default the UI is served from `http://localhost:8080` and the static assets are read from the `webui/` directory. Both can be customised via environment variables:

```bash
WEBUI_PORT=9090 WEBUI_ASSETS_DIR=/custom/assets ./mcpwebui
```

The server exposes `/api/echo` for the UI (or other HTTP clients) and `/healthz` for simple health checks.

## Docker

A minimal Dockerfile is provided for the web UI. Build and run it locally with:

```bash
docker build -t mcp-web-ui .
docker run --rm -p 8080:8080 mcp-web-ui
```

## Project Layout
- `cmd/mcpserver` – Entrypoint that wires STDIN/STDOUT into the server runtime.
- `internal/mcp` – Core JSON-RPC handling and example `echo` tool implementation.
- `cmd/mcpwebui` – Lightweight HTTP server that serves the UI and exposes the echo tool over REST.
- `webui/` – Static assets for the in-browser console.
- `.github/workflows/ci.yml` – Continuous integration pipeline running formatting, vetting, and tests.

## Development

Run the included checks before opening a pull request:

```bash
go test ./...
go vet ./...
go fmt ./...
```

The CI workflow enforces formatting, vetting, and unit tests to keep the codebase consistent and reliable.

## Deployment Pipeline

A GitHub Actions workflow in `.github/workflows/deploy.yml` handles automated deployment of the MCP server once changes reach the `main` branch. The pipeline performs formatting checks, runs the Go test suite, and, on success, builds the `mcpserver` binary and ships it to the remote host before restarting the service.

Before enabling deployments, create the following repository secrets so the workflow can authenticate with your target server:

| Secret | Description |
| --- | --- |
| `MCP_SERVER_HOST` | Public hostname or IP address of the deployment target (for example `64.188.97.146`). |
| `MCP_SERVER_USER` | SSH user with permission to deploy the service (for example `codex`). |
| `MCP_SERVER_PASSWORD` | Password for the SSH user. |

Once the secrets are configured, pushes to `main` or manual workflow dispatches will upload the freshly built binary to `/tmp/mcpserver` on the target, install it to `/usr/local/bin`, and restart the associated `systemd` service.
