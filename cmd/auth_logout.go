package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/grigri/pv/internal/service"
)

// AuthLogoutCmd is the auth logout command type
type AuthLogoutCmd = *cobra.Command

// authLogout holds the dependencies for the logout command
type authLogout struct {
	service service.AuthService
}

// NewAuthLogoutCommand creates the auth logout command
func NewAuthLogoutCommand(authService service.AuthService) AuthLogoutCmd {
	al := &authLogout{service: authService}

	return &cobra.Command{
		Use:   "logout",
		Short: "Logout and clear authentication",
		Long:  `Clear your stored GitHub authentication token.`,
		RunE:  al.execute,
	}
}

// execute runs the logout command
func (al *authLogout) execute(cmd *cobra.Command, args []string) error {
	if err := al.service.Logout(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to logout: %v\n", err)
		return err
	}

	fmt.Println(service.GetLogoutMessage())

	return nil
}
