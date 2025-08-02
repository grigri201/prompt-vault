package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type RootCmd = *cobra.Command

func NewRootCommand(lc ListCmd) RootCmd {
	root := &cobra.Command{
		Use:   "pv",
		Short: "Prompt Vault CLI",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello, pv!")
		},
	}
	root.AddCommand(lc)
	return root
}
