package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type RootCmd = *cobra.Command

func NewRootCommand(lc ListCmd, addCmd AddCmd, deleteCmd DeleteCmd, getCmd GetCmd, syncCmd SyncCmd, authCmd AuthCmd, shareCmd *cobra.Command) RootCmd {
	root := &cobra.Command{
		Use:   "pv",
		Short: "Prompt Vault CLI",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello, pv!")
		},
	}
	root.AddCommand(lc, addCmd, deleteCmd, getCmd, syncCmd, authCmd, shareCmd)
	
	// Create 'del' alias for delete command
	delCmd := &cobra.Command{
		Use:     "del [keyword|gist-url]",
		Short:   "删除存储的提示 (delete 命令的别名)",
		Long:    deleteCmd.Long,
		Example: deleteCmd.Example,
		Args:    deleteCmd.Args,
		Run:     deleteCmd.Run,
	}
	root.AddCommand(delCmd)
	
	return root
}
