package main

import (
	"log"
	"pv/internal/di"
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
