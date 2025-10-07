#pragma once

#include <istream>
#include <optional>
#include <ostream>
#include <string>

namespace mcp {

class MessageReader {
public:
    explicit MessageReader(std::istream &input);

    std::optional<std::string> next();

private:
    std::istream &input_;
};

class MessageWriter {
public:
    explicit MessageWriter(std::ostream &output);

    void write(const std::string &payload);

private:
    std::ostream &output_;
};

}  // namespace mcp
