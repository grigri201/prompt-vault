package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUploadCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		setupFile  func(t *testing.T) string
		wantOutput []string
		wantErr    bool
	}{
		{
			name: "requires file argument",
			args: []string{"upload"},
			wantOutput: []string{
				"accepts 1 arg",
			},
			wantErr: true,
		},
		{
			name: "handles non-existent file",
			args: []string{"upload", "nonexistent.yaml"},
			wantOutput: []string{
				"file not found",
			},
			wantErr: true,
		},
		{
			name: "validates yaml format",
			args: []string{"upload"},
			setupFile: func(t *testing.T) string {
				dir := t.TempDir()
				file := filepath.Join(dir, "invalid.yaml")
				content := `invalid yaml content: [`
				if err := os.WriteFile(file, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return file
			},
			wantOutput: []string{
				"invalid",
			},
			wantErr: true,
		},
		{
			name: "validates required fields",
			args: []string{"upload"},
			setupFile: func(t *testing.T) string {
				dir := t.TempDir()
				file := filepath.Join(dir, "incomplete.yaml")
				content := `---
name: Test Prompt
---
Content without required fields`
				if err := os.WriteFile(file, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return file
			},
			wantOutput: []string{
				"required",
			},
			wantErr: true,
		},
		{
			name: "shows success message for valid file",
			args: []string{"upload"},
			setupFile: func(t *testing.T) string {
				dir := t.TempDir()
				file := filepath.Join(dir, "valid.yaml")
				content := `---
name: Test Prompt
author: testuser
category: testing
tags: [test, example]
---
Hello {name}!`
				if err := os.WriteFile(file, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return file
			},
			wantOutput: []string{
				"Uploading prompt",
			},
			wantErr: false, // Would fail in real test due to auth
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip test that requires authentication
			if tt.name == "shows success message for valid file" {
				t.Skip("Skipping test that requires authentication")
			}

			cmd := NewRootCmd()
			
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			
			// Setup file if needed
			args := tt.args
			if tt.setupFile != nil {
				file := tt.setupFile(t)
				args = append(args, file)
			}
			
			cmd.SetArgs(args)
			
			// Execute
			err := cmd.Execute()
			
			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			// Check output
			output := out.String()
			// For argument errors, check the error message
			if err != nil && tt.wantErr {
				errStr := err.Error()
				for _, want := range tt.wantOutput {
					if !strings.Contains(errStr, want) && !strings.Contains(output, want) {
						t.Errorf("Output/Error missing %q\nGot output: %s\nGot error: %v", want, output, err)
					}
				}
			} else {
				for _, want := range tt.wantOutput {
					if !strings.Contains(output, want) {
						t.Errorf("Output missing %q\nGot: %s", want, output)
					}
				}
			}
		})
	}
}

func TestUploadCommand_Help(t *testing.T) {
	cmd := NewRootCmd()
	
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"upload", "--help"})
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	
	output := out.String()
	expectedStrings := []string{
		"Upload a prompt template",
		"YAML format",
		"front matter",
	}
	
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Help text missing %q", expected)
		}
	}
}