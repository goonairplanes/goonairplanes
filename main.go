package main

import (
	"flag"
	"goonairplanes/core"
	"log"
)

func main() {

	port := flag.String("port", core.AppConfig.Port, "Port to run the server on")
	flag.Parse()

	core.AppConfig.Port = *port

	app := core.NewApp()

	err := app.Init()
	if err != nil {
		log.Fatalf("Failed to initialize Go on Airplanes: %v", err)
	}

	err = app.Start()
	if err != nil {
		log.Fatalf("Go on Airplanes server error: %v", err)
	}
}
