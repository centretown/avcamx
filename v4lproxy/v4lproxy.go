package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/centretown/avcamx"
)

func main() {
	var (
		remoteAddr      = "http://192.168.10.197:8080"
		remoteAddrUsage = "remote camera ip address"
		host            = avcamx.NewAvHost("", "9000")
	)

	flag.StringVar(&remoteAddr, "remote", remoteAddr, remoteAddrUsage)
	flag.StringVar(&remoteAddr, "r", remoteAddr, remoteAddrUsage)
	flag.Parse()

	remote, err := host.FetchRemote(remoteAddr)
	if err != nil {
		log.Fatal(err)
	}
	host.MakeProxy(remote)

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
