package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/pedrospdc/datadog-mcp/internal/config"
	"github.com/pedrospdc/datadog-mcp/internal/datadog"
	"github.com/pedrospdc/datadog-mcp/internal/tools"
)

const (
	serverName    = "datadog-mcp"
	serverVersion = "v1.0.0"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create Datadog client
	ddClient := datadog.NewClient(cfg)

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)

	// Register all tools
	tools.RegisterAll(server, ddClient)

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	// Run server with stdio transport
	log.Printf("Starting %s %s", serverName, serverVersion)
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
