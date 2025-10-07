#include "third_party/httplib.h"

#include <cstdint>
#include <cstdlib>
#include <filesystem>
#include <iostream>
#include <stdexcept>
#include <string>
#include <string_view>

#include <nlohmann/json.hpp>

namespace {
std::filesystem::path resolve_assets_dir() {
    if (const char* override_path = std::getenv("WEBUI_ASSETS_DIR")) {
        std::filesystem::path candidate{override_path};
        if (std::filesystem::exists(candidate)) {
            return candidate;
        }
        std::cerr << "[warn] WEBUI_ASSETS_DIR=" << candidate << " does not exist. Falling back to compiled default.\n";
    }

#ifdef WEBUI_ASSETS_DIR
    std::filesystem::path compiled{WEBUI_ASSETS_DIR};
    if (std::filesystem::exists(compiled)) {
        return compiled;
    }
#endif

    auto current = std::filesystem::current_path();
    auto fallback = current / "webui";
    if (std::filesystem::exists(fallback)) {
        return fallback;
    }

    throw std::runtime_error("Unable to locate web UI assets directory");
}

std::uint16_t resolve_port() {
    if (const char* port_str = std::getenv("WEBUI_PORT")) {
        try {
            int parsed = std::stoi(port_str);
            if (parsed > 0 && parsed < 65536) {
                return static_cast<std::uint16_t>(parsed);
            }
            std::cerr << "[warn] Ignoring invalid WEBUI_PORT value: " << port_str << "\n";
        } catch (const std::exception& ex) {
            std::cerr << "[warn] Failed to parse WEBUI_PORT='" << port_str << "': " << ex.what() << "\n";
        }
    }
    return 8080;
}

nlohmann::json make_echo_response(std::string_view message) {
    return nlohmann::json{{"result", nlohmann::json{{"tool", "echo"}, {"message", message}}}};
}

} // namespace

int main() {
    httplib::Server server;

    const auto assets_dir = resolve_assets_dir();
    if (!server.set_mount_point("/", assets_dir.string().c_str())) {
        std::cerr << "[error] Failed to mount web assets at " << assets_dir << "\n";
        return 1;
    }

    server.Get("/healthz", [](const httplib::Request&, httplib::Response& res) {
        res.set_content("ok", "text/plain");
    });

    server.Post("/api/echo", [](const httplib::Request& req, httplib::Response& res) {
        try {
            auto json = nlohmann::json::parse(req.body, nullptr, true, true);
            const auto message_it = json.find("message");
            if (message_it == json.end() || !message_it->is_string()) {
                res.status = 400;
                res.set_content("{\"error\":\"Missing message\"}", "application/json");
                res.set_header("Access-Control-Allow-Origin", "*");
                return;
            }

            const auto response = make_echo_response(message_it->get<std::string>());
            res.set_content(response.dump(2), "application/json");
            res.set_header("Access-Control-Allow-Origin", "*");
        } catch (const std::exception& ex) {
            nlohmann::json error{{"error", ex.what()}};
            res.status = 400;
            res.set_content(error.dump(), "application/json");
            res.set_header("Access-Control-Allow-Origin", "*");
        }
    });

    server.Options(R"(/api/echo)", [](const httplib::Request&, httplib::Response& res) {
        res.set_header("Access-Control-Allow-Origin", "*");
        res.set_header("Access-Control-Allow-Headers", "Content-Type");
        res.set_header("Access-Control-Allow-Methods", "POST, OPTIONS");
        res.status = 200;
    });

    server.set_logger([](const auto& req, const auto& res) {
        std::clog << req.method << ' ' << req.path << " -> " << res.status << '\n';
    });

    server.set_error_handler([](const httplib::Request& req, httplib::Response& res) {
        nlohmann::json payload{{"error", {{"status", res.status}, {"path", req.path}}}};
        res.set_content(payload.dump(), "application/json");
    });

    const auto port = resolve_port();
    std::cout << "Serving MCP web UI from " << assets_dir << " on 0.0.0.0:" << port << '\n';
    std::cout << "Visit http://localhost:" << port << " in your browser.\n";

    server.listen("0.0.0.0", port);
    return 0;
}
