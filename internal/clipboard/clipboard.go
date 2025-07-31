package clipboard

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/grigri201/prompt-vault/internal/errors"
)

// Copy copies the given text to the system clipboard
func Copy(text string) error {
	if text == "" {
		return errors.NewValidationErrorMsg("Copy", "cannot copy empty text to clipboard")
	}

	switch runtime.GOOS {
	case "darwin":
		return copyDarwin(text)
	case "linux":
		return copyLinux(text)
	case "windows":
		return copyWindows(text)
	default:
		return errors.NewValidationErrorMsg("Copy", fmt.Sprintf("unsupported platform: %s", runtime.GOOS))
	}
}

// copyDarwin copies text to clipboard on macOS
func copyDarwin(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)

	if err := cmd.Run(); err != nil {
		return errors.NewFileSystemError("copyDarwin", err)
	}

	return nil
}

// copyLinux copies text to clipboard on Linux
func copyLinux(text string) error {
	// Try xclip first
	if isCommandAvailable("xclip") {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)

		if err := cmd.Run(); err != nil {
			return errors.NewFileSystemError("copyLinux", err)
		}

		return nil
	}

	// Try xsel as fallback
	if isCommandAvailable("xsel") {
		cmd := exec.Command("xsel", "--clipboard", "--input")
		cmd.Stdin = strings.NewReader(text)

		if err := cmd.Run(); err != nil {
			return errors.NewFileSystemError("copyLinux", err)
		}

		return nil
	}

	// Try wl-copy for Wayland
	if isCommandAvailable("wl-copy") {
		cmd := exec.Command("wl-copy")
		cmd.Stdin = strings.NewReader(text)

		if err := cmd.Run(); err != nil {
			return errors.NewFileSystemError("copyLinux", err)
		}

		return nil
	}

	return errors.NewValidationErrorMsg("copyLinux", "no clipboard utility found (xclip, xsel, or wl-copy required)")
}

// copyWindows copies text to clipboard on Windows
func copyWindows(text string) error {
	// Use PowerShell to access Windows clipboard
	cmd := exec.Command("powershell", "-command", "Set-Clipboard", "-Value", text)

	if err := cmd.Run(); err != nil {
		// Fallback to clip.exe for older Windows versions
		clipCmd := exec.Command("clip")
		clipCmd.Stdin = strings.NewReader(text)

		if err := clipCmd.Run(); err != nil {
			return errors.NewFileSystemError("copyWindows", err)
		}
	}

	return nil
}

// isCommandAvailable checks if a command is available in the system PATH
func isCommandAvailable(name string) bool {
	cmd := exec.Command("which", name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// IsAvailable checks if clipboard functionality is available on the current platform
func IsAvailable() bool {
	switch runtime.GOOS {
	case "darwin":
		return isCommandAvailable("pbcopy")
	case "linux":
		return isCommandAvailable("xclip") || isCommandAvailable("xsel") || isCommandAvailable("wl-copy")
	case "windows":
		// Windows always has clipboard support
		return true
	default:
		return false
	}
}
