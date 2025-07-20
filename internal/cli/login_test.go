package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoginCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		input      string
		wantOutput []string
		wantErr    bool
	}{
		{
			name:  "shows GitHub token instructions",
			args:  []string{"login"},
			input: "\n", // Just press enter without token
			wantOutput: []string{
				"GitHub Personal Access Token Setup",
				"https://github.com/settings/tokens",
				"Generate new token",
				"gist",
				"Enter your GitHub Personal Access Token:",
			},
			wantErr: true,
		},
		{
			name:  "accepts valid token format",
			args:  []string{"login"},
			input: "ghp_validtoken123\n",
			wantOutput: []string{
				"Enter your GitHub Personal Access Token:",
				"Validating token...",
			},
			wantErr: false, // Would fail in real test due to API call
		},
		{
			name:  "rejects empty token",
			args:  []string{"login"},
			input: "\n",
			wantOutput: []string{
				"Enter your GitHub Personal Access Token:",
				"Token cannot be empty",
			},
			wantErr: true,
		},
		{
			name:  "shows success message on valid token",
			args:  []string{"login"},
			input: "ghp_testtoken123\n",
			wantOutput: []string{
				"Enter your GitHub Personal Access Token:",
			},
			wantErr: false, // Would fail in real test due to API call
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that require API calls
			if tt.name == "accepts valid token format" || tt.name == "shows success message on valid token" {
				t.Skip("Skipping test that requires API call")
			}

			cmd := NewRootCmd()

			var out bytes.Buffer
			var in bytes.Buffer

			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetIn(&in)

			// Write input
			in.WriteString(tt.input)

			// Set args
			cmd.SetArgs(tt.args)

			// Execute
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

func TestLoginCommand_Help(t *testing.T) {
	cmd := NewRootCmd()

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"login", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := out.String()
	expectedStrings := []string{
		"Authenticate with GitHub using a Personal Access Token",
		"https://github.com/settings/tokens",
		"gist",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Help text missing %q", expected)
		}
	}
}
