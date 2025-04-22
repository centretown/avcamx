package main

import (
	"avcamx"
	"encoding/json"
	"log"
	"net/http"
)

var remote = "http://192.168.10.7:8080"

func main() {
	resp, err := http.Get(remote)
	if err != nil {
		log.Fatal(err)
	}
	remote, err := avcamx.ReadRemote(resp.Body)
	buf, err := json.MarshalIndent(remote, "", "  ")
	log.Print(string(buf))

	host := avcamx.NewAvHost("", "9000")
	host.MakeProxy(remote, 2)
	buf, err = json.MarshalIndent(host, "", "  ")
	log.Print(string(buf))
}
