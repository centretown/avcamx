package avcamx

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func FetchRemote(remoteAddr string) (host *AvHost, err error) {
	host = NewAvHost("", "9000")
	var (
		resp   *http.Response
		remote *AvHost
	)

	resp, err = http.Get(remoteAddr)
	if err != nil {
		log.Print(err)
		return
	}

	remote, err = ReadRemote(resp.Body)
	if err != nil {
		log.Print(err)
		return
	}
	host.MakeProxy(remote, 2)
	return
}

func ReadRemote(r io.Reader) (host *AvHost, err error) {
	host = &AvHost{}
	var buf []byte
	buf, err = io.ReadAll(r)
	if err != nil {
		log.Print(err)
		return
	}
	err = json.Unmarshal(buf, host)
	return
}
