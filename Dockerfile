# syntax=docker/dockerfile:1.6

FROM alpine:3.19 AS build

RUN apk add --no-cache \
    build-base \
    cmake \
    git

WORKDIR /app

COPY . .

RUN cmake -S . -B build -DCMAKE_BUILD_TYPE=Release \
    && cmake --build build --target mcp_web_ui --config Release

FROM alpine:3.19

RUN apk add --no-cache libstdc++

WORKDIR /opt/mcp

COPY --from=build /app/build/mcp_web_ui /usr/local/bin/mcp_web_ui
COPY webui /opt/mcp/webui

ENV WEBUI_ASSETS_DIR=/opt/mcp/webui \
    WEBUI_PORT=8080

EXPOSE 8080

CMD ["mcp_web_ui"]
