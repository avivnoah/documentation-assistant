package main

import (
	"context"
	"fmt"
	"helpers"
	"logging"

	"os"

	"github.com/avivnoah/documentation-assistant/server"
)

func main() {
	helpers.LoadDotEnv(".env")
	logger := logging.New()
	ctx := context.Background()

	config := server.Config{
		PineconeHost: os.Getenv("PINECONE_HOST"),
		Port:         os.Getenv("PORT"),
	}

	srv, err := server.NewServer(ctx, logger, config)
	if err != nil {
		logger.Error(ctx, "Failed to create server", map[string]any{"error": err.Error()})
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := srv.Start(ctx); err != nil {
		logger.Error(ctx, "Server stopped with error", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
}
