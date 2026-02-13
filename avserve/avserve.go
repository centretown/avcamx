package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/centretown/avcamx"
)

func main() {
	avFlags := avcamx.NewAvFlags()
	exists := avFlags.HasFile()
	if exists {
		avFlags.Load()
	}

	avFlags.Parse()
	err := avFlags.Save()
	if err != nil {
		log.Printf("Error saving configuration file %s. %s", avcamx.ConfigName, err)
	} else if exists {
		log.Print("Saved configuration file. ", avcamx.ConfigName)
	} else {
		log.Print("Created configuration file. ", avcamx.ConfigName)
	}

	avFlags.Print()

	host := avcamx.NewAvHost(avFlags.HostAddr, avFlags.HostPort, avFlags.Remotes, 1000)

	log.Printf("\nServing %s...", host.Url)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	sig := <-sigs
	log.Printf("Interrupted: %v", sig)
	host.Quit()
}
