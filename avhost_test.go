package avcamx

import (
	"net/http"
	"testing"
	"time"
)

func TestScan(t *testing.T) {

	host := NewAvHost("", "", []string{}, 1000)

	for range 60 {
		time.Sleep(time.Second * 2)
		if len(host.Streams) > 0 && host.Streams[0].IsOpened() {
			testCmd(t, host, host.Streams[0])
		}
	}

	for _, item := range host.Streams {
		t.Logf("host.Item=%s %s %v", item.Url, item.Source.Path(), item.IsOpened())
	}

	host.Quit()
}

var (
	cmds = []string{
		"/reset",
		// "/zoomin", "/zoomin", "/zoomin",
		// "/panleft", "/panleft", "/panleft",
		// "/tiltup", "/tiltup", "/tiltup",
	}
	cmdCount int
)

func testCmd(t *testing.T, host *AvHost, avStream *AvStream) {
	if cmdCount >= len(cmds) {
		cmdCount = 0
	}

	resp, err := http.Get("http://" + host.Url + avStream.Url + cmds[cmdCount])
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resp.Status, resp.Body)
	cmdCount++
}
