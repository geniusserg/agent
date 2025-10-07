#include "mcp/McpServer.hpp"

#include <iostream>
#include <stdexcept>

namespace mcp {

namespace {
constexpr int kParseError = -32700;
constexpr int kInvalidRequest = -32600;
constexpr int kMethodNotFound = -32601;
constexpr int kInternalError = -32603;
}  // namespace

McpServer::McpServer(std::istream &input, std::ostream &output)
    : reader_(input), writer_(output) {}

int McpServer::run() {
    while (auto raw = reader_.next()) {
        try {
            auto message = nlohmann::json::parse(*raw);
            handle(message);
        } catch (const nlohmann::json::exception &error) {
            std::cerr << "[mcp] failed to parse incoming payload: " << error.what() << '\n';
            nlohmann::json response{
                {"jsonrpc", "2.0"},
                {"id", nullptr},
                {"error", {
                             {"code", kParseError},
                             {"message", "Unable to parse JSON payload"},
                         }},
            };
            writer_.write(response.dump());
        } catch (const std::exception &error) {
            std::cerr << "[mcp] unexpected exception: " << error.what() << '\n';
        }
    }

    return 0;
}

void McpServer::handle(const nlohmann::json &message) {
    if (!message.contains("jsonrpc") || message.at("jsonrpc") != "2.0") {
        std::cerr << "[mcp] dropping message without jsonrpc version\n";
        return;
    }

    if (message.contains("method")) {
        if (message.contains("id")) {
            handleRequest(message);
        } else {
            handleNotification(message);
        }
        return;
    }

    if (message.contains("result") || message.contains("error")) {
        // This skeleton server does not currently initiate requests, so responses are ignored.
        return;
    }

    std::cerr << "[mcp] unrecognized message shape: " << message.dump() << '\n';
}

void McpServer::handleRequest(const nlohmann::json &request) {
    const auto &method = request.at("method").get_ref<const std::string &>();
    const auto &id = request.at("id");

    try {
        if (method == "initialize") {
            initialized_ = true;
            respond(id, {{"capabilities", makeCapabilities()},
                         {"serverInfo", {
                                            {"name", "cpp-mcp-server"},
                                            {"version", "0.1.0"},
                                        }}} );
            return;
        }

        if (!initialized_) {
            respondWithError(id, kInvalidRequest, "Server has not been initialized");
            return;
        }

        if (method == "ping") {
            respond(id, {{"result", "pong"}});
            return;
        }

        if (method == "tools/list") {
            respond(id, listTools());
            return;
        }

        if (method == "tools/call") {
            if (!request.contains("params")) {
                respondWithError(id, kInvalidRequest, "Missing params for tools/call");
                return;
            }
            respond(id, callTool(request.at("params")));
            return;
        }

        if (method == "shutdown") {
            respond(id, nlohmann::json::object());
            return;
        }

        respondWithError(id, kMethodNotFound, "Method not implemented: " + method);
    } catch (const nlohmann::json::exception &error) {
        respondWithError(id, kInternalError, error.what());
    }
}

void McpServer::handleNotification(const nlohmann::json &notification) {
    const auto &method = notification.at("method").get_ref<const std::string &>();

    if (method == "notifications/initialized") {
        std::cerr << "[mcp] client signalled that initialization is complete\n";
        return;
    }

    std::cerr << "[mcp] ignoring notification: " << method << '\n';
}

void McpServer::respond(const nlohmann::json &id, const nlohmann::json &result) {
    nlohmann::json response{
        {"jsonrpc", "2.0"},
        {"id", id},
        {"result", result},
    };
    writer_.write(response.dump());
}

void McpServer::respondWithError(const nlohmann::json &id, int code, const std::string &message) {
    nlohmann::json response{
        {"jsonrpc", "2.0"},
        {"id", id},
        {"error", {
                      {"code", code},
                      {"message", message},
                  }},
    };
    writer_.write(response.dump());
}

nlohmann::json McpServer::makeCapabilities() const {
    return {
        {"tools", {
                      {"list", true},
                      {"call", true},
                  }},
    };
}

nlohmann::json McpServer::listTools() const {
    return {
        {"tools",
         nlohmann::json::array(
             {nlohmann::json{
                 {"name", "echo"},
                 {"description", "Return the same text that the caller provides."},
                 {"inputSchema",
                  {
                      {"type", "object"},
                      {"properties",
                       {
                           {"text",
                            {
                                {"type", "string"},
                                {"description", "Text to echo back to the caller."},
                            }},
                       }},
                      {"required", nlohmann::json::array({"text"})},
                  }},
             }})},
    };
}

nlohmann::json McpServer::callTool(const nlohmann::json &params) {
    const auto name = params.at("name").get<std::string>();

    if (name != "echo") {
        throw std::runtime_error("Unknown tool: " + name);
    }

    const auto &arguments = params.at("arguments");

    const auto text = arguments.value("text", "");

    return {
        {"content",
         nlohmann::json::array(
             {nlohmann::json{
                 {"type", "text"},
                 {"text", text},
             }})},
    };
}

}  // namespace mcp
