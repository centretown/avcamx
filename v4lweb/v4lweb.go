package main

import (
	avcam "avcamx"
	"encoding/json"
	"log"
	"os"
	"os/signal"
)

func main() {
	host := avcam.NewAvHost("", "")
	host.Load()

	buf, err := json.MarshalIndent(host, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s\n", string(buf))

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
