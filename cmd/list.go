package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/grigri/pv/internal/infra"
)

type ListCmd *cobra.Command

type list struct {
	store infra.Store
}

func (lc *list) execute(cmd *cobra.Command, args []string) {
	var prompts, err = lc.store.List()
	if err != nil {
		log.Fatalln("error in get prompts")
	}
	for i := range prompts {
		var prompt = prompts[i]
		fmt.Printf("%s by: %s", prompt.Name, prompt.Author)
	}
}

func NewListCommand(store infra.Store) ListCmd {
	lc := &list{store: store}
	return &cobra.Command{
		Use: "list",
		Run: lc.execute,
	}
}
