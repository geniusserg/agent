package main

import (
	"log"
	"os"

	"myredis/internal/mcp"
)

func main() {
	log.SetFlags(0)
	server := mcp.NewServer(os.Stdin, os.Stdout)
	if err := server.Run(); err != nil {
		log.Printf("mcp server exited with error: %v", err)
		os.Exit(1)
	}
}
