package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/grigri/pv/internal/service"
)

// AuthStatusCmd is the auth status command type
type AuthStatusCmd = *cobra.Command

// authStatus holds the dependencies for the status command
type authStatus struct {
	service service.AuthService
}

// NewAuthStatusCommand creates the auth status command
func NewAuthStatusCommand(authService service.AuthService) AuthStatusCmd {
	as := &authStatus{service: authService}

	return &cobra.Command{
		Use:   "status",
		Short: "Check authentication status",
		Long:  `Check your current GitHub authentication status.`,
		RunE:  as.execute,
	}
}

// execute runs the status command
func (as *authStatus) execute(cmd *cobra.Command, args []string) error {
	status, err := as.service.GetStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to check authentication status: %v\n", err)
		return err
	}

	fmt.Println(service.GetStatusMessage(status))

	return nil
}
