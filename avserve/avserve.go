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

	avFlags.Print()

	host := avcamx.NewAvHost(avFlags.HostAddr, avFlags.Connect, avFlags.Remotes, 1000, nil)

	log.Printf("\nServing %s...", host.Url)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	sig := <-sigs
	log.Printf("Interrupted: %v", sig)
	host.Quit()
}
