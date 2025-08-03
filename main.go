package main

import (
	"log"

	"github.com/grigri/pv/internal/di"
)

func main() {
	root, err := di.BuildCLI()
	if err != nil {
		log.Fatal(err)
	}
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
