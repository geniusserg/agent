#include "mcp/Message.hpp"

#include <algorithm>
#include <cctype>
#include <charconv>
#include <iostream>

namespace {

std::string trim(const std::string &value) {
    auto start = value.begin();
    while (start != value.end() && std::isspace(static_cast<unsigned char>(*start))) {
        ++start;
    }
    auto finish = value.end();
    while (finish != start && std::isspace(static_cast<unsigned char>(*(finish - 1)))) {
        --finish;
    }
    return {start, finish};
}

}  // namespace

namespace mcp {

MessageReader::MessageReader(std::istream &input) : input_(input) {}

std::optional<std::string> MessageReader::next() {
    std::string line;
    std::size_t content_length = 0;

    while (std::getline(input_, line)) {
        if (!line.empty() && line.back() == '\r') {
            line.pop_back();
        }

        if (line.empty()) {
            break;
        }

        const auto colon = line.find(':');
        if (colon == std::string::npos) {
            continue;
        }

        const auto header = trim(line.substr(0, colon));
        const auto value = trim(line.substr(colon + 1));
        if (header == "Content-Length") {
            auto result = std::from_chars(value.data(), value.data() + value.size(), content_length);
            if (result.ec != std::errc{}) {
                content_length = 0;
            }
        }
    }

    if (input_.eof() && content_length == 0) {
        return std::nullopt;
    }

    if (content_length == 0) {
        return std::nullopt;
    }

    std::string payload(content_length, '\0');
    input_.read(payload.data(), static_cast<std::streamsize>(content_length));
    if (input_.gcount() != static_cast<std::streamsize>(content_length)) {
        return std::nullopt;
    }

    return payload;
}

MessageWriter::MessageWriter(std::ostream &output) : output_(output) {}

void MessageWriter::write(const std::string &payload) {
    output_ << "Content-Length: " << payload.size() << "\r\n\r\n" << payload;
    output_.flush();
}

}  // namespace mcp
