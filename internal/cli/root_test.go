package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantOutput []string
		wantErr    bool
	}{
		{
			name: "shows help with no args",
			args: []string{},
			wantOutput: []string{
				"Prompt Vault",
				"Available Commands:",
				"login",
				"upload",
				"list",
				"get",
				"delete",
				"sync",
			},
			wantErr: false,
		},
		{
			name: "shows help with --help flag",
			args: []string{"--help"},
			wantOutput: []string{
				"Prompt Vault",
			},
			wantErr: false,
		},
		{
			name: "shows version with --version flag",
			args: []string{"--version"},
			wantOutput: []string{
				"pv version",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create root command
			cmd := NewRootCmd()

			// Capture output
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs(tt.args)

			// Execute command
			err := cmd.Execute()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check output
			output := out.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing %q\nGot: %s", want, output)
				}
			}
		})
	}
}

func TestCommandStructure(t *testing.T) {
	cmd := NewRootCmd()

	// Check that all expected commands exist
	expectedCommands := []string{
		"login",
		"upload",
		"list",
		"get",
		"delete",
		"sync",
	}

	commands := cmd.Commands()
	commandMap := make(map[string]bool)
	for _, c := range commands {
		commandMap[c.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandMap[expected] {
			t.Errorf("Missing expected command: %s", expected)
		}
	}
}

func TestCommandAliases(t *testing.T) {
	cmd := NewRootCmd()

	// Test command aliases
	aliasTests := []struct {
		command string
		alias   string
	}{
		{"list", "ls"},
		{"upload", "up"},
		{"delete", "del"},
	}

	for _, tt := range aliasTests {
		t.Run(tt.command, func(t *testing.T) {
			var found bool
			for _, c := range cmd.Commands() {
				if c.Name() == tt.command {
					for _, alias := range c.Aliases {
						if alias == tt.alias {
							found = true
							break
						}
					}
					break
				}
			}
			if !found {
				t.Errorf("Command %s missing alias %s", tt.command, tt.alias)
			}
		})
	}
}
