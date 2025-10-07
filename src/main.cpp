#include "mcp/McpServer.hpp"

#include <iostream>

int main() {
    mcp::McpServer server(std::cin, std::cout);
    return server.run();
}
