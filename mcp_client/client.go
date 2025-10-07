package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Simple MCP (Message Communication Protocol) client example
func main() {
	ctx := context.Background()

	// Create an MCP client with HTTP transport
	serverURL := "http://localhost:8080/mcp" // MCP server URL
	transport := &mcp.HTTPTransport{
		URL: serverURL,
		Client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}

	client := mcp.NewClient(transport)

	// Initialize connection (handshake)
	if err := client.Initialize(ctx, &mcp.InitializeParams{
		ClientInfo: mcp.Implementation{
			Name:    "greeter-client",
			Version: "1.0.0",
		},
		nil,
	}); err != nil {
		log.Fatalf("init failed: %v", err)
	}
	defer client.Close()

	// List available tools
	tools, err := client.ListTools(ctx)
	if err != nil {
		log.Fatalf("list tools: %v", err)
	}
	fmt.Println("Available tools:")
	for _, t := range tools.Tools {
		fmt.Printf("- %s: %s\n", t.Name, t.Description)
	}

	// Call the “greet” tool
	params := map[string]any{"name": "Alice"}
	resp, err := client.CallTool(ctx, "greet", params)
	if err != nil {
		log.Fatalf("call tool: %v", err)
	}

	// Print the server’s response
	fmt.Printf("Server response: %+v\n", resp)
}