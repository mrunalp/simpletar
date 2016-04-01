package main

import (
	"log"
	"os"

	"github.com/mrunalp/simpletar"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: simpletar <source> <destination>\n")
	}

	if err := simpletar.Tar(os.Args[1], os.Args[2]); err != nil {
		log.Fatalf("Failed to create tar file: %v", err)
	}
}
