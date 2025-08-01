package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/centretown/avcamx"
)

func main() {
	host := avcamx.NewAvHost("", "")
	host.MakeLocal(nil)

	httpErr := make(chan error, 1)
	go func() {
		httpErr <- host.ListenAndServe()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-httpErr:
		log.Printf("failed to serve http: %v", err)
	case sig := <-sigs:
		log.Printf("terminating: %v", sig)
	}

	host.Quit()
}
