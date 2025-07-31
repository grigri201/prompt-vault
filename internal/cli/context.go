package cli

import (
	"github.com/grigri201/prompt-vault/internal/container"
)

// CommandContext holds the shared dependencies for all commands
type CommandContext struct {
	Container *container.Container
}

// NewCommandContext creates a new command context with the given container
func NewCommandContext(c *container.Container) *CommandContext {
	return &CommandContext{
		Container: c,
	}
}

// Global command context - initialized in Execute()
var cmdContext *CommandContext

// GetCommandContext returns the global command context
func GetCommandContext() *CommandContext {
	if cmdContext == nil {
		panic("command context not initialized")
	}
	return cmdContext
}

// SetCommandContext sets the global command context
func SetCommandContext(ctx *CommandContext) {
	cmdContext = ctx
}
