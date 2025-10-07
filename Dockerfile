FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/mcpwebui ./cmd/mcpwebui

FROM alpine:3.19
RUN adduser -S -D -H mcp
USER mcp
WORKDIR /app
COPY --from=builder /out/mcpwebui /usr/local/bin/mcpwebui
COPY webui /app/webui
ENV WEBUI_ASSETS_DIR=/app/webui
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/mcpwebui"]
