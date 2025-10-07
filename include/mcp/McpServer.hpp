#pragma once

#include "mcp/Message.hpp"

#include <nlohmann/json.hpp>

#include <istream>
#include <optional>
#include <ostream>
#include <string>

namespace mcp {

class McpServer {
public:
    McpServer(std::istream &input, std::ostream &output);

    int run();

private:
    void handle(const nlohmann::json &message);
    void handleRequest(const nlohmann::json &request);
    void handleNotification(const nlohmann::json &notification);
    void respond(const nlohmann::json &id, const nlohmann::json &result);
    void respondWithError(const nlohmann::json &id, int code, const std::string &message);
    nlohmann::json makeCapabilities() const;
    nlohmann::json listTools() const;
    nlohmann::json callTool(const nlohmann::json &params);

    MessageReader reader_;
    MessageWriter writer_;
    bool initialized_{false};
};

}  // namespace mcp
