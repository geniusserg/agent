# C++ MCP Server Skeleton

This repository contains a minimal [Model Context Protocol](https://github.com/modelcontextprotocol/specification) server written in C++ and designed to run with a very small resource footprint. It speaks JSON-RPC 2.0 over STDIN/STDOUT, exposes a single sample tool (`echo`), and provides a clear starting point for expanding the server with your own capabilities.

## Features
- Modern CMake build configured for C++20.
- Thin message transport layer that understands `Content-Length` framed JSON over STDIN/STDOUT.
- Basic JSON-RPC dispatch with initialization handshake, tool listing, and tool invocation flow.
- Header-only dependency management via `FetchContent` to keep the final binary lean while avoiding manual vendoring.

## Getting Started

### Prerequisites
- A C++20 compiler (GCC 11+, Clang 13+, or MSVC 19.30+).
- CMake 3.16 or newer.
- Git (used by CMake to fetch the `nlohmann/json` dependency).

### Build
```bash
cmake -S . -B build
cmake --build build
```

The resulting executable will be located at `build/mcp_server`.

### Run (JSON-RPC transport)

The `mcp_server` executable communicates over STDIN/STDOUT using MCP's JSON-RPC framing, so it is typically launched by an MCP-compatible client. For manual experimentation you can pipe JSON messages:

```bash
printf 'Content-Length: 85\r\n\r\n{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"clientInfo":{"name":"demo"}}}\n' \
  | build/mcp_server
```

You should receive a matching JSON-RPC response containing the advertised capabilities.

### Run (web UI)

The repository also ships with a small web console that exposes the demo `echo` tool over HTTP. Launch it with:

```bash
cmake --build build --target mcp_web_ui
./build/mcp_web_ui
```

By default the UI is served from `http://localhost:8080` and the static assets are read from the `webui/` directory. Both can be customised via environment variables:

```bash
WEBUI_PORT=9090 WEBUI_ASSETS_DIR=/custom/assets ./build/mcp_web_ui
```

The server exposes `/api/echo` for the UI (or other HTTP clients) and `/healthz` for simple health checks.

## Project Layout
- `src/main.cpp` – Entrypoint that wires STDIN/STDOUT into the server runtime.
- `src/WebUiServer.cpp` – Lightweight HTTP server that serves the UI and exposes the echo tool over REST.
- `include/mcp` – Public headers for the message transport and server logic.
- `include/third_party` – Vendored single-header dependency (`cpp-httplib`) used by the web UI.
- `src/McpServer.cpp` – Core JSON-RPC handling and example `echo` tool implementation.
- `webui/` – Static assets for the in-browser console.

## Extending the Server
1. Add new tools by updating `McpServer::listTools` and `McpServer::callTool`.
2. Expand `makeCapabilities` to advertise new MCP surfaces (resources, prompts, etc.).
3. If you need to initiate requests toward the client, add helper methods that build and send notifications via `writer_`.

## Deploying the Web UI

The repository includes a `Dockerfile` that builds the `mcp_web_ui` binary and serves the static assets. Build and run it locally:

```bash
docker build -t mcp-web-ui .
docker run --rm -p 8080:8080 mcp-web-ui
```

To publish the UI to the internet you can deploy the image to any container hosting platform (Fly.io, Railway, Render, etc.). Expose port `8080` and point your domain or generated URL to the running container. Health probes can target the `/healthz` endpoint, while end users interact with the UI at `/`.

The structure keeps allocations minimal and defers heavy work until it is genuinely needed, making it suitable for deployment on constrained hardware.
