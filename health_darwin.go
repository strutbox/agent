package main

import (
	"log"
	"os"
)

func handleHealthFailure() {
	log.Println("health: unable to connect to service")
	os.Exit(0)
}
