package avcamx

import (
	"net/http"
	"testing"
	"time"
)

func TestScan(t *testing.T) {

	host := NewAvHost("", "", []string{})
	done := make(chan int)
	go host.Monitor(done)

	for range 60 {
		time.Sleep(time.Second * 2)
		if len(host.Items) > 0 && host.Items[0].IsOpened() {
			testCmd(t, host, host.Items[0])
		}
	}

	done <- 1
	time.Sleep(3 * time.Second)

	for _, item := range host.Items {
		t.Logf("host.Item=%s %s %v", item.Url, item.Source.Path(), item.IsOpened())
	}

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
