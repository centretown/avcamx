package main

import (
	"avcamx"
	"encoding/json"
	"log"
	"os"
	"os/signal"
)

func main() {
	var remote = "http://192.168.10.197:8080"
	host := avcamx.NewAvHost("", "9000")

	err := host.FetchRemote(remote)
	if err != nil {
		log.Fatal(err)
	}

	buf, err := json.MarshalIndent(host, "", "  ")
	if err != nil {
		log.Print(err)
		return
	}

	started := true

	log.Print(string(buf))

	if started {

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
}
