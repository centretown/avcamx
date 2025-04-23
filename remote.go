package avcamx

import (
	"encoding/json"
	"io"
	"log"
)

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
