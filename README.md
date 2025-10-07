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

### Run

The server communicates over STDIN/STDOUT using MCP's JSON-RPC framing, so it is typically launched by an MCP-compatible client. For manual experimentation you can pipe JSON messages:

```bash
printf 'Content-Length: 85\r\n\r\n{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"clientInfo":{"name":"demo"}}}\n' \
  | build/mcp_server
```

You should receive a matching JSON-RPC response containing the advertised capabilities.

## Project Layout
- `src/main.cpp` – Entrypoint that wires STDIN/STDOUT into the server runtime.
- `include/mcp` – Public headers for the message transport and server logic.
- `src/McpServer.cpp` – Core JSON-RPC handling and example `echo` tool implementation.

## Extending the Server
1. Add new tools by updating `McpServer::listTools` and `McpServer::callTool`.
2. Expand `makeCapabilities` to advertise new MCP surfaces (resources, prompts, etc.).
3. If you need to initiate requests toward the client, add helper methods that build and send notifications via `writer_`.

The structure keeps allocations minimal and defers heavy work until it is genuinely needed, making it suitable for deployment on constrained hardware.
