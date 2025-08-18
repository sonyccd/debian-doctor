package main

import (
	"log"
	"os"

	"github.com/debian-doctor/debian-doctor/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}