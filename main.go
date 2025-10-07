package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Args struct{ 
	Name string `json:"name" jsonschema:"Name to greet"`
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "v1.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "greet", Description: "say hi"}, 
	func(ctx context.Context, req *mcp.CallToolRequest, args Args) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Hi" + args.Name + "!"},
		},
	}, nil, nil
  },)

	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{JSONResponse: true})

	mux := http.NewServeMux()
	mux.Handle("/mcp", handler)

	srv := &http.Server{
		Addr:    ":8080", // public port
		Handler: mux,
	}

	fmt.Println("Listening on http://0.0.0.0:8080/mcp")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
