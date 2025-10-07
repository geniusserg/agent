cmake_minimum_required(VERSION 3.16)

project(mcp_server LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

include(FetchContent)

FetchContent_Declare(
    nlohmann_json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG v3.11.3
)

FetchContent_MakeAvailable(nlohmann_json)

add_executable(mcp_server
    src/main.cpp
    src/McpServer.cpp
    src/Message.cpp
)

target_include_directories(mcp_server PRIVATE include)

target_link_libraries(mcp_server PRIVATE nlohmann_json::nlohmann_json)
