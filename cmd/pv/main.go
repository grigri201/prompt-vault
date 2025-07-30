package main

import (
	"context"
	"fmt"
	"os"

	"github.com/grigri201/prompt-vault/internal/cli"
)

func main() {
	// Create container using wire
	c := buildContainer()
	if err := c.Initialize(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing: %v\n", err)
		os.Exit(1)
	}
	defer c.Cleanup()
	
	// Set global command context
	cli.SetCommandContext(cli.NewCommandContext(c))
	
	// Execute CLI
	if err := cli.NewRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
