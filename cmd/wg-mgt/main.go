package main

import (
	"log"

	"github.com/slchris/wg-mgt/internal/app"
)

func main() {
	webFS, err := GetWebFS()
	if err != nil {
		log.Fatalf("Failed to get web filesystem: %v", err)
	}

	application, err := app.New(webFS)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
